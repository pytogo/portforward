"""
Go native module / Python C Extension
"""


def forward(namespace: str, pod_or_service: str, from_port: int, to_port: int, config_path: str, log_level: int, kube_context: str) -> None:
    pass


def stop(namespace: str, pod_or_service: str, to_port: int) -> None:
    pass
