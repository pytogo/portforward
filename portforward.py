"""
portforward gives the super-power of port-forward from Go's Kubernetes library.
"""

import os
import ctypes

libname = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "portforward/portforward.so")
)

libc = ctypes.CDLL(libname)
libc.PortForward.argtypes = [ctypes.c_char_p, ctypes.c_char_p]


def port_forward(namespace: str, pod_name: str):
    """
    Creates a port_forward to a pod in the given namespace.

    :param namespace:
    :param pod_name:
    :return:
    """

    _validate("namespace", namespace)
    _validate("pod_name", pod_name)

    ns = namespace.encode("utf-8")
    pod = pod_name.encode("utf-8")

    libc.PortForward(ns, pod)


def _validate(arg_name, arg):
    if arg is None or not isinstance(arg, str):
        raise ValueError(f"{arg_name}={arg} is not a valid str")
    if len(arg) == 0:
        raise ValueError(f"{arg_name} cannot be an empty str")
