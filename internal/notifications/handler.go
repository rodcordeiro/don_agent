package notifications

import (
	"context"
	"fmt"

	"donagent/internal/events"
)

// Message represents a desktop notification request.
type Message struct {
	Title       string
	Description string
	ImageURL    string
}

// Notifier displays a notification to the user.
type Notifier interface {
	Notify(context.Context, Message) error
}

// Handler processes notification events.
type Handler struct {
	notifier Notifier
}

// NewHandler creates a notification event handler.
func NewHandler(notifier Notifier) Handler {
	return Handler{notifier: notifier}
}

// HandleNotification validates and dispatches a notification event.
func (handler Handler) HandleNotification(ctx context.Context, _ events.Event, payload events.NotificationPayload) error {
	if handler.notifier == nil {
		return fmt.Errorf("notification notifier is required")
	}

	return handler.notifier.Notify(ctx, Message{
		Title:       payload.Title,
		Description: payload.Description,
		ImageURL:    payload.ImageURL,
	})
}

