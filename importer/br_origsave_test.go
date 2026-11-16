package importer

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/asset/orig"
	"gobox/ecl"
	"gobox/rules"
	"gobox/rulesxxvc"
)

var brSavedGames = []struct {
	id  string
	exe BREXE
}{{"br1", BR1EXE}, {"br2", BR2EXE}}

// brSavedGame reads game's shipped SAVGAMA.DAT with the data segment the import reads the ship type names from.
func brSavedGame(t *testing.T, game string, exe BREXE) (b, ds []byte) {
	t.Helper()
	dos := filepath.Join("..", "data", game, "sources", "dos")
	b, err := os.ReadFile(filepath.Join(dos, "SAVGAMA.DAT"))
	if err != nil {
		t.Skip("no originals:", err)
	}
	im, err := orig.Load(filepath.Join(dos, "START.EXE"), exe.DS)
	if err != nil {
		t.Fatal(err)
	}
	return b, im.Mem[im.DS:]
}

// brFileOffset is the byte of script address a in the file (saveload.pas SaveGame's order: the set, zone 0, zone 1, zone 2).
func brFileOffset(a int) int {
	switch {
	case a >= brZone0Base && a < brZone0Base+brZone0Size:
		return 1 + a - brZone0Base
	case a >= brZone1Base && a < brZone1Base+brZone1Size:
		return 1 + brZone0Size + a - brZone1Base
	case a >= brZone2Base && a < brZone2Base+brZone2Size:
		return 1 + brZone0Size + brZone1Size + a - brZone2Base
	}
	panic("address in no zone")
}

// brSetCell sets the layout field name of the file b to v (a byte, a word or a dword little-endian; a text from its first byte).
func brSetCell(t *testing.T, l *ecl.Layout, b []byte, name string, v ...int) {
	t.Helper()
	f, ok := brField(l, name)
	if !ok {
		t.Fatalf("no layout field %s", name)
	}
	at := brFileOffset(f.From)
	switch f.Width {
	case ecl.WidthWord:
		binary.LittleEndian.PutUint16(b[at:], uint16(v[0]))
	case ecl.WidthDword:
		binary.LittleEndian.PutUint32(b[at:], uint32(v[0]))
	default:
		for i, c := range v {
			b[at+i] = byte(c)
		}
	}
}

// TestBRSavedGameShipped: the shipped saved game of both games imports: the six members, the position (7,13) facing 0 on the 3D
// view, the data set and the world's words, the clock at its start, the party's ship.
func TestBRSavedGameShipped(t *testing.T) {
	for _, g := range brSavedGames {
		t.Run(g.id, func(t *testing.T) {
			b, ds := brSavedGame(t, g.id, g.exe)
			s, err := ImportBRSave(b, embeddedScripts(g.id), ds, g.exe)
			if err != nil {
				t.Fatal(err)
			}
			if want := []string{"CHRDATA1", "CHRDATA2", "CHRDATA3", "CHRDATA4", "CHRDATA5", "CHRDATA6"}; !slices.Equal(s.Party, want) {
				t.Errorf("party %v, want %v", s.Party, want)
			}
			if s.X != 7 || s.Y != 13 || s.Facing != 0 || s.Set != 1 || s.ViewMode != 1 {
				t.Errorf("position %d,%d facing %d set %d view %d", s.X, s.Y, s.Facing, s.Set, s.ViewMode)
			}
			w := s.World
			if w == nil || w.In3D != 1 || w.Flying != 0 || w.SymbolSet != 1 || w.MsgDelay != 4 || w.ViewFlags != 3 || w.RocketX != 9 || w.RocketY != 8 || w.RocketHeading != 0 || w.AreaKind != 0 {
				t.Errorf("world %+v", w)
			}
			if got := s.Fields["TravelKind"]; !slices.Equal(got, []int64{1}) {
				t.Errorf("TravelKind %v", got)
			}
			r := rulesxxvc.New()
			if err := r.Load(s.Rules); err != nil {
				t.Fatal(err)
			}
			if a := r.Check(rules.ClockIs{Part: rules.ClockMinuteOfDay}, nil); a.Value != 0 {
				t.Errorf("clock %d", a.Value)
			}
			hull, err := r.Fact(rules.NoChar, rules.ShipCellFact{Part: rules.ShipSystemStatus, Index: 0})
			if err != nil || hull.(rules.Number) != 0x258 {
				t.Errorf("hull %v %v", hull, err)
			}
		})
	}
}

