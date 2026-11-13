package present

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/gpu"
	"gobox/ui"
)

// TestOneCellSize (CLAUDE.md, one glyph cell size): every menu, pop-up,
// question and tooltip draws at K, every box (text pane, picture and view
// window) its ribbon at a whole multiple of K, on each kind of screen at
// several window sizes. Portrait cards are not drawn through these.
func TestOneCellSize(t *testing.T) {
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	pane := &front.Pane{Lines: [][]front.PaneRun{{{Text: "THE PARTY HAS WON."}}}}
	menu := &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "PRESS ENTER", Commands: []front.Command{{Text: "CONTINUE"}, {Text: "LEAVE"}}}}
	yesNo := &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "ARE YOU SURE?", Commands: []front.Command{{Text: "YES", Key: 'Y'}, {Text: "NO", Key: 'N'}}}}
	typing := &front.Ask{Kind: front.AskText, Menu: front.Menu{Prompt: "Character name:"}, Max: 15}
	party := []game.Member{{Name: "ANNA", HP: 8, MaxHP: 8}, {Name: "BERT", HP: 5, MaxHP: 9}, {Name: "CARL", HP: 3, MaxHP: 3}}
	screens := map[string]front.View{
		"pane":               {Pane: pane},
		"pane, menu":         {Pane: pane, Ask: menu},
		"typing":             {Ask: typing},
		"yes/no":             {Ask: yesNo},
		"adventure pane":     {Adventure: true, Pane: pane},
		"adventure menu":     {Adventure: true, Pane: pane, Ask: menu},
		"adventure question": {Adventure: true, Pane: pane, Ask: yesNo},
		"adventure typing":   {Adventure: true, Pane: pane, Ask: typing},
	}
	defer func() { drawnAt = nil }()
	for _, sz := range [][2]int{{800, 500}, {1280, 800}, {1920, 1200}, {2560, 1600}, {3840, 2400}} {
		dst := gpu.NewImage(sz[0], sz[1])
		K := float64(ui.K(dst.Bounds()))
		for name, v := range screens {
			n := 0
			drawnAt = func(what drawnKind, on *gpu.Image, k float64) {
				if on != dst { // a menu drawn on an image of its own (a card's) is placed by its caller
					return
				}
				n++
				if what == drawnMenu && k != K || what == drawnBox && (k < K || k/K != math.Trunc(k/K)) {
					t.Errorf("%v %s: a %v drawn at %v, K %v", sz, name, what, k, K)
				}
			}
			p := &Presenter{}
			if v.Adventure { // the cards are the party's
				p = partyOn(party...)
			}
			p.Draw(dst, note, v)
			if n == 0 {
				t.Errorf("%v %s: no menu or box drawn", sz, name)
			}
		}
		dst.Deallocate()
	}
}

// TestOneK: the UI's text scale is computed in one place, ui.K. No other
// file divides a window's size by the page (400 x 250) to get a scale of
// its own; the combat UI is not drawn at a fraction of K.
func TestOneK(t *testing.T) {
	own := regexp.MustCompile(`(Dx|Dy)\(\)\s*/\s*(400|250|invW|invH|ui\.PageW|ui\.PageH|PageW|PageH)\b|scale\(\w+\)\)?\s*/\s*2|int\(k\)\s*/\s*2`)
	root := filepath.Join("..", "..")
	for _, dir := range []string{"front", "combat", "ui", "blobber", "game", "pool"} {
		filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, filepath.Join("ui", "fullpage.go")) {
				return err
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for i, l := range strings.Split(string(b), "\n") {
				if own.MatchString(l) && !strings.Contains(l, "the portrait card") {
					t.Errorf("%s:%d computes its own UI scale (use ui.K / scale): %s", path, i+1, strings.TrimSpace(l))
				}
			}
			return nil
		})
	}
}
