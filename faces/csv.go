package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Player struct {
	ID   string
	Name string
}

// loadPlayerCSV reads a semicolon-delimited CSV with at least Id and Name columns.
// Returns a map of ID -> name.
func loadPlayerCSV(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("opening CSV %s: %v", path, err)
	}
	defer f.Close()

	// Skip UTF-8 BOM if present
	r := skipBOM(f)

	cr := csv.NewReader(r)
	cr.Comma = ';'
	cr.LazyQuotes = true

	header, err := cr.Read()
	if err != nil {
		log.Fatalf("reading CSV header: %v", err)
	}

	idIdx, nameIdx := -1, -1
	for i, col := range header {
		switch strings.TrimSpace(col) {
		case "Id":
			idIdx = i
		case "Name":
			nameIdx = i
		}
	}
	if idIdx < 0 || nameIdx < 0 {
		log.Fatalf("CSV must have Id and Name columns, found: %v", header)
	}

	players := make(map[string]string)
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("reading CSV row: %v", err)
		}
		if idIdx < len(row) && nameIdx < len(row) {
			players[strings.TrimSpace(row[idIdx])] = strings.TrimSpace(row[nameIdx])
		}
	}

	return players
}

// loadPlayerList reads a CSV and returns a slice of Players (preserves order).
func loadPlayerList(path string) []Player {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("opening CSV %s: %v", path, err)
	}
	defer f.Close()

	r := skipBOM(f)

	cr := csv.NewReader(r)
	cr.Comma = ';'
	cr.LazyQuotes = true

	header, err := cr.Read()
	if err != nil {
		log.Fatalf("reading CSV header: %v", err)
	}

	idIdx, nameIdx := -1, -1
	for i, col := range header {
		switch strings.TrimSpace(col) {
		case "Id":
			idIdx = i
		case "Name":
			nameIdx = i
		}
	}
	if idIdx < 0 || nameIdx < 0 {
		log.Fatalf("CSV must have Id and Name columns, found: %v", header)
	}

	var players []Player
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("reading CSV row: %v", err)
		}
		if idIdx < len(row) && nameIdx < len(row) {
			players = append(players, Player{
				ID:   strings.TrimSpace(row[idIdx]),
				Name: strings.TrimSpace(row[nameIdx]),
			})
		}
	}

	return players
}

// loadPlayersFromFolders reads folder names as player data.
// Each folder should contain exactly one subdirectory whose name is the player ID.
func loadPlayersFromFolders(folderPath string) []Player {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		log.Fatalf("reading folder %s: %v", folderPath, err)
	}

	var players []Player
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		playerName := entry.Name()
		subdirs, err := os.ReadDir(filepath.Join(folderPath, playerName))
		if err != nil {
			log.Printf("warning: reading subdirs of %s: %v", playerName, err)
			continue
		}

		var dirs []string
		for _, sd := range subdirs {
			if sd.IsDir() {
				dirs = append(dirs, sd.Name())
			}
		}

		if len(dirs) == 1 {
			players = append(players, Player{ID: dirs[0], Name: playerName})
		} else {
			log.Printf("warning: skipping '%s': expected 1 subdirectory, found %d", playerName, len(dirs))
		}
	}

	return players
}

// skipBOM returns a reader that skips a UTF-8 BOM if present.
func skipBOM(f *os.File) io.Reader {
	bom := make([]byte, 3)
	n, _ := f.Read(bom)
	if n >= 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		return f // BOM consumed, continue from current position
	}
	// No BOM, seek back to start
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.Fatalf("seeking file: %v", err)
	}
	return f
}

func lookupPlayer(players map[string]string, id string) string {
	if name, ok := players[id]; ok {
		return fmt.Sprintf("%s (%s)", name, id)
	}
	return fmt.Sprintf("unknown (%s)", id)
}
