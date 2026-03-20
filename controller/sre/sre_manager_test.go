package sre

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

func TestManager_DefaultConfig(t *testing.T) {
	manager := NewManager()
	config := sretypes.DefaultSREConfig()

	assert.NotNil(t, config)
	assert.NotNil(t, config.DeploymentStrategy)
	assert.Equal(t, sretypes.DeploymentStrategyCanary, config.DeploymentStrategy.Type)
	assert.NotNil(t, config.DeploymentStrategy.Canary)
	assert.Greater(t, len(config.DeploymentStrategy.Canary.Steps), 0)
	assert.NotNil(t, config.CircuitBreaker)
	assert.True(t, config.CircuitBreaker.Enabled)
	assert.NotNil(t, config.RateLimiting)
	assert.NotNil(t, config.IncidentPolicy)
	assert.NotNil(t, manager)
}

func TestManager_RegisterApp(t *testing.T) {
	manager := NewManager()

	// Register with default config
	manager.RegisterApp("test-app", nil)

	config := manager.GetConfig("test-app")
	assert.NotNil(t, config)
	assert.Equal(t, sretypes.DeploymentStrategyCanary, config.DeploymentStrategy.Type)
}

func TestManager_PreSyncCheck(t *testing.T) {
	manager := NewManager()
	manager.RegisterApp("test-app", nil)

	allowed, _ := manager.PreSyncCheck("test-app")
	assert.True(t, allowed)
}

func TestManager_SyncLifecycle(t *testing.T) {
	manager := NewManager()
	manager.RegisterApp("test-app", nil)

	// Start sync
	manager.OnSyncStarted("test-app")

	// Complete sync successfully
	manager.OnSyncCompleted("test-app", true)

	// Check SLO status
	statuses := manager.SLOMonitor().GetStatus("test-app")
	assert.NotEmpty(t, statuses)
}

func TestManager_SyncFailureCreatesIncident(t *testing.T) {
	manager := NewManager()
	manager.RegisterApp("test-app", nil)

	manager.OnSyncStarted("test-app")
	manager.OnSyncCompleted("test-app", false)

	incidents := manager.IncidentManager().GetActiveIncidents("test-app")
	assert.NotEmpty(t, incidents)
	assert.Equal(t, "Open", incidents[0].State)
}

func TestManager_CanaryDeployment(t *testing.T) {
	manager := NewManager()
	manager.RegisterApp("test-app", nil)

	status, err := manager.StartCanaryDeployment("test-app", "v1", "v2")
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "v1", status.StableRevision)
	assert.Equal(t, "v2", status.CanaryRevision)
}

func TestManager_GetSREStatus(t *testing.T) {
	manager := NewManager()
	manager.RegisterApp("test-app", nil)

	status := manager.GetSREStatus("test-app", true, true)
	assert.NotNil(t, status)
	assert.NotNil(t, status.HealthScore)
	assert.NotNil(t, status.CircuitBreaker)
	assert.NotEmpty(t, status.SLOs)
	assert.NotNil(t, status.LastUpdated)
}

func TestManager_HealthScore(t *testing.T) {
	manager := NewManager()
	manager.RegisterApp("test-app", nil)

	status := manager.GetSREStatus("test-app", true, true)
	assert.NotNil(t, status.HealthScore)
	assert.Greater(t, status.HealthScore.Overall, int32(0))
	assert.Greater(t, status.HealthScore.Availability, int32(0))
	assert.Greater(t, status.HealthScore.Performance, int32(0))
	assert.Greater(t, status.HealthScore.Reliability, int32(0))
}
