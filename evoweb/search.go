package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var threadLinkRegex = regexp.MustCompile(`/threads/[a-zA-Z0-9_\-%]+\.[0-9]+/`)

type SearchResult struct {
	Query   string   `json:"query"`
	Threads []string `json:"threads"`
}

func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("q", "", "search query (required)")
	user := fs.String("user", "", "restrict to posts by this username (optional)")
	output := fs.String("output", "", "output JSON file path (required)")
	maxPages := fs.Int("max-pages", 3, "max search-results pages to scan")
	delay := fs.Duration("delay", 500*time.Millisecond, "delay between page requests")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *query == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "usage: evoweb search -q <query> [--user <name>] --output <file>")
		fs.PrintDefaults()
		os.Exit(1)
	}

	client, err := buildClient("")
	if err != nil {
		log.Fatal(err)
	}

	// Step 1: POST /search/search to start a search session, follow the redirect
	form := url.Values{
		"keywords":  {*query},
		"c[title_only]": {"0"},
		"order":     {"date"},
	}
	if *user != "" {
		form.Set("users", *user)
	}

	req, err := http.NewRequest("POST", baseURL+"/search/search", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Need _xfToken
	token, err := extractCSRFToken(client)
	if err != nil {
		log.Fatal(err)
	}
	form.Set("_xfToken", token)
	req.Body = io.NopCloser(strings.NewReader(form.Encode()))
	req.ContentLength = int64(len(form.Encode()))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	finalURL := resp.Request.URL.String()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	log.Printf("search POST landed at: %s (HTTP %d, %d bytes)", finalURL, resp.StatusCode, len(body))

	threads := map[string]struct{}{}
	for _, m := range threadLinkRegex.FindAllString(string(body), -1) {
		threads[m] = struct{}{}
	}

	// Paginate
	if strings.Contains(finalURL, "/search/") {
		for page := 2; page <= *maxPages; page++ {
			time.Sleep(*delay)
			pageURL := strings.TrimSuffix(finalURL, "/") + fmt.Sprintf("/page-%d", page)
			req2, _ := http.NewRequest("GET", pageURL, nil)
			req2.Header.Set("User-Agent", "Mozilla/5.0")
			r2, err := client.Do(req2)
			if err != nil || r2.StatusCode != 200 {
				if r2 != nil {
					r2.Body.Close()
				}
				break
			}
			b2, _ := io.ReadAll(r2.Body)
			r2.Body.Close()
			before := len(threads)
			for _, m := range threadLinkRegex.FindAllString(string(b2), -1) {
				threads[m] = struct{}{}
			}
			log.Printf("page %d: +%d threads", page, len(threads)-before)
		}
	}

	var sorted []string
	for t := range threads {
		sorted = append(sorted, baseURL+t)
	}
	sort.Strings(sorted)

	writeJSON(SearchResult{Query: *query, Threads: sorted}, *output)
	log.Printf("found %d unique thread URLs", len(sorted))
}
