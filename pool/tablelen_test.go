package pool

import (
	"bytes"
	"os"
	"rules1e"
	"testing"

	"gobox/game/play"
)

// TestGroundTableLen (BUG-CLASSES 42): groundTable is START.EXE's DS:2744
// table record for record, all 66 (the last two overlap footprintTable's
// unused size 0 at DS:2844).
func TestGroundTableLen(t *testing.T) {
	exe, err := os.ReadFile("../data/por/sources/original/START.EXE")
	if err != nil {
		t.Skip("no START.EXE")
	}
	var b []byte
	for _, r := range rules1e.GroundTable {
		b = append(b, fileBytes(r[:])...)
	}
	i := bytes.Index(exe, b[:64*4])
	if i < 0 {
		t.Fatal("groundTable's 64 first records are not in START.EXE")
	}
	if !bytes.Equal(exe[i:i+len(b)], b) {
		t.Errorf("groundTable (%d records) differs from START.EXE's", len(rules1e.GroundTable))
	}
	if n := len(rules1e.GroundTable); n != 66 {
		t.Errorf("groundTable has %d records, START.EXE 66", n)
	}
}

// TestFightArtResolves (BUG-CLASSES 42): every square's tile, effect and
// foe icon a fight publishes resolves (gobox skips unresolved art
// silently), across the dev fights.
func TestFightArtResolves(t *testing.T) {
	for _, sc := range []string{"combat", "missile", "castfight", "wildcombat", "autocombat"} {
		e := testEngine(t)
		e.Explore = true
		e.DevAction = e.DevScene(sc)
		seed := uint32(1)
		e.FixedSeed = &seed
		e.Init()
		e.Start(e.Play)
		fights := 0
		for range 300 {
			st := e.State()
			if f, ok := st.(play.Fighting); ok {
				fights++
				c := f.Field
				art := e.Art()
				for _, q := range c.Squares {
					if q.Tile >= len(c.Tiles) || c.Tiles[q.Tile] == "" {
						t.Fatalf("%s: square %v tile %d has no art", sc, q.At, q.Tile)
					}
					if _, ok := art[c.Tiles[q.Tile]]; !ok {
						t.Fatalf("%s: tile %q not in Art()", sc, c.Tiles[q.Tile])
					}
				}
				for _, f := range c.Clouds {
					if f.Art.Pixels == nil {
						t.Fatalf("%s: an effect at %d,%d has no picture", sc, f.X, f.Y)
					}
				}
				for _, u := range c.Units {
					if u.Foe && !u.Dead && u.Art.Pixels == nil {
						t.Fatalf("%s: foe %s has no icon", sc, u.Name)
					}
				}
			}
			switch {
			case commonOf(st).Ask != nil:
				if _, err := e.Do(play.Reply{}); err != nil {
					e.Do(play.Continue{})
				}
			case playAllows(st, play.Quick{}):
				e.Do(play.Quick{})
			default:
				e.Step()
				press(e)
			}
		}
		if fights == 0 {
			t.Errorf("%s: no fight", sc)
		}
	}
}
