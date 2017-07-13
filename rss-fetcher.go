package main

import (
	"github.com/mmcdole/gofeed"
	"time"
	"log"
)

type Aggregator interface{
	AddStories(s []Story)
}

type RSSFetcher struct{
	URL string
	Name string
	Aggregator Aggregator
}

func NewRSSFetcher(url string, name string, aggregator Aggregator) *RSSFetcher {
	return &RSSFetcher{
		URL: url,
		Name: name,
		Aggregator:aggregator,
	}
}

func (r *RSSFetcher) GetStories() (stories []Story) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(r.URL)

	for _, article := range feed.Items {

		datePublished := article.PublishedParsed
		if datePublished==nil{
			datePublished = &time.Now()
		}

		stories = append(stories, Story{
			Title: article.Title,
			Description: article.Description,
			Link: article.Link,
			Source:r.Name,
			Date: datePublished,
		})
	}

	return
}

func (r *RSSFetcher) ListenForUpdates(){
	tkr := time.NewTicker(5 * time.Second)

	for _ = range tkr.C{
		log.Printf("Checking %s for new content\n", r.Name)
		r.Aggregator.AddStories(r.GetStories())
		log.Printf("Fetched content for %s\n", r.Name)
	}
}