// Package sre provides SRE (Site Reliability Engineering) types for Argo CD.
// This package extends Argo CD with canary deployment, circuit breaker,
// traffic management, SLO/SLI tracking, auto-rollback, and incident management.
package sre

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ============================================================================
// Deployment Strategy Types
// ============================================================================

// DeploymentStrategyType defines the type of deployment strategy
type DeploymentStrategyType string

const (
	// DeploymentStrategyCanary performs a canary deployment (default)
	DeploymentStrategyCanary DeploymentStrategyType = "Canary"
	// DeploymentStrategyBlueGreen performs a blue-green deployment
	DeploymentStrategyBlueGreen DeploymentStrategyType = "BlueGreen"
	// DeploymentStrategyRolling performs a standard rolling update
	DeploymentStrategyRolling DeploymentStrategyType = "Rolling"
	// DeploymentStrategyAB performs A/B testing deployment
	DeploymentStrategyAB DeploymentStrategyType = "ABTesting"
)

// DeploymentStrategy defines the deployment strategy configuration
type DeploymentStrategy struct {
	// Type is the deployment strategy type. Defaults to Canary.
	// +optional
	Type DeploymentStrategyType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`
	// Canary holds canary deployment configuration
	// +optional
	Canary *CanaryStrategy `json:"canary,omitempty" protobuf:"bytes,2,opt,name=canary"`
	// BlueGreen holds blue-green deployment configuration
	// +optional
	BlueGreen *BlueGreenStrategy `json:"blueGreen,omitempty" protobuf:"bytes,3,opt,name=blueGreen"`
	// ABTesting holds A/B testing deployment configuration
	// +optional
	ABTesting *ABTestingStrategy `json:"abTesting,omitempty" protobuf:"bytes,4,opt,name=abTesting"`
}

// CanaryStrategy defines canary deployment configuration
type CanaryStrategy struct {
	// Steps defines the canary progression steps
	Steps []CanaryStep `json:"steps,omitempty" protobuf:"bytes,1,rep,name=steps"`
	// MaxWeight is the maximum traffic weight for canary (0-100). Default: 100
	// +optional
	MaxWeight int32 `json:"maxWeight,omitempty" protobuf:"varint,2,opt,name=maxWeight"`
	// TrafficRouting defines how traffic is routed between canary and stable
	// +optional
	TrafficRouting *TrafficRouting `json:"trafficRouting,omitempty" protobuf:"bytes,3,opt,name=trafficRouting"`
	// Analysis defines the analysis to run during canary progression
	// +optional
	Analysis *CanaryAnalysis `json:"analysis,omitempty" protobuf:"bytes,4,opt,name=analysis"`
	// AutoRollback defines automatic rollback conditions
	// +optional
	AutoRollback *AutoRollback `json:"autoRollback,omitempty" protobuf:"bytes,5,opt,name=autoRollback"`
}

// CanaryStep defines a single step in canary progression
type CanaryStep struct {
	// SetWeight sets the canary traffic weight percentage (0-100)
	// +optional
	SetWeight *int32 `json:"setWeight,omitempty" protobuf:"varint,1,opt,name=setWeight"`
	// Pause pauses the canary progression
	// +optional
	Pause *CanaryPause `json:"pause,omitempty" protobuf:"bytes,2,opt,name=pause"`
	// Analysis runs analysis during this step
	// +optional
	Analysis *CanaryStepAnalysis `json:"analysis,omitempty" protobuf:"bytes,3,opt,name=analysis"`
	// SetHeaderRouting sets header-based routing for canary
	// +optional
	SetHeaderRouting *SetHeaderRouting `json:"setHeaderRouting,omitempty" protobuf:"bytes,4,opt,name=setHeaderRouting"`
}

// CanaryPause defines pause configuration
type CanaryPause struct {
	// Duration is how long to pause. If omitted, pauses indefinitely until manually promoted.
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty" protobuf:"bytes,1,opt,name=duration"`
}

// CanaryStepAnalysis defines analysis configuration for a canary step
type CanaryStepAnalysis struct {
	// Metrics defines the metrics to analyze
	Metrics []AnalysisMetric `json:"metrics,omitempty" protobuf:"bytes,1,rep,name=metrics"`
}

// SetHeaderRouting defines header-based routing
type SetHeaderRouting struct {
	// Match defines header match rules
	Match []HeaderMatch `json:"match,omitempty" protobuf:"bytes,1,rep,name=match"`
}

// HeaderMatch defines a header matching rule
type HeaderMatch struct {
	// HeaderName is the name of the header
	HeaderName string `json:"headerName" protobuf:"bytes,1,opt,name=headerName"`
	// HeaderValue is the regex value to match
	HeaderValue string `json:"headerValue" protobuf:"bytes,2,opt,name=headerValue"`
}

