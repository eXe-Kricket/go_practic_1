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
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				os.Stdout.Sync()
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
				os.Stdout.Sync()
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
				os.Stdout.Sync()
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

		// Парсим значения
		la, _ := strconv.ParseFloat(parts[0], 64)
		memTotal, _ := strconv.ParseFloat(parts[1], 64)
		memUsed, _ := strconv.ParseFloat(parts[2], 64)
		diskTotal, _ := strconv.ParseFloat(parts[3], 64)
		diskUsed, _ := strconv.ParseFloat(parts[4], 64)
		netTotal, _ := strconv.ParseFloat(parts[5], 64)
		netUsed, _ := strconv.ParseFloat(parts[6], 64)

		hasOutput := false

		// 1. Load Average > 30
		if la > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", la)
			hasOutput = true
		}

		// 2. Memory > 80%
		if memTotal > 0 {
			usage := memUsed / memTotal
			if usage > 0.8 {
				percent := int(usage * 100)
				fmt.Printf("Memory usage too high: %d%%\n", percent)
				hasOutput = true
			}
		}

		// 3. Disk > 90%
		if diskTotal > 0 {
			usage := diskUsed / diskTotal
			if usage >= 0.9 {
				freeMB := int((diskTotal - diskUsed) / (1024 * 1024))
				fmt.Printf("Free disk space is too low: %d Mb left\n", freeMB)
				hasOutput = true
			}
		}

		// 4. Network > 90%
		if netTotal > 0 {
			usage := netUsed / netTotal
			if usage > 0.9 {
				freeMbits := int((netTotal - netUsed) / 1000000)
				fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbits)
				hasOutput = true
			}
		}

		// Синхронизируем вывод
		if hasOutput {
			os.Stdout.Sync()
		}

		time.Sleep(60 * time.Second)
	}
}
