package service

import (
	"context"
	"fmt"
	"math/rand"
	"titles-mcp/config"
)

const (
	// maxGameRetries is the number of different TMDB pages to try before giving up.
	maxGameRetries = 5
)

// GetNextGameMovieOutput is the response for the movie game endpoint.
type GetNextGameMovieOutput struct {
	Status   string        `json:"status"`
	Metadata *TMDBMetadata `json:"metadata,omitempty"`
	Message  string        `json:"message,omitempty"`
}

type tmdbPopularResponse struct {
	Results []tmdbSearchResult `json:"results"`
}

var lastGameGenre int

// GetNextGameMovie picks a random popular movie from TMDB that the user has not
// yet tracked. It retries across multiple attempts with randomized genres and decades.
func (t *titleService) GetNextGameMovie(ctx context.Context) (GetNextGameMovieOutput, error) {
	apiKey := config.AppConfig.TMDBConfig.APIKey
	if apiKey == "" {
		return GetNextGameMovieOutput{
			Status:  "error",
			Message: "TMDB_API_KEY not found in environment.",
		}, nil
	}

	for range maxGameRetries {
		movies, err := t.discoverGameMovies(apiKey)
		if err != nil {
			return GetNextGameMovieOutput{}, err
		}
		if len(movies) == 0 {
			continue
		}

		movie, found, err := t.findUntrackedMovie(ctx, apiKey, movies)
		if err != nil {
			return GetNextGameMovieOutput{}, err
		}
		if found {
			return GetNextGameMovieOutput{Status: "success", Metadata: movie}, nil
		}
	}

	return GetNextGameMovieOutput{
		Status:  "error",
		Message: "Could not find new movies after several attempts. You may have tracked most popular titles!",
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

