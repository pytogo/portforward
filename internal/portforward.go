package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	// Auth plugins - common and cloud provider
	_ "github.com/Azure/go-autorest/autorest/adal"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// ===== Management of open connections =====

/*
Thoughts:

Global states are bad but should any reference to memory exists in
the Python and Go space? Who should when free the memory?

Every space should keep the ownership of its memory allocations.
Parameters are passed from Python to Go but Go never owns them.
*/
var (
	activeForwards  = make(map[string]chan struct{})
	podsForServices = make(map[string]string)
	mutex           sync.Mutex
)

// registerForwarding adds a forwarding to the active forwards.
func registerForwarding(namespace, pod, podOrService string, toPort int, stopCh chan struct{}) {
	key := fmt.Sprintf("%s/%s/%d", namespace, pod, toPort)
	debugPortforward(fmt.Sprintf("Register pod key %s", key))

	mutex.Lock()
	defer mutex.Unlock()

	if otherCh, ok := activeForwards[key]; ok {
		close(otherCh)
	}

	// When they are not equal then we received a service name.
	if pod != podOrService {
		serviceKey := fmt.Sprintf("%s/%s/%d", namespace, podOrService, toPort)
		debugPortforward(fmt.Sprintf("Register service key %s with %s", serviceKey, pod))
		podsForServices[serviceKey] = pod
	}

	activeForwards[key] = stopCh
}

// StopForwarding closes a port forwarding.
func StopForwarding(namespace, podOrService string, toPort int) {
	key := fmt.Sprintf("%s/%s/%d", namespace, podOrService, toPort)
	debugPortforward(fmt.Sprintf("Look up pod key %s", key))

	mutex.Lock()
	defer mutex.Unlock()

	if stopChannel, ok := activeForwards[key]; ok {
		close(stopChannel)
		delete(activeForwards, key)
		debugPortforward(fmt.Sprintf("Stopped forward for key %s", key))
	}

	// We did not find a stopChannel. Was it maybe a service
	// and was registered with the actual target pod name?

	serviceKey := fmt.Sprintf("%s/%s/%d", namespace, podOrService, toPort)
	debugPortforward(fmt.Sprintf("Look up service key %s", serviceKey))

	if podForService, ok := podsForServices[serviceKey]; ok {
		key = fmt.Sprintf("%s/%s/%d", namespace, podForService, toPort)
		debugPortforward(fmt.Sprintf("Look up pod key %s", key))

		if stopChannel, ok := activeForwards[key]; ok {
			close(stopChannel)
			delete(activeForwards, key)
			debugPortforward(fmt.Sprintf("Stopped forward for key %s", key))
		}

		delete(podsForServices, serviceKey)
	}
}

// ===== Port forwarding =====

// Forward connects to a pod/service and tunnels traffic from a local port to this pod.
func Forward(namespace, podOrService string, fromPort, toPort int, configPath string, logLevel int, kubeContext string) error {
	// LOGGING
	log := newLogger(logLevel)
	overwriteLog(log)

	// Based on example https://github.com/kubernetes/client-go/issues/51#issuecomment-436200428

	// CONFIG
	config, err := loadConfig(configPath, kubeContext, log)

	if err != nil {
		return err
	}

	// PREPARE
	// Check & prepare name
	// PortForward must be started in a go-routine, therefore we have
	// to check manually if the pod or service exists and is reachable.
	err = checkPort(fromPort, log)

	if err != nil {
		return err
	}

	pod, err := getPod(config, namespace, podOrService, log)

	if err != nil {
		return err
	}

	// DIALER
	dialer, err := newDialer(config, namespace, pod)

	if err != nil {
		return err
	}

	// PORT FORWARD
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)

	ports := fmt.Sprintf("%d:%d", fromPort, toPort)

	if err = startForward(dialer, ports, stopChan, readyChan, log); err != nil {
		return err
	}

	// HANDLE CLOSING
	registerForwarding(namespace, pod.Name, podOrService, toPort, stopChan)
	closeOnSigterm(namespace, pod.Name, toPort)

	return nil
}

// loadConfig fetches the config from .kube config folder inside the home dir.
// It tries to load in-cluster-config when an empty path was provided.
func loadConfig(kubeconfigPath string, kubeContext string, log logger) (*rest.Config, error) {
	if kubeconfigPath == "" {
		log.Debug("An empty config path was provide - will try to use in-cluster-config")
		return rest.InClusterConfig()
	}

	var configOverrides *clientcmd.ConfigOverrides

	if kubeContext != "" {
		log.Debug("Override kube context with " + kubeContext)
		configOverrides = &clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{Server: ""}, CurrentContext: kubeContext,
		}
	} else {
		configOverrides = &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		configOverrides).ClientConfig()

	if err != nil {
		return nil, err
	}

	return config, nil
}

// checkPort checks if a local port is free.
func checkPort(port int, log logger) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// We do not care when closing fails because we do not expect incoming traffic
	err = l.Close()
	if err != nil {
		log.Warn(err.Error())
	}

	return nil
}

