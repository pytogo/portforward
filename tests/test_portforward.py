"""
Tests for `portforward` package.

!!! It will only test the Python side !!!
"""
import time
from typing import List
import uuid
import sys

sys.path.append(".")

import pytest
from pytest_kind import KindCluster
import requests
from pykube import Namespace

import portforward


TEST_NAMESPACE = "pftest"
TEST_CONTEXT = "kind-pytest-kind"


def test_pod_portforward_with_success(kind_cluster: KindCluster):
    # Arrange
    _create_test_resources(kind_cluster)

    # Act & Assert
    pod_name = "test-pod"
    local_port = 9000  # from port
    pod_port = 3000  # to port
    context = TEST_CONTEXT
    config = str(kind_cluster.kubeconfig_path.absolute())

    url = f"http://localhost:{local_port}/ping"

    with portforward.forward(
        TEST_NAMESPACE,
        pod_name,
        local_port,
        pod_port,
        config_path=config,
        kube_context=context,
    ):
        response: requests.Response = requests.get(url)
        assert response.status_code == 200

    with pytest.raises(requests.exceptions.ConnectionError):
        response: requests.Response = requests.get(url)
        pytest.fail("Portforward should be closed after leaving the context manager")


def test_service_portforward_with_success(kind_cluster: KindCluster):
    # Arrange
    _create_test_resources(kind_cluster)

    # Act & Assert
    service_name = "test-service"
    local_port = 9001  # from port
    pod_port = 3000  # to port
    context = TEST_CONTEXT
    config = str(kind_cluster.kubeconfig_path.absolute())

    url = f"http://localhost:{local_port}/ping"

    with portforward.forward(
        TEST_NAMESPACE,
        service_name,
        local_port,
        pod_port,
        config_path=config,
        kube_context=context,
    ):
        response: requests.Response = requests.get(url)
        assert response.status_code == 200

    with pytest.raises(requests.exceptions.ConnectionError):
        response: requests.Response = requests.get("http://localhost:9000/ping")
        pytest.fail("Portforward should be closed after leaving the context manager")


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
    for namespace in Namespace.objects(kind_cluster.api).filter():
        # When test namespace already exists then the other resource
        # should also already exists.
        if namespace.name == TEST_NAMESPACE:
            return

    kind_cluster.kubectl("create", "ns", TEST_NAMESPACE)

    for _ in range(0, 100):
        try:
            kind_cluster.kubectl("apply", "-f", "tests/resources.yaml")
            break
        except:
            print("Could not yet create resources")
            time.sleep(1.0)
