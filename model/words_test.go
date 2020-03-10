package model

import (
	"math/rand"
	"testing"
)

var sampleWords = Words{
	Word{Wotd: "one"},
	Word{Wotd: "two"},
	Word{Wotd: "three"},
	Word{Wotd: "four"},
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
	got := sampleWords.PickRandomWord().Wotd
	expected := "two"
	if got != expected {
		t.Errorf("Got %s and expected %s", got, expected)
	}
}
