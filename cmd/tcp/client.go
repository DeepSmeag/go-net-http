package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var mut = sync.Mutex{}
var numClients = 0

func startClient(wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second)
	if err != nil {
		log.Println("Client: could not connect to the server")
		return
	}
	defer conn.Close()
	data := make([]byte, 1024)
	fmt.Println("Client: connected to server")
	// reader := bufio.NewReader(os.Stdin)
	for true {
		//loop until we guess correctly
		// guess, err := reader.ReadString('\n')
		// guess = strings.TrimSpace(guess)
		guess := strconv.Itoa(rand.Intn(11)) // if using stdin to input manually uncomment the other lines
		// if err != nil {
		// 	log.Println("Could not read from stdin")
		// 	return
		// }
		// not validating, assuming number is good
		conn.Write([]byte(guess))
		num, err := conn.Read(data)
		if err != nil {
			log.Println("Client: error on reading data from server", err)
			return
		}
		rec := string(data[:num])
		// fmt.Printf("Client: Received %v bytes: %s\n", num, rec)
		if rec == "Correct!" {
			fmt.Println("Client: Guessed correctly, ending execution.")
			mut.Lock()
			numClients++
			mut.Unlock()
			return
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <numClients>")
	}
	numClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid <numClients>. Need number")
	}
	wg := sync.WaitGroup{}
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go startClient(&wg)
		// time.Sleep(time.Millisecond * 1)
	}
	wg.Wait()

}
