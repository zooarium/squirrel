# syntax=docker/dockerfile:1
# Build stage
ARG GO_VERSION=1.26.3
FROM golang:${GO_VERSION}-alpine AS builder

# Install build dependencies for CGO (needed for SQLite)
RUN apk add --no-cache build-base

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
COPY go.sum ./

# Copy vendor directory
COPY vendor/ vendor/

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for the standard SQLite driver.
# BuildKit cache mount persists the Go build cache across image builds (incremental recompiles).
# Dropped `-a -installsuffix cgo` — those forced a full rebuild and defeated the cache.
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" GOOS=linux \
    go build -mod=vendor -o squirrel ./cmd/api/main.go

# Final stage
FROM alpine:3.22

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/squirrel .
COPY CHANGELOG.md .

# Expose port 8081
EXPOSE 8081

# Command to run the executable
CMD ["./squirrel"]
