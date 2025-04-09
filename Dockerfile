# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
# CGO_ENABLED=0 for static linking
# -ldflags="-s -w" to strip debug information and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/server ./cmd/server/main.go

# Development stage with hot reload (used in docker-compose for development)
FROM golang:1.24-alpine AS dev

# Install development tools
RUN apk add --no-cache git

# Install CompileDaemon for hot reloading
RUN go install github.com/githubnemo/CompileDaemon@latest

# Set working directory
WORKDIR /app

# Copy air configuration
COPY .air.toml .

# The source code will be mounted from the host in docker-compose

# Runtime stage
FROM gcr.io/distroless/static-debian11 AS production

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/server /app/server

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/server"]