// CanaryAnalysis defines analysis configuration for the canary
type CanaryAnalysis struct {
	// Metrics defines the metrics to analyze
	Metrics []AnalysisMetric `json:"metrics,omitempty" protobuf:"bytes,1,rep,name=metrics"`
	// Interval is the time between analysis runs
	// +optional
	Interval *metav1.Duration `json:"interval,omitempty" protobuf:"bytes,2,opt,name=interval"`
	// FailureLimit is the max number of failed analyses before rollback
	// +optional
	FailureLimit *int32 `json:"failureLimit,omitempty" protobuf:"varint,3,opt,name=failureLimit"`
	// SuccessCondition defines when analysis is considered successful
	// +optional
	SuccessCondition string `json:"successCondition,omitempty" protobuf:"bytes,4,opt,name=successCondition"`
}

// AnalysisMetric defines a metric for canary analysis
type AnalysisMetric struct {
	// Name is the name of the metric
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Provider defines the metric provider
	Provider MetricProvider `json:"provider" protobuf:"bytes,2,opt,name=provider"`
	// SuccessCondition is a condition expression for metric success
	SuccessCondition string `json:"successCondition,omitempty" protobuf:"bytes,3,opt,name=successCondition"`
	// FailureCondition is a condition expression for metric failure
	FailureCondition string `json:"failureCondition,omitempty" protobuf:"bytes,4,opt,name=failureCondition"`
	// Interval is the time between metric measurements
	// +optional
	Interval *metav1.Duration `json:"interval,omitempty" protobuf:"bytes,5,opt,name=interval"`
	// FailureLimit is how many failures before the metric is considered failed
	// +optional
	FailureLimit *int32 `json:"failureLimit,omitempty" protobuf:"varint,6,opt,name=failureLimit"`
}

// MetricProvider defines a source for metric data
type MetricProvider struct {
	// Prometheus defines a Prometheus metric provider
	// +optional
	Prometheus *PrometheusMetricProvider `json:"prometheus,omitempty" protobuf:"bytes,1,opt,name=prometheus"`
	// Datadog defines a Datadog metric provider
	// +optional
	Datadog *DatadogMetricProvider `json:"datadog,omitempty" protobuf:"bytes,2,opt,name=datadog"`
	// CloudWatch defines an AWS CloudWatch metric provider
	// +optional
	CloudWatch *CloudWatchMetricProvider `json:"cloudWatch,omitempty" protobuf:"bytes,3,opt,name=cloudWatch"`
	// Custom defines a custom webhook metric provider
	// +optional
	Custom *CustomMetricProvider `json:"custom,omitempty" protobuf:"bytes,4,opt,name=custom"`
}

// PrometheusMetricProvider defines Prometheus metric collection
type PrometheusMetricProvider struct {
	// Address is the Prometheus server address
	Address string `json:"address" protobuf:"bytes,1,opt,name=address"`
	// Query is the PromQL query
	Query string `json:"query" protobuf:"bytes,2,opt,name=query"`
}

// DatadogMetricProvider defines Datadog metric collection
type DatadogMetricProvider struct {
	// Query is the Datadog metric query
	Query string `json:"query" protobuf:"bytes,1,opt,name=query"`
	// APIKey is a reference to the Datadog API key secret
	APIKey string `json:"apiKey,omitempty" protobuf:"bytes,2,opt,name=apiKey"`
	// AppKey is a reference to the Datadog App key secret
	AppKey string `json:"appKey,omitempty" protobuf:"bytes,3,opt,name=appKey"`
}

// CloudWatchMetricProvider defines AWS CloudWatch metric collection
type CloudWatchMetricProvider struct {
	// MetricName is the CloudWatch metric name
	MetricName string `json:"metricName" protobuf:"bytes,1,opt,name=metricName"`
	// Namespace is the CloudWatch metric namespace
	Namespace string `json:"namespace" protobuf:"bytes,2,opt,name=namespace"`
	// Period is the CloudWatch metric period in seconds
	Period int32 `json:"period,omitempty" protobuf:"varint,3,opt,name=period"`
	// Statistic is the CloudWatch statistic (Average, Sum, etc.)
	Statistic string `json:"statistic,omitempty" protobuf:"bytes,4,opt,name=statistic"`
}

// CustomMetricProvider defines a custom webhook-based metric provider
type CustomMetricProvider struct {
	// URL is the webhook URL to call
	URL string `json:"url" protobuf:"bytes,1,opt,name=url"`
	// Method is the HTTP method to use (GET, POST)
	Method string `json:"method,omitempty" protobuf:"bytes,2,opt,name=method"`
	// Headers are additional HTTP headers
	Headers map[string]string `json:"headers,omitempty" protobuf:"bytes,3,opt,name=headers"`
	// Body is the request body template (for POST)
	Body string `json:"body,omitempty" protobuf:"bytes,4,opt,name=body"`
	// JSONPath is the JSON path to extract the metric value from the response
	JSONPath string `json:"jsonPath,omitempty" protobuf:"bytes,5,opt,name=jsonPath"`
}

