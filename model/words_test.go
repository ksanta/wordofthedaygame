package model

import (
	"math/rand"
	"testing"
)

var sampleWords = Words{
	Word{Word: "one", WordType: "noun"},
	Word{Word: "two", WordType: "noun"},
	Word{Word: "three", WordType: "noun"},
	Word{Word: "four", WordType: "noun"},
}

func TestWords_PickRandomType(t *testing.T) {
	rand.Seed(1)
	got := PickRandomType()
	expected := "adjective"
	if got != expected {
		t.Errorf("Got %s and expected %s", got, expected)
	}
}

func TestWords_PickRandomWords_NoWords(t *testing.T) {
	rand.Seed(1)

	got := sampleWords.PickRandomWords(0)
	expected := Words{}
	if len(got) != len(expected) {
		t.Errorf("Got length %d and expected %d", len(got), len(expected))
	}
}

func TestWords_PickRandomWords_OneWord(t *testing.T) {
	rand.Seed(1)

	got := sampleWords.PickRandomWords(1)
	expectedWord := sampleWords[1]
	expected := Words{expectedWord}
	if len(got) != len(expected) {
		t.Errorf("Got length %d and expected %d", len(got), len(expected))
	}
	if got[0] != expectedWord {
		t.Errorf("Got word %s and expected %s", got[0], expectedWord)
	}
}

func TestWords_PickRandomWords_PickTooMany(t *testing.T) {
	rand.Seed(1)

	got := sampleWords.PickRandomWords(5) // Only four words in sample
	expectedLength := len(sampleWords)
	if len(got) != expectedLength {
		t.Errorf("Got length %d and expected %d", len(got), expectedLength)
	}
}
