import {beforeEach, describe, expect, it, vi} from 'vitest';
import {
    exportableConfig,
    initialConfig,
    loadStoredConfig,
    parseConfigJson,
} from './edgeConfig';

function installStorage() {
    const values = new Map<string, string>();
    const storage = {
        get length() {
            return values.size;
        },
        clear() {
            values.clear();
        },
        getItem(key: string) {
            return values.get(key) ?? null;
        },
        key(index: number) {
            return Array.from(values.keys())[index] ?? null;
        },
        removeItem(key: string) {
            values.delete(key);
        },
        setItem(key: string, value: string) {
            values.set(key, value);
        },
    } as Storage;

    vi.stubGlobal('localStorage', storage);
}

describe('edge config storage', () => {
    beforeEach(() => {
        installStorage();
    });

    it('migrates legacy stored config away from unsafe and noisy defaults', () => {
        localStorage.setItem('swiftn2n:edge-config:v1', JSON.stringify({
            edgePath: '/tmp/custom-edge',
            allowCustomEdgePath: true,
            headerEncryption: true,
            verbose: 3,
            key: 'secret',
            authPassword: 'auth-secret',
            supernodes: ['sn.example.net:7777'],
        }));

        const config = loadStoredConfig();

        expect(config.edgePath).toBe('');
        expect(config.allowCustomEdgePath).toBe(false);
        expect(config.headerEncryption).toBe(false);
        expect(config.verbose).toBe(0);
        expect(config.key).toBe('');
        expect(config.authPassword).toBe('');
        expect(config.supernodes).toEqual(['sn.example.net:7777']);
    });

    it('preserves explicit custom edge path only in v2 stored config', () => {
        localStorage.setItem('swiftn2n:edge-config:v2', JSON.stringify({
            ...initialConfig,
            edgePath: '/opt/n2n/edge',
            allowCustomEdgePath: true,
            headerEncryption: true,
            verbose: 2,
            key: 'secret',
            authPassword: 'auth-secret',
        }));

        const config = loadStoredConfig();

        expect(config.edgePath).toBe('/opt/n2n/edge');
        expect(config.allowCustomEdgePath).toBe(true);
        expect(config.headerEncryption).toBe(true);
        expect(config.verbose).toBe(2);
        expect(config.key).toBe('');
        expect(config.authPassword).toBe('');
    });

    it('exports profiles without secrets and without inactive custom edge path', () => {
        const exported = exportableConfig({
            ...initialConfig,
            edgePath: '/tmp/custom-edge',
            allowCustomEdgePath: false,
            key: 'secret',
            authPassword: 'auth-secret',
        }) as Record<string, unknown>;

        expect(exported.schemaVersion).toBe(2);
        expect(exported.edgePath).toBe('');
        expect(exported.key).toBeUndefined();
        expect(exported.authPassword).toBeUndefined();
    });

    it('imports profiles without carrying executable paths or secrets', () => {
        const config = parseConfigJson(JSON.stringify({
            edgePath: '/tmp/custom-edge',
            allowCustomEdgePath: true,
            key: 'secret',
            authPassword: 'auth-secret',
            supernodes: ['sn.example.net:7777'],
        }));

        expect(config.edgePath).toBe('');
        expect(config.allowCustomEdgePath).toBe(false);
        expect(config.key).toBe('');
        expect(config.authPassword).toBe('');
        expect(config.supernodes).toEqual(['sn.example.net:7777']);
    });
});

