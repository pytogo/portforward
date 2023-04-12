//! portforwards contains the "nice" Rust API.

use std::{collections::HashMap, sync::Mutex};

use once_cell::sync::Lazy;
use tokio::sync::RwLock;
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
pub async fn forward(config: ForwardConfig) {}

/// Stops a connection to a pod.
pub async fn stop(namespace: String, pod_or_service: String) {}

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
