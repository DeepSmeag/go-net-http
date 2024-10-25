package main

import (
	"log"
	"net/http"
)

func main() {
	var router = http.NewServeMux()
	router.HandleFunc("/item/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.Write([]byte("Item ID: " + id))
	})

	server := http.Server{Addr: ":8080", Handler: router}
	log.Println("Server started at :8080")
	server.ListenAndServe()
}
