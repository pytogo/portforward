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
    sync::oneshot::{self, Receiver},
    sync::{oneshot::Sender, RwLock},
};
use tokio_stream::wrappers::TcpListenerStream;
use tracing::*;
use typed_builder::TypedBuilder;

static OPEN_PORTFORWARDS: Lazy<RwLock<HashMap<String, Sender<()>>>> = Lazy::new(|| {
    let m = HashMap::new();
    RwLock::new(m)
});

#[derive(TypedBuilder)]
pub struct ForwardConfig {
    namespace: String,
    pod_or_service: String,
    from_port: u16,
    to_port: u16,
    config_path: String,
    log_level: u64,
    kube_context: String,
}

/// Creates a connection to a pod. It returns the actual pod name for the portforward.
/// It differs from `pod_or_service` when `pod_or_service` represents a service.
pub async fn forward(config: ForwardConfig) -> String {
    tracing_subscriber::fmt::init();

    let q_name = QualifiedName::new(config.namespace, config.pod_or_service);
    let target_pod = q_name.pod_name.clone();

    let kube_config = kube::config::Kubeconfig::read_from(config.config_path).unwrap();
    let mut options = kube::config::KubeConfigOptions::default();
    options.context = Some(config.kube_context);
    let client_config = kube::config::Config::from_custom_kubeconfig(kube_config, &options)
        .await
        .unwrap();
    let client = Client::try_from(client_config).unwrap();
    let pods: Api<Pod> = Api::namespaced(client, &q_name.namespace);

    let running = await_condition(pods.clone(), &q_name.pod_name, is_pod_running());
    let _ = tokio::time::timeout(std::time::Duration::from_secs(30), running)
        .await
        .unwrap();

    let (tx, rx): (Sender<()>, Receiver<()>) = oneshot::channel();
    register_forward(&q_name, tx).await;

    let addr = SocketAddr::from(([127, 0, 0, 1], config.from_port));

    tokio::spawn(async move {
        let server = TcpListenerStream::new(TcpListener::bind(addr).await.unwrap())
            .take_until(rx)
            .try_for_each(|client_conn| async {
                let pods = pods.clone();
                let pod_port = config.to_port;
                let pod_name = q_name.pod_name.clone();

                tokio::spawn(async move {
                    if let Err(e) =
                        forward_connection(&pods, &pod_name, pod_port, client_conn).await
                    {
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
    });

    return target_pod;
}

/// Stops a connection to a pod.
pub async fn stop(namespace: String, pod_or_service: String) {
    let q_name = QualifiedName::new(namespace, pod_or_service);
    let key = q_name.to_string();

    if let Some(tx) = OPEN_PORTFORWARDS.write().await.remove(&key) {
        tx.send(()).unwrap();
    }
}

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
    Ok(())
}

// ===== ===== HELPER ===== =====

struct QualifiedName {
    namespace: String,
    pod_name: String,
}

impl QualifiedName {
    fn new(namespace: String, name: String) -> Self {
        Self {
            namespace,
            pod_name: name,
        }
    }

    fn to_string(&self) -> String {
        format!("{}/{}", self.namespace, self.pod_name)
    }
}

async fn register_forward(qualified_name: &QualifiedName, sender: Sender<()>) {
    OPEN_PORTFORWARDS
        .write()
        .await
        .insert(qualified_name.to_string(), sender);
}
