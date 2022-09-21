package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const AppName = "slink"

type Config struct {
	Environment   string `envconfig:"ENVIRONMENT" required:"true" json:"environment"`
	ProjectID     string `envconfig:"PROJECT_ID" required:"true" json:"project_id"`
	Port          int    `envconfig:"PORT" required:"true" json:"port"`
	Secret        string `envconfig:"SECRET" required:"true" json:"-"`
	AlgoliaAppID  string `envconfig:"ALGOLIA_APP_ID" required:"true" json:"-"`
	AlgoliaAPIKey string `envconfig:"ALGOLIA_API_KEY" required:"true" json:"-"`
	AsyncTopic    string `envconfig:"ASYNC_TOPIC" json:"async_topic"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process(AppName, &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to process config")
	}

	return &cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsLocal() bool {
	return c.IsDevelopment()
}

func (c *Config) IsHosted() bool {
	return c.IsStaging() || c.IsProduction()
}

func (c *Config) Collection(name string) string {
	return fmt.Sprintf("%s.%s", c.Environment, name)
}

func (c *Config) NewLogger() (*zap.Logger, error) {
	if c.IsHosted() {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.MessageKey = "message"
		encoderConfig.LevelKey = "severity"
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig = encoderConfig

		return cfg.Build()
	} else {
		return zap.NewDevelopment()
	}
}
