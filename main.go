package main

import (
	"flag"
	"fmt"
	"github.com/ksanta/wordofthedaygame/cache"
	"github.com/ksanta/wordofthedaygame/game"
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/scraper"
	"math/rand"
	"time"
)

var (
	cacheFile          = flag.String("cache", "words.cache", "Cache file name")
	limit              = flag.Int("limit", 3000, "The number of definitions to scrape")
	questionsPerGame   = flag.Int("questionsPerGame", 5, "Number of questions per game")
	optionsPerQuestion = flag.Int("optionsPerQuestion", 3, "Number of options per question")
)

func main() {
	// Parse command line args
	flag.Parse()

	// Randomise the random number generator
	rand.Seed(time.Now().Unix())

	words := obtainWordsOfTheDay()

	theGame := game.Game{
		WordEntries:        words,
		QuestionsPerGame:   *questionsPerGame,
		OptionsPerQuestion: *optionsPerQuestion,
	}

	theGame.PlayGame()
}

func obtainWordsOfTheDay() []model.WordDetail {
	myCache := cache.NewFileCache(*cacheFile)

	if myCache.DoesNotExists() {
		return populateCacheFromScraper(myCache)
	} else {
		return myCache.LoadWordsFromCache()
	}
}

func populateCacheFromScraper(myCache cache.Cache) []model.WordDetail {
	fmt.Println("Scraping words from the web (please wait)")

	var allDetails = make([]model.WordDetail, 0, *limit)

	// Start a producer of words
	myScraper := scraper.NewMeriamScraper(*limit)
	incomingWordChannel := myScraper.Scrape()

	// Create a channel that will be used to write words to the cache
	cacheChannel := myCache.CreateCacheWriter()

	// Start a consumer that will show percentage progress to the user
	progressChannel := createConsumerThatShowsPercentageComplete(*limit)

	// Capture the word into an array, and send it onwards to the CSV writer
	for wordDetail := range incomingWordChannel {
		allDetails = append(allDetails, wordDetail)
		cacheChannel <- wordDetail
		progressChannel <- true
	}
	close(cacheChannel)
	close(progressChannel)

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
