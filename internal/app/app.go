package app

import (
	"context"
	"fmt"

	"donagent/internal/config"
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

	fmt.Printf("DonAgent started. Queue: %s\n", cfg.QueueName)

	err = consumer.Consume(ctx, func(_ context.Context, delivery rabbitmq.Delivery) error {
		fmt.Printf("DonAgent received message. Bytes: %d\n", len(delivery.Body))
		return nil
	})
	if err == context.Canceled {
		return nil
	}

	return err
}