// BlueGreenStrategy defines blue-green deployment configuration
type BlueGreenStrategy struct {
	// ActiveService is the service name for the active (stable) version
	ActiveService string `json:"activeService" protobuf:"bytes,1,opt,name=activeService"`
	// PreviewService is the service name for the preview (new) version
	PreviewService string `json:"previewService" protobuf:"bytes,2,opt,name=previewService"`
	// AutoPromotionEnabled automatically promotes the preview to active
	// +optional
	AutoPromotionEnabled *bool `json:"autoPromotionEnabled,omitempty" protobuf:"varint,3,opt,name=autoPromotionEnabled"`
	// AutoPromotionSeconds is the delay before auto promotion
	// +optional
	AutoPromotionSeconds int32 `json:"autoPromotionSeconds,omitempty" protobuf:"varint,4,opt,name=autoPromotionSeconds"`
	// PrePromotionAnalysis defines analysis to run before promotion
	// +optional
	PrePromotionAnalysis *CanaryAnalysis `json:"prePromotionAnalysis,omitempty" protobuf:"bytes,5,opt,name=prePromotionAnalysis"`
	// PostPromotionAnalysis defines analysis to run after promotion
	// +optional
	PostPromotionAnalysis *CanaryAnalysis `json:"postPromotionAnalysis,omitempty" protobuf:"bytes,6,opt,name=postPromotionAnalysis"`
	// ScaleDownDelaySeconds defines delay before scaling down old version
	// +optional
	ScaleDownDelaySeconds *int32 `json:"scaleDownDelaySeconds,omitempty" protobuf:"varint,7,opt,name=scaleDownDelaySeconds"`
}

// ABTestingStrategy defines A/B testing deployment configuration
type ABTestingStrategy struct {
	// HeaderRouting configures header-based routing for A/B testing
	HeaderRouting *SetHeaderRouting `json:"headerRouting,omitempty" protobuf:"bytes,1,opt,name=headerRouting"`
	// WeightA is the traffic weight for version A (0-100)
	WeightA int32 `json:"weightA,omitempty" protobuf:"varint,2,opt,name=weightA"`
	// WeightB is the traffic weight for version B (0-100)
	WeightB int32 `json:"weightB,omitempty" protobuf:"varint,3,opt,name=weightB"`
	// Analysis defines the analysis to run for A/B comparison
	Analysis *CanaryAnalysis `json:"analysis,omitempty" protobuf:"bytes,4,opt,name=analysis"`
}

// TrafficRouting defines how traffic is routed
type TrafficRouting struct {
	// Istio configures Istio traffic routing
	// +optional
	Istio *IstioTrafficRouting `json:"istio,omitempty" protobuf:"bytes,1,opt,name=istio"`
	// Nginx configures NGINX Ingress traffic routing
	// +optional
	Nginx *NginxTrafficRouting `json:"nginx,omitempty" protobuf:"bytes,2,opt,name=nginx"`
	// ALB configures AWS ALB traffic routing
	// +optional
	ALB *ALBTrafficRouting `json:"alb,omitempty" protobuf:"bytes,3,opt,name=alb"`
	// SMI configures SMI traffic routing
	// +optional
	SMI *SMITrafficRouting `json:"smi,omitempty" protobuf:"bytes,4,opt,name=smi"`
}

// IstioTrafficRouting configures Istio for canary traffic management
type IstioTrafficRouting struct {
	// VirtualService references the Istio VirtualService
	VirtualService IstioVirtualService `json:"virtualService" protobuf:"bytes,1,opt,name=virtualService"`
	// DestinationRule references the Istio DestinationRule
	// +optional
	DestinationRule *IstioDestinationRule `json:"destinationRule,omitempty" protobuf:"bytes,2,opt,name=destinationRule"`
}

// IstioVirtualService references an Istio VirtualService
type IstioVirtualService struct {
	// Name is the name of the VirtualService
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Routes are the route names within the VirtualService
	Routes []string `json:"routes,omitempty" protobuf:"bytes,2,rep,name=routes"`
}

// IstioDestinationRule references an Istio DestinationRule
type IstioDestinationRule struct {
	// Name is the name of the DestinationRule
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// CanarySubsetName is the subset name for canary
	CanarySubsetName string `json:"canarySubsetName,omitempty" protobuf:"bytes,2,opt,name=canarySubsetName"`
	// StableSubsetName is the subset name for stable
	StableSubsetName string `json:"stableSubsetName,omitempty" protobuf:"bytes,3,opt,name=stableSubsetName"`
}

// NginxTrafficRouting configures NGINX Ingress for canary traffic
type NginxTrafficRouting struct {
	// StableIngress is the name of the stable ingress resource
	StableIngress string `json:"stableIngress" protobuf:"bytes,1,opt,name=stableIngress"`
	// AnnotationPrefix is the prefix for canary annotations. Default: nginx.ingress.kubernetes.io
	// +optional
	AnnotationPrefix string `json:"annotationPrefix,omitempty" protobuf:"bytes,2,opt,name=annotationPrefix"`
}

