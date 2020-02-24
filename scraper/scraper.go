// Scrapes websites for words of the day
package scraper

import "github.com/ksanta/wordofthedaygame/model"

type Scraper interface {
	// Scrape all the word-of-the-days from a site
	Scrape(outputChan chan model.PageDetails)

	// Scrape only newer words from a site, stopping with the given stopWord
	ScrapeNewWords(outputChan chan model.PageDetails, stopWord string)
}
