package main

import (
	"context"
	"log/slog"
	"testing"
)

// TestLogHandler is a custom handler to capture log messages for testing
type TestLogHandler struct {
	logs []TestLogRecord
}

type TestLogRecord struct {
	Level   slog.Level
	Message string
}

func (h *TestLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

func (h *TestLogHandler) Handle(_ context.Context, r slog.Record) error {
	h.logs = append(h.logs, TestLogRecord{
		Level:   r.Level,
		Message: r.Message,
	})
	return nil
}

func (h *TestLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TestLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *TestLogHandler) hasWarn() bool {
	for _, log := range h.logs {
		if log.Level == slog.LevelWarn {
			return true
		}
	}
	return false
}

func TestParseConfigs_DefaultConfig(t *testing.T) {
	handler := &TestLogHandler{}
	logger := slog.New(handler)
	app := NewApplication(logger)

	config := app.ParseConfigs([]string{})

	// Assert default values
	if config.env != "development" {
		t.Errorf("expected env to be 'development', got '%s'", config.env)
	}

	if config.server.port != 8080 {
		t.Errorf("expected port to be 8080, got %d", config.server.port)
	}
}

func TestParseConfigs_EnvFlag(t *testing.T) {
	handler := &TestLogHandler{}
	logger := slog.New(handler)
	app := NewApplication(logger)

	config := app.ParseConfigs([]string{"-env", "production"})

	if config.env != "production" {
		t.Errorf("expected env to be 'production', got '%s'", config.env)
	}

	// Port should still be default
	if config.server.port != 8080 {
		t.Errorf("expected port to be 8080, got %d", config.server.port)
	}
}

func TestParseConfigs_PortFlag(t *testing.T) {
	handler := &TestLogHandler{}
	logger := slog.New(handler)
	app := NewApplication(logger)

	config := app.ParseConfigs([]string{"-port", "3000"})

	if config.server.port != 3000 {
		t.Errorf("expected port to be 3000, got %d", config.server.port)
	}

	// Env should still be default
	if config.env != "development" {
		t.Errorf("expected env to be 'development', got '%s'", config.env)
	}
}

func TestParseConfigs_BothFlags(t *testing.T) {
	handler := &TestLogHandler{}
	logger := slog.New(handler)
	app := NewApplication(logger)

	config := app.ParseConfigs([]string{"-env", "production", "-port", "9000"})

	if config.env != "production" {
		t.Errorf("expected env to be 'production', got '%s'", config.env)
	}

	if config.server.port != 9000 {
		t.Errorf("expected port to be 9000, got %d", config.server.port)
	}
}

func TestParseConfigs_InvalidEnv(t *testing.T) {
	handler := &TestLogHandler{}
	logger := slog.New(handler)
	app := NewApplication(logger)

	config := app.ParseConfigs([]string{"-env", "staging"})

	// Should default to development
	if config.env != "development" {
		t.Errorf("expected env to be 'development' for invalid env, got '%s'", config.env)
	}

	// Should have logged a warning
	if !handler.hasWarn() {
		t.Error("expected warning log for invalid env value")
	}
}

func TestParseConfigs_InvalidEnvProduction(t *testing.T) {
	handler := &TestLogHandler{}
	logger := slog.New(handler)
	app := NewApplication(logger)

	// Test that "development" is valid (no warning)
	config := app.ParseConfigs([]string{"-env", "development"})

	if config.env != "development" {
		t.Errorf("expected env to be 'development', got '%s'", config.env)
	}

	if handler.hasWarn() {
		t.Error("should not log warning for valid 'development' env")
	}
}
