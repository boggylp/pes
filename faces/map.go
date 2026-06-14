package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

type PlayerMapping struct {
	SrcPlayerID   string
	SrcPlayerName string
	DestPlayerID  string
}

func mapFaces(sourceCSV, destCSV, sourceFolder, destFolder string, skipExisting bool) {
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
	updateFacesStructure(sourceFolder, destFolder, mapping, skipExisting)

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

func updateFacesStructure(srcFolder, destFolder string, mapping []PlayerMapping, skipExisting bool) {
	log.Printf("Updating faces structure from %s to %s", srcFolder, destFolder)

	var processed, relinked, skipped, missingSrc, failed int64

	// Buffered channel as a work queue; workers consume items in parallel.
	// Disk I/O dominates each iteration, so 2× CPU is a reasonable default.
	workers := runtime.NumCPU() * 2
	if workers < 4 {
		workers = 4
	}
	ch := make(chan PlayerMapping, workers*2)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				switch processOne(srcFolder, destFolder, item, skipExisting) {
				case resultProcessed:
					atomic.AddInt64(&processed, 1)
				case resultRelinked:
					atomic.AddInt64(&relinked, 1)
				case resultSkippedExisting:
					atomic.AddInt64(&skipped, 1)
				case resultMissingSrc:
					atomic.AddInt64(&missingSrc, 1)
				case resultError:
					atomic.AddInt64(&failed, 1)
				}
			}
		}()
	}

	for _, item := range mapping {
		ch <- item
	}
	close(ch)
	wg.Wait()

	log.Printf("Installed %d faces (%d direct, %d relinked); skipped %d existing, %d source-not-found, %d errors",
		processed+relinked, processed, relinked, skipped, missingSrc, failed)
}

type processResult int

const (
	resultProcessed processResult = iota
	resultRelinked
	resultSkippedExisting
	resultMissingSrc
	resultError
)

func processOne(srcFolder, destFolder string, item PlayerMapping, skipExisting bool) processResult {
	pathDirect := filepath.Join(srcFolder, item.SrcPlayerID)
	pathNested := filepath.Join(srcFolder, item.SrcPlayerName, item.SrcPlayerID)
	srcPath := pathDirect
	if _, err := os.Stat(pathDirect); os.IsNotExist(err) {
		srcPath = pathNested
	}

	destPath := filepath.Join(destFolder, item.DestPlayerID)

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		// Common when the source CSV is much larger than the source folder
		// (e.g. running an identity install with the full game CSV). Counted
		// rather than logged per-row to keep output usable for those cases.
		return resultMissingSrc
	}

	if skipExisting {
		if _, err := os.Stat(destPath); err == nil {
			return resultSkippedExisting
		}
	}

	if err := os.MkdirAll(destPath, 0o755); err != nil {
		log.Printf("Error creating directory %s: %v", destPath, err)
		return resultError
	}

	if err := copyDir(srcPath, destPath); err != nil {
		log.Printf("Error copying %s to %s: %v", srcPath, destPath, err)
		return resultError
	}

	if len(item.SrcPlayerID) == len(item.DestPlayerID) {
		fpkPath := filepath.Join(destPath, "#Win", "face.fpk")
		if err := hexReplace(fpkPath, item.SrcPlayerID, item.DestPlayerID); err != nil {
			log.Printf("Error hex-replacing %s: %v", fpkPath, err)
			return resultError
		}
		return resultProcessed
	}
	// Different-length IDs need an FPK repack, not an in-place byte swap.
	switch err := relinkFaceFolder(destPath, item.DestPlayerID); {
	case err == nil:
		return resultRelinked
	case errors.Is(err, errNoFpk), errors.Is(err, errNoEmbeddedID):
		return resultProcessed // copied; no embedded id to rewrite
	default:
		log.Printf("Error relinking %s: %v", destPath, err)
		return resultError
	}
}

// hexReplace rewrites every literal occurrence of oldID with newID inside the
// file at filePath. The two IDs must have equal byte length (the caller is
// responsible) so the file size stays the same and the FPK's length-prefixed
// path table is not corrupted.
func hexReplace(filePath, oldID, newID string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Some source folders ship without a face.fpk (textures only); skip silently.
		return nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if !bytes.Contains(data, []byte(oldID)) {
		return nil
	}
	data = bytes.ReplaceAll(data, []byte(oldID), []byte(newID))
	return os.WriteFile(filePath, data, 0o644)
}

// copyDir mirrors src into dst recursively. Files are streamed via io.Copy so
// large textures don't have to fit fully in memory.
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
