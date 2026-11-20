package config

import (
	"strings"
	"testing"
	"time"
)

const (
	logLevelEnv     = "LOG_LEVEL"
	databaseMaxEnv  = "DATABASE_MAX_CONNS"
	databaseMinEnv  = "DATABASE_MIN_CONNS"
	databaseURLEnv  = "DATABASE_URL"
	testDatabaseURL = "postgres://localhost/tripgo"
	pushHTTPURLEnv  = "PUSH_HTTP_URL"
	testPushHTTPURL = "http://localhost:23809"
)

func cleanConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"HTTP_ADDR", logLevelEnv, "SHUTDOWN_TIMEOUT", databaseURLEnv,
		databaseMaxEnv, databaseMinEnv, "DATABASE_MAX_CONN_LIFETIME",
		"DATABASE_CONNECT_TIMEOUT", "DATABASE_QUERY_TIMEOUT", "OPERATION_TIMEOUT",
		"HTTP_READ_HEADER_TIMEOUT", "HTTP_READ_TIMEOUT", "HTTP_WRITE_TIMEOUT", "HTTP_IDLE_TIMEOUT",
		"POSITION_CHECK_INTERVAL", "POSITION_STALE_AFTER", "POSITION_WORKERS", "POSITION_QUEUE_SIZE", "PUSH_REQUEST_TIMEOUT",
	} {
		t.Setenv(name, "")
	}
	t.Setenv(databaseURLEnv, testDatabaseURL)
	t.Setenv(pushHTTPURLEnv, testPushHTTPURL)
}

func TestLoadDefaults(t *testing.T) {
	cleanConfigEnv(t)
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := Config{
		HTTPAddr: ":8080", DatabaseURL: testDatabaseURL, Loglevel: "info",
		ShutdownTimeout: 10 * time.Second, DatabaseMaxConns: 10, DatabaseMinConns: 2,
		DatabaseMaxConnLifetime: 30 * time.Minute, DatabaseConnectTimeout: 5 * time.Second,
		DatabaseQueryTimeout: 3 * time.Second, OperationTimeout: 10 * time.Second,
		HTTPReadHeaderTimeout: 5 * time.Second, HTTPReadTimeout: 10 * time.Second,
		HTTPWriteTimeout: 15 * time.Second, HTTPIdleTimeout: 60 * time.Second,
		PositionCheckInterval: 30 * time.Second, PositionStaleAfter: time.Minute,
		PositionWorkers: 4, PositionQueueSize: 100, PushHTTPURL: testPushHTTPURL,
		PushRequestTimeout: time.Second,
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadInvalidSettings(t *testing.T) {
	for _, tt := range []struct{ name, value string }{
		{databaseURLEnv, ""},
		{pushHTTPURLEnv, ""},
		{logLevelEnv, "verbose"},
		{databaseMaxEnv, "0"},
		{databaseMinEnv, "11"},
		{databaseMaxEnv, "abc"},
		{databaseMinEnv, "-1"},
		{"DATABASE_QUERY_TIMEOUT", "0s"},
		{"OPERATION_TIMEOUT", "-1s"},
		{"HTTP_READ_TIMEOUT", "invalid"},
	} {
		t.Run(tt.name+"="+tt.value, func(t *testing.T) {
			cleanConfigEnv(t)
			t.Setenv(tt.name, tt.value)
			_, err := Load()
			if err == nil {
				t.Fatal("expected configuration error")
			}
			if tt.name == logLevelEnv && !strings.Contains(err.Error(), "debug, info, warn or error") {
				t.Fatal(err)
			}
		})
	}
}

func TestLoadOverrides(t *testing.T) {
	cleanConfigEnv(t)
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv(databaseMinEnv, "0")
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
