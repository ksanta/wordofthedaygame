package player

import (
	"bufio"
	"fmt"
	"github.com/ksanta/wordofthedaygame/model"
	"os"
	"strings"
	"time"
)

type ConsoleCommunication struct {
	PlayerDetails model.PlayerDetails
}

func NewConsoleCommunication() Comms {
	return &ConsoleCommunication{}
}

func (consoleComm *ConsoleCommunication) GetPlayerDetails() string {
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func (consoleComm *ConsoleCommunication) DisplayIntro(questionsPerGame int) {
	fmt.Println("Playing", questionsPerGame, "rounds")
}

func (consoleComm *ConsoleCommunication) PresentQuestion(round int, wordToGuess string, definitions []string, timeoutChan <-chan time.Time) string {
	fmt.Println()
	fmt.Printf("Round %d!\n", round)
	fmt.Println("The word of the day is:", strings.ToUpper(wordToGuess))

	for i, definition := range definitions {
		fmt.Printf("%d) %s\n", i+1, definition)
	}
	fmt.Print("\nEnter your best guess: ")

	return consoleComm.getAnswerFromPlayer(timeoutChan)
}

func (consoleComm *ConsoleCommunication) DisplayCorrect() {
	fmt.Println("Correct 🎉")
}

func (consoleComm *ConsoleCommunication) DisplayWrong() {
	fmt.Println("Wrong! 💀💀💀")
}

func (consoleComm *ConsoleCommunication) DisplayProgress(points int) {
	fmt.Printf("Earned %d points\n", points)
}

func (consoleComm *ConsoleCommunication) DisplaySummary(totalPoints int) {
	fmt.Println()
	fmt.Println("You scored", totalPoints, "points!")
}

func (consoleComm *ConsoleCommunication) getAnswerFromPlayer(timeoutChan <-chan time.Time) string {
	stdinChannel := make(chan string, 1)

	// Get the answer from the player in a different goroutine and send to the channel
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		stdinChannel <- scanner.Text()
	}()

	select {
	case response := <-stdinChannel:
		return response
	case <-timeoutChan:
		fmt.Println("💥 Too slow! 💥")
		// On timeout, the goroutine is still blocked waiting for user input.
		// Prompt the player to hit enter to finish the goroutine
		fmt.Print("Hit enter to move to the next question")
		<-stdinChannel
		return ""
	}
}
