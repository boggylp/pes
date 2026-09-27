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

	bestMatch, bestScore, bestCount := pickBestCandidate(targetParts, targetNorm, candidates)
	if bestMatch == "" {
		return "", false
	}
	// A relaxed (differing part-count) match is only trusted when unambiguous:
	// two players sharing first+last name must not be silently conflated.
	if bestScore < scoreSamePartCount && bestCount > 1 {
		return "", false
	}
	return bestMatch, true
}

func pickBestCandidate(targetParts []string, targetNorm string, candidates map[string]struct{}) (string, int, int) {
	bestMatch := ""
	bestScore := -1
	bestCount := 0
	for candidate := range candidates {
		if !keepCandidate(targetParts, targetNorm, candidate) {
			continue
		}
		score, ok := calculateNameMatchScore(targetParts, targetNorm, candidate)
		if !ok {
			continue
		}
		switch {
		case score > bestScore:
			bestScore = score
			bestMatch = candidate
			bestCount = 1
		case score == bestScore:
			bestCount++
		}
	}
	return bestMatch, bestScore, bestCount
}

func keepCandidate(targetParts []string, targetNorm, candidate string) bool {
	if abs(len(candidate)-len(targetNorm)) > 10 {
		return false
	}
	// Surname must start with same character
	candidateParts := strings.Fields(candidate)
	if len(candidateParts) == 0 {
		return false
	}
	targetSurname := targetParts[len(targetParts)-1]
	candidateSurname := candidateParts[len(candidateParts)-1]
	return targetSurname[0] == candidateSurname[0]
}

// Score tiers, ascending. Exact-name and equal-part-count matches outrank a
// relaxed first+last match, so the latter only wins when nothing better exists.
const (
	scoreRelaxed       = 30 // differing part count: surname exact + first name compatible
	scoreSamePartCount = 50 // base for an equal-part-count, surname-exact match
)

func calculateNameMatchScore(targetParts []string, targetNorm, candidateNorm string) (int, bool) {
	if targetNorm == candidateNorm {
		return 1000, true
	}
	candidateParts := strings.Fields(candidateNorm)
	if len(targetParts) != len(candidateParts) {
		return scoreRelaxedMatch(targetParts, candidateParts)
	}
	return scoreSamePartCountMatch(targetParts, candidateParts)
}

// scoreRelaxedMatch keeps a middle name on one side (e.g. "Dion Drena Beljo"
// vs "Dion Beljo") from blocking the match. It requires the surname to match
// exactly and the first name to be compatible; the ambiguity guard in
// getBestMatch rejects first+last collisions.
func scoreRelaxedMatch(targetParts, candidateParts []string) (int, bool) {
	if len(targetParts) < 2 || len(candidateParts) < 2 {
		return 0, false
	}
	if targetParts[len(targetParts)-1] != candidateParts[len(candidateParts)-1] {
		return 0, false
	}
	if !firstNameCompatible(targetParts[0], candidateParts[0]) {
		return 0, false
	}
	return scoreRelaxed, true
}

func scoreSamePartCountMatch(targetParts, candidateParts []string) (int, bool) {
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
		partScore, ok := scoreGivenNamePart(targetParts[i], candidateParts[i])
		if !ok {
			return 0, false
		}
		score += partScore
	}
	return score + scoreSamePartCount, true
}

func scoreGivenNamePart(p1, p2 string) (int, bool) {
	if len(p1) == 1 || len(p2) == 1 {
		return 1, p1[0] == p2[0]
	}
	return 10, p1 == p2
}

// firstNameCompatible reports whether two first names plausibly denote the same
// person: equal, a shared initial when either is abbreviated, or one a prefix of
// the other. Used only by the relaxed differing-part-count path.
func firstNameCompatible(a, b string) bool {
	if a == b {
		return true
	}
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 1 || len(rb) == 1 {
		return ra[0] == rb[0] // initial match, rune-safe for non-ASCII first names
	}
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
