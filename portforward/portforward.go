package main

import "C"
import "fmt"

//export PortForward
// PortForward creates a connection to a pod in a namespace.
func PortForward(namespace *C.char, podName *C.char, fromPort, toPort int) {
	var ns string = C.GoString(namespace)
	var pod string = C.GoString(podName)

	fmt.Printf("PortForward %s/%s (%d:%d)\n", ns, pod, fromPort, toPort)
}

func main() {}
