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

	"gobox/core"
)

// Format is the manifest's format version; it changes when Manifest does.
const Format = 5

// ManifestFile is the manifest's name in an adventure folder.
const ManifestFile = "manifest.json"

// Contents are the entries every imported folder holds besides the manifest.
var Contents = []string{"scripts", "art", "data", "data/pictures.json", ManifestFile}

// CheckContents reports the first of Contents missing in folder dir: each
// game's import test runs it after Run, so an import step that clears what an
// earlier one wrote (por's data/ over data/pictures.json) fails there.
func CheckContents(dir string) error {
	for _, c := range Contents {
		if _, err := os.Stat(filepath.Join(dir, c)); err != nil {
			return fmt.Errorf("adventure folder %s has no %s: %w", dir, c, err)
		}
	}
	return nil
}

// Game is one game's importer.
type Game struct {
	ID      string            // the folder's name, adventures/<ID>
	Title   string            // shown in the list of adventures
	Ruleset string            // the ruleset compiled into the binary that plays it (rules1e, rulesxxvc)
	Record  core.RecordLayout // the layout of the adventure's character records (core.NoRecord: none)
	Version int               // the game's import version: bumped when what Import writes changes, so older folders read as stale (Read)
	Sources []string          // the originals' file patterns the import reads (filepath.Match, case-insensitive), EXE files included
	// Import writes the folder's files from the originals in dir src into
	// dir dst (empty, existing).
	Import func(src, dst string) error
	// Start and End give the title and end sequences from the originals in
	// dir src (nil func or nil Sequence: none); Fixed is one known without them.
	Start, End func(src string) (*Sequence, error)
}

// Fixed is a Game.Start or End that is s whatever the originals.
func Fixed(s *Sequence) func(string) (*Sequence, error) {
	return func(string) (*Sequence, error) { return s, nil }
}

// sequence is f's Sequence from src (nil f: none).
func sequence(f func(string) (*Sequence, error), src string) (*Sequence, error) {
	if f == nil {
		return nil, nil
	}
	return f(src)
}

