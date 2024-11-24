package client

import (
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/scalog/scalog/logger"
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
	totalRecord := 0 // Tracks the total number of records appended
	totalLatency := 0.0
	for {
		select {
		case <-stopChan:
			// Handle stop signal by sending results and exiting the function
			avgLatency := totalLatency / float64(totalRecord)
			fmt.Fprintf(os.Stderr, "Client-%d: total Record: %v, avg Latency: %v s\n", index, totalRecord, avgLatency)
			sendResults(totalRecord, totalLatency, totalRecordChan, totalLatencyChan)

			return

		default:
			// Generate a log record specific to the client and record number
			record := fmt.Sprintf("client-%d:log_entry_number_%v", index, totalRecord)

			// Append the record using the client's AppendOne method
			start := time.Now()
			gsn, shard, err := it.client.AppendOne(record)
			elapse := time.Since(start)
			totalLatency += elapse.Seconds()
			if err != nil {
				// Log error to standard error and send results before exiting
				fmt.Fprintf(os.Stderr, "Client-%d encountered an error: %v\n", index, err)
				sendResults(totalRecord, totalLatency, totalRecordChan, totalLatencyChan)
				return
			}

			// Log the successful append operation to standard error
			fmt.Fprintf(os.Stdout, "Client-%d Append result: { GSN: %d, Shard: %d, Record: %s }\n", index, gsn, shard, record)
			totalRecord++ // Increment the total record count
		}
	}
}

func sendResults(totalRecord int, totalLatency float64, totalRecordChan chan int, totalLatencyChan chan float64) {
	totalRecordChan <- totalRecord
	totalLatencyChan <- totalLatency
}

func Perf() {
	threads := int(viper.GetInt("threads"))
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

	// Launch goroutines
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			singleClientPerf(clientArray[index], index, stopChan, totalRecordChan, totalLatencyChan)
		}(i)
	}
	duration := 10 * time.Second
	// Let the test run for the specified duration
	time.Sleep(duration)

	// Signal all goroutines to stop
	close(stopChan)

	// Wait for all goroutines to complete
	wg.Wait()
	totalRecords := 0
	totalLatency := 0.0
	for i := 0; i < threads; i++ {
		totalRecords += <-totalRecordChan
		totalLatency += <-totalLatencyChan
	}

	// Print final results
	fmt.Fprintf(os.Stderr, "Total Records Processed: %d\n", totalRecords)
	fmt.Fprintf(os.Stderr, "Average Latency (across all threads): %.2f records/second\n", totalLatency/float64(totalRecords))
}
