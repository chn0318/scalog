package client

import (
	"fmt"
	"sync"
	"time"
	"os"
	"gopkg.in/yaml.v3"

	log "github.com/chn0318/scalog/logger"
	"github.com/chn0318/scalog/util"
	"github.com/spf13/viper"
)

type Config struct {
    JsonName       string `yaml:"json_name"`
    ClientIP       string `yaml:"client_ip"`
}

func Start() {
	it, err := NewIt()
	if err != nil {
		log.Fatalf("%v", err)
		return

	}
	it.Start()
}

func singleClientPerf(it It, index int, stopChan chan struct{}, stats *Stats, duration int64) {
	messageSize := viper.GetInt("size")
	totalRecord := 0 // Tracks the total number of records appended
	totalLatency := 0.0
	log.Infof("At the beginning of singleClientPerf!\n")
	for {
		select {
		case <-stopChan:
			// Handle stop signal by sending results and exiting the function
			avgLatency := totalLatency / float64(totalRecord)
			log.Infof("Client-%d: total Record: %v, avg Latency: %v ms\n", index, totalRecord, avgLatency)
			stats.ExportResults(duration)
			// Print final results
			fmt.Printf("Total Records Processed: %d\n", totalRecord)
			fmt.Printf("Total BandWidth: %.2f writes/s\n", float64(totalRecord)/float64(duration))
			fmt.Printf("Average Latency (across all threads): %.2f ms\n", totalLatency/float64(totalRecord))
			return
		default:
			// Generate a log record
			record := util.GenerateRandomString(messageSize)

			// Append the record using the client's AppendOne method
			start := time.Now()
			_, _, err := it.client.AppendOne(record)
			elapse := time.Since(start)
			totalLatency += float64(elapse.Milliseconds())
			log.Infof("Avg Lat: %.2fms \n", totalLatency)
			if err != nil {
				// Log error to standard error and send results before exiting
				log.Errorf("Client-%d encountered an error: %v\n", index, err)
				stats.ExportResults(duration)
				return
			}
			totalRecord++ // Increment the total record count
			stats.AddOp()
			stats.AddDuration(float64(elapse.Milliseconds()))
		}
	}
}

func Perf() {
	log.Infof("At the beginning of the Perf function!\n")
    	data, err := os.ReadFile("data.yaml")
    	if err != nil {
    	    log.Fatalf("error: %v", err)
    	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
	    log.Fatalf("error: %v", err)
	}

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
	var wg sync.WaitGroup // WaitGroup to synchronize goroutines

	duration := viper.GetDuration("duration")
	log.Infof("Duration: %ds\n", duration)
	// Launch goroutines
	for i := 0; i < threads; i++ {
		wg.Add(1)
		stats := NewStats(config.JsonName, config.ClientIP)
		go func(index int) {
			defer wg.Done()
			singleClientPerf(clientArray[index], index, stopChan, stats, int64(duration))
		}(i)
		stats.Close()
	}

	// Let the test run for the specified duration
	time.Sleep(duration)

	// Signal all goroutines to stop
	close(stopChan)

	// Wait for all goroutines to complete
	wg.Wait()
}
