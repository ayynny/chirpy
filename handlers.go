package main

import (
	"fmt"
	"net/http"
)

// Tracks how many times a request has been made to the fileserver
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Simple health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Returns how many times Chirpy has been visited
func (cfg *apiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	htmlTemplate := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	sprintfOutput := fmt.Sprintf(htmlTemplate, cfg.fileserverHits.Load())
	w.Write([]byte(sprintfOutput))
}
