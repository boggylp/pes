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
		folder := entry.Name()

		if !numericFolderRe.MatchString(folder) {
			issues = append(issues, checkNonNumeric(facesDir, folder, players))
			continue
		}

		baseID := suffixRe.ReplaceAllString(folder, "")
		isDuplicate := folder != baseID

		if _, ok := players[baseID]; !ok {
			issues = append(issues, Issue{
				Type:   "orphan",
				Folder: folder,
				BaseID: baseID,
			})
			continue
		}

		fpkPath := filepath.Join(facesDir, folder, "#Win", "face.fpk")
		if _, err := os.Stat(fpkPath); os.IsNotExist(err) {
			issues = append(issues, Issue{
				Type:         "no_fpk",
				Folder:       folder,
				BaseID:       baseID,
				FolderPlayer: players[baseID],
			})
			continue
		}

		embeddedID := extractFPKInternalID(fpkPath)
		if embeddedID != "" && embeddedID != baseID {
			issues = append(issues, Issue{
				Type:           "id_mismatch",
				Folder:         folder,
				BaseID:         baseID,
				FolderPlayer:   players[baseID],
				EmbeddedID:     embeddedID,
				EmbeddedPlayer: players[embeddedID],
				IsDuplicate:    isDuplicate,
			})
		}
	}

	return issues
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

	var mismatches, orphans, nonNumeric, noFPK []Issue
	for _, i := range issues {
		switch i.Type {
		case "id_mismatch":
			mismatches = append(mismatches, i)
		case "orphan":
			orphans = append(orphans, i)
		case "non_numeric":
			nonNumeric = append(nonNumeric, i)
		case "no_fpk":
			noFPK = append(noFPK, i)
		}
	}

	fmt.Printf("Found %d issue(s):\n\n", len(issues))

	if len(mismatches) > 0 {
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

	if len(orphans) > 0 {
		fmt.Printf("== ORPHAN FOLDERS (%d) ==\n", len(orphans))
		fmt.Println("Folder ID does not exist in the player CSV. Face won't be used by the game.")
		fmt.Println()
		for _, o := range orphans {
			fmt.Printf("  %s (ID %s not in CSV)\n", o.Folder, o.BaseID)
		}
		fmt.Println()
	}

	if len(nonNumeric) > 0 {
		fmt.Printf("== NON-NUMERIC FOLDERS (%d) ==\n", len(nonNumeric))
		fmt.Println("Folder name is not a valid player ID. The game won't load these.")
		fmt.Println()
		for _, n := range nonNumeric {
			info := ""
			if n.EmbeddedID != "" {
				name := ""
				if n.EmbeddedPlayer != "" {
					name = " (" + n.EmbeddedPlayer + ")"
				}
				info = fmt.Sprintf(" -> FPK references ID %s%s", n.EmbeddedID, name)
			} else if !n.HasFPK {
				info = " (no face.fpk)"
			}
			fmt.Printf("  %s%s\n", n.Folder, info)
		}
		fmt.Println()
	}

	if len(noFPK) > 0 {
		fmt.Printf("== MISSING FPK (%d) ==\n", len(noFPK))
		fmt.Println("Face folder exists but has no face.fpk file.")
		fmt.Println()
		for _, n := range noFPK {
			fmt.Printf("  %s (%s)\n", n.Folder, n.FolderPlayer)
		}
		fmt.Println()
	}
}
