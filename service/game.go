package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"titles-mcp/clients"
)

const (
	// maxGameRetries is the number of different TMDB pages to try before giving up.
	maxGameRetries = 5
)

type TitleNameAndReleaseYear struct {
	Name        string `json:"movie_name"`
	ReleaseYear string `json:"release_year"`
}

func (t *TitleNameAndReleaseYear) UnmarshalJSON(data []byte) error {
	var aux struct {
		Name        string      `json:"movie_name"`
		ReleaseYear interface{} `json:"release_year"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.Name = aux.Name
	switch v := aux.ReleaseYear.(type) {
	case string:
		t.ReleaseYear = v
	case float64:
		t.ReleaseYear = fmt.Sprintf("%.0f", v)
	}
	return nil
}

type genre string

const (
	Action    genre = "Action"
	Adventure genre = "Adventure"
	Animation genre = "Animation"
	Comedy    genre = "Comedy"
	Crime     genre = "Crime"
	Drama     genre = "Drama"
	Fantasy   genre = "Fantasy"
	SciFi     genre = "Sci-Fi"
	Thriller  genre = "Thriller"
	Bollywood genre = "Bollywood"
)

var allGenres = []genre{
	Action, Adventure, Animation, Comedy, Crime, Drama, Fantasy, SciFi, Thriller, Bollywood,
}

var (
	exclusionList = make([]string, 0, 2000)
)

// NextGameMovie is the response for the movie game endpoint.
type NextGameMovie struct {
	Status   string                `json:"status"`
	Metadata *clients.TMDBMetadata `json:"metadata,omitempty"`
	Message  string                `json:"message,omitempty"`
}

var nextGameTitlesPool = NewGameNextTitlesPool()

func (t *titleService) GetNextGameMovie(ctx context.Context) (NextGameMovie, error) {
	err := t.updateExclusionList(ctx)
	if err != nil {
		return NextGameMovie{Status: "Error", Metadata: nil, Message: "Failed to update exclusion list"}, err
	}
	poolSize, err := nextGameTitlesPool.GetPoolSize(ctx)
	if err != nil {
		return NextGameMovie{Status: "Error", Metadata: nil, Message: "Failed to get current game titles pool size"}, err
	}
	if poolSize == 0 {
		t.updateNextGamesTitlesPool(ctx)
	} else if poolSize < 10 {
		go t.updateNextGamesTitlesPool(ctx)
	}
	return t.getNextMovieUsingLLM(ctx)
}

func (t *titleService) updateNextGamesTitlesPool(ctx context.Context) error {
	prompt := t.generatePrompt()
	llmOutput, err := t.llmClient.GenerateContent(ctx, prompt, nil)
	log.Print(llmOutput)
	if err != nil {
		return err
	}
	var rawMovies []TitleNameAndReleaseYear
	err = json.Unmarshal([]byte(llmOutput), &rawMovies)
	if err != nil {
		log.Printf("JSON unmarshalling failed for llm output: %v with error: %v", llmOutput, err)
		return err
	}
	return nextGameTitlesPool.AddTitles(ctx, rawMovies)
}

func (t *titleService) generatePrompt() string {
	genre := t.getGenreForNextMovie()

	prompt := fmt.Sprintf("Suggest 20 popular unique movie name and its release year of genre %s. Return in json format array [{\"movie_name\": \"nameOfTheMovie\", \"release_year\": \"ReleaseYear\"}]. Ensure release_year is a string. There should be no text before opening `[` and after closing `]`.", genre)
	prompt += "Do not include these movies: ["
	var builder strings.Builder
	builder.Grow(len(exclusionList) * 20)
	for _, title := range exclusionList {
		builder.WriteString("`")
		builder.WriteString(title)
		builder.WriteString("`,")
	}
	builder.WriteString("].")
	prompt += builder.String()
	return prompt
}

func (t *titleService) updateExclusionList(ctx context.Context) error {
	pageSize := 100
	for page := range 10 {
		titles, total, err := t.repository.GetAllTitles(ctx, page, pageSize)
		if err != nil {
			return err
		}
		for _, title := range titles {
			exclusionList = append(exclusionList, title.Name)
		}
		if total < int64(pageSize) {
			break
		}
	}
	return nil
}

func (t *titleService) getNextMovieUsingLLM(ctx context.Context) (NextGameMovie, error) {
	var nextMovie TitleNameAndReleaseYear
	var err error
	for {
		nextMovie, err = nextGameTitlesPool.GetNextTitle(ctx)
		if err != nil {
			return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
		}
		if nextMovie.Name != "" {
			break
		}
	}
	bestMatch, err := t.tmdbClient.SearchTMDB(nextMovie.Name, &nextMovie.ReleaseYear)
	if err != nil {
		return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
	}
	metadata, err := t.tmdbClient.FetchTMDBDetails(bestMatch.ID, bestMatch.MediaType)
	if err != nil {
		return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
	}
	metadata.NormalizeIMDbID()
	return NextGameMovie{Status: "Success", Metadata: metadata, Message: ""}, nil
}

func (t *titleService) getGenreForNextMovie() genre {
	return allGenres[rand.Intn(len(allGenres))]
}
