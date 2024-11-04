package main

import (
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./cmd/http3/index.html")
	})
	router.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Yes it works"))
	})
	log.Println("Starting server...")
	// if I don't do TLS to listen on TLS/TCP at the same time it doesn't work; there's something about QUIC connections I'm not getting
	err := http3.ListenAndServeTLS("localhost:8080", "./cmd/http3/server.crt", "./cmd/http3/server.key", router)
	if err != nil {
		log.Println(err)
	}
}
