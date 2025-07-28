package main

import (
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) countHandler(w http.ResponseWriter, r *http.Request) {
	count := int(cfg.fileserverHits.Load())
	str_count := strconv.Itoa(count)

	base := []byte("Hits: ")
	b := append(base, str_count...)
	w.Write(b)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = atomic.Int32{}
}

func main() {
	mux := http.NewServeMux() // 1. create a new "router" (ServeMux) which decides what code (handler) to run for various paths

	s := http.Server{ // 2. set up http.Server struct
		Handler: mux,
		Addr:    ":8080",
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	fileServer := http.FileServer(http.Dir(".")) // returns a http.Handler

	fileServerStripped := http.StripPrefix("/app", fileServer)

	// mux.Handle("/app/", http.StripPrefix("/app", fileServer))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerStripped)) //register handler before server starts serving requests

	mux.HandleFunc("/metrics", apiCfg.countHandler)
	mux.HandleFunc("/healthz", myHandler)
	mux.HandleFunc("/reset", apiCfg.resetHandler)

	err := s.ListenAndServe()
	if err != nil {
		return
	}

}