// ALBTrafficRouting configures AWS ALB for canary traffic
type ALBTrafficRouting struct {
	// Ingress is the name of the ALB ingress
	Ingress string `json:"ingress" protobuf:"bytes,1,opt,name=ingress"`
	// ServicePort is the port of the service
	ServicePort int32 `json:"servicePort" protobuf:"varint,2,opt,name=servicePort"`
}

// SMITrafficRouting configures SMI TrafficSplit for canary
type SMITrafficRouting struct {
	// RootService is the name of the root service
	RootService string `json:"rootService,omitempty" protobuf:"bytes,1,opt,name=rootService"`
	// TrafficSplitName is the name of the TrafficSplit resource
	TrafficSplitName string `json:"trafficSplitName,omitempty" protobuf:"bytes,2,opt,name=trafficSplitName"`
}

// AutoRollback defines automatic rollback configuration
type AutoRollback struct {
	// Enabled enables automatic rollback on failure
	Enabled bool `json:"enabled" protobuf:"varint,1,opt,name=enabled"`
	// OnFailure triggers rollback when analysis fails
	// +optional
	OnFailure bool `json:"onFailure,omitempty" protobuf:"varint,2,opt,name=onFailure"`
	// OnError triggers rollback on errors
	// +optional
	OnError bool `json:"onError,omitempty" protobuf:"varint,3,opt,name=onError"`
	// OnHealthCheckFailed triggers rollback when health checks fail
	// +optional
	OnHealthCheckFailed bool `json:"onHealthCheckFailed,omitempty" protobuf:"varint,4,opt,name=onHealthCheckFailed"`
	// OnSLOViolation triggers rollback when SLO violations are detected
	// +optional
	OnSLOViolation bool `json:"onSLOViolation,omitempty" protobuf:"varint,5,opt,name=onSLOViolation"`
	// RollbackWindow defines how far back rollback can go
	// +optional
	RollbackWindow *RollbackWindow `json:"rollbackWindow,omitempty" protobuf:"bytes,6,opt,name=rollbackWindow"`
}

// RollbackWindow defines the rollback window
type RollbackWindow struct {
	// Revisions is the number of revisions to keep for rollback
	Revisions int32 `json:"revisions" protobuf:"varint,1,opt,name=revisions"`
}

// ============================================================================
// Circuit Breaker Types
// ============================================================================

// CircuitBreakerState represents the current state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "Closed"
	CircuitBreakerOpen     CircuitBreakerState = "Open"
	CircuitBreakerHalfOpen CircuitBreakerState = "HalfOpen"
)

// CircuitBreaker defines circuit breaker configuration for an application
type CircuitBreaker struct {
	// Enabled enables the circuit breaker
	Enabled bool `json:"enabled" protobuf:"varint,1,opt,name=enabled"`
	// FailureThreshold is the number of failures before opening the circuit
	FailureThreshold int32 `json:"failureThreshold,omitempty" protobuf:"varint,2,opt,name=failureThreshold"`
	// SuccessThreshold is the number of successes to close the circuit
	SuccessThreshold int32 `json:"successThreshold,omitempty" protobuf:"varint,3,opt,name=successThreshold"`
	// Timeout is the duration the circuit stays open before testing
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty" protobuf:"bytes,4,opt,name=timeout"`
	// HalfOpenMaxRequests is the max requests allowed in half-open state
	HalfOpenMaxRequests int32 `json:"halfOpenMaxRequests,omitempty" protobuf:"varint,5,opt,name=halfOpenMaxRequests"`
	// MonitorInterval is the interval for checking circuit breaker conditions
	// +optional
	MonitorInterval *metav1.Duration `json:"monitorInterval,omitempty" protobuf:"bytes,6,opt,name=monitorInterval"`
}

// CircuitBreakerStatus represents the current status of the circuit breaker
type CircuitBreakerStatus struct {
	// State is the current state of the circuit breaker
	State CircuitBreakerState `json:"state" protobuf:"bytes,1,opt,name=state"`
	// FailureCount is the current number of consecutive failures
	FailureCount int32 `json:"failureCount" protobuf:"varint,2,opt,name=failureCount"`
	// SuccessCount is the current number of consecutive successes (in half-open)
	SuccessCount int32 `json:"successCount" protobuf:"varint,3,opt,name=successCount"`
	// LastStateChange is when the circuit breaker last changed state
	LastStateChange *metav1.Time `json:"lastStateChange,omitempty" protobuf:"bytes,4,opt,name=lastStateChange"`
	// LastFailure is the time of the last failure
	LastFailure *metav1.Time `json:"lastFailure,omitempty" protobuf:"bytes,5,opt,name=lastFailure"`
}

// ============================================================================
// Rate Limiting Types
// ============================================================================

