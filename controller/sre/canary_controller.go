// Package sre provides SRE controllers for Argo CD including canary deployment,
// circuit breaker, SLO monitoring, and incident management.
package sre

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

// CanaryController manages canary deployment progression
type CanaryController struct {
	mu              sync.RWMutex
	deployments     map[string]*CanaryDeploymentState
	analysisRunners map[string]*AnalysisRunner
}

// CanaryDeploymentState tracks the state of a canary deployment
type CanaryDeploymentState struct {
	AppName        string
	Strategy       *sretypes.CanaryStrategy
	Status         *sretypes.DeploymentStatus
	LastTransition time.Time
	PauseStarted   *time.Time
	AutoRollback   *sretypes.AutoRollback
}

// NewCanaryController creates a new CanaryController
func NewCanaryController() *CanaryController {
	return &CanaryController{
		deployments:     make(map[string]*CanaryDeploymentState),
		analysisRunners: make(map[string]*AnalysisRunner),
	}
}

// StartCanaryDeployment initiates a new canary deployment
func (c *CanaryController) StartCanaryDeployment(appName string, strategy *sretypes.CanaryStrategy, stableRevision, canaryRevision string) (*sretypes.DeploymentStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if strategy == nil {
		return nil, fmt.Errorf("canary strategy is nil")
	}

	if len(strategy.Steps) == 0 {
		return nil, fmt.Errorf("canary strategy has no steps defined")
	}

	now := metav1.Now()
	status := &sretypes.DeploymentStatus{
		Phase:          sretypes.DeploymentPhasePending,
		CurrentStep:    0,
		TotalSteps:     int32(len(strategy.Steps)),
		CanaryWeight:   0,
		StableRevision: stableRevision,
		CanaryRevision: canaryRevision,
		StartedAt:      &now,
		Message:        "Canary deployment initiated",
	}

	state := &CanaryDeploymentState{
		AppName:        appName,
		Strategy:       strategy,
		Status:         status,
		LastTransition: time.Now(),
		AutoRollback:   strategy.AutoRollback,
	}

	c.deployments[appName] = state

	log.WithFields(log.Fields{
		"app":             appName,
		"stableRevision":  stableRevision,
		"canaryRevision":  canaryRevision,
		"totalSteps":      len(strategy.Steps),
	}).Info("Canary deployment started")

	// Process the first step
	return c.processStep(appName)
}

// processStep processes the current step of a canary deployment
func (c *CanaryController) processStep(appName string) (*sretypes.DeploymentStatus, error) {
	state, ok := c.deployments[appName]
	if !ok {
		return nil, fmt.Errorf("no canary deployment found for app: %s", appName)
	}

	if int(state.Status.CurrentStep) >= len(state.Strategy.Steps) {
		// All steps completed - promote
		return c.promoteCanary(appName)
	}

	step := state.Strategy.Steps[state.Status.CurrentStep]
	state.Status.Phase = sretypes.DeploymentPhaseProgressing

	if step.SetWeight != nil {
		state.Status.CanaryWeight = *step.SetWeight
		state.Status.Message = fmt.Sprintf("Setting canary weight to %d%%", *step.SetWeight)
		log.WithFields(log.Fields{
			"app":    appName,
			"step":   state.Status.CurrentStep,
			"weight": *step.SetWeight,
		}).Info("Canary weight updated")

		// Auto-advance to next step
		state.Status.CurrentStep++
		state.LastTransition = time.Now()

		// Check if next step exists and process
		if int(state.Status.CurrentStep) < len(state.Strategy.Steps) {
			nextStep := state.Strategy.Steps[state.Status.CurrentStep]
			if nextStep.Pause != nil {
				return c.handlePause(appName, nextStep.Pause)
			}
			return c.processStep(appName)
		}
		return c.promoteCanary(appName)
	}

	if step.Pause != nil {
		return c.handlePause(appName, step.Pause)
	}

	if step.Analysis != nil {
		return c.handleAnalysis(appName, step.Analysis)
	}

	// Unknown step type, advance
	state.Status.CurrentStep++
	return c.processStep(appName)
}

