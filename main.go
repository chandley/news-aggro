package main

import (
"github.com/mmcdole/gofeed"
"fmt"
"net/http"
"log"
)

func main() {
	feedServer := NewFeed()

	feedServer.AddStories(StoriesFromFeed("http://www.londonstockexchange.com/exchange/CompanyNewsRSS.html?newsSource=RNS&indexSymbol=UKX"))
	feedServer.AddStories(StoriesFromFeed("http://feeds.bbci.co.uk/news/rss.xml?edition=uk"))

	router := http.NewServeMux()
	router.Handle("/", feedServer)

	fmt.Println("Listening on 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

func StoriesFromFeed(url string) (stories []Story){
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	for _, article := range feed.Items{
		stories = append(stories, Story{Title: article.Title, Description:article.Description})
	}

	return

}

type Story struct{
	Title, Description string
}

func NewFeed() *Feed{
	return new(Feed)
}

type Feed struct {
	Stories []Story
}

func (f *Feed) AddStory(s Story) {
	f.Stories = append(f.Stories, s)
}

func (f *Feed) AddStories(s []Story) {
	f.Stories = append(f.Stories, s...)
}

func (f *Feed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, story := range f.Stories{
		fmt.Fprintf(w, "Title: %s, Description: %s\n", story.Title, story.Description)
	}
}