// RateLimiting defines rate limiting configuration
type RateLimiting struct {
	// Enabled enables rate limiting for sync operations
	Enabled bool `json:"enabled" protobuf:"varint,1,opt,name=enabled"`
	// MaxSyncsPerHour limits the number of sync operations per hour
	MaxSyncsPerHour int32 `json:"maxSyncsPerHour,omitempty" protobuf:"varint,2,opt,name=maxSyncsPerHour"`
	// MaxParallelSyncs limits concurrent sync operations
	MaxParallelSyncs int32 `json:"maxParallelSyncs,omitempty" protobuf:"varint,3,opt,name=maxParallelSyncs"`
	// CooldownPeriod is the minimum time between syncs
	// +optional
	CooldownPeriod *metav1.Duration `json:"cooldownPeriod,omitempty" protobuf:"bytes,4,opt,name=cooldownPeriod"`
	// BurstLimit allows temporary burst of syncs
	BurstLimit int32 `json:"burstLimit,omitempty" protobuf:"varint,5,opt,name=burstLimit"`
}

// ============================================================================
// SLO/SLI Types
// ============================================================================

// SLOSpec defines Service Level Objective configuration
type SLOSpec struct {
	// Name is the name of the SLO
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Description is a human-readable description
	Description string `json:"description,omitempty" protobuf:"bytes,2,opt,name=description"`
	// Target is the SLO target percentage (e.g., 99.9)
	Target float64 `json:"target" protobuf:"fixed64,3,opt,name=target"`
	// Window is the time window for the SLO measurement
	Window *metav1.Duration `json:"window" protobuf:"bytes,4,opt,name=window"`
	// SLI defines the Service Level Indicator for this SLO
	SLI SLISpec `json:"sli" protobuf:"bytes,5,opt,name=sli"`
	// AlertPolicy defines alerting for SLO violations
	// +optional
	AlertPolicy *SLOAlertPolicy `json:"alertPolicy,omitempty" protobuf:"bytes,6,opt,name=alertPolicy"`
}

