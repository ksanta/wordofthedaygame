package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/ksanta/wordofthedaygame/model"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var timeoutChan = make(chan struct{})

func main() {
	flag.Parse()

	// Direct the interrupt (ctrl-c) signal into a channel for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn := connectToServer()
	defer conn.Close()

	// Read loop in goroutine
	done := startReadLoop(conn)

	// Write loop in current thread
	startWriteLoop(conn, done, interrupt)
}

func connectToServer() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/game"}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	return conn
}

func startReadLoop(conn *websocket.Conn) chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var msg model.Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("read error:", err)
				return
			}

			// Delegate to handlers depending on message contents
			if msg.PlayerDetailsReq != nil {
				handlePlayerDetailsReqMessage(conn)

			} else if msg.Intro != nil {
				handleIntroMessage(msg.Intro)

			} else if msg.PresentQuestion != nil {
				handlePresentQuestionMessage(conn, msg.PresentQuestion)

			} else if msg.Timeout != nil {
				handleTimeoutMessage()

			} else if msg.Correct != nil {
				handleCorrect()

			} else if msg.Wrong != nil {
				handleWrong()

			} else if msg.Progress != nil {
				handleProgress(msg.Progress)

			} else if msg.Summary != nil {
				handleSummary(msg.Summary)

			} else {
				log.Fatal("Unsupported message", msg)
			}
		}
	}()
	return done
}

func handleCorrect() {
	fmt.Println("Correct 🎉")
}

func handleWrong() {
	fmt.Println("Wrong! 💀💀💀")
}

func handleProgress(progress *model.Progress) {
	fmt.Printf("Earned %d points\n", progress.Points)
}

func handleSummary(summary *model.Summary) {
	fmt.Println()
	fmt.Println("You scored", summary.TotalPoints, "points!")
}

func handleTimeoutMessage() {
	timeoutChan <- struct{}{}
}

func handlePresentQuestionMessage(conn *websocket.Conn, q *model.PresentQuestion) {
	fmt.Println()
	fmt.Printf("Round %d!\n", q.Round)
	fmt.Println("The word of the day is:", strings.ToUpper(q.WordToGuess))

	for i, definition := range q.Definitions {
		fmt.Printf("%d) %s\n", i+1, definition)
	}
	fmt.Print("\nEnter your best guess: ")

	response := getAnswerFromPlayer()

	err := conn.WriteJSON(model.Message{
		PlayerResponse: &model.PlayerResponse{
			Response: response,
		},
	})
	if err != nil {
		log.Fatal("Send PlayerResponse err", err)
	}
}

func handlePlayerDetailsReqMessage(conn *websocket.Conn) {
	fmt.Print("Enter your name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	playerDetailsResp := model.PlayerDetailsResp{Name: scanner.Text()}

	err := conn.WriteJSON(model.Message{
		PlayerDetailsResp: &playerDetailsResp,
	})
	if err != nil {
		log.Fatal("Send PlayerDetailsResponse err", err)
	}
}

func handleIntroMessage(intro *model.Intro) {
	fmt.Println("Playing", intro.QuestionsPerGame, "rounds")
}

func startWriteLoop(conn *websocket.Conn, done chan struct{}, interrupt chan os.Signal) {
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close error:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func getAnswerFromPlayer() string {
	stdinChannel := make(chan string, 1)

	// Get the answer from the player in a different goroutine and send to the channel
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		stdinChannel <- scanner.Text()
	}()

	select {
	case response := <-stdinChannel:
		return response
	case <-timeoutChan:
		fmt.Println("💥 Too slow! 💥")
		// On timeout, the goroutine is still blocked waiting for user input.
		// Prompt the player to hit enter to finish the goroutine
		fmt.Print("Hit enter to move to the next question")
		<-stdinChannel
		return ""
	}
}
