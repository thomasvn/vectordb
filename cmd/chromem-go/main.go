package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/mmcdole/gofeed"
	"github.com/philippgille/chromem-go"
	"github.com/tmc/langchaingo/textsplitter"
	"google.golang.org/api/option"
)

var (
	OpenaiApiKey              = os.Getenv("OPENAI_API_KEY")               // ""
	RssFeeds                  = os.Getenv("RSS_FEEDS")                    // Comma-separated list of links "https://thomasvn.dev/feed/,https://golangweekly.com/rss/,https://kubernetes.io/feed.xml"
	ServiceKeyCredentialsFile = os.Getenv("SERVICE_KEY_CREDENTIALS_FILE") // Optional. "./service-key.json"
	GcsBucketName             = "thomasvn-rss-search"
)

func main() {
	if OpenaiApiKey == "" || RssFeeds == "" {
		log.Fatal("OPENAI_API_KEY and RSS_FEEDS environment variables must be set")
	}
	if len(os.Args) < 2 {
		log.Fatal("Please provide a search query as a command line argument")
	}
	query := os.Args[1]

	db := InitChromemDB()
	parser := InitRssFeedParser()

	feeds := parser.ParseAllFeeds(RssFeeds)

	db.Insert(feeds)

	results := db.Query(query)
	log.Printf("Found %d results\n", len(results))
	log.Printf("Results: %v\n", results)
}

type ChromemDB struct {
	db          *chromem.DB
	concurrency int
}

func InitChromemDB() *ChromemDB {
	cdb := ChromemDB{
		db:          chromem.NewDB(),
		concurrency: 500,
	}
	cdb.db.CreateCollection("RssFeeds", nil, nil)

	if ServiceKeyCredentialsFile != "" {
		cdb.ImportFromGCS()
	} else {
		cdb.Import()
	}

	return &cdb
}

func (cdb *ChromemDB) Insert(feeds []RSSFeedProperties) {
	ids := []string{}
	metadatas := []map[string]string{}
	contents := []string{}

	rssCollection := cdb.db.GetCollection("RssFeeds", nil)

	for _, feed := range feeds {
		// Do not insert if it already exists
		if _, err := rssCollection.GetByID(context.TODO(), feed.UID); err == nil {
			continue
		}

		ids = append(ids, feed.UID)
		metadatas = append(metadatas, map[string]string{
			"title":     feed.Title,
			"link":      feed.Link,
			"updated":   feed.Updated,
			"published": feed.Published,
		})
		contents = append(contents, feed.Content)
	}

	if len(ids) == 0 {
		log.Printf("No new entries to insert\n")
		return
	}

	err := rssCollection.AddConcurrently(context.TODO(), ids, nil, metadatas, contents, cdb.concurrency)
	if err != nil {
		log.Fatalf("Error inserting entries: %s", err.Error())
	}
	log.Printf("Inserted %d entries\n", len(ids))

	if ServiceKeyCredentialsFile != "" {
		cdb.ExportToGCS()
	} else {
		cdb.Export()
	}
}

func (cdb *ChromemDB) Query(query string) []string {
	maxQueryResults := 5

	rssCollection := cdb.db.GetCollection("RssFeeds", nil)

	results, _ := rssCollection.Query(context.TODO(), query, maxQueryResults, nil, nil)

	formattedResults := []string{}
	for _, result := range results {
		formattedResults = append(formattedResults, fmt.Sprintf("Title: %s\nSimilarity: %f\nLink: %s\n", result.Metadata["title"], result.Similarity, result.Metadata["link"]))
	}
	return formattedResults
}

func (cdb *ChromemDB) Export() {
	if err := cdb.db.ExportToFile("chromem-go.gob.gz", true, ""); err != nil {
		log.Printf("WARN: Error exporting DB: %s", err.Error())
		return
	}
	log.Printf("Exported DB to file chromem-go.gob.gz")
}

func (cdb *ChromemDB) Import() {
	if err := cdb.db.ImportFromFile("chromem-go.gob.gz", ""); err != nil {
		log.Printf("WARN: Error importing DB: %s", err.Error())
		return
	}
	log.Printf("Imported DB from file chromem-go.gob.gz")
}

func (cdb *ChromemDB) ExportToGCS() {
	client, err := storage.NewClient(context.Background(), option.WithCredentialsFile(ServiceKeyCredentialsFile))
	if err != nil {
		log.Printf("WARN: failed to create storage client: %s", err.Error())
		return
	}
	defer client.Close()

	obj := client.Bucket(GcsBucketName).Object("chromem-go.gob.gz")
	writer := obj.NewWriter(context.Background())
	defer writer.Close()

	if err := cdb.db.ExportToWriter(writer, true, ""); err != nil {
		log.Printf("WARN: failed to write to GCS: %s", err.Error())
	}
	log.Printf("Exported chromem-go.gob.gz to GCS")
}

