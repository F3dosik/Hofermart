package config

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/F3dosik/Hofermart/internal/logger"
	"github.com/caarlos0/env/v6"
)

const (
	defaultServiceAddress = "localhost:8081"
	defaultAccrualAddress = "localhost:8080"
	defaultLogLevel       = string(logger.ModeDevelopment)
)

type Config struct {
	ServiceAddress string
	DatabaseURI    string
	AccrualAddress string
	LogLevel       string
}

func LoadConfig() (*Config, error) {
	envConf := parseEnvConfig()
	flagConf := parseFlagConfig()

	var config *Config
	config = mergeConfigs(envConf, flagConf)
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

type envConfig struct {
	ServiceAddress string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`
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
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.ServiceAddress, "a", defaultServiceAddress, "service launch address and port")
	flag.StringVar(&config.DatabaseURI, "d", "", "database connection address")
	flag.StringVar(&config.AccrualAddress, "r", defaultAccrualAddress, "address of the accrual calculation system")
	flag.StringVar(&config.LogLevel, "l", defaultLogLevel, "logging levels")

	flag.Parse()

	return &config
}

func mergeConfigs(envConfig *envConfig, flagConfig *flagConfig) *Config {
	var config Config
	config.ServiceAddress = resolveString(envConfig.ServiceAddress, flagConfig.ServiceAddress)
	config.DatabaseURI = resolveString(envConfig.DatabaseURI, flagConfig.DatabaseURI)
	config.AccrualAddress = resolveString(envConfig.AccrualAddress, flagConfig.AccrualAddress)
	config.LogLevel = resolveString(envConfig.LogLevel, flagConfig.LogLevel)
	return &config
}

func resolveString(envVal, flagVal string) string {
	if envVal != "" {
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

	switch c.LogLevel {
	case string(logger.ModeDevelopment), string(logger.ModeProduction):
	default:
		return fmt.Errorf("invalid log mode: %s, allowed: development, production", c.LogLevel)
	}

	return nil
}
