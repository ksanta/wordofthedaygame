package player

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ksanta/wordofthedaygame/model"
	"time"
)

// WebsocketCommunication is an implementation that communicates to the player via Websockets
type WebsocketCommunication struct {
	conn *websocket.Conn
}

func NewWebsocketCommunication(conn *websocket.Conn) Comms {
	return &WebsocketCommunication{conn: conn}
}

func (wsComm *WebsocketCommunication) GetPlayerDetails() string {
	request := model.Message{
		PlayerDetailsReq: &model.PlayerDetailsReq{},
	}
	err := wsComm.conn.WriteJSON(request)
	if err != nil {
		panic(fmt.Sprint("Send PlayerDetailsReq error:", err))
	}

	var response model.Message
	err = wsComm.conn.ReadJSON(&response)
	if err != nil {
		panic(fmt.Sprint("Receive PlayerDetails error:", err))
	}

	return response.PlayerDetailsResp.Name
}

func (wsComm *WebsocketCommunication) DisplayIntro(questionsPerGame int) {
	msg := model.Message{
		Intro: &model.Intro{
			QuestionsPerGame: questionsPerGame,
		},
	}
	err := wsComm.conn.WriteJSON(msg)
	if err != nil {
		panic(fmt.Sprint("Display intro error:", err))
	}
}

func (wsComm *WebsocketCommunication) PresentQuestion(round int, wordToGuess string, definitions []string, timeoutChan <-chan time.Time) string {
	request := model.Message{
		PresentQuestion: &model.PresentQuestion{
			Round:       round,
			WordToGuess: wordToGuess,
			Definitions: definitions,
		},
	}
	err := wsComm.conn.WriteJSON(request)
	if err != nil {
		panic(fmt.Sprint("Error presenting question:", err))
	}

	responseChan := make(chan string)
	errorChan := make(chan error)
	go func() {
		var response model.Message
		err = wsComm.conn.ReadJSON(&response)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- response.PlayerResponse.Response
	}()

	select {
	case err := <-errorChan:
		panic(fmt.Sprint("Error receiving player response:", err))
		return ""
	case response := <-responseChan:
		return response
	case <-timeoutChan:
		timeoutMsg := model.Message{
			Timeout: &model.Timeout{},
		}
		err := wsComm.conn.WriteJSON(timeoutMsg)
		if err != nil {
			panic(fmt.Sprint("Error sending timeout:", err))
		}
		// Still wait for user to respond after timeout
		<-responseChan
		return ""
	}
}

func (wsComm *WebsocketCommunication) DisplayCorrect() {
	msg := model.Message{
		Correct: &model.Correct{},
	}
	err := wsComm.conn.WriteJSON(msg)
	if err != nil {
		panic(fmt.Sprint("Correct error:", err))
	}
}

func (wsComm *WebsocketCommunication) DisplayWrong() {
	msg := model.Message{
		Wrong: &model.Wrong{},
	}
	err := wsComm.conn.WriteJSON(msg)
	if err != nil {
		panic(fmt.Sprint("Wrong error:", err))
	}
}

func (wsComm *WebsocketCommunication) DisplayProgress(points int) {
	msg := model.Message{
		Progress: &model.Progress{
			Points: points,
		},
	}
	err := wsComm.conn.WriteJSON(msg)
	if err != nil {
		panic(fmt.Sprint("Progress error:", err))
	}
}

func (wsComm *WebsocketCommunication) DisplaySummary(totalPoints int) {
	msg := model.Message{
		Summary: &model.Summary{
			TotalPoints: totalPoints,
		},
	}
	err := wsComm.conn.WriteJSON(msg)
	if err != nil {
		panic(fmt.Sprint("Summary error:", err))
	}
}
