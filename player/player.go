package player

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Player struct {
	Comms
	*log.Logger
	name   string
	points int
}

func NewPlayer(comms Comms) *Player {
	return &Player{
		Comms:  comms,
		Logger: log.New(os.Stdout, "[New player] ", log.Ldate|log.Ltime),
	}
}

func (p *Player) AddPoints(points int) {
	p.points += points
}

func (p *Player) GetPoints() int {
	return p.points
}

func (p *Player) SetName(name string) {
	p.name = name
	p.Logger.SetPrefix(fmt.Sprintf("[%s] ", name))
}

type Comms interface {
	// GetPlayerDetails will prompt the player for some details and return them
	GetPlayerDetails() string

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
	DisplaySummary(totalPoints int)
}
