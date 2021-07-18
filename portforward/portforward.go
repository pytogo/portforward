package main

import "C"
import "fmt"

//export PortForward
func PortForward(namespace *C.char, podName *C.char) {
	var ns string = C.GoString(namespace)
	var pod string = C.GoString(podName)

	fmt.Printf("PortForward %s/%s\n", ns, pod)
}

func main() {}
