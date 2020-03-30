package player

import (
	"time"
)

type Player interface {
	// GetPlayerDetails will prompt the player for some details and save them
	GetPlayerDetails()

	// DisplayIntro will let the player know how many questions will be asked
	DisplayIntro(questionsPerGame int)

	// PresentQuestion will present to the player a word and ask them which definition they
	// think is the correct one.
	// If the player takes too long, the timeoutChan will fire.
	PresentQuestion(round int, wordToGuess string, definitions []string, timeoutChan <-chan time.Time) string

	DisplayCorrect()

	DisplayWrong()

	//DisplayProgress shows the player the number of points they received in a round
	DisplayProgress(points int)

	// DisplaySummary will show the user the number of points they got in the game
	DisplaySummary()

	AddPoints(points int)

	GetPoints() int
}
