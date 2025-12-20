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
	load := int64(v[0])

	memTotal := int64(v[1])
	memUsed := int64(v[2])

	diskTotal := int64(v[3])
	diskUsed := int64(v[4])

	netTotal := int64(v[5])
	netUsed := int64(v[6])

	if load > 30 {
		fmt.Printf("Load Average is too high: %d\n", load)
	}

	memPercent := memUsed * 100 / memTotal
	if memPercent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", memPercent)
	}

	if diskUsed*100/diskTotal > 90 {
		diskFreeMB := (diskTotal - diskUsed) / (1024 * 1024)
		fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMB)
	}

	if netUsed*100/netTotal > 90 {
		netFreeMbit := (netTotal - netUsed) / 1_000_000
		fmt.Printf(
			"Network bandwidth usage high: %d Mbit/s available\n",
			netFreeMbit,
		)
	}
}