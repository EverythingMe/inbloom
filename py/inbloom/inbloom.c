#define PY_SSIZE_T_CLEAN
#include <Python.h>
#include <arpa/inet.h>
#include "../vendor/libbloom/bloom.h"
#include "crc32.c"

static char module_docstring[] = "Python wrapper for libbloom";

typedef struct {
    PyObject_HEAD;
    struct bloom *_bloom_struct;
} Filter;

#if PY_MAJOR_VERSION >=3
#define MY_PyObject_HEAD_INIT(p, ob_size) PyVarObject_HEAD_INIT(p, ob_size)
#else
#define MY_PyObject_HEAD_INIT(p, ob_size) PyObject_HEAD_INIT(p) \
ob_size,
#endif

static PyTypeObject FilterType = {
    MY_PyObject_HEAD_INIT(NULL, 0)
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

#if PY_MAJOR_VERSION >=3
#define INSTANTIATE_ARGS_FORMAT "(idy#)"
#else
#define INSTANTIATE_ARGS_FORMAT "(ids#)"
#endif

static PyObject *
instantiate_filter(uint32_t cardinality, uint16_t error_rate, const char *data, int datalen)
{
    PyObject *args = Py_BuildValue(INSTANTIATE_ARGS_FORMAT, cardinality, 1.0 / error_rate, data, datalen);
    PyObject *obj = FilterType.tp_new(&FilterType, args, NULL);
    if (FilterType.tp_init(obj, args, NULL) < 0) {
        Py_DECREF(obj);
        obj = NULL;
    }
    Py_DECREF(args);
    return obj;
}

/* helpers */
static uint16_t
compute_checksum(const char *buf, size_t len)
{
    uint32_t checksum32 = crc32(0, buf, len);
    return (checksum32 & 0xFFFF) ^ (checksum32 >> 16);
}

static uint16_t
read_uint16(const char **buffer)
{
    uint16_t ret = ntohs(*((uint16_t *)*buffer));
    *buffer += sizeof(uint16_t);
    return ret;
}

static uint32_t
read_uint32(const char **buffer)
{
    uint32_t ret = ntohl(*((uint32_t *)*buffer));
    *buffer += sizeof(uint32_t);
    return ret;
}

#if PY_MAJOR_VERSION >=3
#define LOAD_ARGS_FORMAT "y#"
#else
#define LOAD_ARGS_FORMAT "s#"
#endif

/* serialization */
static PyObject *
load(PyObject *self, PyObject *args)
{
    const char *buffer;
    Py_ssize_t buflen;
    if (!PyArg_ParseTuple(args, LOAD_ARGS_FORMAT, &buffer, &buflen)) {
        return NULL;
    }

    if ((int)buflen < sizeof(struct serialized_filter_header) + 1) {
        PyErr_SetString(InBloomError, "incomplete payload");
        return NULL;
    }

    struct serialized_filter_header header;
    header.checksum = read_uint16(&buffer);
    header.error_rate = read_uint16(&buffer);
    header.cardinality = read_uint32(&buffer);
    const char *data = buffer;
    size_t datalen = (int)buflen - sizeof(struct serialized_filter_header);
    uint16_t expected_checksum = compute_checksum(data, datalen);
    if (expected_checksum != header.checksum) {
        PyErr_SetString(InBloomError, "checksum mismatch");
        return NULL;
    }
    return instantiate_filter(header.cardinality, header.error_rate, data, datalen);
}

static PyObject *
dump(PyObject *self, PyObject *args)
{
    Filter *filter;
    if (!PyArg_ParseTuple(args, "O", &filter)) {
        return NULL;
    }
    uint16_t checksum = compute_checksum((const char *)filter->_bloom_struct->bf, filter->_bloom_struct->bytes);

    struct serialized_filter_header header = {htons(checksum), htons(1.0 / filter->_bloom_struct->error), htonl(filter->_bloom_struct->entries)};
    PyObject *serial_header = PyBytes_FromStringAndSize((const char *)&header, sizeof(struct serialized_filter_header));
    PyObject *serial_data = PyBytes_FromStringAndSize((const char *)filter->_bloom_struct->bf, filter->_bloom_struct->bytes);
    PyBytes_Concat(&serial_header, serial_data);
    Py_DECREF(serial_data);
    return serial_header;
}

static PyMethodDef module_methods[] = {
    {"load", (PyCFunction)load, METH_VARARGS,
     "load a serialized filter"},
    {"dump", (PyCFunction)dump, METH_VARARGS,
     "dump a filter into a string"},
    {NULL}
};

/* Filter methods */
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
    return PyBytes_FromStringAndSize((const char *)self->_bloom_struct->bf, self->_bloom_struct->bytes);
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

#if PY_MAJOR_VERSION >=3
#define MY_Py_TYPE(t) Py_TYPE(t)
#else
#define MY_Py_TYPE(t) t->ob_type
#endif

static void
Filter_dealloc(Filter* self)
{
    bloom_free(self->_bloom_struct);
    free(self->_bloom_struct);
    MY_Py_TYPE(self)->tp_free((PyObject*)self);
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

#if PY_MAJOR_VERSION >=3
#define INIT_ARGS_FORMAT "id|y#"
#else
#define INIT_ARGS_FORMAT "id|s#"
#endif

static int
Filter_init(Filter *self, PyObject *args, PyObject *kwargs)
{
    static char *kwlist[] = {"entries", "error", "data", NULL};
    int entries, success;
    double error;
    const char *data = NULL;
    Py_ssize_t len;
    if (!PyArg_ParseTupleAndKeywords(args, kwargs, INIT_ARGS_FORMAT, kwlist, &entries, &error, &data, &len)) {
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

#if PY_MAJOR_VERSION >=3
#define INIT_ENTRY PyInit_inbloom
#define RESULT(arg) arg
#else
#define INIT_ENTRY initinbloom
#define RESULT(arg)
#endif

PyMODINIT_FUNC
INIT_ENTRY(void)
{
    PyObject *m;
    FilterType.tp_new = Filter_new;
    FilterType.tp_init = (initproc)Filter_init;
    FilterType.tp_methods = Filter_methods;
    FilterType.tp_dealloc = (destructor)Filter_dealloc;
    if (PyType_Ready(&FilterType) < 0)
        return RESULT(NULL);

#if PY_MAJOR_VERSION >=3
	static struct PyModuleDef moduledef = {
    PyModuleDef_HEAD_INIT,
        "inbloom",
        module_docstring,
        -1,
        module_methods
    };
    m = PyModule_Create(&moduledef);
#else
    m = Py_InitModule3("inbloom", module_methods, module_docstring);
#endif
    Py_INCREF(&FilterType);
    PyModule_AddObject(m, "Filter", (PyObject *)&FilterType);

    InBloomError = PyErr_NewException("inbloom.error", NULL, NULL);
    Py_INCREF(InBloomError);
    PyModule_AddObject(m, "error", InBloomError);
    return RESULT(m);
}
