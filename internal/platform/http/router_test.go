package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"vyaya/internal/category"
	"vyaya/internal/transaction"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	// Create handlers with nil services - just testing routing to public endpoints
	catHandler := category.NewCategoryHandler(nil)
	txHandler := transaction.NewTransactionHandler(nil)
	router := NewRouter(catHandler, txHandler)

	tests := []struct {
		name           string
		method         string
		url            string
		wantStatusCode int
	}{
		{"Health check", "GET", "/health", http.StatusOK},
		{"NotFound check", "GET", "/non-existent", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}
