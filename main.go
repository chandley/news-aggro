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
		NewRSSFetcher("http://feeds.bbci.co.uk/news/rss.xml?edition=uk", "bbc news", feedServer),
		NewRSSFetcher("http://www.londonstockexchange.com/exchange/CompanyNewsRSS.html?newsSource=RNS&indexSymbol=UKX", "LSE FTSE whatever", feedServer),
		NewRSSFetcher("http://www.topazworld.com/en/rss/news", "Topaz", feedServer),
		//NewRSSFetcher("https://twitrss.me/twitter_search_to_rss/?term=bieber", "Bieber News from Twitter", feedServer),
	}

	for _, fetcher := range sources {
		feedServer.AddStories(fetcher.GetStories())
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
	Title, Description, Source, Link string
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

	log.Println("Number of contents is", len(f.Stories))

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
	<form action="/" method="post">
		<ul>
			{{range .}}
			<li>
				<h2>{{.Title}}</h2><h3>{{.Source}}</h3>{{.Description}} {{.Date}}
				<a href="{{.Link}}">story</a>
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
