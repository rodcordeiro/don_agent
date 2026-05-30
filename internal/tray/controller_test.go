package tray

import (
	"context"
	"testing"
	"time"
)

func TestControllerDefaults(t *testing.T) {
	controller := NewController(Config{})

	if controller.AppName() != "DonAgent" {
		t.Fatalf("AppName() = %q", controller.AppName())
	}
	if controller.IconPath() != "assets/logo.png" {
		t.Fatalf("IconPath() = %q", controller.IconPath())
	}
	if controller.Status() != StatusStarting {
		t.Fatalf("Status() = %q", controller.Status())
	}
}

func TestControllerPauseResume(t *testing.T) {
	controller := NewController(Config{})
	controller.SetStatus(StatusRunning)
	controller.Pause()

	if controller.Status() != StatusPaused {
		t.Fatalf("Status() = %q", controller.Status())
	}

	resumed := make(chan error, 1)
	go func() {
		resumed <- controller.WaitIfPaused(context.Background())
	}()

	select {
	case err := <-resumed:
		t.Fatalf("WaitIfPaused() returned before resume: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	controller.Resume()

	select {
	case err := <-resumed:
		if err != nil {
			t.Fatalf("WaitIfPaused() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitIfPaused() did not return after resume")
	}
	if controller.Status() != StatusRunning {
		t.Fatalf("Status() = %q", controller.Status())
	}
}

func TestControllerStopUnblocksPausedWait(t *testing.T) {
	controller := NewController(Config{})
	controller.Pause()

	stopped := make(chan error, 1)
	go func() {
		stopped <- controller.WaitIfPaused(context.Background())
	}()

	controller.Stop()

	select {
	case err := <-stopped:
		if err == nil {
			t.Fatal("WaitIfPaused() error = nil")
		}
	case <-time.After(time.Second):
		t.Fatal("WaitIfPaused() did not return after stop")
	}
	if controller.Status() != StatusStopping {
		t.Fatalf("Status() = %q", controller.Status())
	}
}
