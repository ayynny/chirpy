package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux() // 1. create a new "router" (ServeMux) which decides what code (handler) to run for various paths

	s := http.Server{ // 2. set up http.Server struct
		Handler: mux,
		Addr:    ":8080",
	}

	apiConfig := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	fileServer := http.FileServer(http.Dir(".")) // returns a http.Handler

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app", fileServer))) // register handler for the pattern /app/, before server starts serving requests
	mux.HandleFunc("/healthz", myHandler)

	err := s.ListenAndServe() // listens for and serves requests
	if err != nil {
		return
	}

}
