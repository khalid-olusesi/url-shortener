package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	var wg sync.WaitGroup

	start := time.Now()

	var durations []time.Duration

	var mu sync.Mutex

	for i := 0; i < 100; i++ {

		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			fmt.Println("Request number:", i+1)

			requestStart := time.Now()

			resp, err := client.Get("http://localhost:8080/0KSEH5")

			if err != nil {
				fmt.Println("Request failed:", err)
				return
			}

			requestDuration := time.Since(requestStart)
			mu.Lock()
			durations = append(durations, requestDuration)
			mu.Unlock()
			fmt.Println("Status:", resp.StatusCode)

			if resp.StatusCode == http.StatusFound {
				fmt.Println("Redirect successful")
			} else {
				fmt.Println("Unexpected status code")
			}

			resp.Body.Close()
		}(i)
	}

	wg.Wait()
	fmt.Println("Measurements recorded:", len(durations))

	elapsed := time.Since(start)
	fmt.Println("Total request duration:", elapsed)

	var totalLatency time.Duration

	for _, duration := range durations {
		totalLatency += duration
	}

	if len(durations) > 0 {
		averageLatency := totalLatency / time.Duration(len(durations))
		fmt.Println("Average latency:", averageLatency)
	}

	requestsPerSecond := float64(len(durations)) / elapsed.Seconds()
	fmt.Println("Requests per second:", requestsPerSecond)
}
