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

		// Парсим значения с проверкой ошибок
		la, err1 := strconv.ParseFloat(parts[0], 64)
		memTotal, err2 := strconv.ParseFloat(parts[1], 64)
		memUsed, err3 := strconv.ParseFloat(parts[2], 64)
		diskTotal, err4 := strconv.ParseFloat(parts[3], 64)
		diskUsed, err5 := strconv.ParseFloat(parts[4], 64)
		netTotal, err6 := strconv.ParseFloat(parts[5], 64)
		netUsed, err7 := strconv.ParseFloat(parts[6], 64)

		// Проверяем ошибки парсинга
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil ||
			err5 != nil || err6 != nil || err7 != nil {
			time.Sleep(60 * time.Second)
			continue
		}

		// 1. Load Average > 30
		if la > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", la)
		}

		// 2. Memory > 80% (округляем вниз)
		if memTotal > 0 {
			memUsage := memUsed / memTotal
			if memUsage > 0.8 {
				percent := int(memUsage * 100)
				fmt.Printf("Memory usage too high: %d%%\n", percent)
			}
		}

		// 3. Disk > 90% (округляем вниз)
		if diskTotal > 0 {
			diskUsage := diskUsed / diskTotal
			if diskUsage > 0.9 {
				freeMB := int((diskTotal - diskUsed) / (1024 * 1024))
				fmt.Printf("Free disk space is too low: %d Mb left\n", freeMB)
			}
		}

		// 4. Network > 90% (в Mbit/s)
		if netTotal > 0 {
			netUsage := netUsed / netTotal
			if netUsage > 0.9 {
				freeMbits := netTotal - netUsed
				fmt.Printf("Network bandwidth usage high: %.1f Mbit/s available\n", freeMbits)
			}
		}

		time.Sleep(60 * time.Second)
	}
}
