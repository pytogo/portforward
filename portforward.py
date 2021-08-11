"""
Kubernetes Port-Forward Go-Edition For Python
"""

import _portforward


class PortforwardError(Exception):
    """ Will be raised when something went wrong while the port-forward process. """


def forward(namespace: str, pod: str, from_port: int, to_port: int) -> None:
    """
    Connects to a Pod and tunnels traffic from a local port to this pod.
    It uses the kubectl kube config from the home dir. The portforward will
    be closed by SIGTERM.

    Example:
        >>> import portforward
        >>> portforward.forward("test", "web", 9000, 80)

    :param namespace: Target namespace
    :param pod: Name of target Pod
    :param from_port: Local port
    :param to_port: Port inside the pod
    :return: None
    """

    _validate_str("namespace", namespace)
    _validate_str("pod", pod)

    _validate_port("from_port", from_port)
    _validate_port("to_port", to_port)

    try:
        _portforward.forward(namespace, pod, from_port, to_port)
    except RuntimeError as err:
        # Suppress extension exception
        raise PortforwardError(err) from None


# ===== PRIVATE =====


def _validate_str(arg_name, arg):
    if arg is None or not isinstance(arg, str):
        raise ValueError(f"{arg_name}={arg} is not a valid str")
    if len(arg) == 0:
        raise ValueError(f"{arg_name} cannot be an empty str")


def _validate_port(arg_name, arg):
    in_range = 0 < arg < 65536
    if arg is None or not isinstance(arg, int) or not in_range:
        raise ValueError(f"{arg_name}={arg} is not a valid port")
