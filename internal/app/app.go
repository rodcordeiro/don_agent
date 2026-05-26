package app

import (
	"context"
	"fmt"

	"donagent/internal/config"
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

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Printf("DonAgent started. Queue: %s\n", cfg.QueueName)
		return nil
	}
}
