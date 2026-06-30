export type EdgeState = 'idle' | 'starting' | 'connecting' | 'connected' | 'failed' | 'stopping' | string;

export type EdgeConfig = {
    edgePath: string;
    supernodes: string[];
    community: string;
    address: string;
    key: string;
    cipher: string;
    headerEncryption: boolean;
    mac: string;
    deviceName: string;
    mtu: number;
    localPort: string;
    managementPort: number;
    verbose: number;
    authUsername: string;
    authPassword: string;
    federationPublicKey: string;
    supernodeOnly: string;
    compression: string;
    acceptMulticast: boolean;
    enableRouting: boolean;
    routes: string[];
    trafficRules: string[];
    windowsMetric: number;
}

export type EdgeStatus = {
    state: EdgeState;
    message: string;
    pid: number;
    edgePath: string;
    startedAt: string;
    updatedAt: string;
    lastError: string;
}

export type EdgeLogEvent = {
    time: string;
    stream: 'stdout' | 'stderr' | 'system' | string;
    line: string;
}

export type EdgeExitEvent = {
    time: string;
    code: number;
    message: string;
}

export type EnvironmentStatus = {
    os: string;
    arch: string;
    elevated: boolean;
    helperAvailable: boolean;
    edgeFound: boolean;
    edgePath: string;
    edgeVersion: string;
    missing: string[];
    message: string;
}

const storageKey = 'swiftn2n:edge-config:v1';

export const initialStatus: EdgeStatus = {
    state: 'idle',
    message: 'edge is idle',
    pid: 0,
    edgePath: '',
    startedAt: '',
    updatedAt: '',
    lastError: '',
};

export const initialConfig: EdgeConfig = {
    edgePath: '',
    supernodes: ['127.0.0.1:7777'],
    community: 'swift-lan',
    address: '10.10.10.12',
    key: '',
    cipher: 'chacha20',
    headerEncryption: false,
    mac: '',
    deviceName: '',
    mtu: 1290,
    localPort: '',
    managementPort: 0,
    verbose: 0,
    authUsername: '',
    authPassword: '',
    federationPublicKey: '',
    supernodeOnly: '',
    compression: '',
    acceptMulticast: false,
    enableRouting: false,
    routes: [],
    trafficRules: [],
    windowsMetric: 0,
};

export function parseLines(value: string) {
    return value
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter(Boolean);
}

export function formatTime(value: string) {
    if (!value) {
        return '--:--:--';
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return '--:--:--';
    }
    return date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit', second: '2-digit'});
}

export function firstMeaningfulLine(value: string) {
    const line = value
        .split(/\r?\n/)
        .map((item) => item.trim())
        .find(Boolean);
    return line || 'edge responded without version text';
}

export function loadStoredConfig(): EdgeConfig {
    try {
        const raw = localStorage.getItem(storageKey);
        if (!raw) {
            return initialConfig;
        }
        const parsed = JSON.parse(raw) as Partial<EdgeConfig>;
        return {
            ...initialConfig,
            ...parsed,
            key: '',
            authPassword: '',
            supernodes: Array.isArray(parsed.supernodes) ? parsed.supernodes : initialConfig.supernodes,
            routes: Array.isArray(parsed.routes) ? parsed.routes : [],
            trafficRules: Array.isArray(parsed.trafficRules) ? parsed.trafficRules : [],
        };
    } catch {
        return initialConfig;
    }
}

export function saveStoredConfig(config: EdgeConfig) {
    const safeConfig = exportableConfig(config);
    try {
        localStorage.setItem(storageKey, JSON.stringify(safeConfig));
    } catch {
        // Storage can be unavailable in locked-down WebViews; runtime config still works.
    }
}

export function exportableConfig(config: EdgeConfig) {
    const {key: _key, authPassword: _authPassword, ...safeConfig} = config;
    return safeConfig;
}

export function serializeConfig(config: EdgeConfig) {
    return JSON.stringify(exportableConfig(config), null, 2);
}

export function parseConfigJson(value: string): EdgeConfig {
    const parsed = JSON.parse(value) as Partial<EdgeConfig>;
    return {
        ...initialConfig,
        ...parsed,
        key: '',
        authPassword: '',
        supernodes: Array.isArray(parsed.supernodes) ? parsed.supernodes : initialConfig.supernodes,
        routes: Array.isArray(parsed.routes) ? parsed.routes : [],
        trafficRules: Array.isArray(parsed.trafficRules) ? parsed.trafficRules : [],
    };
}

export function appendBoundedLog(current: EdgeLogEvent[], next: EdgeLogEvent) {
    return [...current, next].slice(-1000);
}
