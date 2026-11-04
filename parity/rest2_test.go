package parity

import (
	"maps"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
)

// rest2Scenes are the dev scenes the rest2 parity tests walk.
var rest2Scenes = []string{"", "journal", "lockeddoor", "portrait", "cast", "inventory", "hall", "bestiary", "autocombat"}

// rest2Opens are what is done first in each scene: nothing, a step forward, the party member's sheet.
var rest2Opens = [][]front.Action{nil, {{Kind: front.ActForward}}, {{Kind: front.ActMember}}, {{Kind: front.ActMember}, {Kind: front.ActSheet}}}

// stop is one stop of the walk: por's View, its State and its art then.
type stop struct {
	v   front.View
	s   game.State
	art map[string]game.Picture
}

var (
	stopsOnce sync.Once
	stops     = map[string]stop{}
)

// walk runs check at each stop of the dev scenes, Enter on. The walk is made once for every check.
func walk(t *testing.T, check func(name string, st stop)) {
	stopsOnce.Do(func() {
		for _, scene := range rest2Scenes {
			for k, open := range rest2Opens {
				e := engine(t, true, scene)
				settled(e, front.Action{})
				settled(e, open...)
				for i := 0; i < 8; i++ {
					stops[scene+strconv.Itoa(k)+"/"+strconv.Itoa(i)] = stop{e.View(), e.Rules().State(), e.Data().Art}
					settled(e, front.Action{Kind: front.ActOK})
				}
			}
		}
	})
	for _, name := range slices.Sorted(maps.Keys(stops)) {
		check(name, stops[name])
	}
}

// TestPlaceParity: the window, area, square and facing, blanked and rebuilt.
func TestPlaceParity(t *testing.T) {
	walk(t, func(name string, st stop) {
		v, s := st.v, st.s
		b := v
		b.Window, b.Area, b.X, b.Y, b.Facing = 0, 0, 0, 0, 0
		c := classic.Place(b, s)
		if c.Window != v.Window || c.Area != v.Area || c.X != v.X || c.Y != v.Y || c.Facing != v.Facing {
			t.Errorf("%s: place %v %v %v,%v %v, por %v %v %v,%v %v", name, c.Window, c.Area, c.X, c.Y, c.Facing, v.Window, v.Area, v.X, v.Y, v.Facing)
		}
	})
}

// TestJournalParity: the quests and the lore.
func TestJournalParity(t *testing.T) {
	q, l := 0, 0
	walk(t, func(name string, st stop) {
		v, s := st.v, st.s
		if got := classic.Quests(s); !reflect.DeepEqual(got, v.Quests) {
			t.Errorf("%s: quests\n got %+v\nwant %+v", name, got, v.Quests)
		}
		if got := classic.Lore(s); !reflect.DeepEqual(got, v.Lore) {
			t.Errorf("%s: lore\n got %+v\nwant %+v", name, got, v.Lore)
		}
		q, l = q+len(v.Quests), l+len(v.Lore)
	})
	t.Logf("quests %d lore %d compared", q, l)
}

// TestWaitsParity: what the game waits on, the door's ways and the power pick.
func TestWaitsParity(t *testing.T) {
	seen := map[string]int{}
	walk(t, func(name string, st stop) {
		v, s := st.v, st.s
		b := v
		b.Pending, b.Door, b.Powers = 0, nil, nil
		c := classic.Waits(b, s)
		if c.Pending != v.Pending || !reflect.DeepEqual(c.Door, v.Door) || !reflect.DeepEqual(c.Powers, v.Powers) {
			t.Errorf("%s: waits %v %+v %+v, por %v %+v %+v", name, c.Pending, c.Door, c.Powers, v.Pending, v.Door, v.Powers)
		}
		seen[map[bool]string{true: "pending", false: "none"}[v.Pending != 0]]++
		if v.Door != nil {
			seen["door"]++
		}
		if v.Powers != nil {
			seen["powers"]++
		}
	})
	if seen["pending"] == 0 || seen["door"] == 0 {
		t.Errorf("not seen: %v", seen)
	}
	t.Logf("waits compared: %v", seen)
}

// TestPartyParity: the cards and the purse.
func TestPartyParity(t *testing.T) {
	cards, heads := 0, 0
	walk(t, func(name string, st stop) {
		v, s := st.v, st.s
		if got := classic.Party(s, st.art); !reflect.DeepEqual(got, v.Party) {
			t.Errorf("%s: party\n got %+v\nwant %+v", name, got, v.Party)
		}
		if got := classic.Purse(s); !reflect.DeepEqual(got, v.Purse) {
			t.Errorf("%s: purse got %+v want %+v", name, got, v.Purse)
		}
		cards += len(v.Party)
		for _, m := range v.Party {
			if m.HeadPix.Pix != nil {
				heads++
			}
		}
	})
	if cards == 0 {
		t.Error("no cards seen")
	}
	t.Logf("cards %d, with heads %d", cards, heads)
}

// TestSheetParity: the sheet, its page and the boxes.
func TestSheetParity(t *testing.T) {
	seen := map[string]int{}
	check := func(name string, st stop) {
		v, s := st.v, st.s
		b := v
		b.Sheet, b.SheetPage, b.Portrait, b.IconBox = nil, nil, front.CellRect{}, front.CellRect{}
		c := classic.Sheet(b, s)
		if !reflect.DeepEqual(c.Sheet, v.Sheet) || !reflect.DeepEqual(c.SheetPage, v.SheetPage) || c.Portrait != v.Portrait || c.IconBox != v.IconBox {
			t.Errorf("%s: sheet\n got %+v %+v %v %v\nwant %+v %+v %v %v", name, c.Sheet, c.SheetPage, c.Portrait, c.IconBox, v.Sheet, v.SheetPage, v.Portrait, v.IconBox)
		}
		if v.Sheet != nil {
			seen["sheet"]++
		}
		if v.SheetPage != nil {
			seen["page"]++
		}
		if v.Portrait != (front.CellRect{}) {
			seen["portrait"]++
		}
	}
	walk(t, check)
	if seen["page"] == 0 {
		t.Errorf("no sheet page seen: %v", seen)
	}
	t.Logf("sheet compared: %v", seen)
}
