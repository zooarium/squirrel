// Package pagination provides a shared helper for parsing limit/offset
// query parameters on list endpoints.
package pagination

import (
	"net/http"
	"strconv"
)

// Default and bound values for list endpoints.
const (
	DefaultLimit = 50
	MaxLimit     = 500
	MinLimit     = 1

	DefaultOffset = 0
	MinOffset     = 0
)

// Pagination holds limit/offset values for list queries.
type Pagination struct {
	Limit  int
	Offset int
}

// Parse extracts and clamps the `limit` and `offset` query params.
// Non-integer or out-of-range values fall back to defaults / clamped bounds.
func Parse(r *http.Request) Pagination {
	p := Pagination{Limit: DefaultLimit, Offset: DefaultOffset}

	if val := r.URL.Query().Get("limit"); val != "" {
		if limit, err := strconv.Atoi(val); err == nil {
			switch {
			case limit < MinLimit:
				p.Limit = MinLimit
			case limit > MaxLimit:
				p.Limit = MaxLimit
			default:
				p.Limit = limit
			}
		}
	}

	if val := r.URL.Query().Get("offset"); val != "" {
		if offset, err := strconv.Atoi(val); err == nil {
			if offset < MinOffset {
				p.Offset = MinOffset
			} else {
				p.Offset = offset
			}
		}
	}

	return p
}
