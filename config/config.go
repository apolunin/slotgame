package config

import "fmt"

const (
	TextFormat LogFormat = `text`
	JSONFormat LogFormat = `json`
)

type (
	LogFormat string

	SlotGameConfig struct {
		HTTPCfg   HTTPConfig
		DBCfg     DatabaseConfig
		LogCfg    LogConfig
		JWTSecret string `env:"JWT_SECRET"`
	}

	HTTPConfig struct {
		Port         int `env:"HTTP_PORT" envDefault:"8080"`
		RateLimitCfg RateLimitConfig
	}

	RateLimitConfig struct {
		Rate  int `env:"LIMIT_RATE" envDefault:"1000"`
		Burst int `env:"LIMIT_BURST" envDefault:"5"`
	}

	DatabaseConfig struct {
		Host     string `env:"DB_HOST"`
		Port     int    `env:"DB_PORT"`
		Database string `env:"DB_DATABASE"`
		User     string `env:"DB_USR"`
		Password string `env:"DB_PWD"`
	}

	LogConfig struct {
		Level     string    `env:"LOG_LEVEL" envDefault:"info"`
		Format    LogFormat `env:"LOG_FORMAT" envDefault:"text"`
		AddSource bool      `env:"ADD_SOURCE" envDefault:"true"`
	}
)

func (cfg DatabaseConfig) URL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}
