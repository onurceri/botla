# Multi-stage Dockerfile with CGO support for go-fitz PDF processing
# Stage 1: Build with Debian (glibc-based) for go-fitz compatibility
FROM golang:1.25-bookworm AS builder

# Install build dependencies required for CGO
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled and fitz tag
ENV CGO_ENABLED=1
RUN go build -tags fitz -ldflags="-s -w" -o server cmd/server/main.go

# Stage 2: Create minimal runtime image with glibc compatibility
FROM debian:bookworm-slim

# Install runtime dependencies including Chromium for go-rod web scraping
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    chromium \
    && rm -rf /var/lib/apt/lists/*

# Set Chromium path for go-rod (skip auto-download)
ENV ROD_BROWSER_PATH=/usr/bin/chromium
ENV SCRAPER_BROWSER_PATH=/usr/bin/chromium

WORKDIR /root/

# Copy the compiled binary from builder stage
COPY --from=builder /app/server .

# Copy database migrations
COPY --from=builder /app/db/migrations ./migrations

# Expose application port
EXPOSE 8080

# Run the server
CMD ["./server"]
