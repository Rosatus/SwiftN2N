import {AdvancedEdgeOptions} from './AdvancedEdgeOptions';
import {FormField} from './FormField';
import type {EdgeConfig, EnvironmentStatus} from '../lib/edgeConfig';

type ConfigPanelProps = {
    config: EdgeConfig;
    supernodesText: string;
    routesText: string;
    rulesText: string;
    advancedOpen: boolean;
    pending: boolean;
    isRunning: boolean;
    isBusy: boolean;
    formError: string;
    version: string;
    environment: EnvironmentStatus | null;
    onConfigChange: <K extends keyof EdgeConfig>(key: K, value: EdgeConfig[K]) => void;
    onSupernodesTextChange: (value: string) => void;
    onRoutesTextChange: (value: string) => void;
    onRulesTextChange: (value: string) => void;
    onAdvancedOpenChange: (value: boolean) => void;
    onCheckEdge: () => void;
    onToggleConnection: () => void;
    onImportConfig: (file: File) => void;
    onExportConfig: () => void;
}

export function ConfigPanel({
    config,
    supernodesText,
    routesText,
    rulesText,
    advancedOpen,
    pending,
    isRunning,
    isBusy,
    formError,
    version,
    environment,
    onConfigChange,
    onSupernodesTextChange,
    onRoutesTextChange,
    onRulesTextChange,
    onAdvancedOpenChange,
    onCheckEdge,
    onToggleConnection,
    onImportConfig,
    onExportConfig,
}: ConfigPanelProps) {
    return (
        <aside className="panel config-panel">
            <div className="panel-heading">
                <div>
                    <span className="section-kicker">Configuration</span>
                    <h2>Edge profile</h2>
                </div>
                <button className="ghost-button" type="button" onClick={onCheckEdge} disabled={pending}>
                    Check Edge
                </button>
            </div>

            <EnvironmentPanel environment={environment}/>

            <div className="field-stack">
                <FormField label="edge binary">
                    <input
                        value={config.edgePath}
                        onChange={(event) => onConfigChange('edgePath', event.target.value)}
                        placeholder="Auto: bin/linux/amd64/edge"
                    />
                </FormField>
                <FormField label="supernode">
                    <textarea
                        value={supernodesText}
                        onChange={(event) => onSupernodesTextChange(event.target.value)}
                        rows={3}
                        placeholder="supernode.example.net:7777"
                    />
                </FormField>
                <div className="field-grid">
                    <FormField label="community">
                        <input
                            value={config.community}
                            onChange={(event) => onConfigChange('community', event.target.value)}
                            placeholder="office-lan"
                        />
                    </FormField>
                    <FormField label="fixed IP">
                        <input
                            value={config.address}
                            onChange={(event) => onConfigChange('address', event.target.value)}
                            placeholder="10.10.10.12"
                        />
                    </FormField>
                </div>
                <FormField label="community key" helper="Sent to edge through N2N_KEY, never persisted locally.">
                    <input
                        value={config.key}
                        onChange={(event) => onConfigChange('key', event.target.value)}
                        placeholder="Optional"
                        type="password"
                    />
                </FormField>
            </div>

            <button
                className="advanced-toggle"
                type="button"
                aria-expanded={advancedOpen}
                onClick={() => onAdvancedOpenChange(!advancedOpen)}
            >
                <span>v3 advanced parameters</span>
                <span className="chevron">{advancedOpen ? 'Close' : 'Open'}</span>
            </button>

            {advancedOpen && (
                <AdvancedEdgeOptions
                    config={config}
                    routesText={routesText}
                    rulesText={rulesText}
                    onConfigChange={onConfigChange}
                    onRoutesTextChange={onRoutesTextChange}
                    onRulesTextChange={onRulesTextChange}
                />
            )}

            {formError && <div className="inline-error">{formError}</div>}
            {version && <div className="version-readout">{version}</div>}

            <div className="profile-actions">
                <label className="ghost-button file-button">
                    <input
                        type="file"
                        accept="application/json,.json"
                        onChange={(event) => {
                            const file = event.target.files?.[0];
                            if (file) {
                                onImportConfig(file);
                                event.target.value = '';
                            }
                        }}
                    />
                    Import
                </label>
                <button type="button" className="ghost-button" onClick={onExportConfig}>
                    Export
                </button>
            </div>

            <button
                className={`connect-button ${isRunning ? 'disconnect' : ''}`}
                type="button"
                onClick={onToggleConnection}
                disabled={isBusy}
            >
                <span className="button-orb"/>
                <span>{isRunning ? 'Disconnect' : 'Connect'}</span>
            </button>
        </aside>
    );
}

function EnvironmentPanel({
    environment,
}: {
    environment: EnvironmentStatus | null;
}) {
    if (!environment) {
        return (
            <div className="environment-panel muted">
                <span className="status-dot"/>
                <div>
                    <strong>checking runtime</strong>
                    <span>waiting for Wails bridge</span>
                </div>
            </div>
        );
    }

    const missing = Array.isArray(environment.missing) ? environment.missing : [];
    const healthy = environment.edgeFound && (environment.elevated || environment.helperAvailable) && missing.length === 0;
    return (
        <div className={`environment-panel ${healthy ? 'ready' : 'attention'}`}>
            <span className="status-dot"/>
            <div className="environment-copy">
                <strong>{environment.message}</strong>
                <span>
                    {environment.os}/{environment.arch}
                    {environment.edgePath ? ` - ${environment.edgePath}` : ''}
                </span>
                {!environment.elevated && environment.helperAvailable && <em>Connect will request Polkit authorization for the edge helper.</em>}
                {missing.length > 0 && <em>{missing.join(', ')}</em>}
            </div>
        </div>
    );
}
