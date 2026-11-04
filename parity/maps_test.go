package parity

import (
	"image"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
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
	if !slices.ContainsFunc(d.Rules().State().Map.Doors, func(w game.Wall) bool { return w.State == game.DoorLocked || w.State == game.DoorWizardLocked }) {
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
	sameCamp(t, "bestiary", b)
	if bs, v := b.Rules().State().Bestiary, b.View().Camp.Bestiary; len(bs) != len(v) || len(v) == 0 || bs[0].Name != v[0].Name || bs[0].Text != v[0].Text {
		t.Errorf("bestiary %d, por %d", len(bs), len(v))
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
		sameCamp(t, "capture", e)
		if len(want) > 0 { // a spell planned: the book's counts and its level's plan
			if err := e.PlanSpell(i, want[0], 1); err != nil {
				t.Fatal(err)
			}
			settled(e, front.Action{})
			if v := e.View().Camp; v.Book[0].Count != 1 {
				t.Errorf("member %d planned %+v", i, v.Book[0])
			}
			sameCamp(t, "planned", e)
		}
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

// sameCamp says classic's camp equals por's.
func sameCamp(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	v, s := e.View(), e.Rules().State()
	if c := classic.View(v, s); !reflect.DeepEqual(c.Camp, v.Camp) {
		t.Errorf("%s: camp\n got %+v\nwant %+v", name, c.Camp, v.Camp)
	}
}

// TestOverlandParity: the overland's terrain, art, marker and seen squares
// in State and Data.Art build the map por's View publishes.
func TestOverlandParity(t *testing.T) {
	e := engine(t, true, "overland")
	v := settled(e, front.Action{})
	s := e.Rules().State()
	if v.AreaMap == nil || len(s.Map.Terrain) != 44*36 || len(s.Map.Known) != 44*36 {
		t.Fatalf("overland: map %v, %d terrain, %d known", v.AreaMap != nil, len(s.Map.Terrain), len(s.Map.Known))
	}
	art := e.Data().Art
	if art["marker/"+strconv.Itoa(s.Where.Facing)].Pixels == nil || art["terrain/"+strconv.Itoa(s.Map.Terrain[0])].Pixels == nil {
		t.Fatal("no overland art")
	}
	if got := classic.AreaMap(v.AreaMap, s, art); !reflect.DeepEqual(got, v.AreaMap) {
		t.Errorf("area map\n got %+v\nwant %+v", got.Party, v.AreaMap.Party)
	}
	s.Map.Known = s.Map.Known[1:] // one square unseen: drawn as none
	if got := classic.AreaMap(v.AreaMap, s, art); got.Seen == nil || got.Seen[0][0] {
		t.Error("unseen square drawn")
	}
}

// TestBestiaryParity: the bestiary classic builds from State (danger, fight
// picture, combat icon) equals por's, and carries the art.
func TestBestiaryParity(t *testing.T) {
	b := engine(t, true, "bestiary")
	settled(b, front.Action{})
	sameCamp(t, "bestiary", b)
	for _, be := range b.Rules().State().Bestiary {
		if be.Danger < 1 || be.Danger > 8 || be.Icon.Pixels == nil {
			t.Errorf("%s: danger %d, icon %v", be.Name, be.Danger, be.Icon.Pixels != nil)
		}
	}
}

// TestItemParity: the inventory party's powers, doll slots and scrolls build
// por's inventory, and every readied slot is a shared name.
func TestItemParity(t *testing.T) {
	e := engine(t, true, "inventory")
	settled(e, front.Action{})
	if err := e.OpenItems(0, true); err != nil {
		t.Fatal(err)
	}
	settled(e, front.Action{})
	members(t, "items", e)
	shared := []string{"", "weapon", "shield", "armor", "gauntlets", "helm", "belt", "robe", "cloak", "boots", "ring", "ring2", "arrows", "bolts"}
	for _, m := range e.Rules().State().Party {
		for _, it := range m.Items {
			if !slices.Contains(shared, it.Slot) {
				t.Errorf("%s: slot %q", it.Name, it.Slot)
			}
		}
	}
}
