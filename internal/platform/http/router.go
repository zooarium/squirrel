package http

import (
	"time"

	_ "vyaya/docs" // Import generated docs
	"vyaya/internal/category"
	"vyaya/internal/transaction"

	"dvarapala/pkg/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter creates a new chi router with default middleware and application routes.
func NewRouter(categoryHandler *category.Handler, transactionHandler *transaction.Handler, jwtManager *auth.JWTManager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	))

	r.Get("/health", HealthHandler)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtManager))
		r.Mount("/categories", categoryHandler.Routes())
		r.Mount("/transactions", transactionHandler.Routes())
	})

	return r
}