// SLISpec defines a Service Level Indicator
type SLISpec struct {
	// Type is the SLI type (Availability, Latency, ErrorRate, Throughput)
	Type SLIType `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Metric defines the metric provider and query for this SLI
	Metric AnalysisMetric `json:"metric" protobuf:"bytes,2,opt,name=metric"`
	// ThresholdValue is the threshold value for the SLI
	ThresholdValue string `json:"thresholdValue,omitempty" protobuf:"bytes,3,opt,name=thresholdValue"`
}

// SLIType defines the type of SLI
type SLIType string

const (
	SLIAvailability SLIType = "Availability"
	SLILatency      SLIType = "Latency"
	SLIErrorRate    SLIType = "ErrorRate"
	SLIThroughput   SLIType = "Throughput"
	SLISaturation   SLIType = "Saturation"
)

// SLOAlertPolicy defines alerting for SLO violations
type SLOAlertPolicy struct {
	// BurnRateThreshold is the burn rate that triggers an alert
	BurnRateThreshold float64 `json:"burnRateThreshold,omitempty" protobuf:"fixed64,1,opt,name=burnRateThreshold"`
	// NotificationTargets defines where to send alerts
	NotificationTargets []NotificationTarget `json:"notificationTargets,omitempty" protobuf:"bytes,2,rep,name=notificationTargets"`
}

// NotificationTarget defines where to send alerts
type NotificationTarget struct {
	// Type is the notification type (slack, email, pagerduty, webhook, opsgenie)
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Target is the target identifier (channel, email, service key, URL)
	Target string `json:"target" protobuf:"bytes,2,opt,name=target"`
}

// SLOStatus represents the current status of an SLO
type SLOStatus struct {
	// Name is the SLO name
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// CurrentValue is the current SLI value
	CurrentValue float64 `json:"currentValue" protobuf:"fixed64,2,opt,name=currentValue"`
	// Target is the target value
	Target float64 `json:"target" protobuf:"fixed64,3,opt,name=target"`
	// Compliance indicates if SLO is being met
	Compliance bool `json:"compliance" protobuf:"varint,4,opt,name=compliance"`
	// ErrorBudgetRemaining is the remaining error budget percentage
	ErrorBudgetRemaining float64 `json:"errorBudgetRemaining" protobuf:"fixed64,5,opt,name=errorBudgetRemaining"`
	// BurnRate is the current error budget burn rate
	BurnRate float64 `json:"burnRate" protobuf:"fixed64,6,opt,name=burnRate"`
	// LastMeasured is when the SLO was last measured
	LastMeasured *metav1.Time `json:"lastMeasured,omitempty" protobuf:"bytes,7,opt,name=lastMeasured"`
}

// ============================================================================
// Health Score Types
// ============================================================================

// HealthScore provides a comprehensive health score for an application
type HealthScore struct {
	// Overall is the overall health score (0-100)
	Overall int32 `json:"overall" protobuf:"varint,1,opt,name=overall"`
	// Availability score (0-100)
	Availability int32 `json:"availability" protobuf:"varint,2,opt,name=availability"`
	// Performance score (0-100)
	Performance int32 `json:"performance" protobuf:"varint,3,opt,name=performance"`
	// Reliability score (0-100)
	Reliability int32 `json:"reliability" protobuf:"varint,4,opt,name=reliability"`
	// Security score (0-100)
	Security int32 `json:"security" protobuf:"varint,5,opt,name=security"`
	// ResourceEfficiency score (0-100)
	ResourceEfficiency int32 `json:"resourceEfficiency" protobuf:"varint,6,opt,name=resourceEfficiency"`
	// LastCalculated is when the health score was last calculated
	LastCalculated *metav1.Time `json:"lastCalculated,omitempty" protobuf:"bytes,7,opt,name=lastCalculated"`
	// Trend indicates the trend direction (Improving, Stable, Degrading)
	Trend string `json:"trend,omitempty" protobuf:"bytes,8,opt,name=trend"`
}

// ============================================================================
// Incident Management Types
// ============================================================================

// IncidentPolicy defines incident management configuration
type IncidentPolicy struct {
	// Enabled enables automatic incident creation
	Enabled bool `json:"enabled" protobuf:"varint,1,opt,name=enabled"`
	// Severity defines the default severity level
	Severity IncidentSeverity `json:"severity,omitempty" protobuf:"bytes,2,opt,name=severity"`
	// AutoCreate automatically creates incidents on failures
	AutoCreate bool `json:"autoCreate,omitempty" protobuf:"varint,3,opt,name=autoCreate"`
	// EscalationPolicy defines how incidents are escalated
	// +optional
	EscalationPolicy *EscalationPolicy `json:"escalationPolicy,omitempty" protobuf:"bytes,4,opt,name=escalationPolicy"`
	// NotificationTargets defines where to send incident notifications
	NotificationTargets []NotificationTarget `json:"notificationTargets,omitempty" protobuf:"bytes,5,rep,name=notificationTargets"`
	// Runbooks are references to runbooks for automated remediation
	Runbooks []RunbookRef `json:"runbooks,omitempty" protobuf:"bytes,6,rep,name=runbooks"`
}

// IncidentSeverity defines incident severity levels
type IncidentSeverity string

const (
	SeverityCritical IncidentSeverity = "Critical"
	SeverityHigh     IncidentSeverity = "High"
	SeverityMedium   IncidentSeverity = "Medium"
	SeverityLow      IncidentSeverity = "Low"
)

// EscalationPolicy defines how incidents are escalated
type EscalationPolicy struct {
	// Levels defines escalation levels
	Levels []EscalationLevel `json:"levels,omitempty" protobuf:"bytes,1,rep,name=levels"`
}

// EscalationLevel defines a single escalation level
type EscalationLevel struct {
	// After is the duration after which to escalate
	After *metav1.Duration `json:"after" protobuf:"bytes,1,opt,name=after"`
	// NotificationTargets for this escalation level
	NotificationTargets []NotificationTarget `json:"notificationTargets" protobuf:"bytes,2,rep,name=notificationTargets"`
}

// RunbookRef references a runbook for automation
type RunbookRef struct {
	// Name is the name of the runbook
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// URL is the URL to the runbook documentation
	URL string `json:"url,omitempty" protobuf:"bytes,2,opt,name=url"`
	// AutoExecute enables automatic execution on incident
	AutoExecute bool `json:"autoExecute,omitempty" protobuf:"varint,3,opt,name=autoExecute"`
	// Actions defines automated remediation actions
	Actions []RemediationAction `json:"actions,omitempty" protobuf:"bytes,4,rep,name=actions"`
}

// RemediationAction defines an automated remediation action
type RemediationAction struct {
	// Type is the action type (Rollback, Scale, Restart, Custom)
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Parameters are action-specific parameters
	Parameters map[string]string `json:"parameters,omitempty" protobuf:"bytes,2,opt,name=parameters"`
	// Condition defines when this action should be triggered
	Condition string `json:"condition,omitempty" protobuf:"bytes,3,opt,name=condition"`
}

// IncidentStatus represents the current status of an incident
type IncidentStatus struct {
	// ID is the unique incident identifier
	ID string `json:"id" protobuf:"bytes,1,opt,name=id"`
	// Severity is the incident severity
	Severity IncidentSeverity `json:"severity" protobuf:"bytes,2,opt,name=severity"`
	// State is the incident state (Open, Acknowledged, Resolved)
	State string `json:"state" protobuf:"bytes,3,opt,name=state"`
	// Summary describes the incident
	Summary string `json:"summary" protobuf:"bytes,4,opt,name=summary"`
	// CreatedAt is when the incident was created
	CreatedAt *metav1.Time `json:"createdAt,omitempty" protobuf:"bytes,5,opt,name=createdAt"`
	// ResolvedAt is when the incident was resolved
	ResolvedAt *metav1.Time `json:"resolvedAt,omitempty" protobuf:"bytes,6,opt,name=resolvedAt"`
	// AffectedResources lists affected resources
	AffectedResources []string `json:"affectedResources,omitempty" protobuf:"bytes,7,rep,name=affectedResources"`
	// RemediationActions taken
	RemediationActions []string `json:"remediationActions,omitempty" protobuf:"bytes,8,rep,name=remediationActions"`
}

// ============================================================================
// Chaos Engineering Types
// ============================================================================

// ChaosPolicy defines chaos engineering configuration
type ChaosPolicy struct {
	// Enabled enables chaos engineering experiments
	Enabled bool `json:"enabled" protobuf:"varint,1,opt,name=enabled"`
	// Experiments defines chaos experiments to run
	Experiments []ChaosExperiment `json:"experiments,omitempty" protobuf:"bytes,2,rep,name=experiments"`
	// Schedule is a cron schedule for running experiments
	Schedule string `json:"schedule,omitempty" protobuf:"bytes,3,opt,name=schedule"`
	// DryRun only simulates experiments without affecting resources
	DryRun bool `json:"dryRun,omitempty" protobuf:"varint,4,opt,name=dryRun"`
}

// ChaosExperiment defines a chaos engineering experiment
type ChaosExperiment struct {
	// Name is the experiment name
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Type is the experiment type (PodKill, NetworkDelay, CPUStress, MemoryStress, DiskFill)
	Type string `json:"type" protobuf:"bytes,2,opt,name=type"`
	// Duration is how long the experiment runs
	Duration *metav1.Duration `json:"duration,omitempty" protobuf:"bytes,3,opt,name=duration"`
	// Target defines which resources to target
	Target ChaosTarget `json:"target" protobuf:"bytes,4,opt,name=target"`
	// Parameters are experiment-specific parameters
	Parameters map[string]string `json:"parameters,omitempty" protobuf:"bytes,5,opt,name=parameters"`
}

// ChaosTarget defines the target of a chaos experiment
type ChaosTarget struct {
	// Kind is the Kubernetes resource kind to target
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	// Labels to match target resources
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,2,opt,name=labels"`
	// Percentage of matching resources to affect (0-100)
	Percentage int32 `json:"percentage,omitempty" protobuf:"varint,3,opt,name=percentage"`
}

