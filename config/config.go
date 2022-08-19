package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const AppName = "slink"

type Config struct {
	Environment   string `envconfig:"ENVIRONMENT" required:"true"`
	ProjectID     string `envconfig:"PROJECT_ID" required:"true"`
	Port          int    `envconfig:"PORT" required:"true"`
	Secret        string `envconfig:"SECRET" required:"true"`
	AlgoliaAppID  string `envconfig:"ALGOLIA_APP_ID" required:"true"`
	AlgoliaAPIKey string `envconfig:"ALGOLIA_API_KEY" required:"true"`
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
	return c.Environment == "staging"
}

func (c *Config) IsLocal() bool {
	return c.IsDevelopment()
}

func (c *Config) IsHosted() bool {
	return !c.IsLocal()
}

func (c *Config) Collection(name string) string {
	return fmt.Sprintf("%s.%s", c.Environment, name)
}

func (c *Config) NewLogger() (*zap.Logger, error) {
	if c.IsHosted() {
		return zap.NewProduction()
	} else {
		return zap.NewDevelopment()
	}
}
