package pool

import (
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/ecl/script"
	"slices"
	"testing"

	"gobox/game/play"
)

// A script's HMENU does not wait: it leaves the VM pending on its menu
// (View.Ask), and the answering decision completes it; an encounter's
// WAIT asks again, a typed number not 0..65536 is asked again.
func TestScriptAsks(t *testing.T) {
	e := quickStarted(t)
	e.vm.Text[1], e.vm.Text[2] = "YES", "NO"
	v := testField(t, e, fieldSurprise) // a script variable
	got := func() int { return int(e.host.Get(v)) }
	e.askChoice("", "~YES ~NO", false, false, askDone{kind: dChoiceVar, dst: v})
	if w := e.eclWait; w == nil || len(w.ask.Replies) != 2 || w.ask.YesNo {
		t.Fatalf("ask %+v", e.eclWait)
	}
	if answer(e, play.Continue{}) {
		t.Fatal("Enter answered a menu without it")
	}
	if !answer(e, play.Reply{N: 1}) || got() != 1 || e.eclWait != nil {
		t.Fatalf("NO gave %d, pending %v", got(), e.eclWait != nil)
	}

	// an encounter's WAIT asks again: the menu stays until another choice
	e.es.enc = &encounter{dst: v}
	e.askChoice("", "~COMBAT ~WAIT", false, false, askDone{kind: dDevParlay})
	if answer(e, play.Reply{N: 1}) || e.eclWait == nil {
		t.Fatal("WAIT ended the menu")
	}
	if answer(e, play.Reply{N: 0}) || e.eclWait == nil || len(e.eclWait.ask.Replies) != 5 {
		t.Fatal("COMBAT did not go on to the talk menu")
	}
	answer(e, play.Reply{N: 0})

	e.host.Set(v, 0)
	e.askNumber(askDone{kind: dNumber, dst: v})
	if answer(e, play.Reply{Text: "99999"}) {
		t.Fatal("99999 taken")
	}
	if !answer(e, play.Reply{Text: "42"}) || e.host.Get(v) != 42 {
		t.Fatalf("number %d", e.host.Get(v))
	}

	e.askMember("WHO?", core.NoChar, askDone{kind: dWho})
	if !answer(e, play.Reply{N: 1}) || e.member(e.selected()) != e.state().Party[1] {
		t.Fatal("the member pick did not take member 2")
	}
}

// answer gives a to the open question (opened first if it is not); true
// when nothing is asked any more.
func answer(e *Engine, a play.Action) bool {
	if e.eclWait.seq == 0 {
		e.openAsk()
	}
	e.answer(a)
	return e.eclWait == nil
}

// A yes/no Scene's reply named "Yes" answers yes: the Scene lists play's
// order (No, Yes) as Reply N takes it (it listed Yes, No while Reply 0
// answered No).
func TestSceneYesNoOrder(t *testing.T) {
	for _, want := range []string{"Yes", "No"} {
		e := quickStarted(t)
		v := testField(t, e, fieldSurprise)
		e.host.Set(v, 7)
		e.askYesNo(yesNo("SURE?", false), false, askDone{kind: dChoiceVar, dst: v})
		e.openAsk()
		sc, ok := e.build().(play.Scene)
		if !ok || sc.Ask == nil {
			t.Fatalf("%T, no Scene question", e.build())
		}
		n := slices.Index(sc.Ask.Replies, want)
		if _, err := e.doNow(play.Reply{N: n}); err != nil {
			t.Fatal(err)
		}
		if got := e.host.Get(v) == 1; got != (want == "Yes") {
			t.Errorf("%s answered %d", want, e.host.Get(v))
		}
	}
}

// TestMenuEscForced: a menu naming no exit row takes no Esc: the choice is forced.
func TestMenuEscForced(t *testing.T) {
	e := quickStarted(t)
	e.vm.Script = &script.Script{Code: []script.Instr{{Op: script.OpHMenu}}}
	e.askChoice("", "~BUY ~EXIT", false, false, askDone{})
	e.scriptExit()
	e.openAsk()
	q := e.question()
	if q == nil || !q.NoEsc {
		t.Fatalf("question %+v", q)
	}
	if _, ok := q.Esc.Reply(); ok {
		t.Fatal("Esc answers a reply")
	}
	w := e.eclWait
	if err := e.leaveAsk(q); err == nil || e.eclWait != w {
		t.Fatalf("Esc answered it (err %v)", err)
	}
}

// TestMenuEscHint: Esc picks the row the instruction's hint names, whatever
// its label; a list likewise.
func TestMenuEscHint(t *testing.T) {
	e := quickStarted(t)
	v := testField(t, e, fieldSurprise)
	e.vm.Script = &script.Script{Code: []script.Instr{{Op: script.OpHMenu, Exit: play.EscTo(0)}}, Layout: e.vm.Script.Layout}
	e.host.Set(v, 7)
	e.askChoice("", "~STAY ~GO", false, false, askDone{kind: dChoiceVar, dst: v})
	e.scriptExit()
	e.openAsk()
	q := e.question()
	if i, ok := q.Esc.Reply(); !ok || i != 0 || q.NoEsc {
		t.Fatalf("question %+v", q)
	}
	if err := e.leaveAsk(q); err != nil || e.host.Get(v) != 0 {
		t.Fatalf("err %v, var %d", err, e.host.Get(v))
	}
}

// TestScriptExitHints: every script's exit= row is reachable: a menu row whose
// hotkey is its own, so Esc answers that row and no other.
func TestScriptExitHints(t *testing.T) {
	e := quickStarted(t)
	hints := 0
	for area := range 40 {
		s, err := e.parseScriptFile(area)
		if err != nil {
			continue
		}
		for i, in := range s.Code {
			row, ok := in.Exit.Reply()
			if !ok {
				continue
			}
			hints++
			rows := script.MenuLabels(in)
			if row >= len(rows) {
				t.Errorf("area %d command %d: row %d of %d", area, i, row, len(rows))
				continue
			}
			if in.Op != script.OpHMenu {
				continue
			}
			menu := ""
			for _, r := range rows {
				menu += "~" + r + " "
			}
			if _, ids := hotMenu("", goldbox.Cut(menu, 0x32)); slices.Index(ids, row) < 0 {
				t.Errorf("area %d command %d: %q has no key of its own in %q", area, i, rows[row], rows)
			}
		}
	}
	if hints == 0 {
		t.Fatal("no hints")
	}
}

// TestPurseWho: the two Who commands keyed by their place are the inn's and the tavern's.
func TestPurseWho(t *testing.T) {
	e := quickStarted(t)
	e.loadScript(0)
	var got []string
	for i, in := range e.vm.Script.Code {
		if e.purseWho[i] && in.Op == script.OpNamed("Who") {
			got = append(got, in.Args[0].S)
		}
	}
	if !slices.Equal(got, []string{"WHO WILL PAY?", "WHO WILL GAMBLE?"}) {
		t.Fatalf("Who at the purse's places: %q", got)
	}
}
