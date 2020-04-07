package model

// MessageToPlayer is sent across the network to the client
type MessageToPlayer struct {
	PlayerDetailsReq *PlayerDetailsReq `json:",omitempty"`
	Intro            *Intro            `json:",omitempty"`
	PresentQuestion  *PresentQuestion  `json:",omitempty"`
	PlayerResult     *PlayerResult     `json:",omitempty"`
	RoundSummary     *RoundSummary     `json:",omitempty"`
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
	TargetScore int
}

// PresentQuestion is sent to the client telling it to pose a question to the player
type PresentQuestion struct {
	WordToGuess    string
	Definitions    []string
	SecondsAllowed int
}

// PlayerResponse is the response from the player
type PlayerResponse struct {
	Response string
}

// PlayerResult is sent to the player telling them their result of the round
type PlayerResult struct {
	Correct bool
	Points  int
}

type RoundSummary struct {
	PlayerStates []PlayerState
}

type PlayerState struct {
	Name   string
	Score  int
	Active bool
}

// Summary is sent to the client at the end telling the player the final result
type Summary struct {
	TotalPoints int
}
