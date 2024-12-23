package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var correctNum = rand.IntN(11)
var rwm = sync.RWMutex{}
var mut = sync.Mutex{}
var numClients = 0

func reshuffle() {
	rwm.Lock()
	correctNum = rand.IntN(11)
	log.Println("Server: Changed number to:", correctNum)
	rwm.Unlock()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run server.go <expectedClients:int>")
	}
	expected, _ := strconv.Atoi(os.Args[1])
	listenAddr, err := net.ResolveUDPAddr("udp", "localhost:8000")
	server, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Fatal("Could not start server on port 8000")
		return
	}
	fmt.Println("Server: Establishing server...", server.LocalAddr())
	for numClients != expected {
		data := make([]byte, 1024)
		server.SetReadDeadline(time.Now().Add(time.Second))
		num, addr, err := server.ReadFromUDP(data)
		if err != nil {
			log.Println("Server: Read timeout")
			continue
		}
		go handlePacket(server, addr, data[:num])
	}
	log.Printf("Served #%v packets", numClients)
}

func handlePacket(
	server *net.UDPConn,
	addr *net.UDPAddr,
	data []byte,
) {
	guess := string(data)
	guessNum, err := strconv.Atoi(guess)
	if err != nil {
		log.Println("Server: error, received from client and can't convert-", guess)
	}
	// log.Println("Server: Received", message)
	var res string
	guessedCorrectly := false
	rwm.RLock()
	if guessNum < correctNum {
		res = "Too low!"
	} else if guessNum > correctNum {
		res = "Too high!"
	} else {
		res = "Correct!"
		mut.Lock()
		numClients++
		mut.Unlock()
		guessedCorrectly = true
	}
	rwm.RUnlock()
	server.WriteToUDP([]byte(res), addr)
	if guessedCorrectly {
		log.Printf("Client %s guessed correctly", addr.String())
		go reshuffle()
	}

}