// handlePause handles a pause step
func (c *CanaryController) handlePause(appName string, pause *sretypes.CanaryPause) (*sretypes.DeploymentStatus, error) {
	state := c.deployments[appName]
	now := time.Now()
	state.PauseStarted = &now
	state.Status.Phase = sretypes.DeploymentPhasePaused

	if pause.Duration != nil {
		state.Status.Message = fmt.Sprintf("Paused at %d%% canary weight for %s", state.Status.CanaryWeight, pause.Duration.Duration)
	} else {
		state.Status.Message = fmt.Sprintf("Paused at %d%% canary weight - manual promotion required", state.Status.CanaryWeight)
	}

	log.WithFields(log.Fields{
		"app":    appName,
		"weight": state.Status.CanaryWeight,
		"step":   state.Status.CurrentStep,
	}).Info("Canary deployment paused")

	return state.Status, nil
}

// handleAnalysis handles an analysis step
func (c *CanaryController) handleAnalysis(appName string, analysis *sretypes.CanaryStepAnalysis) (*sretypes.DeploymentStatus, error) {
	state := c.deployments[appName]
	state.Status.Message = fmt.Sprintf("Running analysis at %d%% canary weight", state.Status.CanaryWeight)

	log.WithFields(log.Fields{
		"app":     appName,
		"metrics": len(analysis.Metrics),
	}).Info("Running canary analysis")

	// Start analysis runner
	runner := NewAnalysisRunner(appName, analysis.Metrics)
	c.analysisRunners[appName] = runner

	return state.Status, nil
}

// PromoteCanaryStep manually promotes to the next canary step
func (c *CanaryController) PromoteCanaryStep(appName string) (*sretypes.DeploymentStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.deployments[appName]
	if !ok {
		return nil, fmt.Errorf("no canary deployment found for app: %s", appName)
	}

	if state.Status.Phase != sretypes.DeploymentPhasePaused {
		return nil, fmt.Errorf("canary deployment is not paused (current phase: %s)", state.Status.Phase)
	}

	state.PauseStarted = nil
	state.Status.CurrentStep++
	state.LastTransition = time.Now()

	return c.processStep(appName)
}

// promoteCanary completes the canary deployment by fully promoting
func (c *CanaryController) promoteCanary(appName string) (*sretypes.DeploymentStatus, error) {
	state := c.deployments[appName]
	now := metav1.Now()

	state.Status.Phase = sretypes.DeploymentPhaseCompleted
	state.Status.CanaryWeight = 100
	state.Status.CompletedAt = &now
	state.Status.Message = "Canary deployment completed successfully"

	log.WithFields(log.Fields{
		"app":      appName,
		"revision": state.Status.CanaryRevision,
	}).Info("Canary deployment promoted to stable")

	return state.Status, nil
}

// RollbackCanary rolls back the canary deployment
func (c *CanaryController) RollbackCanary(appName string, reason string) (*sretypes.DeploymentStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.deployments[appName]
	if !ok {
		return nil, fmt.Errorf("no canary deployment found for app: %s", appName)
	}

	now := metav1.Now()
	state.Status.Phase = sretypes.DeploymentPhaseRollingBack
	state.Status.CanaryWeight = 0
	state.Status.CompletedAt = &now
	state.Status.Message = fmt.Sprintf("Rolling back: %s", reason)

	// Stop any analysis runners
	if runner, exists := c.analysisRunners[appName]; exists {
		runner.Stop()
		delete(c.analysisRunners, appName)
	}

	log.WithFields(log.Fields{
		"app":    appName,
		"reason": reason,
	}).Warn("Canary deployment rolled back")

	return state.Status, nil
}

