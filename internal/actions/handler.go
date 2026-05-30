package actions

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"donagent/internal/events"
)

const TypeOpenURL = "open_url"
const TypeOpenApp = "open_app"

var appAliasPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

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
	appAliases     map[string]string
}

// NewHandler creates an action event handler.
func NewHandler(executor Executor, allowedActions []string, appAliases map[string]string) Handler {
	allowed := make(map[string]struct{}, len(allowedActions))
	for _, action := range allowedActions {
		action = strings.TrimSpace(action)
		if action != "" {
			allowed[action] = struct{}{}
		}
	}

	aliases := make(map[string]string, len(appAliases))
	for alias, target := range appAliases {
		alias = strings.TrimSpace(alias)
		target = strings.TrimSpace(target)
		if alias != "" && target != "" {
			aliases[alias] = target
		}
	}

	return Handler{executor: executor, allowedActions: allowed, appAliases: aliases}
}

// HandleAction validates and dispatches a semantic local action event.
func (handler Handler) HandleAction(ctx context.Context, _ events.Event, payload events.ActionPayload) error {
	if handler.executor == nil {
		return fmt.Errorf("action executor is required")
	}
	if _, ok := handler.allowedActions[payload.Type]; !ok {
		return fmt.Errorf("action type %q is not allowed", payload.Type)
	}
	if len(payload.Args) > 0 {
		return fmt.Errorf("action args are not supported yet")
	}

	target := payload.Target
	switch payload.Type {
	case TypeOpenURL:
		if err := validateHTTPURL(payload.Target); err != nil {
			return err
		}
	case TypeOpenApp:
		aliasTarget, err := handler.resolveAppAlias(payload.Target)
		if err != nil {
			return err
		}
		target = aliasTarget
	default:
		return fmt.Errorf("action type %q is not supported", payload.Type)
	}

	return handler.executor.Execute(ctx, Request{
		Title:  payload.Title,
		Type:   payload.Type,
		Target: target,
		Args:   payload.Args,
	})
}

func (handler Handler) resolveAppAlias(alias string) (string, error) {
	if !appAliasPattern.MatchString(alias) || strings.Contains(alias, "..") {
		return "", fmt.Errorf("app alias %q is invalid", alias)
	}

	target, ok := handler.appAliases[alias]
	if !ok {
		return "", fmt.Errorf("app alias %q is not configured", alias)
	}
	if strings.TrimSpace(target) == "" {
		return "", fmt.Errorf("app alias %q target is empty", alias)
	}

	return target, nil
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
