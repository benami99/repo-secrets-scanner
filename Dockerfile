# Use official Go image as build stage
FROM golang:1.25.3-alpine AS builder

# Install git (needed if scanner fetches git repos)
RUN apk add --no-cache git

# Set workdir inside container
WORKDIR /app

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the Go app
RUN go build -o scanner ./cmd/main.go

# Use a small final image
FROM alpine:latest

# Install ca-certificates
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/scanner .

# Expose port
EXPOSE 8080

# Set env vars defaults (can override at runtime)
ENV HTTP_ADDR=:8080
ENV GITHUB_TOKEN=""

# Run the binary
ENTRYPOINT ["./scanner"]
