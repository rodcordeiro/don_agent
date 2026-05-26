package events

import (
	"context"
	"testing"
)

func TestRouterRoutesNotification(t *testing.T) {
	handler := &notificationHandlerSpy{}
	router := NewRouter(handler)

	event, err := Parse([]byte(`{"event":"notification","payload":{"title":"Hello"}}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if err := router.Route(context.Background(), event); err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if !handler.called {
		t.Fatal("notification handler was not called")
	}
}

func TestRouterRejectsActionUntilImplemented(t *testing.T) {
	router := NewRouter(&notificationHandlerSpy{})

	event, err := Parse([]byte(`{"event":"action","payload":{"type":"open_url","target":"https://example.com"}}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if err := router.Route(context.Background(), event); err == nil {
		t.Fatal("Route() error = nil")
	}
}

type notificationHandlerSpy struct {
	called bool
}

func (handler *notificationHandlerSpy) HandleNotification(_ context.Context, _ Event, _ NotificationPayload) error {
	handler.called = true
	return nil
}

