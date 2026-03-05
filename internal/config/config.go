package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/F3dosik/Hofermart/internal/logger"
	"github.com/caarlos0/env/v6"
)

const (
	defaultServiceAddress = "localhost:8081"
	defaultDatabaseURI    = "postgresql://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable"
	defaultAccrualAddress = "localhost:8080"
	defaultLogLevel       = string(logger.ModeDevelopment)
	defaultWorkerCount    = 3
	defaultPollInterval   = 2 * time.Second
	defaultMaxDelay       = 5 * time.Minute
	defaultJWTSecret      = "secret"
)

type Config struct {
	ServiceAddress string
	DatabaseURI    string
	AccrualAddress string
	LogLevel       string
	JWTSecret      string
	WorkerCount    int
	PollInterval   time.Duration
	MaxDelay       time.Duration
}

func LoadConfig() (*Config, error) {
	envConf := parseEnvConfig()
	flagConf := parseFlagConfig()

	config := mergeConfigs(envConf, flagConf)
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

type envConfig struct {
	ServiceAddress string        `env:"RUN_ADDRESS"`
	DatabaseURI    string        `env:"DATABASE_URI"`
	AccrualAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel       string        `env:"LOG_LEVEL"`
	JWTSecret      string        `env:"JWT_SECRET"`
	WorkerCount    int           `env:"WORKER_COUNT"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	MaxDelay       time.Duration `env:"MAX_DELAY"`
}

func parseEnvConfig() *envConfig {
	var config envConfig
	err := env.Parse(&config)
	if err != nil {
		log.Printf("Warning: failed to parse env config: %v\n", err)
	}

	return &config
}

type flagConfig struct {
	ServiceAddress string
	DatabaseURI    string
	AccrualAddress string
	LogLevel       string
	WorkerCount    int
	PollInterval   time.Duration
	MaxDelay       time.Duration
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.ServiceAddress, "a", defaultServiceAddress, "service launch address and port")
	flag.StringVar(&config.DatabaseURI, "d", defaultDatabaseURI, "database connection address")
	flag.StringVar(&config.AccrualAddress, "r", defaultAccrualAddress, "address of the accrual calculation system")
	flag.StringVar(&config.LogLevel, "l", defaultLogLevel, "logging levels")
	flag.IntVar(&config.WorkerCount, "worker-count", defaultWorkerCount, "number of workers for accrual polling")
	flag.DurationVar(&config.PollInterval, "poll-interval", defaultPollInterval, "base interval between accrual polling attempts")
	flag.DurationVar(&config.MaxDelay, "max-delay", defaultMaxDelay, "maximum delay between accrual polling attempts")

	flag.Parse()

	return &config
}

func mergeConfigs(envConfig *envConfig, flagConfig *flagConfig) *Config {
	var config Config
	config.ServiceAddress = resolveString(envConfig.ServiceAddress, flagConfig.ServiceAddress)
	config.DatabaseURI = resolveString(envConfig.DatabaseURI, flagConfig.DatabaseURI)
	config.AccrualAddress = resolveString(envConfig.AccrualAddress, flagConfig.AccrualAddress)
	config.LogLevel = resolveString(envConfig.LogLevel, flagConfig.LogLevel)
	config.JWTSecret = defaultJWTSecret
	config.WorkerCount = resolveInt(envConfig.WorkerCount, flagConfig.WorkerCount)
	config.PollInterval = resolveDuration(envConfig.PollInterval, flagConfig.PollInterval)
	config.MaxDelay = resolveDuration(envConfig.MaxDelay, flagConfig.MaxDelay)
	return &config
}

func resolveString(envVal, flagVal string) string {
	if envVal != "" {
		return envVal
	}
	return flagVal
}

func resolveInt(envVal, flagVal int) int {
	if envVal != 0 {
		return envVal
	}
	return flagVal
}

func resolveDuration(envVal, flagVal time.Duration) time.Duration {
	if envVal != 0 {
		return envVal
	}
	return flagVal
}

func (c *Config) Validate() error {
	host, port, err := net.SplitHostPort(c.ServiceAddress)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("invalid service address")
	}

	host, port, err = net.SplitHostPort(c.AccrualAddress)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("invalid accrual address")
	}

	if c.DatabaseURI == "" {
		return fmt.Errorf("database address can't be empty")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret can't be empty")
	}

	if c.WorkerCount < 0 {
		return fmt.Errorf("worker count can't be less than zero")
	}

	switch c.LogLevel {
	case string(logger.ModeDevelopment), string(logger.ModeProduction):
	default:
		return fmt.Errorf("invalid log mode: %s, allowed: development, production", c.LogLevel)
	}

	return nil
}
