package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := writeConfig(t, `
rabbit_url = "amqp://localhost:5672"
queue_name = "donagent.events"
exchange_name = "donagent"
log_max_size_mb = 12
tls_enabled = true
allowed_actions = ["open_url", "open_app"]
app_aliases = ["browser=` + testAppTarget() + `"]
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RabbitURL != "amqp://localhost:5672" {
		t.Fatalf("RabbitURL = %q", cfg.RabbitURL)
	}
	if cfg.EventContractVersion != "1" {
		t.Fatalf("EventContractVersion = %q", cfg.EventContractVersion)
	}
	if cfg.LogMaxSizeMB != 12 {
		t.Fatalf("LogMaxSizeMB = %d", cfg.LogMaxSizeMB)
	}
	if cfg.LogPath == "" || cfg.LogPath == "$HOME/.donagent/events.log" {
		t.Fatalf("LogPath = %q", cfg.LogPath)
	}
	if !cfg.TLSEnabled {
		t.Fatal("TLSEnabled = false")
	}
	if len(cfg.AllowedActions) != 2 {
		t.Fatalf("AllowedActions length = %d", len(cfg.AllowedActions))
	}
	if cfg.AppAliases["browser"] != filepath.Clean(testAppTarget()) {
		t.Fatalf("AppAliases[browser] = %q", cfg.AppAliases["browser"])
	}
}

func TestLoadMissingRequiredFields(t *testing.T) {
	path := writeConfig(t, `
queue_name = "donagent.events"
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestLoadRejectsMalformedAppAlias(t *testing.T) {
	path := writeConfig(t, `
rabbit_url = "amqp://localhost:5672"
queue_name = "donagent.events"
exchange_name = "donagent"
allowed_actions = ["open_app"]
app_aliases = ["browser"]
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestLoadRejectsDuplicateAppAlias(t *testing.T) {
	path := writeConfig(t, `
rabbit_url = "amqp://localhost:5672"
queue_name = "donagent.events"
exchange_name = "donagent"
allowed_actions = ["open_app"]
app_aliases = ["browser=` + testAppTarget() + `", "browser=` + testAppTarget() + `"]
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestLoadRejectsInvalidAppAliasName(t *testing.T) {
	path := writeConfig(t, `
rabbit_url = "amqp://localhost:5672"
queue_name = "donagent.events"
exchange_name = "donagent"
allowed_actions = ["open_app"]
app_aliases = ["../browser=` + testAppTarget() + `"]
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestLoadRejectsRelativeAppAliasTarget(t *testing.T) {
	path := writeConfig(t, `
rabbit_url = "amqp://localhost:5672"
queue_name = "donagent.events"
exchange_name = "donagent"
allowed_actions = ["open_app"]
app_aliases = ["browser=firefox"]
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestLoadRejectsScriptAppAliasTarget(t *testing.T) {
	path := writeConfig(t, `
rabbit_url = "amqp://localhost:5672"
queue_name = "donagent.events"
exchange_name = "donagent"
allowed_actions = ["open_app"]
app_aliases = ["script=` + testScriptTarget() + `"]
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestLoadRejectsAppAliasesWhenOpenAppIsNotAllowed(t *testing.T) {
	path := writeConfig(t, `
rabbit_url = "amqp://localhost:5672"
queue_name = "donagent.events"
exchange_name = "donagent"
allowed_actions = ["open_url"]
app_aliases = ["browser=` + testAppTarget() + `"]
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}

func testAppTarget() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\notepad.exe`
	}

	return "/usr/bin/nano"
}

func testScriptTarget() string {
	if runtime.GOOS == "windows" {
		return `C:\Temp\script.ps1`
	}

	return "/tmp/script.ps1"
}
