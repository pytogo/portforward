package main

// #include <Python.h>
// int PyArg_ParseTuple_ssll(PyObject*, char**, char**, int*, int*);
import "C"
import (
	"bytes"
	"fmt"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/homedir"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

//export forward
func forward(self *C.PyObject, args *C.PyObject) *C.PyObject {
	// Interface for C extension and only part that contains C.
	var namespace *C.char
	var podName *C.char

	var fromPort C.int
	var toPort C.int

	if C.PyArg_ParseTuple_ssll(args, &namespace, &podName, &fromPort, &toPort) == 0 {
		fmt.Println("Could not parse args")

		C.Py_IncRef(C.Py_None)
		return C.Py_None
	}

	var ns string = C.GoString(namespace)
	var pod string = C.GoString(podName)

	if err := portForward(ns, pod, int(fromPort), int(toPort)); err != nil {
		panic(err)
	}

	C.Py_IncRef(C.Py_None)
	return C.Py_None
}

// portForward connects to a Pod and tunnels traffic from a local port to this pod.
func portForward(namespace, podName string, fromPort, toPort int) error {

	fmt.Printf("PortForward %s/%s (%d:%d)\n", namespace, podName, fromPort, toPort)

	// CONFIG
	var config *rest.Config

	if c, err := configFromHome(); err != nil {
		return err
	} else {
		config = c
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

	return nil
}

// Based on example https://github.com/kubernetes/client-go/issues/51#issuecomment-436200428

// configFromHome fetches the config from .kube config folder inside the home dir.
func configFromHome() (*rest.Config, error) {
	var configPath string
	if home := homedir.HomeDir(); home != "" {
		configPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
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
	if err = forwarder.ForwardPorts(); err != nil {
		return err
	}

	return nil
}

func main() {}
