package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr                string
	DatabaseURL             string
	Loglevel                string
	ShutdownTimeout         time.Duration
	DatabaseMaxConns        int32
	DatabaseMinConns        int32
	DatabaseMaxConnLifetime time.Duration
	DatabaseConnectTimeout  time.Duration
	DatabaseQueryTimeout    time.Duration
	OperationTimeout        time.Duration
	HTTPReadHeaderTimeout   time.Duration
	HTTPReadTimeout         time.Duration
	HTTPWriteTimeout        time.Duration
	HTTPIdleTimeout         time.Duration
}

func LoadInt32(name string, fallback int32) (int32, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	data, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", name, err)
	}
	if data < 0 {
		return -1, fmt.Errorf("%s must be nonnegative", name)
	}
	return int32(data), nil
}

func LoadTime(name string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	data, err := time.ParseDuration(raw)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", name, err)
	}
	if data <= 0 {
		return -1, fmt.Errorf("%s must be positive", name)
	}
	return data, nil
}

func Load() (Config, error) {
	var cfg Config

	cfg.HTTPAddr = os.Getenv("HTTP_ADDR")
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	cfg.Loglevel = os.Getenv("LOG_LEVEL")
	if cfg.Loglevel == "" {
		cfg.Loglevel = "info"
	}
	switch cfg.Loglevel {
	case "debug", "info", "warn", "error":
	default:
		return Config{}, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error")
	}

	timeout, err := LoadTime("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout = timeout

	max_conns, err := LoadInt32("DATABASE_MAX_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	cfg.DatabaseMaxConns = max_conns

	min_conns, err := LoadInt32("DATABASE_MIN_CONNS", 2)
	if err != nil {
		return Config{}, err
	}
	cfg.DatabaseMinConns = min_conns

	if max_conns < min_conns {
		return Config{}, fmt.Errorf("max connection number should be more than min connection number")
	}
	if max_conns == 0 {
		return Config{}, fmt.Errorf("DATABASE_MAX_CONNS must be greater than zero")
	}

	max_lifetime, err := LoadTime("DATABASE_MAX_CONN_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	cfg.DatabaseMaxConnLifetime = max_lifetime

	conn_timeout, err := LoadTime("DATABASE_CONNECT_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.DatabaseConnectTimeout = conn_timeout

	query_timeout, err := LoadTime("DATABASE_QUERY_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.DatabaseQueryTimeout = query_timeout

	operationTimeout, err := LoadTime("OPERATION_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.OperationTimeout = operationTimeout

	read_header_timeout, err := LoadTime("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPReadHeaderTimeout = read_header_timeout

	read_timeout, err := LoadTime("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPReadTimeout = read_timeout

	write_timeout, err := LoadTime("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPWriteTimeout = write_timeout

	idle_timeout, err := LoadTime("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPIdleTimeout = idle_timeout
	return cfg, nil
}
