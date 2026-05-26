package notifications

import (
	"context"
	"errors"
	"testing"

	"donagent/internal/events"
)

func TestHandlerDispatchesNotification(t *testing.T) {
	notifier := &notifierSpy{}
	handler := NewHandler(notifier)

	err := handler.HandleNotification(context.Background(), events.Event{}, events.NotificationPayload{
		Title:       "Hello",
		Description: "World",
		ImageURL:    "https://example.com/image.png",
	})
	if err != nil {
		t.Fatalf("HandleNotification() error = %v", err)
	}

	if !notifier.called {
		t.Fatal("notifier was not called")
	}
	if notifier.message.Title != "Hello" {
		t.Fatalf("Title = %q", notifier.message.Title)
	}
}

func TestHandlerReturnsNotifierError(t *testing.T) {
	want := errors.New("temporary notification failure")
	handler := NewHandler(&notifierSpy{err: want})

	err := handler.HandleNotification(context.Background(), events.Event{}, events.NotificationPayload{Title: "Hello"})
	if !errors.Is(err, want) {
		t.Fatalf("HandleNotification() error = %v, want %v", err, want)
	}
}

func TestHandlerRequiresNotifier(t *testing.T) {
	handler := NewHandler(nil)

	if err := handler.HandleNotification(context.Background(), events.Event{}, events.NotificationPayload{Title: "Hello"}); err == nil {
		t.Fatal("HandleNotification() error = nil")
	}
}

type notifierSpy struct {
	called  bool
	message Message
	err     error
}

func (notifier *notifierSpy) Notify(_ context.Context, message Message) error {
	notifier.called = true
	notifier.message = message
	return notifier.err
}

