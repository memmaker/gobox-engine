package ui

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestOneCellSize (CLAUDE.md, one glyph cell size): every menu, pop-up,
// question and tooltip draws at K, every box (text pane, picture and view
// window) its ribbon at a whole multiple of K, on each kind of screen at
// several window sizes. Portrait cards are not drawn through these.
func TestOneCellSize(t *testing.T) {
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	pane := []string{"THE PARTY HAS WON."}
	menu := &game.Question{Text: "PRESS ENTER", Replies: []string{"CONTINUE", "LEAVE"}}
	yesNo := &game.Question{Text: "ARE YOU SURE?", Replies: []string{"YES", "NO"}, Keys: []rune{'Y', 'N'}}
	typing := &game.Question{Text: "Character name:", Typed: true, Max: 15}
	party := []game.Member{{Name: "ANNA", HP: 8, MaxHP: 8}, {Name: "BERT", HP: 5, MaxHP: 9}, {Name: "CARL", HP: 3, MaxHP: 3}}
	scene := func(c game.Common) game.State { return game.Scene{Common: c} }
	adventure := func(c game.Common) game.State { // the cards are the party's
		c.Adventure, c.Party = true, party
		return game.Exploring{Common: c}
	}
	screens := map[string]game.State{
		"pane":               scene(game.Common{Text: pane}),
		"pane, menu":         scene(game.Common{Text: pane, Ask: menu}),
		"typing":             scene(game.Common{Ask: typing}),
		"yes/no":             scene(game.Common{Ask: yesNo}),
		"adventure pane":     adventure(game.Common{Text: pane}),
		"adventure menu":     adventure(game.Common{Text: pane, Ask: menu}),
		"adventure question": adventure(game.Common{Text: pane, Ask: yesNo}),
		"adventure typing":   adventure(game.Common{Text: pane, Ask: typing}),
	}
	defer func() { drawnAt = nil }()
	for _, sz := range [][2]int{{800, 500}, {1280, 800}, {1920, 1200}, {2560, 1600}, {3840, 2400}} {
		dst := gpu.NewImage(sz[0], sz[1])
		K := float64(K(dst.Bounds()))
		for name, st := range screens {
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
			p, _ := presenterOn(st)
			p.Draw(dst, note)
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
	root := filepath.Join("..")
	for _, dir := range []string{"ui", "game", "engine"} {
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
