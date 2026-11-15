package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/adventure"
	"gobox/rules"
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

// placeNames are the place, rest and training words (their script field names), in the order of
// shippedPlaceWords.
var placeNames = [...]string{"NoMagic", "In3D", "MapX", "MapY", "Boss", "MarkFlags4", "Mark3", "Mark4Off", "MarkFlags57", "Mark4",
	"Sky", "SkyInside", "RestEvery", "RestChance", "PartyExp"}

// enginePlaceWords are the place words as the engine sees them, in placeNames' order.
func enginePlaceWords(e *Engine) [len(placeNames)]int {
	return [...]int{
		int(e.world.NoMagic), int(e.world.In3D), int(e.world.MapX), int(e.world.MapY), int(e.world.Boss),
		int(e.world.MarkFlags4), int(e.world.Mark3), int(e.world.Mark4Off), int(e.world.MarkFlags57), int(e.world.Mark4),
		int(e.world.Sky), int(e.world.SkyInside), int(e.world.RestEvery), int(e.world.RestChance), int(e.world.PartyExp),
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

// shippedPlaceWords is each shipped slot's place, rest and training words as the engine saw them
// before they moved from the ruleset's records into the engine (recorded with RECORD_WORLD=1),
// in placeNames' order.
var shippedPlaceWords = map[string][len(placeNames)]int{
	"por/back-for-good": {0, 1, 13, 27, 0, 0, 0, 0, 0, 0, 11, 10, 0, 0, 114},
	"por/CAPTURE":       {0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	"por/Hobgoblins":    {0, 1, 13, 27, 0, 0, 0, 0, 0, 0, 11, 8, 0, 0, 114},
	"por/in-Slums":      {0, 1, 13, 27, 0, 0, 0, 0, 0, 0, 11, 9, 0, 0, 116},
	"por/library":       {0, 1, 13, 27, 0, 0, 0, 0, 0, 0, 11, 9, 0, 0, 120},
	"por/Podol":         {0, 1, 13, 27, 0, 0, 0, 0, 0, 0, 8, 10, 0, 0, 114},
	"por/SHOPPING":      {0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 11, 10, 0, 0, 0},
	"por/Sokol Keep":    {0, 1, 13, 27, 0, 0, 0, 0, 0, 0, 11, 10, 0, 0, 120},
	"por/STARTER":       {0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 11, 10, 0, 0, 116},
	"por/tavern":        {0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 11, 10, 0, 0, 0},
	"por/weird-scribe":  {0, 1, 13, 27, 0, 0, 0, 0, 0, 0, 11, 8, 0, 100, 120},
	"por/WORLDMAP":      {0, 0, 4, 31, 0, 0, 0, 0, 0, 0, 11, 10, 0, 0, 114},
	"sotsb/GAME A":      {0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
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
				fmt.Printf("RECORDSESSIONSLOT %q: %v,\n", key, strings.NewReplacer("[", "{", "]", "}", " ", ", ").Replace(fmt.Sprint(sessionSlotWords(e))))
				fmt.Printf("RECORD %q: %v,\n", key, strings.NewReplacer("[", "{", "]", "}", " ", ", ").Replace(fmt.Sprint(got)))
				fmt.Printf("RECORDPLACE %q: %v,\n", key, strings.NewReplacer("[", "{", "]", "}", " ", ", ").Replace(fmt.Sprint(enginePlaceWords(e))))
				return
			}
			if want, ok := shippedPlaceWords[key]; !ok || enginePlaceWords(e) != want {
				t.Errorf("place words %v, want %v (recorded)", enginePlaceWords(e), want)
			}
			if want, ok := shippedSessionWords[key]; !ok || sessionSlotWords(e) != want {
				t.Errorf("session words %v, want %v (recorded)", sessionSlotWords(e), want)
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
			if again := sessionSlotWords(e); again != shippedSessionWords[key] {
				t.Errorf("after save and load, session words %v, want %v", again, shippedSessionWords[key])
			}
			if again := enginePlaceWords(e); again != shippedPlaceWords[key] {
				t.Errorf("after save and load, place words %v, want %v", again, shippedPlaceWords[key])
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

// A save of any step of the words' move loads whole: a word the ruleset's
// part still has comes from there, the others from the save's World (bug
// class: a save written between two steps lost the words moved later).
func TestLoadWorldMergesOldBlob(t *testing.T) {
	e := &Engine{}
	e.loadWorld(adventure.Save{
		World: &rules.World{Geo: 5, Morale: 1},
		Rules: json.RawMessage(`{"Area":{"NoMagic":1,"Clock":[0,0,3,4,0,0,0]},"Game":{"Morale":60,"RestEvery":7},"Purse":{}}`),
	})
	if w := e.world; w.Geo != 5 || w.NoMagic != 1 || w.Morale != 60 || w.RestEvery != 7 {
		t.Errorf("loaded %+v, want Geo 5 (World), NoMagic 1, Morale 60, RestEvery 7 (the ruleset's part)", w)
	}
	e.loadWorld(adventure.Save{World: &rules.World{Geo: 9}})
	if e.world != (rules.World{Geo: 9}) {
		t.Errorf("a save with no ruleset part: %+v", e.world)
	}
}

// sessionNames are the words of the game's session, in the order of shippedSessionWords.
var sessionNames = [...]string{"minuteOfDay", "partyCount", "phase"}

// sessionSlotWords are the clock's minute of the day, the party's count and the game's phase.
func sessionSlotWords(e *Engine) [len(sessionNames)]int {
	return [...]int{obsMinute(e), obsCount(e), obsPhase(e)}
}

// shippedSessionWords is each shipped slot's session words as the engine saw them before the
// generic words left the ruleset (recorded with RECORD_WORLD=1), in sessionNames' order.
var shippedSessionWords = map[string][len(sessionNames)]int{
	"por/CAPTURE":       {0, 6, 4},
	"por/Hobgoblins":    {840, 8, 4},
	"por/Podol":         {263, 8, 4},
	"por/SHOPPING":      {426, 6, 4},
	"por/STARTER":       {551, 6, 4},
	"por/Sokol Keep":    {778, 6, 4},
	"por/WORLDMAP":      {465, 8, 3},
	"por/back-for-good": {818, 8, 4},
	"por/in-Slums":      {1113, 8, 4},
	"por/library":       {1115, 8, 4},
	"por/tavern":        {442, 6, 4},
	"por/weird-scribe":  {797, 6, 4},
	"sotsb/GAME A":      {0, 6, 0},
}
