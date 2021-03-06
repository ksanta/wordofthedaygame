package player

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ksanta/wordofthedaygame/model"
	"log"
	"os"
	"time"
)

// Player sits between the Game and the Websocket connection. It monitors the
// Websocket connection for incoming messages and forwards them to the game via a
// channel. It monitors a channel from the game and sends messages out the Websocket
// connection.
type Player struct {
	// Embedded Logger allows the Player struct to attach Logger methods
	*log.Logger
	// The Websocket connection
	conn *websocket.Conn
	// Whether the player has an active connection. Connection could go dead mid-game
	// and the show must go on!
	Active bool
	// WaitingFOrResponse means the player has been sent the question and hasn't received response yet
	WaitingForResponse bool
	// Posting here will terminate the TCP connection
	disconnectChan chan struct{}
	// Posting here will send the message to the game hub
	sendToGameChan chan PlayerMessage
	// SendToClientChan will send received messages to the Websocket connection
	SendToClientChan chan model.MessageToPlayer
	// Name of this player
	name string
	// Client-specific icon to represent the player
	Icon string
	// Points for this player
	points int
	// Time tracks when a player started to answer a question
	startTime time.Time
}

// PlayerMessage is sent from the Player to the Game, so the player knows which
// player sent which message to the game
type PlayerMessage struct {
	Player  *Player
	Message model.MessageFromPlayer
}

func NewPlayer(conn *websocket.Conn, disconnectChan chan struct{}, sendToGameChan chan PlayerMessage) *Player {
	return &Player{
		Logger:           log.New(os.Stdout, "[New player] ", 0),
		conn:             conn,
		disconnectChan:   disconnectChan,
		sendToGameChan:   sendToGameChan,
		SendToClientChan: make(chan model.MessageToPlayer),
		name:             "New player",
	}
}

// WritePump listens on channels and writes messages to the Websocket connection.
// This is to be started as a goroutine.
func (p *Player) WritePump() {
	defer func() {
		p.Println("Closing TCP connection")
		p.disconnectChan <- struct{}{}
	}()

	for {
		message, ok := <-p.SendToClientChan
		if !ok {
			// The hub closed the channel.
			p.Println("Sending close message")
			p.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		err := p.sendJSON(message)
		if err != nil {
			p.Println("SendJSON error:", err)
			return
		}
	}
}

// ReadPump listens to the Websocket connection and delegates handling of messages.
// This is to be started as a goroutine.
func (p *Player) ReadPump() {
	defer func() {
		// Notify the game that the player disconnected. That will handle shutdown.
		p.Println("Unregistering", p.name)
		p.sendToGameChan <- PlayerMessage{
			Player: p,
			Message: model.MessageFromPlayer{
				Disconnected: &model.Disconnected{},
			},
		}
	}()

	for {
		message, err := p.receiveJSON()
		if err != nil {
			// Client closed the connection
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				p.Println("Unexpected close error:", err)
			}
			return
		}
		p.sendToGameChan <- PlayerMessage{
			Player:  p,
			Message: message,
		}
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

func (p *Player) GetName() string {
	return p.name
}

func (p *Player) StartTimer() {
	p.startTime = time.Now()
}

func (p *Player) StopTimer() time.Duration {
	return time.Since(p.startTime)
}

func (p *Player) sendJSON(request model.MessageToPlayer) error {
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	p.Print("<- ", string(jsonBytes))

	err = p.conn.WriteMessage(websocket.TextMessage, jsonBytes)
	if err != nil {
		return err
	}

	return nil
}

func (p *Player) receiveJSON() (model.MessageFromPlayer, error) {
	_, jsonBytes, err := p.conn.ReadMessage()
	if err != nil {
		return model.MessageFromPlayer{}, err
	}
	p.Print("-> ", string(jsonBytes))

	var response model.MessageFromPlayer
	err = json.Unmarshal(jsonBytes, &response)
	if err != nil {
		return model.MessageFromPlayer{}, err
	}
	return response, nil
}

func (p *Player) PlayerState() model.PlayerState {
	return model.PlayerState{
		Name:   p.name,
		Score:  p.GetPoints(),
		Active: p.Active,
		Icon:   p.Icon,
	}
}
