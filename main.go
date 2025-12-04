package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	errors := 0

	for {
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil || resp.StatusCode != 200 {
			errors++
			if errors >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(time.Minute)
			continue
		}

		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		parts := strings.Split(strings.TrimSpace(string(data)), ",")
		if len(parts) == 7 {
			errors = 0

			la, _ := strconv.ParseFloat(parts[0], 64)
			memT, _ := strconv.ParseFloat(parts[1], 64)
			memU, _ := strconv.ParseFloat(parts[2], 64)
			diskT, _ := strconv.ParseFloat(parts[3], 64)
			diskU, _ := strconv.ParseFloat(parts[4], 64)
			netT, _ := strconv.ParseFloat(parts[5], 64)
			netU, _ := strconv.ParseFloat(parts[6], 64)

			// Load Average
			if la > 30 {
				fmt.Printf("Load Average is too high: %.0f\n", la)
			}

			// Memory (округляем вниз)
			if memT > 0 && memU/memT > 0.8 {
				fmt.Printf("Memory usage too high: %d%%\n", int(memU/memT*100))
			}

			// Disk (округляем вниз, проверяем >= 0.9)
			if diskT > 0 && diskU/diskT >= 0.9 {
				free := (diskT - diskU) / (1024 * 1024)
				// Используем math.Floor для надежного округления вниз
				fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Floor(free))
			}

			// Network (делим на 1,000,000 как в тесте)
			if netT > 0 && netU/netT > 0.9 {
				free := (netT - netU) / 1000000
				fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", math.Floor(free))
			}
		}

		time.Sleep(time.Minute)
	}
}
