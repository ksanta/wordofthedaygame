package main

import (
	"flag"
	"github.com/ksanta/wordofthedaygame/cache"
	"github.com/ksanta/wordofthedaygame/game"
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

	c := cache.NewCache(*cacheFile, *limit)
	words := c.ObtainWordsOfTheDay()

	theGame := game.Game{
		WordEntries:        words,
		QuestionsPerGame:   *questionsPerGame,
		OptionsPerQuestion: *optionsPerQuestion,
	}

	theGame.PlayGame()
}
