// Package fixture builds a save fixture for a trace scenario: a saved game
// with its party, moved into another area. The game supplies how its save is
// read and edited (Game); the flags, the walls/vars parsing and the file
// writing are shared. See API.md.
package fixture

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// State is a loaded save the fixture edits.
type State interface {
	// EnterArea replaces the saved script with ecl and loads GEO block geo;
	// walls are the WALLDEF blocks of wall slots 1..3 (255 none).
	EnterArea(disk, block, geo int, ecl []byte, walls [3]int)
	SetPos(x, y, dir int) // negative keeps the saved value
	SetVar(addr int, v uint16) error
	Bytes() []byte
	Party() []string // member file names, without extension
}

// Game is what a game supplies.
type Game struct {
	// Root resolves the -data flag (may be empty) to the game data directory.
	Root func(data string) (string, error)
	// Open reads the save to start from (slot is -slot, may be empty).
	Open func(root, slot string) (State, error)
	// ECL finds ECL block n: its disk and the script bytes without the 2-byte header.
	ECL func(root string, block int) (disk int, ecl []byte, err error)
	// Member reads the party member file name+ext (.SAV .SPC .ITM) from the save dir.
	Member func(root, name, ext string) ([]byte, error)
	// Strong makes a .SAV win fights.
	Strong func([]byte) []byte
	// ParseDecision (optional) is the game's engine.ParseDecision; with
	// -decisions FILE every line must parse.
	ParseDecision func(line string) error
	// NeedSlot makes -slot required.
	NeedSlot bool
}

type sets []string

func (s *sets) String() string     { return strings.Join(*s, ",") }
func (s *sets) Set(v string) error { *s = append(*s, v); return nil }

// Run parses args (os.Args[1:]) and writes the fixture.
func Run(g Game, args []string) error {
	fs := flag.NewFlagSet("fixture", flag.ContinueOnError)
	data := fs.String("data", "", "game data directory")
	out := fs.String("out", "", "fixture directory to write")
	block := fs.Int("block", -1, "ECL block to enter (-1 keeps the saved area)")
	x := fs.Int("x", -1, "party x")
	y := fs.Int("y", -1, "party y")
	slot := fs.String("slot", "", "the named save to start from")
	dir := fs.Int("dir", -1, "facing 0 N, 2 E, 4 S, 6 W")
	geo := fs.Int("geo", -1, "GEO block to load (default: the ECL block)")
	walls := fs.String("walls", "", "WALLDEF blocks of wall slots 1,2,3 (255 none)")
	strong := fs.Bool("strong", false, "make the party win fights")
	dec := fs.String("decisions", "", "decision file to validate with the game's ParseDecision")
	var vars sets
	fs.Var(&vars, "set", "VAR=VAL, hex, repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *out == "" || (g.NeedSlot && *slot == "") {
		return fmt.Errorf("-out%s required", map[bool]string{true: " and -slot are", false: " is"}[g.NeedSlot])
	}
	if *dec != "" {
		if g.ParseDecision == nil {
			return fmt.Errorf("-decisions: game has no ParseDecision")
		}
		b, err := os.ReadFile(*dec)
		if err != nil {
			return err
		}
		for i, l := range strings.Split(string(b), "\n") {
			if l = strings.TrimSpace(l); l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			if err := g.ParseDecision(l); err != nil {
				return fmt.Errorf("%s:%d: %v", *dec, i+1, err)
			}
		}
	}
	root, err := g.Root(*data)
	if err != nil {
		return err
	}
	s, err := g.Open(root, *slot)
	if err != nil {
		return err
	}
	if *block >= 0 {
		disk, ecl, err := g.ECL(root, *block)
		if err != nil {
			return err
		}
		gb := *geo
		if gb < 0 {
			gb = *block // most areas load the GEO block of their own number
		}
		ws := strings.Split(*walls, ",")
		if len(ws) != 3 {
			return fmt.Errorf("-block needs -walls A,B,C")
		}
		var ids [3]int
		for i, w := range ws {
			if ids[i], err = strconv.Atoi(w); err != nil {
				return fmt.Errorf("bad -walls %q", *walls)
			}
		}
		s.EnterArea(disk, *block, gb, ecl, ids)
	}
	s.SetPos(*x, *y, *dir)
	for _, kv := range vars {
		k, v, ok := strings.Cut(kv, "=")
		a, err1 := strconv.ParseUint(k, 16, 16)
		n, err2 := strconv.ParseUint(v, 16, 16)
		if !ok || err1 != nil || err2 != nil {
			return fmt.Errorf("bad -set %q", kv)
		}
		if err := s.SetVar(int(a), uint16(n)); err != nil {
			return fmt.Errorf("-set %s: %v", k, err)
		}
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "SAVGAMA.DAT"), s.Bytes(), 0o644); err != nil {
		return err
	}
	for _, name := range s.Party() {
		for _, ext := range []string{".SAV", ".SPC", ".ITM"} {
			b, err := g.Member(root, name, ext)
			if err != nil {
				continue
			}
			if ext == ".SAV" && *strong {
				b = g.Strong(b)
			}
			if err := os.WriteFile(filepath.Join(*out, name+ext), b, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}
