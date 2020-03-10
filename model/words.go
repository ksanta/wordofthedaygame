package model

import "math/rand"

// Words is simply a slice of WordDetail, with handy methods
type Words []WordDetail

// GroupByType groups this word slice into a map keyed by the type
func (words Words) GroupByType() map[string][]WordDetail {
	wordsByType := make(map[string][]WordDetail)

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
func (words Words) PickRandomWords(wordsToChoose int) []WordDetail {
	randomWords := make([]WordDetail, 0, wordsToChoose)
	alreadyPickedWords := make(map[string]interface{})

	// todo: prevent infinite loops if there aren't enough words in the slice to pick from
	for len(randomWords) < wordsToChoose {
		wordDetail := words.PickRandomWord()
		if _, alreadyPicked := alreadyPickedWords[wordDetail.Wotd]; !alreadyPicked {
			randomWords = append(randomWords, wordDetail)
			alreadyPickedWords[wordDetail.Wotd] = struct{}{}
		}
	}

	return randomWords
}

func (words Words) PickRandomWord() WordDetail {
	return words[rand.Intn(len(words))]
}
