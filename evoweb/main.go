package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "login":
		cmdLogin()
	case "scrape":
		cmdScrape(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: evoweb <command> [args]

commands:
  login              save evoweb.uk credentials
  scrape [flags] URL scrape a XenForo thread`)
}

func cmdLogin() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		log.Fatal("username and password are required")
	}

	creds := Credentials{Username: username, Password: password}

	fmt.Print("Logging in to evoweb.uk... ")
	_, err := loginToForum(creds)
	if err != nil {
		log.Fatalf("failed: %v", err)
	}
	fmt.Println("OK")

	if err := saveCredentials(creds); err != nil {
		log.Fatalf("saving credentials: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Credentials saved to %s\n", credentialsPath())
}

func cmdScrape(args []string) {
	fs := flag.NewFlagSet("scrape", flag.ExitOnError)
	cookie := fs.String("cookie", "", "cookie string (e.g. 'xf_session=abc; xf_user=def')")
	cookieFile := fs.String("cookie-file", "", "path to file containing cookie string")
	maxPages := fs.Int("max-pages", 0, "max pages to scrape (0 = all)")
	lastPages := fs.Int("last-pages", 0, "scrape only the last N pages")
	delay := fs.Duration("delay", 500*time.Millisecond, "delay between page requests")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: evoweb scrape [flags] <thread-url>")
		fs.PrintDefaults()
		os.Exit(1)
	}

	threadURL := fs.Arg(0)

	cookieStr := resolveCookie(*cookie, *cookieFile)
	client, err := buildClient(cookieStr)
	if err != nil {
		log.Fatal(err)
	}

	scrapeURL := threadURL
	scrapeMax := *maxPages

	if *lastPages > 0 {
		var pages int
		log.Printf("Discovering page count for thread...")
		scrapeURL, pages, err = resolveStartURL(client, threadURL, *lastPages, *delay)
		if err != nil {
			log.Fatalf("discovering pages: %v", err)
		}
		scrapeMax = pages
		log.Printf("Scraping last %d pages (starting from %s)", pages, scrapeURL)
	}

	thread, err := scrapeThread(client, scrapeURL, scrapeMax, *delay)
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
