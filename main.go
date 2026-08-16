package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total de requests HTTP",
	}, []string{"method", "path", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duración de requests HTTP",
		Buckets: prometheus.DefBuckets,
	}, []string{"path"})

	startTime = time.Now()
)

func instrument(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		defer timer.ObserveDuration()
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next(rw, r)
		httpRequests.WithLabelValues(r.Method, path, http.StatusText(rw.status)).Inc()
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := os.Getenv("APP_NAME")
	if app == "" {
		app = "test-ddgil"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", instrument("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"app":     app,
			"uptime":  time.Since(startTime).String(),
		})
	}))

	mux.HandleFunc("/api/info", instrument("/api/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"app":     app,
			"team":    "cloud-army",
			"version": "0.1.0",
			"env":     os.Getenv("ENV"),
		})
	}))

	mux.Handle("/metrics", promhttp.Handler())

	log.Printf("%s listening on :%s", app, port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
