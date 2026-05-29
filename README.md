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

## CI

GitHub Actions runs the M1 validation workflow on pushes and pull requests targeting `main` or `develop`.

The workflow uses `actions/setup-go` with the Go version declared in `go.mod`, then runs:

```text
go mod download
go test ./...
go build ./...
```

The job runs on `ubuntu-latest` and `windows-latest` and does not require RabbitMQ or desktop notification services.

### Run the agent

```powershell
docker compose run --rm dev go run ./cmd/donagent
```

### RabbitMQ local

```powershell
docker compose up -d rabbitmq
```

RabbitMQ management UI:

```text
http://localhost:15672
```

Default credentials for local development:

```text
guest / guest
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

The example `rabbit_url` uses the Compose service name `rabbitmq`, which is correct when running the agent through Docker Compose. Use `amqp://guest:guest@localhost:5672/` only when running the agent directly on the host.

## Current scope

Milestone 1 is implemented as the technical MVP.

Included now:

- Go module;
- Docker-based development environment;
- initial `cmd/` and `internal/` structure;
- minimal application entrypoint.
- local config loading with validation and defaults.
- event parsing contract and tests.
- local RabbitMQ service and initial consumer lifecycle.
- conservative ack/nack handling after event validation and routing.
- notification handler behind an internal notifier interface.
- GitHub Actions build/test workflow for Linux and Windows.

Not included yet:

- native desktop notification library;
- tray/background execution;
- local actions;
- installer;
- TLS enforcement;
- log rotation.
