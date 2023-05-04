import sys
import time

from pytest_kind import KindCluster

sys.path.append(".")

import uuid

import pytest

import portforward
import requests


TEST_NAMESPACE = "pftest"


def test_pod_portforward_with_success(kind_cluster: KindCluster):
    # Arrange
    _create_test_resources(kind_cluster)

    # Act & Assert
    pod_name = "nginx"
    local_port = 9000  # from port
    pod_port = 80  # to port
    context = kind_cluster.name
    config = str(kind_cluster.kubeconfig_path.absolute())

    with portforward.forward(
        TEST_NAMESPACE,
        pod_name,
        local_port,
        pod_port,
        config_path=config,
        kube_context=context,
    ):
        response: requests.Response = requests.get("http://localhost:9000")
        assert response.status_code == 200


@pytest.mark.parametrize(
    "namespace,pod,from_port,to_port",
    [
        # Namespace
        ("", "web", 9000, 80),
        ("/test", "web", 9000, 80),
        (1337, "web", 9000, 80),
        (None, "web", 9000, 80),
        # Pod name
        ("test", "", 9000, 80),
        ("test", "web/", 9000, 80),
        ("test", 1337, 9000, 80),
        ("test", None, 9000, 80),
        # From port
        ("test", "web", 1_000_000, 80),
        ("test", "web", 9000.1, 80),
        ("test", "web", -9000, 80),
        ("test", "web", None, 80),
        # To port
        ("test", "web", 9000, 1_000_000),
        ("test", "web", 9000, 80.1),
        ("test", "web", 9000, -80),
        ("test", "web", 9000, None),
    ],
)
def test_forward_invalid_parameter(namespace, pod, from_port, to_port):
    # Arrange
    ...

    # Act and Assert
    with pytest.raises(ValueError):
        with portforward.forward(namespace, pod, from_port, to_port):
            pytest.fail("Should raise error before")


def test_forward_raise_error():
    """Tests the conversion of the C extension error into the Python Error"""

    # Arrange
    namespace = "test" + str(uuid.uuid4())  # Should never exists
    pod = "web"
    from_ = 9000
    to = 80

    # Act and Assert
    with pytest.raises(portforward.PortforwardError):
        with portforward.forward(namespace, pod, from_, to):
            pytest.fail("Should raise error before")


def _create_test_resources(kind_cluster: KindCluster):
    kind_cluster.kubectl("create", "ns", TEST_NAMESPACE)

    for _ in range(0, 100):
        try:
            kind_cluster.kubectl("apply", "-f", "tests/resources.yaml")
            break
        except:
            print("Could not yet create resources")
            time.sleep(1.0)
