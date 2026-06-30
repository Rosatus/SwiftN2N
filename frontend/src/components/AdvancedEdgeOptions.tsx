import {FormField} from './FormField';
import type {EdgeConfig} from '../lib/edgeConfig';

type AdvancedEdgeOptionsProps = {
    config: EdgeConfig;
    routesText: string;
    rulesText: string;
    onConfigChange: <K extends keyof EdgeConfig>(key: K, value: EdgeConfig[K]) => void;
    onRoutesTextChange: (value: string) => void;
    onRulesTextChange: (value: string) => void;
}

export function AdvancedEdgeOptions({
    config,
    routesText,
    rulesText,
    onConfigChange,
    onRoutesTextChange,
    onRulesTextChange,
}: AdvancedEdgeOptionsProps) {
    return (
        <div className="advanced-panel">
            <div className="field-grid">
                <FormField label="cipher">
                    <select value={config.cipher} onChange={(event) => onConfigChange('cipher', event.target.value)}>
                        <option value="">edge default</option>
                        <option value="none">none</option>
                        <option value="twofish">twofish</option>
                        <option value="aes">AES</option>
                        <option value="chacha20">ChaCha20</option>
                        <option value="speck">Speck</option>
                    </select>
                </FormField>
                <FormField label="MTU">
                    <input
                        type="number"
                        min="0"
                        value={config.mtu}
                        onChange={(event) => onConfigChange('mtu', Number(event.target.value))}
                    />
                </FormField>
            </div>
            <div className="field-grid">
                <FormField label="MAC address">
                    <input
                        value={config.mac}
                        onChange={(event) => onConfigChange('mac', event.target.value)}
                        placeholder="DE:AD:BE:EF:10:12"
                    />
                </FormField>
                <FormField label="TAP name">
                    <input
                        value={config.deviceName}
                        onChange={(event) => onConfigChange('deviceName', event.target.value)}
                        placeholder="swiftn2n0"
                    />
                </FormField>
            </div>
            <div className="field-grid">
                <FormField label="local port">
                    <input
                        value={config.localPort}
                        onChange={(event) => onConfigChange('localPort', event.target.value)}
                        placeholder="0.0.0.0:7777"
                    />
                </FormField>
                <FormField label="management port">
                    <input
                        type="number"
                        min="0"
                        max="65535"
                        value={config.managementPort}
                        onChange={(event) => onConfigChange('managementPort', Number(event.target.value))}
                    />
                </FormField>
            </div>
            <div className="field-grid">
                <FormField label="verbose level">
                    <input
                        type="number"
                        min="0"
                        max="5"
                        value={config.verbose}
                        onChange={(event) => onConfigChange('verbose', Number(event.target.value))}
                    />
                </FormField>
                <FormField label="Windows metric">
                    <input
                        type="number"
                        min="0"
                        value={config.windowsMetric}
                        onChange={(event) => onConfigChange('windowsMetric', Number(event.target.value))}
                    />
                </FormField>
            </div>
            <div className="field-grid">
                <FormField label="auth user">
                    <input
                        value={config.authUsername}
                        onChange={(event) => onConfigChange('authUsername', event.target.value)}
                        placeholder="edge identity"
                    />
                </FormField>
                <FormField label="auth password">
                    <input
                        type="password"
                        value={config.authPassword}
                        onChange={(event) => onConfigChange('authPassword', event.target.value)}
                        placeholder="Optional"
                    />
                </FormField>
            </div>
            <FormField label="federation public key">
                <input
                    value={config.federationPublicKey}
                    onChange={(event) => onConfigChange('federationPublicKey', event.target.value)}
                    placeholder="base64 public key"
                />
            </FormField>
            <div className="field-grid">
                <FormField label="supernode only">
                    <select value={config.supernodeOnly} onChange={(event) => onConfigChange('supernodeOnly', event.target.value)}>
                        <option value="">off</option>
                        <option value="udp">UDP</option>
                        <option value="tcp">TCP</option>
                    </select>
                </FormField>
                <FormField label="compression">
                    <select value={config.compression} onChange={(event) => onConfigChange('compression', event.target.value)}>
                        <option value="">off</option>
                        <option value="lzo">LZO</option>
                        <option value="zstd">Zstandard</option>
                    </select>
                </FormField>
            </div>
            <div className="switch-row">
                <label>
                    <input
                        type="checkbox"
                        checked={config.headerEncryption}
                        onChange={(event) => onConfigChange('headerEncryption', event.target.checked)}
                    />
                    Header encryption
                </label>
                <label>
                    <input
                        type="checkbox"
                        checked={config.enableRouting}
                        onChange={(event) => onConfigChange('enableRouting', event.target.checked)}
                    />
                    Routing
                </label>
                <label>
                    <input
                        type="checkbox"
                        checked={config.acceptMulticast}
                        onChange={(event) => onConfigChange('acceptMulticast', event.target.checked)}
                    />
                    Multicast
                </label>
            </div>
            <FormField label="routes" helper="Format: network/prefix:gateway">
                <textarea
                    rows={3}
                    value={routesText}
                    onChange={(event) => onRoutesTextChange(event.target.value)}
                    placeholder="10.42.0.0/16:10.10.10.1"
                />
            </FormField>
            <FormField label="traffic rules">
                <textarea
                    rows={3}
                    value={rulesText}
                    onChange={(event) => onRulesTextChange(event.target.value)}
                    placeholder="accept ip"
                />
            </FormField>
        </div>
    );
}
