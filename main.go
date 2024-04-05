package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	concurrencyPtr := flag.Int("c", 30, "The number of concurrent downloads to run")
	outputFilePtr := flag.String("o", "", "The name and location of the output file")
	flag.Parse()

	url := "https://chaos-data.projectdiscovery.io/index.json"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var data []map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	outputFileName := *outputFilePtr
	if outputFileName == "" {
		// Generate output file name based on current date
		currentDate := time.Now().Format("02-01-2006")
		outputFileName = "chaos-output-" + currentDate + ".txt"
	}

	file, err := os.Create(outputFileName)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer file.Close()

	var wg sync.WaitGroup
	wg.Add(len(data))

	concurrency := make(chan struct{}, *concurrencyPtr)

	for _, obj := range data {
		zipUrl := obj["URL"].(string)

		concurrency <- struct{}{}

		go func(url string) {
			defer func() {
				<-concurrency
				wg.Done()
			}()

			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error making GET request to zip file:", err)
				return
			}
			defer resp.Body.Close()

			zipData, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading zip file contents:", err)
				return
			}

			zipReader := bytes.NewReader(zipData)

			reader, err := zip.NewReader(zipReader, int64(len(zipData)))
			if err != nil {
				fmt.Println("Error creating zip reader:", err)
				return
			}
			for _, zipFile := range reader.File {
				fileContents, err := zipFile.Open()
				if err != nil {
					fmt.Println("Error opening file in zip archive:", err)
					return
				}
				_, err = io.Copy(file, fileContents)
				if err != nil {
					fmt.Println("Error writing to output file:", err)
					return
				}
				fileContents.Close()
			}
			fmt.Printf("Downloaded and extracted %s to %s\n", url, outputFileName)
		}(zipUrl)
	}

	wg.Wait()

	fmt.Println("All downloads completed successfully")
}
