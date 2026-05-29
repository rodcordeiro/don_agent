# Decision - Controlled open_url action

Status: accepted

## Context

Milestone 2 introduces local actions, but arbitrary commands from RabbitMQ events are a high-risk execution path.

## Decision

Start local actions with a semantic `open_url` type only. The event declares intent through `type` and `target`; the agent validates the local allowlist before execution.

The `open_url` target must be an `http` or `https` URL, include a host, and must not include embedded credentials. The OS executor uses platform commands directly and does not invoke a shell.

## Consequences

- Keeps the first action testable without desktop, RabbitMQ or external programs.
- Keeps actions disabled by default through `allowed_actions = []`.
- Defers `open_app`, command aliases and domain allowlists to later hardening work.
- Still requires native validation on Windows and Linux before broad operational use.
