package tray

import (
	"context"
	"sync"
)

const defaultAppName = "DonAgent"
const defaultIconPath = "assets/logo.png"

// Status represents the resident agent lifecycle state.
type Status string

const (
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusPaused   Status = "paused"
	StatusStopping Status = "stopping"
	StatusStopped  Status = "stopped"
	StatusFailed   Status = "failed"
)

// Config contains resident tray metadata.
type Config struct {
	AppName  string
	IconPath string
}

// Controller holds the state that a native tray menu can drive.
type Controller struct {
	appName  string
	iconPath string

	mu      sync.RWMutex
	status  Status
	paused  bool
	resume  chan struct{}
	stopped chan struct{}
	once    sync.Once
}

// NewController creates a resident controller for tray-backed actions.
func NewController(cfg Config) *Controller {
	appName := cfg.AppName
	if appName == "" {
		appName = defaultAppName
	}

	iconPath := cfg.IconPath
	if iconPath == "" {
		iconPath = defaultIconPath
	}

	return &Controller{
		appName:  appName,
		iconPath: iconPath,
		status:   StatusStarting,
		stopped:  make(chan struct{}),
	}
}

// AppName returns the display name used by a native tray implementation.
func (controller *Controller) AppName() string {
	return controller.appName
}

// IconPath returns the icon path used by a native tray implementation.
func (controller *Controller) IconPath() string {
	return controller.iconPath
}

// Status returns the current resident status.
func (controller *Controller) Status() Status {
	controller.mu.RLock()
	defer controller.mu.RUnlock()

	return controller.status
}

// SetStatus updates the resident status.
func (controller *Controller) SetStatus(status Status) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	controller.status = status
}

// Pause pauses message processing.
func (controller *Controller) Pause() {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.paused {
		return
	}

	controller.paused = true
	controller.resume = make(chan struct{})
	controller.status = StatusPaused
}

// Resume resumes message processing.
func (controller *Controller) Resume() {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if !controller.paused {
		return
	}

	close(controller.resume)
	controller.resume = nil
	controller.paused = false
	controller.status = StatusRunning
}

// Stop requests graceful shutdown.
func (controller *Controller) Stop() {
	controller.once.Do(func() {
		controller.mu.Lock()
		controller.status = StatusStopping
		if controller.paused && controller.resume != nil {
			close(controller.resume)
			controller.resume = nil
			controller.paused = false
		}
		controller.mu.Unlock()

		close(controller.stopped)
	})
}

// Done is closed when shutdown is requested.
func (controller *Controller) Done() <-chan struct{} {
	return controller.stopped
}

// WaitIfPaused blocks while the resident controller is paused.
func (controller *Controller) WaitIfPaused(ctx context.Context) error {
	for {
		controller.mu.RLock()
		resume := controller.resume
		controller.mu.RUnlock()

		if resume == nil {
			select {
			case <-controller.stopped:
				return context.Canceled
			default:
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-controller.stopped:
			return context.Canceled
		case <-resume:
		}
	}
}
