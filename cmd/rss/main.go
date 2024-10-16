package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-openapi/strfmt"
	"github.com/mmcdole/gofeed"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

const (
	TABLE_NAME = "RssFeeds"
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

var schema = models.Object{
	Class:      TABLE_NAME,
	Properties: map[string]RSSFeedProperties{},
}

func main() {
	if weaviateURL == "" || openaiApiKey == "" || rssFeeds == "" {
		log.Fatal("WEAVIATE_URL, OPENAI_APIKEY, and RSS_FEEDS environment variables must be set")
	}

	if len(os.Args) < 2 {
		log.Fatal("Please provide a search query as a command line argument")
	}
	query := os.Args[1]

	client := instantiateWeaviate()

	feeds := []RSSFeedProperties{}
	for _, url := range strings.Split(rssFeeds, ",") {
		feed := parseFeed(url)
		feeds = append(feeds, feed...)
	}

	insertRSSFeeds(client, feeds)
	result := searchRSSFeeds(client, query)

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
	log.Println("Configuring Weaviate connection ...")

	exists, _ := client.Schema().ClassExistenceChecker().WithClassName(TABLE_NAME).Do(context.Background())
	if exists {
		log.Printf("Class %s already exists, skipping schema creation ...\n", TABLE_NAME)
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
	log.Printf("Class %s created ...\n", TABLE_NAME)

	return client
}

func insertRSSFeeds(client *weaviate.Client, rssFeeds []RSSFeedProperties) {
	objects := make([]*models.Object, len(rssFeeds))
	for i := range rssFeeds {
		properties := map[string]any{
			"title":   rssFeeds[i].Title,
			"link":    rssFeeds[i].Link,
			"content": rssFeeds[i].Content,
		}
		if rssFeeds[i].Updated != "" {
			if t, err := parseAndFormatDate(rssFeeds[i].Updated); err == nil {
				properties["updated"] = t
			} else {
				log.Printf("Warning: Unable to parse Updated date for %s: %v", rssFeeds[i].Link, err)
			}
		}
		if rssFeeds[i].Published != "" {
			if t, err := parseAndFormatDate(rssFeeds[i].Published); err == nil {
				properties["published"] = t
			} else {
				log.Printf("Warning: Unable to parse Published date for %s: %v", rssFeeds[i].Link, err)
			}
		}

		objects[i] = &models.Object{
			Class:      TABLE_NAME,
			Properties: properties,
			ID:         generateUUID(rssFeeds[i].Link),
		}
	}

	log.Printf("Batch inserting %d objects ...\n", len(objects))
	batchRes, err := client.Batch().ObjectsBatcher().WithObjects(objects...).Do(context.Background())
	if err != nil {
		log.Fatalf("failed to batch write objects: %v", err)
	}
	for _, res := range batchRes {
		if res.Result.Errors != nil {
			errorsJSON, _ := json.MarshalIndent(res.Result.Errors, "", "  ")
			log.Fatalf("failed to batch write objects: %s\n", string(errorsJSON))
		}
	}
}

func generateUUID(input string) strfmt.UUID {
	input = strings.ToLower(input)
	hash := md5.Sum([]byte(input))
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:])
	return strfmt.UUID(uuid)
}

func parseAndFormatDate(dateStr string) (string, error) {
	formats := []string{
		time.RFC3339,
		time.RFC1123,
		time.RFC1123Z,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 +0000",
		"2006-01-02T15:04:05Z",
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t.Format(time.RFC3339), nil
		}
	}

	return "", fmt.Errorf("unable to parse date: %s", dateStr)
}

func searchRSSFeeds(client *weaviate.Client, query string) map[string]models.JSONObject {
	fields := []graphql.Field{
		{Name: "title"},
		{Name: "link"},
		{Name: "updated"},
		{Name: "published"},
		{Name: "content"},
	}

	nearText := client.GraphQL().
		NearTextArgBuilder().
		WithConcepts([]string{query})

	result, err := client.GraphQL().Get().
		WithClassName(TABLE_NAME).
		WithFields(fields...).
		WithNearText(nearText).
		WithLimit(5).
		Do(context.Background())
	if err != nil {
		log.Fatalf("failed to perform semantic search: %v", err)
	}
	if result.Errors != nil {
		errorsJSON, _ := json.MarshalIndent(result.Errors, "", "  ")
		log.Fatalf("failed to perform semantic search: %s\n", string(errorsJSON))
	}

	return result.Data
}
