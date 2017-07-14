package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/boltdb/bolt"
	"io/ioutil"
)

func main() {

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	feed := NewFeed(db)

	sources := NewRSSFetchers(db, feed)

	//sources.Add("http://www.londonstockexchange.com/exchange/CompanyNewsRSS.html?newsSource=RNS&indexSymbol=UKX", "LSE FTSE whatever", ".bg")
	//sources.Add("http://www.topazworld.com/en/rss/news", "Topaz", "#content")
	//sources.Add("https://investing.einnews.com/rss/5hMDxhc02nswfIlH", "EIN feed", ".none")
	//sources.Add("http://lorem-rss.herokuapp.com/feed?unit=second&interval=10", "Lorem ipsum feed", ".none"),

	dwAsLive := DebtwireAsLive{}
	publisher := NewPublisher(&dwAsLive, feed)

	router := http.NewServeMux()
	server := NewServer(feed, &sources)

	router.Handle("/", server)
	router.Handle("/publish", publisher)

	fmt.Println("Listening on 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

type ContentHub interface {
	Publish(body string) error
}

type StoryDatabase interface{
	GetStory(title string) Story
}

type Publisher struct {
	contentHub ContentHub
	storyDatabase StoryDatabase
}

func NewPublisher(contentHub ContentHub, storyDatabase StoryDatabase) *Publisher{
	return &Publisher{
		contentHub:contentHub,
		storyDatabase: storyDatabase,
	}
}

func (p *Publisher) ServeHTTP(w http.ResponseWriter,r *http.Request) {

	if r.Method==http.MethodPost{
		r.ParseForm()
		p.contentHub.Publish(r.FormValue("message-body"))
	}

	publishForm, err := ioutil.ReadFile("./publish-form.html")
	if err != nil {
		panic("problem reading form")
	}

	var body string
	if r.URL.Query().Get("title")!=""{
		body = p.storyDatabase.GetStory(r.URL.Query().Get("title")).Summary
	}

	fmt.Fprintf(w, string(publishForm), body)
}
