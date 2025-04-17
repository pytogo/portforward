use pyo3::exceptions::PyRuntimeError;
use pyo3::prelude::*;
use pyo3::types::PyModule;
use pyo3::Bound;

mod portforward;

/// Creates a connection to a pod.
#[pyfunction]
fn forward<'py>(
    py: Python<'py>,
    namespace: String,
    pod_or_service: String,
    bind_address: String,
    to_port: u16,
    config_path: String,
    log_level: u64,
    kube_context: String,
) -> PyResult<Bound<'py, PyAny>> {
    init_log(log_level);

    let config = portforward::ForwardConfig::builder()
        .namespace(namespace)
        .pod_or_service(pod_or_service)
        .bind_address(bind_address)
        .to_port(to_port)
        .config_path(config_path)
        .kube_context(kube_context)
        .build();

    pyo3_async_runtimes::tokio::future_into_py(py, async {
        portforward::forward(config).await.map_err(|e| {
            let msg = format!("{:?}", e);
            PyRuntimeError::new_err(msg)
        })
    })
}

/// Stops a connection to a pod.
#[pyfunction]
fn stop<'py>(
    py: Python<'py>,
    namespace: String,
    actual_pod: String,
    to_port: u16,
    log_level: u64,
) -> PyResult<Bound<'py, PyAny>> {
    init_log(log_level);

    pyo3_async_runtimes::tokio::future_into_py(py, async move {
        portforward::stop(namespace, actual_pod, to_port).await;
        Ok(()) // () automatically converts to Python None
    })
}

/// A Python module implemented in Rust.
#[pymodule]
fn _portforward<'py>(_py: Python<'py>, m: &Bound<'py, PyModule>) -> PyResult<()> {
    m.add_function(wrap_pyfunction!(forward, m)?)?;
    m.add_function(wrap_pyfunction!(stop, m)?)?;
    Ok(())
}

/*
   DEBUG = 0
   INFO = 1
   WARN = 2
   ERROR = 3
   OFF = 4
*/
// ===== ===== HELPER ===== =====
fn init_log(log_level: u64) {
    let level = match log_level {
        0 => log::LevelFilter::Debug,
        1 => log::LevelFilter::Info,
        2 => log::LevelFilter::Warn,
        3 => log::LevelFilter::Error,
        _ => log::LevelFilter::Off,
    };

    let _ = env_logger::builder().filter_level(level).try_init();
}
