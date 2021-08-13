package internal_portforward

import (
	"bytes"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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
	activeForwards = make(map[string]chan struct{})
	mutex          sync.Mutex
)

// registerForwarding adds a forwarding to the active forwards.
func registerForwarding(namespace, pod string, stopCh chan struct{}) {
	key := fmt.Sprintf("%s/%s", namespace, pod)

	mutex.Lock()
	defer mutex.Unlock()

	if otherCh, ok := activeForwards[key]; ok {
		close(otherCh)
	}

	activeForwards[key] = stopCh
}

// StopForwarding closes a port forwarding.
func StopForwarding(namespace, pod string) {
	key := fmt.Sprintf("%s/%s", namespace, pod)

	mutex.Lock()
	defer mutex.Unlock()

	if otherCh, ok := activeForwards[key]; ok {
		close(otherCh)
		delete(activeForwards, key)
	}
}

// ===== Port forwarding =====

// Forward connects to a Pod and tunnels traffic from a local port to this pod.
func Forward(namespace, podName string, fromPort, toPort int, configPath string) error {
	// Based on example https://github.com/kubernetes/client-go/issues/51#issuecomment-436200428

	// CONFIG
	var config *rest.Config

	if c, err := loadConfig(configPath); err != nil {
		return err
	} else {
		config = c
	}

	// CHECK
	// PortForward must be started in a go-routine, therefore we have
	// to check manually if the pod exists and is reachable.
	if err := checkPodExistence(config, namespace, podName); err != nil {
		return err
	}

	// DIALER
	var dialer httpstream.Dialer

	if d, err := newDialer(config, namespace, podName); err != nil {
		return err
	} else {
		dialer = d
	}

	// PORT FORWARD
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)

	ports := fmt.Sprintf("%d:%d", fromPort, toPort)

	if err := startForward(dialer, ports, stopChan, readyChan); err != nil {
		return err
	}

	// HANDLE CLOSING
	registerForwarding(namespace, podName, stopChan)
	closeOnSigterm(namespace, podName)

	return nil
}

// loadConfig fetches the config from .kube config folder inside the home dir.
func loadConfig(configPath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func checkPodExistence(config *rest.Config, namespace, podName string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	return nil
}

// newDialer creates a dialer that connects to the pod.
func newDialer(config *rest.Config, namespace, podName string) (httpstream.Dialer, error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	hostIP := strings.TrimLeft(config.Host, "https://")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	fmt.Printf("hostIP: %s\n", hostIP)

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	return dialer, nil
}

// startForward runs the port-forwarding.
func startForward(dialer httpstream.Dialer, ports string, stopChan, readyChan chan struct{}) error {
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
			fmt.Println(out.String())
		}
	}()

	// Locks until stopChan is closed.
	go func() {
		if err = forwarder.ForwardPorts(); err != nil {
			panic(err)
		}
	}()

	return nil
}

// closeOnSigterm cares about closing a channel when the OS sends a SIGTERM.
func closeOnSigterm(namespace, podName string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// Received kill signal
		<-sigs

		StopForwarding(namespace, podName)
	}()
}
