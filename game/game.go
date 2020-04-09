package game

import (
	"github.com/ksanta/wordofthedaygame/model"
	"github.com/ksanta/wordofthedaygame/player"
	"log"
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
	MaxPlayerCount      int
	MessageChan         chan player.PlayerMessage
	StartChan           chan struct{}
	players             player.Players
	waitGroup           sync.WaitGroup
	wordsInRound        model.Words
	wordToGuess         string
	gameInProgress      bool
	waitingForAnswers   bool
}

func NewGame(wordsByType map[string]model.Words,
	targetScore int,
	optionsPerQuestion int,
	durationPerQuestion time.Duration,
	maxPlayerCount int) *Game {
	rand.Seed(time.Now().Unix())

	return &Game{
		WordsByType:         wordsByType,
		TargetScore:         targetScore,
		OptionsPerQuestion:  optionsPerQuestion,
		DurationPerQuestion: durationPerQuestion,
		MaxPlayerCount:      maxPlayerCount,
		MessageChan:         make(chan player.PlayerMessage),
		// Buffer on StartChan required because same thread can send/receive
		StartChan:         make(chan struct{}, 1),
		players:           make([]*player.Player, 0, 10),
		waitGroup:         sync.WaitGroup{},
		wordsInRound:      nil,
		wordToGuess:       "",
		gameInProgress:    false,
		waitingForAnswers: false,
	}
}

// Run will start listening on its channels. This is meant to be called as a goroutine.
func (game *Game) Run() {
	for {
		select {
		case playerMessage := <-game.MessageChan:
			switch {
			case playerMessage.Message.PlayerDetailsResp != nil:
				game.handlePlayerReady(playerMessage)

			case playerMessage.Message.PlayerResponse != nil:
				game.handlePlayerResponse(playerMessage.Player, playerMessage.Message.PlayerResponse.Response)

			case playerMessage.Message.Disconnected != nil:
				// Player sent the game a Disconnect msg because the connection was lost
				game.safelyUnregisterPlayer(playerMessage.Player)
			}

		case <-game.StartChan:
			if game.players.NumActivePlayers() > 0 {
				go game.PlayGame()
			}
		}
	}
}

func (game *Game) handlePlayerReady(playerMessage player.PlayerMessage) {
	// Prevent player from registering if there is a game in progress
	if game.gameInProgress {
		messageToPlayer := model.MessageToPlayer{
			Error: &model.GameError{
				Message: "Game is already in progress",
			},
		}
		playerMessage.Player.SendToClientChan <- messageToPlayer
		return
	}

	// Player has sent their name - they are ready to play
	p := playerMessage.Player
	p.SetName(playerMessage.Message.PlayerDetailsResp.Name)
	p.Icon = playerMessage.Message.PlayerDetailsResp.Icon
	p.Active = true
	game.players = append(game.players, p)
	game.sendWelcomeToPlayer(p)

	// Sending round summary now alerts each player when new players join
	game.sendRoundSummaryToEachPlayer()

	// Auto-start the game if there are N players ready
	if game.players.NumActivePlayers() == game.MaxPlayerCount {
		game.StartChan <- struct{}{}
	}
}

func (game *Game) safelyUnregisterPlayer(p *player.Player) {
	close(p.SendToClientChan)
	if game.waitingForAnswers {
		game.waitGroup.Done()
	}
	p.Active = false

	// Reset the game if all players have become inactive
	if game.players.AllInactive() {
		game.reset()
	}
}

func (game *Game) requestPlayerName(p *player.Player) {
	message := model.MessageToPlayer{
		PlayerDetailsReq: &model.PlayerDetailsReq{},
	}
	p.SendToClientChan <- message
}

func (game *Game) sendWelcomeToPlayer(p *player.Player) {
	// Sending messages to the player must be done via channel
	p.SendToClientChan <- model.MessageToPlayer{
		Welcome: &model.Welcome{TargetScore: game.TargetScore},
	}
}

func (game *Game) AlertPlayersGameWillBegin() {
	const waitSeconds = 5

	alertPlayers := func(p *player.Player) {
		p.SendToClientChan <- model.MessageToPlayer{
			AboutToStart: &model.AboutToStart{
				Seconds: waitSeconds,
			},
		}
	}
	game.players.ForActivePlayers(alertPlayers)

	time.Sleep(time.Duration(waitSeconds) * time.Second)
}

// PlayGame orchestrates the rounds of questions and displays the result to all players.
// This runs in a goroutine.
func (game *Game) PlayGame() {
	log.Println("Starting game")

	game.gameInProgress = true
	game.AlertPlayersGameWillBegin()

	maxScore := 0
	for maxScore <= game.TargetScore {
		game.playRound()

		if game.players.AllInactive() {
			break
		}

		maxScore = game.players.PlayerWithHighestPoints().GetPoints()
		time.Sleep(2 * time.Second) // Give the players time to prepare for the next round
	}

	game.sendGameSummaryToPlayers()
	game.gameInProgress = false
	game.reset()
}

func (game *Game) playRound() {
	game.sendQuestionToEachPlayer()

	// Wait for all players to send in their responses
	game.waitingForAnswers = true
	game.waitGroup.Wait()
	game.waitingForAnswers = false

	game.sendRoundSummaryToEachPlayer()
}

func (game *Game) sendGameSummaryToPlayers() {
	winner := game.players.PlayerWithHighestPoints()

	sendSummary := func(p *player.Player) {
		p.SendToClientChan <- model.MessageToPlayer{
			Summary: &model.Summary{
				Winner:      winner.GetName(),
				Icon:        winner.Icon,
				TotalPoints: winner.GetPoints(),
			},
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

// validateResponse returns true if the response is correct
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
	// Player took longer than allowed time - no points!
	if elapsedTime > game.DurationPerQuestion {
		return 0
	}

	correctPoints := 0
	if correct {
		correctPoints += 100
	}

	timePoints := int(50 * (game.DurationPerQuestion - elapsedTime) / game.DurationPerQuestion)
	if timePoints < 0 {
		timePoints = 0
	}

	return correctPoints + timePoints
}

// reset will reset the game state
func (game *Game) reset() {
	log.Println("Removing all players")
	game.players = make([]*player.Player, 0, 10)
}
