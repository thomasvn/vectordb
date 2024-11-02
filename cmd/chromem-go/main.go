package main

import (
	"log"
	"os"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
	"github.com/philippgille/chromem-go"
)

var (
	openaiApiKey = os.Getenv("OPENAI_APIKEY") // ""
	rssFeeds     = os.Getenv("RSS_FEEDS")     // Comma-separated list of links "https://thomasvn.dev/feed/,https://golangweekly.com/rss/,https://kubernetes.io/feed.xml"
)

func main() {
	if openaiApiKey == "" || rssFeeds == "" {
		log.Fatal("OPENAI_APIKEY and RSS_FEEDS environment variables must be set")
	}
	if len(os.Args) < 2 {
		log.Fatal("Please provide a search query as a command line argument")
	}
	query := os.Args[1]

	cdb := chromemDB{}
	cdb.Instantiate()

	rfp := RssFeedParser{}
	rfp.Instantiate()
	feeds := rfp.ParseAllFeeds(rssFeeds)

	log.Printf("Feeds: %v\n", feeds)
	log.Printf("Parsed %d feeds\n", len(feeds))

	cdb.Insert(feeds)

	cdb.Query(query)

	// https://pkg.go.dev/github.com/philippgille/chromem-go#section-readme
}

type chromemDB struct {
	db            *chromem.DB
	rssCollection *chromem.Collection // only a single collection will be used
}

func (cdb *chromemDB) Instantiate() {
	cdb.db = chromem.NewDB()
	cdb.rssCollection, _ = cdb.db.CreateCollection("RssFeeds", nil, nil)
}

func (cdb *chromemDB) Insert(feeds []RSSFeedProperties) {
	// TODO
}

func (cdb *chromemDB) Query(query string) {
	// TODO
}

func (cdb *chromemDB) Export() {
	// TODO
}

func (cdb *chromemDB) Import() {
	// TODO
}

type RssFeedParser struct {
	maxContentLength int
	mdConverter      *md.Converter
}

type RSSFeedProperties struct {
	Title     string
	Link      string
	Updated   string
	Published string
	Content   string
}

func (rfp *RssFeedParser) Instantiate() {
	rfp.maxContentLength = 10000
	rfp.mdConverter = md.NewConverter("", true, nil)
}

func (rfp *RssFeedParser) ParseAllFeeds(rssFeeds string) []RSSFeedProperties {
	results := []RSSFeedProperties{}
	for _, url := range strings.Split(rssFeeds, ",") {
		feed := rfp.parseFeed(url)
		results = append(results, feed...)
	}
	return results
}

func (rfp *RssFeedParser) parseFeed(url string) []RSSFeedProperties {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	results := []RSSFeedProperties{}
	for _, item := range feed.Items {
		descriptionMd, _ := rfp.mdConverter.ConvertString(item.Description)
		contentMd, _ := rfp.mdConverter.ConvertString(item.Content)
		resultMd := descriptionMd + "\n\n" + contentMd
		if len(resultMd) > rfp.maxContentLength {
			resultMd = resultMd[:rfp.maxContentLength] + "..."
		}

		d := RSSFeedProperties{
			Title:     item.Title,
			Link:      item.Link,
			Updated:   item.Updated,
			Published: item.Published,
			Content:   resultMd,
		}
		results = append(results, d)
	}
	return results
}
