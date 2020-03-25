package player

import "time"

type Player interface {
	// DisplayIntro will let the player know how many questions will be asked
	DisplayIntro(questionsPerGame int)

	// PresentQuestion will present to the player a word and ask them which definition they
	// think is the correct one.
	// If the player takes too long, the timeoutChan will fire.
	// The player's answer will be sent through in the response chan.
	PresentQuestion(round int, wordToGuess string, definitions []string, timeoutChan <-chan time.Time, responseChan chan string)

	DisplayCorrect()

	DisplayWrong()

	//DisplayProgress shows the player the number of points they received in a round
	DisplayProgress(points int)

	// DisplaySummary will show the user the number of points they got in the game
	DisplaySummary(totalPoints int)
}
