//! portforwards contains the "nice" Rust API.
//! It is based on [kube-rs examples](https://github.com/kube-rs/kube/blob/33ba90e0a003c915801bb9b6c4961b7d0b721889/examples/pod_portforward_bind.rs).

use anyhow::Context;
use futures::{StreamExt, TryStreamExt};
use k8s_openapi::api::core::v1::Pod;
use kube::{
    api::Api,
    runtime::wait::{await_condition, conditions::is_pod_running},
    Client,
};
use once_cell::sync::Lazy;
use std::collections::HashMap;
use std::net::SocketAddr;
use tokio::{
    io::{AsyncRead, AsyncWrite},
    net::TcpListener,
    sync::RwLock,
};
use tokio_stream::wrappers::TcpListenerStream;
use tracing::*;
use typed_builder::TypedBuilder;

static OPEN_PORTFORWARDS: Lazy<RwLock<HashMap<String, String>>> = Lazy::new(|| {
    let m = HashMap::new();
    RwLock::new(m)
});

#[derive(TypedBuilder)]
pub struct ForwardConfig {
    namespace: String,
    pod_or_service: String,
    from_port: u64,
    to_port: u64,
    config_path: String,
    log_level: u64,
    kube_context: String,
}

/// Creates a connection to a pod.
pub async fn forward(config: ForwardConfig) {
    tracing_subscriber::fmt::init();
    let client = Client::try_default().await.unwrap();

    let pods: Api<Pod> = Api::default_namespaced(client);

    // Wait until the pod is running, otherwise we get 500 error.
    info!("waiting for nginx pod to start");
    let running = await_condition(pods.clone(), "nginx", is_pod_running());
    let _ = tokio::time::timeout(std::time::Duration::from_secs(30), running)
        .await
        .unwrap();

    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    let pod_port = 80;
    info!(local_addr = %addr, pod_port, "forwarding traffic to the pod");
    info!(
        "try opening http://{0} in a browser, or `curl http://{0}`",
        addr
    );
    info!("use Ctrl-C to stop the server and delete the pod");
    let server = TcpListenerStream::new(TcpListener::bind(addr).await.unwrap())
        .take_until(tokio::signal::ctrl_c())
        .try_for_each(|client_conn| async {
            if let Ok(peer_addr) = client_conn.peer_addr() {
                info!(%peer_addr, "new connection");
            }
            let pods = pods.clone();
            tokio::spawn(async move {
                if let Err(e) = forward_connection(&pods, "nginx", 80, client_conn).await {
                    error!(
                        error = e.as_ref() as &dyn std::error::Error,
                        "failed to forward connection"
                    );
                }
            });
            // keep the server running
            Ok(())
        });
    if let Err(e) = server.await {
        error!(error = &e as &dyn std::error::Error, "server error");
    }
}

/// Stops a connection to a pod.
pub async fn stop(namespace: String, pod_or_service: String) {}

async fn forward_connection(
    pods: &Api<Pod>,
    pod_name: &str,
    port: u16,
    mut client_conn: impl AsyncRead + AsyncWrite + Unpin,
) -> anyhow::Result<()> {
    let mut forwarder = pods.portforward(pod_name, &[port]).await?;
    let mut upstream_conn = forwarder
        .take_stream(port)
        .context("port not found in forwarder")?;
    tokio::io::copy_bidirectional(&mut client_conn, &mut upstream_conn).await?;
    drop(upstream_conn);
    forwarder.join().await?;
    info!("connection closed");
    Ok(())
}

// ===== ===== HELPER ===== =====

#[derive(Debug)]
enum TargetType {
    Service,
    Pod,
}

struct QualifiedName {
    namespace: String,
    target_type: TargetType,
    name: String,
}

impl QualifiedName {
    fn new(namespace: String, target_type: TargetType, name: String) -> Self {
        Self {
            namespace,
            target_type,
            name,
        }
    }

    fn to_string(&self) -> String {
        format!("{}/{}", self.namespace, self.name)
    }
}

async fn register_forward(qualified_name: QualifiedName, conn: String) {
    OPEN_PORTFORWARDS
        .write()
        .await
        .insert(qualified_name.to_string(), conn);
}
