CUR_DIR := $(notdir $(shell pwd))

.PHONY: build up down restart logs ps test fmt lint swag clean shell help tidy vet generate vendor coverage coverage-view build-local build-prod sql

# Docker Compose commands
build: vendor
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose restart

refresh: down swag build up

logs:
	docker-compose logs -f

ps:
	docker-compose ps

# Run tests inside the container
test: fmt
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) \
		-e CGO_ENABLED=1 \
		-e CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
		golang:1.26-alpine \
		sh -c "apk add --no-cache build-base && go test -mod=vendor -v ./..."

# Format code and manage imports
fmt:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) golang:1.26-alpine sh -c "go install golang.org/x/tools/cmd/goimports@latest && goimports -w ."

# Run linter using a docker container
lint:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) golangci/golangci-lint:latest golangci-lint run -v --modules-download-mode=vendor

# Generate Swagger documentation
swag:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) golang:latest sh -c "go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/api/main.go --parseDependency --parseInternal"

# Open a shell in the running api container
shell:
	docker-compose exec api sh

# Clean up go.mod and go.sum
tidy:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) golang:1.26-alpine sh -c "apk add git && go mod tidy"

# Run go vet for static analysis
vet:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) golang:1.26-alpine go vet -mod=vendor ./...

# Run go generate for code generation
generate:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) \
		golang:1.26-alpine \
		go generate -mod=vendor ./...

# Create vendor directory
vendor:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) golang:1.26-alpine sh -c "apk add git && go mod tidy && go mod vendor"

# Generate test coverage report
coverage:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) \
		-e CGO_ENABLED=1 \
		-e CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
		golang:1.26-alpine \
		sh -c "apk add --no-cache build-base && go test -mod=vendor -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"

# Open the coverage report in a browser
coverage-view:
	xdg-open coverage.html

# Build the binary locally (requires Go on host)
build-local:
	go build -o bin/api ./cmd/api/main.go

# Build the final binary for production (statically linked, stripped symbols)
build-prod: vendor
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) \
		-e CGO_ENABLED=1 \
		-e CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
		golang:1.26-alpine \
		sh -c "apk add --no-cache build-base && go build -mod=vendor -ldflags='-s -w -extldflags \"-static\"' -o bin/squirrel ./cmd/api/main.go"

# Update Go dependencies
deps-upgrade:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) \
		golang:1.26-alpine \
		sh -c "apk add git && go get -u ./... && go mod tidy && go mod vendor"
	$(MAKE) test

# Upgrade Go version across the project
go-upgrade:
	@if [ -z "$(version)" ]; then echo "Usage: make go-upgrade version=1.x"; exit 1; fi
	sed -i 's/^go [0-9.]*/go $(version)/' go.mod
	sed -i 's/^FROM golang:[0-9.]*-alpine/FROM golang:$(version)-alpine/' Dockerfile
	sed -i 's/golang:[0-9.]*-alpine/golang:$(version)-alpine/g' Makefile
	$(MAKE) build

# Database migrations
migrate-gen:
	docker run --rm -v $(shell pwd)/..:/workspace -w /workspace/$(CUR_DIR) \
		-e CGO_ENABLED=1 \
		-e CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
		golang:1.26-alpine \
		sh -c "apk add --no-cache build-base && go run -mod=vendor ent/migrate/main.go $(name)"


migrate-apply:
	docker-compose run --rm atlas migrate apply \
		--url "sqlite:///data/squirrel.db?_fk=1" \
		--dir "file://ent/migrate/migrations" \
		--allow-dirty

# Run SQL query against the database
sql:
	@if [ -z "$(query)" ]; then echo "Usage: make sql query=\"SQL_QUERY\""; exit 1; fi
	sqlite3 data/squirrel.db "$(query)"

# Clean up containers, images, and volumes
clean:
	docker-compose down --rmi all --volumes --remove-orphans

# Show help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build Docker images"
	@echo "  up            Start services in background"
	@echo "  down          Stop services"
	@echo "  restart       Restart services"
	@echo "  refresh       Rebuild and restart all services"
	@echo "  logs          Follow container logs"
	@echo "  ps            List running containers"
	@echo "  test          Run unit tests"
	@echo "  fmt           Format code (goimports)"
	@echo "  lint          Run linter"
	@echo "  swag          Generate Swagger docs"
	@echo "  tidy          Clean up go.mod"
	@echo "  vet           Run go vet"
	@echo "  generate      Run go generate"
	@echo "  vendor        Create vendor directory"
	@echo "  coverage      Generate test coverage report"
	@echo "  coverage-view Open test coverage report"
	@echo "  build-local   Build binary locally (requires Go)"
	@echo "  build-prod    Build final production binary (static)"
	@echo "  deps-upgrade  Upgrade Go dependencies"
	@echo "  go-upgrade    Upgrade Go version (use version=1.x)"
	@echo "  sql           Run SQL query (use query=...)"
	@echo "  clean         Deep clean containers/images"
	@echo "  help          Show this help message"
