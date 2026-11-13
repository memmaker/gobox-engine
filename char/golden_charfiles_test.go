package char

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"gobox/rules1e"
	"testing"
)

// goldenCharFile is one character as a game's own reader read it (captured
// from the games' rules/ packages before they went): the original files and
// the name, levels, HP and experience the game read from them.
type goldenCharFile struct {
	Record, Items, Affects []byte
	ItemSize               int
	Name                   string
	Levels                 []int
	HP, MaxHP, Exp         int
}

// checkGoldenCharFiles: each file packs into rules1e's character and unpacks
// byte-equal, and the Sheet carries what the game read.
func checkGoldenCharFiles[R Later](t *testing.T, game string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "rules1e", "testdata", game, "charfiles.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var files []goldenCharFile
	if err := json.Unmarshal(b, &files); err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no golden characters")
	}
	for _, f := range files {
		lf := LaterFiles{Record: f.Record, Items: f.Items, Affects: f.Affects, ItemSize: f.ItemSize}
		packed, err := PackLater[R](lf)
		if err != nil {
			t.Fatal(err)
		}
		g, err := UnpackLater[R](packed, f.ItemSize)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(g.Record, f.Record) || !bytes.Equal(g.Items, f.Items) || !bytes.Equal(g.Affects, f.Affects) {
			t.Fatalf("%s: round trip differs", f.Name)
		}
		var j rules1e.CharacterJSON
		if err := json.Unmarshal(packed, &j); err != nil {
			t.Fatal(err)
		}
		s := j.Sheet
		for i, l := range f.Levels {
			if int(s.Levels[i]) != l {
				t.Errorf("%s: level %d is %d, game read %d", f.Name, i, s.Levels[i], l)
			}
		}
		if s.CharName != f.Name || int(s.HP) != f.HP || int(s.MaxHP) != f.MaxHP || int(s.Exp) != f.Exp {
			t.Errorf("sheet %q hp %d/%d xp %d; game read %q %d/%d %d", s.CharName, s.HP, s.MaxHP, s.Exp, f.Name, f.HP, f.MaxHP, f.Exp)
		}
	}
}

func TestGoldenCharFilesSOTSB(t *testing.T) { checkGoldenCharFiles[Record439](t, "sotsb") }
func TestGoldenCharFilesPOD(t *testing.T)   { checkGoldenCharFiles[Record510](t, "pod") }
func TestGoldenCharFilesCOTAB(t *testing.T) { checkGoldenCharFiles[Record422](t, "cotab") }
