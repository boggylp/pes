package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestCleanText_CollapsesWhitespace(t *testing.T) {
	input := "hello  \t \n \t  world"
	want := "hello\nworld"
	if got := cleanText(input); got != want {
		t.Errorf("cleanText() = %q, want %q", got, want)
	}
}

func TestCleanText_CollapsesBlankLines(t *testing.T) {
	input := "hello\n\n\n\n\nworld"
	want := "hello\n\nworld"
	if got := cleanText(input); got != want {
		t.Errorf("cleanText() = %q, want %q", got, want)
	}
}

func TestCleanText_TrimsSpace(t *testing.T) {
	input := "  \n  hello world  \n  "
	want := "hello world"
	if got := cleanText(input); got != want {
		t.Errorf("cleanText() = %q, want %q", got, want)
	}
}

func TestCleanText_EmptyString(t *testing.T) {
	if got := cleanText(""); got != "" {
		t.Errorf("cleanText(\"\") = %q, want \"\"", got)
	}
}

func TestCleanText_PreservesDoubleNewline(t *testing.T) {
	input := "paragraph one\n\nparagraph two"
	want := "paragraph one\n\nparagraph two"
	if got := cleanText(input); got != want {
		t.Errorf("cleanText() = %q, want %q", got, want)
	}
}

func TestNextPageURL_AbsoluteHref(t *testing.T) {
	html := `<html><body><a class="pageNav-jump--next" href="https://evoweb.uk/threads/test.123/page-2">Next</a></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	got := nextPageURL(doc)
	want := "https://evoweb.uk/threads/test.123/page-2"
	if got != want {
		t.Errorf("nextPageURL() = %q, want %q", got, want)
	}
}

func TestNextPageURL_RelativeHref(t *testing.T) {
	html := `<html><head><link rel="canonical" href="https://evoweb.uk/threads/test.123/"/></head>
	<body><a class="pageNav-jump--next" href="/threads/test.123/page-2">Next</a></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	got := nextPageURL(doc)
	want := "https://evoweb.uk/threads/test.123/page-2"
	if got != want {
		t.Errorf("nextPageURL() = %q, want %q", got, want)
	}
}

func TestNextPageURL_NoLink(t *testing.T) {
	html := `<html><body><p>no pagination</p></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	if got := nextPageURL(doc); got != "" {
		t.Errorf("nextPageURL() = %q, want empty", got)
	}
}

func TestParsePost(t *testing.T) {
	html := `<article class="message--post" data-author="TestUser">
		<time class="u-dt" datetime="2024-06-15T10:30:00+0000"></time>
		<article class="message-body"><div class="bbWrapper">Hello world!</div></article>
	</article>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find("article.message--post").First()

	post := parsePost(sel)

	if post.Author != "TestUser" {
		t.Errorf("Author = %q, want %q", post.Author, "TestUser")
	}
	if post.Date != "2024-06-15T10:30:00+0000" {
		t.Errorf("Date = %q, want %q", post.Date, "2024-06-15T10:30:00+0000")
	}
	if post.Content != "Hello world!" {
		t.Errorf("Content = %q, want %q", post.Content, "Hello world!")
	}
}

func TestParsePost_MissingFields(t *testing.T) {
	html := `<article class="message--post">
		<article class="message-body"><div class="bbWrapper">  </div></article>
	</article>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find("article.message--post").First()

	post := parsePost(sel)

	if post.Author != "" {
		t.Errorf("Author = %q, want empty", post.Author)
	}
	if post.Date != "" {
		t.Errorf("Date = %q, want empty", post.Date)
	}
	if post.Content != "" {
		t.Errorf("Content = %q, want empty", post.Content)
	}
}

func TestLastPageNumber_MultiplePages(t *testing.T) {
	html := `<html><body>
		<li class="pageNav-page"><a>1</a></li>
		<li class="pageNav-page"><a>2</a></li>
		<li class="pageNav-page"><a>3</a></li>
		<li class="pageNav-page"><a>10</a></li>
	</body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	if got := lastPageNumber(doc); got != 10 {
		t.Errorf("lastPageNumber() = %d, want 10", got)
	}
}

func TestLastPageNumber_SinglePage(t *testing.T) {
	html := `<html><body><p>no pagination</p></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	if got := lastPageNumber(doc); got != 1 {
		t.Errorf("lastPageNumber() = %d, want 1", got)
	}
}
