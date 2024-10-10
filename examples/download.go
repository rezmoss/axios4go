package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rezmoss/axios4go"
)

func main() {
	url := "https://ash-speed.hetzner.com/1GB.bin"
	outputPath := "1GB.bin"

	startTime := time.Now()
	lastPrintTime := startTime

	resp, err := axios4go.Get(url, &axios4go.RequestOptions{
		MaxContentLength: 5 * 1024 * 1024 * 1024, // 5GB
		Timeout:          60000 * 5,
		OnDownloadProgress: func(bytesRead, totalBytes int64) {
			currentTime := time.Now()
			if currentTime.Sub(lastPrintTime) >= time.Second || bytesRead == totalBytes {
				percentage := float64(bytesRead) / float64(totalBytes) * 100
				downloadedMB := float64(bytesRead) / 1024 / 1024
				totalMB := float64(totalBytes) / 1024 / 1024
				elapsedTime := currentTime.Sub(startTime)
				speed := float64(bytesRead) / elapsedTime.Seconds() / 1024 / 1024 // MB/s

				fmt.Printf("\rDownloaded %.2f%% (%.2f MB / %.2f MB) - Speed: %.2f MB/s",
					percentage, downloadedMB, totalMB, speed)

				lastPrintTime = currentTime
			}
		},
	})

	if err != nil {
		fmt.Printf("\nError downloading file: %v\n", err)
		return
	}

	err = writeResponseToFile(resp, outputPath)
	if err != nil {
		fmt.Printf("\nError writing file: %v\n", err)
		return
	}

	fmt.Println("\nDownload completed successfully!!")
}

func writeResponseToFile(resp *axios4go.Response, outputPath string) error {
	return os.WriteFile(outputPath, resp.Body, 0644)
}
