package rules1e

import (
	"fmt"
	"slices"
	"testing"

	"gobox/rules"
)

// TestAreaHitsAllItShows: an area aimed anywhere takes exactly the figures
// that stand on the squares it shows (AreaPreview.Area against Hit), for
// every radius and with walls cutting it; the cast takes its targets with
// the same buildNear (chooseSpellTargets' fromNear), so the preview is the
// cast's. Bug class: an area drawn bigger or smaller than what it hits.
func TestAreaHitsAllItShows(t *testing.T) {
	var foes []*Character
	for i := range 48 {
		foes = append(foes, newChar(fmt.Sprint("FOE", i), 1))
	}
	e := cotabFirstScene(t, foes...)
	r := &Rules{S: e}
	const cx, cy = 20, 12
	n := 0
	for y := cy - 3; y <= cy+3; y++ { // a 7x7 block of figures round the centre, the caster at its middle
		for x := cx - 3; x <= cx+3; x++ {
			e.Cpos[n+1] = CombatPos{X: byte(x), Y: byte(y), Size: 1}
			n++
		}
	}
	wall := Ground(0)
	for g := Ground(1); g < 0x40 && wall == 0; g++ {
		if e.GroundInfo(g)[0] == 0xFF { // blocks sight (boltPreview's test)
			wall = g
		}
	}
	if wall == 0 {
		t.Fatal("no ground blocks sight")
	}
	for _, w := range [][2]int{{cx + 2, cy}, {cx, cy - 2}} { // walls inside the block: line of sight cuts the area
		e.Cmap.Ground[w[1]][w[0]] = wall
	}
	e.CombatRedrawMap()
	e.Fight.On = true
	for aoe := 1; aoe <= 8; aoe++ {
		for _, at := range []rules.BoardPos{{X: cx, Y: cy}, {X: cx + 1, Y: cy - 1}, {X: cx - 3, Y: cy + 3}} {
			p := r.areaPreview(e.Party[0], rules.AreaFact{At: at, Area: aoe})
			if aoe == 4 && at.X == cx && at.Y == cy && len(p.Area) >= 49 {
				t.Fatalf("the walls cut nothing: %d squares", len(p.Area))
			}
			var want []rules.BoardPos
			for i := 1; i <= e.Ncomb; i++ {
				q := rules.BoardPos{X: int(e.Cpos[i].X), Y: int(e.Cpos[i].Y)}
				if slices.Contains(p.Area, q) {
					want = append(want, q)
				}
			}
			got := slices.Clone(p.Hit)
			cmp := func(a, b rules.BoardPos) int { return a.Y*100 + a.X - b.Y*100 - b.X }
			slices.SortFunc(got, cmp)
			slices.SortFunc(want, cmp)
			if !slices.Equal(got, want) {
				t.Errorf("area %d at %v: hits %v, the figures on its squares %v", aoe, at, got, want)
			}
		}
	}
}
