//go:build !windows && !linux

package actions

import (
	"context"
	"fmt"
	"runtime"
)

// SystemExecutor performs supported actions using operating system facilities.
type SystemExecutor struct{}

// Execute returns an explicit error on unsupported operating systems.
func (SystemExecutor) Execute(context.Context, Request) error {
	return fmt.Errorf("actions are not supported on %s", runtime.GOOS)
}
