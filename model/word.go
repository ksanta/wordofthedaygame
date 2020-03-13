package model

import "fmt"

type Word struct {
	Word       string
	WordType   string
	Definition string
	URL        string
}

func NewFromStringSlice(stringSlice []string) Word {
	return Word{
		Word:       stringSlice[0],
		WordType:   stringSlice[1],
		Definition: stringSlice[2],
		URL:        stringSlice[3],
	}
}

func (d Word) String() string {
	return fmt.Sprintf("%s (%s): %s", d.Word, d.WordType, d.Definition)
}

func (d Word) ToStringSlice() []string {
	return []string{d.Word, d.WordType, d.Definition, d.URL}
}
