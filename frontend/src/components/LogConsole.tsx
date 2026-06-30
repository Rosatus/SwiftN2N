import type {MutableRefObject} from 'react';
import {formatTime} from '../lib/edgeConfig';
import type {EdgeLogEvent} from '../lib/edgeConfig';

type LogConsoleProps = {
    logs: EdgeLogEvent[];
    autoScroll: boolean;
    onAutoScroll: (value: boolean) => void;
    onClear: () => void;
    terminalRef: MutableRefObject<HTMLDivElement | null>;
}

export function LogConsole({
    logs,
    autoScroll,
    onAutoScroll,
    onClear,
    terminalRef,
}: LogConsoleProps) {
    async function copyLogs() {
        const text = logs
            .map((entry) => `[${entry.time}] ${entry.stream}: ${entry.line}`)
            .join('\n');
        if (!text) {
            return;
        }
        await navigator.clipboard.writeText(text);
    }

    return (
        <section className="terminal-panel">
            <div className="terminal-toolbar">
                <div>
                    <span className="section-kicker">Runtime output</span>
                    <h2>edge terminal</h2>
                </div>
                <div className="terminal-actions">
                    <label className="mini-toggle">
                        <input
                            type="checkbox"
                            checked={autoScroll}
                            onChange={(event) => onAutoScroll(event.target.checked)}
                        />
                        Auto
                    </label>
                    <button type="button" className="ghost-button" onClick={copyLogs} disabled={logs.length === 0}>Copy</button>
                    <button type="button" className="ghost-button" onClick={onClear}>Clear</button>
                </div>
            </div>
            <div className="terminal" ref={terminalRef}>
                {logs.length === 0 ? (
                    <div className="terminal-empty">
                        <span>awaiting edge output</span>
                    </div>
                ) : logs.map((entry, index) => (
                    <div className={`log-line ${entry.stream}`} key={`${entry.time}-${index}`}>
                        <span className="log-time">{formatTime(entry.time)}</span>
                        <span className="log-stream">{entry.stream}</span>
                        <code>{entry.line}</code>
                    </div>
                ))}
            </div>
        </section>
    );
}
