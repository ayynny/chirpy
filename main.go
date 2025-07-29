package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

type Middlewares struct {
	handlers []func() http.Handler
}

func middlewareHeaderGetter() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header
		fmt.Print("header")
		for key, value := range header {
			for _, word := range value {
				keyValue := key + ": " + word + "\n"
				w.Write([]byte(keyValue))
				fmt.Println(keyValue)
			}
		}
	})
}

func (cfg *apiConfig) middlewareMetricsInc() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Print("vistor!")
		cfg.fileserverHits.Add(1)
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
	cfg.fileserverHits.Store(0)
}

func (m *Middlewares) add(next func() http.Handler) {
	m.handlers = append(m.handlers, next)
}

func (m *Middlewares) applyMiddlewares(next http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, middlewareHandlers := range m.handlers {
			if middlewareHandlers != nil {
				middlewareHandlers()
			}
		}
		next.ServeHTTP(w, r)
	}
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

	middlewareStack := Middlewares{
		handlers: make([]func() http.Handler, 2),
	}

	middlewareStack.add(apiCfg.middlewareMetricsInc)
	middlewareStack.add(middlewareHeaderGetter)

	mux.HandleFunc("/app/", middlewareStack.applyMiddlewares(fileServerStripped)) //register handler before server starts serving requests

	mux.HandleFunc("GET /metrics", apiCfg.countHandler)
	mux.HandleFunc("GET /healthz", myHandler)
	mux.HandleFunc("POST /reset", apiCfg.resetHandler)
	// mux.Handle("/app/", apiCfg.middlewareHeaderGetter(fileServerStripped))

	err := s.ListenAndServe()
	if err != nil {
		return
	}

}
