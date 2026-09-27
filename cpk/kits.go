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
	defer func() { _ = f.Close() }()
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
	defer func() { _ = f.Close() }()
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

type kitsOptions struct {
	kservSrc string
	out      string
	mapPath  string
	label    string
	leagues  []string
	cpkPath  string
}

func parseKitsFlags(args []string) kitsOptions {
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
	return kitsOptions{kservSrc: *kservSrc, out: *out, mapPath: mp, label: *label, leagues: leagues, cpkPath: fset.Arg(0)}
}

func cmdKits(args []string) {
	opts := parseKitsFlags(args)
	entries, err := parseKitMap(opts.mapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map: %v\n", err)
		os.Exit(1)
	}
	r, err := Open(opts.cpkPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = r.Close() }()

	if err := os.MkdirAll(opts.out, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	pack := kitPack{r: r, fam: buildFamilyIndex(r)}
	result, err := pack.packTeams(entries, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(opts.out, "map.txt"), []byte(result.mapText), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write map.txt: %v\n", err)
		os.Exit(1)
	}
	reportKits(opts, result)
}

func reportKits(opts kitsOptions, result kitPackResult) {
	if result.teams == 0 {
		fmt.Fprintln(os.Stderr, "kits: no teams matched (check --leagues against map.txt league names)")
		os.Exit(1)
	}
	if result.textures.embedded == 0 {
		fmt.Fprintf(os.Stderr, "kits: %d teams matched but 0 textures embedded; is %s the cpk that holds these teams' textures?\n", result.teams, opts.cpkPath)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "built %d teams, %d textures embedded, %d referenced-but-absent\n", result.teams, result.textures.embedded, result.textures.missing)
}

func keepLeague(leagues []string) func(kitMapEntry) bool {
	want := map[string]bool{}
	for _, l := range leagues {
		want[strings.ToLower(l)] = true
	}
	return func(e kitMapEntry) bool {
		return len(want) == 0 || want[strings.ToLower(e.league)]
	}
}

type textureCounts struct {
	embedded int
	missing  int
}

func (c textureCounts) add(o textureCounts) textureCounts {
	return textureCounts{embedded: c.embedded + o.embedded, missing: c.missing + o.missing}
}

type kitPackResult struct {
	mapText  string
	teams    int
	textures textureCounts
}

type kitPack struct {
	r   *Reader
	fam map[string][]string
}

func (p kitPack) packTeams(entries []kitMapEntry, opts kitsOptions) (kitPackResult, error) {
	keep := keepLeague(opts.leagues)
	var mapBuf strings.Builder
	fmt.Fprintf(&mapBuf, "# %s\n# team-id, \"League\\Team\"\n\n", opts.label)
	var result kitPackResult
	for _, e := range entries {
		if !keep(e) {
			continue
		}
		packed, counts, err := p.packTeam(e, opts)
		if err != nil {
			return kitPackResult{}, err
		}
		if !packed {
			continue
		}
		fmt.Fprintf(&mapBuf, "%s, \"%s\"\n", e.id, e.rel)
		result.teams++
		result.textures = result.textures.add(counts)
	}
	mapBuf.WriteString("\n# end of map\n")
	result.mapText = mapBuf.String()
	return result, nil
}

func (p kitPack) packTeam(e kitMapEntry, opts kitsOptions) (bool, textureCounts, error) {
	srcTeam := filepath.Join(opts.kservSrc, e.league, e.team)
	if fi, err := os.Stat(srcTeam); err != nil || !fi.IsDir() {
		return false, textureCounts{}, nil
	}
	dstTeam := filepath.Join(opts.out, e.league, e.team)
	if err := copyTree(srcTeam, dstTeam); err != nil {
		return false, textureCounts{}, fmt.Errorf("copy %s: %w", e.rel, err)
	}
	var total textureCounts
	err := filepath.WalkDir(dstTeam, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "config.txt" {
			return err
		}
		counts, err := p.embedConfigTextures(path)
		total = total.add(counts)
		return err
	})
	if err != nil {
		return false, textureCounts{}, fmt.Errorf("textures %s: %w", e.rel, err)
	}
	return true, total, nil
}

func (p kitPack) embedConfigTextures(configPath string) (textureCounts, error) {
	names, kitFile, err := configTextures(configPath)
	if err != nil {
		return textureCounts{}, err
	}
	want := append(names, p.fam[strings.ToLower(kitFile)]...)
	seen := map[string]bool{}
	var total textureCounts
	for _, n := range want {
		key := strings.ToLower(n)
		if n == "" || seen[key] {
			continue
		}
		seen[key] = true
		counts, err := p.embedTexture(filepath.Dir(configPath), n)
		if err != nil {
			return total, err
		}
		total = total.add(counts)
	}
	return total, nil
}

func (p kitPack) embedTexture(dir, name string) (textureCounts, error) {
	f, ok := p.r.FindFile(uniformTexDir + name + ".ftex")
	if !ok {
		return textureCounts{missing: 1}, nil
	}
	data, err := p.r.ReadFile(*f)
	if err != nil {
		return textureCounts{missing: 1}, nil
	}
	if err := os.WriteFile(filepath.Join(dir, name+".ftex"), data, 0644); err != nil {
		return textureCounts{}, err
	}
	return textureCounts{embedded: 1}, nil
}
