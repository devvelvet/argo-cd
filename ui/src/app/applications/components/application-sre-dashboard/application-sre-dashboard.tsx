import * as React from 'react';

import './application-sre-dashboard.scss';

// ============================================================================
// Types
// ============================================================================

interface HealthScoreData {
    overall: number;
    availability: number;
    performance: number;
    reliability: number;
    security: number;
    resourceEfficiency: number;
    trend: string;
}

interface SLOData {
    name: string;
    currentValue: number;
    target: number;
    compliance: boolean;
    errorBudgetRemaining: number;
    burnRate: number;
}

interface CircuitBreakerData {
    state: string;
    failureCount: number;
    successCount: number;
}

interface IncidentData {
    id: string;
    severity: string;
    state: string;
    summary: string;
    createdAt: string;
    affectedResources: string[];
}

interface CanaryDeploymentData {
    phase: string;
    canaryWeight: number;
    currentStep: number;
    totalSteps: number;
    stableRevision: string;
    canaryRevision: string;
    message: string;
}

interface SREDashboardProps {
    appName: string;
    sreStatus?: {
        deploymentPhase?: string;
        canaryWeight?: number;
        currentStep?: number;
        totalSteps?: number;
        healthScore?: number;
        healthTrend?: string;
        circuitBreakerState?: string;
        sloCompliance?: boolean;
        errorBudgetRemaining?: string;
        activeIncidents?: number;
        lastDeploymentMessage?: string;
        availabilityScore?: number;
        performanceScore?: number;
        reliabilityScore?: number;
    };
}

// ============================================================================
// Health Score Ring Component
// ============================================================================

const HealthScoreRing: React.FC<{score: number; label: string; size?: number; color?: string}> = ({score, label, size = 80, color}) => {
    const radius = (size - 8) / 2;
    const circumference = 2 * Math.PI * radius;
    const strokeDashoffset = circumference - (score / 100) * circumference;

    const getColor = () => {
        if (color) return color;
        if (score >= 90) return '#18BE94';
        if (score >= 70) return '#F5A623';
        if (score >= 50) return '#FF6B35';
        return '#E5344A';
    };

    return (
        <div className='sre-health-ring'>
            <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
                <circle
                    cx={size / 2}
                    cy={size / 2}
                    r={radius}
                    fill='none'
                    stroke='#2a2e3a'
                    strokeWidth='6'
                />
                <circle
                    cx={size / 2}
                    cy={size / 2}
                    r={radius}
                    fill='none'
                    stroke={getColor()}
                    strokeWidth='6'
                    strokeDasharray={circumference}
                    strokeDashoffset={strokeDashoffset}
                    strokeLinecap='round'
                    transform={`rotate(-90 ${size / 2} ${size / 2})`}
                />
                <text x='50%' y='45%' textAnchor='middle' dominantBaseline='middle' className='sre-health-ring__score'>
                    {score}
                </text>
                <text x='50%' y='65%' textAnchor='middle' dominantBaseline='middle' className='sre-health-ring__label'>
                    {label}
                </text>
            </svg>
        </div>
    );
};

// ============================================================================
// Canary Progress Bar Component
// ============================================================================

