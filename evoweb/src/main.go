package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
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

func main() {
	cookie := flag.String("cookie", "", "cookie string (e.g. 'xf_session=abc; xf_user=def')")
	cookieFile := flag.String("cookie-file", "", "path to file containing cookie string")
	maxPages := flag.Int("max-pages", 0, "max pages to scrape (0 = all)")
	delay := flag.Duration("delay", 500*time.Millisecond, "delay between page requests")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: evoweb [flags] <thread-url>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cookieStr := resolveCookie(*cookie, *cookieFile)
	url := flag.Arg(0)

	thread, err := scrapeThread(url, cookieStr, *maxPages, *delay)
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(thread); err != nil {
		log.Fatal(err)
	}
}

func resolveCookie(cookie, cookieFile string) string {
	if cookie != "" {
		return cookie
	}
	if cookieFile != "" {
		data, err := os.ReadFile(cookieFile)
		if err != nil {
			log.Fatalf("reading cookie file: %v", err)
		}
		return strings.TrimSpace(string(data))
	}
	return ""
}

func scrapeThread(url, cookie string, maxPages int, delay time.Duration) (*Thread, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}

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

		doc, err := fetch(client, url, cookie)
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

func fetch(client *http.Client, url, cookie string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

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
	// Handle relative URLs
	if strings.HasPrefix(href, "/") {
		// Extract base from the current page
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
