#include <Python.h>

/* Will come from go */
PyObject* portforward(PyObject* , PyObject*);

/* To shim go's missing variadic function support */
int PyArg_ParseTuple_ll(PyObject* args, int* a, int* b) {
    return PyArg_ParseTuple(args, "II", a, b);
}

static struct PyMethodDef methods[] = {
    {"portforward", (PyCFunction)portforward, METH_VARARGS},
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
