package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/adventure"
)

// worldNames are the engine's own world words (their script field names), in
// the order of the recorded table below.
var worldNames = [...]string{"AreaKind", "PrevX", "PrevY", "RiverOpen", "Search", "HoldMove", "LeftMap", "Geo",
	"WallSplitA", "WallSplitB", "NoCoords", "MsgDelay", "ViewFlags", "MaxDist", "Portrait", "Temple", "Shop", "SymbolSet",
	"Morale", "Surprise", "MonsterHit", "PartyHit", "PartyMove", "NoLoot", "Duel", "MonsterDist", "CombatEnd", "Beaten", "AttackedAlly"}

// engineWords are the world words as the engine sees them, in worldNames' order.
func engineWords(e *Engine) [len(worldNames)]int {
	return [...]int{
		int(e.world.AreaKind), int(e.world.PrevX), int(e.world.PrevY), int(e.world.RiverOpen),
		int(e.world.Search), int(e.world.HoldMove), int(e.world.LeftMap), int(e.world.Geo),
		int(e.world.WallSplitA), int(e.world.WallSplitB), int(e.world.NoCoords), int(e.world.MsgDelay),
		int(e.world.ViewFlags), int(e.world.MaxDist), int(e.world.Portrait), int(e.world.Temple),
		int(e.world.Shop), int(e.world.SymbolSet),
		int(e.world.Morale), int(e.world.Surprise), int(e.world.MonsterHit), int(e.world.PartyHit),
		int(e.world.PartyMove), int(e.world.NoLoot), int(e.world.Duel), int(e.world.MonsterDist),
		int(e.world.CombatEnd), int(e.world.Beaten), int(e.world.AttackedAlly),
	}
}

// shippedSlots are the shipped slot folders by "<game>/<slot>".
func shippedSlots(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, pat := range []string{"../data/*/sources/save/saves/*/" + adventure.GameFile, "../data/*/saves/saves/*/" + adventure.GameFile} {
		m, err := filepath.Glob(pat)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range m {
			dir := filepath.Dir(f)
			out[strings.Split(filepath.ToSlash(dir), "/")[2]+"/"+filepath.Base(dir)] = dir
		}
	}
	return out
}

