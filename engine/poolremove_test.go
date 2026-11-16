package engine

import (
	"os"
	"testing"

	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
)

// The Pool's party menu Remove, as each game's Pascal does it (SaveChar of chargen.pas's 'R'; the Behaviour data
// says which dialect asks what):
//
//	por    chargen.pas 188-191, saveload.pas 223: SaveChar rewrites the file; no question, an NPC is saved like any
//	cotab  chargen.pas 235-241, saveload.pas 481: 'Overwrite <file>? ' (No first); No asks 'New file name: ' and asks
//	       again for that file; an NPC (npc >= $80) gets DropCharacterMenu
//	sotsb  chargen.pas 268-276, saveload.pas 436: the same, the record takes the new name (PStr15(cr)^ := fn)
//	pod    chargen.pas 200-207, saveload.pas 420: the same, the name too (Types.PCharRecord(...)^.name := fileStem)

// poolRemoveGames are the Pool games: the engine, and the dialect's say (SaveExists, RemoveNPCDrops).
var poolRemoveGames = []struct {
	name   string
	engine func(t *testing.T) *Engine
	asks   ecl.SaveExists
	drops  bool
}{
	{"por", testEngine, ecl.SaveOverwrites, false},
	{"cotab", func(t *testing.T) *Engine { return advEngine(t, advEnvCOTAB) }, ecl.SaveAsksYesNo, true},
	{"sotsb", func(t *testing.T) *Engine { return advEngine(t, advEnvSOTSB) }, ecl.SaveAsksYesNo, true},
	{"pod", func(t *testing.T) *Engine { return advEngine(t, advEnvPOD) }, ecl.SaveAsksYesNo, true},
}

// poolPartyMenu is the game's Quick Start party with the party menu up and the first member saved once (its file exists).
func poolPartyMenu(t *testing.T, mk func(*testing.T) *Engine) (e *Engine, who rules.CharID, size int) {
	t.Helper()
	e = quickStartedOn(t, mk(t))
	e.menu.on = true
	who = e.partyIDs()[0]
	if err := e.SaveCharacter(who, ""); err != nil {
		t.Fatal(err)
	}
	return e, who, e.partySize()
}

func TestPoolRemoveExistingFile(t *testing.T) {
	for _, g := range poolRemoveGames {
		t.Run(g.name, func(t *testing.T) {
			e, who, size := poolPartyMenu(t, g.engine)
			if got := e.eclVersion.Behaviour.SaveExists; got != g.asks {
				t.Fatalf("the dialect's SaveExists is %d, the Pascal's %d", got, g.asks)
			}
			stem := e.charBase(who)
			if err := e.removeMember(0); err != nil {
				t.Fatal(err)
			}
			if g.asks == ecl.SaveOverwrites { // por: no question
				if q := e.question(); q != nil || e.partySize() != size-1 {
					t.Fatalf("a Remove that overwrites silently: question %+v party %d", q, e.partySize())
				}
				return
			}
			if q := e.question(); q == nil || q.Text != "Overwrite "+stem+"? " || !q.YesNo || q.Default != 0 {
				t.Fatalf("question %+v", q)
			}
			reply(t, e, 0) // No: a new file name
			if q := e.question(); q == nil || !q.Typed || q.Text != "New file name:" {
				t.Fatalf("name question %+v", q)
			}
			zed := e.savePath("ZED.cha")
			if err := os.WriteFile(zed, []byte{0}, 0o644); err != nil { // a file of that name exists too
				t.Fatal(err)
			}
			if _, err := e.doNow(game.Reply{Text: "Zed"}); err != nil {
				t.Fatal(err)
			}
			if q := e.question(); q == nil || q.Text != "Overwrite ZED? " { // SaveChar's loop asks again for the new file
				t.Fatalf("after the rename: question %+v", q)
			}
			if e.partySize() != size {
				t.Fatalf("left the party before it was saved: %d", e.partySize())
			}
			reply(t, e, 1) // Yes
			if b, err := os.ReadFile(zed); err != nil || len(b) <= 1 || e.partySize() != size-1 {
				t.Errorf("saved over ZED.cha: %d bytes %v, party %d", len(b), err, e.partySize())
			}
		})
	}
}

func TestPoolRemoveOverwriteYes(t *testing.T) {
	for _, g := range poolRemoveGames[1:] {
		t.Run(g.name, func(t *testing.T) {
			e, who, size := poolPartyMenu(t, g.engine)
			file := e.savePath(e.charBase(who) + ".cha")
			if err := os.WriteFile(file, []byte{0}, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := e.removeMember(0); err != nil {
				t.Fatal(err)
			}
			reply(t, e, 1) // Yes
			if b, _ := os.ReadFile(file); len(b) <= 1 || e.partySize() != size-1 {
				t.Errorf("Yes overwrote %d bytes, party %d", len(b), e.partySize())
			}
		})
	}
}

// TestPoolRemoveNPC: a non-player character's Remove is the Drop question in cotab, sotsb and pod (chargen.pas
// DropCharacterMenu); POR's saves it like any member.
func TestPoolRemoveNPC(t *testing.T) {
	for _, g := range poolRemoveGames {
		t.Run(g.name, func(t *testing.T) {
			e, who, size := poolPartyMenu(t, g.engine)
			e.setSelFact(rules.FactControl, 0xB0) // npc >= NON_PLAYER ($80)
			if err := e.removeMember(0); err != nil {
				t.Fatal(err)
			}
			q := e.question()
			if g.drops {
				if q == nil || q.Text != purgeFirst(e.name(who)) {
					t.Fatalf("question %+v", q)
				}
				reply(t, e, 0) // No: it stays
				if e.partySize() != size {
					t.Errorf("a No dropped it: %d", e.partySize())
				}
				return
			}
			if q != nil || e.partySize() != size-1 { // por saves it, asking nothing
				t.Errorf("POR's Remove of an NPC: question %+v, party %d", q, e.partySize())
			}
		})
	}
}
