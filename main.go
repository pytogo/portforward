package main

// #include <Python.h>
// int PyArg_ParseTuple_ssiisi(PyObject* args, char** a, char** b, int* c, int* d, char** e, int* f);
// int PyArg_ParseTuple_ss(PyObject*, char**, char**);
// void raise_exception(char *msg);
import "C"
import (
	"github.com/pytogo/pytogo/portforward"
)

//export forward
func forward(self *C.PyObject, args *C.PyObject) *C.PyObject {
	// Interface for C extension and only part that contains C.

	// Strings should not need to be freed
	// > A pointer to an existing string is stored in the character pointer variable whose address you pass.
	// https://docs.python.org/3/c-api/arg.html
	var namespace *C.char
	var podName *C.char

	var fromPort C.int
	var toPort C.int

	var configPath *C.char
	var logLevel C.int

	if C.PyArg_ParseTuple_ssiisi(args, &namespace, &podName, &fromPort, &toPort, &configPath, &logLevel) == 0 {
		C.raise_exception(C.CString("Could not parse args"))
		return nil
	}

	var ns string = C.GoString(namespace)
	var pod string = C.GoString(podName)

	var cPath string = C.GoString(configPath)
	cLLevel := int(logLevel)

	if err := portforward.Forward(ns, pod, int(fromPort), int(toPort), cPath, cLLevel); err != nil {
		C.raise_exception(C.CString(err.Error()))
		return nil
	}

	C.Py_IncRef(C.Py_None)
	return C.Py_None
}

//export stop
func stop(self *C.PyObject, args *C.PyObject) *C.PyObject {
	// Interface for C extension and only part that contains C.

	// Strings should not need to be freed
	// > A pointer to an existing string is stored in the character pointer variable whose address you pass.
	// https://docs.python.org/3/c-api/arg.html
	var namespace *C.char
	var podName *C.char

	if C.PyArg_ParseTuple_ss(args, &namespace, &podName) == 0 {
		C.raise_exception(C.CString("Could not parse args"))
		return nil
	}

	var ns string = C.GoString(namespace)
	var pod string = C.GoString(podName)

	portforward.StopForwarding(ns, pod)

	C.Py_IncRef(C.Py_None)
	return C.Py_None
}

func main() {}
