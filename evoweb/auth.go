package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	baseURL      = "https://evoweb.uk"
	loginPageURL = baseURL + "/login/"
	loginPostURL = baseURL + "/login/login"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func credentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot determine home directory: %v", err)
	}
	return filepath.Join(home, ".secrets", "evoweb", "credentials")
}

func saveCredentials(creds Credentials) error {
	path := credentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating secrets directory: %w", err)
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func loadCredentials() (Credentials, error) {
	data, err := os.ReadFile(credentialsPath())
	if err != nil {
		return Credentials{}, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return Credentials{}, fmt.Errorf("parsing credentials: %w", err)
	}
	return creds, nil
}

var tokenRegex = regexp.MustCompile(`name="_xfToken"\s+value="([^"]*)"`)

func extractCSRFToken(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", loginPageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching login page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login page returned HTTP %d", resp.StatusCode)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading login page: %w", err)
	}
	body := string(buf)

	matches := tokenRegex.FindStringSubmatch(body)
	if matches == nil {
		return "", fmt.Errorf("CSRF token not found on login page")
	}
	return matches[1], nil
}

func loginToForum(creds Credentials) (*http.Client, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	token, err := extractCSRFToken(client)
	if err != nil {
		return nil, err
	}

	form := url.Values{
		"login":       {creds.Username},
		"password":    {creds.Password},
		"_xfToken":    {token},
		"remember":    {"1"},
		"_xfRedirect": {baseURL + "/"},
	}

	req, err := http.NewRequest("POST", loginPostURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login POST failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for xf_user cookie (indicates successful login)
	u, _ := url.Parse(baseURL)
	for _, c := range jar.Cookies(u) {
		if c.Name == "xf_user" {
			return client, nil
		}
	}

	return nil, fmt.Errorf("login failed: no xf_user cookie received (check username/password)")
}

func newUnauthenticatedClient(cookie string) *http.Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}

	if cookie != "" {
		// Inject raw cookie string by setting it on all requests via a custom transport
		client.Transport = &cookieTransport{cookie: cookie}
	}

	return client
}

type cookieTransport struct {
	cookie string
}

func (t *cookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Cookie", t.cookie)
	return http.DefaultTransport.RoundTrip(req)
}

func buildClient(cookieStr string) (*http.Client, error) {
	if cookieStr != "" {
		return newUnauthenticatedClient(cookieStr), nil
	}

	creds, err := loadCredentials()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no credentials found; run 'evoweb login' first")
		}
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	log.Print("Logging in to evoweb.uk...")
	return loginToForum(creds)
}
