# Contributing

Thanks for helping improve SwiftN2N.

## Development Setup

Install Go, Node.js, npm, Wails v2, and the native dependencies documented in
`README.md`.

Common checks:

```sh
make test
make frontend-test
go vet ./...
make audit
make compliance
```

For Linux development, prepare the bundled edge binary with:

```sh
make edge-linux-amd64
```

## Pull Requests

- Keep changes focused and small enough to review.
- Add or update Go tests when changing command construction, process
  management, log parsing, redaction, or helper behavior.
- Run `gofmt` on Go changes.
- Do not commit local edge binaries, build outputs, private operator notes, or
  secrets.
- Do not put community keys or auth passwords in command-line arguments,
  localStorage, logs, docs, or test fixtures.

## Security-Sensitive Changes

Changes to privileged helper behavior, profile import, custom `edge` paths,
release packaging, signing, or dependency sourcing should explain their threat
model in the PR description.
