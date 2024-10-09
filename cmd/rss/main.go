package main

import (
	"context"
	"fmt"
	"log"
	"os"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

const (
	TABLE_NAME = "rss_feeds"
)

var (
	weaviateURL  = os.Getenv("WEAVIATE_URL")  // "localhost:8080"
	openaiApiKey = os.Getenv("OPENAI_APIKEY") // ""
)

var rssFeeds = []string{
	"https://thomasvn.dev/feed/",
}

type RSSFeedProperties struct {
	Title     string
	Link      string
	Updated   string
	Published string
	Content   string
}

var schema = models.Object{
	Class:      TABLE_NAME,
	Properties: map[string]RSSFeedProperties{},
}

func main() {
	if weaviateURL == "" || openaiApiKey == "" {
		log.Fatal("WEAVIATE_URL and OPENAI_APIKEY environment variables must be set")
	}

	client := instantiateWeaviate()

	feeds := []RSSFeedProperties{}
	for _, url := range rssFeeds {
		feed := parseFeed(url)
		feeds = append(feeds, feed...)
	}

	insertRSSFeeds(client, feeds)
}

func parseFeed(url string) []RSSFeedProperties {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	results := []RSSFeedProperties{}
	for _, item := range feed.Items {
		converter := md.NewConverter("", true, nil)
		contentMd, _ := converter.ConvertString(item.Content)
		d := RSSFeedProperties{
			Title:     item.Title,
			Link:      item.Link,
			Updated:   item.Updated,
			Published: item.Published,
			Content:   contentMd,
		}
		results = append(results, d)
	}
	return results
}

func instantiateWeaviate() *weaviate.Client {
	cfg := weaviate.Config{
		Host:   weaviateURL,
		Scheme: "http",
		Headers: map[string]string{
			"X-OpenAI-Api-Key": openaiApiKey,
		},
	}

	client, _ := weaviate.NewClient(cfg)
	fmt.Println("Configuring Weaviate connection ...")

	exists, _ := client.Schema().ClassExistenceChecker().WithClassName(TABLE_NAME).Do(context.Background())
	if exists {
		fmt.Printf("Class %s already exists, skipping schema creation ...\n", TABLE_NAME)
		return client
	}

	classObj := &models.Class{
		Class:      TABLE_NAME,
		Vectorizer: "text2vec-openai",
		ModuleConfig: map[string]interface{}{
			"text2vec-openai": map[string]interface{}{},
		},
	}
	_ = client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	fmt.Printf("Class %s created ...\n", TABLE_NAME)

	return client
}

func insertRSSFeeds(client *weaviate.Client, rssFeeds []RSSFeedProperties) {
	objects := make([]*models.Object, len(rssFeeds))
	for i := range rssFeeds {
		objects[i] = &models.Object{
			Class: TABLE_NAME,
			Properties: map[string]any{
				"title":     rssFeeds[i].Title,
				"link":      rssFeeds[i].Link,
				"updated":   rssFeeds[i].Updated,
				"published": rssFeeds[i].Published,
				"content":   rssFeeds[i].Content,
			},
		}
	}

	batchRes, err := client.Batch().ObjectsBatcher().WithObjects(objects...).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to batch write objects: %v", err)
	}
	for _, res := range batchRes {
		if res.Result.Errors != nil {
			log.Fatalf("failed to batch write objects: %v", res.Result.Errors.Error)
		}
	}
}

// Function to search RSS feeds
