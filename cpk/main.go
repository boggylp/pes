package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "list":
		cmdList(os.Args[2:])
	case "extract":
		cmdExtract(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: cpk <command> [flags]

commands:
  list    [flags] CPK            print the table of contents
  extract [flags] CPK            extract one file or all files from a CPK`)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	long := fs.Bool("l", false, "long output: include offset and sizes")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: cpk list [-l] <cpk>")
		os.Exit(1)
	}

	r, err := Open(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	for _, f := range r.Files() {
		if *long {
			fmt.Printf("%-12d %-12d %#016x %s\n", f.Size, f.ExtractSize, f.Offset, f.Path())
		} else {
			fmt.Println(f.Path())
		}
	}
}

func cmdExtract(args []string) {
	fs := flag.NewFlagSet("extract", flag.ExitOnError)
	inner := fs.String("file", "", "inner path of a single file to extract (e.g. common/etc/pesdb/Player.bin)")
	out := fs.String("out", "", "output path (file when --file is given, directory for full extraction)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: cpk extract [--file inner-path] [--out path] <cpk>")
		os.Exit(1)
	}

	r, err := Open(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	if *inner != "" {
		f, ok := r.FindFile(*inner)
		if !ok {
			fmt.Fprintf(os.Stderr, "file %q not found in CPK\n", *inner)
			os.Exit(1)
		}
		data, err := r.ReadFile(*f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read: %v\n", err)
			os.Exit(1)
		}
		dest := *out
		if dest == "" {
			dest = filepath.Base(f.Name)
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", dest, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", dest, len(data))
		return
	}

	dir := *out
	if dir == "" {
		fmt.Fprintln(os.Stderr, "extract: --out <dir> required for full extraction")
		os.Exit(1)
	}
	for _, f := range r.Files() {
		data, err := r.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Path(), err)
			continue
		}
		dest := filepath.Join(dir, filepath.FromSlash(f.Path()))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", filepath.Dir(dest), err)
			continue
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", dest, err)
			continue
		}
	}
}
