package notifications

import (
	"context"
	"fmt"
)

// ConsoleNotifier writes notifications to stdout until native desktop support is selected.
type ConsoleNotifier struct{}

// Notify displays the notification as console output.
func (ConsoleNotifier) Notify(_ context.Context, message Message) error {
	fmt.Printf("Notification: %s\n", message.Title)
	if message.Description != "" {
		fmt.Printf("Description: %s\n", message.Description)
	}
	if message.ImageURL != "" {
		fmt.Printf("Image: %s\n", message.ImageURL)
	}

	return nil
}

