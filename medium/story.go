package medium

import (
	"errors"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Story struct {
	Description string
	Title       string
	Url         string `db:"HASH"`
	Author      string
	ImageUrl    string
	Published   string
	Collection  string
}

func GetStory(mediumUrl string) (Story, error) {
	var doc *goquery.Document
	var e error
	var story Story

	// stories must look like https://medium.com/collection-name/story-id
	// or at least https://medium.com/p/story-id
	u, e := url.Parse(mediumUrl)
	if e != nil {
		return story, e
	}

	if doc, e = goquery.NewDocument(mediumUrl); e != nil {
		return story, e
	}

	canonicalUrl, _ := doc.Find("link[rel=\"canonical\"]").Attr("href")
	canonicalUrl = strings.TrimRight(canonicalUrl, "/")
	u, _ = url.Parse(canonicalUrl)

	// parse the canonical url to determine the originating collection
	pathParts := strings.Split(u.Path, "/")
	if len(pathParts) != 3 {
		return story, errors.New("invalid story url: " + canonicalUrl)
	}
	if pathParts[1] != "p" {
		story.Collection = "https://medium.com/" + pathParts[1]
	} else {
		story.Collection = ""
	}

	story.Url = canonicalUrl

	// there can be more than one rel="author" but we just want the medium one
	doc.Find("link[rel=\"author\"]").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			u, err := url.Parse(href)
			if err == nil && u.Host == "medium.com" {
				story.Author = href
			}
		}
	})

	parts := strings.Split(doc.Find("title").Text(), "—")
	if len(parts) > 0 {
		story.Title = strings.TrimSpace(parts[0])
	}

	// stories that lack image headers seem to have broken urls as og:image
	// metadata; they have a path that starts with double slashes ; ignore them
	story.ImageUrl, _ = doc.Find("meta[property=\"og:image\"]").Attr("content")
	u, err := url.Parse(story.ImageUrl)
	if err == nil && strings.HasPrefix(u.Path, "//") {
		story.ImageUrl = ""
	}

	story.Description, _ = doc.Find("meta[name=\"description\"]").Attr("content")
	story.Published, _ = doc.Find("meta[property=\"article:published_time\"]").Attr("content")

	return story, nil
}
