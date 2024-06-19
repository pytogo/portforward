"""
Rust native module / Python C Extension
"""

async def forward(
    namespace: str,
    pod_or_service: str,
    bind_address: str,
    to_port: int,
    config_path: str,
    log_level: int,
    kube_context: str,
) -> None:
    pass

async def stop(namespace: str, actual_pod: str, to_port: int, log_level: int) -> None:
    pass
