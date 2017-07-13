package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"text/template"
	"io"
)

func main() {
	feedServer := NewFeed()

	sources := []*RSSFetcher {
		NewRSSFetcher("http://feeds.bbci.co.uk/news/rss.xml?edition=uk", "bbc news", feedServer),
		NewRSSFetcher("http://www.londonstockexchange.com/exchange/CompanyNewsRSS.html?newsSource=RNS&indexSymbol=UKX", "LSE FTSE whatever", feedServer),
		NewRSSFetcher("http://www.topazworld.com/en/rss/news", "Topaz", feedServer),
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
}


func (f *Feed) AddStories(s []Story) {
	f.Stories = append(f.Stories, s...)
}

func (f *Feed) UnprocessedStories () (stories []Story){
	for _, story := range f.Stories {
		if !story.Processed {
			stories = append(stories, story)
		}
	}
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
