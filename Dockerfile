# Build stage
FROM golang:1.26-alpine AS builder

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
# CGO_ENABLED=1 is required for the standard SQLite driver
RUN CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" GOOS=linux go build -mod=vendor -a -installsuffix cgo -o vyaya ./cmd/api/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/vyaya .

# Expose port 8081
EXPOSE 8081

# Command to run the executable
CMD ["./vyaya"]
