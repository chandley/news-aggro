package main

import (
	"io"
	"time"
	"io/ioutil"
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"text/template"
	"sync"
	"sort"
	"encoding/json"
	"net/http"
)

type Story struct {
	Title, Description, Source, Link, Summary string
	Date *time.Time
	Processed bool
}

type Feed struct {
	Stories []Story
	tmpl *template.Template
	DB *bolt.DB
	sync.Mutex
}

func NewFeed(db *bolt.DB) *Feed {
	storyTemplate, err := ioutil.ReadFile("./storyTemplate.html")
	f := new(Feed)
	tmpl, err := template.New("test").Parse(string(storyTemplate))
	if err != nil { panic(err) }
	f.tmpl = tmpl

	f.DB = db

	db.Update(func(tx *bolt.Tx) error {
		_ , err := tx.CreateBucket([]byte("feed"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("feed"))
		rawStories := b.Get([]byte("stories"))

		if rawStories != nil {
			var storiesFromDisk []Story

			err := json.Unmarshal( rawStories, &storiesFromDisk)
			if err != nil {
				log.Println("problem parsing stories from bolt", err)
				return nil
			}

			log.Println("Loaded", len(storiesFromDisk), "from db", storiesFromDisk[0].Title)
			f.AddStories(storiesFromDisk)

		}
		return nil
	})

	if err != nil {
		log.Println("problem loading stories from bolt", err)
	}

	return f
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
		}
	}

	newStoryCount := len(f.Stories) - startingNumberOfStories

	f.SortStoriesByDate();
	f.SaveStories()

	if newStoryCount > 0 {
		log.Println(newStoryCount, "new stories added")
	}

}

func (f *Feed) SortStoriesByDate() (){
	sort.Slice(f.Stories, func(i, j int) bool { return f.Stories[i].Date.After(*f.Stories[j].Date) })
}


func (f *Feed) SaveStories() {
	err := f.DB.Update(func(tx *bolt.Tx) error {
		storiesAsJSON, _ := json.Marshal(f.Stories)
		b := tx.Bucket([]byte("feed"))
		err := b.Put([]byte("stories"), storiesAsJSON)
		return err
	})

	if err != nil {
		log.Println("Problem persisting stories to bolt")
	}
}

func (f *Feed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		title := r.FormValue("title")
		f.MarkAsProcessed(title)
	}

	f.RenderFeedAsHTML(w)
}

func (f *Feed) RenderFeedAsHTML(out io.Writer) {
	f.tmpl.Execute(out, f.Stories)
}

func (f *Feed) MarkAsProcessed(title string) {
	for i, story := range f.Stories {
		if story.Title == title {
			f.Stories[i].Processed = true
			f.SaveStories()
		}
	}
}



