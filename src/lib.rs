use pyo3::prelude::*;

mod portforward;

/// Creates a connection to a pod.
#[pyfunction]
fn forward(
    namespace: String,
    pod_or_service: String,
    from_port: u64,
    to_port: u64,
    config_path: String,
    log_level: u64,
    kube_context: String,
) -> PyResult<()> {
    let _config = portforward::ForwardConfig::builder()
        .namespace(namespace)
        .pod_or_service(pod_or_service)
        .from_port(from_port)
        .to_port(to_port)
        .config_path(config_path)
        .log_level(log_level)
        .kube_context(kube_context)
        .build();

    Ok(())
}

/// Stops a connection to a pod.
#[pyfunction]
fn stop(namespace: String, pod_or_service: String) -> PyResult<()> {
    Ok(())
}

/// A Python module implemented in Rust.
#[pymodule]
fn _portforward(_py: Python, m: &PyModule) -> PyResult<()> {
    m.add_function(wrap_pyfunction!(forward, m)?)?;
    m.add_function(wrap_pyfunction!(stop, m)?)?;
    Ok(())
}
