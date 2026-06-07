package config

import (
	"log"
	"os"
	"path/filepath"

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

type LLMConfig struct {
	GeminiAPIKey string
}

type Config struct {
	DBConfig   DBConfig
	TMDBConfig TMDBConfig
	LLMConfig  LLMConfig
}

var AppConfig *Config

func LoadConfig() {
	// Try loading from current directory first
	err := godotenv.Load()
	if err != nil {
		// Fallback to executable's directory
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			_ = godotenv.Load(filepath.Join(exeDir, ".env"))
		}
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Fatal("DB_HOST is missing. Please set the required environment variables or provide a .env file.")
	}

	AppConfig = &Config{
		DBConfig: DBConfig{
			Port:     os.Getenv("DB_PORT"),
			Host:     dbHost,
			Name:     os.Getenv("DB_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
		TMDBConfig: TMDBConfig{
			APIKey: os.Getenv("TMDB_API_KEY"),
		},
		LLMConfig: LLMConfig{
			GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		},
	}
}