// getPod returns a pod for a pod name or look up a pod that belongs to a service
func getPod(config *rest.Config, namespace, podOrService string, log logger) (*v1.Pod, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Check for pods
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podOrService, metav1.GetOptions{})
	if err != nil {
		// In case the pod was not found, we want to check next if a service with this name exists.
		if !strings.Contains(err.Error(), "not found") {
			return nil, err
		}
	}

	if pod != nil && pod.Name != "" && pod.Namespace != "" {
		pod, err = waitTillReady(clientset, pod, log)
		if err != nil {
			return nil, err
		}
		return pod, nil
	}

	// Check for services
	service, err := clientset.CoreV1().Services(namespace).Get(context.Background(), podOrService, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	pod, err = searchPodForSvc(clientset, service)
	if err != nil {
		return nil, err
	}

	if pod == nil || pod.Name == "" || pod.Namespace == "" {
		msg := fmt.Sprintf("no pod or service with name %s found", podOrService)
		return nil, errors.New(msg)
	}

	return waitTillReady(clientset, pod, log)
}

func waitTillReady(clientset *kubernetes.Clientset, pod *v1.Pod, log logger) (*v1.Pod, error) {
	var err error

	ns := pod.Namespace
	n := pod.Name

	for i := 0; i < 6; i++ {
		status, ok := isPodReady(pod)

		if ok {
			return pod, nil
		}

		msg := fmt.Sprintf("Wait 10 seconds for pod %s to become ready (current status %s)", pod.Name, status)
		log.Info(msg)
		time.Sleep(10 * time.Second)

		pod, err = clientset.CoreV1().Pods(ns).Get(context.Background(), n, metav1.GetOptions{})

		if err != nil {
			return nil, err
		}
	}

	return nil, errors.New("pod did not become ready in time")
}

func searchPodForSvc(clientset *kubernetes.Clientset, service *v1.Service) (*v1.Pod, error) {
	selector := labels.SelectorFromSet(service.Spec.Selector)

	options := metav1.ListOptions{LabelSelector: selector.String()}

	podList, err := clientset.CoreV1().Pods(service.Namespace).List(context.Background(), options)
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		if _, ok := isPodReady(&pod); ok {
			return &pod, nil
		}
	}

	msg := fmt.Sprintf("no pod available for service %s", service.Name)
	return nil, errors.New(msg)
}

// newDialer creates a dialer that connects to the pod.
func newDialer(config *rest.Config, namespace string, pod *v1.Pod) (httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, pod.Name)
	hostIP := strings.TrimLeft(config.Host, "https://")

	// When there is a "/" in the hostIP, it contains also a path
	if parts := strings.SplitN(hostIP, "/", 2); len(parts) == 2 {
		hostIP = parts[0]
		path = fmt.Sprintf("/%s%s", parts[1], path)
	}

	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	return dialer, nil
}

// startForward runs the port-forwarding.
func startForward(dialer httpstream.Dialer, ports string, stopChan, readyChan chan struct{}, log logger) error {
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	forwarder, err := portforward.New(dialer, []string{ports}, stopChan, readyChan, out, errOut)
	if err != nil {
		return err
	}

	go func() {
		// Kubernetes will close this channel when it has something to tell us.
		for range readyChan {
		}
		if len(errOut.String()) != 0 {
			panic(errOut.String())
		} else if len(out.String()) != 0 {
			log.Debug(out.String())
		}
	}()

	errCh := make(chan error, 1)

	// Locks until stopChan is closed.
	go func() {
		if err = forwarder.ForwardPorts(); err != nil {
			log.Error(err.Error())
			errCh <- err
		}

		close(errCh)
	}()

	// For limited time we will wait if portforward returned an error
	timer := time.After(100 * time.Millisecond)

	select {
	case e := <-errCh:
		err = e
	case <-timer:
		msg := "Error listener timeout - nothing happend"
		log.Debug(msg)
		err = nil
	}

	return err
}

// closeOnSigterm cares about closing a channel when the OS sends a SIGTERM.
func closeOnSigterm(namespace, qualifiedName string, toPort int) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// Received kill signal
		<-sigs

		StopForwarding(namespace, qualifiedName, toPort)
	}()
}

func overwriteLog(log logger) {
	if !log.isOff() {
		return
	}

	debugPortforward("Turned off k8s runtime error handlers")
	runtime.ErrorHandlers = make([]func(error), 0)
}

func isPodReady(pod *v1.Pod) (string, bool) {
	podReady := pod.Status.Phase == v1.PodRunning

	if !podReady {
		return string(pod.Status.Phase), podReady
	}

	// Check if at least one container is ready:
	// We have to guess that this must be required one.
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			return "", status.Ready
		}
	}

	return "NoContainerReady", false
}

// ===== logger =====

const (
	Debug = iota
	Info
	Warn
	Error
	Off
)

type logger struct {
	level int
}

func newLogger(level int) logger {
	debugPortforward(fmt.Sprintf("level=%d", level))
	return logger{level: level}
}

func (l *logger) Debug(msg string) {
	if l.level > Debug {
		return
	}

	fmt.Printf("DEBUG: %s\n", msg)
}

func (l *logger) Info(msg string) {
	if l.level > Info {
		return
	}

	fmt.Printf("INFO: %s\n", msg)
}

func (l *logger) Warn(msg string) {
	if l.level > Warn {
		return
	}

	fmt.Printf("WARN: %s\n", msg)
}

func (l *logger) Error(msg string) {
	if l.level > Error {
		return
	}

	fmt.Printf("ERROR: %s\n", msg)
}

func (l *logger) isOff() bool {
	return l.level == Off
}

func (l *logger) logError(err error) {
	l.Error(err.Error())
}

func debugPortforward(msg string) {
	if os.Getenv("PORTFORWARD_DEBUG") == "YES" {
		fmt.Println(msg)
	}
}
