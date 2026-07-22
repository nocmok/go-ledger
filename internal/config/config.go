package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBConfig DBConfig
}

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	Name     string `env:"DB_NAME"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	MaxConn  int    `env:"DB_MAX_CONN"`
}

func Load() (Config, error) {
	config := Config{}
	if err := cleanenv.ReadConfig(".env", &config); err != nil {
		return Config{}, err
	}
	return config, nil
}
