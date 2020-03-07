package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/ksanta/wordofthedaygame/game"
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/scraper"
	"io"
	"log"
	"os"
)

var (
	cacheFile          = flag.String("cache", "words.cache", "Cache file name")
	limit              = flag.Int("limit", 3000, "The number of definitions to scrape")
	optionsPerQuestion = flag.Int("optionsPerQuestion", 3, "Number of options per question")
	questionsPerGame   = flag.Int("questionsPerGame", 5, "Number of questions per game")
)

func main() {
	// Parse command line args
	flag.Parse()

	words := obtainWordOfTheDays(cacheFile, *limit)

	theGame := game.Game{
		WordEntries:        words,
		QuestionsPerGame:   *questionsPerGame,
		OptionsPerQuestion: *optionsPerQuestion,
	}

	theGame.PlayGame()
}

func obtainWordOfTheDays(cacheFile *string, limit int) []model.PageDetails {
	var allDetails []model.PageDetails

	if fileDoesNotExists(cacheFile) {
		fmt.Println("Scraping words from the web (please wait)")
		allDetails = scrapeWordsToCacheFile(*cacheFile, limit)
	} else {
		allDetails = loadWordsFromCache(cacheFile)
	}

	return allDetails
}

func fileDoesNotExists(name *string) bool {
	_, err := os.Stat(*name)
	return os.IsNotExist(err)
}

func scrapeWordsToCacheFile(cacheFile string, limit int) []model.PageDetails {
	// Start a producer of words
	var myScraper scraper.Scraper = &scraper.MeriamScraper{}
	incomingWordChannel := myScraper.Scrape(limit)

	// Start a consumer that will write words to CSV
	csvChannel := createConsumerThatWritesToCsv(cacheFile)

	// Start a consumer that will show percentage progress to the user
	progressChannel := createConsumerThatShowsPercentageComplete(limit)

	// Capture the word into an array, and send it onwards to the CSV writer
	var allDetails []model.PageDetails
	for details := range incomingWordChannel {
		allDetails = append(allDetails, details)
		progressChannel <- true
		csvChannel <- details
	}
	close(progressChannel)
	close(csvChannel)

	return allDetails
}

func createConsumerThatShowsPercentageComplete(limit int) chan bool {
	progressChannel := make(chan bool)
	countSoFar := 0
	previousPercentage := 0

	go func() {
		for range progressChannel {
			countSoFar++
			currentPercentage := countSoFar * 100 / limit
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

func createConsumerThatWritesToCsv(cacheFile string) chan model.PageDetails {
	wordChannel := make(chan model.PageDetails)

	go func() {
		file, err := os.Create(cacheFile)
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

func loadWordsFromCache(cacheFile *string) []model.PageDetails {
	var allDetails []model.PageDetails
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
