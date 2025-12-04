package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			continue
		}

		errorCount = 0

		data := strings.TrimSpace(string(body))
		parts := strings.Split(data, ",")
		if len(parts) != 7 {
			continue
		}

		// Парсим значения
		var vals [7]float64
		for i, part := range parts {
			v, _ := strconv.ParseFloat(strings.TrimSpace(part), 64)
			vals[i] = v
		}

		la := vals[0]
		memT := vals[1]
		memU := vals[2]
		diskT := vals[3]
		diskU := vals[4]
		netT := vals[5]
		netU := vals[6]

		// Load Average
		if la > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", la)
		}

		// Memory
		if memT > 0 && memU/memT > 0.8 {
			fmt.Printf("Memory usage too high: %d%%\n", int(memU/memT*100))
		}

		// Disk
		if diskT > 0 && diskU/diskT >= 0.9 {
			free := (diskT - diskU) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", int(free))
		}

		// Network
		if netT > 0 && netU/netT > 0.9 {
			free := (netT - netU) / 1000000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", int(free))
		}
	}
}
