package events

import (
	"errors"
	"testing"
)

func TestParseNotificationEvent(t *testing.T) {
	raw := []byte(`{
		"event": "notification",
		"payload": {
			"title": "Build finished",
			"description": "Pipeline completed",
			"image_url": "https://example.com/image.png"
		},
		"metadata": {
			"created_at": "2026-05-08T11:12:00Z",
			"author": "RodCordeiro",
			"origin": "n8n"
		}
	}`)

	event, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if event.Event != TypeNotification {
		t.Fatalf("Event = %q", event.Event)
	}

	payload, err := ParseNotificationPayload(event)
	if err != nil {
		t.Fatalf("ParseNotificationPayload() error = %v", err)
	}
	if payload.Title != "Build finished" {
		t.Fatalf("Title = %q", payload.Title)
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse([]byte(`{"event":`))
	if !errors.Is(err, ErrInvalidJSON) {
		t.Fatalf("Parse() error = %v, want ErrInvalidJSON", err)
	}
}

func TestParseUnknownEvent(t *testing.T) {
	_, err := Parse([]byte(`{"event":"unknown","payload":{}}`))
	if !errors.Is(err, ErrUnknownEventType) {
		t.Fatalf("Parse() error = %v, want ErrUnknownEventType", err)
	}
}

func TestParseActionEvent(t *testing.T) {
	event, err := Parse([]byte(`{"event":"action","payload":{"type":"open_url","target":"https://example.com"}}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if event.Event != TypeAction {
		t.Fatalf("Event = %q", event.Event)
	}

	payload, err := ParseActionPayload(event)
	if err != nil {
		t.Fatalf("ParseActionPayload() error = %v", err)
	}
	if payload.Type != "open_url" {
		t.Fatalf("Type = %q", payload.Type)
	}
}

func TestParseActionWithoutPayload(t *testing.T) {
	_, err := Parse([]byte(`{"event":"action"}`))
	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("Parse() error = %v, want ErrInvalidPayload", err)
	}
}

func TestParseActionWithoutType(t *testing.T) {
	_, err := Parse([]byte(`{"event":"action","payload":{"target":"https://example.com"}}`))
	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("Parse() error = %v, want ErrInvalidPayload", err)
	}
}

func TestParseActionWithoutTarget(t *testing.T) {
	_, err := Parse([]byte(`{"event":"action","payload":{"type":"open_url"}}`))
	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("Parse() error = %v, want ErrInvalidPayload", err)
	}
}

func TestParseInvalidNotificationPayload(t *testing.T) {
	_, err := Parse([]byte(`{"event":"notification","payload":{"description":"missing title"}}`))
	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("Parse() error = %v, want ErrInvalidPayload", err)
	}
}

func TestMetadataCreatedAtTime(t *testing.T) {
	metadata := Metadata{CreatedAt: "2026-05-08T11:12:00Z"}

	createdAt, err := metadata.CreatedAtTime()
	if err != nil {
		t.Fatalf("CreatedAtTime() error = %v", err)
	}
	if createdAt.IsZero() {
		t.Fatal("CreatedAtTime() returned zero time")
	}
}