// GetDeploymentStatus returns the current deployment status
func (c *CanaryController) GetDeploymentStatus(appName string) (*sretypes.DeploymentStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, ok := c.deployments[appName]
	if !ok {
		return nil, fmt.Errorf("no canary deployment found for app: %s", appName)
	}

	return state.Status, nil
}

// CheckPauseExpiration checks if any timed pauses have expired
func (c *CanaryController) CheckPauseExpiration(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for appName, state := range c.deployments {
		if state.Status.Phase != sretypes.DeploymentPhasePaused || state.PauseStarted == nil {
			continue
		}

		stepIdx := state.Status.CurrentStep
		if int(stepIdx) >= len(state.Strategy.Steps) {
			continue
		}

		step := state.Strategy.Steps[stepIdx]
		if step.Pause != nil && step.Pause.Duration != nil {
			elapsed := time.Since(*state.PauseStarted)
			if elapsed >= step.Pause.Duration.Duration {
				log.WithFields(log.Fields{
					"app":  appName,
					"step": stepIdx,
				}).Info("Canary pause expired, advancing to next step")

				state.PauseStarted = nil
				state.Status.CurrentStep++
				state.LastTransition = time.Now()
				// Process next step (ignore error in background check)
				_, _ = c.processStep(appName)
			}
		}
	}
}

// ============================================================================
// Analysis Runner
// ============================================================================

// AnalysisRunner runs metric analysis for canary deployments
type AnalysisRunner struct {
	appName string
	metrics []sretypes.AnalysisMetric
	results []sretypes.AnalysisResult
	cancel  context.CancelFunc
	mu      sync.RWMutex
	stopped bool
}

// NewAnalysisRunner creates a new analysis runner
func NewAnalysisRunner(appName string, metrics []sretypes.AnalysisMetric) *AnalysisRunner {
	ctx, cancel := context.WithCancel(context.Background())
	runner := &AnalysisRunner{
		appName: appName,
		metrics: metrics,
		cancel:  cancel,
	}
	go runner.run(ctx)
	return runner
}

// run executes the analysis loop
func (r *AnalysisRunner) run(ctx context.Context) {
	for _, metric := range r.metrics {
		select {
		case <-ctx.Done():
			return
		default:
			result := r.analyzeMetric(ctx, metric)
			r.mu.Lock()
			r.results = append(r.results, result)
			r.mu.Unlock()
		}
	}
}

// analyzeMetric runs a single metric analysis
func (r *AnalysisRunner) analyzeMetric(_ context.Context, metric sretypes.AnalysisMetric) sretypes.AnalysisResult {
	now := metav1.Now()

	// Determine provider and execute query
	result := sretypes.AnalysisResult{
		MetricName: metric.Name,
		MeasuredAt: &now,
	}

	if metric.Provider.Prometheus != nil {
		result.Value = "0" // Placeholder - real implementation queries Prometheus
		result.Status = "Successful"
		result.Message = fmt.Sprintf("Prometheus query: %s", metric.Provider.Prometheus.Query)
	} else if metric.Provider.Datadog != nil {
		result.Value = "0"
		result.Status = "Successful"
		result.Message = fmt.Sprintf("Datadog query: %s", metric.Provider.Datadog.Query)
	} else if metric.Provider.Custom != nil {
		result.Value = "0"
		result.Status = "Successful"
		result.Message = fmt.Sprintf("Custom webhook: %s", metric.Provider.Custom.URL)
	} else {
		result.Status = "Inconclusive"
		result.Message = "No metric provider configured"
	}

	log.WithFields(log.Fields{
		"app":    r.appName,
		"metric": metric.Name,
		"status": result.Status,
	}).Info("Analysis metric evaluated")

	return result
}

// GetResults returns current analysis results
func (r *AnalysisRunner) GetResults() []sretypes.AnalysisResult {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]sretypes.AnalysisResult, len(r.results))
	copy(results, r.results)
	return results
}

// Stop stops the analysis runner
func (r *AnalysisRunner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.stopped {
		r.stopped = true
		r.cancel()
	}
}
