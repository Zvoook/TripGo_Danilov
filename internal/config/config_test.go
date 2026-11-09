package config

import (
	"strings"
	"testing"
	"time"
)

func cleanConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"HTTP_ADDR", "LOG_LEVEL", "SHUTDOWN_TIMEOUT", "DATABASE_URL",
		"DATABASE_MAX_CONNS", "DATABASE_MIN_CONNS", "DATABASE_MAX_CONN_LIFETIME",
		"DATABASE_CONNECT_TIMEOUT", "DATABASE_QUERY_TIMEOUT", "OPERATION_TIMEOUT",
		"HTTP_READ_HEADER_TIMEOUT", "HTTP_READ_TIMEOUT", "HTTP_WRITE_TIMEOUT", "HTTP_IDLE_TIMEOUT",
	} {
		t.Setenv(name, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	cleanConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/tripgo")
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := Config{
		HTTPAddr: ":8080", DatabaseURL: "postgres://localhost/tripgo", Loglevel: "info",
		ShutdownTimeout: 10 * time.Second, DatabaseMaxConns: 10, DatabaseMinConns: 2,
		DatabaseMaxConnLifetime: 30 * time.Minute, DatabaseConnectTimeout: 5 * time.Second,
		DatabaseQueryTimeout: 3 * time.Second, OperationTimeout: 10 * time.Second,
		HTTPReadHeaderTimeout: 5 * time.Second, HTTPReadTimeout: 10 * time.Second,
		HTTPWriteTimeout: 15 * time.Second, HTTPIdleTimeout: 60 * time.Second,
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadInvalidSettings(t *testing.T) {
	for _, tt := range []struct{ name, value string }{
		{"DATABASE_URL", ""},
		{"LOG_LEVEL", "verbose"},
		{"DATABASE_MAX_CONNS", "0"},
		{"DATABASE_MIN_CONNS", "11"},
		{"DATABASE_MAX_CONNS", "abc"},
		{"DATABASE_MIN_CONNS", "-1"},
		{"DATABASE_QUERY_TIMEOUT", "0s"},
		{"OPERATION_TIMEOUT", "-1s"},
		{"HTTP_READ_TIMEOUT", "invalid"},
	} {
		t.Run(tt.name+"="+tt.value, func(t *testing.T) {
			cleanConfigEnv(t)
			t.Setenv("DATABASE_URL", "postgres://localhost/tripgo")
			t.Setenv(tt.name, tt.value)
			_, err := Load()
			if err == nil {
				t.Fatal("expected configuration error")
			}
			if tt.name == "LOG_LEVEL" && !strings.Contains(err.Error(), "debug, info, warn or error") {
				t.Fatal(err)
			}
		})
	}
}

func TestLoadOverrides(t *testing.T) {
	cleanConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/tripgo")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_MIN_CONNS", "0")
	t.Setenv("DATABASE_QUERY_TIMEOUT", "1s")
	t.Setenv("OPERATION_TIMEOUT", "8s")
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.HTTPAddr != ":9090" || got.DatabaseMinConns != 0 || got.DatabaseQueryTimeout != time.Second || got.OperationTimeout != 8*time.Second {
		t.Fatalf("overrides ignored: %+v", got)
	}
}
