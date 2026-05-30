package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultContractVersion = "1"
	defaultLogMaxSizeMB    = 10
	actionOpenApp          = "open_app"
)

var appAliasPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

// Config holds DonAgent runtime settings loaded from the local config file.
type Config struct {
	RabbitURL            string
	Username             string
	Password             string
	QueueName            string
	ExchangeName         string
	EventContractVersion string
	LogPath              string
	LogMaxSizeMB         int
	TLSEnabled           bool
	AllowedActions       []string
	AppAliases           map[string]string
}

// DefaultPath returns the default config path under the user home directory.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}

	return filepath.Join(home, ".donagent", "config.toml"), nil
}

// Load reads and validates a DonAgent config file.
func Load(path string) (Config, error) {
	values, err := parseFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		RabbitURL:            values["rabbit_url"],
		Username:             values["username"],
		Password:             values["password"],
		QueueName:            values["queue_name"],
		ExchangeName:         values["exchange_name"],
		EventContractVersion: defaultString(values["event_contract_version"], defaultContractVersion),
		LogPath:              expandPath(defaultString(values["log_path"], defaultLogPath())),
		LogMaxSizeMB:         defaultLogMaxSizeMB,
		AllowedActions:       parseStringList(values["allowed_actions"]),
	}

	appAliases, err := parseStringMap(values["app_aliases"])
	if err != nil {
		return Config{}, err
	}
	cfg.AppAliases = appAliases

	if raw := values["log_max_size_mb"]; raw != "" {
		size, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid log_max_size_mb: %w", err)
		}
		cfg.LogMaxSizeMB = size
	}

	if raw := values["tls_enabled"]; raw != "" {
		enabled, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("invalid tls_enabled: %w", err)
		}
		cfg.TLSEnabled = enabled
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks required settings and basic numeric constraints.
func (cfg Config) Validate() error {
	var missing []string

	if cfg.RabbitURL == "" {
		missing = append(missing, "rabbit_url")
	}
	if cfg.QueueName == "" {
		missing = append(missing, "queue_name")
	}
	if cfg.ExchangeName == "" {
		missing = append(missing, "exchange_name")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config fields: %s", strings.Join(missing, ", "))
	}
	if cfg.LogMaxSizeMB <= 0 {
		return errors.New("log_max_size_mb must be greater than zero")
	}
	if len(cfg.AppAliases) > 0 && !contains(cfg.AllowedActions, actionOpenApp) {
		return errors.New("app_aliases requires open_app in allowed_actions")
	}

	return nil
}

func parseFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file %q: %w", path, err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid config line %d", lineNumber)
		}

		values[strings.TrimSpace(key)] = trimValue(value)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	return values, nil
}

func trimValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "\r")
	value = strings.Trim(value, "\"")
	return value
}

func parseStringList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), "\"")
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

func parseStringMap(value string) (map[string]string, error) {
	items := parseStringList(value)
	if len(items) == 0 {
		return nil, nil
	}

	values := make(map[string]string, len(items))
	for _, item := range items {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			return nil, fmt.Errorf("invalid app_aliases item %q", item)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if err := validateAppAlias(key); err != nil {
			return nil, err
		}
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("duplicate app alias %q", key)
		}

		target, err := validateAppTarget(value)
		if err != nil {
			return nil, fmt.Errorf("invalid app alias %q: %w", key, err)
		}
		values[key] = target
	}

	return values, nil
}

func validateAppAlias(alias string) error {
	if !appAliasPattern.MatchString(alias) {
		return fmt.Errorf("invalid app alias %q", alias)
	}
	if alias == "." || alias == ".." || strings.Contains(alias, "..") {
		return fmt.Errorf("invalid app alias %q", alias)
	}

	return nil
}

func validateAppTarget(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("target is required")
	}

	value = filepath.Clean(expandPath(value))
	if !filepath.IsAbs(value) {
		return "", errors.New("target must be an absolute path")
	}

	switch strings.ToLower(filepath.Ext(value)) {
	case ".bat", ".cmd", ".ps1", ".vbs", ".js", ".wsf":
		return "", errors.New("script targets are not allowed")
	}

	return value, nil
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}

	return false
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".donagent", "events.log")
	}

	return filepath.Join(home, ".donagent", "events.log")
}

func expandPath(value string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		value = strings.Replace(value, "$HOME", home, 1)
	}

	return os.ExpandEnv(value)
}
