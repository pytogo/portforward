"""
Tests for `portforward` package.

!!! It will only test the Python side !!!
"""
import uuid

import pytest

import portforward


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
    ]
)
def test_forward_invalid_parameter(namespace, pod, from_port, to_port):
    # Arrange
    ...

    # Act and Assert
    with pytest.raises(ValueError):
        with portforward.forward(namespace, pod, from_port, to_port):
            pytest.fail("Should raise error before")


def test_forward_raise_error():
    """ Tests the conversion of the C extension error into the Python Error """

    # Arrange
    namespace = "test" + str(uuid.uuid4())  # Should never exists
    pod = "web"
    from_ = 9000
    to = 80

    # Act and Assert
    with pytest.raises(portforward.PortforwardError):
        with portforward.forward(namespace, pod, from_, to):
            pytest.fail("Should raise error before")