// ============================================================================
// Composite SRE Config
// ============================================================================

// SREConfig is the top-level SRE configuration for an application
type SREConfig struct {
	// DeploymentStrategy defines the deployment strategy (default: Canary)
	// +optional
	DeploymentStrategy *DeploymentStrategy `json:"deploymentStrategy,omitempty" protobuf:"bytes,1,opt,name=deploymentStrategy"`
	// CircuitBreaker defines circuit breaker configuration
	// +optional
	CircuitBreaker *CircuitBreaker `json:"circuitBreaker,omitempty" protobuf:"bytes,2,opt,name=circuitBreaker"`
	// RateLimiting defines rate limiting configuration
	// +optional
	RateLimiting *RateLimiting `json:"rateLimiting,omitempty" protobuf:"bytes,3,opt,name=rateLimiting"`
	// SLOs defines Service Level Objectives
	// +optional
	SLOs []SLOSpec `json:"slos,omitempty" protobuf:"bytes,4,rep,name=slos"`
	// IncidentPolicy defines incident management configuration
	// +optional
	IncidentPolicy *IncidentPolicy `json:"incidentPolicy,omitempty" protobuf:"bytes,5,opt,name=incidentPolicy"`
	// ChaosPolicy defines chaos engineering configuration
	// +optional
	ChaosPolicy *ChaosPolicy `json:"chaosPolicy,omitempty" protobuf:"bytes,6,opt,name=chaosPolicy"`
}

// SREStatus is the top-level SRE status for an application
type SREStatus struct {
	// DeploymentStatus contains canary/blue-green deployment status
	DeploymentStatus *DeploymentStatus `json:"deploymentStatus,omitempty" protobuf:"bytes,1,opt,name=deploymentStatus"`
	// CircuitBreaker contains circuit breaker status
	CircuitBreaker *CircuitBreakerStatus `json:"circuitBreaker,omitempty" protobuf:"bytes,2,opt,name=circuitBreaker"`
	// SLOs contains current SLO statuses
	SLOs []SLOStatus `json:"slos,omitempty" protobuf:"bytes,3,rep,name=slos"`
	// HealthScore contains the current health score
	HealthScore *HealthScore `json:"healthScore,omitempty" protobuf:"bytes,4,opt,name=healthScore"`
	// Incidents contains current active incidents
	Incidents []IncidentStatus `json:"incidents,omitempty" protobuf:"bytes,5,rep,name=incidents"`
	// LastUpdated is when the SRE status was last updated
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty" protobuf:"bytes,6,opt,name=lastUpdated"`
}

// DeploymentPhase represents the phase of a deployment
type DeploymentPhase string

