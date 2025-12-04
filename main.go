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
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(60 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(60 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(60 * time.Second)
			continue
		}

		errorCount = 0

		data := strings.TrimSpace(string(body))
		parts := strings.Split(data, ",")
		if len(parts) != 7 {
			time.Sleep(60 * time.Second)
			continue
		}

		// Парсим все значения
		var values [7]float64
		for i, part := range parts {
			val, _ := strconv.ParseFloat(strings.TrimSpace(part), 64)
			values[i] = val
		}

		la := values[0]
		memTotal := values[1]
		memUsed := values[2]
		diskTotal := values[3]
		diskUsed := values[4]
		netTotal := values[5]
		netUsed := values[6]

		// Load Average
		if la > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", la)
		}

		// Memory - вычисляем точно как в тесте
		if memTotal > 0 {
			memoryUsage := memUsed / memTotal
			if memoryUsage > 0.8 {
				// Важно: не округлять, а отбрасывать дробную часть
				percent := int(memoryUsage * 100)
				fmt.Printf("Memory usage too high: %d%%\n", percent)
			}
		}

		// Disk
		if diskTotal > 0 {
			diskUsage := diskUsed / diskTotal
			if diskUsage > 0.9 {
				freeMB := int((diskTotal - diskUsed) / (1024 * 1024))
				fmt.Printf("Free disk space is too low: %d Mb left\n", freeMB)
			}
		}

		// Network
		if netTotal > 0 {
			networkUsage := netUsed / netTotal
			if networkUsage > 0.9 {
				freeMbits := int((netTotal - netUsed) / 125000)
				fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbits)
			}
		}

		time.Sleep(60 * time.Second)
	}
}
