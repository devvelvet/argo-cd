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

// SLOMonitor monitors SLOs for applications
type SLOMonitor struct {
	mu       sync.RWMutex
	configs  map[string][]sretypes.SLOSpec
	statuses map[string][]sretypes.SLOStatus
	cancel   context.CancelFunc
}

// NewSLOMonitor creates a new SLO monitor
func NewSLOMonitor() *SLOMonitor {
	return &SLOMonitor{
		configs:  make(map[string][]sretypes.SLOSpec),
		statuses: make(map[string][]sretypes.SLOStatus),
	}
}

// Register registers SLO configurations for an application
func (m *SLOMonitor) Register(appName string, slos []sretypes.SLOSpec) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[appName] = slos

	// Initialize statuses
	statuses := make([]sretypes.SLOStatus, len(slos))
	for i, slo := range slos {
		statuses[i] = sretypes.SLOStatus{
			Name:                 slo.Name,
			Target:               slo.Target,
			CurrentValue:         100.0, // Start with perfect score
			Compliance:           true,
			ErrorBudgetRemaining: 100.0,
			BurnRate:             0.0,
		}
	}
	m.statuses[appName] = statuses

	log.WithFields(log.Fields{
		"app":      appName,
		"sloCount": len(slos),
	}).Info("SLO monitoring registered")
}

// Start starts the SLO monitoring loop
func (m *SLOMonitor) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.evaluateAll()
			}
		}
	}()

	log.Info("SLO monitor started")
}

// Stop stops the SLO monitor
func (m *SLOMonitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

// evaluateAll evaluates all registered SLOs
func (m *SLOMonitor) evaluateAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for appName, slos := range m.configs {
		for i, slo := range slos {
			if i < len(m.statuses[appName]) {
				m.evaluateSLO(appName, &slo, &m.statuses[appName][i])
			}
		}
	}
}

// evaluateSLO evaluates a single SLO
func (m *SLOMonitor) evaluateSLO(appName string, spec *sretypes.SLOSpec, status *sretypes.SLOStatus) {
	now := metav1.Now()
	status.LastMeasured = &now

	// Calculate error budget remaining
	if spec.Target > 0 {
		errorBudget := 100.0 - spec.Target
		if errorBudget > 0 {
			consumed := 100.0 - status.CurrentValue
			status.ErrorBudgetRemaining = ((errorBudget - consumed) / errorBudget) * 100.0
			if status.ErrorBudgetRemaining < 0 {
				status.ErrorBudgetRemaining = 0
			}
		}
	}

	// Check compliance
	status.Compliance = status.CurrentValue >= spec.Target

	if !status.Compliance {
		log.WithFields(log.Fields{
			"app":          appName,
			"slo":          spec.Name,
			"current":      status.CurrentValue,
			"target":       spec.Target,
			"errorBudget":  status.ErrorBudgetRemaining,
		}).Warn("SLO violation detected")
	}
}

// RecordEvent records an event that affects SLIs
func (m *SLOMonitor) RecordEvent(appName string, sloName string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	statuses, ok := m.statuses[appName]
	if !ok {
		return
	}

	for i := range statuses {
		if statuses[i].Name == sloName {
			// Simple exponential moving average for SLI
			alpha := 0.1
			var newValue float64
			if success {
				newValue = 100.0
			}
			statuses[i].CurrentValue = statuses[i].CurrentValue*(1-alpha) + newValue*alpha
			break
		}
	}
}

// GetStatus returns SLO statuses for an application
func (m *SLOMonitor) GetStatus(appName string) []sretypes.SLOStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses, ok := m.statuses[appName]
	if !ok {
		return nil
	}

	result := make([]sretypes.SLOStatus, len(statuses))
	copy(result, statuses)
	return result
}

// IsCompliant checks if all SLOs for an app are compliant
func (m *SLOMonitor) IsCompliant(appName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses, ok := m.statuses[appName]
	if !ok {
		return true // No SLOs means compliant
	}

	for _, s := range statuses {
		if !s.Compliance {
			return false
		}
	}
	return true
}

// GetErrorBudgetSummary returns a summary of error budgets
func (m *SLOMonitor) GetErrorBudgetSummary(appName string) map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]float64)
	statuses, ok := m.statuses[appName]
	if !ok {
		return result
	}

	for _, s := range statuses {
		result[s.Name] = s.ErrorBudgetRemaining
	}
	return result
}

// GetViolations returns SLOs that are currently in violation
func (m *SLOMonitor) GetViolations(appName string) []sretypes.SLOStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var violations []sretypes.SLOStatus
	statuses, ok := m.statuses[appName]
	if !ok {
		return violations
	}

	for _, s := range statuses {
		if !s.Compliance {
			violations = append(violations, s)
		}
	}
	return violations
}

// FormatSLOReport generates a human-readable SLO report
func (m *SLOMonitor) FormatSLOReport(appName string) string {
	statuses := m.GetStatus(appName)
	if len(statuses) == 0 {
		return fmt.Sprintf("No SLOs configured for %s", appName)
	}

	report := fmt.Sprintf("SLO Report for %s:\n", appName)
	for _, s := range statuses {
		compliance := "PASS"
		if !s.Compliance {
			compliance = "FAIL"
		}
		report += fmt.Sprintf("  [%s] %s: %.2f%% (target: %.2f%%) | Error Budget: %.1f%%\n",
			compliance, s.Name, s.CurrentValue, s.Target, s.ErrorBudgetRemaining)
	}
	return report
}
