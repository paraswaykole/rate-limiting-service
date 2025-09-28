package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	url := "http://localhost:8080/api1/hello"
	totalRequests := 1000
	concurrency := 10
	duration := 60 * time.Second

	var wg sync.WaitGroup
	requestsCh := make(chan struct{}, concurrency)
	resultsCh := make(chan time.Duration, totalRequests)
	errorsCh := make(chan error, totalRequests)

	// Calculate interval between requests to fit in total duration
	interval := duration / time.Duration(totalRequests)
	fmt.Printf("Sending %d requests in %v (interval: %v) with concurrency %d\n", totalRequests, duration, interval, concurrency)

	start := time.Now()
	for range totalRequests {
		wg.Add(1)
		requestsCh <- struct{}{}
		go func() {
			defer wg.Done()
			reqStart := time.Now()
			resp, err := http.Get(url)
			elapsed := time.Since(reqStart)
			if err != nil {
				errorsCh <- err
			} else {
				resp.Body.Close()
				resultsCh <- elapsed
			}
			<-requestsCh
		}()

		time.Sleep(interval) // pace requests
	}

	wg.Wait()
	close(resultsCh)
	close(errorsCh)

	// Collect stats
	var totalTime time.Duration
	var count int
	var minTime, maxTime time.Duration
	for t := range resultsCh {
		totalTime += t
		count++
		if minTime == 0 || t < minTime {
			minTime = t
		}
		if t > maxTime {
			maxTime = t
		}
	}

	totalErrors := len(errorsCh)
	end := time.Since(start)

	fmt.Println("\nBenchmark Results:")
	fmt.Printf("Total Requests:\t\t%d\n", totalRequests)
	fmt.Printf("Successful Requests:\t%d\n", count)
	fmt.Printf("Failed Requests:\t%d\n", totalErrors)
	fmt.Printf("Total Time:\t\t%.2fs\n", end.Seconds())
	if count > 0 {
		fmt.Printf("Average Time/Request:\t%.2fms\n", totalTime.Seconds()*1000/float64(count))
		fmt.Printf("Min Time/Request:\t%.2fms\n", minTime.Seconds()*1000)
		fmt.Printf("Max Time/Request:\t%.2fms\n", maxTime.Seconds()*1000)
		fmt.Printf("Requests per second:\t%.2f\n", float64(count)/end.Seconds())
	}
}
