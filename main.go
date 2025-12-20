package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const url = "http://srv.msk01.gigacorp.local/_stats"

func main() {
	errorCount := 0

	for {
		ok := fetchAndProcess()

		if !ok {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
		} else {
			errorCount = 0
		}

		time.Sleep(1 * time.Second)
	}
}

func fetchAndProcess() bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 {
		return false
	}

	values := make([]float64, 7)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return false
		}
		values[i] = v
	}

	process(values)
	return true
}

func process(v []float64) {
	load := v[0]

	memTotal := v[1]
	memUsed := v[2]

	diskTotal := v[3]
	diskUsed := v[4]

	netTotal := v[5]
	netUsed := v[6]

	if load > 30 {
		fmt.Printf("Load Average is too high: %.0f\n", load)
	}

	memPercent := memUsed / memTotal * 100
	if memPercent > 80 {
		fmt.Printf("Memory usage too high: %.0f%%\n", memPercent)
	}

	diskFreeBytes := diskTotal - diskUsed
	diskFreeMB := diskFreeBytes / (1024 * 1024)
	if diskUsed/diskTotal*100 > 90 {
		fmt.Printf("Free disk space is too low: %.0f Mb left\n", diskFreeMB)
	}

	netFree := netTotal - netUsed
	netFreeMbit := netFree * 8 / 1_000_000
	if netUsed/netTotal*100 > 90 {
		fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", netFreeMbit)
	}
}