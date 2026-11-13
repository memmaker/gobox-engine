package engine

import (
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"gobox/asset"
	"gobox/game/play"
	"gobox/rules"
	"gobox/rules1e"
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSlotName(t *testing.T) {
	for in, want := range map[string]string{"Before Boss": "Before Boss", "../x/y": "xy", " a:B ": "aB", "": ""} {
		if got := asset.Name(in); got != want {
			t.Errorf("saves.Name(%q) = %q, want %q", in, got, want)
		}
	}
}

// An F9 save leaves the area block as it was: saveSlot packs the wall slots
// into area+0x3F4.., and area+0x3FE is the view animation flag.
func TestSaveToKeepsArea(t *testing.T) {
	e := &Engine{set: 1}
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	e.data, e.adv = "../data/por/sources", "../data/por/sources/assets" // the fields' layouts
	e.mapPairB = [3]int{5, 6, 7}
	e.saveTo("X")
	if w := e.state().Area.WallsB[2]; w != 0 {
		t.Errorf("area+0x3FE = %d after a save, want 0", w)
	}
	if _, err := os.Stat(filepath.Join(e.slotRoot(), "X", gameFile)); err != nil {
		t.Error("no save written")
	}
	if l := e.slots(); len(l) != 1 || time.Since(l[0].At) > time.Minute {
		t.Errorf("slots = %v, want X at the save's time", l)
	}
}

// A save in the first command input after an area change names the area
// the party is in, not the one it left (the area block still does then).
func TestSaveToRunningArea(t *testing.T) {
	e := &Engine{set: 1, data: "../data/por/sources", adv: "../data/por/sources/assets"} // the save needs the script layout
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	e.state().Area.AreaKind, e.gs.areaKind = 20, 0
	e.saveTo("X")
	g, err := readGameSave(filepath.Join(e.slotRoot(), "X", gameFile))
	if a := blobOf(t, g).Area; err != nil || a.AreaKind != 0 {
		t.Errorf("saved area %d (%v), want 0, the running one", a.AreaKind, err)
	}
}

// A slot holding only an original save is listed once imported, and loads
// the original's state.
func TestImportSlot(t *testing.T) {
	e := &Engine{set: 1, data: "../data/por/sources", adv: "../data/por/sources/assets"}
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	f := rules1e.OriginalSave{Set: 1, X: 7, Y: 9}
	f.Area.AreaKind = 20
	f.Game.Shop = 1
	b, err := binary.Append(nil, binary.LittleEndian, &f)
	if err != nil {
		t.Fatal(err)
	}
	b = rules1e.AppendPurse(b, rules1e.Purse{Gold: 12})
	dir := filepath.Join(e.slotRoot(), "OLD")
	os.MkdirAll(dir, 0o755)
	if err := os.WriteFile(filepath.Join(dir, origSaveFile), b, 0o644); err != nil {
		t.Fatal(err)
	}
	if l := e.slots(); len(l) != 1 || l[0].Name != "OLD" {
		t.Fatalf("slots = %v, want the imported OLD", l)
	}
	g, err := readGameSave(filepath.Join(dir, gameFile))
	if err != nil {
		t.Fatal(err)
	}
	if r := blobOf(t, g); g.X != 7 || g.Y != 9 || r.Area.AreaKind != 20 || r.Game.Shop != 1 || r.Purse.Gold != 12 {
		t.Errorf("imported %+v (%v)", g, err)
	}
}

// The named words sit at the offsets the listings use, as the scripts
// reach them (word i = byte offset 2i), and each record is 0x800 bytes.
func TestVarOffsets(t *testing.T) {
	var a rules1e.AreaVars
	var g rules1e.GameVars
	for off, f := range map[int]*uint16{
		0x186: &a.MapX, 0x18C: &a.Clock[0], 0x1CA: &a.NoMagic, 0x1E0: &a.PrevX, 0x1F6: &a.NoCoords,
		0x2F8: &a.MarkFlags4, 0x318: &a.Mark3, 0x33E: &a.Mark4Off, 0x344: &a.RiverOpen, 0x366: &a.Boss,
		0x3A4: &a.Mark4, 0x3F4: &a.WallsA[0], 0x3FA: &a.WallsB[0], 0x400: &a.V4B00[0],
	} {
		if a.Word(off/2) != f {
			t.Errorf("area word at %#x is not the named field", off)
		}
	}
	for off, f := range map[int]*uint16{
		0x550: &g.PartyExp, 0x580: &g.MaxDist, 0x58C: &g.Morale, 0x5A4: &g.RestEvery, 0x5AA: &g.LeftMap,
		0x5C2: &g.Portrait, 0x5CC: &g.Duel, 0x624: &g.SymbolSet, 0x666: &g.AttackedAlly,
		0x67C: &g.PartyCount, 0x6D8: &g.Shop, 0x6E0: &g.MonsterHit, 0x6E4: &g.PartyMove,
	} {
		if g.Word(off/2) != f {
			t.Errorf("game word at %#x is not the named field", off)
		}
	}
	if n, m, k := len(recordBytes(&a)), len(recordBytes(&g)), len(recordBytes(&rules1e.AuxVars{})); n != 0x800 || m != 0x800 || k != 0x400 {
		t.Errorf("record sizes %#x %#x %#x, want 0x800 0x800 0x400", n, m, k)
	}
}

// The explored squares go with the slot: a load brings back the save's, not
// what was explored since.
func TestSeenFollowsSlot(t *testing.T) {
	e := &Engine{set: 1, data: "../data/por/sources", adv: "../data/por/sources/assets"} // the save needs the script layout
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	e.ext.seen = map[int]*[256]uint8{3: {5: seenVisited}}
	e.saveTo("X")
	b, err := os.ReadFile(filepath.Join(e.slotRoot(), "X", seenFile))
	if err != nil || len(b) == 0 {
		t.Fatal("no explored squares saved")
	}
	e.ext.seen = map[int]*[256]uint8{3: {5: seenVisited, 6: seenSeen}, 9: {}}
	f, _ := os.Open(filepath.Join(e.slotRoot(), "X", seenFile))
	e.ext.seen = nil
	gob.NewDecoder(f).Decode(&e.ext.seen)
	f.Close()
	if len(e.ext.seen) != 1 || e.ext.seen[3][5] != seenVisited || e.ext.seen[3][6] != 0 {
		t.Errorf("seen after reload = %v", e.ext.seen)
	}
}

// A slot saved with the explored squares by script area loads them by map:
// New Phlan's district (area 0) and Training Hall (11) pool on set 3's map 0.
func TestSeenByMap(t *testing.T) {
	e := &Engine{set: 1, data: "../data/por/sources", adv: "../data/por/sources/assets"}
	e.initVM()
	seen := e.seenByMap(map[int]*[256]uint8{0: {5: seenVisited, 6: seenSeen}, 11 << 16: {6: seenVisited, 7: seenSeen}, 20<<16 | 20: {1: seenVisited}})
	want := map[int]*[256]uint8{3 << 16: {5: seenVisited, 6: seenVisited, 7: seenSeen}, 2<<16 | 20: {1: seenVisited}}
	if len(seen) != len(want) {
		t.Fatalf("%d maps, want %d", len(seen), len(want))
	}
	for k, m := range want {
		if got := seen[k]; got == nil || *got != *m {
			t.Errorf("map %#x = %v, want %v", k, got, m)
		}
	}
}

// labels are the current map's labels (square y<<4|x -> label): the
// built-in ones with the player's over them.
func (e *Engine) labels() map[int]mapLabel {
	m := map[int]mapLabel{}
	maps.Copy(m, e.builtinLabels(e.areaKey()))
	maps.Copy(m, e.ext.labels[e.areaKey()])
	return m
}

// New Phlan (set 3, GEO 0) always has its built-in labels; the player's go
// over them and with the slot (built-in copies are not written), a new
// adventure starts without them, and removing one shows the built-in again.
func TestPhlanLabels(t *testing.T) {
	e := &Engine{set: 3, data: "../data/por/sources", adv: "../data/por/sources/assets"} // the save needs the script layout
	e.initVM()
	e.Cfg.SaveDir = t.TempDir()
	if m := e.labels(); m[1<<4|15].text != "Boats" || len(m) != len(e.builtinLabels(3<<16)) {
		t.Fatalf("fresh New Phlan: %d labels, 15,1 = %q", len(m), m[1<<4|15].text)
	}
	e.userLabels()[1<<4|15] = mapLabel{"My boat", poiUser}
	e.userLabels()[1<<4|11] = e.builtinLabels(3 << 16)[1<<4|11]
	e.saveTo("X")
	path := filepath.Join(e.slotRoot(), "X", labelsFile)
	if b, _ := os.ReadFile(path); string(b) != "map 196608\n15 1 0 My boat\n" {
		t.Errorf("saved labels = %q", b)
	}
	e.resetAdventure()
	if l := e.labels()[1<<4|15]; l.text != "Boats" {
		t.Errorf("new adventure's 15,1 = %q, want Boats", l.text)
	}
	e.readLabels(path)
	if m := e.labels(); m[1<<4|15].text != "My boat" || m[4<<4|0].text != "Gate to slums" {
		t.Fatalf("loaded 15,1 = %q, 0,4 = %q", m[1<<4|15].text, m[4<<4|0].text)
	}
	delete(e.userLabels(), 1<<4|15)
	if l := e.labels()[1<<4|15]; l.text != "Boats" {
		t.Errorf("removed 15,1 = %q, want Boats", l.text)
	}
}

// The full map is titled with its map's name from names/pictures.txt.
func TestMapName(t *testing.T) {
	e := &Engine{set: 3, data: "../data/por/sources", adv: "../data/por/sources/assets"}
	e.initVM()
	if n := e.mapName(); n != "New Phlan" {
		t.Errorf("set 3 GEO 0 = %q, want New Phlan", n)
	}
	e.set, e.state().Area.Geo = 5, 3
	if n := e.mapName(); n != "Valjevo Castle 3" {
		t.Errorf("set 5 GEO 3 = %q, want Valjevo Castle 3", n)
	}
	e.set, e.state().Area.Geo = 8, 32
	if n := e.mapName(); n != "Well Bottom" {
		t.Errorf("set 8 GEO 32 = %q, want Well Bottom", n)
	}
	e.state().Area.Geo = 99
	if n := e.mapName(); n != "Map" {
		t.Errorf("unnamed map = %q, want Map", n)
	}
}

// A loaded game may Bash, Pick and Knock again: with none allowed the door
// menu is only Exit, which walls the door up.
func TestLoadAllowsDoorTries(t *testing.T) {
	e := quickStarted(t)
	e.saveTo("X")
	e.run.doorTry = [3]byte{}
	if !e.loadFrom("X") || e.run.doorTry != [3]byte{1, 1, 1} {
		t.Fatalf("doorTry %v after a load, want all allowed", e.run.doorTry)
	}
}

// A locked door is the wait PendDoor : TryDoor answers it and
// the step goes on with the outcome; an option not offered is refused.
func TestDoorWait(t *testing.T) {
	e := quickStarted(t)
	e.run.doorTry = [3]byte{1, 0, 0}
	e.ext.doorThrough = true
	e.lockedDoor(false)
	if d := e.ext.door.opts; e.pending != PendDoor || !d.Bash || d.Pick || e.top() == onAsk {
		t.Fatalf("door %+v pending %d", d, e.pending)
	}
	if e.TryDoor(play.DoorPick) == nil {
		t.Error("Pick taken where it is not offered")
	}
	if err := e.TryDoor(play.DoorExit); err != nil || e.ext.doorThrough || e.pending != PendNone {
		t.Errorf("Exit: err %v through %v pending %d", err, e.ext.doorThrough, e.pending)
	}
}

// recordBytes is a record (a pointer to areaVars, gameVars or auxVars) in
// its file layout, little-endian word for word.
func recordBytes(rec any) []byte {
	b, err := binary.Append(nil, binary.LittleEndian, rec)
	if err != nil {
		panic(err)
	}
	return b
}

// blobOf is g's ruleset part.
func blobOf(t *testing.T, g gameSave) rules1e.SaveBlob {
	t.Helper()
	var b rules1e.SaveBlob
	if err := json.Unmarshal(g.Rules, &b); err != nil {
		t.Fatal(err)
	}
	return b
}

// A save written before the ruleset's part was a blob (Area, Game, Purse at
// the top level) loads unchanged.
func TestLegacySaveLoads(t *testing.T) {
	var old rules1e.SaveBlob
	old.Area.AreaKind, old.Area.Geo, old.Game.Shop, old.Purse.Gold = 20, 3, 1, 1<<40
	b, err := json.Marshal(struct {
		Set rules.Set
		rules1e.SaveBlob
		X byte
	}{2, old, 7})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), gameFile)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	g, err := readGameSave(path)
	if err != nil {
		t.Fatal(err)
	}
	var s rules1e.RuleState
	if err := (&rules1e.Rules{S: &s}).Load(g.Rules); err != nil {
		t.Fatal(err)
	}
	if g.Set != 2 || g.X != 7 || s.Area != old.Area || s.Game != old.Game || s.Purse != old.Purse {
		t.Errorf("loaded set %d x %d %+v, want the old save's", g.Set, g.X, s.Purse)
	}
}

// TestMapLabelsPerGame: a game without MapLabels shows none on set 3, GEO 0
// (POR's New Phlan labels were on every game's map there).
func TestMapLabelsPerGame(t *testing.T) {
	e := &Engine{set: 3, given: &GameTables{}}
	if len(e.builtinLabels(e.areaKey())) != 0 || e.areaSong() != 0 {
		t.Fatal("another game's area 3/0 has POR's labels or songs")
	}
}
