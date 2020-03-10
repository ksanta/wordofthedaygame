// Scrapes websites for words of the day
package scraper

import "github.com/ksanta/wordofthedaygame/model"

type Scraper interface {
	// Scrape will scrape a website for word definitions and send them to a channel for consumption
	Scrape() chan model.Word
}
