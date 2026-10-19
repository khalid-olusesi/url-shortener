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

	const totalRequests = 1000
	const concurrency = 100

	var wg sync.WaitGroup
	var mu sync.Mutex

	var durations []time.Duration
	var successful int
	var failed int

	start := time.Now()

	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)

		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			requestStart := time.Now()

			resp, err := client.Get("http://localhost:8080/M829KTl2")

			requestDuration := time.Since(requestStart)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				failed++
				return
			}
			defer resp.Body.Close()

			durations = append(durations, requestDuration)

			if resp.StatusCode == http.StatusFound {
				successful++
			} else {
				failed++
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)

	var totalLatency time.Duration
	for _, duration := range durations {
		totalLatency += duration
	}

	fmt.Println("Total requests:", totalRequests)
	fmt.Println("Successful redirects:", successful)
	fmt.Println("Failed requests:", failed)
	fmt.Println("Completed requests:", len(durations))
	fmt.Println("Total duration:", elapsed)

	if len(durations) > 0 {
		averageLatency := totalLatency / time.Duration(len(durations))
		fmt.Println("Average latency:", averageLatency)
	}

	fmt.Printf("Requests per second: %.2f\n", float64(len(durations))/elapsed.Seconds())
}
