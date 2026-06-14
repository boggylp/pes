package main

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Cristiano Ronaldo", "cristiano ronaldo"},
		{"Fábio", "fabio"},
		{"Jürgen Klopp", "jurgen klopp"},
		{"R. Santa Cruz", "r santa cruz"},
		{"Ange Postecoglou", "ange postecoglou"},
		{"", ""},
	}

	for _, tt := range tests {
		got := normalize(tt.input)
		if got != tt.want {
			t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetBestMatch(t *testing.T) {
	candidates := map[string]struct{}{
		"cristiano ronaldo":  {},
		"lionel messi":       {},
		"jurgen klopp":       {},
		"r santa cruz":       {},
		"ange postecoglou":   {},
		"erik ten hag":       {},
	}

	tests := []struct {
		name    string
		wantOK  bool
		wantVal string
	}{
		{"Cristiano Ronaldo", true, "cristiano ronaldo"},
		{"Jürgen Klopp", true, "jurgen klopp"},
		{"R. Santa Cruz", true, "r santa cruz"},
		{"NonExistent Player", false, ""},
	}

	for _, tt := range tests {
		got, ok := getBestMatch(tt.name, candidates)
		if ok != tt.wantOK {
			t.Errorf("getBestMatch(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
		}
		if got != tt.wantVal {
			t.Errorf("getBestMatch(%q) = %q, want %q", tt.name, got, tt.wantVal)
		}
	}
}

func TestGetBestMatchRelaxed(t *testing.T) {
	// Unique relaxed match succeeds.
	uniq := map[string]struct{}{"dion drena beljo": {}, "luka modric": {}}
	if got, ok := getBestMatch("Dion Beljo", uniq); !ok || got != "dion drena beljo" {
		t.Errorf("relaxed unique = (%q,%v), want (dion drena beljo,true)", got, ok)
	}
	// Ambiguous relaxed match (two players share first+last) is rejected.
	ambig := map[string]struct{}{"dion drena beljo": {}, "dion mario beljo": {}}
	if got, ok := getBestMatch("Dion Beljo", ambig); ok {
		t.Errorf("relaxed ambiguous = (%q,true), want rejected", got)
	}
	// Exact match still wins over a relaxed candidate.
	mixed := map[string]struct{}{"dion beljo": {}, "dion drena beljo": {}}
	if got, ok := getBestMatch("Dion Beljo", mixed); !ok || got != "dion beljo" {
		t.Errorf("exact-over-relaxed = (%q,%v), want (dion beljo,true)", got, ok)
	}
}

func TestCalculateNameMatchScore(t *testing.T) {
	tests := []struct {
		target    string
		candidate string
		wantScore int
		wantOK    bool
	}{
		{"cristiano ronaldo", "cristiano ronaldo", 1000, true},
		{"ronaldo", "ronaldo", 1000, true},
		{"r ronaldo", "r ronaldo", 1000, true},
		{"cristiano ronaldo", "cristiano messi", 0, false},
		{"cristiano ronaldo", "messi", 0, false},
		// Non-exact matches
		{"r ronaldo", "roberto ronaldo", 51, true},
		{"cristiano ronaldo", "cristian ronaldo", 0, false},
		// Relaxed differing-part-count: middle name ignored, surname exact
		{"dion beljo", "dion drena beljo", scoreRelaxed, true},
		{"timi elsnik", "timi max elsnik", scoreRelaxed, true},
		// Differing part count but incompatible first name or surname: no match
		{"marko beljo", "dion drena beljo", 0, false},
		{"dion beljo", "dion drena vida", 0, false},
	}

	for _, tt := range tests {
		targetParts := splitFields(tt.target)
		score, ok := calculateNameMatchScore(targetParts, tt.target, tt.candidate)
		if ok != tt.wantOK || score != tt.wantScore {
			t.Errorf("calculateNameMatchScore(%q, %q) = (%d, %v), want (%d, %v)",
				tt.target, tt.candidate, score, ok, tt.wantScore, tt.wantOK)
		}
	}
}

func splitFields(s string) []string {
	var parts []string
	start := -1
	for i, c := range s {
		if c == ' ' {
			if start >= 0 {
				parts = append(parts, s[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		parts = append(parts, s[start:])
	}
	return parts
}
