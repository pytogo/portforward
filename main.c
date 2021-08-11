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
    PyObject *m;

    m = PyModule_Create(&module);
    if (m == NULL)
        return NULL;

    /* Initialize new exception object */
    PortforwardError = PyErr_NewException("_portforward.PortforwardError", PyExc_RuntimeError, NULL);
    Py_XINCREF(PortforwardError);

    /* Add exception object to your module */
    if (PyModule_AddObject(m, "error", PortforwardError) < 0) {
        Py_XDECREF(PortforwardError);
        Py_CLEAR(PortforwardError);
        Py_DECREF(m);
        return NULL;
    }

    return m;
}

// ===== END PYTHON PART =====

void raise_exception(char *msg) {
    PyErr_SetString(PortforwardError, msg);
}
