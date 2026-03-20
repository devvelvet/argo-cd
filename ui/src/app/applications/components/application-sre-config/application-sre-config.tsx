import * as React from 'react';

import './application-sre-config.scss';

// ============================================================================
// Types
// ============================================================================

export interface SREConfigData {
    deploymentStrategy?: {
        type?: string;
        canarySteps?: Array<{weight: number; pauseDuration?: string}>;
        autoRollbackOnFailure?: boolean;
        autoRollbackOnSLOViolation?: boolean;
    };
    circuitBreaker?: {
        enabled?: boolean;
        failureThreshold?: number;
        successThreshold?: number;
        timeoutSeconds?: number;
    };
    rateLimiting?: {
        enabled?: boolean;
        maxSyncsPerHour?: number;
        maxParallelSyncs?: number;
    };
    slos?: Array<{
        name?: string;
        target?: string;
        type?: string;
    }>;
    incidentPolicy?: {
        enabled?: boolean;
        autoCreate?: boolean;
        defaultSeverity?: string;
    };
}

interface SREConfigPanelProps {
    config: SREConfigData;
    onChange: (config: SREConfigData) => void;
    readOnly?: boolean;
}

// ============================================================================
// Deployment Strategy Config
// ============================================================================

const DeploymentStrategyConfig: React.FC<{
    strategy: SREConfigData['deploymentStrategy'];
    onChange: (strategy: SREConfigData['deploymentStrategy']) => void;
    readOnly: boolean;
}> = ({strategy, onChange, readOnly}) => {
    const strategyType = strategy?.type || 'Canary';
    const steps = strategy?.canarySteps || [{weight: 10}, {weight: 30}, {weight: 50}, {weight: 80}, {weight: 100}];

    return (
        <div className='sre-config-section'>
            <h3 className='sre-config-section__title'>Deployment Strategy</h3>

            <div className='sre-config-field'>
                <label className='sre-config-field__label'>Strategy Type</label>
                <select className='sre-config-field__select' value={strategyType} disabled={readOnly} onChange={e => onChange({...strategy, type: e.target.value})}>
                    <option value='Canary'>Canary (Default)</option>
                    <option value='BlueGreen'>Blue-Green</option>
                    <option value='Rolling'>Rolling Update</option>
                    <option value='ABTesting'>A/B Testing</option>
                </select>
            </div>

            {strategyType === 'Canary' && (
                <>
                    <div className='sre-config-field'>
                        <label className='sre-config-field__label'>Canary Steps</label>
                        <div className='sre-config-steps'>
                            {steps.map((step, i) => (
                                <div key={i} className='sre-config-step'>
                                    <span className='sre-config-step__number'>{i + 1}</span>
                                    <input
                                        type='number'
                                        className='sre-config-field__input sre-config-field__input--small'
                                        value={step.weight}
                                        min={0}
                                        max={100}
                                        disabled={readOnly}
                                        onChange={e => {
                                            const newSteps = [...steps];
                                            newSteps[i] = {...newSteps[i], weight: parseInt(e.target.value) || 0};
                                            onChange({...strategy, canarySteps: newSteps});
                                        }}
                                    />
                                    <span className='sre-config-step__unit'>%</span>
                                    <input
                                        type='text'
                                        className='sre-config-field__input sre-config-field__input--medium'
                                        value={step.pauseDuration || ''}
                                        placeholder='Manual gate'
                                        disabled={readOnly}
                                        onChange={e => {
                                            const newSteps = [...steps];
                                            newSteps[i] = {...newSteps[i], pauseDuration: e.target.value};
                                            onChange({...strategy, canarySteps: newSteps});
                                        }}
                                    />
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className='sre-config-field sre-config-field--inline'>
                        <label className='sre-config-toggle'>
                            <input
                                type='checkbox'
                                checked={strategy?.autoRollbackOnFailure !== false}
                                disabled={readOnly}
                                onChange={e => onChange({...strategy, autoRollbackOnFailure: e.target.checked})}
                            />
                            <span className='sre-config-toggle__slider' />
                            <span className='sre-config-toggle__label'>Auto-rollback on failure</span>
                        </label>
                    </div>

                    <div className='sre-config-field sre-config-field--inline'>
                        <label className='sre-config-toggle'>
                            <input
                                type='checkbox'
                                checked={strategy?.autoRollbackOnSLOViolation || false}
                                disabled={readOnly}
                                onChange={e => onChange({...strategy, autoRollbackOnSLOViolation: e.target.checked})}
                            />
                            <span className='sre-config-toggle__slider' />
                            <span className='sre-config-toggle__label'>Auto-rollback on SLO violation</span>
                        </label>
                    </div>
                </>
            )}
        </div>
    );
};

// ============================================================================
// Circuit Breaker Config
// ============================================================================

const CircuitBreakerConfig: React.FC<{
    config: SREConfigData['circuitBreaker'];
    onChange: (config: SREConfigData['circuitBreaker']) => void;
    readOnly: boolean;
}> = ({config, onChange, readOnly}) => (
    <div className='sre-config-section'>
        <h3 className='sre-config-section__title'>
            Circuit Breaker
            <label className='sre-config-toggle sre-config-toggle--header'>
                <input type='checkbox' checked={config?.enabled || false} disabled={readOnly} onChange={e => onChange({...config, enabled: e.target.checked})} />
                <span className='sre-config-toggle__slider' />
            </label>
        </h3>

        {config?.enabled && (
            <div className='sre-config-grid'>
                <div className='sre-config-field'>
                    <label className='sre-config-field__label'>Failure Threshold</label>
                    <input
                        type='number'
                        className='sre-config-field__input'
                        value={config.failureThreshold || 5}
                        min={1}
                        disabled={readOnly}
                        onChange={e => onChange({...config, failureThreshold: parseInt(e.target.value) || 5})}
                    />
                </div>
                <div className='sre-config-field'>
                    <label className='sre-config-field__label'>Success Threshold</label>
                    <input
                        type='number'
                        className='sre-config-field__input'
                        value={config.successThreshold || 3}
                        min={1}
                        disabled={readOnly}
                        onChange={e => onChange({...config, successThreshold: parseInt(e.target.value) || 3})}
                    />
                </div>
                <div className='sre-config-field'>
                    <label className='sre-config-field__label'>Timeout (seconds)</label>
                    <input
                        type='number'
                        className='sre-config-field__input'
                        value={config.timeoutSeconds || 60}
                        min={1}
                        disabled={readOnly}
                        onChange={e => onChange({...config, timeoutSeconds: parseInt(e.target.value) || 60})}
                    />
                </div>
            </div>
        )}
    </div>
);

// ============================================================================
// Rate Limiting Config
// ============================================================================

const RateLimitingConfig: React.FC<{
    config: SREConfigData['rateLimiting'];
    onChange: (config: SREConfigData['rateLimiting']) => void;
    readOnly: boolean;
}> = ({config, onChange, readOnly}) => (
    <div className='sre-config-section'>
        <h3 className='sre-config-section__title'>
            Rate Limiting
            <label className='sre-config-toggle sre-config-toggle--header'>
                <input type='checkbox' checked={config?.enabled || false} disabled={readOnly} onChange={e => onChange({...config, enabled: e.target.checked})} />
                <span className='sre-config-toggle__slider' />
            </label>
        </h3>

        {config?.enabled && (
            <div className='sre-config-grid'>
                <div className='sre-config-field'>
                    <label className='sre-config-field__label'>Max Syncs / Hour</label>
                    <input
                        type='number'
                        className='sre-config-field__input'
                        value={config.maxSyncsPerHour || 10}
                        min={1}
                        disabled={readOnly}
                        onChange={e => onChange({...config, maxSyncsPerHour: parseInt(e.target.value) || 10})}
                    />
                </div>
                <div className='sre-config-field'>
                    <label className='sre-config-field__label'>Max Parallel Syncs</label>
                    <input
                        type='number'
                        className='sre-config-field__input'
                        value={config.maxParallelSyncs || 2}
                        min={1}
                        disabled={readOnly}
                        onChange={e => onChange({...config, maxParallelSyncs: parseInt(e.target.value) || 2})}
                    />
                </div>
            </div>
        )}
    </div>
);

// ============================================================================
// Incident Policy Config
// ============================================================================

const IncidentPolicyConfig: React.FC<{
    config: SREConfigData['incidentPolicy'];
    onChange: (config: SREConfigData['incidentPolicy']) => void;
    readOnly: boolean;
}> = ({config, onChange, readOnly}) => (
    <div className='sre-config-section'>
        <h3 className='sre-config-section__title'>
            Incident Management
            <label className='sre-config-toggle sre-config-toggle--header'>
                <input type='checkbox' checked={config?.enabled || false} disabled={readOnly} onChange={e => onChange({...config, enabled: e.target.checked})} />
                <span className='sre-config-toggle__slider' />
            </label>
        </h3>

        {config?.enabled && (
            <div className='sre-config-grid'>
                <div className='sre-config-field sre-config-field--inline'>
                    <label className='sre-config-toggle'>
                        <input type='checkbox' checked={config.autoCreate || false} disabled={readOnly} onChange={e => onChange({...config, autoCreate: e.target.checked})} />
                        <span className='sre-config-toggle__slider' />
                        <span className='sre-config-toggle__label'>Auto-create incidents</span>
                    </label>
                </div>
                <div className='sre-config-field'>
                    <label className='sre-config-field__label'>Default Severity</label>
                    <select
                        className='sre-config-field__select'
                        value={config.defaultSeverity || 'High'}
                        disabled={readOnly}
                        onChange={e => onChange({...config, defaultSeverity: e.target.value})}>
                        <option value='Critical'>Critical</option>
                        <option value='High'>High</option>
                        <option value='Medium'>Medium</option>
                        <option value='Low'>Low</option>
                    </select>
                </div>
            </div>
        )}
    </div>
);

// ============================================================================
// Main SRE Config Panel
// ============================================================================

export const ApplicationSREConfigPanel: React.FC<SREConfigPanelProps> = ({config, onChange, readOnly = false}) => {
    return (
        <div className='sre-config'>
            <div className='sre-config__header'>
                <h2 className='sre-config__title'>SRE Configuration</h2>
                <span className='sre-config__subtitle'>Configure deployment strategy, reliability controls, and incident management</span>
            </div>

            <DeploymentStrategyConfig strategy={config.deploymentStrategy} onChange={deploymentStrategy => onChange({...config, deploymentStrategy})} readOnly={readOnly} />

            <CircuitBreakerConfig config={config.circuitBreaker} onChange={circuitBreaker => onChange({...config, circuitBreaker})} readOnly={readOnly} />

            <RateLimitingConfig config={config.rateLimiting} onChange={rateLimiting => onChange({...config, rateLimiting})} readOnly={readOnly} />

            <IncidentPolicyConfig config={config.incidentPolicy} onChange={incidentPolicy => onChange({...config, incidentPolicy})} readOnly={readOnly} />
        </div>
    );
};

export default ApplicationSREConfigPanel;
