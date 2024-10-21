package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	cmd := exec.Command("/users/haonan/Project-Go/src/github.com/scalog/scalog/scalog", "client", "--config", "/users/haonan/Project-Go/src/github.com/scalog/scalog/.scalog.yaml")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error getting stdin:", err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error getting stderr", err)
		return
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting command:", err)
		return
	}

	writer := bufio.NewWriter(stdin)
	reader := bufio.NewReader(stderr)

	duration := 10 * time.Second
	startTime := time.Now()
	logCount := 0
	totalExectime := time.Duration(0)

	for time.Since(startTime) < duration {

		logMessage := fmt.Sprintf("append log_entry_number_%d\n", logCount+1)
		localStartTime := time.Now()
		_, err := writer.WriteString(logMessage)
		if err != nil {
			fmt.Println("Error writing to stdin:", err)
			break
		}

		writer.Flush()

		for {
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading from stderr:", err)
				break
			}
			if strings.Contains(response, "Append result") {
				localEndTime := time.Now()
				latency := localEndTime.Sub(localStartTime)
				fmt.Printf("Log entry %d appended successfully. Latency: %v. Response: %s\n", logCount+1, latency, response)
				totalExectime += latency
				logCount++
				break
			}
		}
	}

	_, err = writer.WriteString("exit\n")
	if err != nil {
		fmt.Println("Error writing exit to stdin:", err)
	}
	writer.Flush()

	stdin.Close()

	err = cmd.Wait()
	if err != nil {
		fmt.Println("Command finished with error:", err)
	}

	if logCount > 0 {
		fmt.Printf("Total execution time: %d ms\n", totalExectime.Milliseconds())
		fmt.Printf("Total logs written: %d\n", logCount)
	} else {
		fmt.Println("No logs were successfully written.")
	}
}