func (cdb *ChromemDB) ImportFromGCS() {
	client, err := storage.NewClient(context.Background(), option.WithCredentialsFile(ServiceKeyCredentialsFile))
	if err != nil {
		log.Printf("WARN: failed to create storage client: %s", err.Error())
		return
	}
	defer client.Close()

	obj := client.Bucket(GcsBucketName).Object("chromem-go.gob.gz")
	reader, err := obj.NewReader(context.Background())
	if err != nil {
		log.Printf("WARN: failed to read from GCS: %s", err.Error())
		return
	}
	defer reader.Close()

	seekableReader, err := NewSeekableReader(reader)
	if err != nil {
		log.Printf("WARN: failed to create seekable reader: %s", err.Error())
		return
	}

	if err := cdb.db.ImportFromReader(seekableReader, ""); err != nil {
		log.Printf("WARN: failed to import from GCS: %s", err.Error())
	}
	log.Printf("Imported chromem-go.gob.gz from GCS")
}

type RssFeedParser struct {
	maxContentLength int // TODO: Chunking
}

type RSSFeedProperties struct {
	UID       string
	Title     string
	Link      string
	Updated   string
	Published string
	Content   string
}

func InitRssFeedParser() *RssFeedParser {
	return &RssFeedParser{
		maxContentLength: 10000,
	}
}

func (rfp *RssFeedParser) ParseAllFeeds(rssFeeds string) []RSSFeedProperties {
	results := []RSSFeedProperties{}

	// var wg sync.WaitGroup
	// var mu sync.Mutex

	for _, url := range strings.Split(rssFeeds, ",") {
		feed := rfp.parseFeed(url)
		results = append(results, feed...)

		// wg.Add(1)
		// go func(url string) {
		// 	defer wg.Done()
		// 	feed := rfp.parseFeed(url)
		// 	mu.Lock()
		// 	results = append(results, feed...)
		// 	mu.Unlock()
		// }(url)
	}

	// wg.Wait()
	log.Printf("Parsed %d entries\n", len(results))
	return results
}

func (rfp *RssFeedParser) parseFeed(url string) []RSSFeedProperties {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	results := []RSSFeedProperties{}
	for _, item := range feed.Items {
		titleMd, _ := md.ConvertString(item.Title)
		descriptionMd, _ := md.ConvertString(item.Description)
		contentMd, _ := md.ConvertString(item.Content)
		resultMd := titleMd + "\n\n" + descriptionMd + "\n\n" + contentMd

		// TODO: Chunking

		// https://pkg.go.dev/github.com/pkoukk/tiktoken-go
		// tkm, _ := tiktoken.GetEncoding("cl100k_base")
		// token := tkm.Encode(resultMd, nil, nil)
		// log.Printf("Token Length: %d\n", len(token))

		// https://pkg.go.dev/github.com/tmc/langchaingo/textsplitter
		splitter := textsplitter.NewMarkdownTextSplitter(
			textsplitter.WithChunkSize(5000),
			textsplitter.WithChunkOverlap(1000),
			textsplitter.WithCodeBlocks(true),
			textsplitter.WithEncodingName("cl100k_base"),
		)
		chunks, _ := splitter.SplitText(contentMd)
		log.Printf("NumChunks1: %d\n", len(chunks))

		splitter2 := textsplitter.NewTokenSplitter(
			textsplitter.WithChunkSize(5000),
			textsplitter.WithChunkOverlap(1000),
			textsplitter.WithEncodingName("cl100k_base"),
		)
		chunks2, _ := splitter2.SplitText(contentMd)
		log.Printf("NumChunks2: %d\n", len(chunks2))

		if len(resultMd) > rfp.maxContentLength {
			resultMd = resultMd[:rfp.maxContentLength] + "..."
		}

		d := RSSFeedProperties{
			UID:       rfp.generateUID(item.Link),
			Title:     item.Title,
			Link:      item.Link,
			Updated:   item.Updated,
			Published: item.PublishedParsed.String(),
			Content:   resultMd,
		}
		results = append(results, d)
	}
	return results
}

func (rfp *RssFeedParser) generateUID(input string) string {
	input = strings.ToLower(input)
	hash := md5.Sum([]byte(input))
	uid := fmt.Sprintf("%x-%x-%x-%x-%x", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:])
	return uid
}

// Embeds a bytes.Reader which implements io.ReadSeeker. chromem-go requires it
// but GCS Reader does not implement it
type seekableReader struct {
	*bytes.Reader
	original *storage.Reader
}

func NewSeekableReader(r *storage.Reader) (*seekableReader, error) {
	// Read all content into buffer
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Create bytes.Reader which implements io.ReadSeeker
	return &seekableReader{
		Reader:   bytes.NewReader(content),
		original: r,
	}, nil
}

func (sr *seekableReader) Close() error {
	return sr.original.Close()
}
