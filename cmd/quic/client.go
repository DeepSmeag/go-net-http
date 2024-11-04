package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	quic "github.com/quic-go/quic-go"
)

func main() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // testing only
		NextProtos:         []string{"h3", "http/1.1"},
	}
	url := "localhost:8080"

	ctx := context.Background()
	conn, err := quic.DialAddr(ctx, url, tlsConfig, nil)
	if err != nil {
		println(err.Error())
		return
	}
	log.Println("Client: Connected to server")
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		log.Fatal("Client: Could not open stream sync")
	}
	defer stream.Close()
	data := make([]byte, 1024)
	num, err := stream.Read(data)
	if err != nil {
		log.Println("Could not read from stream")
	}
	greeting := string(data[:num])
	log.Println("Client: received", greeting)
	stream.Write([]byte("Hello to you as well"))
	time.Sleep(time.Millisecond) // if we don't have this here, the stream gets closed before the write gets a chance to flush the buffer to network so the server receives something
}
