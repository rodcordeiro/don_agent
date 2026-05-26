# Decision - Desktop notification library

Status: pending

## Context

Milestone 1 needs a notification handler that receives `title`, `description` and optional `image_url` without coupling event routing to a specific operating system library.

Native desktop notification behavior differs between Windows and Linux, and validation cannot be fully represented inside the development container.

## Current decision

Use an internal `Notifier` interface and a `ConsoleNotifier` implementation for the Milestone 1 runtime path.

The concrete desktop notification library remains pending until native validation is available on Windows and Linux.

## Trade-offs

- Keeps event processing testable without desktop dependencies.
- Avoids adding a GUI/native dependency before validating Windows/Linux behavior.
- Does not display real OS notifications yet.
- Makes the future library swap local to `internal/notifications`.

## Selection criteria for the real library

- Supports Windows and Linux.
- Can show title and description.
- Handles missing `image_url` safely.
- Has acceptable maintenance activity.
- Does not require broad OS permissions beyond notifications.
- Can fail gracefully without crashing the agent.

