// Package importer is gobox's one import entry: `gobox import <game>
// <originals>` turns a game's original files into its adventure folder,
// adventures/<id>/, once. Each game registers one Game; its Import may know
// the game's formats, nothing at run time may. The folder ends with a
// manifest (manifest.json) naming the game, its ruleset and the originals it
// came from (Manifest.Source).
package importer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Format is the manifest's format version; it changes when Manifest does.
const Format = 1

// ManifestFile is the manifest's name in an adventure folder.
const ManifestFile = "manifest.json"

// Game is one game's importer.
type Game struct {
	ID      string   // the folder's name, adventures/<ID>
	Title   string   // shown in the list of adventures
	Ruleset string   // the ruleset compiled into the binary that plays it (rules1e, rulesxxvc)
	Sources []string // the originals' file patterns the import reads (filepath.Match, case-insensitive), EXE files included
	// Import writes the folder's files from the originals in dir src into
	// dir dst (empty, existing).
	Import func(src, dst string) error
}

// Manifest is an adventure folder's manifest.json.
type Manifest struct {
	ID      string
	Title   string
	Ruleset string
	Format  int
	Start   string            `json:",omitempty"` // the title sequence (hook, filled with the declarative sequences)
	End     string            `json:",omitempty"` // the end sequence (hook)
	Source  map[string]string // each original read: its name, its SHA-256
}

var games = map[string]Game{}

// Register adds g; a second Game with the same ID panics.
func Register(g Game) {
	if _, ok := games[g.ID]; ok {
		panic("importer: " + g.ID + " registered twice")
	}
	games[g.ID] = g
}

// IDs are the registered games, sorted.
func IDs() []string {
	ids := make([]string, 0, len(games))
	for id := range games {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// Run imports game id from the originals in src into root/<id>/, replacing
// the folder; the manifest is written last, so a failed import leaves no
// valid folder.
func Run(id, src, root string) error {
	g, ok := games[id]
	if !ok {
		return fmt.Errorf("import: no game %q (have %s)", id, strings.Join(IDs(), ", "))
	}
	key, err := SourceKey(src, g.Sources)
	if err != nil {
		return err
	}
	if len(key) == 0 {
		return fmt.Errorf("import %s: no originals in %s", id, src)
	}
	dst := filepath.Join(root, id)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if err := g.Import(src, dst); err != nil {
		return fmt.Errorf("import %s: %w", id, err)
	}
	b, err := json.MarshalIndent(Manifest{ID: g.ID, Title: g.Title, Ruleset: g.Ruleset, Format: Format, Source: key}, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dst, ManifestFile), b, 0o644)
}

// SourceKey is each file of dir matching one of patterns (case-insensitive),
// by name, with its SHA-256.
func SourceKey(dir string, patterns []string) (map[string]string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	key := map[string]string{}
	for _, e := range ents {
		if e.IsDir() || !matches(e.Name(), patterns) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(b)
		key[e.Name()] = hex.EncodeToString(sum[:])
	}
	return key, nil
}

func matches(name string, patterns []string) bool {
	for _, p := range patterns {
		if ok, _ := filepath.Match(strings.ToUpper(p), strings.ToUpper(name)); ok {
			return true
		}
	}
	return false
}

// Main is the import command's body: args are [-out dir] <game> <originals>.
func Main(args []string) error {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	out := fs.String("out", "adventures", "the adventures directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: import [-out dir] <game> <originals>; games: %s", strings.Join(IDs(), ", "))
	}
	return Run(fs.Arg(0), fs.Arg(1), *out)
}

// SameTree reports the first difference between the files under a and the
// files under b (names and bytes), nil when they hold the same: an
// importer's equality test against the decode the game reads today.
func SameTree(a, b string) error {
	want := map[string][]byte{}
	err := filepath.WalkDir(b, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(b, p)
		want[rel], err = os.ReadFile(p)
		return err
	})
	if err != nil {
		return err
	}
	err = filepath.WalkDir(a, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(a, p)
		got, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		w, ok := want[rel]
		if !ok {
			return fmt.Errorf("%s: only in %s", rel, a)
		}
		if !bytes.Equal(got, w) {
			return fmt.Errorf("%s differs", rel)
		}
		delete(want, rel)
		return nil
	})
	if err != nil {
		return err
	}
	for rel := range want {
		return fmt.Errorf("%s: only in %s", rel, b)
	}
	return nil
}
