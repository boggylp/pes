package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "detect":
		cmdDetect(os.Args[2:])
	case "map":
		cmdMap(os.Args[2:])
	case "relink":
		cmdRelink(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: faces <command> [args]

commands:
  detect [flags]  detect mismatched player faces in livecpk folder
  map    [flags]  map player faces between game versions
  relink [flags]  rewrite a face folder's embedded ID to a new ID (length-changing)`)
}

func cmdRelink(args []string) {
	fs := flag.NewFlagSet("relink", flag.ExitOnError)
	folder := fs.String("folder", "", "face folder containing #Win/*.fpk|*.fpkd packages (required)")
	newID := fs.String("id", "", "new player ID to embed (required)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if *folder == "" || *newID == "" {
		fmt.Fprintln(os.Stderr, "usage: faces relink --folder <faceDir> --id <newID>")
		fs.PrintDefaults()
		os.Exit(1)
	}
	if err := relinkFaceFolder(*folder, *newID); err != nil {
		fmt.Fprintf(os.Stderr, "relink failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("relinked %s -> id %s\n", *folder, *newID)
}

func cmdDetect(args []string) {
	fs := flag.NewFlagSet("detect", flag.ExitOnError)
	facesDir := fs.String("faces-dir", "", "path to the livecpk faces directory (e.g. .../face/real/)")
	playerCSV := fs.String("player-csv", "", "CSV file with player IDs and names (Id;Name;... format)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *facesDir == "" || *playerCSV == "" {
		fmt.Fprintln(os.Stderr, "usage: faces detect --faces-dir <path> --player-csv <file>")
		fs.PrintDefaults()
		os.Exit(1)
	}

	issues := detect(*facesDir, *playerCSV)
	printReport(issues)

	if len(issues) > 0 {
		os.Exit(1)
	}
}

func cmdMap(args []string) {
	fs := flag.NewFlagSet("map", flag.ExitOnError)
	sourceCSV := fs.String("source-csv", "", "source CSV file with player names and IDs (optional, uses folder names if not provided)")
	destCSV := fs.String("destination-csv", "", "destination CSV file with player names and IDs (required)")
	sourceFolder := fs.String("source-folder", "", "source folder containing player face directories (required)")
	destFolder := fs.String("dest-folder", "", "destination folder for mapped player faces (required)")
	skipExisting := fs.Bool("skip-existing", false, "skip mapping when destination ID folder already exists (do not overwrite)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *destCSV == "" || *sourceFolder == "" || *destFolder == "" {
		fmt.Fprintln(os.Stderr, "usage: faces map --destination-csv <file> --source-folder <path> --dest-folder <path> [--source-csv <file>] [--skip-existing]")
		fs.PrintDefaults()
		os.Exit(1)
	}

	mapFaces(*sourceCSV, *destCSV, *sourceFolder, *destFolder, *skipExisting)
}
