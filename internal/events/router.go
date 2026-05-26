package events

import (
	"context"
	"fmt"
)

// NotificationHandler processes a validated notification event.
type NotificationHandler interface {
	HandleNotification(context.Context, Event, NotificationPayload) error
}

// Router sends validated events to the correct handler.
type Router struct {
	notifications NotificationHandler
}

// NewRouter creates an event router.
func NewRouter(notifications NotificationHandler) Router {
	return Router{notifications: notifications}
}

// Route dispatches one parsed event.
func (router Router) Route(ctx context.Context, event Event) error {
	switch event.Event {
	case TypeNotification:
		if router.notifications == nil {
			return fmt.Errorf("%w: notification handler is required", ErrInvalidPayload)
		}

		payload, err := ParseNotificationPayload(event)
		if err != nil {
			return err
		}

		return router.notifications.HandleNotification(ctx, event, payload)
	case TypeAction:
		return fmt.Errorf("%w: action events are not implemented yet", ErrUnknownEventType)
	default:
		return fmt.Errorf("%w: %q", ErrUnknownEventType, event.Event)
	}
}

