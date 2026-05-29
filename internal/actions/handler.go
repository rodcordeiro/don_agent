package actions

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"donagent/internal/events"
)

const TypeOpenURL = "open_url"

// Request is a validated local action execution request.
type Request struct {
	Title  string
	Type   string
	Target string
	Args   []string
}

// Executor performs a local action after policy validation.
type Executor interface {
	Execute(context.Context, Request) error
}

// Handler validates action policy before delegating local execution.
type Handler struct {
	executor       Executor
	allowedActions map[string]struct{}
}

// NewHandler creates an action event handler.
func NewHandler(executor Executor, allowedActions []string) Handler {
	allowed := make(map[string]struct{}, len(allowedActions))
	for _, action := range allowedActions {
		action = strings.TrimSpace(action)
		if action != "" {
			allowed[action] = struct{}{}
		}
	}

	return Handler{executor: executor, allowedActions: allowed}
}

// HandleAction validates and dispatches a semantic local action event.
func (handler Handler) HandleAction(ctx context.Context, _ events.Event, payload events.ActionPayload) error {
	if handler.executor == nil {
		return fmt.Errorf("action executor is required")
	}
	if _, ok := handler.allowedActions[payload.Type]; !ok {
		return fmt.Errorf("action type %q is not allowed", payload.Type)
	}
	if payload.Type != TypeOpenURL {
		return fmt.Errorf("action type %q is not supported", payload.Type)
	}
	if err := validateHTTPURL(payload.Target); err != nil {
		return err
	}

	return handler.executor.Execute(ctx, Request{
		Title:  payload.Title,
		Type:   payload.Type,
		Target: payload.Target,
		Args:   payload.Args,
	})
}

func validateHTTPURL(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid action target url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("action target url scheme %q is not allowed", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("action target url host is required")
	}
	if parsed.User != nil {
		return fmt.Errorf("action target url credentials are not allowed")
	}

	return nil
}
