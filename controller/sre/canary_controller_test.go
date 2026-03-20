package sre

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

func TestCanaryController_StartDeployment(t *testing.T) {
	ctrl := NewCanaryController()

	weight10 := int32(10)
	weight50 := int32(50)
	weight100 := int32(100)

	strategy := &sretypes.CanaryStrategy{
		MaxWeight: 100,
		Steps: []sretypes.CanaryStep{
			{SetWeight: &weight10},
			{Pause: &sretypes.CanaryPause{}},
			{SetWeight: &weight50},
			{Pause: &sretypes.CanaryPause{}},
			{SetWeight: &weight100},
		},
	}

	status, err := ctrl.StartCanaryDeployment("test-app", strategy, "v1", "v2")
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, sretypes.DeploymentPhasePaused, status.Phase)
	assert.Equal(t, int32(10), status.CanaryWeight)
	assert.Equal(t, "v1", status.StableRevision)
	assert.Equal(t, "v2", status.CanaryRevision)
}

func TestCanaryController_PromoteStep(t *testing.T) {
	ctrl := NewCanaryController()

	weight10 := int32(10)
	weight100 := int32(100)

	strategy := &sretypes.CanaryStrategy{
		Steps: []sretypes.CanaryStep{
			{SetWeight: &weight10},
			{Pause: &sretypes.CanaryPause{}},
			{SetWeight: &weight100},
		},
	}

	_, err := ctrl.StartCanaryDeployment("test-app", strategy, "v1", "v2")
	require.NoError(t, err)

	// Promote from paused state
	status, err := ctrl.PromoteCanaryStep("test-app")
	require.NoError(t, err)
	assert.Equal(t, sretypes.DeploymentPhaseCompleted, status.Phase)
	assert.Equal(t, int32(100), status.CanaryWeight)
}

func TestCanaryController_Rollback(t *testing.T) {
	ctrl := NewCanaryController()

	weight10 := int32(10)

	strategy := &sretypes.CanaryStrategy{
		Steps: []sretypes.CanaryStep{
			{SetWeight: &weight10},
			{Pause: &sretypes.CanaryPause{}},
		},
	}

	_, err := ctrl.StartCanaryDeployment("test-app", strategy, "v1", "v2")
	require.NoError(t, err)

	status, err := ctrl.RollbackCanary("test-app", "analysis failed")
	require.NoError(t, err)
	assert.Equal(t, sretypes.DeploymentPhaseRollingBack, status.Phase)
	assert.Equal(t, int32(0), status.CanaryWeight)
}

func TestCanaryController_NilStrategy(t *testing.T) {
	ctrl := NewCanaryController()

	_, err := ctrl.StartCanaryDeployment("test-app", nil, "v1", "v2")
	assert.Error(t, err)
}

func TestCanaryController_EmptySteps(t *testing.T) {
	ctrl := NewCanaryController()

	strategy := &sretypes.CanaryStrategy{
		Steps: []sretypes.CanaryStep{},
	}

	_, err := ctrl.StartCanaryDeployment("test-app", strategy, "v1", "v2")
	assert.Error(t, err)
}

func TestCanaryController_PromoteNotPaused(t *testing.T) {
	ctrl := NewCanaryController()
	_, err := ctrl.PromoteCanaryStep("nonexistent")
	assert.Error(t, err)
}
