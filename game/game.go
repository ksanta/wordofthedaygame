package game

import (
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/player"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Game struct {
	WordsByType         map[string]model.Words
	QuestionsPerGame    int
	OptionsPerQuestion  int
	DurationPerQuestion time.Duration
	RegisterChan        chan *player.Player
	UnregisterChan      chan *player.Player
	MessageChan         chan player.PlayerMessage
	StartChan           chan struct{}
	players             map[*player.Player]bool
	waitGroup           sync.WaitGroup
	wordsInRound        model.Words
	wordToGuess         string
}

func NewGame(wordsByType map[string]model.Words,
	questionsPerGame int,
	optionsPerQuestion int,
	durationPerQuestion time.Duration) *Game {
	rand.Seed(time.Now().Unix())

	return &Game{
		WordsByType:         wordsByType,
		QuestionsPerGame:    questionsPerGame,
		OptionsPerQuestion:  optionsPerQuestion,
		DurationPerQuestion: durationPerQuestion,
		RegisterChan:        make(chan *player.Player),
		UnregisterChan:      make(chan *player.Player),
		MessageChan:         make(chan player.PlayerMessage),
		StartChan:           make(chan struct{}),
		players:             make(map[*player.Player]bool),
		waitGroup:           sync.WaitGroup{},
		wordsInRound:        nil,
		wordToGuess:         "",
	}
}

// Run will start listening on its channels. This is meant to be called as a goroutine.
func (game *Game) Run() {
	for {
		select {
		case p := <-game.RegisterChan:
			game.players[p] = true
			game.requestPlayerName(p)
		case p := <-game.UnregisterChan:
			if _, ok := game.players[p]; ok {
				delete(game.players, p)
				close(p.SendToClientChan)
			}
		case <-game.StartChan:
			go game.PlayGame()
		case playerMessage := <-game.MessageChan:
			switch {
			case playerMessage.Message.PlayerDetailsResp != nil:
				p := playerMessage.Player
				p.SetName(playerMessage.Message.PlayerDetailsResp.Name)
				game.sendIntroToPlayer(p)
			case playerMessage.Message.PlayerResponse != nil:
				game.handlePlayerResponse(playerMessage.Player, playerMessage.Message.PlayerResponse.Response)
			}
		}
	}
}

func (game *Game) requestPlayerName(p *player.Player) {
	message := model.MessageToPlayer{
		PlayerDetailsReq: &model.PlayerDetailsReq{},
	}
	p.SendToClientChan <- message
}

func (game *Game) sendIntroToPlayer(p *player.Player) {
	// Sending messages to the player must be done via channel
	p.SendToClientChan <- model.MessageToPlayer{
		Intro: &model.Intro{QuestionsPerGame: game.QuestionsPerGame},
	}
}

// PlayGame orchestrates the rounds of questions and displays the result to all players
func (game *Game) PlayGame() {
	for i := 1; i <= game.QuestionsPerGame; i++ {
		game.playRound(i)
		time.Sleep(1 * time.Second) // Give the players time to prepare for the next round
	}

	game.sendSummaryToPlayers()
}

func (game *Game) sendSummaryToPlayers() {
	for p := range game.players {
		p.SendToClientChan <- model.MessageToPlayer{
			Summary: &model.Summary{TotalPoints: p.GetPoints()},
		}
	}
}

func (game *Game) playRound(round int) {
	wordType := model.PickRandomType()

	game.wordsInRound = game.WordsByType[wordType].PickRandomWords(game.OptionsPerQuestion)
	game.wordToGuess = game.wordsInRound.PickRandomWord().Word

	definitions := game.wordsInRound.GetDefinitions()

	// Set Game wait group to player count
	game.waitGroup.Add(len(game.players))

	// For each player
	for p := range game.players {
		//   Start player timer
		p.StartTimer()

		//   Send question
		p.SendToClientChan <- model.MessageToPlayer{
			PresentQuestion: &model.PresentQuestion{
				Round:       round,
				WordToGuess: game.wordToGuess,
				Definitions: definitions,
			},
		}
	}
	// Wait for wait group
	game.waitGroup.Wait()
}

func (game *Game) handlePlayerResponse(p *player.Player, response string) {
	elapsedTime := p.StopTimer()

	correct := game.validateResponse(response)

	if correct {
		p.SendToClientChan <- model.MessageToPlayer{
			Correct: &model.Correct{},
		}
	} else {
		p.SendToClientChan <- model.MessageToPlayer{
			Wrong: &model.Wrong{},
		}
	}

	points := game.calculatePoints(correct, elapsedTime)
	p.AddPoints(points)

	p.SendToClientChan <- model.MessageToPlayer{
		Progress: &model.Progress{Points: points},
	}

	game.waitGroup.Done()
}

func (game *Game) validateResponse(response string) bool {
	// If the response doesn't convert to an integer, it's wrong
	responseNum, err := strconv.Atoi(strings.TrimSpace(response))
	if err != nil {
		return false
	}

	index := responseNum - 1

	// If the response is out of range, it's wrong
	if index < 0 || index >= len(game.wordsInRound) {
		return false
	}

	// Compare the response to the correct answer
	return game.wordsInRound[index].Word == game.wordToGuess
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
