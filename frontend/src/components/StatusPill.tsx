import type {EdgeStatus} from '../lib/edgeConfig';

export function StatusPill({status}: { status: EdgeStatus }) {
    return (
        <div className={`status-pill ${status.state}`}>
            <span className="status-dot"/>
            <div>
                <strong>{status.state}</strong>
                <span>{status.message}</span>
            </div>
        </div>
    );
}
