package game

import (
	"testing"
	"time"
)

func TestGame_CalculatePoints(t *testing.T) {
	g := Game{
		Words:               nil,
		QuestionsPerGame:    5,
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
