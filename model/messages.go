package model

// MessageToPlayer is sent across the network to the client
type MessageToPlayer struct {
	PlayerDetailsReq *PlayerDetailsReq `json:",omitempty"`
	Intro            *Intro            `json:",omitempty"`
	PresentQuestion  *PresentQuestion  `json:",omitempty"`
	Timeout          *Timeout          `json:",omitempty"`
	Correct          *Correct          `json:",omitempty"`
	Wrong            *Wrong            `json:",omitempty"`
	Progress         *Progress         `json:",omitempty"`
	Summary          *Summary          `json:",omitempty"`
}

// MessageFromPlayer is received from the network from the client
type MessageFromPlayer struct {
	PlayerDetailsResp *PlayerDetails  `json:",omitempty"`
	PlayerResponse    *PlayerResponse `json:",omitempty"`
}

// PlayerDetailsReq is sent to the client telling it to get the player's details
type PlayerDetailsReq struct{}

// PlayerDetails is sent to the server with player details when they start the game
type PlayerDetails struct {
	Name string
}

// Intro tells the client to display an intro to the player
type Intro struct {
	QuestionsPerGame int
}

// PresentQuestion is sent to the client telling it to pose a question to the player
type PresentQuestion struct {
	Round       int
	WordToGuess string
	Definitions []string
}

// PlayerResponse is the response from the player
type PlayerResponse struct {
	Response string
}

// Timeout is sent to the client telling it the player took too long to answer the question
type Timeout struct{}

// Correct is sent to the client telling it the player guessed correctly
type Correct struct{}

// Correct is sent to the client telling it the player guessed wrong
type Wrong struct{}

// Progress is sent to the client telling it the incremental progress between rounds
type Progress struct {
	Points int
}

// Summary is sent to the client at the end telling the player the final result
type Summary struct {
	TotalPoints int
}