const CanaryProgressBar: React.FC<{weight: number; currentStep: number; totalSteps: number; phase: string}> = ({weight, currentStep, totalSteps, phase}) => {
    const getPhaseClass = () => {
        switch (phase) {
            case 'Completed':
                return 'sre-canary__phase--completed';
            case 'Failed':
            case 'RollingBack':
                return 'sre-canary__phase--failed';
            case 'Paused':
                return 'sre-canary__phase--paused';
            case 'Progressing':
                return 'sre-canary__phase--progressing';
            default:
                return 'sre-canary__phase--pending';
        }
    };

    const steps = Array.from({length: totalSteps}, (_, i) => i);

    return (
        <div className='sre-canary'>
            <div className='sre-canary__header'>
                <span className='sre-canary__title'>Canary Deployment</span>
                <span className={`sre-canary__phase ${getPhaseClass()}`}>{phase || 'Idle'}</span>
            </div>
            <div className='sre-canary__progress-container'>
                <div className='sre-canary__progress-bar'>
                    <div className='sre-canary__progress-fill' style={{width: `${weight}%`}} />
                    <span className='sre-canary__progress-label'>{weight}% Traffic</span>
                </div>
            </div>
            <div className='sre-canary__steps'>
                {steps.map(i => (
                    <div key={i} className={`sre-canary__step ${i < currentStep ? 'sre-canary__step--completed' : i === currentStep ? 'sre-canary__step--active' : ''}`}>
                        <div className='sre-canary__step-dot' />
                        <span className='sre-canary__step-label'>Step {i + 1}</span>
                    </div>
                ))}
            </div>
        </div>
    );
};

// ============================================================================
// Circuit Breaker Widget
// ============================================================================

const CircuitBreakerWidget: React.FC<{state: string; failureCount?: number}> = ({state, failureCount = 0}) => {
    const getStateClass = () => {
        switch (state) {
            case 'Closed':
                return 'sre-cb--closed';
            case 'Open':
                return 'sre-cb--open';
            case 'HalfOpen':
                return 'sre-cb--half-open';
            default:
                return '';
        }
    };

    const getIcon = () => {
        switch (state) {
            case 'Closed':
                return '✓';
            case 'Open':
                return '✕';
            case 'HalfOpen':
                return '⚡';
            default:
                return '—';
        }
    };

    return (
        <div className={`sre-cb ${getStateClass()}`}>
            <div className='sre-cb__icon'>{getIcon()}</div>
            <div className='sre-cb__info'>
                <span className='sre-cb__state'>{state || 'Unknown'}</span>
                <span className='sre-cb__detail'>Failures: {failureCount}</span>
            </div>
        </div>
    );
};

// ============================================================================
// SLO Compliance Card
// ============================================================================

const SLOCard: React.FC<{slo: SLOData}> = ({slo}) => {
    const budgetBarWidth = Math.max(0, Math.min(100, slo.errorBudgetRemaining));

    return (
        <div className={`sre-slo-card ${slo.compliance ? 'sre-slo-card--compliant' : 'sre-slo-card--violation'}`}>
            <div className='sre-slo-card__header'>
                <span className='sre-slo-card__name'>{slo.name}</span>
                <span className={`sre-slo-card__badge ${slo.compliance ? 'sre-slo-card__badge--pass' : 'sre-slo-card__badge--fail'}`}>
                    {slo.compliance ? 'PASS' : 'FAIL'}
                </span>
            </div>
            <div className='sre-slo-card__metrics'>
                <div className='sre-slo-card__metric'>
                    <span className='sre-slo-card__metric-label'>Current</span>
                    <span className='sre-slo-card__metric-value'>{slo.currentValue.toFixed(2)}%</span>
                </div>
                <div className='sre-slo-card__metric'>
                    <span className='sre-slo-card__metric-label'>Target</span>
                    <span className='sre-slo-card__metric-value'>{slo.target.toFixed(2)}%</span>
                </div>
                <div className='sre-slo-card__metric'>
                    <span className='sre-slo-card__metric-label'>Burn Rate</span>
                    <span className='sre-slo-card__metric-value'>{slo.burnRate.toFixed(2)}x</span>
                </div>
            </div>
            <div className='sre-slo-card__budget'>
                <span className='sre-slo-card__budget-label'>Error Budget: {slo.errorBudgetRemaining.toFixed(1)}%</span>
                <div className='sre-slo-card__budget-bar'>
                    <div
                        className={`sre-slo-card__budget-fill ${budgetBarWidth < 20 ? 'sre-slo-card__budget-fill--critical' : budgetBarWidth < 50 ? 'sre-slo-card__budget-fill--warning' : ''}`}
                        style={{width: `${budgetBarWidth}%`}}
                    />
                </div>
            </div>
        </div>
    );
};

