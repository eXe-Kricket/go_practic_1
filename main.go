package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	errorCount := 0

	for {
		// Отправляем запрос
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				println("Unable to fetch server statistic")
			}
			time.Sleep(60 * time.Second)
			continue
		}

		// Проверяем статус
		if resp.StatusCode != 200 {
			resp.Body.Close()
			errorCount++
			if errorCount >= 3 {
				println("Unable to fetch server statistic")
			}
			time.Sleep(60 * time.Second)
			continue
		}

		// Читаем данные
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				println("Unable to fetch server statistic")
			}
			time.Sleep(60 * time.Second)
			continue
		}

		// Сбрасываем счетчик ошибок
		errorCount = 0

		// Парсим данные
		stats := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(stats) != 7 {
			time.Sleep(60 * time.Second)
			continue
		}

		// Парсим значения
		la, _ := strconv.ParseFloat(stats[0], 64)
		memTotal, _ := strconv.ParseFloat(stats[1], 64)
		memUsed, _ := strconv.ParseFloat(stats[2], 64)
		diskTotal, _ := strconv.ParseFloat(stats[3], 64)
		diskUsed, _ := strconv.ParseFloat(stats[4], 64)
		netTotal, _ := strconv.ParseFloat(stats[5], 64)
		netUsed, _ := strconv.ParseFloat(stats[6], 64)

		// Проверяем пороги

		// Load Average
		if la > 30 {
			printf("Load Average is too high: %.0f\n", la)
		}

		// Memory
		if memTotal > 0 {
			usage := memUsed / memTotal
			if usage > 0.8 {
				printf("Memory usage too high: %.0f%%\n", usage*100)
			}
		}

		// Disk
		if diskTotal > 0 {
			usage := diskUsed / diskTotal
			if usage > 0.9 {
				freeMB := (diskTotal - diskUsed) / (1024 * 1024)
				printf("Free disk space is too low: %.0f Mb left\n", freeMB)
			}
		}

		// Network
		if netTotal > 0 {
			usage := netUsed / netTotal
			if usage > 0.9 {
				freeMbit := (netTotal - netUsed) / 125000
				printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
			}
		}

		time.Sleep(60 * time.Second)
	}
}

// println и printf для вывода без буферизации
func println(s string) {
	print(s + "\n")
}

func printf(format string, args ...interface{}) {
	print(format, args...)
	print("\n")
}

func print(args ...interface{}) {
	// Простой вывод
	s := fmt.Sprint(args...)
	os.Stdout.Write([]byte(s))
}
