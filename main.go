package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"titles-mcp/clients"
	"titles-mcp/config"
	"titles-mcp/database"
	"titles-mcp/handler"
	"titles-mcp/service"
	"titles-mcp/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	mode := flag.String("mode", "mcp", "Run it in either mcp or http mode")
	flag.Parse()
	config.LoadConfig()
	db = database.NewDb()

	switch *mode {
	case "mcp":
		startMCPServer()
	case "http":
		slog.Info("sarting http server")
		startHTTPServer()
	default:
		slog.Info("Expected mode as mcp or http.", "Found:", *mode)
	}
}

func startMCPServer() {
	server := mcp.NewServer(&mcp.Implementation{Name: "title-mcp", Version: "v1.0.0"}, nil)

	titleService := service.NewTitleService(database.NewRepository(db), clients.NewTMDB())
	titleTool := tools.NewTitleTool(titleService)
	titleTool.Register(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func startHTTPServer() {
	mux := http.NewServeMux()

	titleService := service.NewTitleService(database.NewRepository(db), clients.NewTMDB())
	handlers := handler.NewHandler(titleService)
	handlers.Register(mux)
	srv := &http.Server{
		Addr:              ":3369",
		Handler:           mux,
		ReadTimeout:       5 * time.Second,   // Max time to read the entire request
		WriteTimeout:      10 * time.Second,  // Max time to write the response
		IdleTimeout:       120 * time.Second, // Max time for connections using TCP Keep-Alive
		ReadHeaderTimeout: 5 * time.Second,   // Max time to read headers (Mitigates Slowloris attacks)
	}

	go func() {
		slog.Info("HTTP Server is starting", "port", srv.Addr)
		// ErrServerClosed is expected when we gracefully shut down, so we ignore it
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// 5. Setup Graceful Shutdown
	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)

	// SIGINT (Ctrl+C) and SIGTERM (Docker/Kubernetes shutdown)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-quit
	slog.Info("Shutdown signal received, initiating graceful shutdown...")

	// Create a context with a timeout for the shutdown process
	// This gives active requests 30 seconds to finish before forcing the server down
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited properly")
}