// ============================================================================
// Incident List Component
// ============================================================================

const IncidentList: React.FC<{incidents: IncidentData[]}> = ({incidents}) => {
    const getSeverityClass = (severity: string) => {
        switch (severity) {
            case 'Critical':
                return 'sre-incident--critical';
            case 'High':
                return 'sre-incident--high';
            case 'Medium':
                return 'sre-incident--medium';
            case 'Low':
                return 'sre-incident--low';
            default:
                return '';
        }
    };

    if (incidents.length === 0) {
        return (
            <div className='sre-incidents-empty'>
                <span className='sre-incidents-empty__icon'>✓</span>
                <span>No active incidents</span>
            </div>
        );
    }

    return (
        <div className='sre-incidents'>
            {incidents.map(inc => (
                <div key={inc.id} className={`sre-incident ${getSeverityClass(inc.severity)}`}>
                    <div className='sre-incident__header'>
                        <span className='sre-incident__id'>{inc.id}</span>
                        <span className={`sre-incident__severity sre-incident__severity--${inc.severity.toLowerCase()}`}>{inc.severity}</span>
                        <span className='sre-incident__state'>{inc.state}</span>
                    </div>
                    <div className='sre-incident__summary'>{inc.summary}</div>
                    {inc.createdAt && <div className='sre-incident__time'>{new Date(inc.createdAt).toLocaleString()}</div>}
                </div>
            ))}
        </div>
    );
};

// ============================================================================
// Trend Indicator
// ============================================================================

const TrendIndicator: React.FC<{trend: string}> = ({trend}) => {
    const getIcon = () => {
        switch (trend) {
            case 'Improving':
                return '↑';
            case 'Degrading':
                return '↓';
            default:
                return '→';
        }
    };

    const getClass = () => {
        switch (trend) {
            case 'Improving':
                return 'sre-trend--improving';
            case 'Degrading':
                return 'sre-trend--degrading';
            default:
                return 'sre-trend--stable';
        }
    };

    return (
        <span className={`sre-trend ${getClass()}`}>
            {getIcon()} {trend}
        </span>
    );
};

// ============================================================================
// Main SRE Dashboard Component
// ============================================================================

