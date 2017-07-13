package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"text/template"
	"io"
	"sync"
	"sort"
)

func main() {
	feedServer := NewFeed()

	sources := []*RSSFetcher {
		NewRSSFetcher("http://feeds.bbci.co.uk/news/rss.xml?edition=uk", "bbc news", ".story-body__inner", feedServer),
		NewRSSFetcher("http://www.londonstockexchange.com/exchange/CompanyNewsRSS.html?newsSource=RNS&indexSymbol=UKX", "LSE FTSE whatever", ".bg", feedServer),
		NewRSSFetcher("http://www.topazworld.com/en/rss/news", "Topaz", "#content", feedServer),
		NewRSSFetcher("https://twitrss.me/twitter_search_to_rss/?term=bieber", "Bieber News from Twitter", ".tweet-text", feedServer),
		NewRSSFetcher("https://twitrss.me/twitter_user_to_rss/?user=quii", "Chris James News from Twitter", ".tweet-text",feedServer),
		NewRSSFetcher("https://twitrss.me/twitter_user_to_rss/?user=chrisrhandley", "Chris Handley News from Twitter", ".tweet-text",feedServer),
		NewRSSFetcher("http://lorem-rss.herokuapp.com/feed?unit=second&interval=10", "Lorem ipsum feed", ".none",feedServer),
	}

	for _, fetcher := range sources {
		go feedServer.AddStories(fetcher.GetStories())
		go fetcher.ListenForUpdates()
	}

	router := http.NewServeMux()
	router.Handle("/", feedServer)

	fmt.Println("Listening on 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

type Story struct {
	Title, Description, Source, Link, Summary string
	Date *time.Time
	Processed bool
}

func NewFeed() *Feed {
	f := new(Feed)
	tmpl, err := template.New("test").Parse(storyTemplate)
	if err != nil { panic(err) }
	f.tmpl = tmpl
	return f
}

type Feed struct {
	Stories []Story
	tmpl *template.Template
	sync.Mutex
}


func (f *Feed) AddStories(s []Story) {

	f.Lock()
	defer f.Unlock()
	startingNumberOfStories := len(f.Stories)

	for _, newStory := range s {
		duplicate := false
		for _, existingStory := range f.Stories {
			if newStory.Title == existingStory.Title && newStory.Source == existingStory.Source {
				duplicate = true
			}
		}
		if !duplicate {
			f.Stories = append(f.Stories, newStory)
		} else {
			log.Println("duplicate detected:", newStory.Title, newStory.Source)
		}
	}

	newStoryCount := len(f.Stories) - startingNumberOfStories

	if newStoryCount > 0 {
		log.Println(newStoryCount, "new stories added")
	}

}

func (f *Feed) UnprocessedStories () (stories []Story){
	for _, story := range f.Stories {
		if !story.Processed {
			stories = append(stories, story)
		}
	}
	sort.Slice(stories, func(i, j int) bool { return stories[i].Date.After(*stories[j].Date) })
	return


}

func (f *Feed) MarkAsProcessed(title string) {
	for i, story := range f.Stories {
		if story.Title == title {
			f.Stories[i].Processed = true
		}
	}
}

func (f *Feed) RenderFeedAsHTML(out io.Writer) {
	f.tmpl.Execute(out, f.UnprocessedStories())
}

const storyTemplate = `

<html>
<meta http-equiv="refresh" content="5; URL=http://localhost:8080">
	<form action="/" method="post">
		<ul>
			{{range .}}
			<li>
				<h2>{{.Title}}</h2><h3>{{.Source}}</h3>{{.Description}} {{.Date}}
				<a href="{{.Link}}">story</a>
				<p>{{.Summary}}</p>
				<button name="title" type="submit" value="{{.Title}}">Processed</button>
			</li>
			{{end}}
		</ul>
	</form>
</html>
`

func (f *Feed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		title := r.FormValue("title")
		f.MarkAsProcessed(title)
	}

	f.RenderFeedAsHTML(w)

}
