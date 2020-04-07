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
		Words:               nil,
		TargetScore:         5,
		OptionsPerQuestion:  3,
		DurationPerQuestion: 10 * time.Second,
	}

	gotPoints := g.calculatePoints(true, 2*time.Second)
	expectedPoints := 100 + 80
	if gotPoints != expectedPoints {
		t.Errorf("Got %d points but expected %d", gotPoints, expectedPoints)
	}

	gotPoints = g.calculatePoints(false, 8*time.Second)
	expectedPoints = 0 + 20
	if gotPoints != expectedPoints {
		t.Errorf("Got %d points but expected %d", gotPoints, expectedPoints)
	}
}

func TestGame_ValidateResponse(t *testing.T) {
	// Validate responses
	if !validateResponse("1", words, "hello") {
		t.Error("Happy case 1 fail")
	}

	if !validateResponse("2", words, "greetings") {
		t.Error("Happy case 2 fail")
	}

	if !validateResponse("02", words, "greetings") {
		t.Error("Happy case 02 fail")
	}

	if !validateResponse(" 2", words, "greetings") {
		t.Error("Happy case ' 2' fail")
	}

	if !validateResponse("2 ", words, "greetings") {
		t.Error("Happy case '2 ' fail")
	}

	// Invalid responses
	if validateResponse("0", words, "hello") {
		t.Error("Zero response fail")
	}

	if validateResponse("-1", words, "hello") {
		t.Error("Negative response fail")
	}

	if validateResponse("", words, "hello") {
		t.Error("Blank response fail")
	}

	if validateResponse("A", words, "hello") {
		t.Error("Alpha response fail")
	}

}
