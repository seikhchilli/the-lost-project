package handler

import (
	"net/http"
	"titles-mcp/service"
)

type Handler interface {
	Register(mux *http.ServeMux)
}

type handler struct {
	titleService service.TitleService
}

func NewHandler(titleService service.TitleService) Handler {
	return &handler{
		titleService: titleService,
	}
}

func (h *handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", Health)

	mux.HandleFunc("POST /api/titles", h.AddTitles)
	mux.HandleFunc("GET /api/titles", h.ListAllTitles)
	mux.HandleFunc("GET /api/titles/watched", h.ListWatchedTitles)
	mux.HandleFunc("GET /api/titles/search", h.SearchTitles)
	mux.HandleFunc("GET /api/titles/details", h.GetTitleDetails)
	mux.HandleFunc("GET /api/titles/game/next", h.GetNextGameMovie)
	mux.HandleFunc("POST /api/titles/bulk", h.GetTitlesByIds)

	mux.HandleFunc("PUT /api/titles/{id}/watch", h.MarkTitleAsWatched)
	mux.HandleFunc("DELETE /api/titles/{id}/watch", h.RemoveTitleFromWatched)

	mux.HandleFunc("PUT /api/titles/{id}/wish", h.MarkTitleAsWished)
	mux.HandleFunc("DELETE /api/titles/{id}/wish", h.RemoveTitleFromWished)

	mux.HandleFunc("DELETE /api/titles/{id}", h.DeleteTitle)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)
}
