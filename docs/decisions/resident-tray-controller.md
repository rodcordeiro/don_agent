# Decision - Resident tray controller foundation

Status: accepted

## Context

Milestone 2 requires a resident agent with tray actions for status, pause, resume and graceful shutdown. A native tray UI requires operating-system dependencies and native validation on Windows and Linux, which can break the current Docker/CI path if introduced prematurely.

## Decision

Create an internal `tray.Controller` first and wire the application lifecycle to it.

The controller owns:

- resident status;
- pause and resume state;
- graceful stop signal;
- app display name;
- icon path pointing to `assets/logo.png`.

RabbitMQ prefetch is limited to one message so pause can stop further processing without acknowledging additional messages. The native GUI tray library remains a future adapter over this controller.

## Consequences

- Keeps CI/build free of GUI dependencies.
- Creates a stable seam for a native tray implementation.
- Pause/resume can be tested without desktop APIs.
- The user-visible tray menu is still pending native Windows/Linux validation.
