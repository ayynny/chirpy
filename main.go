package main

import (
	"encoding/json"
	"fmt"
	"log"
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

// Simple health check handler
func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Metrics handler
func (cfg *apiConfig) countHandler(w http.ResponseWriter, r *http.Request) {
	count := int(cfg.fileserverHits.Load())
	w.Write([]byte("Hits: " + strconv.Itoa(count)))
}

// Reset handler
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Reset done"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	htmlTemplate := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	sprintfOutput := fmt.Sprintf(htmlTemplate, cfg.fileserverHits.Load())
	w.Write([]byte(sprintfOutput))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type request struct { // request body
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)      // create a new decoder to read the request json data
	requestInstance := request{}            // create instance of the request body
	err := decoder.Decode(&requestInstance) // pass &requestInstance to modify original struct, not the copy (requestInstance).
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	type responseErr struct { // response body if there's an error
		Err string `json:"error"`
	}

	type responseValid struct { // response body for validity
		Valid bool `json:"valid"`
	}

	responseErrInstance := responseErr{}
	responseValidInstance := responseValid{}

	w.Header().Set("Content-Type", "application/json")

	// check for length of request's body
	if len(requestInstance.Body) > 140 {
		responseErrInstance.Err = "Chirp is too long"
		w.WriteHeader(400)
		dat, err := json.Marshal(responseErrInstance) // json.Marshal converts Go data structures (like structs) into compact JSON byte arrays
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat) // write out the error, in json format
	} else {
		w.WriteHeader(200)
		responseValidInstance.Valid = true
		dat, err := json.Marshal(responseValidInstance)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat)
	}

}

func main() {
	mux := http.NewServeMux()

	apiCfg := apiConfig{}

	// Create file server handler
	fileServer := http.FileServer(http.Dir("."))
	fileServerStripped := http.StripPrefix("/app", fileServer)

	// // Build middleware stack
	// middlewareStack := Middlewares{}
	// middlewareStack.add(apiCfg.middlewareMetricsInc())
	// middlewareStack.add(middlewareHeaderGetter())

	// // Apply middleware to file server
	// mux.Handle("/app/", middlewareStack.applyMiddlewares(fileServerStripped))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerStripped))

	// Register other endpoints
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("GET /api/healthz", myHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	// Start server
	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	fmt.Println("Server listening on http://localhost:8080")
	if err := s.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
}
