package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"titles-mcp/config"
)

const (
	// maxGameRetries is the number of different TMDB pages to try before giving up.
	maxGameRetries = 5
	llmMode        = true
)

// NextGameMovie is the response for the movie game endpoint.
type NextGameMovie struct {
	Status   string        `json:"status"`
	Metadata *TMDBMetadata `json:"metadata,omitempty"`
	Message  string        `json:"message,omitempty"`
}

type tmdbPopularResponse struct {
	Results []tmdbSearchResult `json:"results"`
}

var lastGameGenre int
var genreIterator int
var nextGameMovieCache = make(map[GetTitleDetailsInput]struct{})

// NextGameMovie picks a random popular movie from TMDB that the user has not
// yet tracked. It retries across multiple attempts with randomized genres and decades.
func (t *titleService) GetNextGameMovie(ctx context.Context) (NextGameMovie, error) {
	if llmMode {
		return t.getNextMovieUsingLLM()
	}
	apiKey := config.AppConfig.TMDBConfig.APIKey
	if apiKey == "" {
		return NextGameMovie{
			Status:  "error",
			Message: "TMDB_API_KEY not found in environment.",
		}, nil
	}

	for range maxGameRetries {
		movies, err := t.discoverGameMovies(apiKey)
		if err != nil {
			return NextGameMovie{}, err
		}
		if len(movies) == 0 {
			continue
		}

		movie, found, err := t.findUntrackedMovie(ctx, apiKey, movies)
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

func (t *titleService) getNextMovieUsingLLM() (NextGameMovie, error) {
	ctx := context.Background()
	genres := []string{"Action", "Adventure", "Animation", "Comedy", "Crime", "Drama", "Fantasy", "Sci-Fi", "Thriller"}
	genre := genres[genreIterator]
	genreIterator++
	genreIterator %= len(genres)

	// Time Hopping: Pick a random decade from 1980s to 2020s
	// startYear := 1980 + (rand.Intn(5) * 10)
	// endYear := startYear + 9
	// dateGte := fmt.Sprintf("%d-01-01", startYear)
	// dateLte := fmt.Sprintf("%d-12-31", endYear)
	if len(nextGameMovieCache) <= 20 {
		exclusionList := []string{}
		pageSize := 100
		for page := 0; page < 4; page++ {
			titles, total, err := t.repository.GetAllTitles(ctx, page, pageSize)
			if err != nil {
				return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
			}
			for _, title := range titles {
				exclusionList = append(exclusionList, title.Name)
			}
			if total < int64(pageSize) {
				break
			}
		}
		prompt := fmt.Sprintf("Suggest 40 popular unique movie name and its release year of genre %s. Return in json format array [{\"movie_name\": \"nameOfTheMovie\", \"release_year\": \"ReleaseYear\"}]. Ensure release_year is a string. There should be no text before opening `[` and after closing `]`. Do not include these movies: %v", genre, exclusionList)
		llmOutput, err := t.llmClient.GenerateContent(ctx, prompt, nil)
		log.Print(llmOutput)
		if err != nil {
			return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
		}
		var rawMovies []struct {
			MovieName   string      `json:"movie_name"`
			ReleaseYear interface{} `json:"release_year"`
		}
		err = json.Unmarshal([]byte(llmOutput), &rawMovies)
		for _, rm := range rawMovies {
			var yearStr string
			if y, ok := rm.ReleaseYear.(string); ok {
				yearStr = y
			} else if y, ok := rm.ReleaseYear.(float64); ok {
				yearStr = fmt.Sprintf("%.0f", y)
			}
			var yearPtr *string
			if yearStr != "" {
				yearPtr = &yearStr
			}
			nextGameMovieCache[GetTitleDetailsInput{MovieName: rm.MovieName, ReleaseYear: yearPtr}] = struct{}{}
		}
		if err != nil {
			log.Printf("JSON unmarshalling failed for llm output: %v with error: %v", llmOutput, err)
			return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
		}
	}

	for nextMovie, _ := range nextGameMovieCache {
		bestMatch, err := t.searchTMDB(config.AppConfig.TMDBConfig.APIKey, nextMovie.MovieName, nextMovie.ReleaseYear)
		if err != nil {
			return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
		}
		metadata, err := t.fetchTMDBDetails(config.AppConfig.TMDBConfig.APIKey, bestMatch.ID, bestMatch.MediaType)
		if err != nil {
			return NextGameMovie{Status: "Error", Metadata: nil, Message: "error"}, err
		}
		metadata.NormalizeIMDbID()
		delete(nextGameMovieCache, nextMovie)
		return NextGameMovie{Status: "Success", Metadata: metadata, Message: ""}, nil
	}
	return NextGameMovie{
		Status:  "error",
		Message: "Could not find new movies. Seems to be some issue with LLM!",
	}, nil
}

// discoverGameMovies fetches a page of movies using TMDB's discover endpoint,
// applying randomized genre rotation, time hopping, and quality filters.
func (t *titleService) discoverGameMovies(apiKey string) ([]tmdbSearchResult, error) {
	// Popular genres: Action(28), Adventure(12), Animation(16), Comedy(35), Crime(80), Drama(18), Fantasy(14), Sci-Fi(878), Thriller(53)
	genres := []int{28, 12, 16, 35, 80, 18, 14, 878, 53}

	// Genre Rotation: Pick a random genre that isn't the same as the last one
	var genre int
	for {
		genre = genres[rand.Intn(len(genres))]
		if genre != lastGameGenre {
			break
		}
	}
	lastGameGenre = genre

	// Time Hopping: Pick a random decade from 1980s to 2020s
	startYear := 1980 + (rand.Intn(5) * 10)
	endYear := startYear + 9
	dateGte := fmt.Sprintf("%d-01-01", startYear)
	dateLte := fmt.Sprintf("%d-12-31", endYear)

	// Due to strict filters, there won't be many pages. Pick page 1 or 2.
	page := rand.Intn(2) + 1

	url := fmt.Sprintf("%s/discover/movie?api_key=%s&page=%d&with_genres=%d&primary_release_date.gte=%s&primary_release_date.lte=%s&vote_count.gte=3000&vote_average.gte=7.0&sort_by=popularity.desc",
		tmdbBaseURL, apiKey, page, genre, dateGte, dateLte)

	var resp tmdbPopularResponse
	if err := t.getAndDecodeJSON(url, &resp); err != nil {
		return nil, nil
	}
	return resp.Results, nil
}

// findUntrackedMovie filters out movies already in the DB, shuffles the
// remainder, and returns full metadata for the first untracked movie it finds.
func (t *titleService) findUntrackedMovie(ctx context.Context, apiKey string, movies []tmdbSearchResult) (*TMDBMetadata, bool, error) {
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

		metadata, err := t.fetchTMDBDetails(apiKey, m.ID, "movie")
		if err != nil {
			return nil, false, err
		}
		metadata.NormalizeIMDbID()
		return metadata, true, nil
	}

	return nil, false, nil
}
