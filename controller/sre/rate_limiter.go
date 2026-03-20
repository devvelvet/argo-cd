package sre

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	sretypes "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1/sre"
)

// RateLimiterController manages rate limiting for sync operations
type RateLimiterController struct {
	mu      sync.RWMutex
	configs map[string]*sretypes.RateLimiting
	// Track sync history per app
	syncHistory map[string][]time.Time
	// Track active syncs
	activeSyncs map[string]int32
}

// NewRateLimiterController creates a new rate limiter controller
func NewRateLimiterController() *RateLimiterController {
	return &RateLimiterController{
		configs:     make(map[string]*sretypes.RateLimiting),
		syncHistory: make(map[string][]time.Time),
		activeSyncs: make(map[string]int32),
	}
}

// Register registers rate limiting configuration for an application
func (r *RateLimiterController) Register(appName string, config *sretypes.RateLimiting) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if config == nil || !config.Enabled {
		delete(r.configs, appName)
		return
	}

	r.configs[appName] = config
	log.WithFields(log.Fields{
		"app":              appName,
		"maxSyncsPerHour":  config.MaxSyncsPerHour,
		"maxParallelSyncs": config.MaxParallelSyncs,
	}).Info("Rate limiter registered")
}

// CanSync checks if a sync operation is allowed
func (r *RateLimiterController) CanSync(appName string) (bool, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.configs[appName]
	if !ok {
		return true, "" // No rate limiting
	}

	// Check parallel syncs
	if config.MaxParallelSyncs > 0 {
		active := r.activeSyncs[appName]
		if active >= config.MaxParallelSyncs {
			return false, fmt.Sprintf("Max parallel syncs reached (%d/%d)", active, config.MaxParallelSyncs)
		}
	}

	// Check hourly rate
	if config.MaxSyncsPerHour > 0 {
		history := r.syncHistory[appName]
		oneHourAgo := time.Now().Add(-time.Hour)
		recentCount := int32(0)
		for _, t := range history {
			if t.After(oneHourAgo) {
				recentCount++
			}
		}
		if recentCount >= config.MaxSyncsPerHour {
			return false, fmt.Sprintf("Hourly sync limit reached (%d/%d)", recentCount, config.MaxSyncsPerHour)
		}
	}

	// Check cooldown period
	if config.CooldownPeriod != nil {
		history := r.syncHistory[appName]
		if len(history) > 0 {
			lastSync := history[len(history)-1]
			elapsed := time.Since(lastSync)
			if elapsed < config.CooldownPeriod.Duration {
				remaining := config.CooldownPeriod.Duration - elapsed
				return false, fmt.Sprintf("Cooldown period active (%.0fs remaining)", remaining.Seconds())
			}
		}
	}

	return true, ""
}

// RecordSync records a sync operation
func (r *RateLimiterController) RecordSync(appName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.syncHistory[appName] = append(r.syncHistory[appName], time.Now())
	r.activeSyncs[appName]++

	// Clean old history (keep last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	var cleaned []time.Time
	for _, t := range r.syncHistory[appName] {
		if t.After(cutoff) {
			cleaned = append(cleaned, t)
		}
	}
	r.syncHistory[appName] = cleaned
}

// CompletedSync marks a sync as completed
func (r *RateLimiterController) CompletedSync(appName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeSyncs[appName] > 0 {
		r.activeSyncs[appName]--
	}
}

// GetStats returns rate limiting stats for an application
func (r *RateLimiterController) GetStats(appName string) RateLimitStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := RateLimitStats{}

	config, ok := r.configs[appName]
	if !ok {
		return stats
	}

	stats.MaxSyncsPerHour = config.MaxSyncsPerHour
	stats.MaxParallelSyncs = config.MaxParallelSyncs
	stats.ActiveSyncs = r.activeSyncs[appName]

	oneHourAgo := time.Now().Add(-time.Hour)
	for _, t := range r.syncHistory[appName] {
		if t.After(oneHourAgo) {
			stats.SyncsInLastHour++
		}
	}

	stats.RemainingBudget = config.MaxSyncsPerHour - stats.SyncsInLastHour
	if stats.RemainingBudget < 0 {
		stats.RemainingBudget = 0
	}

	return stats
}

// RateLimitStats contains rate limiting statistics
type RateLimitStats struct {
	MaxSyncsPerHour  int32
	MaxParallelSyncs int32
	ActiveSyncs      int32
	SyncsInLastHour  int32
	RemainingBudget  int32
}
