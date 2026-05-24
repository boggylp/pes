// Package main implements editsave, a wrapper around zlac/ejogc327's
// decrypter21.exe and encrypter21.exe that decrypts and re-encrypts PES 2021
// / Football Life 2026 EDIT00000000 saves, plus a round-trip verifier.
//
// The encryption itself is Konami's; the tools are external. This binary
// owns the orchestration: temp directories, argument quoting, structural
// integrity checks. It does NOT modify save contents.
package main

import (
	"fmt"
	"os"
)

const usage = `usage: editsave <command> [flags]

commands:
  decrypt    --tools-dir DIR --out DIR INPUT            wrap decrypter21.exe
  encrypt    --tools-dir DIR --out FILE INPUT_DIR       wrap encrypter21.exe
  roundtrip  --tools-dir DIR INPUT                      decrypt then re-encrypt, compare
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "decrypt":
		os.Exit(cmdDecrypt(os.Args[2:]))
	case "encrypt":
		os.Exit(cmdEncrypt(os.Args[2:]))
	case "roundtrip":
		os.Exit(cmdRoundtrip(os.Args[2:]))
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}
