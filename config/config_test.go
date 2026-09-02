package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()

	if cfg.Port == "" {
		t.Error("expected non-empty port")
	}
	if cfg.JWTAccessExpiry == 0 {
		t.Error("expected non-zero JWT access expiry")
	}
	if cfg.JWTRefreshExpiry == 0 {
		t.Error("expected non-zero JWT refresh expiry")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "production")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("ENV")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Env != "production" {
		t.Errorf("expected env production, got %s", cfg.Env)
	}
}

func TestParseDuration_Valid(t *testing.T) {
	d := parseDuration("30m")
	if d != 30*time.Minute {
		t.Errorf("expected 30m, got %v", d)
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	d := parseDuration("invalid")
	if d != 15*time.Minute {
		t.Errorf("expected default 15m, got %v", d)
	}
}

func TestParseBool_True(t *testing.T) {
	if !parseBool("true") {
		t.Error("expected true")
	}
}

func TestParseBool_False(t *testing.T) {
	if parseBool("false") {
		t.Error("expected false")
	}
}

func TestParseBool_Invalid(t *testing.T) {
	if parseBool("invalid") {
		t.Error("expected false for invalid input")
	}
}

func TestParseInt64_Valid(t *testing.T) {
	n := parseInt64("1024")
	if n != 1024 {
		t.Errorf("expected 1024, got %d", n)
	}
}

func TestParseInt64_Invalid(t *testing.T) {
	n := parseInt64("invalid")
	if n != 52428800 {
		t.Errorf("expected default 52428800, got %d", n)
	}
}

func TestGetEnv_WithFallback(t *testing.T) {
	val := getEnv("NON_EXISTENT_VAR_XYZ", "fallback")
	if val != "fallback" {
		t.Errorf("expected fallback, got %s", val)
	}
}

func TestGetEnv_FromEnv(t *testing.T) {
	os.Setenv("TEST_VAR_XYZ", "hello")
	defer os.Unsetenv("TEST_VAR_XYZ")

	val := getEnv("TEST_VAR_XYZ", "fallback")
	if val != "hello" {
		t.Errorf("expected hello, got %s", val)
	}
}
