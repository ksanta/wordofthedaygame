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
	TargetScore         int
	OptionsPerQuestion  int
	DurationPerQuestion time.Duration
	RegisterChan        chan *player.Player
	UnregisterChan      chan *player.Player
	MessageChan         chan player.PlayerMessage
	StartChan           chan struct{}
	players             player.Players
	waitGroup           sync.WaitGroup
	wordsInRound        model.Words
	wordToGuess         string
	gameInProgress      bool
}

func NewGame(wordsByType map[string]model.Words,
	targetScore int,
	optionsPerQuestion int,
	durationPerQuestion time.Duration) *Game {
	rand.Seed(time.Now().Unix())

	return &Game{
		WordsByType:         wordsByType,
		TargetScore:         targetScore,
		OptionsPerQuestion:  optionsPerQuestion,
		DurationPerQuestion: durationPerQuestion,
		RegisterChan:        make(chan *player.Player),
		UnregisterChan:      make(chan *player.Player),
		MessageChan:         make(chan player.PlayerMessage),
		StartChan:           make(chan struct{}),
		players:             make([]*player.Player, 0, 10),
		waitGroup:           sync.WaitGroup{},
		wordsInRound:        nil,
		wordToGuess:         "",
		gameInProgress:      false,
	}
}

// Run will start listening on its channels. This is meant to be called as a goroutine.
func (game *Game) Run() {
	for {
		select {
		case p := <-game.RegisterChan:
			if game.gameInProgress {
				// todo message/call to player to try again later
			}
			game.players = append(game.players, p)
			game.requestPlayerName(p)
		case p := <-game.UnregisterChan:
			p.Active = false
			close(p.SendToClientChan)
			game.waitGroup.Done()
		case <-game.StartChan:
			go game.PlayGame()
		case playerMessage := <-game.MessageChan:
			switch {
			case playerMessage.Message.PlayerDetailsResp != nil:
				p := playerMessage.Player
				p.SetName(playerMessage.Message.PlayerDetailsResp.Name)
				p.Active = true
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
		Intro: &model.Intro{TargetScore: game.TargetScore},
	}
}

// PlayGame orchestrates the rounds of questions and displays the result to all players
func (game *Game) PlayGame() {
	maxScore := 0
	for maxScore <= game.TargetScore {
		game.playRound()
		maxScore = game.players.MaxScore()
		time.Sleep(2 * time.Second) // Give the players time to prepare for the next round
	}

	game.sendSummaryToPlayers()

	// todo: remove all players at the end of the game?
	// todo: split up static/dynamic game fields?
}

func (game *Game) playRound() {
	game.sendQuestionToEachPlayer()

	// Wait for all players to send in their responses
	game.waitGroup.Wait()

	game.sendRoundSummaryToEachPlayer()
}

func (game *Game) sendSummaryToPlayers() {
	sendSummary := func(p *player.Player) {
		p.SendToClientChan <- model.MessageToPlayer{
			Summary: &model.Summary{TotalPoints: p.GetPoints()},
		}
	}

	game.players.ForActivePlayers(sendSummary)
}

func (game *Game) sendQuestionToEachPlayer() {
	wordType := model.PickRandomType()

	game.wordsInRound = game.WordsByType[wordType].PickRandomWords(game.OptionsPerQuestion)
	game.wordToGuess = game.wordsInRound.PickRandomWord().Word

	definitions := game.wordsInRound.GetDefinitions()

	// Wait group keeps track of how many responses to wait for
	game.waitGroup.Add(game.players.NumActivePlayers())

	questionMsg := model.MessageToPlayer{
		PresentQuestion: &model.PresentQuestion{
			WordToGuess:    game.wordToGuess,
			Definitions:    definitions,
			SecondsAllowed: int(game.DurationPerQuestion.Seconds()),
		},
	}

	sendQuestion := func(p *player.Player) {
		p.StartTimer()
		p.SendToClientChan <- questionMsg
	}

	game.players.ForActivePlayers(sendQuestion)
}

func (game *Game) sendRoundSummaryToEachPlayer() {

	playerStates := make([]model.PlayerState, 0, len(game.players))

	for _, p := range game.players {
		playerStates = append(playerStates, p.PlayerState())
	}

	roundSummary := model.MessageToPlayer{
		RoundSummary: &model.RoundSummary{
			PlayerStates: playerStates,
		},
	}

	sendRoundSummary := func(p *player.Player) {
		p.SendToClientChan <- roundSummary
	}

	game.players.ForActivePlayers(sendRoundSummary)
}

func (game *Game) handlePlayerResponse(p *player.Player, response string) {
	elapsedTime := p.StopTimer()

	correct := game.validateResponse(response)

	points := game.calculatePoints(correct, elapsedTime)
	p.AddPoints(points)

	// Immediately send the result to the player
	p.SendToClientChan <- model.MessageToPlayer{
		PlayerResult: &model.PlayerResult{
			Correct: correct,
			Points:  points,
		},
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
	correctPoints := 0

	if correct {
		correctPoints += 100
	}

	timePoints := int(100 * (game.DurationPerQuestion - elapsedTime) / game.DurationPerQuestion)

	if timePoints < 0 {
		timePoints = 0
	}

	return correctPoints + timePoints
}
