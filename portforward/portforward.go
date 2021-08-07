package main

// #include <Python.h>
// int PyArg_ParseTuple_ll(PyObject*, int*, int*);
import "C"
import "fmt"

//export portforward
func portforward(self *C.PyObject, args *C.PyObject) {
	var fromPort C.int
	var toPort C.int

	if C.PyArg_ParseTuple_ll(args, &fromPort, &toPort) == 0 {
		fmt.Println("Could not parse args")
		return
	}

	fmt.Printf("Portforward from %d to %d", fromPort, toPort)
}

func main() {}
