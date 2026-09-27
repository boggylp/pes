package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Issue struct {
	Type           string // "id_mismatch", "orphan", "non_numeric", "no_fpk"
	Folder         string
	BaseID         string
	FolderPlayer   string
	EmbeddedID     string
	EmbeddedPlayer string
	HasFPK         bool
	IsDuplicate    bool
}

var (
	numericFolderRe = regexp.MustCompile(`^\d+(_\d+)?$`)
	fpkPathRe       = regexp.MustCompile(`face/real/(\d+)`)
	suffixRe        = regexp.MustCompile(`_\d+$`)
)

func detect(facesDir, playerCSV string) []Issue {
	players := loadPlayerCSV(playerCSV)
	log.Printf("Loaded %d players from CSV", len(players))
	log.Printf("Scanning %s", facesDir)

	entries, err := os.ReadDir(facesDir)
	if err != nil {
		log.Fatalf("reading faces directory: %v", err)
	}

	var issues []Issue
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if issue, ok := checkFolder(facesDir, entry.Name(), players); ok {
			issues = append(issues, issue)
		}
	}
	return issues
}

func checkFolder(facesDir, folder string, players map[string]string) (Issue, bool) {
	if !numericFolderRe.MatchString(folder) {
		return checkNonNumeric(facesDir, folder, players), true
	}

	baseID := suffixRe.ReplaceAllString(folder, "")
	if _, ok := players[baseID]; !ok {
		return Issue{Type: "orphan", Folder: folder, BaseID: baseID}, true
	}

	fpkPath := filepath.Join(facesDir, folder, "#Win", "face.fpk")
	if _, err := os.Stat(fpkPath); os.IsNotExist(err) {
		return Issue{Type: "no_fpk", Folder: folder, BaseID: baseID, FolderPlayer: players[baseID]}, true
	}

	embeddedID := extractFPKInternalID(fpkPath)
	if embeddedID == "" || embeddedID == baseID {
		return Issue{}, false
	}
	return Issue{
		Type:           "id_mismatch",
		Folder:         folder,
		BaseID:         baseID,
		FolderPlayer:   players[baseID],
		EmbeddedID:     embeddedID,
		EmbeddedPlayer: players[embeddedID],
		IsDuplicate:    folder != baseID,
	}, true
}

func checkNonNumeric(facesDir, folder string, players map[string]string) Issue {
	fpkPath := filepath.Join(facesDir, folder, "#Win", "face.fpk")
	hasFPK := false
	if _, err := os.Stat(fpkPath); err == nil {
		hasFPK = true
	}

	var embeddedID, playerName string
	if hasFPK {
		embeddedID = extractFPKInternalID(fpkPath)
		if embeddedID != "" {
			playerName = players[embeddedID]
		}
	}

	return Issue{
		Type:           "non_numeric",
		Folder:         folder,
		EmbeddedID:     embeddedID,
		EmbeddedPlayer: playerName,
		HasFPK:         hasFPK,
	}
}

func extractFPKInternalID(fpkPath string) string {
	data, err := os.ReadFile(fpkPath)
	if err != nil {
		log.Printf("warning: reading %s: %v", fpkPath, err)
		return ""
	}

	// The FPK contains ASCII paths like "face/real/65342/sourceimages/"
	// Search for the pattern in the raw bytes (ASCII subset)
	text := toASCII(data)
	match := fpkPathRe.FindStringSubmatch(text)
	if match != nil {
		return match[1]
	}
	return ""
}

// toASCII converts bytes to a string, replacing non-printable/non-ASCII with spaces.
func toASCII(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))
	for _, c := range data {
		if c >= 32 && c <= 126 {
			b.WriteByte(c)
		} else {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func printReport(issues []Issue) {
	if len(issues) == 0 {
		fmt.Println("No issues found. All faces are correctly mapped.")
		return
	}

	byType := map[string][]Issue{}
	for _, i := range issues {
		byType[i.Type] = append(byType[i.Type], i)
	}

	fmt.Printf("Found %d issue(s):\n\n", len(issues))
	printMismatches(byType["id_mismatch"])
	printIssueSection(byType["orphan"], "ORPHAN FOLDERS",
		"Folder ID does not exist in the player CSV. Face won't be used by the game.",
		func(o Issue) string { return fmt.Sprintf("%s (ID %s not in CSV)", o.Folder, o.BaseID) })
	printIssueSection(byType["non_numeric"], "NON-NUMERIC FOLDERS",
		"Folder name is not a valid player ID. The game won't load these.",
		func(n Issue) string { return n.Folder + describeNonNumeric(n) })
	printIssueSection(byType["no_fpk"], "MISSING FPK",
		"Face folder exists but has no face.fpk file.",
		func(n Issue) string { return fmt.Sprintf("%s (%s)", n.Folder, n.FolderPlayer) })
}

func printMismatches(mismatches []Issue) {
	if len(mismatches) == 0 {
		return
	}
	fmt.Printf("== FPK ID MISMATCHES (%d) ==\n", len(mismatches))
	fmt.Println("Face folder says one player, but the FPK file internally references another.")
	fmt.Println()
	for _, m := range mismatches {
		dup := ""
		if m.IsDuplicate {
			dup = " (duplicate folder)"
		}
		fmt.Printf("  %s%s\n", m.Folder, dup)
		fmt.Printf("    Folder expects: %s\n", lookupPlayer(map[string]string{m.BaseID: m.FolderPlayer}, m.BaseID))
		embPlayers := map[string]string{}
		if m.EmbeddedPlayer != "" {
			embPlayers[m.EmbeddedID] = m.EmbeddedPlayer
		}
		fmt.Printf("    FPK contains:   %s\n", lookupPlayer(embPlayers, m.EmbeddedID))
		fmt.Println()
	}
}

func printIssueSection(issues []Issue, title, blurb string, line func(Issue) string) {
	if len(issues) == 0 {
		return
	}
	fmt.Printf("== %s (%d) ==\n", title, len(issues))
	fmt.Println(blurb)
	fmt.Println()
	for _, i := range issues {
		fmt.Printf("  %s\n", line(i))
	}
	fmt.Println()
}

func describeNonNumeric(n Issue) string {
	if n.EmbeddedID == "" {
		if !n.HasFPK {
			return " (no face.fpk)"
		}
		return ""
	}
	name := ""
	if n.EmbeddedPlayer != "" {
		name = " (" + n.EmbeddedPlayer + ")"
	}
	return fmt.Sprintf(" -> FPK references ID %s%s", n.EmbeddedID, name)
}