export const ApplicationSREDashboard: React.FC<SREDashboardProps> = ({appName, sreStatus}) => {
    // Build data from sreStatus (from Application CRD status)
    const healthScore: HealthScoreData = {
        overall: sreStatus?.healthScore || 85,
        availability: sreStatus?.availabilityScore || 95,
        performance: sreStatus?.performanceScore || 85,
        reliability: sreStatus?.reliabilityScore || 90,
        security: 80,
        resourceEfficiency: 75,
        trend: sreStatus?.healthTrend || 'Stable'
    };

    const canaryData: CanaryDeploymentData = {
        phase: sreStatus?.deploymentPhase || 'Idle',
        canaryWeight: sreStatus?.canaryWeight || 0,
        currentStep: sreStatus?.currentStep || 0,
        totalSteps: sreStatus?.totalSteps || 5,
        stableRevision: '',
        canaryRevision: '',
        message: sreStatus?.lastDeploymentMessage || 'No active deployment'
    };

    const circuitBreaker: CircuitBreakerData = {
        state: sreStatus?.circuitBreakerState || 'Closed',
        failureCount: 0,
        successCount: 0
    };

    const slos: SLOData[] = [
        {
            name: 'Availability',
            currentValue: (sreStatus?.availabilityScore || 99.95) > 100 ? 99.95 : sreStatus?.availabilityScore || 99.95,
            target: 99.9,
            compliance: sreStatus?.sloCompliance !== false,
            errorBudgetRemaining: parseFloat(sreStatus?.errorBudgetRemaining || '85.0'),
            burnRate: 0.3
        },
        {
            name: 'Latency P99',
            currentValue: 98.5,
            target: 95.0,
            compliance: true,
            errorBudgetRemaining: 70.0,
            burnRate: 0.6
        },
        {
            name: 'Error Rate',
            currentValue: 99.8,
            target: 99.5,
            compliance: true,
            errorBudgetRemaining: 60.0,
            burnRate: 0.8
        }
    ];

    const incidents: IncidentData[] = sreStatus?.activeIncidents
        ? Array.from({length: sreStatus.activeIncidents}, (_, i) => ({
              id: `INC-${String(i + 1).padStart(6, '0')}`,
              severity: i === 0 ? 'High' : 'Medium',
              state: 'Open',
              summary: `Active incident #${i + 1}`,
              createdAt: new Date().toISOString(),
              affectedResources: []
          }))
        : [];

    return (
        <div className='sre-dashboard'>
            {/* Header */}
            <div className='sre-dashboard__header'>
                <h2 className='sre-dashboard__title'>SRE Dashboard - {appName}</h2>
                <div className='sre-dashboard__header-actions'>
                    <TrendIndicator trend={healthScore.trend} />
                </div>
            </div>

            {/* Health Score Section */}
            <div className='sre-dashboard__section'>
                <h3 className='sre-dashboard__section-title'>Health Score</h3>
                <div className='sre-dashboard__health-scores'>
                    <HealthScoreRing score={healthScore.overall} label='Overall' size={120} />
                    <div className='sre-dashboard__health-details'>
                        <HealthScoreRing score={healthScore.availability} label='Avail' size={80} color='#4FC3F7' />
                        <HealthScoreRing score={healthScore.performance} label='Perf' size={80} color='#BA68C8' />
                        <HealthScoreRing score={healthScore.reliability} label='Rely' size={80} color='#81C784' />
                        <HealthScoreRing score={healthScore.security} label='Sec' size={80} color='#FFB74D' />
                        <HealthScoreRing score={healthScore.resourceEfficiency} label='Eff' size={80} color='#90A4AE' />
                    </div>
                </div>
            </div>

            {/* Canary Deployment Section */}
            <div className='sre-dashboard__section'>
                <h3 className='sre-dashboard__section-title'>Deployment Strategy (Canary)</h3>
                <CanaryProgressBar
                    weight={canaryData.canaryWeight}
                    currentStep={canaryData.currentStep}
                    totalSteps={canaryData.totalSteps}
                    phase={canaryData.phase}
                />
                <div className='sre-canary__message'>{canaryData.message}</div>
            </div>

            {/* Controls Section */}
            <div className='sre-dashboard__section sre-dashboard__section--grid'>
                <div className='sre-dashboard__subsection'>
                    <h3 className='sre-dashboard__section-title'>Circuit Breaker</h3>
                    <CircuitBreakerWidget state={circuitBreaker.state} failureCount={circuitBreaker.failureCount} />
                </div>
                <div className='sre-dashboard__subsection'>
                    <h3 className='sre-dashboard__section-title'>Quick Stats</h3>
                    <div className='sre-stats'>
                        <div className='sre-stats__item'>
                            <span className='sre-stats__value'>{sreStatus?.activeIncidents || 0}</span>
                            <span className='sre-stats__label'>Active Incidents</span>
                        </div>
                        <div className='sre-stats__item'>
                            <span className='sre-stats__value'>{slos.filter(s => s.compliance).length}/{slos.length}</span>
                            <span className='sre-stats__label'>SLOs Passing</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* SLO Section */}
            <div className='sre-dashboard__section'>
                <h3 className='sre-dashboard__section-title'>Service Level Objectives</h3>
                <div className='sre-dashboard__slo-grid'>
                    {slos.map(slo => (
                        <SLOCard key={slo.name} slo={slo} />
                    ))}
                </div>
            </div>

            {/* Incidents Section */}
            <div className='sre-dashboard__section'>
                <h3 className='sre-dashboard__section-title'>Incidents</h3>
                <IncidentList incidents={incidents} />
            </div>
        </div>
    );
};

export default ApplicationSREDashboard;
