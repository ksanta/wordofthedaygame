package model

import "math/rand"

// Words is simply a slice of Word, with handy methods
type Words []Word

// PickRandomType returns one of four random word types
func PickRandomType() string {
	wordTypes := []string{"noun", "adjective", "verb", "adverb"}
	randomIndex := rand.Intn(len(wordTypes))
	return wordTypes[randomIndex]
}

// GroupByType groups this word slice into a map keyed by the type
func (words Words) GroupByType() map[string]Words {
	wordsByType := make(map[string]Words)

	for _, word := range words {
		wordsByType[word.WordType] = append(wordsByType[word.WordType], word)
	}

	return wordsByType
}

// PickRandomWords will pick n unique random words from this word slice. If it
// happens to pick the same word twice, it will re-pick until a unique word is picked.
func (words Words) PickRandomWords(numberToChoose int) Words {
	// Limit the odd case if there just isn't enough words to choose from
	if numberToChoose >= len(words) {
		return words
	}

	chosenWords := make(Words, 0, numberToChoose)
	pickedIndexes := make(map[int]interface{})

	for len(chosenWords) < numberToChoose {
		index := words.PickRandomIndex()
		if _, present := pickedIndexes[index]; !present {
			chosenWords = append(chosenWords, words[index])
			pickedIndexes[index] = struct{}{}
		}
	}

	return chosenWords
}

func (words Words) PickRandomIndex() int {
	return rand.Intn(len(words))
}

func (words Words) GetDefinitions() []string {
	definitions := make([]string, len(words))
	for i, word := range words {
		definitions[i] = word.Definition
	}
	return definitions
}
