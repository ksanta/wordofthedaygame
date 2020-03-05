// Scrapes websites for words of the day
package scraper

import "github.com/ksanta/wordofthedaygame/model"

type Scraper interface {
	// Scrape will scrape a website for word definitions (up to "limit" words) and send them to a channel for consumption
	Scrape(limit int) chan model.PageDetails
}
