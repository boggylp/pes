package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	Delimiter = ';'
	Encoding  = "utf-8"
)

type PlayerMapping struct {
	SrcPlayerID  string
	DestPlayerID string
}

func readCSV(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = Delimiter
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	// Skip header row
	for i := 1; i < len(records); i++ {
		if len(records[i]) >= 2 {
			playerID := records[i][0]
			playerName := records[i][1]
			data[playerName] = playerID
		}
	}
	return data, nil
}

func normalize(fullName string) string {
	// NFD normalization
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, fullName)

	// Keep only alphanumeric and spaces
	var builder strings.Builder
	for _, char := range result {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || unicode.IsSpace(char) {
			builder.WriteRune(char)
		}
	}

	return strings.ToLower(builder.String())
}

func calculateNameMatchScore(name1, name2 string) *int {
	parts1 := strings.Fields(normalize(name1))
	parts2 := strings.Fields(normalize(name2))

	if len(parts1) != len(parts2) {
		return nil
	}

	if len(parts1) == 1 {
		if parts1[0] == parts2[0] {
			score := 100
			return &score
		}
		return nil
	}

	// Surname must match exactly
	if parts1[len(parts1)-1] != parts2[len(parts2)-1] {
		return nil
	}

	score := 0

	// Check all first names
	for i := 0; i < len(parts1)-1; i++ {
		p1 := parts1[i]
		p2 := parts2[i]

		if len(p1) == 1 || len(p2) == 1 {
			// Initial match
			if p1[0] != p2[0] {
				return nil
			}
			score += 1
		} else {
			// Full name match
			if p1 != p2 {
				return nil
			}
			score += 10
		}
	}

	score += 50 // Bonus for surname match
	return &score
}

func getBestMatch(targetName string, candidateNames []string) string {
	var bestMatch string
	bestScore := -1

	for _, candidate := range candidateNames {
		score := calculateNameMatchScore(targetName, candidate)
		if score != nil && *score > bestScore {
			bestScore = *score
			bestMatch = candidate
		}
	}

	return bestMatch
}

func getPlayerMapping(sourceCSV, destinationCSV string) ([]PlayerMapping, error) {
	sourceData, err := readCSV(sourceCSV)
	if err != nil {
		return nil, err
	}

	destinationData, err := readCSV(destinationCSV)
	if err != nil {
		return nil, err
	}

	var playerMapping []PlayerMapping
	destinationNames := make([]string, 0, len(destinationData))
	for name := range destinationData {
		destinationNames = append(destinationNames, name)
	}

	for playerName, srcID := range sourceData {
		candidateName := getBestMatch(playerName, destinationNames)
		if candidateName != "" {
			playerMapping = append(playerMapping, PlayerMapping{
				SrcPlayerID:  srcID,
				DestPlayerID: destinationData[candidateName],
			})
		}
	}

	return playerMapping, nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func hexReplace(filePath, oldID, newID string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	data = bytes.ReplaceAll(data, []byte(oldID), []byte(newID))

	return os.WriteFile(filePath, data, 0644)
}

func updateFacesStructure(srcFolderPath, destFolderPath string, mapping []PlayerMapping) error {
	facePath := "Asset/model/character/face/real"

	for _, item := range mapping {
		srcPath := filepath.Join(srcFolderPath, facePath, item.SrcPlayerID)
		destPath := filepath.Join(destFolderPath, facePath, item.DestPlayerID)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		if err := os.MkdirAll(destPath, 0755); err != nil {
			return err
		}

		if err := copyDir(srcPath, destPath); err != nil {
			return err
		}

		fpkPath := filepath.Join(destPath, "#Win", "face.fpk")
		if err := hexReplace(fpkPath, item.SrcPlayerID, item.DestPlayerID); err != nil {
			fmt.Printf("Warning: Could not replace hex in %s: %v\n", fpkPath, err)
		}
	}

	return nil
}

func main() {
	sourceCSV := "samples/BPB-2023-players.csv"
	destinationCSV := "samples/FL26_players.csv"
	srcFolderPath := "samples"
	destFolderPath := "livecpk/root"

	mapping, err := getPlayerMapping(sourceCSV, destinationCSV)
	if err != nil {
		fmt.Printf("Error getting player mapping: %v\n", err)
		os.Exit(1)
	}

	resultPath := filepath.Join("result", destFolderPath)
	if err := updateFacesStructure(srcFolderPath, resultPath, mapping); err != nil {
		fmt.Printf("Error updating faces structure: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Finished processing")
}
