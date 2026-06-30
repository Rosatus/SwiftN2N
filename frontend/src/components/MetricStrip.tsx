import {formatTime} from '../lib/edgeConfig';
import type {EdgeConfig, EdgeStatus} from '../lib/edgeConfig';

type MetricStripProps = {
    status: EdgeStatus;
    config: EdgeConfig;
}

export function MetricStrip({status, config}: MetricStripProps) {
    return (
        <div className="details-strip">
            <Metric label="PID" value={status.pid ? String(status.pid) : 'not running'}/>
            <Metric label="cipher" value={config.cipher || 'edge default'}/>
            <Metric label="community" value={config.community || 'unset'}/>
            <Metric label="updated" value={formatTime(status.updatedAt)}/>
        </div>
    );
}

function Metric({label, value}: { label: string; value: string }) {
    return (
        <div className="metric">
            <span>{label}</span>
            <strong>{value}</strong>
        </div>
    );
}
