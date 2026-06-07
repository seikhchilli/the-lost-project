package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"titles-mcp/clients"
)

const (
	// maxGameRetries is the number of different TMDB pages to try before giving up.
	maxGameRetries = 5
	llmMode        = true
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

type genre struct {
	tmdbId int
	name   string
}

var (
	genres = []genre{
		{name: "Action", tmdbId: 28},
		{name: "Adventure", tmdbId: 12},
		{name: "Animation", tmdbId: 16},
		{name: "Comedy", tmdbId: 35},
		{name: "Crime", tmdbId: 80},
		{name: "Drama", tmdbId: 18},
		{name: "Fantasy", tmdbId: 14},
		{name: "Sci-Fi", tmdbId: 878},
		{name: "Thriller", tmdbId: 53},
		{name: "Bollywood", tmdbId: -1},
	}
	exclusionList = []string{}
)

// NextGameMovie is the response for the movie game endpoint.
type NextGameMovie struct {
	Status   string                `json:"status"`
	Metadata *clients.TMDBMetadata `json:"metadata,omitempty"`
	Message  string                `json:"message,omitempty"`
}

var lastGameGenre int
var genreIterator int
var nextGameTitlesPool = NewGameNextTitlesPool()

func (t *titleService) GetNextGameMovie(ctx context.Context) (NextGameMovie, error) {
	err := t.updateExclusionList(ctx)
	if err != nil {
		return NextGameMovie{Status: "Error", Metadata: nil, Message: "Failed to update exclusion list"}, err
	}
	if llmMode {
		poolSize, err := nextGameTitlesPool.GetPoolSize(ctx)
		if err != nil {
			return NextGameMovie{Status: "Error", Metadata: nil, Message: "Failed to update exclusion list"}, err
		}
		if poolSize == 0 {
			t.updateNextGamesTitlesPoolUsingLLM(ctx)
		} else if poolSize < 20 {
			go t.updateNextGamesTitlesPoolUsingLLM(ctx)
		}
		return t.getNextMovieUsingLLM(ctx)
	} else {
		return t.getNextMovieUsingTMDB(ctx)
	}
}

func (t *titleService) updateNextGamesTitlesPoolUsingLLM(ctx context.Context) error {
	genre := t.getGenreForNextMovie()

	prompt := fmt.Sprintf("Suggest 20 popular unique movie name and its release year of genre %s. Return in json format array [{\"movie_name\": \"nameOfTheMovie\", \"release_year\": \"ReleaseYear\"}]. Ensure release_year is a string. There should be no text before opening `[` and after closing `]`. Do not include these movies: %v", genre.name, exclusionList)
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

func (t *titleService) updateExclusionList(ctx context.Context) error {
	exclusionList := []string{}
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

func (t *titleService) getNextMovieUsingTMDB(ctx context.Context) (NextGameMovie, error) {
	for range maxGameRetries {
		movies, err := t.tmdbClient.DiscoverGameMovies(t.getGenreForNextMovie().tmdbId)
		if err != nil {
			return NextGameMovie{}, err
		}
		if len(movies) == 0 {
			continue
		}

		movie, found, err := t.findUntrackedMovie(ctx, movies)
		if err != nil {
			return NextGameMovie{}, err
		}
		if found {
			return NextGameMovie{Status: "success", Metadata: movie}, nil
		}
	}

	return NextGameMovie{
		Status:  "error",
		Message: "Could not find new movies after several attempts. You may have tracked most popular titles!",
	}, nil
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
	var nextGenre genre
	for {
		nextGenre = genres[rand.Intn(len(genres))]
		if !llmMode && nextGenre.tmdbId == -1 {
			continue
		}
		return nextGenre
	}
}

// findUntrackedMovie filters out movies already in the DB, shuffles the
// remainder, and returns full metadata for the first untracked movie it finds.
func (t *titleService) findUntrackedMovie(ctx context.Context, movies []clients.TMDBSearchResult) (*clients.TMDBMetadata, bool, error) {
	ids := make([]string, len(movies))
	for i, m := range movies {
		ids[i] = fmt.Sprintf("%d", m.ID)
	}

	existingIds, err := t.repository.GetExistingTmdbIds(ctx, ids)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check existing titles: %w", err)
	}

	existing := make(map[string]bool, len(existingIds))
	for _, id := range existingIds {
		existing[id] = true
	}

	rand.Shuffle(len(movies), func(i, j int) {
		movies[i], movies[j] = movies[j], movies[i]
	})

	for _, m := range movies {
		if existing[fmt.Sprintf("%d", m.ID)] {
			continue
		}

		metadata, err := t.tmdbClient.FetchTMDBDetails(m.ID, "movie")
		if err != nil {
			return nil, false, err
		}
		metadata.NormalizeIMDbID()
		return metadata, true, nil
	}

	return nil, false, nil
}
