"""
portforward gives the super-power of port-forward from Go's Kubernetes library.
"""

import os
import ctypes

libname = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "portforward/portforward.so")
)

libc = ctypes.CDLL(libname)
libc.PortForward.argtypes = [
    ctypes.c_char_p,
    ctypes.c_char_p,
    ctypes.c_int,
    ctypes.c_int,
]


def port_forward(namespace: str, pod_name: str, from_port: int, to_port: int):
    """
    Creates a port_forward to a pod in the given namespace.

    :param namespace:
    :param pod_name:
    :param from_port:
    :param to_port:
    :return:
    """

    _validate_str("namespace", namespace)
    _validate_str("pod_name", pod_name)
    _validate_port("from_port", from_port)
    _validate_port("to_port", to_port)

    ns = namespace.encode("utf-8")
    pod = pod_name.encode("utf-8")

    libc.PortForward(ns, pod, from_port, to_port)


def _validate_str(arg_name, arg):
    if arg is None or not isinstance(arg, str):
        raise ValueError(f"{arg_name}={arg} is not a valid str")
    if len(arg) == 0:
        raise ValueError(f"{arg_name} cannot be an empty str")


def _validate_port(arg_name, arg):
    in_range = 80 < arg < 65536
    if arg is None or not isinstance(arg, int) or not in_range:
        raise ValueError(f"{arg_name}={arg} is not a valid port")
