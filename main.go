package main

import (
	"log"

	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		log.Println(e.Text)
		s, er := e.DOM.Html()
		log.Println("lalal:", s, er)
	})

	c.OnResponse(func(resp *colly.Response) {
		log.Println("lalal2:", string(resp.Body))
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Referer", "http://scuinfo.com/")
	})

	c.Visit("http://scuinfo.com/api/posts?pageSize=15")
	c.Wait()
}
