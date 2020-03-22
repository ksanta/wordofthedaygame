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

	fmt.Println("Playing", game.QuestionsPerGame, "rounds")
	for i := 1; i <= game.QuestionsPerGame; i++ {
		fmt.Printf("\nRound %v!\n", i)

		wordType := game.Words.PickRandomType()

		randomWords := wordsByType[wordType].PickRandomWords(game.OptionsPerQuestion)

		totalPoints += game.askQuestionAndCheckResponse(randomWords)
	}
	fmt.Println()
	fmt.Println("You scored", totalPoints, "points!")
}

func (game *Game) askQuestionAndCheckResponse(words model.Words) int {
	randomWord := words.PickRandomWord()
	fmt.Println("The word of the day is:", strings.ToUpper(randomWord.Word))
	for i, word := range words {
		fmt.Printf("%d) %s\n", i+1, word.Definition)
	}
	fmt.Print("\nEnter your best guess: ")
	startTime := time.Now()

	response, stdinChannel := game.getAnswerFromPlayer()
	if stdinChannel != nil {
		fmt.Println("💥 Too slow! 💥")
		// Very important for the user to hit enter to close the goroutine listening on stdin
		fmt.Print("Hit enter to move to the next question")
		<-stdinChannel
		return 0
	} else {
		correct := validateResponse(response, words, randomWord.Word)
		if correct {
			fmt.Println("Correct 🎉")
		} else {
			fmt.Println("Wrong! 💀💀💀")
		}
		elapsedTime := time.Since(startTime)

		points := game.calculatePoints(correct, elapsedTime)
		fmt.Printf("Earned %d points\n", points)

		return points
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

func (game *Game) getAnswerFromPlayer() (response string, channelForReuse chan string) {
	stdinChannel := make(chan string, 1)

	// Get the answer  from the player in a different goroutine and send to the channel
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		stdinChannel <- scanner.Text()
	}()

	select {
	case response = <-stdinChannel:
		return response, nil
	case <-time.After(game.DurationPerQuestion):
		// On timeout, the goroutine is still blocked waiting for user input.
		// In this case, return it so the user can be prompted to hit enter to finish the goroutine
		return "", stdinChannel
	}
}
