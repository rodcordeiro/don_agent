package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	TypeNotification = "notification"
	TypeAction       = "action"
)

var (
	ErrInvalidJSON      = errors.New("invalid event json")
	ErrUnknownEventType = errors.New("unknown event type")
	ErrInvalidPayload   = errors.New("invalid event payload")
)

// Event is the raw event envelope consumed from RabbitMQ.
type Event struct {
	Event    string          `json:"event"`
	Payload  json.RawMessage `json:"payload"`
	Metadata Metadata        `json:"metadata"`
}

// Metadata identifies event provenance.
type Metadata struct {
	CreatedAt string `json:"created_at"`
	Author    string `json:"author"`
	Origin    string `json:"origin"`
}

// NotificationPayload contains data required to display a desktop notification.
type NotificationPayload struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

// Parse validates the event envelope and event-specific payloads.
func Parse(data []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return Event{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	event.Event = strings.TrimSpace(event.Event)
	switch event.Event {
	case TypeNotification:
		if _, err := ParseNotificationPayload(event); err != nil {
			return Event{}, err
		}
	case TypeAction:
		if len(event.Payload) == 0 || string(event.Payload) == "null" {
			return Event{}, fmt.Errorf("%w: action payload is required", ErrInvalidPayload)
		}
	default:
		return Event{}, fmt.Errorf("%w: %q", ErrUnknownEventType, event.Event)
	}

	return event, nil
}

// ParseNotificationPayload validates and returns a notification payload.
func ParseNotificationPayload(event Event) (NotificationPayload, error) {
	if event.Event != TypeNotification {
		return NotificationPayload{}, fmt.Errorf("%w: expected notification event", ErrInvalidPayload)
	}

	var payload NotificationPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return NotificationPayload{}, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	payload.Title = strings.TrimSpace(payload.Title)
	payload.Description = strings.TrimSpace(payload.Description)
	payload.ImageURL = strings.TrimSpace(payload.ImageURL)

	if payload.Title == "" {
		return NotificationPayload{}, fmt.Errorf("%w: notification title is required", ErrInvalidPayload)
	}

	return payload, nil
}

// CreatedAtTime parses metadata creation time when present.
func (metadata Metadata) CreatedAtTime() (time.Time, error) {
	if strings.TrimSpace(metadata.CreatedAt) == "" {
		return time.Time{}, nil
	}

	createdAt, err := time.Parse(time.RFC3339, metadata.CreatedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid metadata created_at: %w", err)
	}

	return createdAt, nil
}

