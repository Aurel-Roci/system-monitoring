package models

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment        string        `env:"ENV" envDefault:"local"`
	CollectionInterval time.Duration `env:"COLLECTION_INTERVAL" envDefault:"5s"`

	Database   DatabaseConfig   `envPrefix:"DATABASE_"`
	Monitoring MonitoringConfig `envPrefix:"MONITORING_"`
	Logging    LoggingConfig    `envPrefix:"LOG_"`
}

type DatabaseConfig struct {
	Type string `env:"TYPE" envDefault:"victoriametrics"`
	URL  string `env:"URL" envDefault:"localhost"`
	Port int    `env:"PORT" envDefault:"8428"`
}

type MonitoringConfig struct {
	EnableCPU         bool `env:"ENABLE_CPU_MONITORING" envDefault:"true"`
	EnableMemory      bool `env:"ENABLE_MEMORY_MONITORING" envDefault:"true"`
	EnableTemperature bool `env:"ENABLE_TEMPERATURE_MONITORING" envDefault:"true"`
	EnableLoad        bool `env:"ENABLE_LOAD_MONITORING" envDefault:"true"`
}

type LoggingConfig struct {
	Level string `env:"LEVEL" envDefault:"info"`
	File  string `env:"FILE" envDefault:"stdout"`
}

func (c *Config) GetVictoriaMetricsURL() string {
	return fmt.Sprintf("%s:%d", c.Database.URL, c.Database.Port)
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Info("No .env file found, using system environment variables")
	}

	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, err
	}
	return config, nil
}
