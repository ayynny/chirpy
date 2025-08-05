package main

// import (
// 	"fmt"
// 	"net/http"
// )

// // Middleware signature: takes and returns an http.Handler
// type Middlewares struct {
// 	handlers []func(http.Handler) http.Handler
// }

// // Middleware: prints headers
// func middlewareHeaderGetter() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			header := r.Header
// 			fmt.Println("Headers:")
// 			for key, values := range header {
// 				for _, val := range values {
// 					line := key + ": " + val + "\n"
// 					w.Write([]byte(line))
// 					fmt.Print(line)
// 				}
// 			}
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// // Middleware: increments hit counter
// func (cfg *apiConfig) middlewareMetricsInc() func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			fmt.Println("Visitor!")
// 			cfg.fileserverHits.Add(1)
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// // Add middleware to stack
// func (m *Middlewares) add(mw func(http.Handler) http.Handler) {
// 	m.handlers = append(m.handlers, mw)
// }

// // Apply middlewares in reverse order
// func (m *Middlewares) applyMiddlewares(final http.Handler) http.Handler {
// 	for i := len(m.handlers) - 1; i >= 0; i-- {
// 		final = m.handlers[i](final)
// 	}
// 	return final
// }

// func main(){

// // Build middleware stack
// middlewareStack := Middlewares{}
// middlewareStack.add(apiCfg.middlewareMetricsInc())
// middlewareStack.add(middlewareHeaderGetter())

// // Apply middleware to file server
// mux.Handle("/app/", middlewareStack.applyMiddlewares(fileServerStripped))
// }
