//go:build linux

package actions

import (
	"context"
	"os/exec"
)

// SystemExecutor performs supported actions using operating system facilities.
type SystemExecutor struct{}

// Execute runs a previously validated action without invoking a shell.
func (SystemExecutor) Execute(ctx context.Context, request Request) error {
	switch request.Type {
	case TypeOpenURL:
		return exec.CommandContext(ctx, "xdg-open", request.Target).Start()
	case TypeOpenApp:
		return exec.CommandContext(ctx, request.Target).Start()
	default:
		return nil
	}
}
