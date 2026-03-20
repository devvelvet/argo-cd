import * as React from 'react';

import './sre-overview.scss';

// ============================================================================
// Types
// ============================================================================

interface AppSREOverview {
    name: string;
    healthScore: number;
    healthTrend: string;
    deploymentPhase: string;
    canaryWeight: number;
    circuitBreakerState: string;
    sloCompliance: boolean;
    activeIncidents: number;
    availability: number;
    performance: number;
    reliability: number;
}

// ============================================================================
// Score Badge Component
// ============================================================================

const ScoreBadge: React.FC<{score: number; size?: 'sm' | 'md' | 'lg'}> = ({score, size = 'md'}) => {
    const getClass = () => {
        if (score >= 90) return 'sre-score-badge--excellent';
        if (score >= 70) return 'sre-score-badge--good';
        if (score >= 50) return 'sre-score-badge--warning';
        return 'sre-score-badge--critical';
    };

    return <span className={`sre-score-badge sre-score-badge--${size} ${getClass()}`}>{score}</span>;
};

// ============================================================================
// Status Dot
// ============================================================================

const StatusDot: React.FC<{status: string}> = ({status}) => {
    const getClass = () => {
        switch (status) {
            case 'Closed':
            case 'Completed':
            case 'Healthy':
                return 'sre-status-dot--green';
            case 'Open':
            case 'Failed':
            case 'RollingBack':
                return 'sre-status-dot--red';
            case 'HalfOpen':
            case 'Paused':
            case 'Warning':
                return 'sre-status-dot--yellow';
            case 'Progressing':
                return 'sre-status-dot--blue';
            default:
                return 'sre-status-dot--gray';
        }
    };

    return <span className={`sre-status-dot ${getClass()}`} />;
};

// ============================================================================
// Summary Cards
// ============================================================================

interface SummaryCardProps {
    title: string;
    value: string | number;
    subtitle: string;
    color: string;
    icon: string;
}

const SummaryCard: React.FC<SummaryCardProps> = ({title, value, subtitle, color, icon}) => (
    <div className='sre-summary-card' style={{'--accent-color': color} as React.CSSProperties}>
        <div className='sre-summary-card__icon'>{icon}</div>
        <div className='sre-summary-card__content'>
            <span className='sre-summary-card__value'>{value}</span>
            <span className='sre-summary-card__title'>{title}</span>
            <span className='sre-summary-card__subtitle'>{subtitle}</span>
        </div>
    </div>
);

// ============================================================================
// Application Row
// ============================================================================

const AppRow: React.FC<{app: AppSREOverview}> = ({app}) => (
    <tr className='sre-table__row'>
        <td className='sre-table__cell sre-table__cell--name'>
            <StatusDot status={app.healthScore >= 90 ? 'Healthy' : app.healthScore >= 70 ? 'Warning' : 'Failed'} />
            <span>{app.name}</span>
        </td>
        <td className='sre-table__cell sre-table__cell--center'>
            <ScoreBadge score={app.healthScore} size='sm' />
        </td>
        <td className='sre-table__cell'>
            <span className={`sre-trend-mini sre-trend-mini--${app.healthTrend.toLowerCase()}`}>
                {app.healthTrend === 'Improving' ? '↑' : app.healthTrend === 'Degrading' ? '↓' : '→'}
            </span>
        </td>
        <td className='sre-table__cell'>
            <span className={`sre-phase-badge sre-phase-badge--${app.deploymentPhase.toLowerCase()}`}>{app.deploymentPhase}</span>
        </td>
        <td className='sre-table__cell sre-table__cell--center'>
            {app.canaryWeight > 0 ? (
                <div className='sre-mini-progress'>
                    <div className='sre-mini-progress__fill' style={{width: `${app.canaryWeight}%`}} />
                    <span>{app.canaryWeight}%</span>
                </div>
            ) : (
                <span className='sre-text-muted'>—</span>
            )}
        </td>
        <td className='sre-table__cell sre-table__cell--center'>
            <StatusDot status={app.circuitBreakerState} />
            <span className='sre-text-sm'>{app.circuitBreakerState}</span>
        </td>
        <td className='sre-table__cell sre-table__cell--center'>
            <span className={`sre-slo-badge ${app.sloCompliance ? 'sre-slo-badge--pass' : 'sre-slo-badge--fail'}`}>{app.sloCompliance ? 'PASS' : 'FAIL'}</span>
        </td>
        <td className='sre-table__cell sre-table__cell--center'>
            {app.activeIncidents > 0 ? <span className='sre-incident-count'>{app.activeIncidents}</span> : <span className='sre-text-muted'>0</span>}
        </td>
        <td className='sre-table__cell sre-table__cell--scores'>
            <span className='sre-mini-score sre-mini-score--avail'>{app.availability}</span>
            <span className='sre-mini-score sre-mini-score--perf'>{app.performance}</span>
            <span className='sre-mini-score sre-mini-score--rely'>{app.reliability}</span>
        </td>
    </tr>
);

// ============================================================================
// Main SRE Overview Page
// ============================================================================

