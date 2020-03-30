package player

import (
	"github.com/gorilla/websocket"
	"github.com/ksanta/wordofthedaygame/model"
	"log"
	"time"
)

// WebsocketPlayer is an implementation that communicates to the player via Websockets
type WebsocketPlayer struct {
	conn          *websocket.Conn
	PlayerDetails model.PlayerDetailsResp
	Points
}

func NewWebsocketPlayer(conn *websocket.Conn) Player {
	return &WebsocketPlayer{conn: conn}
}

func (player *WebsocketPlayer) GetPlayerDetails() {
	request := model.Message{
		PlayerDetailsReq: &model.PlayerDetailsReq{},
	}
	err := player.conn.WriteJSON(request)
	if err != nil {
		log.Fatal("Send PlayerDetailsReq error:", err)
	}

	var response model.Message
	err = player.conn.ReadJSON(&response)
	if err != nil {
		log.Fatal("Receive PlayerDetailsResp error:", err)
	}

	player.PlayerDetails = *response.PlayerDetailsResp
}

func (player *WebsocketPlayer) DisplayIntro(questionsPerGame int) {
	msg := model.Message{
		Intro: &model.Intro{
			QuestionsPerGame: questionsPerGame,
		},
	}
	err := player.conn.WriteJSON(msg)
	if err != nil {
		log.Fatal("Display intro error:", err)
	}
}

func (player *WebsocketPlayer) PresentQuestion(round int, wordToGuess string, definitions []string, timeoutChan <-chan time.Time) string {
	request := model.Message{
		PresentQuestion: &model.PresentQuestion{
			Round:       round,
			WordToGuess: wordToGuess,
			Definitions: definitions,
		},
	}
	err := player.conn.WriteJSON(request)
	if err != nil {
		log.Fatal("Error presenting question:", err)
	}

	responseChan := make(chan string)
	go func() {
		var response model.Message
		err = player.conn.ReadJSON(&response)
		if err != nil {
			log.Fatal("Receive PlayerDetailsResp error:", err)
		}
		responseChan <- response.PlayerResponse.Response
	}()

	select {
	case response := <-responseChan:
		return response
	case <-timeoutChan:
		timeoutMsg := model.Message{
			Timeout: &model.Timeout{},
		}
		err := player.conn.WriteJSON(timeoutMsg)
		if err != nil {
			log.Fatal("Error sending timeout:", err)
		}
		// Still wait for user to respond after timeout
		<-responseChan
		return ""
	}
}

func (player *WebsocketPlayer) DisplayCorrect() {
	msg := model.Message{
		Correct: &model.Correct{},
	}
	err := player.conn.WriteJSON(msg)
	if err != nil {
		log.Fatal("Correct error:", err)
	}
}

func (player *WebsocketPlayer) DisplayWrong() {
	msg := model.Message{
		Wrong: &model.Wrong{},
	}
	err := player.conn.WriteJSON(msg)
	if err != nil {
		log.Fatal("Wrong error:", err)
	}
}

func (player *WebsocketPlayer) DisplayProgress(points int) {
	msg := model.Message{
		Progress: &model.Progress{
			Points: points,
		},
	}
	err := player.conn.WriteJSON(msg)
	if err != nil {
		log.Fatal("Progress error:", err)
	}
}

func (player *WebsocketPlayer) DisplaySummary() {
	msg := model.Message{
		Summary: &model.Summary{
			TotalPoints: player.GetPoints(),
		},
	}
	err := player.conn.WriteJSON(msg)
	if err != nil {
		log.Fatal("Summary error:", err)
	}
}
