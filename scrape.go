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

func main() {
	// Parse command line args
	cacheFile := flag.String("cache", "words.cache", "Cache file name")
	//updateWords := flag.Bool("updateWords", false, "Whether to check web for new words")
	flag.Parse()

	// Randomise the random number generator
	rand.Seed(time.Now().Unix())

	allDetails := obtainWordOfTheDays(cacheFile)
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
		fmt.Printf("%d) %s\n", i, detail.Definition)
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
	if randomDetail.Wotd == randomDetails[responseNum].Wotd {
		fmt.Printf("Correct!")
	} else {
		fmt.Println("Wrong!")
	}
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
