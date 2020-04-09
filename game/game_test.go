package game

import (
	"github.com/ksanta/wordofthedaygame/model"
	"testing"
	"time"
)

var words = model.Words{
	{Word: "hello"},
	{Word: "greetings"},
	{Word: "hej"},
}

func TestGame_CalculatePoints(t *testing.T) {
	g := Game{
		WordsByType:         nil,
		TargetScore:         500,
		OptionsPerQuestion:  3,
		DurationPerQuestion: 10 * time.Second,
	}

	gotPoints := g.calculatePoints(true, 2*time.Second)
	expectedPoints := 100 + 40
	if gotPoints != expectedPoints {
		t.Errorf("Got %d points but expected %d", gotPoints, expectedPoints)
	}

	gotPoints = g.calculatePoints(false, 8*time.Second)
	expectedPoints = 0 + 10
	if gotPoints != expectedPoints {
		t.Errorf("Got %d points but expected %d", gotPoints, expectedPoints)
	}
}

func TestGame_ValidateResponse(t *testing.T) {
	g := Game{
		WordsByType:         nil,
		TargetScore:         500,
		OptionsPerQuestion:  3,
		DurationPerQuestion: 10 * time.Second,
		wordsInRound:        words,
	}

	// Validate responses
	g.wordToGuess = "greetings"
	if g.validateResponse("1") {
		t.Error("Happy case 1 fail")
	}

	if !g.validateResponse("2") {
		t.Error("Happy case 2 fail")
	}

	if !g.validateResponse("02") {
		t.Error("Happy case 02 fail")
	}

	if !g.validateResponse(" 2") {
		t.Error("Happy case ' 2' fail")
	}

	if !g.validateResponse("2 ") {
		t.Error("Happy case '2 ' fail")
	}

	// Invalid responses
	g.wordToGuess = "hello"
	if g.validateResponse("0") {
		t.Error("Zero response fail")
	}

	if g.validateResponse("-1") {
		t.Error("Negative response fail")
	}

	if g.validateResponse("") {
		t.Error("Blank response fail")
	}

	if g.validateResponse("A") {
		t.Error("Alpha response fail")
	}
}
