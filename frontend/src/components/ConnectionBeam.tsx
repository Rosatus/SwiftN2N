import type {EdgeStatus} from '../lib/edgeConfig';

type ConnectionBeamProps = {
    status: EdgeStatus;
    supernode: string;
    address: string;
}

export function ConnectionBeam({status, supernode, address}: ConnectionBeamProps) {
    const active = status.state === 'connected';
    const pending = status.state === 'starting' || status.state === 'connecting';
    const failed = status.state === 'failed';

    return (
        <section className={`beam-panel ${active ? 'active' : ''} ${pending ? 'pending' : ''} ${failed ? 'failed' : ''}`}>
            <div className="node local-node">
                <span>Local</span>
                <strong>{address || 'auto address'}</strong>
            </div>
            <div className="beam-track" aria-hidden="true">
                <span className="beam-line"/>
                <span className="beam-pulse one"/>
                <span className="beam-pulse two"/>
                <span className="beam-pulse three"/>
            </div>
            <div className="node super-node">
                <span>Supernode</span>
                <strong>{supernode}</strong>
            </div>
            <div className="state-copy">
                <span>state</span>
                <strong>{status.state}</strong>
            </div>
        </section>
    );
}
