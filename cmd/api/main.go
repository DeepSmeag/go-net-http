package main

import (
	"log"
	"net/http"
	"time"

	"github.com/DeepSmeag/go-net-http/internal/handlers"
	"github.com/DeepSmeag/go-net-http/internal/middleware"
)

type wrapperWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrapperWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrapperWriter{ResponseWriter: w, statusCode: http.StatusOK} // local variable, we take its address and pass it forward as a pointer to this local variable; the variable no longer exists once we go out of this middleware
		log.Printf("Wrapped %p", wrapped)
		next.ServeHTTP(wrapped, r)

		log.Println(wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))
	})
}
func EnsureLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("EnsureLoggedIn")
		next.ServeHTTP(w, r)
	})
}

func main() {
	var router = http.NewServeMux()
	router.HandleFunc("/item/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.Write([]byte("Item ID: " + id))
	})

	// without middleware chain
	// server := http.Server{Addr: ":8080", Handler: LoggingMiddleware(router)}

	//with middleware chain
	chain := middleware.CreateStack(LoggingMiddleware, EnsureLoggedIn)

	usersRouter := handlers.UserHandler()
	router.Handle("/users/", http.StripPrefix("/users", usersRouter))

	server := http.Server{Addr: ":8080", Handler: chain(router)}
	log.Println("Server started at :8080")
	server.ListenAndServe()
}
