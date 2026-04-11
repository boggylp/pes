package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// normalize removes accents, non-alphanumeric characters (keeping spaces), and casefolds.
func normalize(name string) string {
	// NFD decomposition: separates base characters from combining marks
	nfd := norm.NFD.String(name)

	var b strings.Builder
	for _, r := range nfd {
		if unicode.Is(unicode.Mn, r) {
			continue // skip combining marks (accents)
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}

	return strings.ToLower(b.String())
}

// getBestMatch finds the best matching normalized name from candidates.
// Returns the matched normalized name and true, or empty string and false.
func getBestMatch(targetName string, candidates map[string]struct{}) (string, bool) {
	targetNorm := normalize(targetName)

	// Quick exact match
	if _, ok := candidates[targetNorm]; ok {
		return targetNorm, true
	}

	targetParts := strings.Fields(targetNorm)
	if len(targetParts) == 0 {
		return "", false
	}
	targetSurname := targetParts[len(targetParts)-1]

	bestMatch := ""
	bestScore := -1

	for candidate := range candidates {
		if abs(len(candidate)-len(targetNorm)) > 10 {
			continue
		}

		// Surname must start with same character
		candidateParts := strings.Fields(candidate)
		if len(candidateParts) == 0 {
			continue
		}
		candidateSurname := candidateParts[len(candidateParts)-1]
		if targetSurname[0] != candidateSurname[0] {
			continue
		}

		score, ok := calculateNameMatchScore(targetParts, targetNorm, candidate)
		if ok && score > bestScore {
			bestScore = score
			bestMatch = candidate
		}
	}

	if bestMatch == "" {
		return "", false
	}
	return bestMatch, true
}

func calculateNameMatchScore(targetParts []string, targetNorm, candidateNorm string) (int, bool) {
	if targetNorm == candidateNorm {
		return 1000, true
	}

	candidateParts := strings.Fields(candidateNorm)

	if len(targetParts) != len(candidateParts) {
		return 0, false
	}

	if len(targetParts) == 1 {
		if targetParts[0] == candidateParts[0] {
			return 100, true
		}
		return 0, false
	}

	// Surname must match exactly
	if targetParts[len(targetParts)-1] != candidateParts[len(candidateParts)-1] {
		return 0, false
	}

	score := 0
	for i := 0; i < len(targetParts)-1; i++ {
		p1 := targetParts[i]
		p2 := candidateParts[i]

		if len(p1) == 1 || len(p2) == 1 {
			if p1[0] != p2[0] {
				return 0, false
			}
			score++
		} else {
			if p1 != p2 {
				return 0, false
			}
			score += 10
		}
	}

	return score + 50, true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
