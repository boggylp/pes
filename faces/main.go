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
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: faces <command> [args]

commands:
  detect [flags]  detect mismatched player faces in livecpk folder
  map    [flags]  map player faces between game versions`)
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
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *destCSV == "" || *sourceFolder == "" || *destFolder == "" {
		fmt.Fprintln(os.Stderr, "usage: faces map --destination-csv <file> --source-folder <path> --dest-folder <path> [--source-csv <file>]")
		fs.PrintDefaults()
		os.Exit(1)
	}

	mapFaces(*sourceCSV, *destCSV, *sourceFolder, *destFolder)
}
