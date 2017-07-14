package main

import (
	"github.com/mmcdole/gofeed"
	"time"
	"github.com/JesusIslam/tldr"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"log"
	"github.com/boltdb/bolt"
	"fmt"
	"encoding/json"
)

type Aggregator interface{
	AddStories(s []Story)
}

type RSSFetchers struct {
	Sources []*RSSFetcher
	Aggregator Aggregator
	DB *bolt.DB
}

func NewRSSFetchers(db *bolt.DB, aggregator Aggregator) RSSFetchers {
	f := RSSFetchers{Aggregator: aggregator, DB: db}

	db.Update(func(tx *bolt.Tx) error {
		_ , err := tx.CreateBucket([]byte("sources"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("sources"))
		rawSources := b.Get([]byte("sources"))

		if rawSources != nil {
			var sourcesFromDisk []RSSFetcher

			err := json.Unmarshal( rawSources, &sourcesFromDisk )
			if err != nil {
				log.Println("problem parsing sources from bolt", err)
				return nil
			}

			log.Println("Loaded", len(sourcesFromDisk), "from db", sourcesFromDisk[0].Name)

			for _, source := range sourcesFromDisk {
				f.Add(source.URL, source.Name, source.BodySelector)
			}
		}
		return nil
	})

	if err != nil {
		log.Println("problem loading stories from bolt", err)
	}

	return f
}

func (f *RSSFetchers) GetNames() (names []string) {
	for _, fetcher := range f.Sources {
		names = append(names, fetcher.Name);
	}
	log.Println("Sources %p", f)
	return
}

func (f *RSSFetchers) Add(url string, name string, selector string) {
	log.Println("Adding to Sources %p", f)

	log.Println("Adding new source", url, name, selector)
	newFetcher := NewRSSFetcher(url, name, selector)

	f.Sources = append(f.Sources, newFetcher)
	f.SaveSources()
	go newFetcher.GiveNewStoriesTo(f.Aggregator)
}

func (f *RSSFetchers) SaveSources() {
	err := f.DB.Update(func(tx *bolt.Tx) error {
		sourcesAsJSON, _ := json.Marshal(f.Sources)
		b := tx.Bucket([]byte("sources"))
		err := b.Put([]byte("sources"), sourcesAsJSON)
		return err
	})

	if err != nil {
		log.Println("Problem persisting sources to bolt")
	}
}

type RSSFetcher struct{
	URL string
	Name string
	Titles map[string]int
	BodySelector string
}

func NewRSSFetcher(url string, name string, selector string) *RSSFetcher {
	titles := make(map[string]int)
	return &RSSFetcher{
		URL: url,
		Name: name,
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
			Processed: false,
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

func (r *RSSFetcher) GiveNewStoriesTo(aggregator Aggregator){
	tkr := time.NewTicker(5 * time.Second)

	for _ = range tkr.C{
		aggregator.AddStories(r.GetStories())
	}
}