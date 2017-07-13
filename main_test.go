package main

//import (
//	"testing"
//	"github.com/stretchr/testify/assert"
//)

//var testStory = Story{
//	Title: "test",
//	Source: "bbc",
//}


//func TestFeed_AddStories(t *testing.T) {
//	t.Run("adds a story", func(t *testing.T) {
//		var feed = NewFeed();
//		feed.AddStories([]Story{testStory})
//		assert.Len(t, feed.Stories, 1, "adds a story")
//	})
//
//	t.Run("does not add a story with matching title and source", func(t *testing.T) {
//		var feed = NewFeed();
//		feed.AddStories([]Story{testStory})
//		feed.AddStories([]Story{testStory})
//		assert.Len(t, feed.Stories, 1, "does not add the same story twice")
//	})
//
//	t.Run("adds story with same title but different source", func(t *testing.T) {
//		var feed = NewFeed();
//		feed.AddStories([]Story{testStory})
//		feed.AddStories([]Story{{Title: "test", Source: "reuters"}})
//		assert.Len(t, feed.Stories, 2, "does not add the same story twice")
//	})
//}