const (
	DeploymentPhasePending    DeploymentPhase = "Pending"
	DeploymentPhaseProgressing DeploymentPhase = "Progressing"
	DeploymentPhasePaused     DeploymentPhase = "Paused"
	DeploymentPhasePromoting  DeploymentPhase = "Promoting"
	DeploymentPhaseCompleted  DeploymentPhase = "Completed"
	DeploymentPhaseFailed     DeploymentPhase = "Failed"
	DeploymentPhaseRollingBack DeploymentPhase = "RollingBack"
)

// DeploymentStatus contains the status of a canary/blue-green deployment
type DeploymentStatus struct {
	// Phase is the current deployment phase
	Phase DeploymentPhase `json:"phase" protobuf:"bytes,1,opt,name=phase"`
	// CurrentStep is the current step index (for canary)
	CurrentStep int32 `json:"currentStep" protobuf:"varint,2,opt,name=currentStep"`
	// TotalSteps is the total number of steps
	TotalSteps int32 `json:"totalSteps" protobuf:"varint,3,opt,name=totalSteps"`
	// CanaryWeight is the current canary traffic weight
	CanaryWeight int32 `json:"canaryWeight" protobuf:"varint,4,opt,name=canaryWeight"`
	// StableRevision is the revision of the stable version
	StableRevision string `json:"stableRevision,omitempty" protobuf:"bytes,5,opt,name=stableRevision"`
	// CanaryRevision is the revision of the canary version
	CanaryRevision string `json:"canaryRevision,omitempty" protobuf:"bytes,6,opt,name=canaryRevision"`
	// StartedAt is when the deployment started
	StartedAt *metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,7,opt,name=startedAt"`
	// CompletedAt is when the deployment completed
	CompletedAt *metav1.Time `json:"completedAt,omitempty" protobuf:"bytes,8,opt,name=completedAt"`
	// AnalysisResults contains the results of analysis runs
	AnalysisResults []AnalysisResult `json:"analysisResults,omitempty" protobuf:"bytes,9,rep,name=analysisResults"`
	// Message is a human-readable message about the current status
	Message string `json:"message,omitempty" protobuf:"bytes,10,opt,name=message"`
}

// AnalysisResult contains the result of a single analysis run
type AnalysisResult struct {
	// MetricName is the name of the metric
	MetricName string `json:"metricName" protobuf:"bytes,1,opt,name=metricName"`
	// Value is the metric value
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
	// Status is the analysis status (Successful, Failed, Inconclusive)
	Status string `json:"status" protobuf:"bytes,3,opt,name=status"`
	// MeasuredAt is when the analysis was performed
	MeasuredAt *metav1.Time `json:"measuredAt,omitempty" protobuf:"bytes,4,opt,name=measuredAt"`
	// Message provides details about the analysis result
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// DefaultSREConfig returns a default SRE configuration with canary deployment
func DefaultSREConfig() *SREConfig {
	weight10 := int32(10)
	weight30 := int32(30)
	weight50 := int32(50)
	weight80 := int32(80)
	weight100 := int32(100)
	failureLimit := int32(3)

	return &SREConfig{
		DeploymentStrategy: &DeploymentStrategy{
			Type: DeploymentStrategyCanary,
			Canary: &CanaryStrategy{
				MaxWeight: 100,
				Steps: []CanaryStep{
					{SetWeight: &weight10},
					{Pause: &CanaryPause{}}, // Manual gate at 10%
					{SetWeight: &weight30},
					{Pause: &CanaryPause{}}, // Manual gate at 30%
					{SetWeight: &weight50},
					{Pause: &CanaryPause{}}, // Manual gate at 50%
					{SetWeight: &weight80},
					{Pause: &CanaryPause{}}, // Manual gate at 80%
					{SetWeight: &weight100},
				},
				AutoRollback: &AutoRollback{
					Enabled:             true,
					OnFailure:           true,
					OnError:             true,
					OnHealthCheckFailed: true,
					OnSLOViolation:      true,
				},
			},
		},
		CircuitBreaker: &CircuitBreaker{
			Enabled:             true,
			FailureThreshold:    5,
			SuccessThreshold:    3,
			HalfOpenMaxRequests: 1,
		},
		RateLimiting: &RateLimiting{
			Enabled:          true,
			MaxSyncsPerHour:  10,
			MaxParallelSyncs: 2,
			BurstLimit:       3,
		},
		IncidentPolicy: &IncidentPolicy{
			Enabled:    true,
			Severity:   SeverityHigh,
			AutoCreate: true,
		},
		SLOs: []SLOSpec{
			{
				Name:        "availability",
				Description: "Application availability SLO",
				Target:      99.9,
				SLI: SLISpec{
					Type: SLIAvailability,
					Metric: AnalysisMetric{
						Name:             "availability",
						SuccessCondition: "result[0] >= 99.9",
						FailureLimit:     &failureLimit,
					},
				},
			},
		},
	}
}
