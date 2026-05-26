package app

import (
	"context"
	"fmt"

	"donagent/internal/config"
	"donagent/internal/events"
	"donagent/internal/rabbitmq"
)

// Run starts the DonAgent application lifecycle.
func Run(ctx context.Context) error {
	configPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	consumer, err := rabbitmq.NewConsumer(rabbitmq.Config{
		URL:       cfg.RabbitURL,
		QueueName: cfg.QueueName,
	})
	if err != nil {
		return err
	}
	defer consumer.Close()

	router := events.NewRouter(logNotificationHandler{})

	fmt.Printf("DonAgent started. Queue: %s\n", cfg.QueueName)

	err = consumer.Consume(ctx, func(ctx context.Context, delivery rabbitmq.Delivery) error {
		return processDelivery(ctx, router, rabbitDelivery{delivery: delivery})
	})
	if err == context.Canceled {
		return nil
	}

	return err
}

type logNotificationHandler struct{}

func (logNotificationHandler) HandleNotification(_ context.Context, event events.Event, payload events.NotificationPayload) error {
	fmt.Printf("Notification event received. Title: %s Origin: %s\n", payload.Title, event.Metadata.Origin)
	return nil
}

type rabbitDelivery struct {
	delivery rabbitmq.Delivery
}

func (delivery rabbitDelivery) Ack(multiple bool) error {
	return delivery.delivery.Ack(multiple)
}

func (delivery rabbitDelivery) Nack(multiple bool, requeue bool) error {
	return delivery.delivery.Nack(multiple, requeue)
}

func (delivery rabbitDelivery) Reject(requeue bool) error {
	return delivery.delivery.Reject(requeue)
}

func (delivery rabbitDelivery) Body() []byte {
	return delivery.delivery.Body
}
