package config

import (
	"github.com/caarlos0/env/v10"
)

type Config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN,required"`
	DatabaseURL      string `env:"DATABASE_URL,required"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
