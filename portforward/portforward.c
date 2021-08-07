#include <Python.h>

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
    {"forward", (PyCFunction)forward, METH_VARARGS},
    {NULL, NULL}
};

static struct PyModuleDef module = {
    PyModuleDef_HEAD_INIT,
    "portforward",
    NULL,
    -1,
    methods
};

PyMODINIT_FUNC PyInit_portforward(void) {
    return PyModule_Create(&module);
}
