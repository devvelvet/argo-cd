// Package sre provides the SRE API server for Argo CD.
// It exposes REST endpoints for managing canary deployments, circuit breakers,
// SLO monitoring, incident management, and health scoring.
package sre

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	srecontroller "github.com/argoproj/argo-cd/v3/controller/sre"
)

// Server provides the SRE REST API
type Server struct {
	manager *srecontroller.Manager
}

// NewServer creates a new SRE API server
func NewServer(manager *srecontroller.Manager) *Server {
	return &Server{manager: manager}
}

// RegisterRoutes registers SRE API routes with the given mux
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// SRE Status
	mux.HandleFunc("/api/v1/sre/applications/", s.handleSREStatus)

	// Canary operations
	mux.HandleFunc("/api/v1/sre/canary/promote", s.handleCanaryPromote)
	mux.HandleFunc("/api/v1/sre/canary/rollback", s.handleCanaryRollback)
	mux.HandleFunc("/api/v1/sre/canary/start", s.handleCanaryStart)

	// Incident operations
	mux.HandleFunc("/api/v1/sre/incidents", s.handleIncidents)
	mux.HandleFunc("/api/v1/sre/incidents/acknowledge", s.handleIncidentAcknowledge)
	mux.HandleFunc("/api/v1/sre/incidents/resolve", s.handleIncidentResolve)

	// SLO operations
	mux.HandleFunc("/api/v1/sre/slos", s.handleSLOs)

	// Health score
	mux.HandleFunc("/api/v1/sre/health", s.handleHealthScore)

	// Dashboard aggregate
	mux.HandleFunc("/api/v1/sre/dashboard", s.handleDashboard)

	log.Info("SRE API routes registered")
}

// SREStatusResponse is the response for SRE status
type SREStatusResponse struct {
	AppName          string                     `json:"appName"`
	DeploymentPhase  string                     `json:"deploymentPhase"`
	CanaryWeight     int32                      `json:"canaryWeight"`
	CurrentStep      int32                      `json:"currentStep"`
	TotalSteps       int32                      `json:"totalSteps"`
	HealthScore      *HealthScoreResponse       `json:"healthScore,omitempty"`
	CircuitBreaker   *CircuitBreakerResponse    `json:"circuitBreaker,omitempty"`
	SLOs             []SLOStatusResponse        `json:"slos,omitempty"`
	ActiveIncidents  []IncidentResponse         `json:"activeIncidents,omitempty"`
	Message          string                     `json:"message"`
}

// HealthScoreResponse is the health score response
type HealthScoreResponse struct {
	Overall            int32  `json:"overall"`
	Availability       int32  `json:"availability"`
	Performance        int32  `json:"performance"`
	Reliability        int32  `json:"reliability"`
	Security           int32  `json:"security"`
	ResourceEfficiency int32  `json:"resourceEfficiency"`
	Trend              string `json:"trend"`
}

// CircuitBreakerResponse is the circuit breaker response
type CircuitBreakerResponse struct {
	State        string `json:"state"`
	FailureCount int32  `json:"failureCount"`
	SuccessCount int32  `json:"successCount"`
}

// SLOStatusResponse is the SLO status response
type SLOStatusResponse struct {
	Name                 string  `json:"name"`
	CurrentValue         float64 `json:"currentValue"`
	Target               float64 `json:"target"`
	Compliance           bool    `json:"compliance"`
	ErrorBudgetRemaining float64 `json:"errorBudgetRemaining"`
	BurnRate             float64 `json:"burnRate"`
}

// IncidentResponse is the incident response
type IncidentResponse struct {
	ID                string   `json:"id"`
	Severity          string   `json:"severity"`
	State             string   `json:"state"`
	Summary           string   `json:"summary"`
	CreatedAt         string   `json:"createdAt,omitempty"`
	AffectedResources []string `json:"affectedResources,omitempty"`
}

// DashboardResponse is the aggregate dashboard response
type DashboardResponse struct {
	Applications []SREStatusResponse `json:"applications"`
	TotalApps    int                 `json:"totalApps"`
	HealthyApps  int                 `json:"healthyApps"`
	WarningApps  int                 `json:"warningApps"`
	CriticalApps int                 `json:"criticalApps"`
	OpenIncidents int               `json:"openIncidents"`
	SLOViolations int               `json:"sloViolations"`
}

