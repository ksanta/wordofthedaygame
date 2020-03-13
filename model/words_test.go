package model

import (
	"math/rand"
	"testing"
)

var sampleWords = Words{
	Word{Word: "one"},
	Word{Word: "two"},
	Word{Word: "three"},
	Word{Word: "four"},
}

func TestWords_PickRandomType(t *testing.T) {
	rand.Seed(1)
	got := sampleWords.PickRandomType()
	expected := "adjective"
	if got != expected {
		t.Errorf("Got %s and expected %s", got, expected)
	}
}

func TestWords_PickRandomWord(t *testing.T) {
	rand.Seed(1)
	got := sampleWords.PickRandomWord().Word
	expected := "two"
	if got != expected {
		t.Errorf("Got %s and expected %s", got, expected)
	}
}