// TestBRSavedGameConstructed: a game built by layout names (several script cells, a damaged ship, a later clock, another
// position and view) imports as that state.
func TestBRSavedGameConstructed(t *testing.T) {
	for _, g := range brSavedGames {
		t.Run(g.id, func(t *testing.T) {
			b, ds := brSavedGame(t, g.id, g.exe)
			l := readBRLayout(t, g.id)
			b = slices.Clone(b)
			brSetCell(t, l, b, "Word4C31", 5)
			brSetCell(t, l, b, "Word7B04", 'A')
			brSetCell(t, l, b, "Word4CE6", 70000)
			brSetCell(t, l, b, "PrevX", 3)
			brSetCell(t, l, b, "ScriptBlock", 38)
			brSetCell(t, l, b, "Geo", 5)
			brSetCell(t, l, b, "TravelKind", areaOrbit)
			brSetCell(t, l, b, "ShipHull", 300)            // damaged: 600 whole
			brSetCell(t, l, b, "FoeHull", 7)               // the last foe
			brSetCell(t, l, b, "SpaceOwnShip", 1)          // zone 0 shipCombatOutcome
			b[brFileOffset(0x4BB6)] = 4                    // difficulty: Ace
			for i, v := range []int{3, 4, 5, 6, 7, 8, 9} { // the clock, smallest unit first (the day, month and year within their scales)
				b[brFileOffset(0x4BC6+i)] = byte(v)
			}
			b[brFileOffset(0x4BB9)], b[brFileOffset(0x4BBB)], b[brFileOffset(0x4BBD)] = 12, 200, spaceShipFrame+3
			b[brFileOffset(0x4D78)] = 2                        // the last foe's ship type
			b[brFileOffset(0x7EB1)] = 6                        // a member-window cell: no state, ignored
			pos := 1 + brZone0Size + brZone1Size + brZone2Size // the position follows zone 2
			copy(b[pos:], []byte{3, 5, 4, 1, 2})
			b[pos+5], b[pos+6] = 4, ifcOrbitMap // previous, interface
			binary.LittleEndian.PutUint16(b[pos+7:], 3)
			binary.LittleEndian.PutUint16(b[pos+9:], 1)
			s, err := ImportBRSave(b, embeddedScripts(g.id), ds, g.exe)
			if err != nil {
				t.Fatal(err)
			}
			if s.X != 3 || s.Y != 5 || s.Facing != 4 || s.Pos3 != 1 || s.Pos4 != 2 || s.ViewMode != 2 {
				t.Errorf("position %d,%d facing %d view %d", s.X, s.Y, s.Facing, s.ViewMode)
			}
			if !slices.Equal(s.Fields["Word4C31"], []int64{5}) || !slices.Equal(s.Fields["Word7B04"], []int64{'A'}) || !slices.Equal(s.Fields["Word4CE6"], []int64{70000}) || !slices.Equal(s.Fields["SpaceOwnShip"], []int64{1}) {
				t.Errorf("fields %v", s.Fields)
			}
			w := *s.World
			if w.PrevX != 3 || w.AreaKind != 38 || w.Geo != 5 || w.Flying != 1 || w.In3D != 0 || w.OrbitTick != 12 || w.OrbitStep != 200 || w.RocketHeading != 3 {
				t.Errorf("world %+v", w)
			}
			r := rulesxxvc.New()
			if err := r.Load(s.Rules); err != nil {
				t.Fatal(err)
			}
			if a := r.Check(rules.ClockIs{Part: rules.ClockHours}, nil); a.Value != 6 {
				t.Errorf("hours %d", a.Value)
			}
			if a := r.Check(rules.ClockIs{Part: rules.ClockMinuteOfDay}, nil); a.Value != 6*60+5*10+4 {
				t.Errorf("minute of day %d", a.Value)
			}
			hull, _ := r.Fact(rules.NoChar, rules.ShipCellFact{Part: rules.ShipSystemStatus, Index: 0})
			foe, _ := r.Fact(rules.NoChar, rules.ShipCellFact{Foe: true, Part: rules.ShipSystemStatus, Index: 0})
			if hull.(rules.Number) != 300 || foe.(rules.Number) != 7 {
				t.Errorf("hull %v, foe's %v", hull, foe)
			}
			var blob struct{ Difficulty rules.Difficulty }
			if err := json.Unmarshal(s.Rules, &blob); err != nil || blob.Difficulty != rules.DifficultyHardest {
				t.Errorf("difficulty %v %v", blob.Difficulty, err)
			}
			taken, err := r.Resolve(nil, rules.TakeSavedWalls{}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if walls := taken.Effects[0].(rules.SavedWalls); walls.A != [3]uint16{3, noWall, noWall} || walls.B != [3]uint16{1, noWall, noWall} {
				t.Errorf("walls %v", walls)
			}
		})
	}
}

// TestBRSavedGameRefusals: a byte no range of the table covers fails the import, and so do a state with no view and a file of
// the wrong size; a member-window byte is no state.
func TestBRSavedGameRefusals(t *testing.T) {
	for _, g := range brSavedGames {
		t.Run(g.id, func(t *testing.T) {
			b, ds := brSavedGame(t, g.id, g.exe)
			try := func(name string, edit func(b []byte) []byte, wantErr bool) {
				t.Helper()
				_, err := ImportBRSave(edit(slices.Clone(b)), embeddedScripts(g.id), ds, g.exe)
				if (err != nil) != wantErr {
					t.Errorf("%s: error %v, want error %v", name, err, wantErr)
				}
			}
			l := readBRLayout(t, g.id)
			zone0, zone2 := brUnnamed(l, brZone0Base, brZone0Size), brUnnamed(l, brZone2Base, brZone2Size)
			try("unnamed byte", func(b []byte) []byte { b[brFileOffset(zone0)] = 1; return b }, true)
			try("spare byte of the foe's block", func(b []byte) []byte { b[brFileOffset(0x4D7D)] = 1; return b }, true)
			try("unnamed zone 2 byte", func(b []byte) []byte { b[brFileOffset(zone2)] = 1; return b }, true)
			try("member window", func(b []byte) []byte { b[brFileOffset(0x7C10)] = 1; return b }, false)
			try("interface state: ship", func(b []byte) []byte { b[1+brZone0Size+brZone1Size+brZone2Size+6] = 8; return b }, true)
			try("short file", func(b []byte) []byte { return b[:len(b)-1] }, true)
			try("difficulty past Ace", func(b []byte) []byte { b[brFileOffset(0x4BB6)] = 5; return b }, true)
			try("clock past its scale", func(b []byte) []byte { b[brFileOffset(0x4BC8)] = 6; return b }, true)
		})
	}
}

// brUnnamed is an address of the zone that no layout field and no range of the table covers.
func brUnnamed(l *ecl.Layout, base, size int) int {
	for a := base; a < base+size; a++ {
		free := !slices.ContainsFunc(brSaveRanges, func(r brRange) bool { return a >= r.From && a < r.To })
		for _, f := range l.Fields {
			if span, err := brSpan(f); err == nil && a >= f.From && a < f.From+span {
				free = false
			}
		}
		if free {
			return a
		}
	}
	panic("every address of the zone is covered")
}
