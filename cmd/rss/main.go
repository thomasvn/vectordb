package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

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
		feed := parseFeed(url, MAX_CONTENT_LENGTH)
		feeds = append(feeds, feed...)
	}

	client.InsertRSSFeeds(feeds)

	result := client.SearchRSSFeeds(query)

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	log.Printf("Query: %s\nResults: %s\n", query, string(jsonData))
}
