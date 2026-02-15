# Vyaya Project Guide for Gemini CLI

This document provides a comprehensive overview of the Vyaya project, its architecture, development workflows, and technical details to assist Gemini CLI in understanding and maintaining the codebase.

## Project Overview
Vyaya is a microservice for expense management, providing RESTful APIs for category creation and management. It is built with Go, uses SQLite for persistence, and is containerized with Docker.

## Technical Stack
- **Language**: Go v1.26
- **Database**: SQLite v3.51.2
- **ORM**: [Ent](https://entgo.io/)
- **Router**: [chi](https://github.com/go-chi/chi)
- **Validation**: [validator v10](https://github.com/go-playground/validator)
- **Documentation**: Swagger (via `swag`)
- **Logging**: Structured logging with `log/slog`
- **Rate Limiting**: `httprate` (100 req/min per IP)
- **Migrations**: Atlas (integrated with Ent)

## Directory Structure
```text
/
├── cmd/
│   └── api/
│       └── main.go         # Application entry point
├── config/                 # YAML configuration files
│   ├── config.yaml         # Base configuration
│   ├── config.development.yaml
│   ├── config.test.yaml
│   ├── config.cat.yaml
│   └── config.production.yaml
├── internal/
│   ├── category/           # Category domain logic
│   │   ├── handler.go      # HTTP handlers
│   │   ├── service.go      # Business logic
│   │   ├── repository.go   # Data access logic
│   │   └── model.go        # Domain & Request/Response models
│   ├── platform/           # Cross-cutting concerns
│   │   ├── http/           # Router & Middleware
│   │   └── render/         # Standard API responses
│   └── db/
│       └── sqlite.go       # SQLite client initialization
├── ent/                    # Ent ORM generated code & schema
│   └── schema/
│       └── category.go     # Category database schema definition
├── pkg/                    # Shared packages
│   └── config/             # Configuration loader (viper)
├── data/                   # SQLite database file (persisted via volume)
├── log/                    # Application logs (persisted via volume)
├── docs/                   # Swagger documentation
├── Dockerfile              # Docker build configuration
├── docker-compose.yml      # Service orchestration
└── Makefile                # Development automation
```

## Configuration
The application uses `viper` for configuration management, supporting multiple environments via the `ENVIRONMENT` key in config files or the `ENVIRONMENT` environment variable.
- Configuration is loaded from `config/config.yaml` and merged with environment-specific overrides (e.g., `config.development.yaml`).
- Environment variables can override configuration values using the prefix-less, underscore-separated format (e.g., `SERVER_ADDR` for `server.addr`).

## Architecture & Design Patterns
- **Directional Dependencies**: HTTP (Handler) → Service → Repository.
- **Dependency Injection**: Used to decouple components and facilitate testing.
- **Interface Segregation**: Core logic is defined through interfaces.
- **Standardized Responses**: All API responses follow a consistent JSON format defined in `internal/platform/render`.
- **Context Propagation**: `context.Context` is passed through all layers for cancellation and timeouts.
- **Graceful Shutdown**: The API server handles `SIGINT` and `SIGTERM` for graceful termination.

## Naming Conventions
- **Packages**: Short, lowercase, single-word names (e.g., `category`, `auth`). Avoid underscores or mixedCaps.
- **Files**: Lowercase, using underscores only if necessary (e.g., `handler.go`, `service_test.go`).
- **Variables & Constants**: Use `CamelCase` (`MixedCaps` for exported, `mixedCaps` for unexported). Keep acronyms consistent (e.g., `categoryID`, `APIKey`).
- **Receivers**: Use short, consistent names (1-3 letters) representing the type (e.g., `func (u *Category) ...`).
- **Interfaces**: Name based on behavior, often ending in `-er` for single-action interfaces (e.g., `Reader`), or use descriptive nouns for domain logic (e.g., `Service`, `Repository`).
- **REST API Components**:
    - **Handlers**: `Handler` (e.g., `category.Handler`).
    - **Services**: `Service` (e.g., `category.Service`).
    - **Repositories**: `Repository` (e.g., `category.Repository`).
    - **Models**: Use generic names in packages and `[Action][Entity]Request/Response` for DTOs.
- **Database**: Table names and Ent schemas **must** be singular (e.g., `category`).

## Development Workflow

### Command Preference
Always prefer using `make` commands defined in the `Makefile` over direct `docker` or `go` commands. The `Makefile` ensures a consistent environment (using specific Go versions and dependencies) by running tools inside Docker containers.

### Mandatory Workflow for Every Change
To ensure codebase health and consistency, the following steps **must** be completed for every modification or new feature:
1. **Follow Naming Conventions**: Adhere to the project's naming conventions for packages, files, variables, and API components as defined in this document.
2.  **Structured Logging**: Add or update structured logging (using `slog`) to capture important events, business logic milestones, and error conditions.
3.  **Write Unit Tests**: Every new feature or bug fix must include corresponding unit tests (e.g., `*_test.go`).
4.  **Update Makefile**: If new development commands are required, add them to the `Makefile` and update the documentation accordingly.
5.  **Run Formatter**: Ensure code style and imports are consistent by running `make fmt`.
6.  **Run Linter**: Ensure code quality by running `make lint` after code and test changes.
7.  **Update Swagger Documentation**: If any API endpoints are added or modified, regenerate documentation using `make swag`.
8.  **Update README.md**: Ensure any new features, endpoints, or configuration changes are documented in `README.md`.
9.  **Update GEMINI.md**: Ensure this project guide is updated to reflect any changes in architecture, workflows, or documentation standards.
10.  **Run All Tests**: Verify that all tests pass by running `make test`.

### Common Commands (Makefile)
- `make build`: Build Docker images.
- `make up`: Start services in the background.
- `make down`: Stop services.
- `make deps-upgrade`: Update Go dependencies using a Docker container.
- `make fmt`: Format code and organize imports using `goimports`.
- `make tidy`: Clean up `go.mod` and `go.sum` files.
- `make vet`: Run `go vet` for static analysis.
- `make generate`: Run `go generate` for all packages.
- `make vendor`: Create and update the `vendor` directory.
- `make coverage`: Generate an HTML test coverage report.
- `make coverage-view`: Open the HTML coverage report in your default browser.
- `make build-local`: Build the API binary on the host machine.
- `make help`: Display all available Makefile commands.
- `make test`: Run unit tests in a fresh Go container.
- `make logs`: Follow container logs.
- `make swag`: Regenerate Swagger documentation.
- `make migrate-gen name=NAME`: Generate a new database migration.
- `make migrate-apply`: Apply pending migrations.
- `make sql query=QUERY`: Run a SQL query against the SQLite database.

### Database Migrations
1.  **Modify Schema**: Edit `ent/schema/category.go`.
2.  **Singular Table Names**: All database table names **must** be in singular format. Use `entsql.Annotation{Table: "singular_name"}` in the schema definition's `Annotations()` method.
3.  **Generate Code**: `make generate`
4.  **Generate Migration**: `make migrate-gen name=change_description`.
5.  **Apply**: `make migrate-apply` (or restart the app for auto-migration).

## Database Schema

### Category Table
| Field      | Type      | Description                          |
|------------|-----------|--------------------------------------|
| ID         | int       | Primary Key (Auto-increment)         |
| UserID     | int       | Owner user ID                        |
| Name       | string    | Category name                        |
| Status     | int8      | Status (1: Active, 0: Inactive)      |
| CreatedAt  | datetime  | Creation timestamp                   |
| UpdatedAt  | datetime  | Last update timestamp                |

### Transaction Table
| Field      | Type      | Description                          |
|------------|-----------|--------------------------------------|
| ID         | int       | Primary Key (Auto-increment)         |
| UserID     | int       | Owner user ID                        |
| Amount     | float     | Transaction amount                   |
| Type       | enum      | Type (income, expense)               |
| CategoryID | int       | Foreign Key to Category (Optional)   |
| CreatedAt  | datetime  | Creation timestamp                   |
| UpdatedAt  | datetime  | Last update timestamp                |

## API Endpoints
- `GET /health`: Check service health.
- `POST /categories`: Create a new category.
- `GET /categories`: List all categories.
- `GET /categories/{id}`: Get category by ID.
- `POST /categories/{id}`: Update category by ID.
- `DELETE /categories/{id}`: Delete category by ID.
- `POST /transactions`: Create a new transaction.
- `GET /transactions`: List all transactions.
- `GET /transactions/{id}`: Get transaction by ID.
- `POST /transactions/{id}`: Update transaction by ID.
- `DELETE /transactions/{id}`: Delete transaction by ID.
- `GET /swagger/*`: Swagger UI.

## Logging & Monitoring
- Logs are written to **stdout** and `./log/api.log`.
- Log format is JSON (structured).
- Levels: `INFO` for normal operations, `WARN` for client errors/auth failures, `ERROR` for system failures.

## Persistence & Volumes
- **Database**: `./data/vyaya.db` mapped to `/app/data/vyaya.db`.
- **Logs**: `./log/` mapped to `/app/log/`.
- **Environment**:
  - `GO_ENV`: Controls which configuration file is loaded (e.g., `development`, `production`).
  - `SERVER_ADDR`: Overrides the server address (defaults to `:8081`).
  - `DB_PATH`: Overrides the database path (defaults to `data/vyaya.db`).
  - `LOG_DIR`: Overrides the log directory (defaults to `log`).
