package app

import (
	"context"
	"fmt"

	"donagent/internal/actions"
	"donagent/internal/config"
	"donagent/internal/events"
	"donagent/internal/notifications"
	"donagent/internal/rabbitmq"
	"donagent/internal/tray"
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

	resident := tray.NewController(tray.Config{
		AppName:  "DonAgent",
		IconPath: "assets/logo.png",
	})
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-resident.Done():
			cancel()
		case <-runCtx.Done():
		}
	}()

	consumer, err := rabbitmq.NewConsumer(rabbitmq.Config{
		URL:       cfg.RabbitURL,
		QueueName: cfg.QueueName,
	})
	if err != nil {
		return err
	}
	defer consumer.Close()

	router := events.NewRouter(
		notifications.NewHandler(notifications.ConsoleNotifier{}),
		actions.NewHandler(actions.SystemExecutor{}, cfg.AllowedActions, cfg.AppAliases),
	)

	resident.SetStatus(tray.StatusRunning)
	defer resident.SetStatus(tray.StatusStopped)

	fmt.Printf("DonAgent started. Queue: %s. Status: %s\n", cfg.QueueName, resident.Status())

	err = consumer.Consume(runCtx, func(ctx context.Context, delivery rabbitmq.Delivery) error {
		if err := resident.WaitIfPaused(ctx); err != nil {
			return err
		}
		return processDelivery(ctx, router, rabbitDelivery{delivery: delivery})
	})
	if err == context.Canceled {
		return nil
	}

	return err
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
