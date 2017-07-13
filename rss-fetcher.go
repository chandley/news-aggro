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
	Titles map[string]int
}

func NewRSSFetcher(url string, name string, aggregator Aggregator) *RSSFetcher {
	titles := make(map[string]int)
	return &RSSFetcher{
		URL: url,
		Name: name,
		Aggregator:aggregator,
		Titles:titles,
	}
}

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