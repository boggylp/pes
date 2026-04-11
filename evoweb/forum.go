package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ForumThread struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Author     string `json:"author"`
	Date       string `json:"date"`
	Prefix     string `json:"prefix,omitempty"`
	Replies    string `json:"replies"`
	Views      string `json:"views"`
	LastDate   string `json:"last_date"`
	LastAuthor string `json:"last_author"`
	Sticky     bool   `json:"sticky,omitempty"`
}

type Forum struct {
	Title   string        `json:"title"`
	URL     string        `json:"url"`
	Threads []ForumThread `json:"threads"`
}

func scrapeForum(client *http.Client, url string, maxPages int, delay time.Duration) (*Forum, error) {
	forum := &Forum{URL: url}
	page := 0

	for url != "" {
		page++
		if maxPages > 0 && page > maxPages {
			break
		}
		if page > 1 {
			time.Sleep(delay)
		}

		doc, err := fetch(client, url)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}

		if forum.Title == "" {
			forum.Title = strings.TrimSpace(doc.Find("h1.p-title-value").Text())
		}

		doc.Find(".structItem--thread").Each(func(_ int, s *goquery.Selection) {
			forum.Threads = append(forum.Threads, parseForumThread(s))
		})

		url = nextPageURL(doc)
		log.Printf("page %d: %d threads total", page, len(forum.Threads))
	}

	return forum, nil
}

func parseForumThread(s *goquery.Selection) ForumThread {
	titleLink := s.Find(".structItem-title a").Last()
	title := strings.TrimSpace(titleLink.Text())
	href, _ := titleLink.Attr("href")
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}

	author := strings.TrimSpace(s.Find(".structItem-minor .username").First().Text())

	var date string
	if t := s.Find(".structItem-minor time"); t.Length() > 0 {
		date, _ = t.First().Attr("datetime")
	}

	prefix := strings.TrimSpace(s.Find(".label").Text())

	var lastDate string
	if t := s.Find(".structItem-cell--latest time"); t.Length() > 0 {
		lastDate, _ = t.First().Attr("datetime")
	}
	lastAuthor := strings.TrimSpace(s.Find(".structItem-cell--latest .username").Text())

	dds := s.Find(".structItem-cell--meta dd")
	var replies, views string
	if dds.Length() >= 1 {
		replies = strings.TrimSpace(dds.Eq(0).Text())
	}
	if dds.Length() >= 2 {
		views = strings.TrimSpace(dds.Eq(1).Text())
	}

	sticky := s.Find(".structItem-status--sticky").Length() > 0

	return ForumThread{
		Title:      title,
		URL:        href,
		Author:     author,
		Date:       date,
		Prefix:     prefix,
		Replies:    replies,
		Views:      views,
		LastDate:   lastDate,
		LastAuthor: lastAuthor,
		Sticky:     sticky,
	}
}
