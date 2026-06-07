package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			*s = append(*s, p)
		}
	}
	return nil
}

const uniformTexDir = "Asset/model/character/uniform/texture/#windx11/"

var (
	kitMapLineRe = regexp.MustCompile(`^\s*(\d+)\s*,\s*"([^"]+)"`)
	texKeyRe     = regexp.MustCompile(`(?i)^(KitFile|BackNumbersFile|ChestNumbersFile|LegNumbersFile|NameFontFile)=(.*)$`)
	texBaseRe    = regexp.MustCompile(`^u\d+[pg]\d+`)
)

// The game loads _srm/_name_ex (and some _leg) by naming convention without
// listing them in config.txt, so the pack must carry each KitFile's whole
// u<id><slot>* family, not only config-referenced textures.
func buildFamilyIndex(r *Reader) map[string][]string {
	prefix := normalizePath(uniformTexDir)
	fam := map[string][]string{}
	for _, f := range r.Files() {
		p := normalizePath(f.Path())
		if !strings.HasPrefix(p, prefix) || !strings.HasSuffix(p, ".ftex") {
			continue
		}
		full := f.Path()
		name := full[strings.LastIndexAny(full, "/\\")+1 : len(full)-len(".ftex")]
		if base := texBaseRe.FindString(strings.ToLower(name)); base != "" {
			fam[base] = append(fam[base], name)
		}
	}
	return fam
}

type kitMapEntry struct{ id, league, team, rel string }

func parseKitMap(path string) ([]kitMapEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []kitMapEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		m := kitMapLineRe.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		parts := strings.SplitN(m[2], `\`, 2)
		if len(parts) != 2 {
			continue
		}
		out = append(out, kitMapEntry{id: m[1], league: parts[0], team: parts[1], rel: m[2]})
	}
	return out, sc.Err()
}

func configTextures(path string) (names []string, kitFile string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		m := texKeyRe.FindStringSubmatch(strings.TrimRight(sc.Text(), "\r"))
		if m == nil {
			continue
		}
		v := strings.TrimRight(m[2], " \t")
		if v == "" {
			continue
		}
		names = append(names, v)
		if strings.EqualFold(m[1], "KitFile") {
			kitFile = v
		}
	}
	return names, kitFile, sc.Err()
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

func cmdKits(args []string) {
	fset := flag.NewFlagSet("kits", flag.ExitOnError)
	kservSrc := fset.String("kserv-src", "", "kitserver config tree: dir holding map.txt and <League>/<Team>/ folders")
	out := fset.String("out", "", "output pack directory")
	mapPath := fset.String("map", "", "map.txt path (default <kserv-src>/map.txt)")
	label := fset.String("label", "kits, kitserver format (textures embedded)", "first comment line of the pack map.txt")
	var leagues stringList
	fset.Var(&leagues, "leagues", "league names to include, comma-separated and/or repeatable; empty = all")
	if err := fset.Parse(args); err != nil {
		os.Exit(1)
	}
	if fset.NArg() != 1 || *kservSrc == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: cpk kits --kserv-src <dir> --out <dir> [--map <file>] [--leagues A,B ...] <uniform.cpk>")
		os.Exit(1)
	}

	mp := *mapPath
	if mp == "" {
		mp = filepath.Join(*kservSrc, "map.txt")
	}
	entries, err := parseKitMap(mp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map: %v\n", err)
		os.Exit(1)
	}
	r, err := Open(fset.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	want := map[string]bool{}
	for _, l := range leagues {
		want[strings.ToLower(l)] = true
	}
	if err := os.MkdirAll(*out, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	fam := buildFamilyIndex(r)

	var mapBuf strings.Builder
	fmt.Fprintf(&mapBuf, "# %s\n# team-id, \"League\\Team\"\n\n", *label)
	var teams, tex, missing int
	for _, e := range entries {
		if len(want) > 0 && !want[strings.ToLower(e.league)] {
			continue
		}
		srcTeam := filepath.Join(*kservSrc, e.league, e.team)
		if fi, err := os.Stat(srcTeam); err != nil || !fi.IsDir() {
			continue
		}
		dstTeam := filepath.Join(*out, e.league, e.team)
		if err := copyTree(srcTeam, dstTeam); err != nil {
			fmt.Fprintf(os.Stderr, "copy %s: %v\n", e.rel, err)
			os.Exit(1)
		}
		err := filepath.WalkDir(dstTeam, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || d.Name() != "config.txt" {
				return err
			}
			names, kitFile, err := configTextures(p)
			if err != nil {
				return err
			}
			want := append(names, fam[strings.ToLower(kitFile)]...)
			seen := map[string]bool{}
			for _, n := range want {
				key := strings.ToLower(n)
				if n == "" || seen[key] {
					continue
				}
				seen[key] = true
				f, ok := r.FindFile(uniformTexDir + n + ".ftex")
				if !ok {
					missing++
					continue
				}
				data, err := r.ReadFile(*f)
				if err != nil {
					missing++
					continue
				}
				if err := os.WriteFile(filepath.Join(filepath.Dir(p), n+".ftex"), data, 0644); err != nil {
					return err
				}
				tex++
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "textures %s: %v\n", e.rel, err)
			os.Exit(1)
		}
		fmt.Fprintf(&mapBuf, "%s, \"%s\"\n", e.id, e.rel)
		teams++
	}
	mapBuf.WriteString("\n# end of map\n")
	if err := os.WriteFile(filepath.Join(*out, "map.txt"), []byte(mapBuf.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write map.txt: %v\n", err)
		os.Exit(1)
	}
	if teams == 0 {
		fmt.Fprintln(os.Stderr, "kits: no teams matched (check --leagues against map.txt league names)")
		os.Exit(1)
	}
	if tex == 0 {
		fmt.Fprintf(os.Stderr, "kits: %d teams matched but 0 textures embedded; is %s the cpk that holds these teams' textures?\n", teams, fset.Arg(0))
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "built %d teams, %d textures embedded, %d referenced-but-absent\n", teams, tex, missing)
}
