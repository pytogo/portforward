package main

// #include <Python.h>
// int PyArg_ParseTuple_ssll(PyObject*, char**, char**, int*, int*);
// int PyArg_ParseTuple_ss(PyObject*, char**, char**);
// void raise_exception(char *msg);
import "C"
import (
	"github.com/pytogo/portforward/internal_portforward"
)

//export forward
func forward(self *C.PyObject, args *C.PyObject) *C.PyObject {
	// Interface for C extension and only part that contains C.
	var namespace *C.char
	var podName *C.char

	var fromPort C.int
	var toPort C.int

	if C.PyArg_ParseTuple_ssll(args, &namespace, &podName, &fromPort, &toPort) == 0 {
		C.raise_exception(C.CString("Could not parse args"))
		return nil
	}

	var ns string = C.GoString(namespace)
	var pod string = C.GoString(podName)

	if err := internal_portforward.Forward(ns, pod, int(fromPort), int(toPort)); err != nil {
		C.raise_exception(C.CString(err.Error()))
		return nil
	}

	C.Py_IncRef(C.Py_None)
	return C.Py_None
}

//export stop
func stop(self *C.PyObject, args *C.PyObject) *C.PyObject {
	// Interface for C extension and only part that contains C.
	var namespace *C.char
	var podName *C.char

	if C.PyArg_ParseTuple_ss(args, &namespace, &podName) == 0 {
		C.raise_exception(C.CString("Could not parse args"))
		return nil
	}

	var ns string = C.GoString(namespace)
	var pod string = C.GoString(podName)

	internal_portforward.StopForwarding(ns, pod)

	C.Py_IncRef(C.Py_None)
	return C.Py_None
}

func main() {}