func (s *Server) handleSREStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appName := r.URL.Query().Get("name")
	if appName == "" {
		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
		return
	}

	status := s.manager.GetSREStatus(appName, true, true)
	resp := SREStatusResponse{
		AppName: appName,
		Message: "SRE status retrieved",
	}

	if status.DeploymentStatus != nil {
		resp.DeploymentPhase = string(status.DeploymentStatus.Phase)
		resp.CanaryWeight = status.DeploymentStatus.CanaryWeight
		resp.CurrentStep = status.DeploymentStatus.CurrentStep
		resp.TotalSteps = status.DeploymentStatus.TotalSteps
		resp.Message = status.DeploymentStatus.Message
	}

	if status.HealthScore != nil {
		resp.HealthScore = &HealthScoreResponse{
			Overall:            status.HealthScore.Overall,
			Availability:       status.HealthScore.Availability,
			Performance:        status.HealthScore.Performance,
			Reliability:        status.HealthScore.Reliability,
			Security:           status.HealthScore.Security,
			ResourceEfficiency: status.HealthScore.ResourceEfficiency,
			Trend:              status.HealthScore.Trend,
		}
	}

	if status.CircuitBreaker != nil {
		resp.CircuitBreaker = &CircuitBreakerResponse{
			State:        string(status.CircuitBreaker.State),
			FailureCount: status.CircuitBreaker.FailureCount,
			SuccessCount: status.CircuitBreaker.SuccessCount,
		}
	}

	for _, slo := range status.SLOs {
		resp.SLOs = append(resp.SLOs, SLOStatusResponse{
			Name:                 slo.Name,
			CurrentValue:         slo.CurrentValue,
			Target:               slo.Target,
			Compliance:           slo.Compliance,
			ErrorBudgetRemaining: slo.ErrorBudgetRemaining,
			BurnRate:             slo.BurnRate,
		})
	}

	for _, inc := range status.Incidents {
		createdAt := ""
		if inc.CreatedAt != nil {
			createdAt = inc.CreatedAt.Format("2006-01-02T15:04:05Z")
		}
		resp.ActiveIncidents = append(resp.ActiveIncidents, IncidentResponse{
			ID:                inc.ID,
			Severity:          string(inc.Severity),
			State:             inc.State,
			Summary:           inc.Summary,
			CreatedAt:         createdAt,
			AffectedResources: inc.AffectedResources,
		})
	}

	writeJSON(w, resp)
}

type canaryRequest struct {
	AppName  string `json:"appName"`
	Reason   string `json:"reason,omitempty"`
	Stable   string `json:"stableRevision,omitempty"`
	Canary   string `json:"canaryRevision,omitempty"`
}

func (s *Server) handleCanaryPromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req canaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := s.manager.PromoteCanary(req.AppName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, status)
}

func (s *Server) handleCanaryRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req canaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reason := req.Reason
	if reason == "" {
		reason = "Manual rollback"
	}

	status, err := s.manager.RollbackCanary(req.AppName, reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, status)
}

func (s *Server) handleCanaryStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req canaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := s.manager.StartCanaryDeployment(req.AppName, req.Stable, req.Canary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, status)
}

type incidentRequest struct {
	AppName    string `json:"appName"`
	IncidentID string `json:"incidentId"`
	Actions    []string `json:"actions,omitempty"`
}

func (s *Server) handleIncidents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appName := r.URL.Query().Get("name")
	if appName == "" {
		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
		return
	}

	incidents := s.manager.IncidentManager().GetAllIncidents(appName)
	writeJSON(w, incidents)
}

func (s *Server) handleIncidentAcknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req incidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.manager.IncidentManager().AcknowledgeIncident(req.AppName, req.IncidentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"status": "acknowledged"})
}

func (s *Server) handleIncidentResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req incidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.manager.IncidentManager().ResolveIncident(req.AppName, req.IncidentID, req.Actions); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"status": "resolved"})
}

func (s *Server) handleSLOs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appName := r.URL.Query().Get("name")
	if appName == "" {
		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
		return
	}

	statuses := s.manager.SLOMonitor().GetStatus(appName)
	writeJSON(w, statuses)
}

func (s *Server) handleHealthScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appName := r.URL.Query().Get("name")
	if appName == "" {
		http.Error(w, "Missing 'name' query parameter", http.StatusBadRequest)
		return
	}

	score := s.manager.HealthScorer().GetScore(appName)
	if score == nil {
		score = s.manager.HealthScorer().CalculateScore(appName, true, true)
	}

	writeJSON(w, score)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := DashboardResponse{}
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n"))
}
