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
use log::*;
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
use typed_builder::TypedBuilder;

#[derive(TypedBuilder, Clone, Debug)]
pub struct ForwardConfig {
    namespace: String,
    pod_or_service: String,
    from_port: u16,
    to_port: u16,
    config_path: String,
    kube_context: String,
}

/// Creates a connection to a pod. It returns the actual pod name for the portforward.
/// It differs from `pod_or_service` when `pod_or_service` represents a service.
pub async fn forward(config: ForwardConfig) -> anyhow::Result<String> {
    debug!("{:?}", config);

    let q_name = QualifiedName::new(config.namespace, config.pod_or_service, config.to_port);
    let target_pod = q_name.pod_name.clone();

    let kube_config = kube::config::Kubeconfig::read_from(config.config_path.clone())?;
    let mut options = kube::config::KubeConfigOptions::default();
    options.context = Some(config.kube_context);
    let client_config = kube::config::Config::from_custom_kubeconfig(kube_config, &options).await?;
    let client = Client::try_from(client_config)?;
    let pods: Api<Pod> = Api::namespaced(client, &q_name.namespace);

    let running = await_condition(pods.clone(), &q_name.pod_name, is_pod_running());
    let _ = tokio::time::timeout(std::time::Duration::from_secs(30), running).await?;

    let (tx, rx): (Sender<()>, Receiver<()>) = oneshot::channel();

    let forwarding = Forwarding::builder().cancel_sender(tx).build();

    PORTFORWARD_REGISTRY.register(&q_name, forwarding).await;

    let addr = SocketAddr::from(([127, 0, 0, 1], config.from_port));
    let tcp_listener = TcpListener::bind(addr).await?;
    let forward_task =
        setup_forward_task(tcp_listener, rx, pods, config.to_port, target_pod.clone());

    tokio::spawn(forward_task);

    return Ok(target_pod);
}

async fn setup_forward_task(
    tcp_listener: TcpListener,
    rx: Receiver<()>,
    pods: Api<Pod>,
    pod_port: u16,
    pod_name: String,
) {
    let server = TcpListenerStream::new(tcp_listener)
        .take_until(rx)
        .try_for_each(|client_conn| async {
            let pods = pods.clone();
            let pod_name = pod_name.clone();

            tokio::spawn(async move {
                let forwarding = forward_connection(&pods, &pod_name, pod_port, client_conn);
                if let Err(e) = forwarding.await {
                    error!("failed to forward connection: {}", e);
                }
            });
            // keep the server running
            Ok(())
        });
    if let Err(e) = server.await {
        error!("server error: {}", e);
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

/// Stops a connection to a pod.
pub async fn stop(namespace: String, actual_pod: String, to_port: u16) {
    let q_name = QualifiedName::new(namespace, actual_pod, to_port);

    PORTFORWARD_REGISTRY.stop(&q_name).await;
}

// ===== ===== ============ ===== =====
// ===== ===== REGISTRATION ===== =====
// ===== ===== ============ ===== =====

static PORTFORWARD_REGISTRY: Lazy<Registry> = Lazy::new(|| Registry::new());

#[derive(TypedBuilder)]
struct Forwarding {
    cancel_sender: Sender<()>,
}

struct Registry {
    register: RwLock<HashMap<QualifiedName, Forwarding>>,
}

impl Registry {
    fn new() -> Self {
        let m = HashMap::new();
        Self {
            register: RwLock::new(m),
        }
    }

    async fn register(&self, qualified_name: &QualifiedName, forwarding: Forwarding) {
        self.register
            .write()
            .await
            .insert(qualified_name.clone(), forwarding);
    }

    async fn stop(&self, qualified_name: &QualifiedName) {
        if let Some(forwarding) = self.register.write().await.remove(&qualified_name) {
            if let Err(_) = forwarding.cancel_sender.send(()) {
                warn!("Unable to close port-forwarding");
            }
        }
    }
}

#[derive(Eq, PartialEq, Hash, Clone)]
struct QualifiedName {
    namespace: String,
    pod_name: String,
    target_port: u16,
}

impl QualifiedName {
    fn new(namespace: String, name: String, target_port: u16) -> Self {
        Self {
            namespace,
            pod_name: name,
            target_port,
        }
    }
}
