package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"titles-mcp/database"
	"titles-mcp/service"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"status": "error", "message": err.Error()})
}

func parsePagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 15
	q := r.URL.Query()
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 {
		pageSize = ps
	}
	return page, pageSize
}

func (h *handler) AddTitles(w http.ResponseWriter, r *http.Request) {
	var input service.AddTitlesInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := h.titleService.AddTitles(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, output)
}

func (h *handler) ListAllTitles(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	output, err := h.titleService.ListAllTitles(r.Context(), service.GetAllTitlesInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (h *handler) ListWatchedTitles(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	output, err := h.titleService.ListWatchedTitles(r.Context(), service.ListWatchedTitlesInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (h *handler) SearchTitles(w http.ResponseWriter, r *http.Request) {
	var input service.SearchTitlesInput

	// Parse query params to input
	q := r.URL.Query()

	if names := q["title_names"]; len(names) > 0 {
		input.TitleNames = &names
	}

	if from, to := q.Get("release_year_from"), q.Get("release_year_to"); from != "" || to != "" {
		input.ReleaseYearRange = &database.ReleaseYearRange{}
		if f, err := strconv.ParseUint(from, 10, 16); err == nil {
			input.ReleaseYearRange.From = uint16(f)
		}
		if t, err := strconv.ParseUint(to, 10, 16); err == nil {
			input.ReleaseYearRange.To = uint16(t)
		}
	}

	if watched := q.Get("watched"); watched != "" {
		wVal := watched == "true"
		input.Watched = &wVal
	}

	if wished := q.Get("wished"); wished != "" {
		wVal := wished == "true"
		input.Wished = &wVal
	}

	page, pageSize := parsePagination(r)
	input.Page = page
	input.PageSize = pageSize

	output, err := h.titleService.SearchTitles(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (h *handler) GetTitleDetails(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("movie_name")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "movie_name is required"})
		return
	}

	var input service.GetTitleDetailsInput
	input.MovieName = name

	if year := q.Get("release_year"); year != "" {
		input.ReleaseYear = &year
	}

	output, err := h.titleService.GetTitleDetails(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (h *handler) GetTitlesByIds(w http.ResponseWriter, r *http.Request) {
	var input service.GetTitlesByIdsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := h.titleService.GetTitlesByIds(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

// Helpers for extracting ID from path
func getID(r *http.Request) (uint, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	return uint(id), err
}

func (h *handler) MarkTitleAsWatched(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := h.titleService.MarkTitleAsWatched(r.Context(), service.MarkTitleAsWatchedInput{TitleId: id})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (h *handler) MarkTitleAsWished(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := h.titleService.MarkTitleAsWished(r.Context(), service.MarkTitleAsWishedInput{TitleId: id})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (h *handler) RemoveTitleFromWatched(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := h.titleService.RemoveTitleFromWatched(r.Context(), service.RemoveTitleFromWatchedInput{TitleId: id})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}

func (h *handler) RemoveTitleFromWished(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := h.titleService.RemoveTitleFromWished(r.Context(), service.RemoveTitleFromWishedInput{TitleId: id})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, output)
}
