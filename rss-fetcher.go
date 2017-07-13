package main

import (
	"github.com/mmcdole/gofeed"
	"time"
	"github.com/JesusIslam/tldr"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

type Aggregator interface{
	AddStories(s []Story)
}

type RSSFetcher struct{
	URL string
	Name string
	Aggregator Aggregator
	Titles map[string]int
	BodySelector string
}

func NewRSSFetcher(url string, name string, selector string, aggregator Aggregator) *RSSFetcher {
	titles := make(map[string]int)
	return &RSSFetcher{
		URL: url,
		Name: name,
		Aggregator:aggregator,
		Titles:titles,
		BodySelector:selector,
	}
}

const numberOfSentences = 3

func (r *RSSFetcher) GetStories() (stories []Story) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(r.URL)

	for _, article := range feed.Items {

		if _, exists := r.Titles[article.Title]; exists {
			continue
		}

		r.Titles[article.Title] = 0

		datePublished := article.PublishedParsed
		if datePublished==nil{
			now := time.Now()
			datePublished = &now
		}

		stories = append(stories, Story{
			Title: article.Title,
			Description: article.Description,
			Link: article.Link,
			Source:r.Name,
			Date: datePublished,
			Summary:createSummary(article.Link, r.BodySelector),
		})
	}

	return
}

func createSummary(url, selector string) string {

	bag := tldr.New()
	bag.Set(300, tldr.DEFAULT_DAMPING, tldr.DEFAULT_TOLERANCE, tldr.DEFAULT_THRESHOLD, tldr.DEFAULT_SENTENCES_DISTANCE_THRESHOLD, tldr.DEFAULT_ALGORITHM, tldr.DEFAULT_WEIGHING)


	res, _ := http.Get(url)
	defer res.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(res.Body)

	doc.Find("script").Each(func(i int, el *goquery.Selection) {
		el.Remove()
	})

	summary, _ := bag.Summarize(doc.Find(selector).Text(), numberOfSentences)
	return summary

}

func (r *RSSFetcher) ListenForUpdates(){
	tkr := time.NewTicker(5 * time.Second)

	for _ = range tkr.C{
		r.Aggregator.AddStories(r.GetStories())
	}
}