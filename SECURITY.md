# Security Notes

Known issues to address before production hardening.

## Critical

### Hardcoded JWT Secret
- **File**: `pkg/config/config.go`, `config/config.yaml`
- **Issue**: JWT secret hardcoded as default value and committed to repo. Forged tokens possible if source is exposed. Squirrel vendors `keeper/pkg/auth` — same secret used across both services.
- **Fix**: Remove default. Require `SQUIRREL_AUTH_JWT_SECRET` env var. Fail on boot if missing.

## Medium

### Internal Error Messages Leaked to Clients
- **File**: `internal/transaction/handler.go`, `internal/category/handler.go`
- **Issue**: `render.Error(w, http.StatusInternalServerError, err.Error())` — ent constraint errors and internal details returned to caller.
- **Fix**: Return generic `"internal server error"` to client. Keep `slog.Error` with full detail server-side only.

### No Request Body Size Limit
- **File**: All handlers using `json.NewDecoder(r.Body).Decode(&req)`
- **Issue**: No cap on request body size. Large payloads cause unbounded memory usage.
- **Fix**: Add `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` (1MB) before decoding.

## Low

### Weak JWT Claims
- **File**: Inherited from `keeper/pkg/auth` (vendored)
- **Issue**: Only `ExpiresAt` set. No `Issuer`, `Audience`, or `Subject`. Tokens issued by keeper are accepted by squirrel with no audience restriction.
- **Fix**: Add `Audience: ["squirrel"]` validation in squirrel's auth middleware, or create a squirrel-specific JWT manager with audience enforcement.

### Missing Indexes on sqrl_transaction
- **File**: `ent/schema/transaction.go`
- **Issue**: Queries filter by `app_id`, `user_id`, `category_id`, `dated`, `type` — none have explicit indexes. Filter-heavy `buildFilteredQuery` does full table scans.
- **Fix**: Add `Indexes()` method with composite indexes on `(app_id, user_id)` and `(app_id, dated)`.

## Performance (Non-Security)

### No Pagination on List Endpoints
- `GET /categories` and `GET /transactions` return entire filtered result set. Add `limit`/`offset` query params.

### Stats Fires 3 DB Queries per List Request
- `service.List()` calls `repo.List()` + `repo.GetStats()`. `GetStats()` runs 2 queries (aggregation + top 10). Consider caching or combining on high-traffic endpoints.
