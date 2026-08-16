FROM golang:1.26-bookworm AS builder

WORKDIR /app

# Download Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the binary
RUN go build -o titles-mcp main.go

# Stage 2: Setup runtime with Python and Playwright
FROM mcr.microsoft.com/playwright/python:v1.42.0-jammy

WORKDIR /app

# Install python dependencies for the scraper
RUN pip install --no-cache-dir playwright

# Copy the built Go binary and required files from the builder
COPY --from=builder /app/titles-mcp .
COPY --from=builder /app/static ./static
COPY --from=builder /app/yts_scraper.py .

# Expose the HTTP port
EXPOSE 3369

# Run the application in HTTP mode
ENTRYPOINT ["./titles-mcp", "-mode=http"]
