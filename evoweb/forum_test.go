package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseForumThread(t *testing.T) {
	html := `<div class="structItem--thread">
		<div class="structItem-title">
			<span class="label">Gameplay</span>
			<a href="/threads/test-thread.12345/">Test Thread Title</a>
		</div>
		<div class="structItem-minor">
			<span class="username">AuthorName</span>
			<time datetime="2025-10-15T12:00:00+0200"></time>
		</div>
		<div class="structItem-cell--meta">
			<dd>42</dd>
			<dd>1K</dd>
		</div>
		<div class="structItem-cell--latest">
			<time datetime="2026-04-01T10:00:00+0200"></time>
			<span class="username">LastPoster</span>
		</div>
	</div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".structItem--thread").First()

	ft := parseForumThread(sel)

	if ft.Title != "Test Thread Title" {
		t.Errorf("Title = %q, want %q", ft.Title, "Test Thread Title")
	}
	if ft.URL != "https://evoweb.uk/threads/test-thread.12345/" {
		t.Errorf("URL = %q, want %q", ft.URL, "https://evoweb.uk/threads/test-thread.12345/")
	}
	if ft.Author != "AuthorName" {
		t.Errorf("Author = %q, want %q", ft.Author, "AuthorName")
	}
	if ft.Date != "2025-10-15T12:00:00+0200" {
		t.Errorf("Date = %q, want %q", ft.Date, "2025-10-15T12:00:00+0200")
	}
	if ft.Prefix != "Gameplay" {
		t.Errorf("Prefix = %q, want %q", ft.Prefix, "Gameplay")
	}
	if ft.Replies != "42" {
		t.Errorf("Replies = %q, want %q", ft.Replies, "42")
	}
	if ft.Views != "1K" {
		t.Errorf("Views = %q, want %q", ft.Views, "1K")
	}
	if ft.LastDate != "2026-04-01T10:00:00+0200" {
		t.Errorf("LastDate = %q, want %q", ft.LastDate, "2026-04-01T10:00:00+0200")
	}
	if ft.LastAuthor != "LastPoster" {
		t.Errorf("LastAuthor = %q, want %q", ft.LastAuthor, "LastPoster")
	}
	if ft.Sticky {
		t.Error("Sticky = true, want false")
	}
}

func TestParseForumThread_Sticky(t *testing.T) {
	html := `<div class="structItem--thread">
		<div class="structItem-title">
			<a href="https://evoweb.uk/threads/sticky.999/">Sticky Thread</a>
		</div>
		<div class="structItem-minor">
			<span class="username">Admin</span>
		</div>
		<div class="structItem-cell--meta"></div>
		<div class="structItem-cell--latest"></div>
		<span class="structItem-status--sticky"></span>
	</div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".structItem--thread").First()

	ft := parseForumThread(sel)

	if !ft.Sticky {
		t.Error("Sticky = false, want true")
	}
	if ft.Title != "Sticky Thread" {
		t.Errorf("Title = %q, want %q", ft.Title, "Sticky Thread")
	}
}

func TestParseForumThread_MissingFields(t *testing.T) {
	html := `<div class="structItem--thread">
		<div class="structItem-title">
			<a href="/threads/minimal.1/">Minimal</a>
		</div>
		<div class="structItem-minor"></div>
		<div class="structItem-cell--meta"></div>
		<div class="structItem-cell--latest"></div>
	</div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".structItem--thread").First()

	ft := parseForumThread(sel)

	if ft.Title != "Minimal" {
		t.Errorf("Title = %q, want %q", ft.Title, "Minimal")
	}
	if ft.Author != "" {
		t.Errorf("Author = %q, want empty", ft.Author)
	}
	if ft.Prefix != "" {
		t.Errorf("Prefix = %q, want empty", ft.Prefix)
	}
}
