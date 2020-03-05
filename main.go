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

var cacheFile = flag.String("cache", "words.cache", "Cache file name")
var limit = flag.Int("limit", 3000, "The number of definitions to scrape")
var numOptions = flag.Int("numOptions", 3, "Number of options per question")

func main() {
	// Parse command line args
	flag.Parse()

	// Randomise the random number generator
	rand.Seed(time.Now().Unix())

	allDetails := obtainWordOfTheDays(cacheFile, *limit)

	wordType := pickRandomWordType()

	wordsByType := filterWordsByType(allDetails, wordType)

	randomWords := pickRandomWords(wordsByType, *numOptions)

	playTheGame(randomWords)
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

func pickRandomWords(wordsByType []model.PageDetails, numberToPick int) []model.PageDetails {
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

func playTheGame(words []model.PageDetails) {
	randomWord := words[rand.Intn(len(words))]

	fmt.Println("The word of the day is:", strings.ToUpper(randomWord.Wotd))
	for i, detail := range words {
		fmt.Printf("%d) %s\n", i+1, detail.Definition)
	}
	response, timeout := promptAndGetAnswerFromPlayer()
	if timeout {
		fmt.Println("💥 Too slow! 💥")
	} else {
		correct := validateResponse(response, words, randomWord.Wotd)
		if correct {
			fmt.Println("Correct 🎉")
		} else {
			fmt.Println("Wrong! 💀💀💀")
		}
	}
}

func validateResponse(response string, words []model.PageDetails, correctWord string) bool {
	responseNum, err := strconv.Atoi(response)
	if err != nil {
		return false
	}

	index := responseNum - 1
	return index >= 0 && index < len(words) && words[index].Wotd == correctWord
}

func promptAndGetAnswerFromPlayer() (response string, timeout bool) {
	fmt.Print("Enter your best guess: ")

	answerChannel := make(chan string, 1)

	// Read from player in different goroutine and send to channel
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answerChannel <- scanner.Text()
	}()

	// Slightly evil: the timeout period goes from 10-20 seconds
	randomisedWait := time.Duration(10 + rand.Intn(6))
	select {
	case response = <-answerChannel:
		return response, false
	case <-time.After(randomisedWait * time.Second):
		return "", true
	}
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
