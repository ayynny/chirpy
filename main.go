package main

import "net/http"

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux() // 1. create a new "router" (ServeMux) which decides what code (handler) to run for various paths

	s := http.Server{ // 2. set up http.Server struct
		Handler: mux,
		Addr:    ":8080",
	}

	fileServer := http.FileServer(http.Dir(".")) // returns a http.Handler

	mux.Handle("/app/", http.StripPrefix("/app", fileServer)) //register handler before server starts serving requests

	mux.HandleFunc("/healthz", myHandler)

	err := s.ListenAndServe()
	if err != nil {
		return
	}

}
