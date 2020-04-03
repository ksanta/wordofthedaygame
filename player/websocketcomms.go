package player

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/ksanta/wordofthedaygame/model"
	"log"
	"time"
)

// WebsocketCommunication is an implementation that communicates to the player via Websockets
type WebsocketCommunication struct {
	conn *websocket.Conn
	*log.Logger
}

func NewWebsocketCommunication(conn *websocket.Conn, logger *log.Logger) Comms {
	return &WebsocketCommunication{conn: conn, Logger: logger}
}

func (wsComm *WebsocketCommunication) GetPlayerDetails() string {
	request := model.Message{
		PlayerDetailsReq: &model.PlayerDetailsReq{},
	}
	wsComm.sendJSON(request)

	response := wsComm.receiveJSON()

	return response.PlayerDetailsResp.Name
}

func (wsComm *WebsocketCommunication) DisplayIntro(questionsPerGame int) {
	msg := model.Message{
		Intro: &model.Intro{
			QuestionsPerGame: questionsPerGame,
		},
	}
	wsComm.sendJSON(msg)
}

func (wsComm *WebsocketCommunication) PresentQuestion(round int, wordToGuess string, definitions []string, timeoutChan <-chan time.Time) string {
	request := model.Message{
		PresentQuestion: &model.PresentQuestion{
			Round:       round,
			WordToGuess: wordToGuess,
			Definitions: definitions,
		},
	}
	wsComm.sendJSON(request)

	errorChan := make(chan interface{})
	responseChan := make(chan model.Message)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- r
			}
		}()

		responseChan <- wsComm.receiveJSON()
	}()

	select {
	case err := <-errorChan:
		panic(err)
		return ""
	case response := <-responseChan:
		return response.PlayerResponse.Response
	case <-timeoutChan:
		timeoutMsg := model.Message{
			Timeout: &model.Timeout{},
		}
		wsComm.sendJSON(timeoutMsg)
		// Still wait for user to respond after timeout
		<-responseChan
		return ""
	}
}

func (wsComm *WebsocketCommunication) DisplayCorrect() {
	msg := model.Message{
		Correct: &model.Correct{},
	}
	wsComm.sendJSON(msg)
}

func (wsComm *WebsocketCommunication) DisplayWrong() {
	msg := model.Message{
		Wrong: &model.Wrong{},
	}
	wsComm.sendJSON(msg)
}

func (wsComm *WebsocketCommunication) DisplayProgress(points int) {
	msg := model.Message{
		Progress: &model.Progress{
			Points: points,
		},
	}
	wsComm.sendJSON(msg)
}

func (wsComm *WebsocketCommunication) DisplaySummary(totalPoints int) {
	msg := model.Message{
		Summary: &model.Summary{
			TotalPoints: totalPoints,
		},
	}
	wsComm.sendJSON(msg)
}

func (wsComm *WebsocketCommunication) sendJSON(request model.Message) {
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	wsComm.Print("<- ", string(jsonBytes))

	err = wsComm.conn.WriteMessage(websocket.TextMessage, jsonBytes)
	if err != nil {
		panic(err)
	}
}

func (wsComm *WebsocketCommunication) receiveJSON() model.Message {
	_, jsonBytes, err := wsComm.conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	wsComm.Print("-> ", string(jsonBytes))

	var response model.Message
	err = json.Unmarshal(jsonBytes, &response)
	if err != nil {
		panic(err)
	}
	return response
}
