"""
Easy Kubernetes Port-Forward For Python
"""

__version__ = "0.7.0"

import asyncio
import contextlib
import ipaddress
import os
from enum import Enum
from pathlib import Path
from typing import Generator, Optional, Union

from portforward import _portforward


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
    log_level: LogLevel = LogLevel.INFO,
    kube_context: str = "",
    bind_ip: Union[ipaddress.IPv4Address, ipaddress.IPv6Address, str, None] = None,
) -> Generator["PortForwarder", None, None]:
    """
    Connects to a **pod or service** and tunnels traffic from a local port to
    this target. It uses the kubectl kube config from the home dir if no path
    is provided.

    The libary will figure out for you if it has to target a pod or service.

    It will fall back to in-cluster-config in case no kube config file exists.

    (Best consumed as context manager.)

    Example:
        >>> import portforward
        >>> with portforward.forward("test", "web-svc", 9000, 80):
        >>>     # Do work

    :param namespace: Target namespace
    :param pod_or_service: Name of target Pod or service
    :param from_port: Local port, or 0 to use any free port
    :param to_port: Port inside the pod
    :param config_path: Path for loading kube config
    :param waiting: Delay in seconds
    :param log_level: Level of logging
    :param kube_context: Target kubernetes context (fallback is current context)
    :param bind_ip: To which IP shall the portforward be bind
    :return: forwarder to manual stop the forwarding
    """

    forwarder = PortForwarder(
        namespace,
        pod_or_service,
        from_port,
        to_port,
        config_path,
        waiting,
        log_level,
        kube_context,
        bind_ip,
    )

    try:
        forwarder.forward()

        yield forwarder

    except RuntimeError as err:
        # Suppress extension exception
        raise PortforwardError(err) from None

    finally:
        forwarder.stop()


class PortForwarder:
    """Use the same args as the `portforward.forward` method."""

    def __init__(
        self,
        namespace: str,
        pod_or_service: str,
        from_port: int,
        to_port: int,
        config_path: Optional[str] = None,
        waiting: float = 0.1,
        log_level: LogLevel = LogLevel.INFO,
        kube_context: str = "",
        bind_ip: Union[ipaddress.IPv4Address, ipaddress.IPv6Address, str, None] = None,
    ) -> None:
        self._async_forwarder = AsyncPortForwarder(
            namespace,
            pod_or_service,
            from_port,
            to_port,
            config_path,
            waiting,
            log_level,
            kube_context,
            bind_ip,
        )

    def forward(self):
        asyncio.run(self._async_forwarder.forward())

    def stop(self):
        asyncio.run(self._async_forwarder.stop())

    @property
    def is_stopped(self):
        return self._async_forwarder.is_stopped

    @property
    def from_port(self):
        """The local port that was actually used for the portforward."""
        return self._async_forwarder.from_port


class AsyncPortForwarder:
    """Use the same args as the `portforward.forward` method."""

    def __init__(
        self,
        namespace: str,
        pod_or_service: str,
        from_port: int,
        to_port: int,
        config_path: Optional[str] = None,
        waiting: float = 0.1,
        log_level: LogLevel = LogLevel.INFO,
        kube_context: str = "",
        bind_ip: Union[ipaddress.IPv4Address, ipaddress.IPv6Address, str, None] = None,
    ) -> None:
        self.namespace: str = _validate_str("namespace", namespace)
        self.pod_or_service: str = _validate_str("pod_or_service", pod_or_service)
        self.to_port: int = _validate_port("to_port", to_port)
        self.log_level: LogLevel = _validate_log(log_level)
        self.waiting: float = waiting

        self.config_path: str = _config_path(config_path)
        self.kube_context: str = _kube_context(kube_context)

        _validate_port("from_port", from_port)
        bind_ip = _validate_ip_address(bind_ip)

        self.actual_pod_name: str = ""
        self.from_port: int = 0
        self._is_stopped: bool = False
        self.bind_address: str = f"{bind_ip}:{from_port}"

    async def forward(self):
        (self.actual_pod_name, self.from_port) = await _portforward.forward(
            self.namespace,
            self.pod_or_service,
            self.bind_address,
            self.to_port,
            self.config_path,
            self.log_level.value,
            self.kube_context,
        )
        self._is_stopped = False

    async def stop(self):
        await _portforward.stop(
            self.namespace, self.actual_pod_name, self.to_port, self.log_level.value
        )
        self._is_stopped = True

    def is_stopped(self):
        return self._is_stopped


# ===== PRIVATE =====


def _validate_str(arg_name, arg) -> str:
    if arg is None or not isinstance(arg, str):
        raise ValueError(f"{arg_name}={arg} is not a valid str")

    if len(arg) == 0:
        raise ValueError(f"{arg_name} cannot be an empty str")

    if "/" in arg:
        raise ValueError(f"{arg_name} contains illegal character '/'")

    return arg


def _validate_port(arg_name, arg) -> int:
    in_range = arg is not None and 0 <= arg < 65536
    if arg is None or not isinstance(arg, int) or not in_range:
        raise ValueError(f"{arg_name}={arg} is not a valid port")

    return arg


def _validate_log(log_level):
    if not isinstance(log_level, LogLevel):
        raise ValueError(f"log_level={log_level} is not a valid LogLevel")

    return log_level


def _validate_ip_address(ip_address):
    if not ip_address:
        return "127.0.0.1"

    if isinstance(ip_address, ipaddress.IPv4Address) or isinstance(
        ip_address, ipaddress.IPv4Address
    ):
        return str(ip_address)

    return str(ipaddress.ip_address(ip_address))


def _config_path(config_path_arg) -> str:
    if config_path_arg and not isinstance(config_path_arg, str):
        raise ValueError(f"config_path={config_path_arg} is not a valid str")

    elif config_path_arg:
        return config_path_arg

    alt_path = str(Path.home() / ".kube" / "config")

    config_path = os.environ.get("KUBECONFIG", alt_path)

    return config_path if os.path.isfile(config_path) else ""


def _kube_context(context):
    if not context:
        return ""

    if not isinstance(context, str):
        raise ValueError(f"kube_context={context} is not a valid str")

    if "/" in context:
        raise ValueError("kube_context contains illegal character '/'")

    return context
