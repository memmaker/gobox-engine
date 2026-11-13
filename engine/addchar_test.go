package engine

import (
	"gobox/game"
	"gobox/rules"
	"slices"
	"testing"

	"gobox/game/play"
)

// TestAddChar: on the party menu with saved characters, AddCharacter{Roster: 0}
// adds the first of them (BUG-CLASSES 30); the Add picker lists them in
// BuildCharList order.
func TestAddChar(t *testing.T) {
	e := testEngine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	addChar(t, e)
}

// TestAddCharCOTAB: the same on cotab's imported folder.
func TestAddCharCOTAB(t *testing.T) { addChar(t, advStarted(t, advEnvCOTAB)) }

// addChar is TestAddChar on e, at the party menu.
func addChar(t *testing.T, e *Engine) {
	t.Helper()
	if _, err := e.doNow(play.QuickStart{}); err != nil { // a party to save, then dropped
		t.Fatal(err)
	}
	for _, c := range e.state().Party {
		e.SaveCharacter(e.charID(c), MakeFileName(c.Name()))
	}
	e.state().Party = e.state().Party[:0]
	if len(e.BuildCharList()) == 0 {
		t.Fatal("no saved characters")
	}
	n := len(e.state().Party)
	if _, err := e.doNow(play.OpenPicker{Kind: PickAdd}); err != nil || e.picker == nil {
		t.Fatalf("OpenPicker PickAdd: %v", err)
	}
	if p := commonOf(e.build()).Picker; p == nil || !slices.Equal(p.Labels, e.BuildCharList()) {
		t.Fatalf("Add picker labels %v, want %v", p, e.BuildCharList())
	}
	if _, err := e.doNow(play.AddCharacter{Roster: 0}); err != nil {
		t.Fatalf("AddCharacter refused: %v", err)
	}
	if len(e.state().Party) != n+1 {
		t.Fatalf("party %d, want %d", len(e.state().Party), n+1)
	}
	if _, err := e.doNow(play.AddCharacter{Roster: 0}); err == nil { // BUG-CLASSES 31: added twice, nothing changes
		t.Error("the same character added again was accepted")
	}
}

// TestNoSilentNoop (BUG-CLASSES 31): a call the rules can't take is refused,
// not accepted with nothing changed.
func TestNoSilentNoop(t *testing.T) {
	e := quickStarted(t)
	for _, a := range []play.Action{play.Order{From: 0, To: 99}, play.RemoveNote{ID: 99}} {
		if _, err := e.doNow(a); err == nil {
			t.Errorf("%T%+v accepted with no effect", a, a)
		}
	}
}

// TestDropMember: the party menu's Drop drops the member at once; the
// original's two questions (ovr17:1F5A, 1F93) are gobox's to ask first.
func TestDropMember(t *testing.T) {
	e := quickStarted(t)
	e.Advance() // the party menu comes up
	n := len(e.state().Party)
	if _, err := e.doNow(play.DropMember{Member: 0}); err != nil || len(e.state().Party) != n-1 || commonOf(e.build()).Ask != nil {
		t.Fatalf("Drop: %v, party %d want %d, ask %+v", err, len(e.state().Party), n-1, commonOf(e.build()).Ask)
	}
}

// TestAddFromOtherGame: a game that lists another (Data.Imports) adds that game's characters
// through the one canonical format (Config.Foreign): Roster -1 opens its list, a Roster adds
// that one; another game it does not list, or a host with no source, is refused.
func TestAddFromOtherGame(t *testing.T) {
	e := testEngine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	if _, err := e.doNow(play.QuickStart{}); err != nil {
		t.Fatal(err)
	}
	var files []rules.CharacterFile
	for _, c := range e.state().Party[:2] {
		f, err := e.ruleset().Export(e.charID(c))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	e.state().Party = e.state().Party[:0]
	e.gameData.Imports = []string{"other"}
	e.gameData.Commands = []game.PartyCommand{{Label: "Add from Other", Key: 'O', From: "other"}}
	if _, err := e.doNow(play.AddCharacter{Roster: -1, From: "other"}); err == nil {
		t.Fatal("a host with no source was accepted")
	}
	e.Cfg.Foreign = func(from string) ([]rules.CharacterFile, error) { return files, nil }
	if _, err := e.doNow(play.AddCharacter{Roster: -1, From: "elsewhere"}); err == nil {
		t.Error("a game the data does not list was accepted")
	}
	if _, err := e.doNow(play.AddCharacter{Roster: -1, From: "other"}); err != nil || e.picker == nil || e.picker.kind != PickAddFrom || e.picker.n != 2 {
		t.Fatalf("open the list: %v, picker %+v", err, e.picker)
	}
	if _, err := e.doNow(play.AddCharacter{Roster: 1, From: "other"}); err != nil || len(e.state().Party) != 1 {
		t.Fatalf("add the second: %v, party %d", err, len(e.state().Party))
	}
	if _, err := e.doNow(play.AddCharacter{Roster: 1, From: "other"}); err == nil {
		t.Error("the same character added again was accepted")
	}
}
