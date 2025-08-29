package benchmark

import (
	"fmt"
	"os"
	"time"
)

const logFile = "benchmarks.txt"

func LogResult(testName string, duration time.Duration, notes string) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening Benchmark log file: %v\n", err)
		return
	}
	defer f.Close()

	timeStamp := time.Now().Format("2006-01-02 15:04:05")

	logEntry := fmt.Sprintf("[%s] %-25s | Duration: %-15s | Notes: %s\n",
		timeStamp,
		testName,
		duration.String(),
		notes,
	)
	if _, err := f.WriteString(logEntry); err != nil {
		fmt.Printf("Error writing to benchmark log file: %v\n", err)
	}

}
