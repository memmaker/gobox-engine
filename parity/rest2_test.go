package parity

import (
	"image"
	"maps"
	"slices"
	"strconv"
	"sync"
	"testing"

	"gobox/front"
	"gobox/game"
)

// rest2Scenes are the dev scenes the rest2 parity tests walk.
var rest2Scenes = []string{"", "journal", "lockeddoor", "portrait", "cast", "inventory", "hall", "bestiary", "autocombat"}

// rest2Opens are what is done first in each scene: nothing, a step forward, the party member's sheet.
var rest2Opens = [][]front.Action{nil, {{Kind: front.ActForward}}, {{Kind: front.ActMember}}, {{Kind: front.ActMember}, {Kind: front.ActSheet}}}

// stop is one stop of the walk: por's State then.
type stop struct{ s game.State }

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
					stops[scene+strconv.Itoa(k)+"/"+strconv.Itoa(i)] = stop{e.Rules().State()}
					settled(e, front.Action{Kind: front.ActOK})
				}
			}
		}
	})
	for _, name := range slices.Sorted(maps.Keys(stops)) {
		check(name, stops[name])
	}
}

// TestStopsParity: the window, area, square and facing, the quests and lore,
// what the game waits on, the party cards and purse, and the sheet with its
// page and boxes, at each stop of the walk, are por's.
func TestStopsParity(t *testing.T) {
	seen := map[string]int{}
	walk(t, func(name string, st stop) {
		atKey(t, "TestStopsParity/"+name, groups(st.s))
		s := st.s
		seen["quests"] += len(s.Quests)
		seen["lore"] += len(s.Lore)
		seen["cards"] += len(s.Party)
		for k, on := range map[string]bool{"pending": s.Pending != nil, "door": s.Door != nil, "powers": s.PowerPick != nil, "sheet": s.Sheet != nil, "page": s.SheetPage != nil, "portrait": s.Portrait != (image.Rectangle{})} {
			if on {
				seen[k]++
			}
		}
	})
	if seen["pending"] == 0 || seen["door"] == 0 || seen["cards"] == 0 || seen["page"] == 0 {
		t.Errorf("not seen: %v", seen)
	}
	t.Logf("compared: %v", seen)
}
