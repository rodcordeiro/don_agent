package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"donagent/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "donagent: %v\n", err)
		os.Exit(1)
	}
}

