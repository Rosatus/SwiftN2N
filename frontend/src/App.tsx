import {useEffect, useMemo, useRef, useState} from 'react';
import {CheckEnvironment, GetEdgeVersion, GetStatus, StartEdge, StopEdge, ValidateConfig} from '../wailsjs/go/main/App';
import {EventsOff, EventsOn} from '../wailsjs/runtime/runtime';
import {ConfigPanel} from './components/ConfigPanel';
import {ConnectionBeam} from './components/ConnectionBeam';
import {LogConsole} from './components/LogConsole';
import {MetricStrip} from './components/MetricStrip';
import {StatusPill} from './components/StatusPill';
import {edgeEvents} from './lib/events';
import {
    appendBoundedLog,
    firstMeaningfulLine,
    initialConfig,
    initialStatus,
    loadStoredConfig,
    parseConfigJson,
    parseLines,
    saveStoredConfig,
    serializeConfig,
} from './lib/edgeConfig';
import type {EdgeConfig, EdgeExitEvent, EdgeLogEvent, EdgeStatus, EnvironmentStatus} from './lib/edgeConfig';

type WailsWindow = Window & {
    go?: {
        main?: {
            App?: unknown;
        };
    };
    runtime?: unknown;
}

function App() {
    const storedConfig = useMemo(() => loadStoredConfig(), []);
    const [config, setConfig] = useState<EdgeConfig>(storedConfig);
    const [supernodesText, setSupernodesText] = useState(storedConfig.supernodes.join('\n'));
    const [routesText, setRoutesText] = useState(storedConfig.routes.join('\n'));
    const [rulesText, setRulesText] = useState(storedConfig.trafficRules.join('\n'));
    const [status, setStatus] = useState<EdgeStatus>(initialStatus);
    const [logs, setLogs] = useState<EdgeLogEvent[]>([]);
    const [environment, setEnvironment] = useState<EnvironmentStatus | null>(null);
    const [advancedOpen, setAdvancedOpen] = useState(false);
    const [pending, setPending] = useState(false);
    const [formError, setFormError] = useState('');
    const [version, setVersion] = useState('');
    const [autoScroll, setAutoScroll] = useState(true);
    const terminalRef = useRef<HTMLDivElement | null>(null);

    const isRunning = status.state === 'starting' || status.state === 'connecting' || status.state === 'connected';
    const isBusy = pending || status.state === 'starting' || status.state === 'stopping';

    const normalizedConfig = useMemo<EdgeConfig>(() => ({
        ...config,
        supernodes: parseLines(supernodesText),
        routes: parseLines(routesText),
        trafficRules: parseLines(rulesText),
    }), [config, routesText, rulesText, supernodesText]);

    useEffect(() => {
        let disposed = false;
        let retryTimer: number | undefined;
        const offCallbacks: Array<() => void> = [];

        function attachBridge() {
            if (disposed) {
                return;
            }
            if (!wailsBridgeReady()) {
                setBridgeUnavailable('Wails bridge is not ready. Open the Wails dev URL or the native app, not the raw Vite page.');
                retryTimer = window.setTimeout(attachBridge, 300);
                return;
            }

            try {
                GetStatus()
                    .then((next) => setStatus(next as EdgeStatus))
                    .catch((error: unknown) => appendLocalLog('stderr', String(error)));
                refreshEnvironment(normalizedConfig);

                const offStatus = EventsOn(edgeEvents.status, (payload: EdgeStatus) => setStatus(payload));
                const offLog = EventsOn(edgeEvents.log, (payload: EdgeLogEvent) => {
                    setLogs((current) => appendBoundedLog(current, payload));
                });
                const offExit = EventsOn(edgeEvents.exit, (payload: EdgeExitEvent) => {
                    setLogs((current) => appendBoundedLog(current, {
                        time: payload.time,
                        stream: 'system',
                        line: `edge exited with code ${payload.code}: ${payload.message}`,
                    }));
                });
                for (const off of [offStatus, offLog, offExit]) {
                    if (typeof off === 'function') {
                        offCallbacks.push(off);
                    }
                }
            } catch (error) {
                setBridgeUnavailable(String(error));
            }
        }

        attachBridge();
        return () => {
            disposed = true;
            if (retryTimer !== undefined) {
                window.clearTimeout(retryTimer);
            }
            for (const off of offCallbacks) {
                off();
            }
            if (offCallbacks.length === 0 && wailsBridgeReady()) {
                try {
                    EventsOff(edgeEvents.status);
                    EventsOff(edgeEvents.log);
                    EventsOff(edgeEvents.exit);
                } catch {
                    // Runtime bridge is already gone.
                }
            }
        };
    }, []);

    useEffect(() => {
        saveStoredConfig(normalizedConfig);
    }, [normalizedConfig]);

    useEffect(() => {
        if (autoScroll && terminalRef.current) {
            terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
        }
    }, [autoScroll, logs]);

    function updateConfig<K extends keyof EdgeConfig>(key: K, value: EdgeConfig[K]) {
        setConfig((current) => ({...current, [key]: value}));
    }

    function appendLocalLog(stream: string, line: string) {
        setLogs((current) => appendBoundedLog(current, {
            time: new Date().toISOString(),
            stream,
            line,
        }));
    }

    function setBridgeUnavailable(message: string) {
        setEnvironment({
            os: 'unknown',
            arch: 'unknown',
            elevated: false,
            helperAvailable: false,
            edgeFound: false,
            edgePath: '',
            edgeVersion: '',
            missing: ['Wails bridge'],
            message,
        });
    }

    async function importConfig(file: File) {
        setFormError('');
        try {
            const text = await file.text();
            const next = parseConfigJson(text);
            setConfig(next);
            setSupernodesText(next.supernodes.join('\n'));
            setRoutesText(next.routes.join('\n'));
            setRulesText(next.trafficRules.join('\n'));
            appendLocalLog('system', `imported profile ${file.name}`);
            await refreshEnvironment(next);
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            setFormError(`Could not import profile: ${message}`);
            appendLocalLog('stderr', `Could not import profile: ${message}`);
        }
    }

    function exportConfig() {
        const blob = new Blob([serializeConfig(normalizedConfig)], {type: 'application/json'});
        const url = URL.createObjectURL(blob);
        const anchor = document.createElement('a');
        anchor.href = url;
        anchor.download = `${normalizedConfig.community || 'swiftn2n'}-profile.json`;
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
        URL.revokeObjectURL(url);
        appendLocalLog('system', 'exported profile without secrets');
    }

    async function refreshEnvironment(nextConfig: EdgeConfig) {
        if (!wailsBridgeReady()) {
            setBridgeUnavailable('Wails bridge is not ready.');
            return;
        }
        try {
            const next = await CheckEnvironment(nextConfig);
            setEnvironment(next as EnvironmentStatus);
        } catch (error) {
            setEnvironment({
                os: 'unknown',
                arch: 'unknown',
                elevated: false,
                helperAvailable: false,
                edgeFound: false,
                edgePath: '',
                edgeVersion: '',
                missing: ['Wails bridge'],
                message: String(error),
            });
        }
    }

    async function toggleConnection() {
        setFormError('');
        setPending(true);
        try {
            ensureWailsBridge();
            if (isRunning) {
                const next = await StopEdge();
                setStatus(next as EdgeStatus);
                await refreshEnvironment(normalizedConfig);
                return;
            }
            await ValidateConfig(normalizedConfig);
            const next = await StartEdge(normalizedConfig);
            setStatus(next as EdgeStatus);
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            setFormError(message);
            appendLocalLog('stderr', message);
        } finally {
            setPending(false);
        }
    }

    async function loadVersion() {
        setVersion('');
        setPending(true);
        try {
            ensureWailsBridge();
            await refreshEnvironment(normalizedConfig);
            const text = await GetEdgeVersion(normalizedConfig);
            setVersion(firstMeaningfulLine(text));
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            setVersion(message);
            appendLocalLog('stderr', message);
        } finally {
            setPending(false);
        }
    }

    return (
        <main className="shell">
            <section className="topbar">
                <div>
                    <p className="eyebrow">n2n v3 desktop edge client</p>
                    <h1>SwiftN2N</h1>
                </div>
                <StatusPill status={status}/>
            </section>

            <section className="dashboard">
                <ConfigPanel
                    config={config}
                    supernodesText={supernodesText}
                    routesText={routesText}
                    rulesText={rulesText}
                    advancedOpen={advancedOpen}
                    pending={pending}
                    isRunning={isRunning}
                    isBusy={isBusy}
                    formError={formError}
                    version={version}
                    environment={environment}
                    onConfigChange={updateConfig}
                    onSupernodesTextChange={setSupernodesText}
                    onRoutesTextChange={setRoutesText}
                    onRulesTextChange={setRulesText}
                    onAdvancedOpenChange={setAdvancedOpen}
                    onCheckEdge={loadVersion}
                    onToggleConnection={toggleConnection}
                    onImportConfig={importConfig}
                    onExportConfig={exportConfig}
                />

                <section className="workbench">
                    <ConnectionBeam
                        status={status}
                        supernode={normalizedConfig.supernodes[0] || 'supernode'}
                        address={normalizedConfig.address || initialConfig.address}
                    />
                    <MetricStrip status={status} config={normalizedConfig}/>
                    <LogConsole
                        logs={logs}
                        autoScroll={autoScroll}
                        onAutoScroll={setAutoScroll}
                        onClear={() => setLogs([])}
                        terminalRef={terminalRef}
                    />
                </section>
            </section>
        </main>
    );
}

function wailsBridgeReady() {
    const candidate = window as WailsWindow;
    return Boolean(candidate.runtime && candidate.go?.main?.App);
}

function ensureWailsBridge() {
    if (!wailsBridgeReady()) {
        throw new Error('Wails bridge is not ready. Use the native app or http://localhost:34115 during development.');
    }
}

export default App;
