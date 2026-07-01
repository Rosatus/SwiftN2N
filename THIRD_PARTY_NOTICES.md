# Third-Party Notices

SwiftN2N includes or builds on third-party software. This file is a
human-maintained summary; release archives should include this file together
with `LICENSE`.

## Bundled n2n edge binary

SwiftN2N locates and launches the official n2n v3 `edge` binary. Release
archives may include an `edge` binary under `bin/<os>/<arch>/`.

- Upstream: https://github.com/ntop/n2n
- License: GNU General Public License v3.0
- Linux amd64 source package: `n2n_3.0.0-1038_amd64.deb` from the upstream
  GitHub release tag `3.0`
- Windows/macOS build source: upstream Git tag `3.0`, verified at commit
  `66f557af97b9c2ad42537516101fd04df2639ef0`

SwiftN2N does not modify n2n source code. If you distribute a SwiftN2N archive
that includes `edge`, keep this notice with the archive and provide the
corresponding n2n source under the GPLv3 terms.

## Application framework and libraries

SwiftN2N uses Wails v2, Go modules, React, TypeScript, and Vite. Exact versions
are recorded in:

- `go.mod`
- `go.sum`
- `frontend/package.json`
- `frontend/package-lock.json`

Before publishing a release, review dependency licenses and generated package
metadata. At minimum, run:

```sh
go version
go list -m all
cd frontend && npm ci && npm audit --audit-level=high
node scripts/generate-compliance-reports.mjs
```

Release packages include generated `DEPENDENCY_LICENSES.md` and
`sbom.cdx.json` files. These reports are release aids and should be reviewed
before publishing.
