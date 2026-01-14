package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// Configuration
	// Ensure this URL matches the one defined in your LOCALTESTING.md
	url := "http://localhost:8080/api/v1/checkin-checkout"
	contentType := "application/json"

	numEmployees := 5000
	requestsPerEmployee := 2
	totalRequests := numEmployees * requestsPerEmployee
	concurrency := 50 // Number of concurrent requests to avoid local port exhaustion

	fmt.Printf("Starting load test: %d employees (%d requests each) to %s with concurrency %d\n", numEmployees, requestsPerEmployee, url, concurrency)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency) // Semaphore to limit concurrency

	var successCount int64
	var failCount int64

	startTime := time.Now()

	for i := 0; i < numEmployees; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire token

		employeeID := fmt.Sprintf("load-test-emp-%d", i)

		go func(empID string) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

			payload := []byte(fmt.Sprintf(`{"employeeId": "%s"}`, empID))

			for j := 0; j < requestsPerEmployee; j++ {
				// Create a new request for each iteration
				resp, err := http.Post(url, contentType, bytes.NewBuffer(payload))
				if err != nil {
					atomic.AddInt64(&failCount, 1)
					// fmt.Printf("Connection error: %v\n", err)
					continue
				}

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failCount, 1)
				}
				resp.Body.Close()
			}
		}(employeeID)
	}

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Println("\n--- Load Test Results ---")
	fmt.Printf("Total Duration: %v\n", duration)
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful:     %d\n", successCount)
	fmt.Printf("Failed:         %d\n", failCount)
	fmt.Printf("Requests/Sec:   %.2f\n", float64(totalRequests)/duration.Seconds())
}
