package importer

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"gobox/adventure"
	"gobox/audio"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
)

// TestSwitchData: the folder writeSwitchData writes loads through the shared
// engine's readers (pool.ReadGameTables, rules1e.ReadLayout, ecl.ReadEntries,
// game.Load) with the values pod's own engine uses (the pod import).
func TestSwitchData(t *testing.T) {
	dst := t.TempDir()
	if err := writeSwitchData(dst); err != nil {
		t.Fatal(err)
	}
	g, err := adventure.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	d, err := adventure.ReadGameData(dst)
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "pod" || len(d.QuickParty) == 0 || d.AdLib == nil || len(d.Rests) != 2 {
		t.Errorf("game data %+v", d)
	}
	snd, err := os.ReadFile(filepath.Join(dst, adventure.DataDir, adventure.SoundFile))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := audio.NewDriver(snd, 'P'); err != nil { // AdLib only: the tables are silent, but valid
		t.Errorf("sound.json: %v", err)
	}
	if g.Script != ecl.V3 || g.Dialect != Dialect || !g.AddNPCIcon {
		t.Errorf("dialect %v script %p AddNPCIcon %v", g.Dialect, g.Script, g.AddNPCIcon)
	}
	if got := g.Script.Behaviour.NoMonsterCombat; !reflect.DeepEqual(got, []ecl.CombatScreen{ecl.CombatVault, ecl.CombatTemple}) {
		t.Errorf("behaviour %+v", g.Script.Behaviour)
	}
	monsters, icons, count, control := ReserveTable()
	a := g.Reserve
	if a.Count != count || a.Control != control || len(a.BySet) != len(monsters) {
		t.Fatalf("area monsters %+v", a)
	}
	for i, m := range a.BySet {
		if m.Monster != int(monsters[i]) || m.Icon != int(icons[i]) {
			t.Errorf("screen %d: %+v", i, m)
		}
	}
	for n := 0; n <= 255; n++ {
		song, music, digitized := SoundKind(byte(n))
		s, ok := g.ScriptSounds[byte(n)]
		want := adventure.ScriptSound{Music: music, Song: song}
		if digitized {
			want.Cue = rules.SoundCue(strconv.Itoa(n))
		}
		if ok != (music || digitized) || (ok && s != want) {
			t.Errorf("sound %d: %+v (%v), want %+v", n, s, ok, want)
		}
	}
	data := filepath.Join(dst, adventure.DataDir)
	l, err := rules1e.ReadLayout(filepath.Join(data, rules1e.LayoutFile))
	if err != nil || !reflect.DeepEqual(l, PodLayout1e()) {
		t.Errorf("layout %+v (%v), want %+v", l, err, PodLayout1e())
	}
	rules1e.Rules1e(noSource{}, nil) // the ruleset the layout and the tables go to builds

	es, err := ecl.ReadEntries(podScriptData()) // the entries are the script import's (ImportScripts), not writeSwitchData's
	if err != nil {
		t.Fatal(err)
	}
	book, err := game.LoadJournal(os.DirFS(filepath.Join(dst, adventure.JournalDir)), "JOURNAL.TXT")
	if err != nil {
		t.Fatal(err)
	}
	want := JournalEntries()
	if len(es) != len(want) || len(es) == 0 {
		t.Fatalf("%d entries, want %d", len(es), len(want))
	}
	for _, e := range es {
		if e.ID == 13 || !book.Has(e.Title) {
			t.Errorf("entry %+v: pod has no entry 13, every entry has its text", e)
		}
	}
}

// TestStartSpellsAreIDs: creationStart's mage spells are unified ids that
// pod numbers (a game number written as an id broke Quick Start's save).
func TestStartSpellsAreIDs(t *testing.T) {
	for lvl, ids := range PodCreationStart("").MageSpells.BySlot {
		for _, id := range ids {
			if !slices.ContainsFunc(slices.Collect(maps.Values(rules1e.PodSpellIDs)), func(v rules1e.SpellID) bool { return v == id }) {
				t.Errorf("level %d: id %d has no pod number", lvl, id)
			}
		}
	}
}
