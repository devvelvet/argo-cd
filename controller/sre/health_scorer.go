package sre

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

// HealthScorer calculates comprehensive health scores for applications
type HealthScorer struct {
	mu     sync.RWMutex
	scores map[string]*sretypes.HealthScore
	// Component scorers
	sloMonitor      *SLOMonitor
	circuitBreaker  *CircuitBreakerController
	incidentManager *IncidentManager
}

// NewHealthScorer creates a new health scorer
func NewHealthScorer(sloMonitor *SLOMonitor, cb *CircuitBreakerController, im *IncidentManager) *HealthScorer {
	return &HealthScorer{
		scores:          make(map[string]*sretypes.HealthScore),
		sloMonitor:      sloMonitor,
		circuitBreaker:  cb,
		incidentManager: im,
	}
}

// CalculateScore calculates the health score for an application
func (h *HealthScorer) CalculateScore(appName string, syncHealthy bool, resourcesHealthy bool) *sretypes.HealthScore {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := metav1.Now()
	score := &sretypes.HealthScore{
		LastCalculated: &now,
	}

	// Availability score based on sync and resource health
	score.Availability = h.calculateAvailability(appName, syncHealthy, resourcesHealthy)

	// Performance score based on SLO latency metrics
	score.Performance = h.calculatePerformance(appName)

	// Reliability score based on SLO compliance and circuit breaker state
	score.Reliability = h.calculateReliability(appName)

	// Security score (base score, can be enhanced with actual security checks)
	score.Security = h.calculateSecurity(appName)

	// Resource efficiency (base score)
	score.ResourceEfficiency = h.calculateResourceEfficiency(appName)

	// Overall score is weighted average
	score.Overall = (score.Availability*30 + score.Performance*20 + score.Reliability*25 + score.Security*15 + score.ResourceEfficiency*10) / 100

	// Calculate trend
	prevScore, hasPrev := h.scores[appName]
	if hasPrev && prevScore != nil {
		if score.Overall > prevScore.Overall+2 {
			score.Trend = "Improving"
		} else if score.Overall < prevScore.Overall-2 {
			score.Trend = "Degrading"
		} else {
			score.Trend = "Stable"
		}
	} else {
		score.Trend = "Stable"
	}

	h.scores[appName] = score

	log.WithFields(log.Fields{
		"app":          appName,
		"overall":      score.Overall,
		"availability": score.Availability,
		"performance":  score.Performance,
		"reliability":  score.Reliability,
		"trend":        score.Trend,
	}).Debug("Health score calculated")

	return score
}

// calculateAvailability calculates the availability score
func (h *HealthScorer) calculateAvailability(appName string, syncHealthy bool, resourcesHealthy bool) int32 {
	score := int32(100)

	if !syncHealthy {
		score -= 30
	}
	if !resourcesHealthy {
		score -= 40
	}

	// Check for active incidents
	if h.incidentManager != nil {
		incidents := h.incidentManager.GetActiveIncidents(appName)
		for _, inc := range incidents {
			switch inc.Severity {
			case sretypes.SeverityCritical:
				score -= 30
			case sretypes.SeverityHigh:
				score -= 20
			case sretypes.SeverityMedium:
				score -= 10
			case sretypes.SeverityLow:
				score -= 5
			}
		}
	}

	if score < 0 {
		score = 0
	}
	return score
}

// calculatePerformance calculates the performance score
func (h *HealthScorer) calculatePerformance(appName string) int32 {
	score := int32(85) // Base performance score

	if h.sloMonitor != nil {
		statuses := h.sloMonitor.GetStatus(appName)
		for _, s := range statuses {
			if s.Name == "latency" || s.Name == "throughput" {
				if s.Compliance {
					score += 5
				} else {
					score -= 15
				}
			}
		}
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

// calculateReliability calculates the reliability score
func (h *HealthScorer) calculateReliability(appName string) int32 {
	score := int32(90) // Base reliability score

	// Check SLO compliance
	if h.sloMonitor != nil {
		if !h.sloMonitor.IsCompliant(appName) {
			violations := h.sloMonitor.GetViolations(appName)
			score -= int32(len(violations)) * 10
		}
	}

	// Check circuit breaker state
	if h.circuitBreaker != nil {
		cbStatus := h.circuitBreaker.GetStatus(appName)
		if cbStatus != nil {
			switch cbStatus.State {
			case sretypes.CircuitBreakerOpen:
				score -= 30
			case sretypes.CircuitBreakerHalfOpen:
				score -= 15
			}
		}
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

// calculateSecurity calculates the security score
func (h *HealthScorer) calculateSecurity(_ string) int32 {
	// Base security score - can be enhanced with actual security checks
	return 80
}

// calculateResourceEfficiency calculates the resource efficiency score
func (h *HealthScorer) calculateResourceEfficiency(_ string) int32 {
	// Base resource efficiency - can be enhanced with actual resource metrics
	return 75
}

// GetScore returns the current health score for an application
func (h *HealthScorer) GetScore(appName string) *sretypes.HealthScore {
	h.mu.RLock()
	defer h.mu.RUnlock()

	score, ok := h.scores[appName]
	if !ok {
		return nil
	}
	return score
}

// GetScoreHistory returns the score trend over time (simplified)
func (h *HealthScorer) GetScoreHistory(_ string) []HealthScoreEntry {
	// In a real implementation, this would query a time-series store
	return nil
}

// HealthScoreEntry represents a point-in-time health score
type HealthScoreEntry struct {
	Timestamp time.Time
	Score     int32
}
