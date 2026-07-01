#!/usr/bin/env node
import {spawnSync} from 'node:child_process';
import {randomUUID} from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');
const outDir = path.resolve(process.argv[2] || path.join(repoRoot, 'dist', 'compliance'));

fs.mkdirSync(outDir, {recursive: true});

function run(command, args, cwd = repoRoot) {
  const result = spawnSync(command, args, {
    cwd,
    encoding: 'utf8',
    shell: process.platform === 'win32',
  });
  if (result.error) {
    throw result.error;
  }
  if (result.status !== 0) {
    throw new Error(`${command} ${args.join(' ')} failed:\n${result.stderr || result.stdout}`);
  }
  return result.stdout.trim();
}

function readJSON(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

function goModules() {
  const output = run('go', ['list', '-m', 'all']);
  return output
    .split(/\r?\n/)
    .filter(Boolean)
    .map((line) => {
      const [name, version = ''] = line.split(/\s+/);
      return {name, version};
    });
}

function npmPackages() {
  const lock = readJSON(path.join(repoRoot, 'frontend', 'package-lock.json'));
  return Object.entries(lock.packages || {})
    .filter(([packagePath]) => packagePath.startsWith('node_modules/'))
    .map(([packagePath, meta]) => {
      const name = packagePath.replace(/^node_modules\//, '');
      return {
        name,
        version: meta.version || '',
        license: meta.license || 'UNKNOWN',
        dev: meta.dev === true,
      };
    })
    .sort((a, b) => a.name.localeCompare(b.name));
}

const generatedAt = new Date().toISOString();
const go = goModules();
const npm = npmPackages();

const licenseReport = [
  '# Dependency License Report',
  '',
  `Generated: ${generatedAt}`,
  '',
  'This report is generated from `go list -m all` and `frontend/package-lock.json`.',
  'It is a release aid, not a substitute for legal review.',
  '',
  '## SwiftN2N',
  '',
  '- License: AGPL-3.0-only',
  '',
  '## Bundled n2n edge',
  '',
  '- Upstream: https://github.com/ntop/n2n',
  '- License: GPL-3.0-only',
  '- Source ref: tag 3.0, commit 66f557af97b9c2ad42537516101fd04df2639ef0',
  '',
  '## Go modules',
  '',
  ...go.map((module) => `- ${module.name}${module.version ? ` ${module.version}` : ''}`),
  '',
  '## npm packages',
  '',
  ...npm.map((pkg) => `- ${pkg.name}@${pkg.version} - ${pkg.license}${pkg.dev ? ' (dev)' : ''}`),
  '',
].join('\n');

fs.writeFileSync(path.join(outDir, 'DEPENDENCY_LICENSES.md'), licenseReport);

const components = [
  {
    type: 'application',
    name: 'SwiftN2N',
    version: process.env.VERSION || 'dev',
    licenses: [{license: {id: 'AGPL-3.0-only'}}],
  },
  {
    type: 'application',
    name: 'n2n-edge',
    version: '3.0',
    licenses: [{license: {id: 'GPL-3.0-only'}}],
    externalReferences: [
      {type: 'vcs', url: 'https://github.com/ntop/n2n'},
    ],
  },
  ...go.map((module) => ({
    type: 'library',
    name: module.name,
    version: module.version || undefined,
    purl: module.version ? `pkg:golang/${module.name}@${module.version}` : undefined,
  })),
  ...npm.map((pkg) => ({
    type: 'library',
    name: pkg.name,
    version: pkg.version,
    scope: pkg.dev ? 'optional' : 'required',
    licenses: pkg.license === 'UNKNOWN' ? undefined : [{license: {name: pkg.license}}],
    purl: `pkg:npm/${encodeURIComponent(pkg.name)}@${pkg.version}`,
  })),
];

const bom = {
  bomFormat: 'CycloneDX',
  specVersion: '1.5',
  serialNumber: `urn:uuid:${randomUUID()}`,
  version: 1,
  metadata: {
    timestamp: generatedAt,
    tools: [
      {
        vendor: 'SwiftN2N',
        name: 'generate-compliance-reports',
        version: '1',
      },
    ],
    component: components[0],
  },
  components,
};

fs.writeFileSync(path.join(outDir, 'sbom.cdx.json'), `${JSON.stringify(bom, null, 2)}\n`);
console.log(outDir);

