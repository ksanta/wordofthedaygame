package scraper

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/ksanta/wordofthedaygame/model"
	"strings"
	"time"
)

const wotdKey = "wotdKey"
const wordTypeKey = "wordTypeKey"
const definitionKey = "definitionKey"

type MeriamScraper struct {
}

func (m *MeriamScraper) Scrape(limit int) chan model.PageDetails {
	outputChan := make(chan model.PageDetails)

	go func() {

		// Instantiate default collector
		c := colly.NewCollector()

		// Scrape the word of the day
		c.OnHTML("h1", func(element *colly.HTMLElement) {
			element.Request.Ctx.Put(wotdKey, element.Text)
		})

		// Scrape the word type (noun, verb, etc)
		c.OnHTML("span.main-attr", func(element *colly.HTMLElement) {
			element.Request.Ctx.Put(wordTypeKey, element.Text)
		})

		// Scrape the word definition
		c.OnHTML("div.wod-definition-container", func(element *colly.HTMLElement) {
			element.ForEachWithBreak("p", func(i int, element *colly.HTMLElement) bool {
				cleanedDefinition := cleanUpDefinition(element.Text)
				element.Request.Ctx.Put(definitionKey, cleanedDefinition)
				// Returning false will break from the loop (ie process only first matching element)
				return false
			})
		})

		c.OnScraped(func(response *colly.Response) {
			wordEntry := model.PageDetails{
				Wotd:       response.Ctx.Get(wotdKey),
				WordType:   response.Ctx.Get(wordTypeKey),
				Definition: response.Ctx.Get(definitionKey),
				URL:        response.Request.URL.String(),
			}
			outputChan <- wordEntry
		})

		q, _ := queue.New(
			20,
			&queue.InMemoryQueueStorage{MaxSize: 10000},
		)

		// Generate URLs based on dates and visit them all
		yesterday := time.Now().AddDate(0, 0, -1)
		for i := 0; i < limit; i++ {
			date := yesterday.AddDate(0, 0, -i)
			formattedDate := date.Format("2006-01-02")
			url := "https://www.merriam-webster.com/word-of-the-day/" + formattedDate
			q.AddURL(url)
		}

		q.Run(c)

		close(outputChan)
	}()

	return outputChan
}

func cleanUpDefinition(rawText string) string {
	firstColonIndex := strings.Index(rawText, ":")
	cleanedDefinition := strings.TrimSpace(rawText[firstColonIndex+1:])
	return cleanedDefinition
}
