# Decision - Controlled local actions

Status: accepted

## Context

Milestone 2 introduces local actions, but arbitrary commands from RabbitMQ events are a high-risk execution path.

## Decision

Start local actions with semantic `open_url` and `open_app` types. The event declares intent through `type` and `target`; the agent validates the local allowlist before execution.

The `open_url` target must be an `http` or `https` URL, include a host, and must not include embedded credentials. The OS executor uses platform commands directly and does not invoke a shell.

The `open_app` target must be a local alias configured in `app_aliases`. The alias resolves to the executable path locally, not from the queue event. Event-provided arguments are rejected for now.

## Consequences

- Keeps the first action testable without desktop, RabbitMQ or external programs.
- Keeps actions disabled by default through `allowed_actions = []`.
- Defers command aliases, event arguments and domain allowlists to later hardening work.
- Still requires native validation on Windows and Linux before broad operational use.
