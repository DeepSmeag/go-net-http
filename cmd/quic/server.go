package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	quic "github.com/quic-go/quic-go"
)

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		return
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		fmt.Println("Failed to create certificate:", err)
		return
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)

	if err != nil {
		log.Fatal("Error loading certificate:", err)
	}
	tlsConfig := &tls.Config{

		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,                                   // For testing purposes only
		NextProtos:         []string{"h3", "http/1.1", "ping/1.1"}, // Enable QUIC and HTTP/3
	}

	quicConfig := &quic.Config{
		Allow0RTT:       true,
		KeepAlivePeriod: time.Minute,
	}
	listener, err := quic.ListenAddr("localhost:8080", tlsConfig, quicConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	wg := sync.WaitGroup{}
	log.Println("QUIC server started on localhost:8080")
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Println("Server: Could not accept connection")
		}
		wg.Add(1)
		go handleConnection(conn, &wg)
	}
	wg.Wait()
}
func handleConnection(conn quic.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Println("Server: Could not accept a stream")
	}
	log.Println("Server: opened stream")
	defer stream.Close()
	stream.Write([]byte("Hey there sir"))
	// data := make([]byte, 1024)
	// num, err := stream.Read(data)
	// if err != nil {
	// 	log.Println("Server: Could not read from stream of client ", conn.LocalAddr().String())
	// }
	// guess := string(data[:num])
	// log.Println("Server: received", guess)
}
