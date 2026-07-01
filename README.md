# SwiftN2N

SwiftN2N is a Wails v2 desktop client for running an official n2n v3 `edge`
binary from a React/TypeScript control surface.

The app does not compile or modify n2n. It locates and starts an `edge` binary,
streams stdout/stderr into the UI, maps known log lines into connection status,
and stops the process tree cleanly.

## Project Status

SwiftN2N is early-stage desktop networking software.

| Platform | Status | Notes |
| --- | --- | --- |
| Linux amd64 | Supported for development and release archives | Uses the bundled n2n `edge` and Polkit/`pkexec` for privileged startup. |
| Windows amd64 | Experimental | CI builds an archive and the manifest requests administrator privileges, but TAP/driver and process-tree behavior still need regular real-machine validation. |
| macOS amd64 | Experimental / not yet recommended | CI builds an app bundle, but a native privileged helper flow is not implemented yet. |

## License

SwiftN2N is licensed under the GNU Affero General Public License v3.0 only.
See `LICENSE`.

Release archives may include an upstream n2n `edge` binary licensed under the
GNU General Public License v3.0. See `THIRD_PARTY_NOTICES.md`.

## Stack

- Wails v2
- Go
- React + TypeScript
- Vite
- Plain CSS

## Project Layout

```text
.
├── app.go                         # Wails facade
├── internal/edge                  # n2n edge config, process manager, parser
├── internal/platform              # Windows/Unix process and privilege helpers
├── frontend/src                   # React UI
├── frontend/wailsjs               # Generated Wails bindings
├── build/windows/wails.exe.manifest
└── bin/{windows,linux,darwin}/{amd64,arm64}/edge(.exe)
```

## Edge Binary Resolution

By default, SwiftN2N uses the bundled `edge` binary packaged next to the app:

```text
bin/<goos>/<goarch>/edge
```

During development it also checks the repository-local `bin/` directory.

Advanced users can enable a custom `edge` binary path in the UI. For safety,
custom `edge` paths are disabled for the Linux privileged helper by default; the
helper is intended to start the bundled binary, not arbitrary executables. For
local development only, set:

```sh
SWIFTN2N_ALLOW_CUSTOM_EDGE_PATH=1
```

The legacy resolution order used by lower-level helpers is:

1. GUI `edgePath`
2. `SWIFTN2N_EDGE_PATH`
3. `bin/<goos>/<goarch>/edge`
4. `bin/<goos>/edge`
5. The app executable directory
6. `PATH`

Windows expects `edge.exe`.

For Linux amd64 development builds, download the official ntop/n2n GitHub
release binary with:

```sh
make edge-linux-amd64
```

This downloads `n2n_3.0.0-1038_amd64.deb` from:

```text
https://github.com/ntop/n2n/releases/download/3.0/n2n_3.0.0-1038_amd64.deb
```

The script verifies this SHA256 before extracting `/usr/sbin/edge`:

```text
dd29694cf8c22487618218687c0b81961736ae1afe1c43bcfde6d48f017f16f6
```

The local binary is written to:

```text
bin/linux/amd64/edge
```

`make build-linux` copies that binary to:

```text
build/bin/bin/linux/amd64/edge
```

so the packaged app can locate it relative to the executable.

For Windows and macOS release builds, ntop/n2n v3.0 does not publish official
`edge.exe` or macOS `edge` assets in the GitHub release. The CI workflow builds
`edge` from the upstream `ntop/n2n` `3.0-stable` branch on the native runner and
bundles the resulting binary into the application package.

## Security Notes

- `community` is passed as `-c <community>`.
- `key` is passed through `N2N_KEY`.
- `authPassword` is passed through `N2N_PASSWORD`.
- `key` and `authPassword` are not persisted to local storage.
- Process output and exit messages are redacted before they are emitted to the UI.
- Imported profiles cannot silently enable a custom `edge` binary path.
- Report vulnerabilities privately; see `SECURITY.md`.

## Development

For a detailed handoff document aimed at future development agents, see:

```text
docs/AGENT_HANDOFF.md
```

Install frontend dependencies:

```sh
cd frontend
npm install
```

Run tests and frontend build:

```sh
make test
make frontend-test
go vet ./...
make audit
make compliance
```

Run the Wails app:

```sh
make dev-linux
```

Build:

```sh
make build-linux
```

Create a redistributable Linux archive:

```sh
make release-linux
```

The archive is written to:

```text
dist/SwiftN2N-<version>-linux-amd64.tar.gz
dist/SwiftN2N-<version>-linux-amd64.tar.gz.sha256
```

Prepare the edge binary for the current platform:

```sh
make prepare-edge
```

Build and package the current platform:

```sh
make build-current
make package-current VERSION=dev
```

The GitHub Actions workflow at `.github/workflows/release.yml` builds:

```text
linux/amd64
windows/amd64
darwin/amd64
```

Pushes to `main`/`master` and pull requests upload workflow artifacts. Pushing a
tag like `v0.1.0` also uploads the generated archives to the GitHub Release.
Each package command also writes a `.sha256` file next to the archive.
Release archives include:

```text
LICENSE
THIRD_PARTY_NOTICES.md
DEPENDENCY_LICENSES.md
sbom.cdx.json
```

On Ubuntu/Pop!_OS 24.04, Wails v2 should be built against WebKitGTK 4.1:

```sh
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
make build-linux
```

`make build-linux` passes `-tags webkit2_41`. Wails Doctor may still report
`libwebkit2gtk-4.0-dev` as missing on newer distributions; the actual build is
validated by `pkg-config webkit2gtk-4.1` and the `webkit2_41` tag.

## Windows Privileges

The Windows manifest requests administrator privileges:

```xml
<requestedExecutionLevel level="requireAdministrator" uiAccess="false"/>
```

Linux and macOS do not have an equivalent manifest-based elevation flow in this
app. On Linux, the GUI stays in the normal user session. When `edge` needs root
privileges, SwiftN2N starts a hidden CLI helper through `pkexec`:

```text
SwiftN2N --helper edge
```

The helper reads the edge config from stdin, starts the bundled `edge` as root,
and streams JSON log/exit events back to the GUI. Polkit owns the password
prompt; SwiftN2N never receives or stores the sudo password. If `pkexec` or a
desktop authentication agent is unavailable, the UI reports that a working
helper is missing instead of attempting to run `edge` directly as an ordinary
user.

## UI Notes

`Check Edge` verifies the currently selected edge binary path, executable
permission, privilege state, and edge version output. It does not start a VPN
connection.

The `edge binary` field lives under advanced settings. With the custom-path
toggle off, the displayed path is the resolved bundled binary and the saved
profile keeps `edgePath` empty so packages remain relocatable.

## Contributing

See `CONTRIBUTING.md` for local checks and contribution expectations. Notable
changes should be recorded in `CHANGELOG.md`. Project spaces follow
`CODE_OF_CONDUCT.md`.
