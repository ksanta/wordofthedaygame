package model

import "fmt"

type PageDetails struct {
	Wotd       string
	WordType   string
	Definition string
	URL        string
}

func NewFromStringSlice(stringSlice []string) PageDetails {
	return PageDetails{
		Wotd:       stringSlice[0],
		WordType:   stringSlice[1],
		Definition: stringSlice[2],
		URL:        stringSlice[3],
	}
}

func (d PageDetails) String() string {
	return fmt.Sprintf("%s (%s): %s", d.Wotd, d.WordType, d.Definition)
}

func (d PageDetails) ToStringSlice() []string {
	return []string{d.Wotd, d.WordType, d.Definition, d.URL}
}
