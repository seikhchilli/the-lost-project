package clients

import "titles-mcp/config"

type TMDB interface {
}

type tmdb struct {
	apiKey string
}

func NewTMDB() TMDB {
	return &tmdb{
		apiKey: config.AppConfig.TMDBConfig.APIKey,
	}
}
