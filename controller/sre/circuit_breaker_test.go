package sre

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

func TestCircuitBreaker_ClosedByDefault(t *testing.T) {
	cb := NewCircuitBreakerController()
	cb.Register("test-app", &sretypes.CircuitBreaker{
		Enabled:          true,
		FailureThreshold: 3,
		SuccessThreshold: 2,
	})

	allowed, _ := cb.CanSync("test-app")
	assert.True(t, allowed)

	status := cb.GetStatus("test-app")
	assert.Equal(t, sretypes.CircuitBreakerClosed, status.State)
}

func TestCircuitBreaker_OpensOnFailures(t *testing.T) {
	cb := NewCircuitBreakerController()
	cb.Register("test-app", &sretypes.CircuitBreaker{
		Enabled:          true,
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          &metav1.Duration{Duration: 30 * time.Second},
	})

	// Record failures
	cb.RecordFailure("test-app")
	cb.RecordFailure("test-app")

	// Still closed after 2 failures (threshold is 3)
	status := cb.GetStatus("test-app")
	assert.Equal(t, sretypes.CircuitBreakerClosed, status.State)

	cb.RecordFailure("test-app")

	// Now open
	status = cb.GetStatus("test-app")
	assert.Equal(t, sretypes.CircuitBreakerOpen, status.State)

	allowed, reason := cb.CanSync("test-app")
	assert.False(t, allowed)
	assert.Contains(t, reason, "OPEN")
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	cb := NewCircuitBreakerController()
	cb.Register("test-app", &sretypes.CircuitBreaker{
		Enabled:             true,
		FailureThreshold:    2,
		SuccessThreshold:    2,
		HalfOpenMaxRequests: 3,
	})

	// Open the circuit
	cb.RecordFailure("test-app")
	cb.RecordFailure("test-app")

	// Transition to half-open via success
	cb.RecordSuccess("test-app")
	status := cb.GetStatus("test-app")
	assert.Equal(t, sretypes.CircuitBreakerHalfOpen, status.State)

	// Close circuit with enough successes
	cb.RecordSuccess("test-app")
	status = cb.GetStatus("test-app")
	assert.Equal(t, sretypes.CircuitBreakerClosed, status.State)
}

func TestCircuitBreaker_UnregisteredApp(t *testing.T) {
	cb := NewCircuitBreakerController()

	allowed, _ := cb.CanSync("unknown-app")
	assert.True(t, allowed)

	status := cb.GetStatus("unknown-app")
	assert.Nil(t, status)
}

func TestCircuitBreaker_DisabledConfig(t *testing.T) {
	cb := NewCircuitBreakerController()
	cb.Register("test-app", &sretypes.CircuitBreaker{
		Enabled: false,
	})

	allowed, _ := cb.CanSync("test-app")
	assert.True(t, allowed)
}
