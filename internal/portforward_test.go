package portforward

import (
	"testing"
	"time"
)

func TestStopForwarding(t *testing.T) {
	// Arrange
	stopCh := make(chan struct{})
	namespace := "test_namespace"
	pod := "another_pod"
	registerForwarding(namespace, pod, stopCh)

	// Act
	StopForwarding(namespace, pod)

	// Assert
	select {
	case <-stopCh:
		// Success
	case <-time.After(5 * time.Second):
		t.Errorf("Channel for stopping the portforward was not closed in time")
	}
}

func TestStopForwardingWhenPortForwardIsNotRegistered(t *testing.T) {
	// Arrange
	namespace := "test_namespace"
	pod := "another_pod"

	// Act
	StopForwarding(namespace, pod)

	// Assert
	// ... should be reached without any panic
}

func TestForwardWithoutValidConfigPath(t *testing.T) {
	// Arrange
	namespace := "any_namespace"
	pod := "any_pod"
	from := 8000
	to := 8000
	invalidPath := "foo/bar"

	// Act
	err := Forward(namespace, pod, from, to, invalidPath, Debug, "")

	// Assert
	if err == nil {
		t.Errorf("Error should be returned when a not valid config path is provided")
	}
}
