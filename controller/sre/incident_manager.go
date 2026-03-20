package sre

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

// IncidentManager manages incidents for applications
type IncidentManager struct {
	mu        sync.RWMutex
	policies  map[string]*sretypes.IncidentPolicy
	incidents map[string][]sretypes.IncidentStatus
	counter   int64
}

// NewIncidentManager creates a new incident manager
func NewIncidentManager() *IncidentManager {
	return &IncidentManager{
		policies:  make(map[string]*sretypes.IncidentPolicy),
		incidents: make(map[string][]sretypes.IncidentStatus),
	}
}

// Register registers an incident policy for an application
func (m *IncidentManager) Register(appName string, policy *sretypes.IncidentPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if policy == nil || !policy.Enabled {
		delete(m.policies, appName)
		return
	}

	m.policies[appName] = policy
	log.WithField("app", appName).Info("Incident policy registered")
}

// CreateIncident creates a new incident
func (m *IncidentManager) CreateIncident(appName string, severity sretypes.IncidentSeverity, summary string, affectedResources []string) (*sretypes.IncidentStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	now := metav1.Now()

	incident := sretypes.IncidentStatus{
		ID:                fmt.Sprintf("INC-%06d", m.counter),
		Severity:          severity,
		State:             "Open",
		Summary:           summary,
		CreatedAt:         &now,
		AffectedResources: affectedResources,
	}

	m.incidents[appName] = append(m.incidents[appName], incident)

	log.WithFields(log.Fields{
		"app":      appName,
		"id":       incident.ID,
		"severity": severity,
		"summary":  summary,
	}).Warn("Incident created")

	// Trigger notifications
	m.triggerNotifications(appName, &incident)

	return &incident, nil
}

// AutoCreateIncident creates an incident based on policy
func (m *IncidentManager) AutoCreateIncident(appName string, summary string, affectedResources []string) (*sretypes.IncidentStatus, error) {
	m.mu.RLock()
	policy, ok := m.policies[appName]
	m.mu.RUnlock()

	if !ok || !policy.AutoCreate {
		return nil, nil // No auto-creation
	}

	severity := policy.Severity
	if severity == "" {
		severity = sretypes.SeverityHigh
	}

	return m.CreateIncident(appName, severity, summary, affectedResources)
}

// AcknowledgeIncident acknowledges an incident
func (m *IncidentManager) AcknowledgeIncident(appName string, incidentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	incidents, ok := m.incidents[appName]
	if !ok {
		return fmt.Errorf("no incidents found for app: %s", appName)
	}

	for i := range incidents {
		if incidents[i].ID == incidentID {
			incidents[i].State = "Acknowledged"
			log.WithFields(log.Fields{
				"app": appName,
				"id":  incidentID,
			}).Info("Incident acknowledged")
			return nil
		}
	}

	return fmt.Errorf("incident %s not found for app %s", incidentID, appName)
}

// ResolveIncident resolves an incident
func (m *IncidentManager) ResolveIncident(appName string, incidentID string, actions []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	incidents, ok := m.incidents[appName]
	if !ok {
		return fmt.Errorf("no incidents found for app: %s", appName)
	}

	for i := range incidents {
		if incidents[i].ID == incidentID {
			now := metav1.Now()
			incidents[i].State = "Resolved"
			incidents[i].ResolvedAt = &now
			incidents[i].RemediationActions = actions
			log.WithFields(log.Fields{
				"app":     appName,
				"id":      incidentID,
				"actions": actions,
			}).Info("Incident resolved")
			return nil
		}
	}

	return fmt.Errorf("incident %s not found for app %s", incidentID, appName)
}

// GetActiveIncidents returns active (non-resolved) incidents
func (m *IncidentManager) GetActiveIncidents(appName string) []sretypes.IncidentStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []sretypes.IncidentStatus
	incidents, ok := m.incidents[appName]
	if !ok {
		return active
	}

	for _, inc := range incidents {
		if inc.State != "Resolved" {
			active = append(active, inc)
		}
	}
	return active
}

// GetAllIncidents returns all incidents for an application
func (m *IncidentManager) GetAllIncidents(appName string) []sretypes.IncidentStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incidents, ok := m.incidents[appName]
	if !ok {
		return nil
	}

	result := make([]sretypes.IncidentStatus, len(incidents))
	copy(result, incidents)
	return result
}

// triggerNotifications sends notifications for an incident
func (m *IncidentManager) triggerNotifications(appName string, incident *sretypes.IncidentStatus) {
	policy, ok := m.policies[appName]
	if !ok {
		return
	}

	for _, target := range policy.NotificationTargets {
		log.WithFields(log.Fields{
			"app":    appName,
			"id":     incident.ID,
			"type":   target.Type,
			"target": target.Target,
		}).Info("Incident notification sent")
	}
}

// ExecuteRunbooks executes automated runbooks for an application
func (m *IncidentManager) ExecuteRunbooks(appName string, incidentID string) []string {
	m.mu.RLock()
	policy, ok := m.policies[appName]
	m.mu.RUnlock()

	if !ok {
		return nil
	}

	var executedActions []string
	for _, runbook := range policy.Runbooks {
		if !runbook.AutoExecute {
			continue
		}
		for _, action := range runbook.Actions {
			actionDesc := fmt.Sprintf("%s: %s (runbook: %s)", action.Type, action.Condition, runbook.Name)
			executedActions = append(executedActions, actionDesc)
			log.WithFields(log.Fields{
				"app":      appName,
				"incident": incidentID,
				"runbook":  runbook.Name,
				"action":   action.Type,
			}).Info("Runbook action executed")
		}
	}
	return executedActions
}

// GetIncidentMetrics returns incident metrics for an application
func (m *IncidentManager) GetIncidentMetrics(appName string) IncidentMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := IncidentMetrics{}
	incidents, ok := m.incidents[appName]
	if !ok {
		return metrics
	}

	for _, inc := range incidents {
		metrics.Total++
		switch inc.State {
		case "Open":
			metrics.Open++
		case "Acknowledged":
			metrics.Acknowledged++
		case "Resolved":
			metrics.Resolved++
			if inc.CreatedAt != nil && inc.ResolvedAt != nil {
				duration := inc.ResolvedAt.Time.Sub(inc.CreatedAt.Time)
				metrics.TotalResolutionTime += duration
			}
		}
		switch inc.Severity {
		case sretypes.SeverityCritical:
			metrics.Critical++
		case sretypes.SeverityHigh:
			metrics.High++
		case sretypes.SeverityMedium:
			metrics.Medium++
		case sretypes.SeverityLow:
			metrics.Low++
		}
	}

	if metrics.Resolved > 0 {
		metrics.AvgResolutionTime = metrics.TotalResolutionTime / time.Duration(metrics.Resolved)
	}

	return metrics
}

// IncidentMetrics contains aggregated incident metrics
type IncidentMetrics struct {
	Total               int
	Open                int
	Acknowledged        int
	Resolved            int
	Critical            int
	High                int
	Medium              int
	Low                 int
	TotalResolutionTime time.Duration
	AvgResolutionTime   time.Duration
}
