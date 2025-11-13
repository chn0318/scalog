package client

import (
	"fmt"
	"sync"
	"time"

	log "github.com/chn0318/scalog/logger"
	"github.com/chn0318/scalog/util"
	"github.com/spf13/viper"
)

func Start() {
	it, err := NewIt()
	if err != nil {
		log.Fatalf("%v", err)
		return
	}
	it.Start()
}

func singleClientPerf(it It, index int, stopChan chan struct{}, totalRecordChan chan int, totalLatencyChan chan float64) {
	messageSize := viper.GetInt("size")
	totalRecord := 0 // Tracks the total number of records appended
	totalLatency := 0.0
	for {
		select {
		case <-stopChan:
			// Handle stop signal by sending results and exiting the function
			avgLatency := totalLatency / float64(totalRecord)
			log.Infof("Client-%d: total Record: %v, avg Latency: %v ms\n", index, totalRecord, avgLatency)
			sendResults(totalRecord, totalLatency, totalRecordChan, totalLatencyChan)

			return

		default:
			// Generate a log record
			record := util.GenerateRandomString(messageSize)

			// Append the record using the client's AppendOne method
			start := time.Now()
			_, _, err := it.client.AppendOne(record)
			elapse := time.Since(start)
			totalLatency += float64(elapse.Milliseconds())
			if err != nil {
				// Log error to standard error and send results before exiting
				log.Errorf("Client-%d encountered an error: %v\n", index, err)
				sendResults(totalRecord, totalLatency, totalRecordChan, totalLatencyChan)
				return
			}

			totalRecord++ // Increment the total record count
		}
	}
}

func sendResults(totalRecord int, totalLatency float64, totalRecordChan chan int, totalLatencyChan chan float64) {
	totalRecordChan <- totalRecord
	totalLatencyChan <- totalLatency
}

func Perf() {
	threads := viper.GetInt("threads")
	var clientArray []It
	for i := 0; i < threads; i++ {
		it, err := NewIt()
		if err != nil {
			log.Fatalf("%v", err)
			return
		}
		clientArray = append(clientArray, *it)
	}

	stopChan := make(chan struct{}) // Channel to signal goroutines to stop
	totalRecordChan := make(chan int, threads)
	totalLatencyChan := make(chan float64, threads)
	var wg sync.WaitGroup // WaitGroup to synchronize goroutines

	start := time.Now()
	// Launch goroutines
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			singleClientPerf(clientArray[index], index, stopChan, totalRecordChan, totalLatencyChan)
		}(i)
	}

	duration := viper.GetDuration("duration")
	// Let the test run for the specified duration
	time.Sleep(duration)

	// Signal all goroutines to stop
	close(stopChan)

	// Wait for all goroutines to complete
	wg.Wait()
	elapse := time.Since(start)
	totalRecords := 0
	totalLatency := 0.0
	for i := 0; i < threads; i++ {
		totalRecords += <-totalRecordChan
		totalLatency += <-totalLatencyChan
	}

	// Print final results
	fmt.Printf("Total Records Processed: %d\n", totalRecords)
	fmt.Printf("Total BandWidth: %.2f writes/s\n", float64(totalRecords)/elapse.Seconds())
	fmt.Printf("Average Latency (across all threads): %.2f ms\n", totalLatency/float64(totalRecords))
}
