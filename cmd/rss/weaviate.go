package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

type WeaviateClient struct {
	client    *weaviate.Client
	tableName string
}

func NewWeaviateClient(cfg weaviate.Config, tableName string) *WeaviateClient {
	client, _ := weaviate.NewClient(cfg)
	log.Println("Configuring Weaviate connection ...")

	exists, _ := client.Schema().ClassExistenceChecker().WithClassName(tableName).Do(context.Background())
	if exists {
		log.Printf("Class %s already exists, skipping schema creation ...\n", tableName)
		return &WeaviateClient{client: client, tableName: tableName}
	}

	classObj := &models.Class{
		Class:      tableName,
		Vectorizer: "text2vec-openai",
		ModuleConfig: map[string]interface{}{
			"text2vec-openai": map[string]interface{}{},
		},
	}
	_ = client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	log.Printf("Class %s created ...\n", tableName)

	return &WeaviateClient{client: client, tableName: tableName}
}

func (w *WeaviateClient) InsertRSSFeeds(rssFeeds []RSSFeedProperties) {
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
			Class:      w.tableName,
			Properties: properties,
			ID:         generateUUID(rssFeeds[i].Link),
		}
	}

	log.Printf("Batch inserting %d objects ...\n", len(objects))
	batchRes, err := w.client.Batch().ObjectsBatcher().WithObjects(objects...).Do(context.Background())
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

func (w *WeaviateClient) SearchRSSFeeds(query string) map[string]models.JSONObject {
	fields := []graphql.Field{
		{Name: "title"},
		{Name: "link"},
		{Name: "updated"},
		{Name: "published"},
		{Name: "content"},
	}

	nearText := w.client.GraphQL().
		NearTextArgBuilder().
		WithConcepts([]string{query})

	result, err := w.client.GraphQL().Get().
		WithClassName(w.tableName).
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
