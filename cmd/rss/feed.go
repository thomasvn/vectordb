package main

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
)

type RSSFeedProperties struct {
	Title     string
	Link      string
	Updated   string
	Published string
	Content   string
}

func parseFeed(url string, maxContentLength int) []RSSFeedProperties {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	results := []RSSFeedProperties{}
	for _, item := range feed.Items {
		converter := md.NewConverter("", true, nil)
		descriptionMd, _ := converter.ConvertString(item.Description)
		contentMd, _ := converter.ConvertString(item.Content)
		resultMd := descriptionMd + "\n\n" + contentMd
		if len(resultMd) > maxContentLength {
			resultMd = resultMd[:maxContentLength] + "..."
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
