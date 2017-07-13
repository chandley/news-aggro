package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"log"
	"net/http"
	"time"
)

func main() {
	feedServer := NewFeed()

	sources := map[string]string {
		"bbc news": "http://feeds.bbci.co.uk/news/rss.xml?edition=uk",
		"LSE FTSE whatever": "http://www.londonstockexchange.com/exchange/CompanyNewsRSS.html?newsSource=RNS&indexSymbol=UKX",
	}

	for name, url := range sources {
		feedServer.AddStories(StoriesFromFeed(url, name))
	}

	router := http.NewServeMux()
	router.Handle("/", feedServer)

	fmt.Println("Listening on 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

func StoriesFromFeed(url string, name string) (stories []Story) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	for _, article := range feed.Items {
		stories = append(stories, Story{
			Title: article.Title,
			Description: article.Description,
			Source:name,
			Date: article.PublishedParsed,
		})
	}

	return

}

type Story struct {
	Title, Description, Source string
	Date *time.Time
	Processed bool
}

func NewFeed() *Feed {
	return new(Feed)
}

type Feed struct {
	Stories []Story
}


func (f *Feed) AddStories(s []Story) {
	f.Stories = append(f.Stories, s...)
}

func (f *Feed) MarkAsProcessed(title string) {
	for i, story := range f.Stories {
		if story.Title == title {
			f.Stories[i].Processed = true
		}
	}
}

func (f *Feed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		title := r.FormValue("title")
		f.MarkAsProcessed(title)
	}

	fmt.Fprintf(w, `<html><form action="/" method="post">`)
	fmt.Fprintf(w, "<ul>")

	for _, story := range f.Stories {
		if story.Processed {
			fmt.Fprintf(w, "<h1>Move along please</h1>")
		} else {
			fmt.Fprintf(w, `<li><h2>%s</h2><h3>%s</h3>%s %s	<button name="title" type="submit" value="%s">Processed</button></li>`, story.Title, story.Source, story.Description, story.Date, story.Title)
		}

	}
	fmt.Fprintf(w, "</ul>")
	fmt.Fprintf(w, "</form></html>")
}
