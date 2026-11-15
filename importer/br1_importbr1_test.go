package importer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/asset/orig"
	"gobox/rules"
	"gobox/rules/xxvc"
	"gobox/rulesxxvc"
)

// TestImportEqualsDecode imports the originals in ../data and checks the
// folder against the originals: the art equals data/assets/art
// (app.ImportArt), the maps and records equal the archives' blocks, the
// rule tables equal ITEMS.
func TestBR1ImportEqualsDecode(t *testing.T) {
	dos := filepath.Join("..", "data", "br1", "sources", "dos")
	l, err := orig.NewLoader(dos)
	if err != nil || !l.Exists("ITEMS") {
		t.Skip("no originals:", err)
	}
	root := t.TempDir()
	if err := Run("br1", dos, root); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "br1")
	if err := CheckContents(dst); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"scripts", "art"} {
		if es, _ := os.ReadDir(filepath.Join(dst, d)); len(es) == 0 {
			t.Fatal(d, "empty")
		}
	}
	// The decode goes to a folder of the test's own, never asset.Dir(dos): that is the repo's art/, which
	// ImportArtIfStale empties first (it deleted this repo's 46 tracked pictures on every run).
	decoded := filepath.Join(t.TempDir(), "art")
	if err := orig.ImportArtIfStale(dos, filepath.Join(filepath.Dir(dos), "amiga"), decoded); err != nil {
		t.Fatal(err)
	}
	if err := SameTree(filepath.Join(dst, "art"), decoded); err != nil {
		t.Error("art:", err)
	}
	data := filepath.Join(dst, "data")
	geo, err := adventure.ReadBlocks(data, "GEO1")
	if err != nil || len(geo) == 0 {
		t.Fatal("GEO1:", err)
	}
	if items, err := adventure.ReadBlocks(data, ItemsFile); err != nil || !bytes.Equal(items[0], brReadFile(t, l.Path("ITEMS"))) {
		t.Error("ITEMS differs:", err)
	}
	ships, err := adventure.ReadBlocks(data, "SHIPS")
	if err != nil || len(ships) == 0 || !bytes.Equal(ships[50], l.BlockShared("SHIPS", 50)) {
		t.Error("SHIPS 50 differs:", err)
	}
	if !bytes.Equal(brReadFile(t, filepath.Join(dst, SavesDir, "CHRDATA1.SAV")), brReadFile(t, l.Path("CHRDATA1.SAV"))) {
		t.Error("saves: CHRDATA1.SAV differs")
	}
	var mons map[int]BRMonster
	if _, err := adventure.ReadJSON(data, "MON0", &mons); err != nil || len(mons) == 0 {
		t.Fatal("MON0:", len(mons), err)
	}
	for id := range 256 {
		if b := l.BlockShared("GEO1", id); !bytes.Equal(b, geo[id]) {
			t.Errorf("GEO1 %d differs", id)
		}
		b := l.BlockShared("MON0CHA", id)
		if b == nil {
			continue
		}
		m := mons[id]
		var want BRRecord
		copy(want[:], b)
		if m.Record != want {
			t.Errorf("MON0 %d: record differs", id)
		}
		var got []byte
		for _, it := range m.Items {
			got = append(got, it[:]...)
		}
		if w := l.BlockShared("MON0ITM", id); !bytes.Equal(got, w[:len(w)/len(BRItem{})*len(BRItem{})]) {
			t.Errorf("MON0 %d: items differ", id)
		}
	}
	var tab xxvc.Tables
	if _, err := adventure.ReadJSON(data, TablesFile, &tab); err != nil {
		t.Fatal(err)
	}
	items, _ := os.ReadFile(l.Path("ITEMS"))
	var raw [itemTypes * itemTypeSize]byte
	copy(raw[:], items)
	for i, it := range tab.ItemTypes {
		if int(it.Slot) != int(raw[i*itemTypeSize+0]) || int(it.AttacksDoubled) != int(raw[i*itemTypeSize+1]) || int(it.Damage.N) != int(raw[i*itemTypeSize+4]) || int(it.Damage.Sides) != int(raw[i*itemTypeSize+5]) || int(it.Range) != int(raw[i*itemTypeSize+7]) {
			t.Errorf("item type %d: %+v, ITEMS % x", i, it, raw[i*itemTypeSize:(i+1)*itemTypeSize])
		}
	}
	// itemNameParts (gdconsts.pas): 44 words from 'Knife' to 'mini'.
	if len(tab.ItemWords) != 44+1 || tab.ItemWords[0] != "" || tab.ItemWords[1] != "Knife" || tab.ItemWords[44] != "mini" {
		t.Errorf("item words %q", tab.ItemWords)
	}
	// ammoPrices (gdconsts.pas): types 5..18.
	if w := tab.ShipWeapons; len(w) != 8 || w[1] != (xxvc.ShipWeapon{MaxRange: 6, Damage: 10, TargetBonus: 1, Name: "Beam laser", Short: "Laser"}) ||
		w[8] != (xxvc.ShipWeapon{MaxRange: 2, ReloadAmmo: 10, Damage: 30, TargetBonus: -2, Name: "Heavy acceleration gun", Short: "hv-Accel"}) {
		t.Error("ShipWeapons:", w)
	}
	if s := tab.Ships; len(s) != 6 || s[xxvc.PartyShip].Name != "Maelstrom Rider" || s[xxvc.PartyShip].Type != "Md cruiser" ||
		s[xxvc.PartyShip].Mounts[0] != (xxvc.ShipMount{Kind: 6, MaxStatus: 2, Status: 2, Ammo: 6, Reloads: 5}) ||
		s[40].Maneuver != 254 || s[40].Race != xxvc.Martian || !s[42].Mounts[0].Crew || s[42].Crew != 135 {
		t.Error("Ships:", s)
	}
	if tab.Explosion != xxvc.ExplosionRolled {
		t.Error("Explosion:", tab.Explosion)
	}
	if a := tab.AmmoPrices; len(a) != 14 || a[5] != 10 || a[6] != 5 || a[8] != 0 || a[11] != 100 || a[18] != 100 {
		t.Errorf("ammo prices %v", a)
	}
	if len(tab.Monsters) != len(mons) {
		t.Errorf("%d typed monsters, %d records", len(tab.Monsters), len(mons))
	}
	for id, m := range mons {
		tm := tab.Monsters[rules.MonsterID(id)]
		if tm.Name != m.Record.Name() || int(tm.Level) != int(m.Record[BROffLevel]) || len(tm.Items) != len(m.Items) {
			t.Errorf("monster %d: %+v, record %q level %d, %d items", id, tm, m.Record.Name(), m.Record[BROffLevel], len(m.Items))
		}
	}
	// MON0CHA stores base values (currents 0): SWAMP HORNET (33) moves 9, AC 52, unarmed 1d6.
	if h := tab.Monsters[33]; h.Name != "SWAMP HORNET" || h.Movement != 9 || h.ArmorClass[0] != 52 || h.Damage[0] != (rules.Dice{N: 1, Sides: 6}) || !h.Hostile {
		t.Errorf("monster 33: %+v", h)
	}
	// The originals' saved characters (CHRDATA1.SAV, .STF): decoded and converted (rulesxxvc.ConvertBRCharacter).
	var c rulesxxvc.Character
	if _, err := adventure.ReadJSON(filepath.Join(data, CharactersDir), "CHRDATA1", &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "FLAVIUS" || c.Level != 2 || c.MaxHP != 26 || len(c.Items) != 2 {
		t.Errorf("CHRDATA1: %+v", c)
	}
	tmpl := l.BlockShared("ITEM0", 1)
	if len(tab.Items) == 0 {
		t.Fatal("no item templates")
	}
	for id, it := range tab.Items {
		if int(id)*16 > len(tmpl) {
			continue // a monster's own item (addTemplates), past ITEM0's
		}
		r := tmpl[(int(id)-1)*16:]
		if int(it.Type) != int(r[0]) || int(it.Words[0]) != int(r[1]) || int(it.Quantity) != int(r[10]) {
			t.Errorf("template %d: %+v, ITEM0 % x", id, it, r[:16])
		}
	}
}

func brReadFile(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
