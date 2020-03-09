package cache

import (
	"encoding/csv"
	"fmt"
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/scraper"
	"io"
	"log"
	"os"
)

type Cache struct {
	cacheFile string
	limit     int
}

// NewCache is a factory method that creates a new Cache object
func NewCache(cacheFile string, limit int) *Cache {
	return &Cache{cacheFile, limit}
}

func (cache *Cache) ObtainWordsOfTheDay() []model.WordDetail {
	var allDetails []model.WordDetail

	if cache.fileDoesNotExists() {
		fmt.Println("Scraping words from the web (please wait)")
		allDetails = cache.scrapeWordsToCacheFile()
	} else {
		allDetails = cache.loadWordsFromCache()
	}

	return allDetails
}

func (cache *Cache) fileDoesNotExists() bool {
	_, err := os.Stat(cache.cacheFile)
	return os.IsNotExist(err)
}

func (cache *Cache) scrapeWordsToCacheFile() []model.WordDetail {
	// Start a producer of words
	myScraper := scraper.NewMeriamScraper(cache.limit)
	incomingWordChannel := myScraper.Scrape()

	// Start a consumer that will write words to CSV
	csvChannel := cache.createConsumerThatWritesToCsv()

	// Start a consumer that will show percentage progress to the user
	progressChannel := cache.createConsumerThatShowsPercentageComplete()

	// Capture the word into an array, and send it onwards to the CSV writer
	var allDetails []model.WordDetail
	for details := range incomingWordChannel {
		allDetails = append(allDetails, details)
		progressChannel <- true
		csvChannel <- details
	}
	close(progressChannel)
	close(csvChannel)

	return allDetails
}

func (cache *Cache) createConsumerThatShowsPercentageComplete() chan bool {
	progressChannel := make(chan bool)
	countSoFar := 0
	previousPercentage := 0

	go func() {
		for range progressChannel {
			countSoFar++
			currentPercentage := countSoFar * 100 / cache.limit
			// Only update the value if there is a change, to minimise flickering
			if currentPercentage != previousPercentage {
				fmt.Printf("\r%5v%%", currentPercentage)
			}
			previousPercentage = currentPercentage
		}
		fmt.Println()
	}()

	return progressChannel
}

func (cache *Cache) createConsumerThatWritesToCsv() chan model.WordDetail {
	wordChannel := make(chan model.WordDetail)

	go func() {
		file, err := os.Create(cache.cacheFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		csvWriter := csv.NewWriter(file)
		for details := range wordChannel {
			err := csvWriter.Write(details.ToStringSlice())
			if err != nil {
				log.Fatal(err)
			}
		}
		csvWriter.Flush()
	}()

	return wordChannel
}

func (cache *Cache) loadWordsFromCache() []model.WordDetail {
	var allDetails []model.WordDetail
	file, err := os.Open(cache.cacheFile)
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
