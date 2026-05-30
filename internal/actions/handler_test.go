package actions

import (
	"context"
	"strings"
	"testing"

	"donagent/internal/events"
)

func TestHandlerExecutesAllowedOpenURL(t *testing.T) {
	executor := &executorSpy{}
	handler := NewHandler(executor, []string{TypeOpenURL}, nil)

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Title:  "Open docs",
		Type:   TypeOpenURL,
		Target: "https://example.com/docs",
	})
	if err != nil {
		t.Fatalf("HandleAction() error = %v", err)
	}
	if !executor.called {
		t.Fatal("executor was not called")
	}
	if executor.request.Target != "https://example.com/docs" {
		t.Fatalf("Target = %q", executor.request.Target)
	}
}

func TestHandlerRejectsActionOutsideAllowlist(t *testing.T) {
	executor := &executorSpy{}
	handler := NewHandler(executor, nil, nil)

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Type:   TypeOpenURL,
		Target: "https://example.com",
	})
	if err == nil {
		t.Fatal("HandleAction() error = nil")
	}
	if executor.called {
		t.Fatal("executor was called")
	}
}

func TestHandlerRejectsUnsupportedActionType(t *testing.T) {
	executor := &executorSpy{}
	handler := NewHandler(executor, []string{"run_command"}, nil)

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Type:   "run_command",
		Target: "notepad",
	})
	if err == nil {
		t.Fatal("HandleAction() error = nil")
	}
	if executor.called {
		t.Fatal("executor was called")
	}
}

func TestHandlerRejectsUnsafeURL(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{name: "javascript scheme", target: "javascript:alert(1)"},
		{name: "missing host", target: "https:///missing-host"},
		{name: "credentials", target: "https://user:pass@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &executorSpy{}
			handler := NewHandler(executor, []string{TypeOpenURL}, nil)

			err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
				Type:   TypeOpenURL,
				Target: tt.target,
			})
			if err == nil {
				t.Fatal("HandleAction() error = nil")
			}
			if executor.called {
				t.Fatal("executor was called")
			}
		})
	}
}

func TestHandlerRequiresExecutor(t *testing.T) {
	handler := NewHandler(nil, []string{TypeOpenURL}, nil)

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Type:   TypeOpenURL,
		Target: "https://example.com",
	})
	if err == nil || !strings.Contains(err.Error(), "executor") {
		t.Fatalf("HandleAction() error = %v", err)
	}
}

func TestHandlerExecutesConfiguredOpenAppAlias(t *testing.T) {
	executor := &executorSpy{}
	handler := NewHandler(executor, []string{TypeOpenApp}, map[string]string{
		"editor": "/usr/bin/nano",
	})

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Type:   TypeOpenApp,
		Target: "editor",
	})
	if err != nil {
		t.Fatalf("HandleAction() error = %v", err)
	}
	if !executor.called {
		t.Fatal("executor was not called")
	}
	if executor.request.Target != "/usr/bin/nano" {
		t.Fatalf("Target = %q", executor.request.Target)
	}
}

func TestHandlerRejectsUnknownOpenAppAlias(t *testing.T) {
	executor := &executorSpy{}
	handler := NewHandler(executor, []string{TypeOpenApp}, map[string]string{
		"editor": "/usr/bin/nano",
	})

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Type:   TypeOpenApp,
		Target: "browser",
	})
	if err == nil {
		t.Fatal("HandleAction() error = nil")
	}
	if executor.called {
		t.Fatal("executor was called")
	}
}

func TestHandlerRejectsInvalidOpenAppAlias(t *testing.T) {
	executor := &executorSpy{}
	handler := NewHandler(executor, []string{TypeOpenApp}, map[string]string{
		"../editor": "/usr/bin/nano",
	})

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Type:   TypeOpenApp,
		Target: "../editor",
	})
	if err == nil {
		t.Fatal("HandleAction() error = nil")
	}
	if executor.called {
		t.Fatal("executor was called")
	}
}

func TestHandlerRejectsArgs(t *testing.T) {
	executor := &executorSpy{}
	handler := NewHandler(executor, []string{TypeOpenApp}, map[string]string{
		"editor": "/usr/bin/nano",
	})

	err := handler.HandleAction(context.Background(), events.Event{}, events.ActionPayload{
		Type:   TypeOpenApp,
		Target: "editor",
		Args:   []string{"file.txt"},
	})
	if err == nil {
		t.Fatal("HandleAction() error = nil")
	}
	if executor.called {
		t.Fatal("executor was called")
	}
}

type executorSpy struct {
	called  bool
	request Request
}

func (executor *executorSpy) Execute(_ context.Context, request Request) error {
	executor.called = true
	executor.request = request
	return nil
}
