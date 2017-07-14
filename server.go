package main

import (
	"net/http"
	"html/template"
	"io/ioutil"
	"log"
)

type StoryFeed interface{
	GetStories() []Story
	MarkAsProcessed(title string)
}

type SourcesList interface{
	GetNames() []string
	Add(url string, name string, selector string)
}

type Server struct {
	feed StoryFeed
	sources SourcesList
	feedTemplate *template.Template
}

func NewServer(feed StoryFeed, sources SourcesList) *Server{
	storyTemplate, err := ioutil.ReadFile("./storyTemplate.html")

	tmpl, err := template.New("test").Parse(string(storyTemplate))
	if err != nil { panic(err) }

	return &Server{
		feed:feed,
		feedTemplate:tmpl,
		sources: sources,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		if r.FormValue("action") =="addFeed" {
			s.sources.Add(r.FormValue("url"), r.FormValue("name"), r.FormValue("selector"))
		} else  {
			log.Println(r.FormValue("action"))
			if len(r.FormValue("publish")) > 0  {
				http.Redirect(w, r, "/publish?title="+r.FormValue("publish"), 301)
				return
			} else {
				s.feed.MarkAsProcessed(r.FormValue("title"))
			}
		}
	}

	type viewModel struct {
		SourcesNames []string
		Stories []Story
	}

	vm := viewModel{s.sources.GetNames(), s.feed.GetStories()}

	s.feedTemplate.Execute(w, vm)
}
