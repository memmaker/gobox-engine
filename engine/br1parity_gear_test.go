package engine

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rulesxxvc"
)

// brCopyItem gives member id a copy of its item h, unreadied, as a new item (the ruleset's save is JSON: its one door for a
// test), the way a kit holds a kind in two stacks.
func brCopyItem(t *testing.T, e *Engine, id rules.CharID, h rules.ItemHandle) {
	t.Helper()
	r := e.ruleset().(*rulesxxvc.Rules)
	b, err := r.Save()
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	c := doc["Chars"].(map[string]any)[strconv.Itoa(int(id))].(map[string]any)
	items := c["Items"].([]any)
	top := 0.0
	for _, it := range items {
		top = max(top, it.(map[string]any)["Handle"].(float64), it.(map[string]any)["Slot"].(float64))
	}
	for _, it := range items {
		if m := it.(map[string]any); rules.ItemHandle(m["Handle"].(float64)) == h {
			c["Items"] = append(items, map[string]any{"Handle": top + 100, "Slot": top + 1, "Readied": false, "Ref": m["Ref"]})
			if b, err = json.Marshal(doc); err != nil {
				t.Fatal(err)
			}
			if err := r.Load(b); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("no item %d", h)
}

// The gear menu's item verbs (viewchar.pas UseItemMenu): ports rules TestClassItemActs (Ready, Trade, Drop, Halve, Join are
// offered) and TestClassJoinedName (a joined stack shows its new count). The rules' own cases (the limits, Overloaded, a
// readied item) are rulesxxvc/gear_test.go's.
func TestBR1GearVerbs(t *testing.T) { brGearVerbs(t, gameBR1) }
func TestBR2GearVerbs(t *testing.T) { brGearVerbs(t, gameBR2) }

func brGearVerbs(t *testing.T, g brGame) {
	e := br1BeginWith(t, brEngine(t, g), 2)
	e.rebuild()
	items := func(m int) []game.Item { return commonOf(e.build()).Party[m].Items }
	offered := func() map[string]bool {
		seen := map[string]bool{}
		for _, a := range commonOf(e.build()).Allowed.Plain {
			switch a.(type) {
			case game.Ready:
				seen["ready"] = true
			case game.Unready:
				seen["unready"] = true
			case game.Drop:
				seen["drop"] = true
			case game.Give:
				seen["trade"] = true
			case game.Halve:
				seen["halve"] = true
			case game.Join:
				seen["join"] = true
			}
		}
		return seen
	}
	seen := offered()
	for _, k := range []string{"drop", "trade", "join"} { // Ready is Unready when the kit is all readied; Halve only with room
		if !seen[k] {
			t.Errorf("the gear menu offers no %s: %v", k, seen)
		}
	}
	if !seen["ready"] && !seen["unready"] {
		t.Errorf("the gear menu offers no Ready or Unready: %v", seen)
	}
	// A joinable kind in two stacks (BR2's stun grenades, copied): Join makes one with the new count (Item.Count: the front draws it beside the name).
	var grenade game.Item
	for _, it := range items(0) {
		if it.Name == "Grenade Stun" {
			grenade = it
		}
	}
	if grenade.Handle == rules.NoItem {
		t.Skipf("%s: the kit has no stun grenades", g.Name)
	}
	for _, it := range items(0) { // room for the halves: the kit's spare swords go
		if it.Name == "Sword Mono" && len(items(0)) > 14 {
			if _, err := e.doNow(game.Drop{Member: 0, Item: it.Handle}); err != nil {
				t.Fatal(err)
			}
		}
	}
	brCopyItem(t, e, e.partyIDs()[0], grenade.Handle)
	e.rebuild()
	if err := e.OpenItems(0, false); err != nil {
		t.Fatal(err)
	}
	e.rebuild()
	total, n0 := 2*grenade.Count, len(items(0))
	if _, err := e.doNow(game.Join{Member: 0, Item: grenade.Handle}); err != nil {
		t.Fatal(err)
	}
	var joined game.Item
	for _, it := range items(0) {
		if it.Count == total && strings.HasSuffix(it.Name, "Grenade Stun") {
			joined = it
		}
	}
	if len(items(0)) != n0-1 || joined.Handle == rules.NoItem {
		t.Fatalf("joined two stacks of %d: %d items (was %d), item %+v", grenade.Count, len(items(0)), n0, joined)
	}
	// Halve: half, rounded down, a new item beside it.
	if _, err := e.doNow(game.Halve{Member: 0, Item: joined.Handle}); err != nil {
		t.Fatal(err)
	}
	var half game.Item
	for _, it := range items(0) {
		if it.Handle != joined.Handle && it.Count == total/2 && strings.HasSuffix(it.Name, "Grenade Stun") {
			half = it
		}
	}
	if len(items(0)) != n0 || half.Handle == rules.NoItem {
		t.Fatalf("halved %d: %d items (was %d), half %+v", total, len(items(0)), n0, half)
	}
	// Trade: the half goes to the second member and leaves the first.
	had := len(items(1))
	if _, err := e.doNow(game.Give{From: 0, To: 1, Item: half.Handle}); err != nil {
		t.Fatal(err)
	}
	if len(items(0)) != n0-1 || len(items(1)) != had+1 {
		t.Errorf("traded: giver %d items (was %d), receiver %d (had %d)", len(items(0)), n0, len(items(1)), had)
	}
}
