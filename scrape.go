package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/scraper"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var limit = flag.Int("limit", 3000, "The number of definitions to scrape")
var cacheFile = flag.String("cache", "words.cache", "Cache file name")

func main() {
	// Parse command line args
	flag.Parse()

	// Randomise the random number generator
	rand.Seed(time.Now().Unix())

	allDetails := obtainWordOfTheDays(cacheFile, *limit)
	fmt.Println("FYI, the cache file contains", len(allDetails), "words")

	wordType := pickRandomWordType()
	fmt.Printf("Today's word is a %s!\n", wordType)

	wordsByType := filterWordsByType(allDetails, wordType)

	threeRandoms := pickSomeAtRandom(wordsByType, 3)

	playTheGame(threeRandoms)
}

func pickRandomWordType() string {
	wordTypes := []string{"noun", "adjective", "verb", "adverb"}
	randomIndex := rand.Intn(len(wordTypes))
	return wordTypes[randomIndex]
}

func filterWordsByType(allDetails []model.PageDetails, wordType string) []model.PageDetails {
	var filteredDetails []model.PageDetails
	for _, details := range allDetails {
		if details.WordType == wordType {
			filteredDetails = append(filteredDetails, details)
		}
	}
	return filteredDetails
}

func pickSomeAtRandom(wordsByType []model.PageDetails, numberToPick int) []model.PageDetails {
	chosenRandoms := make([]model.PageDetails, 0, numberToPick)
	chosenWords := make(map[string]interface{})

	for len(chosenRandoms) < numberToPick {
		randomIndex := rand.Intn(len(wordsByType))
		details := wordsByType[randomIndex]
		if _, present := chosenWords[details.Wotd]; !present {
			chosenRandoms = append(chosenRandoms, details)
			chosenWords[details.Wotd] = struct{}{}
		}
	}

	return chosenRandoms
}

func playTheGame(randomDetails []model.PageDetails) {
	randomDetail := randomDetails[rand.Intn(len(randomDetails))]

	fmt.Println("The word of the day is:", strings.ToUpper(randomDetail.Wotd))
	for i, detail := range randomDetails {
		fmt.Printf("%d) %s\n", i+1, detail.Definition)
	}
	fmt.Print("Enter your best guess: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := scanner.Text()
	responseNum, err := strconv.Atoi(response)
	if err != nil {
		log.Fatal(err)
	}
	// todo response validation
	if randomDetail.Wotd == randomDetails[responseNum-1].Wotd {
		fmt.Printf("Correct!")
	} else {
		fmt.Println("Wrong!")
	}
}

func obtainWordOfTheDays(cacheFile *string, limit int) []model.PageDetails {
	var allDetails []model.PageDetails

	if fileDoesNotExists(cacheFile) {
		fmt.Println(*cacheFile, "does not exist")
		allDetails = scrapeWordsToCacheFile(*cacheFile, limit)
	} else {
		fmt.Println(*cacheFile, "found")
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
	incomingWordChannel := myScraper.StartScraping(limit)

	// Start a consumer that will write words to CSV
	outgoingWordChannel := createConsumerThatWritesToCsv(cacheFile)

	// Capture the word into an array, and send it onwards to the CSV writer
	var allDetails []model.PageDetails
	for details := range incomingWordChannel {
		allDetails = append(allDetails, details)
		outgoingWordChannel <- details
	}

	return allDetails
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
			csvWriter.Flush()
		}
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
