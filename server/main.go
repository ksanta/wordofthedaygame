package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ksanta/wordofthedaygame/cache"
	"github.com/ksanta/wordofthedaygame/game"
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/player"
	"github.com/ksanta/wordofthedaygame/scraper"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	cacheType          = flag.String("cacheType", "file", "Must be 'file' for now")
	cacheFile          = flag.String("cache", "words.cache", "Cache file name")
	limit              = flag.Int("limit", 3000, "The number of definitions to scrape")
	questionsPerGame   = flag.Int("questionsPerGame", 5, "Number of questions per game")
	optionsPerQuestion = flag.Int("optionsPerQuestion", 3, "Number of options per question")
	addr               = flag.String("addr", "localhost:8080", "http service address")
)

var theGame *game.Game

var upgrader websocket.Upgrader

func main() {
	flag.Parse()
	log.SetFlags(0)

	initialiseTheGame()

	http.HandleFunc("/game", handleNewPlayer)
	http.HandleFunc("/start", handleStartGame)
	log.Println("Listening on", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func initialiseTheGame() {
	words := obtainWordsOfTheDay()
	wordsByType := words.GroupByType() // todo: store the words already grouped into the cache

	theGame = game.NewGame(wordsByType, *questionsPerGame, *optionsPerQuestion, 10*time.Second)

	go theGame.Run()
}

func handleNewPlayer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade fail:", err)
		return
	}
	defer conn.Close()

	p := player.NewPlayer(conn, theGame.UnregisterChan, theGame.MessageChan)

	go p.ReadPump()
	go p.WritePump()

	theGame.RegisterChan <- p

	// Lock indefinitely as a test
	channel := make(chan bool)
	<-channel
}

func handleStartGame(w http.ResponseWriter, e *http.Request) {
	log.Println("Starting game")
	theGame.StartChan <- struct{}{}
	_, err := fmt.Fprint(w, "Game started")
	if err != nil {
		panic(err)
	}
}

func obtainWordsOfTheDay() model.Words {
	var myCache cache.Cache
	if *cacheType == "file" {
		myCache = cache.NewFileCache(*cacheFile)
	} else {
		fmt.Println("Invalid cache type provided")
		os.Exit(1)
	}

	if myCache.SetupRequired() {
		return scrapeAndPopulateCache(myCache)
	} else {
		return myCache.LoadWordsFromCache()
	}
}

func scrapeAndPopulateCache(myCache cache.Cache) model.Words {
	fmt.Println("Scraping words from the web (please wait)")

	var words = make(model.Words, 0, *limit)

	// Start a producer of words
	myScraper := scraper.NewMeriamScraper(*limit)
	incomingWordChannel := myScraper.Scrape()

	// Create a channel that will be used to write words to the cache
	cacheChannel := myCache.CreateCacheWriter()

	// Start a consumer that will show percentage progress to the user
	progressChannel := createConsumerThatShowsPercentageComplete(*limit)

	// Capture the word into an array, and send it onwards to the CSV writer
	for word := range incomingWordChannel {
		words = append(words, word)
		cacheChannel <- word
		progressChannel <- true
	}
	close(cacheChannel)
	close(progressChannel)

	return words
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
