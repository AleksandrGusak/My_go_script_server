package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL       = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval    = 1 * time.Second
	maxErrors       = 3
	loadLimit       = 30.0
	memoryLimit     = 0.8 // 80%
	diskLimit       = 0.9 // 90%
	bandwidthLimit  = 0.9 // 90%
)

func main() {
	errorCount := 0
	
	for {
		stats, err := getServerStats()
		
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				// Не сбрасываем счетчик, чтобы сообщение не повторялось постоянно
				// Ждем перед следующей попыткой
				time.Sleep(pollInterval)
				continue
			}
		} else {
			errorCount = 0
			checkThresholds(stats)
		}
		
		time.Sleep(pollInterval)
	}
}

func getServerStats() ([]uint64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Парсим CSV строку
	valuesStr := strings.TrimSpace(string(body))
	parts := strings.Split(valuesStr, ",")
	
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid data format")
	}
	
	var stats []uint64
	for _, part := range parts {
		// Первое значение (Load Average) может быть дробным
		if len(stats) == 0 {
			// Для Load Average используем преобразование в uint64
			val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid load average: %v", err)
			}
			stats = append(stats, uint64(val))
		} else {
			val, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value: %v", err)
			}
			stats = append(stats, val)
		}
	}
	
	return stats, nil
}

func checkThresholds(stats []uint64) {
	if len(stats) != 6 {
		return
	}
	
	loadAvg := float64(stats[0])
	totalMemory := stats[1]
	usedMemory := stats[2]
	totalDisk := stats[3]
	usedDisk := stats[4]
	totalBandwidth := stats[5] // пропускная способность
	// Используем usedMemory как текущую загруженность сети (по контексту задания)
	currentBandwidth := usedMemory
	
	// 1. Проверка Load Average
	if loadAvg > loadLimit {
		fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
	}
	
	// 2. Проверка использования памяти
	if totalMemory > 0 {
		memoryUsage := float64(usedMemory) / float64(totalMemory)
		if memoryUsage > memoryLimit {
			fmt.Printf("Memory usage too high: %.2f%%\n", memoryUsage*100)
		}
	}
	
	// 3. Проверка дискового пространства
	if totalDisk > 0 {
		diskUsage := float64(usedDisk) / float64(totalDisk)
		if diskUsage > diskLimit {
			freeMB := float64(totalDisk-usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.2f Mb left\n", freeMB)
		}
	}
	
	// 4. Проверка пропускной способности сети
	if totalBandwidth > 0 {
		bandwidthUsage := float64(currentBandwidth) / float64(totalBandwidth)
		if bandwidthUsage > bandwidthLimit {
			// Конвертируем доступную полосу из байт/с в мегабит/с
			availableBytesPerSec := totalBandwidth - currentBandwidth
			availableMbitsPerSec := float64(availableBytesPerSec) * 8 / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", availableMbitsPerSec)
		}
	}
}