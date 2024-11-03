package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"sync"
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
	server, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal("Could not start server on port 8000")
		return
	}
	fmt.Println("Server: Establishing server on port 8000...")
	wg := sync.WaitGroup{}
	for i := 0; i < expected; i++ {
		conn, err := server.Accept()
		if err != nil {
			log.Println("Something happened with the client, could not handle it")
			continue
		}
		wg.Add(1)
		go handleClient(conn, &wg)
	}
	wg.Wait()
	log.Printf("Served #%v clients", numClients)
}

func handleClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	// need to read the number the client sent
	data := make([]byte, 1024)
	for true {
		num, err := conn.Read(data)
		if err != nil {
			log.Println("Could not read data from client", conn.RemoteAddr().String())
			return
		}
		// decode data into string
		guess := string(data[:num])
		// fmt.Printf("Server: Received %v bytes from client: %s\n", num, guess)
		numGuess, err := strconv.Atoi(guess)
		if err != nil {
			log.Println("Server: Could not parse int from guess")
			return
		}
		guessedCorrectly := false
		rwm.RLock()
		var response string
		if numGuess < correctNum {
			response = "Too low!"
		} else if numGuess > correctNum {
			response = "Too high!"
		} else {
			response = "Correct!"
			mut.Lock()
			numClients++
			mut.Unlock()
			guessedCorrectly = true
		}
		rwm.RUnlock()
		conn.Write([]byte(response))
		if guessedCorrectly {
			log.Println("Server: A Client guessed correctly; ending connection there")
			go reshuffle()
			return
		}
	}
}
