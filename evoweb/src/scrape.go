package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Post struct {
	Author  string `json:"author"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

type Thread struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Posts []Post `json:"posts"`
}

func scrapeThread(client *http.Client, url string, maxPages int, delay time.Duration) (*Thread, error) {
	thread := &Thread{URL: url}
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

		if thread.Title == "" {
			thread.Title = strings.TrimSpace(doc.Find("h1.p-title-value").Text())
		}

		doc.Find("article.message--post").Each(func(_ int, s *goquery.Selection) {
			thread.Posts = append(thread.Posts, parsePost(s))
		})

		url = nextPageURL(doc)
		log.Printf("page %d: %d posts total", page, len(thread.Posts))
	}

	return thread, nil
}

func fetch(client *http.Client, url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func parsePost(s *goquery.Selection) Post {
	author, _ := s.Attr("data-author")

	var date string
	if t := s.Find("time.u-dt"); t.Length() > 0 {
		date, _ = t.First().Attr("datetime")
	}

	content := cleanText(s.Find("article.message-body .bbWrapper").Text())

	return Post{
		Author:  author,
		Date:    date,
		Content: content,
	}
}

var collapseWhitespace = regexp.MustCompile(`[\t ]*\n[\t ]*`)
var collapseBlankLines = regexp.MustCompile(`\n{3,}`)

func cleanText(s string) string {
	s = collapseWhitespace.ReplaceAllString(s, "\n")
	s = collapseBlankLines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

func nextPageURL(doc *goquery.Document) string {
	href, exists := doc.Find("a.pageNav-jump--next").Attr("href")
	if !exists {
		return ""
	}
	if strings.HasPrefix(href, "/") {
		if base, exists := doc.Find("link[rel=canonical]").Attr("href"); exists {
			parts := strings.SplitN(base, "//", 2)
			if len(parts) == 2 {
				slashIdx := strings.Index(parts[1], "/")
				if slashIdx > 0 {
					return parts[0] + "//" + parts[1][:slashIdx] + href
				}
			}
		}
	}
	return href
}
