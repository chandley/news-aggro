package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

var testStory = Story{
	Title: "test",
	Source: "bbc",
}




func TestFeed_AddStories(t *testing.T) {db, err := bolt.Open("test.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove("test.db")

	t.Run("adds a story", func(t *testing.T) {
		var feed = NewFeed(db);
		feed.AddStories([]Story{testStory})
		assert.Len(t, feed.Stories, 1, "adds a story")
	})

	t.Run("does not add a story with matching title and source", func(t *testing.T) {
		var feed = NewFeed(db);
		feed.AddStories([]Story{testStory})
		feed.AddStories([]Story{testStory})
		assert.Len(t, feed.Stories, 1, "does not add the same story twice")
	})

	t.Run("adds story with same title but different source", func(t *testing.T) {
		var feed = NewFeed(db);
		feed.AddStories([]Story{testStory})
		feed.AddStories([]Story{{Title: "test", Source: "reuters"}})
		assert.Len(t, feed.Stories, 2, "does not add the same story twice")
	})
}

