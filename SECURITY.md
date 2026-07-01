# Security Policy

SwiftN2N starts n2n `edge`, which may create network interfaces and requires
administrator or root privileges on common desktop systems. Please treat bug
reports involving privilege handling, profile import, command execution,
secrets, or release artifacts as security-sensitive.

## Supported Versions

Until the first stable release, security fixes are provided for the `main`
branch and the latest tagged release only.

## Reporting a Vulnerability

Please do not open a public issue for a suspected vulnerability. Report it to
the project maintainers privately through the repository security advisory
feature, or by using the contact method published on the repository page.

Include:

- affected operating system and SwiftN2N version or commit
- whether the app was installed from a release archive or built locally
- reproduction steps
- relevant logs with community keys, auth passwords, and private endpoints
  removed

## Privileged Helper Boundary

On Linux, SwiftN2N uses Polkit/`pkexec` to run a hidden helper as root. The
helper is intended to start the bundled `edge` binary, not arbitrary commands.
Custom `edge` binaries are disabled for the privileged helper by default.

For local development only, maintainers may set:

```sh
SWIFTN2N_ALLOW_CUSTOM_EDGE_PATH=1
```

Do not enable that override for normal users or release packages.

