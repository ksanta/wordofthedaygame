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
	cacheLimit         = flag.Int("cacheLimit", 3000, "The max number of words to cache")
	targetScore        = flag.Int("targetScore", 500, "Player wins when target score is reached")
	optionsPerQuestion = flag.Int("optionsPerQuestion", 3, "Number of options per question")
	addr               = flag.String("addr", ":8080", "http service address")
)

var theGame *game.Game

var upgrader websocket.Upgrader

func main() {
	flag.Parse()
	log.SetFlags(0)

	initialiseTheGame()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/game", handleNewPlayer)
	http.HandleFunc("/start", handleStartGame)
	log.Println("Listening on", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func initialiseTheGame() {
	words := obtainWordsOfTheDay()
	wordsByType := words.GroupByType()

	theGame = game.NewGame(wordsByType, *targetScore, *optionsPerQuestion, 10*time.Second, 7)

	go theGame.Run()
}

func handleNewPlayer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade fail:", err)
		return
	}
	defer conn.Close()

	// This channel will block this goroutine from exiting. If it closes, the connection will close
	disconnectChan := make(chan struct{})

	p := player.NewPlayer(conn, disconnectChan, theGame.MessageChan)

	go p.ReadPump()
	go p.WritePump()

	<-disconnectChan
	conn.Close()
}

func handleStartGame(w http.ResponseWriter, e *http.Request) {
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

	var words = make(model.Words, 0, *cacheLimit)

	// Start a producer of words
	myScraper := scraper.NewMeriamScraper(*cacheLimit)
	incomingWordChannel := myScraper.Scrape()

	// Create a channel that will be used to write words to the cache
	cacheChannel := myCache.CreateCacheWriter()

	// Start a consumer that will show percentage progress to the user
	progressChannel := createConsumerThatShowsPercentageComplete(*cacheLimit)

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
