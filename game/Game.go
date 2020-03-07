package game

import (
	"bufio"
	"fmt"
	"github.com/ksanta/wordofthedaygame/model"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	WordEntries        []model.PageDetails
	QuestionsPerGame   int
	OptionsPerQuestion int
}

func (game *Game) PlayGame() {
	score := 0

	// Randomise the random number generator
	rand.Seed(time.Now().Unix())

	wordsByType := game.groupWordsByType()

	fmt.Println("Playing", game.QuestionsPerGame, "rounds")
	for i := 1; i <= game.QuestionsPerGame; i++ {
		fmt.Printf("\nRound %v!\n", i)

		wordType := game.pickRandomWordType()

		randomWords := game.pickRandomWords(wordsByType[wordType])

		correct := game.askQuestionAndCheckResponse(randomWords)

		if correct {
			score++
		}
	}
	fmt.Println()
	fmt.Println("You scored", score, "out of", game.QuestionsPerGame)
}

// groupWordsByType is a one-time operation that converts the words slice into
// a map keyed by the word type
func (game *Game) groupWordsByType() map[string][]model.PageDetails {
	wordsByType := make(map[string][]model.PageDetails)

	for _, word := range game.WordEntries {
		wordsByType[word.WordType] = append(wordsByType[word.WordType], word)
	}

	return wordsByType
}

// todo: move this function to a method on some "[]words" interface, making it source-specific?
func (game *Game) pickRandomWordType() string {
	wordTypes := []string{"noun", "adjective", "verb", "adverb"}
	randomIndex := rand.Intn(len(wordTypes))
	return wordTypes[randomIndex]
}

func (game *Game) pickRandomWords(wordsByType []model.PageDetails) []model.PageDetails {
	chosenRandoms := make([]model.PageDetails, 0, game.OptionsPerQuestion)
	chosenWords := make(map[string]interface{})

	for len(chosenRandoms) < game.OptionsPerQuestion {
		randomIndex := rand.Intn(len(wordsByType))
		details := wordsByType[randomIndex]
		if _, present := chosenWords[details.Wotd]; !present {
			chosenRandoms = append(chosenRandoms, details)
			chosenWords[details.Wotd] = struct{}{}
		}
	}

	return chosenRandoms
}

func (game *Game) askQuestionAndCheckResponse(words []model.PageDetails) bool {
	randomWord := words[rand.Intn(len(words))]

	fmt.Println("The word of the day is:", strings.ToUpper(randomWord.Wotd))
	for i, detail := range words {
		fmt.Printf("%d) %s\n", i+1, detail.Definition)
	}
	response, timeout := promptAndGetAnswerFromPlayer()
	if timeout {
		fmt.Println("💥 Too slow! 💥")
		return false
	} else {
		correct := validateResponse(response, words, randomWord.Wotd)
		if correct {
			fmt.Println("Correct 🎉")
		} else {
			fmt.Println("Wrong! 💀💀💀")
		}
		return correct
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
		fmt.Println("Closed scanning goroutine!")
	}()

	// Slightly evil: the timeout period is random
	randomisedWait := time.Duration(10 + rand.Intn(6))
	select {
	case response = <-answerChannel:
		return response, false
	case <-time.After(randomisedWait * time.Second):
		//todo defect fix: manually send answer to stdin to force closure of above goroutine
		_, err := os.Stdin.WriteString("x\n")
		if err != nil {
			log.Fatalln(err)
		}

		return "", true
	}
}
