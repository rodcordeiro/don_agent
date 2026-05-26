package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"donagent/internal/events"
)

func TestProcessDeliveryAcksAfterSuccessfulProcessing(t *testing.T) {
	msg := &deliverySpy{body: []byte(`{"event":"notification","payload":{"title":"Hello"}}`)}
	router := events.NewRouter(notificationHandler{})

	if err := processDelivery(context.Background(), router, msg); err != nil {
		t.Fatalf("processDelivery() error = %v", err)
	}
	if !msg.acked {
		t.Fatal("message was not acked")
	}
}

func TestProcessDeliveryRejectsInvalidJSON(t *testing.T) {
	msg := &deliverySpy{body: []byte(`{"event":`)}
	router := events.NewRouter(notificationHandler{})

	if err := processDelivery(context.Background(), router, msg); err != nil {
		t.Fatalf("processDelivery() error = %v", err)
	}
	if !msg.rejected || msg.requeue {
		t.Fatalf("rejected = %v, requeue = %v", msg.rejected, msg.requeue)
	}
}

func TestProcessDeliveryRejectsUnknownEvent(t *testing.T) {
	msg := &deliverySpy{body: []byte(`{"event":"unknown","payload":{}}`)}
	router := events.NewRouter(notificationHandler{})

	if err := processDelivery(context.Background(), router, msg); err != nil {
		t.Fatalf("processDelivery() error = %v", err)
	}
	if !msg.rejected || msg.requeue {
		t.Fatalf("rejected = %v, requeue = %v", msg.rejected, msg.requeue)
	}
}

func TestProcessDeliveryRejectsInvalidPayload(t *testing.T) {
	msg := &deliverySpy{body: []byte(`{"event":"notification","payload":{"description":"missing title"}}`)}
	router := events.NewRouter(notificationHandler{})

	if err := processDelivery(context.Background(), router, msg); err != nil {
		t.Fatalf("processDelivery() error = %v", err)
	}
	if !msg.rejected || msg.requeue {
		t.Fatalf("rejected = %v, requeue = %v", msg.rejected, msg.requeue)
	}
}

func TestProcessDeliveryNacksTemporaryError(t *testing.T) {
	msg := &deliverySpy{body: []byte(`{"event":"notification","payload":{"title":"Hello"}}`)}
	router := events.NewRouter(notificationHandler{err: ErrTemporaryProcessing})

	if err := processDelivery(context.Background(), router, msg); err != nil {
		t.Fatalf("processDelivery() error = %v", err)
	}
	if !msg.nacked || !msg.requeue {
		t.Fatalf("nacked = %v, requeue = %v", msg.nacked, msg.requeue)
	}
}

func TestProcessDeliveryRejectsPermanentProcessingError(t *testing.T) {
	msg := &deliverySpy{body: []byte(`{"event":"notification","payload":{"title":"Hello"}}`)}
	router := events.NewRouter(notificationHandler{err: errors.New("permanent failure")})

	if err := processDelivery(context.Background(), router, msg); err != nil {
		t.Fatalf("processDelivery() error = %v", err)
	}
	if !msg.rejected || msg.requeue {
		t.Fatalf("rejected = %v, requeue = %v", msg.rejected, msg.requeue)
	}
}

func TestSanitizeLogMessageRedactsKnownSecrets(t *testing.T) {
	got := sanitizeLogMessage(errors.New("password=abc token=def secret=ghi amqp://guest:guest@rabbitmq:5672/"))

	for _, leaked := range []string{"abc", "def", "ghi", "guest:guest"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("sanitized message leaked %q: %q", leaked, got)
		}
	}
}

type deliverySpy struct {
	body     []byte
	acked    bool
	nacked   bool
	rejected bool
	requeue  bool
}

func (delivery *deliverySpy) Ack(_ bool) error {
	delivery.acked = true
	return nil
}

func (delivery *deliverySpy) Nack(_ bool, requeue bool) error {
	delivery.nacked = true
	delivery.requeue = requeue
	return nil
}

func (delivery *deliverySpy) Reject(requeue bool) error {
	delivery.rejected = true
	delivery.requeue = requeue
	return nil
}

func (delivery *deliverySpy) Body() []byte {
	return delivery.body
}

type notificationHandler struct {
	err error
}

func (handler notificationHandler) HandleNotification(_ context.Context, _ events.Event, _ events.NotificationPayload) error {
	return handler.err
}
