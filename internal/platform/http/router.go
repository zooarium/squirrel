package http

import (
	"net/http"
	"time"

	_ "squirrel/docs" // Import generated docs
	"squirrel/internal/category"
	"squirrel/internal/transaction"
	"squirrel/pkg/config"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// NewRouter creates a new chi router with default middleware and application routes.
func NewRouter(cfg *config.Config, categoryHandler *category.Handler, transactionHandler *transaction.Handler, authMW func(http.Handler) http.Handler, dbDriver *entsql.Driver) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	r.Use(middleware.RequestID)
	r.Use(RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(MetricsMiddleware)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	))

	r.Get("/health", HealthHandler)
	r.Get("/ready", ReadyHandler(dbDriver))

	// Prometheus metrics endpoint, exempt from JWT auth.
	r.Handle("/metrics", promhttp.Handler())

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMW)
		r.Mount("/categories", categoryHandler.Routes())
		r.Mount("/transactions", transactionHandler.Routes())
	})

	return r
}
