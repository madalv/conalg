package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gookit/slog"
)

var keyPool = []string{
	"key1",
	"key2",
	"key3",
	"key4",
}

func sendRequest(url, payload string, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(time.Duration(rand.Intn(15)) * time.Second)

	slog.Infof("Sending request to %s with payload %s", url, payload)

	data := map[string]interface{}{
		"command": payload,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	_, err = http.Post("http://"+url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request to %s: %v\n", url, err)
		return
	}
}

func main() {
	slog.Print("Starting load generator")
	urls := []string{
		"localhost:8081/propose",
		"localhost:8082/propose",
		"localhost:8083/propose",
		// Add more URLs if needed
	}


	startTime := time.Now()
	timeout := 30 * time.Second
	commandNr := 50
	wg := sync.WaitGroup{}

	for i := 1; i <= commandNr; i++ {
		if time.Since(startTime) >= timeout {
			break
		}
		wg.Add(1)
		randomIndex := rand.Intn(len(urls))
		go sendRequest(urls[randomIndex], chooseName(70), &wg)
	}

	wg.Wait()
	slog.Printf("Sent %d requests in %v", commandNr, time.Since(startTime))
}

func chooseName(probOfConflict int) string {
	if rand.Intn(100) < probOfConflict {
		randomIndex := rand.Intn(len(keyPool))
		return keyPool[randomIndex]
	}
	return fmt.Sprintf("command%d", rand.Intn(1000))
}
