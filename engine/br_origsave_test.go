package engine

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/asset/orig"
	"gobox/importer"
	"gobox/rules"
)

// TestBRShippedSavedGameLoads: the originals' shipped SAVGAMA.DAT is the slot "GAME A" of the shared engine (the importer wrote
// it); loading it gives the party of six with their credits, the party at (7,13) facing 0, the game's first block and the
// clock at its start (saveload.pas LoadGame).
func TestBRShippedSavedGameLoads(t *testing.T) {
	for _, g := range []brGame{gameBR1, gameBR2} {
		t.Run(g.ID, func(t *testing.T) {
			e := brEngine(t, g)
			e.Open("")
			e.titleDone()
			e.Advance()
			if e.SlotIndex("GAME A") < 0 {
				t.Fatal("no slot GAME A")
			}
			if !e.loadFrom("GAME A") {
				t.Fatal("the slot did not load")
			}
			if got := len(e.partyIDs()); got != 6 {
				t.Fatalf("party of %d, want 6", got)
			}
			var credits rules.Copper
			for _, id := range e.partyIDs() {
				credits += e.ruleset().Purse(id)
			}
			if credits == 0 {
				t.Error("the party has no credits")
			}
			if e.posX != 7 || e.posY != 13 || e.facing != 0 {
				t.Errorf("at %d,%d facing %d, want 7,13 facing 0", e.posX, e.posY, e.facing)
			}
			if e.viewMode != 1 || e.world.In3D != 1 || e.set != 1 {
				t.Errorf("view mode %d In3D %d set %d, want the 3D view of set 1", e.viewMode, e.world.In3D, e.set)
			}
			if e.world.AreaKind != 0 { // NO_ECL_SCRIPT: GameLoop starts at the game's first block
				t.Errorf("block %d", e.world.AreaKind)
			}
			if m := e.clockPart(rules.ClockMinuteOfDay); m != 0 {
				t.Errorf("clock at minute %d", m)
			}
			e.GameLoop()
			e.Advance()
			if e.gs.areaKind == 0 { // the game's first block ran (BR2's start script moves on to its briefing)
				t.Error("no block is playing")
			}
		})
	}
}

// TestBRConstructedSavedGameLoads: an original game saved flying the rocket (built from the shipped file: another clock, a
// damaged ship, another position and the system map's view) loads into the same state.
func TestBRConstructedSavedGameLoads(t *testing.T) {
	for _, g := range []brGame{gameBR1, gameBR2} {
		t.Run(g.ID, func(t *testing.T) {
			e := brEngine(t, g)
			e.Open("")
			e.titleDone()
			e.Advance()
			b, err := os.ReadFile(filepath.Join(g.dos(), "SAVGAMA.DAT"))
			if err != nil {
				t.Fatal(err)
			}
			im, err := orig.Load(filepath.Join(g.dos(), "START.EXE"), g.EXE.DS)
			if err != nil {
				t.Fatal(err)
			}
			at := func(addr int) int { return 1 + addr - 0x4B00 } // saveload.pas SaveGame's order: the set, then zone 0 from 4B00
			b[at(0x4BC9)] = 7                                     // the clock's hour
			b[at(0x4BC8)] = 3                                     // its ten-minute turn
			b[at(0x4BE6)] = 0x70                                  // areaKind ORBIT
			b[at(0x4D16)], b[at(0x4D17)] = 0x2C, 1                // the hull's status: 300 of 600
			b[at(0x4BBE)], b[at(0x4BBF)] = 12, 4                  // the rocket's square
			b[at(0x4BBD)] = 90                                    // its icon: heading 3
			p := 1 + 0x400 + 0x400 + 0x200                        // the position follows the three zones
			b[p], b[p+1], b[p+2] = 3, 5, 6                        // the party's position: 3,5 facing 6
			b[p+6] = 9                                            // interfaceState ORBIT_MAP
			s, err := importer.ImportBRSave(b, os.DirFS(filepath.Join(e.adv, "scripts")), im.Mem[im.DS:], g.EXE)
			if err != nil {
				t.Fatal(err)
			}
			slot := filepath.Join(e.slotRoot(), "BUILT")
			if err := os.MkdirAll(slot, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := adventure.WriteSave(filepath.Join(slot, adventure.GameFile), s); err != nil {
				t.Fatal(err)
			}
			for _, m := range s.Party {
				rec, err := os.ReadFile(filepath.Join(e.slotRoot(), "GAME A", m+".sav"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(slot, m+".sav"), rec, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if !e.loadFrom("BUILT") {
				t.Fatal("the slot did not load")
			}
			if e.posX != 3 || e.posY != 5 || e.facing != 6 {
				t.Errorf("at %d,%d facing %d, want 3,5 facing 6", e.posX, e.posY, e.facing)
			}
			if e.world.Flying != 1 || e.world.In3D != 0 || e.viewMode != adventure.ViewModeOverland || e.world.RocketX != 12 || e.world.RocketY != 4 || e.world.RocketHeading != 3 {
				t.Errorf("flying %d In3D %d view mode %d rocket %d,%d heading %d", e.world.Flying, e.world.In3D, e.viewMode, e.world.RocketX, e.world.RocketY, e.world.RocketHeading)
			}
			if m := e.clockPart(rules.ClockMinuteOfDay); m != 7*60+3*10 {
				t.Errorf("clock at minute %d", m)
			}
			if hull := e.fact(rules.NoChar, rules.ShipCellFact{Part: rules.ShipSystemStatus, Index: 0}).(rules.Number); hull != 300 {
				t.Errorf("hull %d, want 300", hull)
			}
			if len(e.partyIDs()) != 6 {
				t.Errorf("party of %d", len(e.partyIDs()))
			}
		})
	}
}
