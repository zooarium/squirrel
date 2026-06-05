package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by method, route and status.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds, labeled by method and route.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

// MetricsMiddleware records request counts and latencies for every request.
// The route label is resolved from the chi route pattern after the handler
// runs so that path parameters are aggregated (e.g. /transactions/{id}).
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		next.ServeHTTP(ww, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unknown"
		}

		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, route).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(ww.Status())).Inc()
	})
}
