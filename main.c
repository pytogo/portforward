#include <Python.h>

// ===== START PYTHON PART =====

/* Will come from go */
PyObject* forward(PyObject* , PyObject*);

/*
To shim go's missing variadic function support.

Ref https://docs.python.org/3/c-api/arg.html
*/
int PyArg_ParseTuple_ssll(PyObject* args, char** a, char** b, int* c, int* d) {
    return PyArg_ParseTuple(args, "ssii", a, b, c, d);
}

static struct PyMethodDef methods[] = {
    {
    "forward_by_home",
    (PyCFunction)forward,
    METH_VARARGS,
    "Connects to a Pod and tunnels traffic from a local port to this pod. It uses the kubectl kube config from the home dir."
    },
    {NULL, NULL, 0, NULL}
};

static struct PyModuleDef module = {
    PyModuleDef_HEAD_INIT,
    "_portforward",
    "Kubernetes Port-Forward Go-Edition For Python",
    -1,
    methods
};

static PyObject *PortforwardError = NULL;

PyMODINIT_FUNC PyInit__portforward(void) {
    /* Assign module value */
    PyObject *module = PyModule_Create(&module);

    /* Initialize new exception object */
    StringTooShortError = PyErr_NewException("_portforward.PortforwardError", PyExc_RuntimeError, NULL);

    /* Add exception object to your module */
    PyModule_AddObject(module, "PortforwardError", PortforwardError);

    return module;
}

// ===== END PYTHON PART =====

void raise_exception(char *msg) {
    PyErr_SetString(PortforwardError, msg)
}
