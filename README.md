# Squirrel (personal expense tracker)

A microservice to manage personal expenses.

# Architecture

## Go packages

- `ent` (https://github.com/ent/ent) ORM
- `chi` (https://github.com/go-chi/chi) for routing
- `testify` (https://github.com/stretchr/testify) for writing and running unit tests
- `viper` (https://github.com/spf13/viper) manage multiple environments i.e dev, test, CAT, prod configurations
- `slog` go standard library for logging
- `swag` (https://github.com/swaggo/swag) generate RESTful API documentation.
- `golangci-lint` (https://github.com/golangci/golangci-lint) linter
- `cors` (https://github.com/go-chi/cors) CORS net/http middleware for Go
- `httprate` (https://github.com/go-chi/httprate) net/http rate limiter middleware
- `validator` (https://github.com/go-playground/validator) field validation, including Cross Field, Cross Struct, Map, Slice and Array diving

## Directory structure

```
/cmd/api/main.go
/config/
  config.yaml
  config.development.yaml
  config.test.yaml
  config.cat.yaml
  config.production.yaml
/internal/
  category/
    handler.go
    service.go
    repository.go
    model.go
  platform/
    http/
    middleware.go
    response.go
    router.go
  db/
    sqlite.go
/pkg/
  config/
```

## Configuration

The application uses `viper` for configuration management. It supports multiple environments via the `SQUIRREL_ENVIRONMENT` environment variable.

### Environment Variable: `SQUIRREL_ENVIRONMENT`
- `development` (default): Uses `config/config.development.yaml`
- `test`: Uses `config/config.test.yaml`
- `cat`: Uses `config/config.cat.yaml`
- `production`: Uses `config/config.production.yaml`

### Configuration Loading Order
1. **Defaults**: Hardcoded in `pkg/config/config.go`.
2. **Base Config**: `config/config.yaml`.
3. **Environment Overrides**: `config/config.{SQUIRREL_ENVIRONMENT}.yaml`.
4. **Environment Variables**: Overrides any of the above using `SQUIRREL_SERVER_ADDR` for `server.addr`, `SQUIRREL_SERVER_HOST` for `server.host`, etc.

| Variable | Description | Default |
|----------|-------------|---------|
| `SQUIRREL_ENVIRONMENT` | Deployment environment (`dev`, `production`) | `production` |
| `SQUIRREL_SERVER_ADDR` | Internal network address the server binds to | `:8081` |
| `SQUIRREL_SERVER_HOST` | Public-facing host/port for Swagger documentation | `localhost:8081` |
| `SQUIRREL_DATABASE_PATH` | Path to the SQLite database file | `data/squirrel.db` |
| `SQUIRREL_LOG_DIR` | Directory where log files are stored | `log` |

### Running on a different Port/Host
- To change the port the server listens on: set `SQUIRREL_SERVER_ADDR=:9090`.
- To change the address used in Swagger documentation: set `SQUIRREL_SERVER_HOST=api.example.com`.

## Code architecture

### Use directional dependencies:
HTTP → Service → Repository

#### Handler (Delivery Layer):
- Only HTTP concerns
- No business logic
```go
type CategoryHandler struct {
    svc *CategoryService
}
```

#### Service (Business Logic):
- Pure Go logic
- No HTTP, no SQL
```go
type CategoryService struct {
    repo *CategoryRepository
}
```

#### Repository (Persistence)
- DB logic only
- Implements interfaces


## Design patterns

- Dependancy injection (DI)
- Interface Segregation (Very Important in Go)
```
type CategoryWriter interface {
    Save(ctx context.Context, u Category) error
}

type CategoryReader interface {
    FindByID(ctx context.Context, id string) (Category, error)
}
```
- Error Handling Pattern (No Exceptions)
Sentinel + Wrapped Errors
```
var ErrCategoryNotFound = errors.New("category not found")

if err != nil {
    return fmt.Errorf("create category: %w", err)
}
```
Translate errors at the boundary (HTTP)
```
if errors.Is(err, ErrCategoryNotFound) {
    http.Error(w, "not found", http.StatusNotFound)
}
```
- Context Propagation (Mandatory)
```
func (s *service) Create(ctx context.Context, u Category) error
```

## Requirement

- Go v1.26
- SQLite v3.51.2

## Development

The project uses Docker and a Makefile for development.

- `make build`: Build the Docker images.
- `make up`: Start the containers in the background.
- `make down`: Stop and remove the containers.
- `make restart`: Restart the services.
- `make logs`: Follow the container logs.
- `make ps`: List the running containers.
- `make deps-upgrade`: Update all Go dependencies to their latest versions and run tests.
- `make go-upgrade version=1.x`: Upgrade the Go version across the project (go.mod, Dockerfile, Makefile) and rebuild.
- `make fmt`: Format code and organize imports using `goimports`.
- `make tidy`: Clean up `go.mod` and `go.sum` files.
- `make vet`: Run `go vet` for static analysis.
- `make generate`: Run `go generate` for all packages.
- `make vendor`: Create and update the `vendor` directory.
- `make coverage`: Generate an HTML test coverage report.
- `make coverage-view`: Open the HTML coverage report in your default browser.
- `make build-local`: Build the API binary on the host machine.
- `make help`: Display all available Makefile commands.
- `make test`: Run all Go tests inside the container.
- `make lint`: Run `golangci-lint` using a dedicated Docker image.
- `make swag`: Generate Swagger documentation.
- `make shell`: Open an interactive shell inside the API container.
- `make migrate-gen name=migration_name`: Generate a new versioned migration file.
- `make migrate-apply`: Apply all pending migrations to the database.
- `make clean`: Deep clean of containers, images, and volumes.

## Upgrading Go Version

To upgrade the Go version used in this project, run the following command with the desired version:

```bash
make go-upgrade version=1.27
```

This command automatically performs the following:
1. **Updates `go.mod`**: Changes the `go` version directive.
2. **Updates `Dockerfile`**: Changes the `FROM golang:<version>-alpine` base image.
3. **Updates `Makefile`**: Updates all `golang:<version>-alpine` image references used for tests and migrations.
4. **Rebuilds Images**: Runs `make build` to apply the changes.

## Database Migrations

This project uses **Ent** with **Atlas** for versioned migrations. Follow these steps when you need to change the database schema:

### 1. Create or Modify the Schema

#### To create a new table:
Initialize a new schema file:
```bash
docker run --rm -v $(pwd):/app -w /app golang:1.26-alpine go run -mod=mod entgo.io/ent/cmd/ent new TableName
```
Then define the fields in `ent/schema/tablename.go`.

#### To modify an existing table:
Update the schema definitions in the `ent/schema/` directory (e.g., `ent/schema/category.go`).

### 2. Generate Ent Code
After modifying the schema, regenerate the Ent runtime code:
```bash
make generate
```

### 3. Generate Migration Files
Generate a new SQL migration file by comparing your schema changes against an in-memory database:
```bash
make migrate-gen name=add_new_field_to_category
```
This will create new `.sql` files in `ent/migrate/migrations/`.

### 4. Apply Migrations
You can manually apply migrations to the database using:
```bash
make migrate-apply
```

Additionally, in the current development setup, the application automatically applies migrations on startup using `client.Schema.Create` in `internal/db/sqlite.go`. You can restart the service to trigger this:
```bash
make restart
```

## Database Persistence

The SQLite database is stored at `/app/data/squirrel.db` inside the container. This path is persisted using a bind mount to the local `./data` directory in the project root.

- **Host Path**: `./data/squirrel.db`
- **Container Path**: `/app/data/squirrel.db`
- **Environment Variables**:
  - `SQUIRREL_ENVIRONMENT`: Environment name (e.g., `development`).
  - `SQUIRREL_DATABASE_PATH`: Overrides the database path (e.g., `/app/data/squirrel.db`).

The database initialization is fully aligned with the Ent migration setup. On every startup, the application verifies the schema against the generated Ent code and applies any necessary changes to the SQLite file, ensuring the physical database always matches your versioned migration files.

## Database schema

### category

- ID - int - primary key - auto increment
- AppID - int
- UserID - int
- Name - varchar
- Status - int8
- Created at
- Updated at

### transaction

- ID - int - primary key - auto increment
- AppID - int
- UserID - int
- Amount - float
- Type - enum (income, expense)
- CategoryID - int (foreign key)
- Recurring - int8 (0 or 1, default 0)
- Created at
- Updated at

## Service URLs

By default, the services are available at:

- **API Gateway**: `http://<SERVER_HOST>`
- **Health Check**: `http://<SERVER_HOST>/health`
- **Swagger UI**: `http://<SERVER_HOST>/swagger/index.html`

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

## Rate Limiting

The API implements rate limiting using `httprate` middleware. By default, it is limited to **100 requests per minute per IP address**. This is configured in `internal/platform/http/router.go`.

## Logging

Structured logging is implemented project-wide using the standard library `log/slog`. Important events such as database initialization, category creation, authentication attempts, and errors are logged with appropriate levels (INFO, WARN, ERROR).

Logs are written to both **stdout** and to a file named `api.log` located in the `log/` directory.

## Persistence

The project uses Docker volumes to persist data and logs outside the container:
- **Database**: Stored in `./data/squirrel.db`.
- **Logs**: Stored in `./log/api.log`.

## Deployment

To ensure the Squirrel binary works correctly upon deployment, follow this checklist:

### 1. Shared Authentication Secret
Squirrel depends on the **Keeper** project for authentication. Both services **must** share the exact same JWT secret.
*   **Action**: Set the `SQUIRREL_AUTH_JWT_SECRET` environment variable on the server.
*   **Requirement**: This value must match the one used by the Keeper service. If they differ, Squirrel will reject all tokens issued by Keeper.

### 2. File System Permissions
The binary requires write access to its local directories for data and logs.
*   **Data Directory**: The binary needs write access to the `data/` directory to manage the SQLite database.
*   **Log Directory**: The binary will attempt to create a `log/` directory to write `api.log`.
*   **Persistence**: Ensure the `data/` directory is mounted to persistent storage (e.g., a Docker volume) to prevent data loss on restarts.

### 3. Runtime Configuration
Provide configuration via environment variables for production environments:
*   `SQUIRREL_ENVIRONMENT=production`: Loads production settings.
*   `SQUIRREL_SERVER_ADDR=:8081`: The port Squirrel listens on.
*   `SQUIRREL_AUTH_JWT_SECRET`: The shared secret key for JWT validation.
*   `SQUIRREL_DATABASE_PATH=data/squirrel.db`: Path to the SQLite file.

### 4. System Dependencies (CGO/SQLite)
Squirrel uses `go-sqlite3`, which requires **CGO**.
*   **Libraries**: The server must have `libc` and `sqlite` libraries installed (e.g., `apk add sqlite-libs` on Alpine).
*   **Build**: Ensure the binary was compiled with `CGO_ENABLED=1`.

### 5. Network Connectivity
*   **Port 8081**: Ensure the server firewall allows incoming traffic on the configured port.
*   **Keeper Service**: While Squirrel validates tokens locally, clients must be able to reach the Keeper service to obtain tokens.

## TODO

- Pagination
- Show total sum of expense and incode with its difference either in red/green
- A new bar chart type wise in transaction
