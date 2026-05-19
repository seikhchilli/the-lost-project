package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Port     string
	Host     string
	Name     string
	User     string
	Password string
}

type TMDBConfig struct {
	APIKey string
}

type Config struct {
	DBConfig   DBConfig
	TMDBConfig TMDBConfig
}

var AppConfig *Config

func LoadConfig() {
	_ = godotenv.Load()

	AppConfig = &Config{
		DBConfig: DBConfig{
			Port:     os.Getenv("DB_PORT"),
			Host:     os.Getenv("DB_HOST"),
			Name:     os.Getenv("DB_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
		TMDBConfig: TMDBConfig{
			APIKey: os.Getenv("TMDB_API_KEY"),
		},
	}
}
