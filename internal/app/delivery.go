package app

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"donagent/internal/events"
)

// ErrTemporaryProcessing marks a message processing error that can be retried.
var ErrTemporaryProcessing = errors.New("temporary processing error")

var (
	sensitiveKeyPattern = regexp.MustCompile(`(?i)(password|token|secret)=([^&\s]+)`)
	amqpCredentialPattern = regexp.MustCompile(`(?i)amqps?://([^:@/\s]+):([^@/\s]+)@`)
)

type delivery interface {
	Ack(multiple bool) error
	Nack(multiple bool, requeue bool) error
	Reject(requeue bool) error
	Body() []byte
}

func processDelivery(ctx context.Context, router events.Router, msg delivery) error {
	event, err := events.Parse(msg.Body())
	if err != nil {
		fmt.Printf("DonAgent rejected invalid message: %s\n", sanitizeLogMessage(err))
		return msg.Reject(false)
	}

	if err := router.Route(ctx, event); err != nil {
		fmt.Printf("DonAgent failed to process message: %s\n", sanitizeLogMessage(err))
		if errors.Is(err, ErrTemporaryProcessing) {
			return msg.Nack(false, true)
		}
		return msg.Reject(false)
	}

	return msg.Ack(false)
}

func sanitizeLogMessage(err error) string {
	if err == nil {
		return ""
	}

	message := err.Error()
	message = sensitiveKeyPattern.ReplaceAllString(message, "$1=<redacted>")
	message = amqpCredentialPattern.ReplaceAllString(message, "amqp://<redacted>@")

	if len(message) > 240 {
		return message[:240] + "..."
	}

	return message
}
