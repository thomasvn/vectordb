package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

const (
	TABLE_NAME         = "RssFeeds"
	MAX_CONTENT_LENGTH = 10000
)

var (
	weaviateURL  = os.Getenv("WEAVIATE_URL")  // "localhost:8080"
	openaiApiKey = os.Getenv("OPENAI_APIKEY") // ""
	rssFeeds     = os.Getenv("RSS_FEEDS")     // Comma-separated list of links "https://thomasvn.dev/feed/,https://golangweekly.com/rss/,https://kubernetes.io/feed.xml"
)

type RSSFeedProperties struct {
	Title     string
	Link      string
	Updated   string
	Published string
	Content   string
}

func main() {
	if weaviateURL == "" || openaiApiKey == "" || rssFeeds == "" {
		log.Fatal("WEAVIATE_URL, OPENAI_APIKEY, and RSS_FEEDS environment variables must be set")
	}

	if len(os.Args) < 2 {
		log.Fatal("Please provide a search query as a command line argument")
	}
	query := os.Args[1]

	cfg := weaviate.Config{
		Host:   weaviateURL,
		Scheme: "http",
		Headers: map[string]string{
			"X-OpenAI-Api-Key": openaiApiKey,
		},
	}
	client := NewWeaviateClient(cfg, TABLE_NAME)

	feeds := []RSSFeedProperties{}
	for _, url := range strings.Split(rssFeeds, ",") {
		feed := parseFeed(url)
		feeds = append(feeds, feed...)
	}

	client.InsertRSSFeeds(feeds)
	result := client.SearchRSSFeeds(query)

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	log.Printf("Query: %s\n", query)
	log.Printf("Results: %s\n", string(jsonData))
}

func parseFeed(url string) []RSSFeedProperties {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	results := []RSSFeedProperties{}
	for _, item := range feed.Items {
		converter := md.NewConverter("", true, nil)
		contentMd, _ := converter.ConvertString(item.Content)
		if len(contentMd) > MAX_CONTENT_LENGTH {
			contentMd = contentMd[:MAX_CONTENT_LENGTH] + "..."
		}

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
