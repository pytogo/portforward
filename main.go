package main

// #include <Python.h>
// int PyArg_ParseTuple_ssll(PyObject*, char**, char**, int*, int*);
import "C"
import (
	"fmt"
	"github.com/pytogo/portforward/portforward"
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

	if err := portforward.PortForward(ns, pod, int(fromPort), int(toPort)); err != nil {
		panic(err)
	}

	C.Py_IncRef(C.Py_None)
	return C.Py_None
}

func main() {}
