package parity

import (
	"image"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"por/rules"
)

// sameMap says classic's map and full map equal por's, and State knows
// the squares por's map shows seen.
func sameMap(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	v, s := e.View(), e.Rules().State()
	if v.Map == nil {
		t.Fatalf("%s: no map", name)
	}
	if c := classic.View(v, s); !reflect.DeepEqual(c.Map, v.Map) {
		t.Errorf("%s: map\n got %+v\nwant %+v", name, c.Map, v.Map)
	}
	for _, at := range []image.Point{{}, {s.Where.X, s.Where.Y}, {9, 0}, {15, 15}, {20, 3}} {
		if len(s.Map.Marked) > 0 && at == (image.Point{20, 3}) {
			at = s.Map.Marked[0]
		}
		f := e.FullMapView(at, nil)
		if c := classic.FullMap(f, s); !reflect.DeepEqual(c, f) {
			t.Errorf("%s: full map at %v\n got %+v\nwant %+v", name, at, c, f)
		}
	}
	for y, row := range v.Map.Squares {
		for x, q := range row {
			if known := slices.Contains(s.Map.Known, image.Pt(x, y)); known != (q.Seen != front.Unseen) {
				t.Errorf("%s: %d,%d known %v, por seen %d", name, x, y, known, q.Seen)
			}
		}
	}
}

func TestMapParity(t *testing.T) {
	e := engine(t, true, "")
	settled(e)
	sameMap(t, "start", e)
	for _, a := range []front.ActKind{front.ActForward, front.ActTurnRight, front.ActForward, front.ActForward, front.ActTurnLeft, front.ActForward} {
		settled(e, front.Action{Kind: a})
		sameMap(t, "walked", e)
	}
	s := e.Rules().State()
	e.Mark(s.Where.X, s.Where.Y, "camp spot")
	settled(e, front.Action{})
	sameMap(t, "marked", e)
	if s := e.Rules().State(); !slices.Contains(s.Map.Notes, "camp spot") {
		t.Errorf("notes %v", s.Map.Notes)
	}

	d := engine(t, true, "lockeddoor")
	settled(d, front.Action{})
	settled(d, front.Action{Kind: front.ActClose})
	if d.View().Map != nil {
		sameMap(t, "locked door", d)
	}
	if !slices.ContainsFunc(d.Rules().State().Map.Doors, func(w game.Wall) bool { return w.Kind == 2 }) {
		t.Error("no locked door in State")
	}

	sameMap(t, "capture", capture(t))
}

// capture is por exploring from the CAPTURE save (New Phlan).
func capture(t *testing.T) *rules.Engine {
	c := engine(t, true, "")
	settled(c)
	src := filepath.Join(porDir(t), "data", "save", "saves", "CAPTURE")
	if err := os.CopyFS(filepath.Join(c.Cfg.SaveDir, "saves", "CAPTURE"), os.DirFS(src)); err != nil {
		t.Skip("no CAPTURE save:", err)
	}
	if !c.Load("CAPTURE", false) {
		t.Fatal("CAPTURE does not load")
	}
	settled(c, front.Action{})
	return c
}

// TestCampState: camp in State as por's View shows it (the book's spells).
func TestCampState(t *testing.T) {
	b := engine(t, true, "bestiary")
	settled(b, front.Action{})
	if b.View().Camp == nil || b.Rules().State().Mode != game.ModeCamp {
		t.Fatalf("bestiary camp: %v", b.Rules().State().Mode)
	}
	e := capture(t)
	booked := false
	for i := range e.Rules().State().Party { // camp shows the selected member's book
		e.Rules().Apply(game.Select{N: i})
		if r := e.Rules().Apply(game.Camp{}); r.Err != nil {
			t.Fatal(r.Err)
		}
		v, s := settled(e, front.Action{}), e.Rules().State()
		if v.Camp == nil || s.Mode != game.ModeCamp || s.Selected != i {
			t.Fatalf("camp: %v member %d, por %+v", s.Mode, s.Selected, v.Camp)
		}
		var book []int
		for _, sp := range s.Party[s.Selected].Spells {
			if !sp.Memorized {
				book = append(book, sp.ID)
			}
		}
		var want []int
		for _, b := range v.Camp.Book {
			want = append(want, b.ID)
		}
		booked = booked || len(want) > 0
		if !slices.Equal(book, want) {
			t.Errorf("member %d book %v, por %v", i, book, want)
		}
		e.LeaveCamp()
		settled(e, front.Action{})
	}
	if !booked {
		t.Error("no one has a spell book")
	}
}
