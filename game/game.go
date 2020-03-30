package game

import (
	"bufio"
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/player"
	"math/rand"
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
	WordsByType         map[string]model.Words
	Player              player.Player
}

func (game *Game) PlayGame() {
	// Randomise the random number generator
	rand.Seed(time.Now().Unix())

	totalPoints := 0
	game.WordsByType = game.Words.GroupByType() // todo: store the words already grouped into the cache

	game.Player.GetPlayerDetails()

	game.Player.DisplayIntro(game.QuestionsPerGame)

	for i := 1; i <= game.QuestionsPerGame; i++ {
		totalPoints += game.playRound(i)
		time.Sleep(1 * time.Second) // Give the player time to prepare for the next round
	}

	game.Player.DisplaySummary(totalPoints)
}

func (game *Game) playRound(round int) int {
	wordType := game.Words.PickRandomType()
	wordsInRound := game.WordsByType[wordType].PickRandomWords(game.OptionsPerQuestion)

	wordToGuess := wordsInRound.PickRandomWord().Word
	definitions := wordsInRound.GetDefinitions()
	timeoutChan := time.After(game.DurationPerQuestion)

	startTime := time.Now()
	response := game.Player.PresentQuestion(round, wordToGuess, definitions, timeoutChan)
	elapsedTime := time.Since(startTime)

	correct := validateResponse(response, wordsInRound, wordToGuess)
	if correct {
		game.Player.DisplayCorrect()
	} else {
		game.Player.DisplayWrong()
	}

	points := game.calculatePoints(correct, elapsedTime)
	game.Player.DisplayProgress(points)

	return points
}

func (game *Game) calculatePoints(correct bool, elapsedTime time.Duration) int {
	points := 0

	if correct {
		points += 100
	}

	points += int(100 * (game.DurationPerQuestion - elapsedTime) / game.DurationPerQuestion)

	if points < 0 {
		points = 0
	}

	return points
}

func validateResponse(response string, words model.Words, correctWord string) bool {
	// If the response doesn't convert to an integer, it's wrong
	responseNum, err := strconv.Atoi(strings.TrimSpace(response))
	if err != nil {
		return false
	}

	index := responseNum - 1

	// If the response is out of range, it's wrong
	if index < 0 || index >= len(words) {
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
