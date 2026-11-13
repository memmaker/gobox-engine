package engine

import (
	"bufio"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestPorWallsAsMaps: for every area POR's hand-made remastered maps drew
// (testdata/por_faces.txt: "area NAME SET GEO BLOCK:SLOT..." then "TEXTURE
// X,Y,SIDE..."), the live 3D view (mapWalls over GameTables.Walls) gives
// the same texture on every face the map had, a free-standing wall's far
// side through Scenery.TwoSided; and the area's flats and scenery come with it.
func TestPorWallsAsMaps(t *testing.T) {
	e := testEngine(t)
	if e.game().Walls == nil {
		t.Skip("no POR wall table imported")
	}
	f, err := os.Open("testdata/por_faces.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e.viewMode = 1
	num := func(s string) int {
		n, err := strconv.Atoi(s)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	var area string
	var faces, areas int
	w := e.mapWalls()
	for sc := bufio.NewScanner(f); sc.Scan(); {
		fl := strings.Fields(sc.Text())
		if fl[0] == "area" {
			area, areas = fl[1], areas+1
			e.set = rules.Set(num(fl[2]))
			e.mapPairA = [rules.WallSlots]int{-1, -1, -1}
			e.LoadGeo(num(fl[3]))
			for _, l := range fl[4:] {
				bs := strings.Split(l, ":")
				e.LoadWalldef(num(bs[0]), num(bs[1]))
			}
			w = e.mapWalls()
			a := e.game().Walls.Area(e.set, num(fl[3]))
			switch {
			case a == nil || a.Floor == "" || a.Sky == "":
				t.Errorf("%s: no flats: %+v", area, a)
			case w.Floor != a.Floor || w.Sky != a.Sky || !reflect.DeepEqual(w.Scenery, a.Scenery):
				t.Errorf("%s: the view's flats or scenery are not the area's", area)
			}
			continue
		}
		for _, k := range fl[1:] {
			p := strings.Split(k, ",")
			x, y, s := num(p[0]), num(p[1]), num(p[2])
			faces++
			got := w.Pool[w.Side[y][x][s]]
			if v := w.Side[y][x][s]; v == 0 { // the far side of a free-standing wall
				nx, ny := x+[4]int{0, 1, 0, -1}[s], y+[4]int{-1, 0, 1, 0}[s]
				if ny < 0 || ny >= rules.AreaSize || nx < 0 || nx >= rules.AreaSize {
					t.Errorf("%s: face %d,%d,%d: open, off the map", area, x, y, s)
				} else if v := w.Side[ny][nx][s^2]; slices.Contains(w.Scenery.TwoSided, v) {
					got = w.Pool[v]
				}
			}
			if got != fl[0] {
				t.Errorf("%s: face %d,%d,%d is %q, the map's %q", area, x, y, s, got, fl[0])
			}
		}
	}
	if areas != 29 || faces < 12000 {
		t.Errorf("%d areas, %d faces checked", areas, faces)
	}
	if a := e.game().Walls.Area(8, 29); a == nil || !slices.ContainsFunc(a.Scenery.Props, func(p game.Prop) bool { return p.Ring != nil }) {
		t.Error("Kuto's well (set 8, geo 29) is gone")
	}
}
