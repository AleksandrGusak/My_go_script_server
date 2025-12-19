package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	errorCount := 0
	
	for {
		time.Sleep(1 * time.Second)
		
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic.")
			}
			continue
		}
		
		if resp.StatusCode != 200 {
			resp.Body.Close()
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic.")
			}
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic.")
			}
			continue
		}
		
		errorCount = 0
		processStats(string(body))
	}
}

func processStats(data string) {
	parts := strings.Split(strings.TrimSpace(data), ",")
	if len(parts) != 6 {
		return
	}
	
	// Load Average
	loadAvg, err := strconv.ParseFloat(parts[0], 64)
	if err == nil && loadAvg > 30 {
		fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
	}
	
	// Memory
	totalMem, _ := strconv.ParseUint(parts[1], 10, 64)
	usedMem, _ := strconv.ParseUint(parts[2], 10, 64)
	if totalMem > 0 {
		memoryUsage := float64(usedMem) / float64(totalMem)
		if memoryUsage > 0.8 {
			fmt.Printf("Memory usage too high: %.2f%%\n", memoryUsage*100)
		}
	}
	
	// Disk
	totalDisk, _ := strconv.ParseUint(parts[3], 10, 64)
	usedDisk, _ := strconv.ParseUint(parts[4], 10, 64)
	if totalDisk > 0 {
		diskUsage := float64(usedDisk) / float64(totalDisk)
		if diskUsage > 0.9 {
			freeMB := float64(totalDisk-usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.2f Mb left\n", freeMB)
		}
	}
	
	// Network bandwidth (assuming parts[5] is total bandwidth)
	totalBW, _ := strconv.ParseUint(parts[5], 10, 64)
	// Using usedMem as current network usage based on context
	if totalBW > 0 {
		bwUsage := float64(usedMem) / float64(totalBW)
		if bwUsage > 0.9 {
			availableMbits := float64(totalBW-usedMem) * 8 / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", availableMbits)
		}
	}
}