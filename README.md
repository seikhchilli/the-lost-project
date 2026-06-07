# 🎬 The Lost Project (titles-mcp)

A robust Go-based Model Context Protocol (MCP) server and Web Application designed to seamlessly manage a database of media titles. It provides both MCP tools for AI agents and a beautiful, responsive Single Page Application (SPA) for end users to browse, search, and discover movies and TV shows.

## ✨ Features

- 🍿 **Watched & Wishlist Tracking**: Manage your media library easily. Add movies or TV shows to your Watched list or Wishlist.
- 🎮 **Interactive Movie Game**: A Tinder-style "Game" tab that fetches random popular movies directly from TMDB, filters out titles you already have in bulk, and lets you quickly Skip, Wishlist, or mark them as Watched.
- 📱 **Rich SPA Interface**: A fully responsive, glassmorphic UI that adapts perfectly to desktop, tablet, and smartphone screens. Mobile users get a highly optimized, compact view.
- 🚀 **Backend TMDB Proxying**: TMDB API integration securely runs entirely on the backend, proxying requests and filtering results via the database to minimize load.
- 🤖 **AI-Powered Discovery**: Integrated LLM backing to provide highly contextual next movie suggestions for the Interactive Game.

## 🏗 Architecture

The project follows a modular, clean architecture:
- **`main.go`**: Entry point, initializes configuration, database, static file serving, HTTP routes, and the MCP server.
- **`config/`**: Handles environment variables (including `TMDB_API_KEY`, `GEMINI_API_KEY`) and database connection settings.
- **`database/`**: Implements the Repository pattern (`repository.go`), Houses GORM models (`models/`), and custom errors (`sentinel/`). Also includes Redis-based caching.
- **`handler/`**: RESTful HTTP handlers for powering the web application frontend.
- **`service/`**: Core business logic decoupling the handlers/tools from the database.
- **`clients/`**: Standalone API clients (`tmdb`, `llm`) cleanly decoupled from the service layer.
- **`tools/`**: MCP tool implementations using dedicated DTOs to cleanly expose backend functionality to AI clients.
- **`static/`**: The frontend Single Page Application (HTML, CSS, Vanilla JS) providing a dynamic UI without page reloads.

## 🛠 Prerequisites

- [Go 1.25.0+](https://go.dev/)
- A PostgreSQL / SQLite database (or whatever your environment provides via GORM)
- Redis server (used for caching title pools)
- A `.env` file in the root directory:

```env
DB_HOST=your_host
DB_PORT=your_port
DB_NAME=your_db_name
DB_USER=your_user
DB_PASSWORD=your_password
TMDB_API_KEY=your_tmdb_api_key
GEMINI_API_KEY=your_gemini_api_key
```

## 🚀 Getting Started

### Build
To compile the project into an executable:
```bash
go build -o titles-mcp main.go
```

### Run
The application can run in two modes:

**1. Web Server Mode (HTTP)**  
Runs the rich web application interface on `localhost:3369`:
```bash
go run main.go -mode=http
```

**2. Model Context Protocol Mode (MCP)**  
Runs over Stdio providing native tools for AI agents:
```bash
go run main.go -mode=mcp
```

### Tests
To run all the backend tests:
```bash
go test ./...
```

## 📝 Development Conventions

- **Repository Pattern**: Data access must be abstracted through the `database.Repository` interface.
- **Service Layer**: HTTP endpoints and MCP tools must utilize the `TitleService` for unified logic.
- **DTO Pattern for Tools**: Tools and Handlers should use dedicated input structs instead of passing GORM models directly.
- **Error Handling**: Use the `database/sentinel` package for cross-layer error comparison.
- **Database Migrations**: Handled automatically using GORM's `AutoMigrate`. Ensure new models are added to `database/db.go`.
