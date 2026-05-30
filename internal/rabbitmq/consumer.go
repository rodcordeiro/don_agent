package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config contains RabbitMQ connection and queue settings.
type Config struct {
	URL       string
	QueueName string
}

// Delivery is the message type delivered by RabbitMQ.
type Delivery = amqp.Delivery

// DeliveryHandler processes one RabbitMQ delivery.
type DeliveryHandler func(context.Context, Delivery) error

// Consumer manages a RabbitMQ connection and queue consumer.
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewConsumer creates a RabbitMQ connection and channel.
func NewConsumer(cfg Config) (*Consumer, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("rabbitmq url is required")
	}
	if cfg.QueueName == "" {
		return nil, fmt.Errorf("rabbitmq queue name is required")
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	if _, err := channel.QueueDeclare(
		cfg.QueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue %q: %w", cfg.QueueName, err)
	}

	if err := channel.Qos(1, 0, false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("configure queue prefetch: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
		queue:   cfg.QueueName,
	}, nil
}

// Consume starts consuming messages from the configured queue until the context is canceled.
func (consumer *Consumer) Consume(ctx context.Context, handler DeliveryHandler) error {
	if handler == nil {
		return fmt.Errorf("delivery handler is required")
	}

	deliveries, err := consumer.channel.ConsumeWithContext(
		ctx,
		consumer.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue %q: %w", consumer.queue, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbitmq delivery channel closed")
			}
			if err := handler(ctx, delivery); err != nil {
				return err
			}
		}
	}
}

// Close releases RabbitMQ resources.
func (consumer *Consumer) Close() error {
	var closeErr error

	if consumer.channel != nil {
		if err := consumer.channel.Close(); err != nil {
			closeErr = err
		}
	}
	if consumer.conn != nil {
		if err := consumer.conn.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}

	return closeErr
}
