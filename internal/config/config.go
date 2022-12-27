// Package config provides configuration for project
package config

import (
	"github.com/cristalhq/aconfig"
)

type Service struct {
	Host string `default:"" usage:"service host"`
	Port int    `default:"9090"  usage:"service port"`
}

type Database struct {
	Host        string `default:"localhost" usage:"database host"`
	Port        int    `default:"5432" usage:"database port"`
	User        string `default:"postgres" usage:"database user"`
	Password    string `default:"postgres" usage:"database password"`
	Name        string `default:"postgres" usage:"database name"`
	SSLMode     string `default:"disable" usage:"database ssl mode"`
	SSLCertPath string `default:"" usage:"database ssl cert path"`
}

type Frontend struct {
	Name string `default:"" usage:"frontend name"`
	Host string `default:"" usage:"public http-server address to bind"`
	Port int    `default:"8080" usage:"public http-server port to bind"`
}

// Config for http server
type Config struct {
	Debug  bool   `default:"false" usage:"enable debug prints"`
	Trace  bool   `default:"false" usage:"enable trace prints"`
	Format string `default:"text" usage:"log format: text, json"`

	Service  Service
	Database Database
	Frontend Frontend
}

// Load configuration
func Load() (*Config, error) {
	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		AllowUnknownFields: true,
		Files: []string{
			"config.toml",
			"config.json",
		},
	})
	err := loader.Load()
	if err != nil {
		return nil, err
	}
	err = cfg.Validate()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate configuration
func (c *Config) Validate() error {
	// implement your config validation here
	return nil
}
