package sre

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

// Manager is the central SRE manager that coordinates all SRE components
type Manager struct {
	mu              sync.RWMutex
	canary          *CanaryController
	circuitBreaker  *CircuitBreakerController
	rateLimiter     *RateLimiterController
	sloMonitor      *SLOMonitor
	incidentManager *IncidentManager
	healthScorer    *HealthScorer
	configs         map[string]*sretypes.SREConfig
}

// NewManager creates a new SRE Manager with all components initialized
func NewManager() *Manager {
	canary := NewCanaryController()
	cb := NewCircuitBreakerController()
	rl := NewRateLimiterController()
	slo := NewSLOMonitor()
	im := NewIncidentManager()
	hs := NewHealthScorer(slo, cb, im)

	return &Manager{
		canary:          canary,
		circuitBreaker:  cb,
		rateLimiter:     rl,
		sloMonitor:      slo,
		incidentManager: im,
		healthScorer:    hs,
		configs:         make(map[string]*sretypes.SREConfig),
	}
}

// Start starts the SRE manager and all its components
func (m *Manager) Start(ctx context.Context) {
	m.sloMonitor.Start(ctx)
	log.Info("SRE Manager started")
}

// Stop stops the SRE manager and all its components
func (m *Manager) Stop() {
	m.sloMonitor.Stop()
	log.Info("SRE Manager stopped")
}

// RegisterApp registers an application with the SRE manager
func (m *Manager) RegisterApp(appName string, config *sretypes.SREConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		config = sretypes.DefaultSREConfig()
	}

	m.configs[appName] = config

	// Register with each component
	if config.CircuitBreaker != nil {
		m.circuitBreaker.Register(appName, config.CircuitBreaker)
	}
	if config.RateLimiting != nil {
		m.rateLimiter.Register(appName, config.RateLimiting)
	}
	if config.SLOs != nil {
		m.sloMonitor.Register(appName, config.SLOs)
	}
	if config.IncidentPolicy != nil {
		m.incidentManager.Register(appName, config.IncidentPolicy)
	}

	log.WithField("app", appName).Info("Application registered with SRE manager")
}

// PreSyncCheck runs all pre-sync checks (circuit breaker, rate limiter)
func (m *Manager) PreSyncCheck(appName string) (bool, string) {
	// Check circuit breaker
	if allowed, reason := m.circuitBreaker.CanSync(appName); !allowed {
		return false, "Circuit breaker: " + reason
	}

	// Check rate limiter
	if allowed, reason := m.rateLimiter.CanSync(appName); !allowed {
		return false, "Rate limiter: " + reason
	}

	return true, ""
}

// OnSyncStarted is called when a sync operation starts
func (m *Manager) OnSyncStarted(appName string) {
	m.rateLimiter.RecordSync(appName)
}

// OnSyncCompleted is called when a sync operation completes
func (m *Manager) OnSyncCompleted(appName string, success bool) {
	m.rateLimiter.CompletedSync(appName)

	if success {
		m.circuitBreaker.RecordSuccess(appName)
		m.sloMonitor.RecordEvent(appName, "availability", true)
	} else {
		m.circuitBreaker.RecordFailure(appName)
		m.sloMonitor.RecordEvent(appName, "availability", false)

		// Auto-create incident on failure
		_, _ = m.incidentManager.AutoCreateIncident(appName, "Sync operation failed", nil)
	}
}

// StartCanaryDeployment initiates a canary deployment
func (m *Manager) StartCanaryDeployment(appName string, stableRevision, canaryRevision string) (*sretypes.DeploymentStatus, error) {
	m.mu.RLock()
	config, ok := m.configs[appName]
	m.mu.RUnlock()

	if !ok || config.DeploymentStrategy == nil || config.DeploymentStrategy.Canary == nil {
		// Use default canary config
		defaultConfig := sretypes.DefaultSREConfig()
		return m.canary.StartCanaryDeployment(appName, defaultConfig.DeploymentStrategy.Canary, stableRevision, canaryRevision)
	}

	return m.canary.StartCanaryDeployment(appName, config.DeploymentStrategy.Canary, stableRevision, canaryRevision)
}

// PromoteCanary promotes the canary to the next step
func (m *Manager) PromoteCanary(appName string) (*sretypes.DeploymentStatus, error) {
	return m.canary.PromoteCanaryStep(appName)
}

// RollbackCanary rolls back the canary deployment
func (m *Manager) RollbackCanary(appName string, reason string) (*sretypes.DeploymentStatus, error) {
	return m.canary.RollbackCanary(appName, reason)
}

// GetSREStatus returns the complete SRE status for an application
func (m *Manager) GetSREStatus(appName string, syncHealthy bool, resourcesHealthy bool) *sretypes.SREStatus {
	now := metav1.Now()
	status := &sretypes.SREStatus{
		LastUpdated: &now,
	}

	// Deployment status
	if deployStatus, err := m.canary.GetDeploymentStatus(appName); err == nil {
		status.DeploymentStatus = deployStatus
	}

	// Circuit breaker status
	status.CircuitBreaker = m.circuitBreaker.GetStatus(appName)

	// SLO statuses
	status.SLOs = m.sloMonitor.GetStatus(appName)

	// Health score
	status.HealthScore = m.healthScorer.CalculateScore(appName, syncHealthy, resourcesHealthy)

	// Active incidents
	status.Incidents = m.incidentManager.GetActiveIncidents(appName)

	return status
}

// GetConfig returns the SRE configuration for an application
func (m *Manager) GetConfig(appName string) *sretypes.SREConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, ok := m.configs[appName]
	if !ok {
		return nil
	}
	return config
}

// Canary returns the canary controller
func (m *Manager) Canary() *CanaryController {
	return m.canary
}

// CircuitBreaker returns the circuit breaker controller
func (m *Manager) CircuitBreaker() *CircuitBreakerController {
	return m.circuitBreaker
}

// RateLimiter returns the rate limiter controller
func (m *Manager) RateLimiter() *RateLimiterController {
	return m.rateLimiter
}

// SLOMonitor returns the SLO monitor
func (m *Manager) SLOMonitor() *SLOMonitor {
	return m.sloMonitor
}

// IncidentManager returns the incident manager
func (m *Manager) IncidentManager() *IncidentManager {
	return m.incidentManager
}

// HealthScorer returns the health scorer
func (m *Manager) HealthScorer() *HealthScorer {
	return m.healthScorer
}
