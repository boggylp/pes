package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
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

type searchOptions struct {
	query    string
	user     string
	output   string
	maxPages int
	delay    time.Duration
}

func cmdSearch(args []string) {
	opts := parseSearchFlags(args)

	client, err := buildClient("")
	if err != nil {
		log.Fatal(err)
	}

	finalURL, body := submitSearch(client, opts)

	threads := extractThreadLinks(body)
	if strings.Contains(finalURL, "/search/") {
		threads = collectResultPages(client, finalURL, opts, threads)
	}

	sorted := sortThreadURLs(threads)
	writeJSON(SearchResult{Query: opts.query, Threads: sorted}, opts.output)
	log.Printf("found %d unique thread URLs", len(sorted))
}

func parseSearchFlags(args []string) searchOptions {
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
	return searchOptions{query: *query, user: *user, output: *output, maxPages: *maxPages, delay: *delay}
}

func submitSearch(client *http.Client, opts searchOptions) (string, []byte) {
	form := url.Values{
		"keywords":      {opts.query},
		"c[title_only]": {"0"},
		"order":         {"date"},
	}
	if opts.user != "" {
		form.Set("users", opts.user)
	}

	token, err := extractCSRFToken(client)
	if err != nil {
		log.Fatal(err)
	}
	form.Set("_xfToken", token)

	req, err := http.NewRequest("POST", baseURL+"/search/search", strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	finalURL := resp.Request.URL.String()
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	log.Printf("search POST landed at: %s (HTTP %d, %d bytes)", finalURL, resp.StatusCode, len(body))
	return finalURL, body
}

func collectResultPages(client *http.Client, finalURL string, opts searchOptions, seen map[string]struct{}) map[string]struct{} {
	threads := seen
	for page := 2; page <= opts.maxPages; page++ {
		time.Sleep(opts.delay)
		pageURL := strings.TrimSuffix(finalURL, "/") + fmt.Sprintf("/page-%d", page)
		body, err := fetchBody(client, pageURL)
		if err != nil {
			break
		}
		merged := mergeThreadSets(threads, extractThreadLinks(body))
		log.Printf("page %d: +%d threads", page, len(merged)-len(threads))
		threads = merged
	}
	return threads
}

func extractThreadLinks(body []byte) map[string]struct{} {
	threads := map[string]struct{}{}
	for _, m := range threadLinkRegex.FindAllString(string(body), -1) {
		threads[m] = struct{}{}
	}
	return threads
}

func mergeThreadSets(a, b map[string]struct{}) map[string]struct{} {
	merged := make(map[string]struct{}, len(a)+len(b))
	maps.Copy(merged, a)
	maps.Copy(merged, b)
	return merged
}

func sortThreadURLs(threads map[string]struct{}) []string {
	var sorted []string
	for t := range threads {
		sorted = append(sorted, baseURL+t)
	}
	sort.Strings(sorted)
	return sorted
}
