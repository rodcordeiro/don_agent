package events

import (
	"context"
	"fmt"
)

// NotificationHandler processes a validated notification event.
type NotificationHandler interface {
	HandleNotification(context.Context, Event, NotificationPayload) error
}

// ActionHandler processes a validated action event.
type ActionHandler interface {
	HandleAction(context.Context, Event, ActionPayload) error
}

// Router sends validated events to the correct handler.
type Router struct {
	notifications NotificationHandler
	actions       ActionHandler
}

// NewRouter creates an event router.
func NewRouter(notifications NotificationHandler, actions ActionHandler) Router {
	return Router{notifications: notifications, actions: actions}
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
		if router.actions == nil {
			return fmt.Errorf("%w: action handler is required", ErrInvalidPayload)
		}

		payload, err := ParseActionPayload(event)
		if err != nil {
			return err
		}

		return router.actions.HandleAction(ctx, event, payload)
	default:
		return fmt.Errorf("%w: %q", ErrUnknownEventType, event.Event)
	}
}
