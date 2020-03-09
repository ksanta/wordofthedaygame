package model

import "fmt"

type WordDetail struct {
	Wotd       string
	WordType   string
	Definition string
	URL        string
}

func NewFromStringSlice(stringSlice []string) WordDetail {
	return WordDetail{
		Wotd:       stringSlice[0],
		WordType:   stringSlice[1],
		Definition: stringSlice[2],
		URL:        stringSlice[3],
	}
}

func (d WordDetail) String() string {
	return fmt.Sprintf("%s (%s): %s", d.Wotd, d.WordType, d.Definition)
}

func (d WordDetail) ToStringSlice() []string {
	return []string{d.Wotd, d.WordType, d.Definition, d.URL}
}