// Manifest is an adventure folder's manifest.json.
type Manifest struct {
	ID      string
	Title   string
	Ruleset string
	Format  int
	Version int               `json:",omitempty"` // the game's import version (Game.Version) it was written by
	Record  core.RecordLayout `json:",omitempty"` // the character records' layout (gobox/char's codec)
	Start   *Sequence         `json:",omitempty"` // the title sequence
	End     *Sequence         `json:",omitempty"` // the end sequence
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
	start, err := sequence(g.Start, src)
	if err != nil {
		return fmt.Errorf("import %s: start: %w", id, err)
	}
	end, err := sequence(g.End, src)
	if err != nil {
		return fmt.Errorf("import %s: end: %w", id, err)
	}
	b, err := json.MarshalIndent(Manifest{ID: g.ID, Title: g.Title, Ruleset: g.Ruleset, Format: Format, Version: g.Version, Record: g.Record, Start: start, End: end, Source: key}, "", "\t")
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

// Sequence is a title or end sequence as data: its steps in order, then the
// backdrop that stays under the main menu. gobox plays it; a key or click
// ends a step's hold and goes no further (front/latch.go).
type Sequence struct {
	Steps    []SequenceStep
	Credits  []Credit  `json:",omitempty"` // the credits pages' lines, placed
	Boxes    []Cells   `json:",omitempty"` // the credits' boxes (nil: one frame on the page's edges)
	Backdrop []Layer   `json:",omitempty"` // under the main menu after the steps (COTAB: painting and logo), the menu zooming and fading in over it
	Remaster [2]string `json:",omitempty"` // two lines as game.Data.RemasterCredit: the remaster's own lines at the bottom of the credits page (game.Data.RemasterCredit)
	Short    []int     `json:",omitempty"` // the steps (indexes) a short intro plays (COTAB: SSI logo only unless the config asks); nil: all
	Full     bool      `json:",omitempty"` // the full intro by default (the original's config: POR's POOL.CFG); the host's "Full intro" option overrides it
	After    []Effect  `json:",omitempty"` // what an end sequence leaves in the game once it is over (the ruleset does it)
	Note     string    `json:",omitempty"` // an end sequence's closing note, shown once it is over
}

// Effect is what an end sequence does to the game when it is over; the ruleset
// carries it out (core.HealParty, core.SetWord).
type Effect int

const (
	EffectHealParty Effect = iota + 1 // every party member back to full hit points and well
	EffectOpenHall                    // the training hall teaches (the party menu offers it)
)

// StepKind is what a sequence step shows.
type StepKind int

const (
	StepPicture StepKind = iota // Layers over the page
	StepCredits                 // credits page Page
	StepAnim                    // a full-screen animation (Anim)
	StepText                    // Text alone, no picture
)

// SequenceStep is one step: what it shows, its tune and how long it holds.
type SequenceStep struct {
	Kind      StepKind
	Layers    []Layer  `json:",omitempty"` // StepPicture
	Text      []string `json:",omitempty"` // StepText, StepPicture: the paragraphs said over the step (colours are gobox's)
	Page      int      `json:",omitempty"` // StepCredits
	Anim      string   `json:",omitempty"` // StepAnim: the animation's base name in the folder's art
	Song      string   `json:",omitempty"` // the tune it starts (a song name or the game's song number)
	HoldMs    int      `json:",omitempty"` // 0: until a key or click
	Lightning Flash    `json:",omitempty"` // when the picture flashes (POR TITLE 2: FlashEGA)
	Whole     bool     `json:",omitempty"` // the layers are one whole-screen page
	Edition   Edition  `json:",omitempty"` // the art it is shown with (SotSB's Amiga credits card: EditionAmiga)
}

// Edition is which release's art a step belongs to.
type Edition uint8

const (
	EditionAny   Edition = iota // shown with every art
	EditionAmiga                // shown only while the Amiga art is up
)

// Shown is s as played with the Amiga art up (amiga) or not: the steps of
// another edition dropped, Short's indexes moved with them, every Edition
// EditionAny. play.TitleScene takes only a sequence passed through Shown.
func (s Sequence) Shown(amiga bool) Sequence {
	out := s
	out.Steps, out.Short = nil, nil
	at := map[int]int{} // old step index -> new
	for i, st := range s.Steps {
		if st.Edition == EditionAmiga && !amiga {
			continue
		}
		at[i] = len(out.Steps)
		st.Edition = EditionAny
		out.Steps = append(out.Steps, st)
	}
	for _, i := range s.Short {
		if n, ok := at[i]; ok {
			out.Short = append(out.Short, n)
		}
	}
	if s.Short != nil && out.Short == nil {
		out.Short = []int{}
	}
	return out
}

// Flash is when a step's picture flashes with lightning.
type Flash int

const (
	FlashNone Flash = iota // never
	FlashEGA               // in EGA graphics only (POR: ovr02 title_screen under the EGA driver)
)

// Graphics is the run's graphics mode, as the game's config picks it.
type Graphics int

const (
	GraphicsOther Graphics = iota // any mode without its own rule here (VGA, Tandy, CGA)
	GraphicsEGA
)

// On reports whether f flashes in graphics mode g.
func (f Flash) On(g Graphics) bool { return f == FlashEGA && g == GraphicsEGA }

// Art is a picture in the folder's art: an archive's block (TITLE 2), or
// with Kind ArtPNG the PNG File in the folder's art/ (POR's TITLE.PNG).
type Art struct {
	File  string
	Block int
	Kind  ArtKind `json:",omitempty"`
}

// ArtKind is what an Art's File names.
type ArtKind uint8

const (
	ArtDAX ArtKind = iota // a DAX archive, Block its block
	ArtPNG                // a PNG picture in the folder's art/
)

// Layer is a picture placed in 40x25 cells, optionally cycling frames.
type Layer struct {
	Art
	At         Cells
	Frames     []int `json:",omitempty"`
	FrameMs    int   `json:",omitempty"`
	Centred    bool  `json:",omitempty"` // the picture at its own size, centred in At (SotSB's Amiga credits card)
	MinParty   int   `json:",omitempty"` // shown only with at least this many party members (POR's finale: one picture per member)
	MsgDelayed bool  `json:",omitempty"` // FrameMs is as at the default message delay: the host's message delay slows or quickens it
}

// Cells is a box in 40x25 cells, Min and Max the corner cells (inclusive).
type Cells struct{ X1, Y1, X2, Y2 int }

// Credit is a credits line at its cell and colour (an EGA index), on page Page.
type Credit struct {
	X, Y, Colour int
	Text         string
	Page         int `json:",omitempty"`
}
