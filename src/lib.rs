use pyo3::prelude::*;

mod portforward;

/// Creates a connection to a pod.
#[pyfunction]
fn forward(
    py: Python<'_>,
    namespace: String,
    pod_or_service: String,
    from_port: u16,
    to_port: u16,
    config_path: String,
    log_level: u64,
    kube_context: String,
) -> PyResult<&PyAny> {
    let config = portforward::ForwardConfig::builder()
        .namespace(namespace)
        .pod_or_service(pod_or_service)
        .from_port(from_port)
        .to_port(to_port)
        .config_path(config_path)
        .log_level(log_level)
        .kube_context(kube_context)
        .build();

    pyo3_asyncio::tokio::future_into_py(py, async {
        portforward::forward(config).await;
        Ok(Python::with_gil(|py| py.None()))
    })
}

/// Stops a connection to a pod.
#[pyfunction]
fn stop(py: Python<'_>, namespace: String, pod_or_service: String) -> PyResult<&PyAny> {
    pyo3_asyncio::tokio::future_into_py(py, async {
        portforward::stop(namespace, pod_or_service).await;
        Ok(Python::with_gil(|py| py.None()))
    })
}

/// A Python module implemented in Rust.
#[pymodule]
fn _portforward(_py: Python, m: &PyModule) -> PyResult<()> {
    m.add_function(wrap_pyfunction!(forward, m)?)?;
    m.add_function(wrap_pyfunction!(stop, m)?)?;
    Ok(())
}
