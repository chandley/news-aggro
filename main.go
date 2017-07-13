package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"log"
	"net/http"
	"time"

	"text/template"
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

const storyTemplate = `<li>
	<h2>{{.Title}}</h2> <h3>{{.Source}}</h3>{{.Description}} {{.Date}} <button name="title" type="submit" value="{{.Title}}">Processed</button>
	</li>`

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
			tmpl, err := template.New("test").Parse(storyTemplate)
			if err != nil { panic(err) }
			err = tmpl.Execute(w, story)
			if err != nil { panic(err) }
		}

	}
	fmt.Fprintf(w, "</ul>")
	fmt.Fprintf(w, "</form></html>")
}