// slotEngine is the engine of slot key's game, with a copy of the slot under a temp save folder
// (loading erases roster copies): tests never touch the shipped files.
func slotEngine(t *testing.T, key, dir string) *Engine {
	t.Helper()
	var e *Engine
	switch strings.Split(key, "/")[0] {
	case "por":
		e = testEngine(t)
	case "sotsb":
		e = advEngine(t, advEnvSOTSB)
	case "pod":
		e = advEngine(t, advEnvPOD)
	case "cotab":
		e = advEngine(t, advEnvCOTAB)
	default:
		t.Skip("no engine for " + key)
	}
	e.Init()
	to := filepath.Join(e.slotRoot(), filepath.Base(dir))
	if err := os.MkdirAll(to, 0o755); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(to, f.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return e
}

// shippedWords is each shipped slot's world words as the engine saw them
// before they moved from the ruleset's records into the engine (recorded
// with RECORD_WORLD=1), in worldNames' order.
var shippedWords = map[string][len(worldNames)]int{
	"por/back-for-good": {18, 3, 13, 0, 0, 0, 0, 18, 0, 0, 0, 1, 3, 1, 255, 0, 0, 1, 60, 1, 0, 0, 0, 0, 0, 0, 0, 6, 0},
	"por/CAPTURE":       {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 0, 255, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"por/Hobgoblins":    {2, 4, 14, 0, 0, 0, 0, 2, 0, 0, 0, 1, 3, 1, 255, 0, 0, 4, 60, 1, 0, 0, 0, 0, 0, 0, 0, 6, 0},
	"por/in-Slums":      {20, 14, 4, 0, 0, 0, 0, 20, 0, 0, 0, 1, 3, 2, 255, 0, 0, 2, 100, 0, 0, 0, 0, 0, 0, 0, 0, 6, 0},
	"por/library":       {15, 11, 14, 0, 1, 0, 0, 15, 0, 0, 0, 1, 3, 1, 255, 0, 0, 2, 99, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"por/Podol":         {18, 12, 6, 0, 0, 0, 0, 18, 0, 0, 0, 1, 3, 2, 255, 0, 0, 1, 40, 0, 0, 0, 0, 0, 0, 2, 0, 14, 0},
	"por/SHOPPING":      {0, 13, 9, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 2, 42, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"por/Sokol Keep":    {21, 8, 5, 0, 1, 0, 0, 21, 0, 0, 0, 1, 3, 1, 255, 0, 0, 4, 50, 0, 0, 0, 0, 0, 0, 1, 0, 39, 0},
	"por/STARTER":       {0, 13, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 1, 255, 0, 0, 3, 50, 0, 0, 0, 0, 0, 0, 0, 0, 17, 0},
	"por/tavern":        {0, 15, 9, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 2, 255, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"por/weird-scribe":  {29, 1, 5, 0, 1, 0, 0, 32, 0, 0, 0, 1, 3, 2, 255, 0, 0, 8, 70, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"por/WORLDMAP":      {25, 15, 11, 0, 1, 0, 0, 18, 0, 0, 0, 1, 3, 0, 255, 0, 0, 6, 60, 0, 0, 0, 0, 0, 0, 2, 0, 3, 0},
	"sotsb/GAME A":      {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// TestShippedSlotsLoad: every shipped slot loads with the world words it
// always had, and a save written after that loads to the same words.
func TestShippedSlotsLoad(t *testing.T) {
	slots := shippedSlots(t)
	if len(slots) == 0 {
		t.Skip("no shipped slots")
	}
	record := os.Getenv("RECORD_WORLD") != ""
	for key, dir := range slots {
		t.Run(key, func(t *testing.T) {
			e := slotEngine(t, key, dir)
			if !e.loadFrom(filepath.Base(dir)) {
				t.Fatal("the slot did not load")
			}
			got := engineWords(e)
			if record {
				fmt.Printf("RECORD %q: %v,\n", key, strings.NewReplacer("[", "{", "]", "}", " ", ", ").Replace(fmt.Sprint(got)))
				return
			}
			if want, ok := shippedWords[key]; !ok || got != want {
				t.Errorf("words %v, want %v (recorded)", got, want)
			}
			e.gs.areaKind = byte(got[0]) // as GameLoop takes it from the world
			if err := e.saveTo("again"); err != nil {
				t.Fatal(err)
			}
			g, err := adventure.ReadSave(filepath.Join(e.slotRoot(), "again", adventure.GameFile))
			if err != nil {
				t.Fatal(err)
			}
			if g.World == nil {
				t.Fatal("a save written after the load has no World")
			}
			if !e.loadFrom("again") {
				t.Fatal("the new save did not load")
			}
			if again := engineWords(e); again != got {
				t.Errorf("after save and load: %v, want %v", again, got)
			}
		})
	}
}

// scriptWrite is a value that fits every field's width.
const scriptWrite = 0x5A

// TestWorldFieldsReachTheEngine: a script's write to a world word through its
// layout field name lands in the engine's world and reads back, and the
// ruleset's saved part holds none of the words.
func TestWorldFieldsReachTheEngine(t *testing.T) {
	e := quickStarted(t)
	e.loadScript(14)
	cells := e.world.Cells()
	bound := 0
	for name, cell := range cells {
		r, ok := e.layoutField(name)
		if !ok {
			continue // the layout has no such field: nothing binds
		}
		bound++
		e.host.Set(r, scriptWrite)
		if *cell != scriptWrite || e.host.Get(r) != scriptWrite {
			t.Errorf("%s: wrote %#x, the world holds %#x, the field reads %#x", name, scriptWrite, *cell, e.host.Get(r))
		}
	}
	if bound < 5 {
		t.Fatalf("only %d world words are fields of POR's layout", bound)
	}
	blob, err := e.ruleset().Save()
	if err != nil {
		t.Fatal(err)
	}
	var parts map[string]map[string]json.RawMessage
	if err := json.Unmarshal(blob, &parts); err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"Area", "Game"} {
		for name := range cells {
			if _, in := parts[part][name]; in {
				t.Errorf("the ruleset's %s still holds %s", part, name)
			}
		}
	}
}
