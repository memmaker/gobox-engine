package engine

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rulesxxvc"
)

// brAtPartyMenu is the game on its party menu with the first roster character in the party, not begun.
func brAtPartyMenu(t *testing.T, g brGame) *Engine {
	t.Helper()
	e := brEngine(t, g)
	e.Open("")
	e.titleDone()
	e.Advance()
	if _, err := e.doNow(game.AddCharacter{Roster: 0}); err != nil {
		t.Fatal(err)
	}
	return e
}

// brSetExp gives member id the experience exp (the ruleset's save is JSON: its one door for a test).
func brSetExp(t *testing.T, e *Engine, id rules.CharID, exp int64) {
	t.Helper()
	r := e.ruleset().(*rulesxxvc.Rules)
	b, err := r.Save()
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	doc["Chars"].(map[string]any)[strconv.Itoa(int(id))].(map[string]any)["Exp"] = exp
	if b, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	if err := r.Load(b); err != nil {
		t.Fatal(err)
	}
}

// brInHall is the game begun and a script's training hall up over its party menu (the script writes the hall flag
// PartyExp, then Program(0)), the first member holding the experience exp.
func brInHall(t *testing.T, g brGame, exp int64) *Engine {
	t.Helper()
	e := brBegun(t, g)
	brSetExp(t, e, e.partyIDs()[0], exp)
	br1OwnBlock(t, e, 201, "    PartyExp = 127\n    Program(0)")
	if !e.ext.hall || !e.menu.on || e.world.PartyExp != 127 {
		t.Fatalf("no hall: hall %v menu %v flag %d", e.ext.hall, e.menu.on, e.world.PartyExp)
	}
	return e
}

func offersTraining(e *Engine) bool {
	for _, f := range e.partyFunctions() {
		if f == partyTraining {
			return true
		}
	}
	return false
}

// TestBR1TrainHall ports rules TestTrainHall (chargen.pas MainMenu: Train Character is enabled while zone1's
// trainingMask is not 0): the party menu offers Train only in a training hall, to a member who would rise.
func TestBR1TrainHall(t *testing.T)      { brTrainHall(t, gameBR1) }
func TestBR2TrainHall(t *testing.T)      { brTrainHall(t, gameBR2) }
func TestBR1TrainHallLeave(t *testing.T) { brTrainHallLeave(t, gameBR1) }
func TestBR2TrainHallLeave(t *testing.T) { brTrainHallLeave(t, gameBR2) }
func TestBR1TrainLevel(t *testing.T)     { brTrainLevel(t, gameBR1) }
func TestBR2TrainLevel(t *testing.T)     { brTrainLevel(t, gameBR2) }
func TestBR1TrainNotEnough(t *testing.T) { brTrainNotEnough(t, gameBR1) }
func TestBR2TrainNotEnough(t *testing.T) { brTrainNotEnough(t, gameBR2) }

func brTrainHall(t *testing.T, g brGame) {
	e := brAtPartyMenu(t, g)
	brSetExp(t, e, e.partyIDs()[0], 1<<30)
	if offersTraining(e) {
		t.Fatal("Train offered outside a hall")
	}
	e = brInHall(t, g, 1<<30)
	if !offersTraining(e) {
		t.Fatal("Train hidden in a hall")
	}
	brSetExp(t, e, e.partyIDs()[0], 0)
	if offersTraining(e) {
		t.Error("Train offered to a member with no experience")
	}
}

// TestBR1TrainHallLeave ports rules TestTrainHallLeave (MainMenu's Begin: zone1 trainingMask := 0).
func brTrainHallLeave(t *testing.T, g brGame) {
	e := brInHall(t, g, 0)
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
	if e.world.PartyExp != 0 || e.ext.hall {
		t.Errorf("hall flag %d hall %v after Begin", e.world.PartyExp, e.ext.hall)
	}
}

// brTrainLevel ports rules TestTrain (chargen.pas TrainingMenu): the offer and its question, a level with its hit
// points, the picks (EditSkills(40, 20, 15)) on the creation's allotment screen, kept by Exit, then the party menu.
func brTrainLevel(t *testing.T, g brGame) {
	e := brInHall(t, g, 1<<30)
	id := e.partyIDs()[0]
	was := e.ruleset().Sheet(id)
	if _, err := e.doNow(game.OpenTraining{Member: 0}); err != nil {
		t.Fatal(err)
	}
	tr, ok := e.build().(game.Trading)
	if !ok || tr.Training == nil || !tr.Training.NoFee {
		t.Fatalf("the hall: %T %+v", e.build(), tr.Training)
	}
	if b := tr.Training.Becomes[0]; len(b) != 1 || !strings.HasPrefix(b[0], "a level "+strconv.Itoa(int(was.Levels[was.Classes[0]])+1)+" ") {
		t.Errorf("becomes %v", b)
	}
	if _, err := e.doNow(game.Train{Member: 0}); err != nil {
		t.Fatalf("%v: %+v %+v", err, tr.Training, game.CommonOf(e.build()).Allowed)
	}
	now := e.ruleset().Sheet(id)
	if now.Levels[now.Classes[0]] != was.Levels[was.Classes[0]]+1 || now.MaxHP <= was.MaxHP || now.Exp >= 1<<30 {
		t.Errorf("level %d hp %d exp %d, was level %d hp %d", now.Levels[now.Classes[0]], now.MaxHP, now.Exp, was.Levels[was.Classes[0]], was.MaxHP)
	}
	as, ok := e.build().(game.Assembling)
	if !ok || as.Draft == nil || !as.Draft.Rise || len(as.Draft.Allotments) != 2 {
		t.Fatalf("the picks: %T %+v", e.build(), as.Draft)
	}
	if a := as.Draft.Allotments; a[0].Left != 40 || a[1].Left != 20 || a[0].Max != 15 {
		t.Errorf("points %d, %d at most %d", a[0].Left, a[1].Left, a[0].Max)
	}
	if len(as.Draft.Fields) == 1 { // a warrior's odd level: the weapon first
		if _, err := e.doNow(game.SetChoice{Field: as.Draft.Fields[0].Name, N: 0}); err != nil {
			t.Fatal(err)
		}
	}
	row := -1
	for i, r := range as.Draft.Allotments[0].Rows {
		if r.Open {
			row = i
			break
		}
	}
	for range 2 {
		if _, err := e.doNow(game.Spend{Allotment: 0, Row: row, By: 1}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := e.doNow(game.KeepCharacter{}); err != nil { // Exit keeps the picks: the screen cannot be abandoned
		t.Fatal(err)
	}
	if e.creation != nil {
		t.Fatal("the picks are still up")
	}
	sum := func(sh rules.Sheet) (n int32) {
		for _, v := range sh.Skills {
			n += v
		}
		return n
	}
	if got, want := sum(e.ruleset().Sheet(id)), sum(now)+2; got != want {
		t.Errorf("skill levels %d, want %d: the two points are kept", got, want)
	}
	if !e.menu.on || e.trainingOpen() {
		t.Error("not back at the party menu")
	}
}

// brTrainNotEnough: a member short of the next level's experience is not offered Train, and the hall refuses it in the
// original's words (TrainingMenu: 'Not enough experience').
func brTrainNotEnough(t *testing.T, g brGame) {
	e := brInHall(t, g, 0)
	id := e.partyIDs()[0]
	card := e.fact(id, rules.TrainFact{Hall: e.hallTerms()}).(rules.TrainCard)
	if card.Why != "Not enough experience" || card.Rises {
		t.Errorf("%+v", card)
	}
}
