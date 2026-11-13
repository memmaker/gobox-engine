package importpod

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"gobox/audio"
	"gobox/char"
	"gobox/ecl/script"
	"gobox/engine"
	"gobox/game/journal"
	"gobox/rules"
	"gobox/rules1e"
)

// TestSwitchData: the folder writeSwitchData writes loads through the shared
// engine's readers (pool.ReadGameTables, rules1e.ReadLayout, script.ReadEntries,
// journal.Load) with the values pod's own engine uses (importpod).
func TestSwitchData(t *testing.T) {
	dst := t.TempDir()
	if err := writeSwitchData(dst); err != nil {
		t.Fatal(err)
	}
	g, err := engine.ReadGameTables(dst)
	if err != nil {
		t.Fatal(err)
	}
	d, err := engine.ReadGameData(dst)
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "pod" || d.Font != "Pools_of_Darkness" || len(d.QuickParty) == 0 || d.AdLib == nil || len(d.Rests) != 2 {
		t.Errorf("game data %+v", d)
	}
	snd, err := os.ReadFile(filepath.Join(dst, engine.DataDir, engine.SoundFile))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := audio.NewDriver(snd, 'P'); err != nil { // AdLib only: the tables are silent, but valid
		t.Errorf("sound.json: %v", err)
	}
	if g.Script != script.V3 || g.Dialect != Dialect || !g.AddNPCIcon {
		t.Errorf("dialect %v script %p AddNPCIcon %v", g.Dialect, g.Script, g.AddNPCIcon)
	}
	if got := g.Script.Behaviour.NoMonsterCombat; !reflect.DeepEqual(got, []script.CombatScreen{script.CombatVault, script.CombatTemple}) || !g.Script.Behaviour.AreaMonstersJoin {
		t.Errorf("behaviour %+v", g.Script.Behaviour)
	}
	monsters, icons, count, status, control := AreaMonsterTable()
	a := g.AreaMonsters
	if a.Count != count || a.Status != status || a.Control != control || len(a.ByScreen) != len(monsters) {
		t.Fatalf("area monsters %+v", a)
	}
	for i, m := range a.ByScreen {
		if m.Monster != int(monsters[i]) || m.Icon != int(icons[i]) {
			t.Errorf("screen %d: %+v", i, m)
		}
	}
	for n := 0; n <= 255; n++ {
		song, music, digitized := SoundKind(byte(n))
		s, ok := g.ScriptSounds[byte(n)]
		want := engine.ScriptSound{Music: music, Song: song}
		if digitized {
			want.Cue = rules.SoundCue(strconv.Itoa(n))
		}
		if ok != (music || digitized) || (ok && s != want) {
			t.Errorf("sound %d: %+v (%v), want %+v", n, s, ok, want)
		}
	}
	data := filepath.Join(dst, engine.DataDir)
	l, err := rules1e.ReadLayout(filepath.Join(data, rules1e.LayoutFile))
	if err != nil || !reflect.DeepEqual(l, Layout1e()) {
		t.Errorf("layout %+v (%v), want %+v", l, err, Layout1e())
	}
	char.Rules1e("", nil, nil, nil, nil) // the ruleset the layout and the tables go to builds

	es, err := script.ReadEntries(os.DirFS(filepath.Join(dst, "scripts")))
	if err != nil {
		t.Fatal(err)
	}
	book, err := journal.Load(os.DirFS(filepath.Join(dst, engine.JournalDir)), "JOURNAL.TXT")
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
	for lvl, ids := range CreationStart("").MageSpells.BySlot {
		for _, id := range ids {
			if !slices.ContainsFunc(slices.Collect(maps.Values(rules1e.PodSpellIDs)), func(v rules1e.SpellID) bool { return v == id }) {
				t.Errorf("level %d: id %d has no pod number", lvl, id)
			}
		}
	}
}
