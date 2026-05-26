# DonAgent

DonAgent is a desktop agent written in Go. It will consume RabbitMQ events to show operating system notifications and, in later milestones, execute authorized local actions.

## Development

Go is executed through Docker. A local Go installation is not required.

### Requirements

- Docker
- Docker Compose

### Commands

```powershell
docker compose run --rm dev go version
docker compose run --rm dev go build ./...
docker compose run --rm dev go test ./...
```

### Run the agent

```powershell
docker compose run --rm dev go run ./cmd/donagent
```

By default, the agent reads configuration from:

```text
$HOME/.donagent/config.toml
```

Use `config.example.toml` as the starting point.

On PowerShell:

```powershell
New-Item -ItemType Directory -Force $HOME/.donagent
Copy-Item config.example.toml $HOME/.donagent/config.toml
```

## Current scope

The current implementation starts the project scaffold for Milestone 1.

Included now:

- Go module;
- Docker-based development environment;
- initial `cmd/` and `internal/` structure;
- minimal application entrypoint.
- local config loading with validation and defaults.

Not included yet:

- RabbitMQ connection;
- event parsing;
- desktop notifications;
- GitHub Actions build workflow.
