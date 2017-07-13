package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	feedServer := NewFeed()

	sources := []*RSSFetcher {
		//NewRSSFetcher("http://feeds.bbci.co.uk/news/rss.xml?edition=uk", "bbc news", ".story-body__inner", feedServer),
		NewRSSFetcher("http://www.londonstockexchange.com/exchange/CompanyNewsRSS.html?newsSource=RNS&indexSymbol=UKX", "LSE FTSE whatever", ".bg", feedServer),
		NewRSSFetcher("http://www.topazworld.com/en/rss/news", "Topaz", "#content", feedServer),
		//NewRSSFetcher("https://twitrss.me/twitter_search_to_rss/?term=bieber", "Bieber News from Twitter", ".tweet-text", feedServer),
		//NewRSSFetcher("https://twitrss.me/twitter_user_to_rss/?user=quii", "Chris James News from Twitter", ".tweet-text",feedServer),
		//NewRSSFetcher("https://twitrss.me/twitter_user_to_rss/?user=chrisrhandley", "Chris Handley News from Twitter", ".tweet-text",feedServer),
		//NewRSSFetcher("http://lorem-rss.herokuapp.com/feed?unit=second&interval=10", "Lorem ipsum feed", ".none",feedServer),
		//NewRSSFetcher("https://www.sec.gov/cgi-bin/browse-edgar?action=getcurrent&type=10-k&company=&dateb=&owner=include&start=0&count=40&output=atom", "SEC 10k", ".none", feedServer),
		//NewRSSFetcher("https://www.sec.gov/Archives/edgar/usgaap.rss.xml", "SEC gaap", ".none", feedServer),
		NewRSSFetcher("https://investing.einnews.com/rss/5hMDxhc02nswfIlH", "EIN feed", ".none", feedServer),
	}

	for _, fetcher := range sources {
		go fetcher.ListenForUpdates()
	}

	router := http.NewServeMux()
	router.Handle("/", feedServer)

	fmt.Println("Listening on 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
