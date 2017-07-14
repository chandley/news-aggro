package main

import (
	"net/http"
	"html/template"
	"io/ioutil"
)

type StoryFeed interface{
	GetStories() []Story
	MarkAsProcessed(title string)
}

type Server struct {
	feed StoryFeed
	feedTemplate *template.Template

}

func NewServer(feed StoryFeed) *Server{
	storyTemplate, err := ioutil.ReadFile("./storyTemplate.html")

	tmpl, err := template.New("test").Parse(string(storyTemplate))
	if err != nil { panic(err) }

	return &Server{
		feed:feed,
		feedTemplate:tmpl,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		s.feed.MarkAsProcessed(r.FormValue("title"))
	}

	s.feedTemplate.Execute(w, s.feed.GetStories())
}
