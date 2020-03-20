package model

import "math/rand"

// Words is simply a slice of Word, with handy methods
type Words []Word

// GroupByType groups this word slice into a map keyed by the type
func (words Words) GroupByType() map[string]Words {
	wordsByType := make(map[string]Words)

	for _, word := range words {
		wordsByType[word.WordType] = append(wordsByType[word.WordType], word)
	}

	return wordsByType
}

// PickRandomType returns one of four random word types
func (words Words) PickRandomType() string {
	wordTypes := []string{"noun", "adjective", "verb", "adverb"}
	randomIndex := rand.Intn(len(wordTypes))
	return wordTypes[randomIndex]
}

// PickRandomWords will pick n unique random words from this word slice. If it
// happens to pick the same word twice, it will re-pick until a unique word is picked.
func (words Words) PickRandomWords(numberToChoose int) Words {
	// Limit the odd case if there just isn't enough words to choose from
	if numberToChoose > len(words) {
		numberToChoose = len(words)
	}

	randomWords := make(Words, 0, numberToChoose)
	alreadyPickedWords := make(map[string]interface{})

	for len(randomWords) < numberToChoose {
		word := words.PickRandomWord()
		if _, alreadyPicked := alreadyPickedWords[word.Word]; !alreadyPicked {
			randomWords = append(randomWords, word)
			alreadyPickedWords[word.Word] = struct{}{}
		}
	}

	return randomWords
}

func (words Words) PickRandomWord() Word {
	return words[rand.Intn(len(words))]
}
