package app

import (
	"context"
	"fmt"
)

// Run starts the DonAgent application lifecycle.
func Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Println("DonAgent started")
		return nil
	}
}

