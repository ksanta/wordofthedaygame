package game

import (
	"bufio"
	"fmt"
	"github.com/ksanta/wordofthedaygame/model"
	"os"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	Words               model.Words
	QuestionsPerGame    int
	OptionsPerQuestion  int
	DurationPerQuestion time.Duration
}

func (game *Game) PlayGame() {
	totalPoints := 0

	wordsByType := game.Words.GroupByType()
	var stdinChannel chan string

	fmt.Println("Playing", game.QuestionsPerGame, "rounds")
	for i := 1; i <= game.QuestionsPerGame; i++ {
		fmt.Printf("\nRound %v!\n", i)

		wordType := game.Words.PickRandomType()

		randomWords := wordsByType[wordType].PickRandomWords(game.OptionsPerQuestion)

		var points int
		points, stdinChannel = game.askQuestionAndCheckResponse(randomWords, stdinChannel)

		totalPoints += points
	}
	fmt.Println()
	fmt.Println("You scored", totalPoints, "points!")
}

func (game *Game) askQuestionAndCheckResponse(words model.Words, stdinChannel chan string) (points int, stdinChan chan string) {
	randomWord := words.PickRandomWord()

	startTime := time.Now()

	fmt.Println("The word of the day is:", strings.ToUpper(randomWord.Word))
	for i, word := range words {
		fmt.Printf("%d) %s\n", i+1, word.Definition)
	}

	response, stdinChannel := promptAndGetAnswerFromPlayer(stdinChannel, game.DurationPerQuestion)
	if stdinChannel != nil {
		fmt.Println("💥 Too slow! 💥")
		return 0, stdinChannel
	} else {
		correct := validateResponse(response, words, randomWord.Word)
		if correct {
			fmt.Println("Correct 🎉")
		} else {
			fmt.Println("Wrong! 💀💀💀")
		}
		elapsedTime := time.Since(startTime)

		points = game.calculatePoints(correct, elapsedTime)
		fmt.Printf("Earned %d points\n", points)

		return points, nil
	}
}

func (game *Game) calculatePoints(correct bool, elapsedTime time.Duration) int {
	points := 0

	if correct {
		points += 100
	}

	points += int(100 * (game.DurationPerQuestion - elapsedTime) / game.DurationPerQuestion)

	return points
}

func validateResponse(response string, words model.Words, correctWord string) bool {
	// If the response doesn't convert to an integer, it's wrong
	responseNum, err := strconv.Atoi(response)
	if err != nil {
		return false
	}

	index := responseNum - 1

	// If the response is out of range, it's wrong
	if index < 0 && index >= len(words) {
		return false
	}

	// Compare the response to the correct answer
	return words[index].Word == correctWord
}

func promptAndGetAnswerFromPlayer(stdinChannel chan string, waitTime time.Duration) (response string, channelForReuse chan string) {
	fmt.Print("\nEnter your best guess: ")

	// If the previous question timed out, a goroutine waiting for user input still exists and must
	// read something to finish, so we reuse it for the next question

	if stdinChannel == nil {
		stdinChannel = make(chan string, 1)

		// Get the answer  from the player in a different goroutine and send to the channel
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			stdinChannel <- scanner.Text()
		}()
	}

	select {
	case response = <-stdinChannel:
		return response, nil
	case <-time.After(waitTime):
		// On timeout, the goroutine is still blocked waiting for user input.
		// In this case, save the channel for the next question
		return "", stdinChannel
	}
}
