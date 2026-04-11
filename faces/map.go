package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PlayerMapping struct {
	SrcPlayerID   string
	SrcPlayerName string
	DestPlayerID  string
}

func mapFaces(sourceCSV, destCSV, sourceFolder, destFolder string) {
	log.Println("=== Starting Player Face Mapping Tool ===")

	var sourceData []Player
	if sourceCSV != "" {
		if _, err := os.Stat(sourceCSV); err == nil {
			log.Printf("Using CSV mode - reading from %s", sourceCSV)
			sourceData = loadPlayerList(sourceCSV)
		}
	}
	if sourceData == nil {
		log.Printf("Using folder mode - reading folder names from %s", sourceFolder)
		sourceData = loadPlayersFromFolders(sourceFolder)
	}

	destData := loadPlayerList(destCSV)

	mapping := getPlayerMapping(sourceData, destData)
	updateFacesStructure(sourceFolder, destFolder, mapping)

	log.Println("=== Finished processing successfully ===")
}

func getPlayerMapping(sourceData, destData []Player) []PlayerMapping {
	log.Println("Starting player mapping process")

	// Pre-normalize all destination names
	normalizedToOriginal := make(map[string]Player)
	available := make(map[string]struct{})
	for _, p := range destData {
		n := normalize(p.Name)
		normalizedToOriginal[n] = p
		available[n] = struct{}{}
	}
	log.Printf("Pre-normalized %d destination players", len(normalizedToOriginal))

	var mapping []PlayerMapping
	for _, player := range sourceData {
		if len(available) == 0 {
			break
		}
		matched, ok := getBestMatch(player.Name, available)
		if ok {
			dest := normalizedToOriginal[matched]
			mapping = append(mapping, PlayerMapping{
				SrcPlayerID:   player.ID,
				SrcPlayerName: player.Name,
				DestPlayerID:  dest.ID,
			})
			delete(available, matched)
		} else {
			log.Printf("No match found for player: %s", player.Name)
		}
	}

	log.Printf("Successfully mapped %d players", len(mapping))
	return mapping
}

func updateFacesStructure(srcFolder, destFolder string, mapping []PlayerMapping) {
	log.Printf("Updating faces structure from %s to %s", srcFolder, destFolder)
	processed := 0

	for _, item := range mapping {
		// ID length must match for hex replacement to work
		if len(item.SrcPlayerID) != len(item.DestPlayerID) {
			continue
		}

		pathDirect := filepath.Join(srcFolder, item.SrcPlayerID)
		pathNested := filepath.Join(srcFolder, item.SrcPlayerName, item.SrcPlayerID)
		srcPath := pathDirect
		if _, err := os.Stat(pathDirect); os.IsNotExist(err) {
			srcPath = pathNested
		}

		destPath := filepath.Join(destFolder, item.DestPlayerID)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			log.Printf("Source path does not exist: %s", srcPath)
			continue
		}

		if err := os.MkdirAll(destPath, 0755); err != nil {
			log.Printf("Error creating directory %s: %v", destPath, err)
			continue
		}

		if err := copyDir(srcPath, destPath); err != nil {
			log.Printf("Error copying %s to %s: %v", srcPath, destPath, err)
			continue
		}

		fpkPath := filepath.Join(destPath, "#Win", "face.fpk")
		hexReplace(fpkPath, item.SrcPlayerID, item.DestPlayerID)
		processed++
	}

	log.Printf("Successfully processed %d player faces", processed)
}

func hexReplace(filePath, oldID, newID string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File does not exist, skipping hex replace: %s", filePath)
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading %s: %v", filePath, err)
		return
	}

	data = []byte(strings.ReplaceAll(string(data), oldID, newID))

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Printf("Error writing %s: %v", filePath, err)
		return
	}
	log.Printf("Hex replaced in %s: %s -> %s", filePath, oldID, newID)
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}
