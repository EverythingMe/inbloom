#include <Python.h>
#include "../vendor/libbloom/bloom.h"
#include "crc32.c"

static char module_docstring[] = "Python wrapper for libbloom";

typedef struct {
    PyObject_HEAD;
    struct bloom *_bloom_struct;
} Filter;

static PyTypeObject FilterType = {
    PyObject_HEAD_INIT(NULL)
    0,                          /*ob_size*/
    "inbloom.Filter",           /*tp_name*/
    sizeof(Filter),             /*tp_basicsize*/
    0,                          /*tp_itemsize*/
    0,                          /*tp_dealloc*/
    0,                          /*tp_print*/
    0,                          /*tp_getattr*/
    0,                          /*tp_setattr*/
    0,                          /*tp_compare*/
    0,                          /*tp_repr*/
    0,                          /*tp_as_number*/
    0,                          /*tp_as_sequence*/
    0,                          /*tp_as_mapping*/
    0,                          /*tp_hash */
    0,                          /*tp_call*/
    0,                          /*tp_str*/
    0,                          /*tp_getattro*/
    0,                          /*tp_setattro*/
    0,                          /*tp_as_buffer*/
    Py_TPFLAGS_DEFAULT,         /*tp_flags*/
    "Filter objects",           /*tp_doc*/
};

struct serialized_filter_header {
    uint16_t checksum;
    uint16_t error_rate;
    uint32_t cardinality;
};

static PyObject *InBloomError;

static PyObject *
instantiate_filter(uint32_t cardinality, uint16_t error_rate, const char *data, int datalen)
{
    PyObject *args = Py_BuildValue("(ids#)", cardinality, 1.0 / error_rate, data, datalen);
    PyObject *obj = FilterType.tp_new(&FilterType, args, NULL);
    if (FilterType.tp_init(obj, args, NULL) < 0) {
        Py_DECREF(obj);
        obj = NULL;
    }
    return obj;
}

static uint16_t compute_checksum(const char *buf, size_t len)
{
    uint32_t checksum32 = crc32(0, buf, len);
    return (checksum32 & 0xFFFF) ^ (checksum32 >> 16);
}

static PyObject *
load(PyObject *self, PyObject *args)
{
    const char *buffer;
    Py_ssize_t buflen;
    if (!PyArg_ParseTuple(args, "s#", &buffer, &buflen)) {
        return NULL;
    }

    const struct serialized_filter_header *header = (const struct serialized_filter_header *)buffer;
    const char *data = buffer + sizeof(struct serialized_filter_header);
    size_t datalen = (int)buflen - sizeof(struct serialized_filter_header);
    uint16_t expected_checksum = compute_checksum(data, datalen);
    if (expected_checksum != header->checksum) {
        PyErr_SetString(InBloomError, "checksum mismatch");
        return NULL;
    }
    return instantiate_filter(header->cardinality, header->error_rate, data, datalen);
}

static PyObject *
dump(PyObject *self, PyObject *args)
{
    Filter *filter;
    if (!PyArg_ParseTuple(args, "O", &filter)) {
        return NULL;
    }
    uint16_t checksum = compute_checksum((const char *)filter->_bloom_struct->bf, filter->_bloom_struct->bytes);

    struct serialized_filter_header header = {checksum, 1.0 / filter->_bloom_struct->error, filter->_bloom_struct->entries};
    PyObject *serial_header = PyString_FromStringAndSize((const char *)&header, sizeof(struct serialized_filter_header));
    PyObject *serial_data = PyString_FromStringAndSize((const char *)filter->_bloom_struct->bf, filter->_bloom_struct->bytes);
    PyString_Concat(&serial_header, serial_data);
    return serial_header;
}

static PyMethodDef module_methods[] = {
    {"load", (PyCFunction)load, METH_VARARGS,
     "load a serialized filter"},
    {"dump", (PyCFunction)dump, METH_VARARGS,
     "dump a filter into a string"},
    {NULL}
};

static PyObject *
Filter_add(Filter *self, PyObject *args)
{
    const char *buffer;
    Py_ssize_t buflen;
    if (!PyArg_ParseTuple(args, "s#", &buffer, &buflen)) {
        return NULL;
    }

    bloom_add(self->_bloom_struct, buffer, buflen);
    Py_RETURN_NONE;
}

static PyObject *
Filter_check(Filter *self, PyObject *args)
{
    const char *buffer;
    Py_ssize_t buflen;
    if (!PyArg_ParseTuple(args, "s#", &buffer, &buflen)) {
        return NULL;
    }

    if (bloom_check(self->_bloom_struct, buffer, buflen))
        Py_RETURN_TRUE;
    else
        Py_RETURN_FALSE;
}

static PyObject *
Filter_buffer(Filter *self, PyObject *args)
{
    return PyString_FromStringAndSize((const char *)self->_bloom_struct->bf, self->_bloom_struct->bytes);
}

static PyMethodDef Filter_methods[] = {
    {"add", (PyCFunction)Filter_add, METH_VARARGS,
     "add a member to the filter"},
    {"contains", (PyCFunction)Filter_check, METH_VARARGS,
     "check if member exists the filter"},
    {"buffer", (PyCFunction)Filter_buffer, METH_NOARGS,
     "get a copy of the internal buffer"},
    {NULL}  /* Sentinel */
};

static void
Filter_dealloc(Filter* self)
{
    bloom_free(self->_bloom_struct);
    free(self->_bloom_struct);
    self->ob_type->tp_free((PyObject*)self);
}

static PyObject *
Filter_new(PyTypeObject *type, PyObject *args, PyObject *kwds)
{
    Filter *self;

    self = (Filter *)type->tp_alloc(type, 0);
    if (self != NULL) {
        self->_bloom_struct = (struct bloom *)malloc(sizeof(struct bloom));
        if (self->_bloom_struct == NULL)
            return PyErr_NoMemory();
    }

    return (PyObject *)self;
}

static int
Filter_init(Filter *self, PyObject *args, PyObject *kwargs)
{
    static char *kwlist[] = {"entries", "error", "data", NULL};
    int entries, success;
    double error;
    const char *data = NULL;
    Py_ssize_t len;
    if (!PyArg_ParseTupleAndKeywords(args, kwargs, "id|s#", kwlist, &entries, &error, &data, &len)) {
        return -1;
    }
    success = bloom_init(self->_bloom_struct, entries, error);
    if (success == 0) {
        if (data != NULL) {
            if ((int)len != self->_bloom_struct->bytes) {
                PyErr_SetString(InBloomError, "invalid data length");
                return -1;
            }
            memcpy(self->_bloom_struct->bf, (const unsigned char *)data, self->_bloom_struct->bytes);
        }
        return 0;
    }
    else {
        PyErr_SetString(InBloomError, "internal initialization failed");
        return -1;
    }
}


#ifndef PyMODINIT_FUNC
#define PyMODINIT_FUND void
#endif
PyMODINIT_FUNC
initinbloom(void)
{
    PyObject *m;
    FilterType.tp_new = Filter_new;
    FilterType.tp_init = (initproc)Filter_init;
    FilterType.tp_methods = Filter_methods;
    FilterType.tp_dealloc = (destructor)Filter_dealloc;
    if (PyType_Ready(&FilterType) < 0)
        return;

    m = Py_InitModule3("inbloom", module_methods, module_docstring);
    Py_INCREF(&FilterType);
    PyModule_AddObject(m, "Filter", (PyObject *)&FilterType);

    InBloomError = PyErr_NewException("inbloom.error", NULL, NULL);
    Py_INCREF(InBloomError);
    PyModule_AddObject(m, "error", InBloomError);
}
