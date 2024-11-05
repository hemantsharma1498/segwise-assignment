package scraper

import "github.com/gocolly/colly"

type Scraper struct {
	colly *colly.Collector
}

func NewScraper() *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains("www.linkedin.com"),
	)
	return c
}
