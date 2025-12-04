package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Константы
const (
	serverURL             = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval          = 60 * time.Second
	loadAverageThreshold  = 30.0
	memoryUsageThreshold  = 0.8 // 80%
	diskUsageThreshold    = 0.9 // 90%
	networkUsageThreshold = 0.9 // 90%
	maxErrors             = 3
	requestTimeout        = 10 * time.Second
)

// Структура для хранения метрик сервера
type ServerStats struct {
	LoadAverage  float64
	TotalMemory  float64
	UsedMemory   float64
	TotalDisk    float64
	UsedDisk     float64
	TotalNetwork float64
	UsedNetwork  float64
}

func main() {
	errorCount := 0
	httpClient := &http.Client{
		Timeout: requestTimeout,
	}

	fmt.Printf("Starting server monitor for %s\n", serverURL)
	fmt.Printf("Polling interval: %v\n\n", pollInterval)

	for {
		// Получение статистики
		stats, err := fetchStats(httpClient)
		if err != nil {
			errorCount++
			fmt.Printf("Error fetching stats: %v\n", err)

			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				// Не выходим, продолжаем попытки
			}

			time.Sleep(pollInterval)
			continue
		}

		// Сброс счетчика ошибок при успешном запросе
		errorCount = 0

		// Проверка метрик
		checkMetrics(stats)

		// Ожидание перед следующим опросом
		time.Sleep(pollInterval)
	}
}

// fetchStats получает статистику с сервера
func fetchStats(client *http.Client) (*ServerStats, error) {
	resp, err := client.Get(serverURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Парсинг данных
	data := strings.TrimSpace(string(body))
	parts := strings.Split(data, ",")

	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(parts))
	}

	// Преобразование строк в числа
	var stats ServerStats
	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value %d '%s': %w", i, part, err)
		}

		switch i {
		case 0:
			stats.LoadAverage = val
		case 1:
			stats.TotalMemory = val
		case 2:
			stats.UsedMemory = val
		case 3:
			stats.TotalDisk = val
		case 4:
			stats.UsedDisk = val
		case 5:
			stats.TotalNetwork = val
		case 6:
			stats.UsedNetwork = val
		}
	}

	return &stats, nil
}

// checkMetrics проверяет метрики на превышение порогов
func checkMetrics(stats *ServerStats) {
	// 1. Проверка Load Average
	if stats.LoadAverage > loadAverageThreshold {
		fmt.Printf("Load Average is too high: %.1f\n", stats.LoadAverage)
	}

	// 2. Проверка использования памяти
	if stats.TotalMemory > 0 {
		memoryUsage := stats.UsedMemory / stats.TotalMemory
		if memoryUsage > memoryUsageThreshold {
			fmt.Printf("Memory usage too high: %.1f%%\n", memoryUsage*100)
		}
	}

	// 3. Проверка использования диска
	if stats.TotalDisk > 0 {
		diskUsage := stats.UsedDisk / stats.TotalDisk
		if diskUsage > diskUsageThreshold {
			freeMB := (stats.TotalDisk - stats.UsedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.1f Mb left\n", freeMB)
		}
	}

	// 4. Проверка использования сети
	if stats.TotalNetwork > 0 {
		networkUsage := stats.UsedNetwork / stats.TotalNetwork
		if networkUsage > networkUsageThreshold {
			freeMbits := (stats.TotalNetwork - stats.UsedNetwork) / 125000 // байты/с -> мегабиты/с
			fmt.Printf("Network bandwidth usage high: %.1f Mbit/s available\n", freeMbits)
		}
	}
}
