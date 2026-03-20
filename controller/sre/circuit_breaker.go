package sre

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

// CircuitBreakerController manages circuit breakers for applications
type CircuitBreakerController struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreakerInstance
}

// CircuitBreakerInstance represents a single circuit breaker
type CircuitBreakerInstance struct {
	Config           *sretypes.CircuitBreaker
	Status           *sretypes.CircuitBreakerStatus
	LastAttempt      time.Time
	ConsecutiveFails int32
	ConsecutiveOK    int32
}

// NewCircuitBreakerController creates a new circuit breaker controller
func NewCircuitBreakerController() *CircuitBreakerController {
	return &CircuitBreakerController{
		breakers: make(map[string]*CircuitBreakerInstance),
	}
}

// Register registers a circuit breaker for an application
func (c *CircuitBreakerController) Register(appName string, config *sretypes.CircuitBreaker) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if config == nil || !config.Enabled {
		delete(c.breakers, appName)
		return
	}

	now := metav1.Now()
	c.breakers[appName] = &CircuitBreakerInstance{
		Config: config,
		Status: &sretypes.CircuitBreakerStatus{
			State:           sretypes.CircuitBreakerClosed,
			LastStateChange: &now,
		},
	}

	log.WithField("app", appName).Info("Circuit breaker registered")
}

// CanSync checks if a sync operation is allowed by the circuit breaker
func (c *CircuitBreakerController) CanSync(appName string) (bool, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instance, ok := c.breakers[appName]
	if !ok {
		return true, "" // No circuit breaker, allow
	}

	switch instance.Status.State {
	case sretypes.CircuitBreakerOpen:
		// Check if timeout has elapsed
		if instance.Config.Timeout != nil && instance.Status.LastStateChange != nil {
			elapsed := time.Since(instance.Status.LastStateChange.Time)
			if elapsed >= instance.Config.Timeout.Duration {
				// Transition to half-open (done in RecordResult)
				return true, "Circuit breaker transitioning to half-open"
			}
		}
		return false, fmt.Sprintf("Circuit breaker is OPEN for %s - sync blocked after %d consecutive failures", appName, instance.Status.FailureCount)

	case sretypes.CircuitBreakerHalfOpen:
		if instance.ConsecutiveOK < instance.Config.HalfOpenMaxRequests {
			return true, "Circuit breaker is HALF-OPEN - limited requests allowed"
		}
		return false, "Circuit breaker HALF-OPEN max requests reached"

	default: // Closed
		return true, ""
	}
}

// RecordSuccess records a successful sync
func (c *CircuitBreakerController) RecordSuccess(appName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	instance, ok := c.breakers[appName]
	if !ok {
		return
	}

	instance.ConsecutiveFails = 0
	instance.ConsecutiveOK++
	instance.Status.SuccessCount++
	instance.LastAttempt = time.Now()

	switch instance.Status.State {
	case sretypes.CircuitBreakerHalfOpen:
		if instance.ConsecutiveOK >= instance.Config.SuccessThreshold {
			c.transitionState(appName, instance, sretypes.CircuitBreakerClosed)
		}
	case sretypes.CircuitBreakerOpen:
		// Transition to half-open on success during timeout check
		c.transitionState(appName, instance, sretypes.CircuitBreakerHalfOpen)
	}
}

// RecordFailure records a failed sync
func (c *CircuitBreakerController) RecordFailure(appName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	instance, ok := c.breakers[appName]
	if !ok {
		return
	}

	now := metav1.Now()
	instance.ConsecutiveOK = 0
	instance.ConsecutiveFails++
	instance.Status.FailureCount++
	instance.Status.LastFailure = &now
	instance.LastAttempt = time.Now()

	switch instance.Status.State {
	case sretypes.CircuitBreakerClosed:
		if instance.ConsecutiveFails >= instance.Config.FailureThreshold {
			c.transitionState(appName, instance, sretypes.CircuitBreakerOpen)
		}
	case sretypes.CircuitBreakerHalfOpen:
		c.transitionState(appName, instance, sretypes.CircuitBreakerOpen)
	}
}

// transitionState transitions the circuit breaker to a new state
func (c *CircuitBreakerController) transitionState(appName string, instance *CircuitBreakerInstance, newState sretypes.CircuitBreakerState) {
	oldState := instance.Status.State
	now := metav1.Now()
	instance.Status.State = newState
	instance.Status.LastStateChange = &now
	instance.ConsecutiveFails = 0
	instance.ConsecutiveOK = 0

	log.WithFields(log.Fields{
		"app":      appName,
		"oldState": oldState,
		"newState": newState,
	}).Info("Circuit breaker state transition")
}

// GetStatus returns the current circuit breaker status
func (c *CircuitBreakerController) GetStatus(appName string) *sretypes.CircuitBreakerStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instance, ok := c.breakers[appName]
	if !ok {
		return nil
	}

	return instance.Status
}

// GetAllStatuses returns all circuit breaker statuses
func (c *CircuitBreakerController) GetAllStatuses() map[string]*sretypes.CircuitBreakerStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*sretypes.CircuitBreakerStatus, len(c.breakers))
	for name, instance := range c.breakers {
		result[name] = instance.Status
	}
	return result
}
