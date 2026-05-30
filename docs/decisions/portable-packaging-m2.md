# Decision - Initial portable packaging for M2

Status: accepted

## Context

Milestone 2 needs an initial portable distribution path before a native installer exists. The package must include the DonAgent binary, `config.example.toml` and `assets/logo.png`, without changing product code or adding dependencies.

## Decision

Provide script-based portable packaging under `scripts/`:

- `scripts/package-portable.ps1` for PowerShell environments.
- `scripts/package-portable.sh` for Linux or POSIX shell environments.

The scripts cross-compile `windows/amd64` and `linux/amd64` with `CGO_ENABLED=0` and create:

- `dist/donagent-<version>-windows-amd64.zip`
- `dist/donagent-<version>-linux-amd64.tar.gz`

Each archive contains:

- `donagent` or `donagent.exe`
- `config.example.toml`
- `assets/logo.png`

The builder defaults to Docker when the Docker daemon is reachable, falling back to local Go. Docker execution uses `docker compose run --no-deps --rm dev` so packaging does not start RabbitMQ.

## Consequences

- Keeps M2 packaging reversible and independent from installer choices.
- Does not require Go on the host when Docker is available.
- Does not provide install, uninstall, auto-start, signing or auto-update behavior.
- Native runtime validation remains necessary on Windows and Linux.