export const SREOverview: React.FC = () => {
    // Demo data - in production, this comes from the API
    const apps: AppSREOverview[] = [
        {
            name: 'frontend-app',
            healthScore: 95,
            healthTrend: 'Stable',
            deploymentPhase: 'Completed',
            canaryWeight: 0,
            circuitBreakerState: 'Closed',
            sloCompliance: true,
            activeIncidents: 0,
            availability: 99,
            performance: 92,
            reliability: 95
        },
        {
            name: 'api-gateway',
            healthScore: 82,
            healthTrend: 'Improving',
            deploymentPhase: 'Progressing',
            canaryWeight: 50,
            circuitBreakerState: 'Closed',
            sloCompliance: true,
            activeIncidents: 0,
            availability: 95,
            performance: 78,
            reliability: 85
        },
        {
            name: 'payment-service',
            healthScore: 68,
            healthTrend: 'Degrading',
            deploymentPhase: 'Paused',
            canaryWeight: 30,
            circuitBreakerState: 'HalfOpen',
            sloCompliance: false,
            activeIncidents: 2,
            availability: 85,
            performance: 60,
            reliability: 65
        },
        {
            name: 'user-service',
            healthScore: 91,
            healthTrend: 'Stable',
            deploymentPhase: 'Completed',
            canaryWeight: 0,
            circuitBreakerState: 'Closed',
            sloCompliance: true,
            activeIncidents: 0,
            availability: 98,
            performance: 88,
            reliability: 92
        },
        {
            name: 'notification-service',
            healthScore: 45,
            healthTrend: 'Degrading',
            deploymentPhase: 'RollingBack',
            canaryWeight: 10,
            circuitBreakerState: 'Open',
            sloCompliance: false,
            activeIncidents: 3,
            availability: 60,
            performance: 40,
            reliability: 35
        }
    ];

    const totalApps = apps.length;
    const healthyApps = apps.filter(a => a.healthScore >= 90).length;
    const warningApps = apps.filter(a => a.healthScore >= 50 && a.healthScore < 90).length;
    const criticalApps = apps.filter(a => a.healthScore < 50).length;
    const totalIncidents = apps.reduce((sum, a) => sum + a.activeIncidents, 0);
    const sloViolations = apps.filter(a => !a.sloCompliance).length;
    const avgHealth = Math.round(apps.reduce((sum, a) => sum + a.healthScore, 0) / totalApps);

    return (
        <div className='sre-overview'>
            {/* Header */}
            <div className='sre-overview__header'>
                <h1 className='sre-overview__title'>SRE Command Center</h1>
                <div className='sre-overview__subtitle'>Real-time reliability monitoring and deployment management</div>
            </div>

            {/* Summary Cards */}
            <div className='sre-overview__summary'>
                <SummaryCard title='Total Applications' value={totalApps} subtitle='Managed by ArgoCD' color='#4FC3F7' icon='◎' />
                <SummaryCard title='Avg Health Score' value={avgHealth} subtitle={`${healthyApps} healthy`} color='#18BE94' icon='♥' />
                <SummaryCard title='Active Incidents' value={totalIncidents} subtitle={`${criticalApps} critical apps`} color='#E5344A' icon='⚠' />
                <SummaryCard title='SLO Violations' value={sloViolations} subtitle={`of ${totalApps} applications`} color='#F5A623' icon='◉' />
            </div>

            {/* Status Distribution */}
            <div className='sre-overview__distribution'>
                <div className='sre-distribution'>
                    <div className='sre-distribution__bar'>
                        <div className='sre-distribution__segment sre-distribution__segment--healthy' style={{width: `${(healthyApps / totalApps) * 100}%`}}>
                            {healthyApps}
                        </div>
                        <div className='sre-distribution__segment sre-distribution__segment--warning' style={{width: `${(warningApps / totalApps) * 100}%`}}>
                            {warningApps}
                        </div>
                        <div className='sre-distribution__segment sre-distribution__segment--critical' style={{width: `${(criticalApps / totalApps) * 100}%`}}>
                            {criticalApps}
                        </div>
                    </div>
                    <div className='sre-distribution__legend'>
                        <span className='sre-distribution__legend-item sre-distribution__legend-item--healthy'>Healthy (≥90)</span>
                        <span className='sre-distribution__legend-item sre-distribution__legend-item--warning'>Warning (50-89)</span>
                        <span className='sre-distribution__legend-item sre-distribution__legend-item--critical'>Critical (&lt;50)</span>
                    </div>
                </div>
            </div>

            {/* Application Table */}
            <div className='sre-overview__table-container'>
                <table className='sre-table'>
                    <thead>
                        <tr className='sre-table__header-row'>
                            <th className='sre-table__header'>Application</th>
                            <th className='sre-table__header sre-table__header--center'>Health</th>
                            <th className='sre-table__header'>Trend</th>
                            <th className='sre-table__header'>Deploy Phase</th>
                            <th className='sre-table__header sre-table__header--center'>Canary</th>
                            <th className='sre-table__header sre-table__header--center'>Circuit Breaker</th>
                            <th className='sre-table__header sre-table__header--center'>SLO</th>
                            <th className='sre-table__header sre-table__header--center'>Incidents</th>
                            <th className='sre-table__header'>Sub-Scores</th>
                        </tr>
                    </thead>
                    <tbody>
                        {apps.map(app => (
                            <AppRow key={app.name} app={app} />
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default SREOverview;
