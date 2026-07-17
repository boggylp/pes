package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"time"
)

func cmdDownload(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	cookie := fs.String("cookie", "", "cookie string (e.g. 'xf_session=abc; xf_user=def')")
	cookieFile := fs.String("cookie-file", "", "path to file containing cookie string")
	output := fs.String("output", "", "output file path (required)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() < 1 || *output == "" {
		fmt.Fprintln(os.Stderr, "usage: evoweb download --output <file> <attachment-url>")
		fs.PrintDefaults()
		os.Exit(1)
	}

	cookieStr := resolveCookie(*cookie, *cookieFile)
	client, err := buildClient(cookieStr)
	if err != nil {
		log.Fatal(err)
	}
	client.Timeout = 10 * time.Minute

	name, size, err := downloadFile(client, fs.Arg(0), *output)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%d bytes, server filename %q)", *output, size, name)
}

func downloadFile(client *http.Client, fileURL, outputPath string) (serverName string, size int64, err error) {
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("fetching %s: %w", fileURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("%s returned HTTP %d", fileURL, resp.StatusCode)
	}

	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil {
		serverName = params["filename"]
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = out.Close() }()

	size, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("writing %s: %w", outputPath, err)
	}
	if err := out.Close(); err != nil {
		return "", 0, fmt.Errorf("writing %s: %w", outputPath, err)
	}
	return serverName, size, nil
}
