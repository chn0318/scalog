package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// 启动 scalog client 进程
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

	// 启动进程
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting command:", err)
		return
	}

	writer := bufio.NewWriter(stdin)
	reader := bufio.NewReader(stderr)

	// 设置执行时间为10秒
	duration := 10 * time.Second
	startTime := time.Now()
	logCount := 0
	totallatency := time.Duration(0)

	// 不断发送 append 命令，直到达到10秒
	for time.Since(startTime) < duration {
		// 写入日志内容到 scalog client 进程
		logMessage := fmt.Sprintf("append log_entry_number_%d\n", logCount+1)
		fmt.Println("Send Request: Append %s", logMessage)
		localStartTime := time.Now()
		_, err := writer.WriteString(logMessage)
		if err != nil {
			fmt.Println("Error writing to stdin:", err)
			break
		}

		// 刷新缓冲区，确保命令发送出去
		writer.Flush()

		// 读取并等待 "Append result" 输出
		for {
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading from stderr:", err)
				break
			}
			fmt.Printf("Receive Response: %s", response)
			if strings.Contains(response, "Append result") {
				localEndTime := time.Now()
				latency := localEndTime.Sub(localStartTime)
				totallatency += latency

				fmt.Printf("Log entry %d appended successfully. Latency: %v. Response: %s\n", logCount+1, latency, response)

				logCount++
				break
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}

	// 发送退出命令
	_, err = writer.WriteString("exit\n")
	if err != nil {
		fmt.Println("Error writing exit to stdin:", err)
	}
	writer.Flush()

	// 关闭stdin，结束scalog客户端进程
	stdin.Close()

	// 等待命令执行完成
	err = cmd.Wait()
	if err != nil {
		fmt.Println("Command finished with error:", err)
	}

	// 输出总执行时间和平均时延
	totalTime := time.Since(startTime)
	fmt.Printf("Total execution time: %v\n", totalTime)
	if logCount > 0 {
		averageLatency := totallatency / time.Duration(logCount)
		fmt.Printf("Average latency: %v ms\n", averageLatency.Milliseconds())
	} else {
		fmt.Println("No logs were successfully written.")
	}
	fmt.Printf("Total logs written: %d\n", logCount)
}
