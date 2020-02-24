package scraper

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/ksanta/wordofthedaygame/model"
)

const wotdKey = "wotdKey"
const wordTypeKey = "wordTypeKey"
const definitionKey = "definitionKey"

func Scrape(outputChan chan model.PageDetails) {
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.merriam-webster.com"),
		colly.MaxDepth(10),
		colly.Async(true),
	)

	c.OnRequest(func(request *colly.Request) {
		fmt.Println(request.URL.Path)
	})

	// Scrape the word of the day
	c.OnHTML("h1", func(element *colly.HTMLElement) {
		element.Request.Ctx.Put(wotdKey, element.Text)
	})

	// Scrape the word type (noun, verb, etc)
	c.OnHTML("span.main-attr", func(element *colly.HTMLElement) {
		element.Request.Ctx.Put(wordTypeKey, element.Text)
	})

	// Scrape the word definition
	c.OnHTML("div.wod-definition-container > p", func(element *colly.HTMLElement) {
		element.Request.Ctx.Put(definitionKey, element.Text)
	})

	// Find the link to the previous word of the day
	c.OnHTML("a.prev-wod-arrow", func(element *colly.HTMLElement) {
		link := element.Attr("href")
		c.Visit(element.Request.AbsoluteURL(link))
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

	// Start scraping
	startPageLink := findStartPage()
	c.Visit(startPageLink)
	c.Wait()

	close(outputChan)
}

// Finds the "complete" URL for the most recent word of the day
func findStartPage() string {
	prevLink := findPreviousLink()
	return findStartLink(prevLink)
}

func findPreviousLink() string {
	var prevLink string

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.merriam-webster.com"),
	)

	// Find the link to the previous word of the day
	c.OnHTML("a.prev-wod-arrow", func(element *colly.HTMLElement) {
		prevLink = element.Attr("href")
		prevLink = element.Request.AbsoluteURL(prevLink)
	})

	c.Visit("https://www.merriam-webster.com/word-of-the-day")

	return prevLink
}

func findStartLink(prevLink string) string {
	var startLink string

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.merriam-webster.com"),
	)

	// Find the link to the previous word of the day
	c.OnHTML("a.next-wod-arrow", func(element *colly.HTMLElement) {
		startLink = element.Attr("href")
		startLink = element.Request.AbsoluteURL(startLink)
	})

	c.Visit(prevLink)

	return startLink
}
