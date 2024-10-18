//! portforwards contains the "nice" Rust API.
//! It is based on [kube-rs examples](https://github.com/kube-rs/kube/blob/33ba90e0a003c915801bb9b6c4961b7d0b721889/examples/pod_portforward_bind.rs).

use anyhow::anyhow;
use anyhow::Context;
use futures::{StreamExt, TryStreamExt};
use k8s_openapi::api::core::v1::{Pod, Service, ServiceSpec};
use kube::{
    api::{Api, ListParams},
    runtime::wait::{await_condition, conditions::is_pod_running},
    Client,
};
use log::*;
use once_cell::sync::Lazy;
use std::net::SocketAddr;
use std::{collections::HashMap, path::Path};
use std::str::FromStr;
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
    bind_address: String,
    to_port: u16,
    config_path: String,
    kube_context: String,
}

/// Creates a connection to a pod. It returns a `(pod_name, from_port)` tuple
/// with the actual pod name and local port used for the portforward.
/// It differs from `pod_or_service` when `pod_or_service` represents a service.
pub async fn forward(config: ForwardConfig) -> anyhow::Result<(String, u16)> {
    debug!("{:?}", config);

    let client_config = load_config(&config.config_path, &config.kube_context).await?;
    let client = Client::try_from(client_config)?;

    let target_pod = find_pod(&client, &config.namespace, &config.pod_or_service).await?;

    let q_name: QualifiedName = QualifiedName::new(&config.namespace, &target_pod, config.to_port);

    let pods: Api<Pod> = Api::namespaced(client, &config.namespace);
    let running = await_condition(pods.clone(), &q_name.pod_name, is_pod_running());
    let _ = tokio::time::timeout(std::time::Duration::from_secs(30), running).await?;

    let (tx, rx): (Sender<()>, Receiver<()>) = oneshot::channel();

    let forwarding = Forwarding::builder().cancel_sender(tx).build();

    PORTFORWARD_REGISTRY.register(&q_name, forwarding).await;

    let addr = SocketAddr::from_str(&config.bind_address).with_context(move || config.bind_address)?;
    let tcp_listener = TcpListener::bind(addr).await?;
    let from_port = tcp_listener.local_addr()?.port();
    let forward_task = setup_forward_task(
        tcp_listener,
        rx,
        pods,
        config.to_port,
        q_name.pod_name.clone(),
    );

    tokio::spawn(forward_task);

    return Ok((q_name.pod_name, from_port));
}

async fn load_config(
    config_path: &str,
    kube_context: &str,
) -> anyhow::Result<kube::config::Config> {
    // When no config file exists we assume that we should use incluster config.
    if !Path::new(config_path).exists() {
        let incluster_config = kube::config::Config::incluster()?;
        return Ok(incluster_config);
    }

    let kube_config = kube::config::Kubeconfig::read_from(config_path)?;
    let mut options = kube::config::KubeConfigOptions::default();

    // "" is the sign for using default context
    if kube_context != "" {
        options.context = Some(kube_context.to_string());
    }

    let client_config = kube::config::Config::from_custom_kubeconfig(kube_config, &options).await?;

    Ok(client_config)
}

/// Tries to find a pod by the given or name or checks if it is a service
/// and if a pod matches the selector of the service.
///
/// It will return the first thing that it can find.
async fn find_pod(
    client: &Client,
    namespace: &str,
    pod_or_service: &str,
) -> anyhow::Result<String> {
    let pods: Api<Pod> = Api::namespaced(client.clone(), namespace);

    if let Some(_) = pods.get_opt(pod_or_service).await? {
        return Ok(pod_or_service.to_string());
    }

    let services: Api<Service> = Api::namespaced(client.clone(), namespace);

    let service = services.get(pod_or_service).await?;
    let selector = service
        .spec
        .and_then(to_label_selector)
        .ok_or_else(|| anyhow!("No selector could be found for service {}", pod_or_service))?;

    let lp = ListParams::default().labels(&selector);

    let pods_for_svc = pods.list(&lp).await?;

    for pod in pods_for_svc {
        if let Some(name) = pod.metadata.name {
            return Ok(name);
        }
    }

    Err(anyhow!(
        "No pod could be found for service {}",
        pod_or_service
    ))
}

fn to_label_selector(service_spec: ServiceSpec) -> Option<String> {
    let selector = service_spec.selector?;
    let selector = selector
        .iter()
        .map(|(key, val)| format!("{key}={val}"))
        .collect::<Vec<String>>()
        .join(",");

    Some(selector)
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
    let q_name = QualifiedName::new(&namespace, &actual_pod, to_port);

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
    fn new(namespace: &str, name: &str, target_port: u16) -> Self {
        Self {
            namespace: namespace.to_string(),
            pod_name: name.to_string(),
            target_port,
        }
    }
}
