"""
Kubernetes Port-Forward Go-Edition For Python
"""

__version__ = "0.4.0"

import contextlib
import os
import time
from enum import Enum
from pathlib import Path
from typing import Generator, Optional

import _portforward


class PortforwardError(Exception):
    """Will be raised when something went wrong while the port-forward process."""


class LogLevel(Enum):
    DEBUG = 0
    INFO = 1
    WARN = 2
    ERROR = 3
    OFF = 4


@contextlib.contextmanager
def forward(
    namespace: str,
    pod_or_service: str,
    from_port: int,
    to_port: int,
    config_path: Optional[str] = None,
    waiting: float = 0.1,
    log_level: LogLevel = LogLevel.DEBUG,
    kube_context: str = "",
) -> Generator[None, None, None]:
    """
    Connects to a **pod or service** and tunnels traffic from a local port to this target.
    It uses the kubectl kube config from the home dir if no path is provided.

    The libary will figure out for you if it has to target a pod or service.

    Caution: Go and the port-forwarding needs some ms to be ready. ``waiting``
    can be used to wait until the port-forward is ready.

    (Best consumed as context manager.)

    Example:
        >>> import portforward
        >>> with portforward.forward("test", "web-svc", 9000, 80):
        >>>     # Do work

    :param namespace: Target namespace
    :param pod_or_service: Name of target Pod or service
    :param from_port: Local port
    :param to_port: Port inside the pod
    :param config_path: Path for loading kube config
    :param waiting: Delay in seconds
    :param log_level: Level of logging
    :param kube_context: Target kubernetes context (fallback is current context)
    :return: None
    """

    _validate_str("namespace", namespace)
    _validate_str("pod_or_service", pod_or_service)

    _validate_port("from_port", from_port)
    _validate_port("to_port", to_port)

    _validate_log(log_level)

    config_path = _config_path(config_path)

    kube_context = kube_context if kube_context else ""
    _kube_context(kube_context)

    try:
        _portforward.forward(
            namespace,
            pod_or_service,
            from_port,
            to_port,
            config_path,
            log_level.value,
            kube_context,
        )

        # Go and the port-forwarding needs some ms to be ready
        time.sleep(waiting)

        yield None

    except RuntimeError as err:
        # Suppress extension exception
        raise PortforwardError(err) from None

    finally:
        _portforward.stop(namespace, pod_or_service)


# ===== PRIVATE =====


def _validate_str(arg_name, arg):
    if arg is None or not isinstance(arg, str):
        raise ValueError(f"{arg_name}={arg} is not a valid str")

    if len(arg) == 0:
        raise ValueError(f"{arg_name} cannot be an empty str")

    if "/" in arg:
        raise ValueError(f"{arg_name} contains illegal character '/'")


def _validate_port(arg_name, arg):
    in_range = arg and 0 < arg < 65536
    if arg is None or not isinstance(arg, int) or not in_range:
        raise ValueError(f"{arg_name}={arg} is not a valid port")


def _validate_log(log_level):
    if not isinstance(log_level, LogLevel):
        raise ValueError(f"log_level={log_level} is not a valid LogLevel")


def _config_path(config_path_arg) -> str:
    if config_path_arg and not isinstance(config_path_arg, str):
        raise ValueError(f"config_path={config_path_arg} is not a valid str")

    elif config_path_arg:
        return config_path_arg

    alt_path = str(Path.home() / ".kube" / "config")

    return os.environ.get("KUBECONFIG", alt_path)


def _kube_context(arg):
    if arg is None or not isinstance(arg, str):
        raise ValueError(f"kube_context={arg} is not a valid str")

    if "/" in arg:
        raise ValueError(f"kube_context contains illegal character '/'")
