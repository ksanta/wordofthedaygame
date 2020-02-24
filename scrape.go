package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/scraper"
	"io"
	"log"
	"os"
)

func main() {
	// Parse command line args
	cacheFile := flag.String("cache", "words.cache", "Cache file name")
	//updateWords := flag.Bool("updateWords", false, "Whether to check web for new words")
	flag.Parse()

	allDetails := obtainWordOfTheDays(cacheFile)

	fmt.Println("Cache contains", len(allDetails), "words")
}

func obtainWordOfTheDays(cacheFile *string) []model.PageDetails {
	// If there is no cache file, scrape from the web
	if fileDoesNotExists(cacheFile) {
		fmt.Println(*cacheFile, "does not exist")
		scrapeWordsToCacheFile(*cacheFile)
	} else {
		fmt.Println(*cacheFile, "found")
	}

	// Read all the words from the cache file
	allDetails := []model.PageDetails{}
	file, err := os.Open(*cacheFile)
	if err != nil {
		log.Fatal(err)
	}
	wordReader := csv.NewReader(file)
	for {
		record, err := wordReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		details := model.NewFromStringSlice(record)
		allDetails = append(allDetails, details)
	}

	return allDetails
}

func fileDoesNotExists(name *string) bool {
	_, err := os.Stat(*name)
	return os.IsNotExist(err)
}

func scrapeWordsToCacheFile(cacheFile string) {
	// Create the cache file
	file, err := os.Create(cacheFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Run the scraper in a goroutine
	wordChannel := make(chan model.PageDetails)
	go scraper.Scrape(wordChannel)

	// Receive a stream of words into a CSV file until the channel is closed
	csvWriter := csv.NewWriter(file)
	for details := range wordChannel {
		err := csvWriter.Write(details.ToStringSlice())
		if err != nil {
			log.Fatal(err)
		}
	}
	csvWriter.Flush()
}
