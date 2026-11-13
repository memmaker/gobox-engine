package engine

import (
	"gobox/core"
	"image"
	"reflect"
	"gobox/rules1e"
	"slices"
	"testing"
	"time"

	"gobox/game/play"
)

// The engine is a play.Game: exploring after a Quick Start, a step through
// RunEvent{Before}, Arrive and RunEvent, a turn, camp, a refusal and a
// Save/Load round trip.
func TestPlayGame(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	x, ok := e.build().(play.Exploring)
	if !ok || len(x.Party) == 0 {
		t.Fatalf("state %T, want Exploring with a party", e.build())
	}
	if len(e.state().Party) > 1 {
		e.sel = 1
	}
	sel := e.sel
	if s1, s2 := e.build(), e.build(); !reflect.DeepEqual(s1, s2) || e.sel != sel { // bug class 1: reading changes nothing
		t.Fatalf("State twice differs, or moved the selection %d -> %d", sel, e.sel)
	}
	if _, err := e.doNow(play.Buy{}); err == nil {
		t.Error("Buy while exploring: no refusal")
	}
	f := x.Where.Facing
	turn := (f + 2) % 8
	if _, err := e.doNow(play.Face{Facing: turn}); err != nil {
		t.Fatal("Face:", err)
	}
	if w := e.build().(play.Exploring).Where; w.Facing != turn {
		t.Errorf("facing %d, want %d", w.Facing, turn)
	}
	b, err := e.Save()
	if err != nil {
		t.Fatal("Save:", err)
	}
	if _, err := e.doNow(play.MakeCamp{}); err != nil {
		t.Fatal("MakeCamp:", err)
	}
	if _, ok := e.build().(play.Camping); !ok {
		t.Fatalf("after MakeCamp %T, want Camping", e.build())
	}
	if len(e.state().Party) > 1 {
		first := e.state().Party[0]
		if _, err := e.doNow(play.Order{From: 0, To: 1}); err != nil {
			t.Fatal("Order:", err)
		}
		if e.state().Party[1] != first {
			t.Error("Order did not move the first member")
		}
	}
	if _, err := e.doNow(play.AdvanceTimeBy{D: time.Hour}); err != nil {
		t.Fatal("AdvanceTimeBy:", err)
	}
	if _, err := e.doNow(play.BreakCamp{}); err != nil {
		t.Fatal("BreakCamp:", err)
	}
	if err := e.Load(b); err != nil {
		t.Fatal("Load:", err)
	}
	if w := e.build().(play.Exploring).Where; w.Facing != turn {
		t.Errorf("after Load facing %d, want %d", w.Facing, turn)
	}
}

// PromptRow24 outside a fight waits with keyWait, never the fight's pause
// (BUG-CLASSES 11: that panics under test): a file error while exploring.
func TestPromptRow24Explore(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.fileError("read", "NOFILE.DAT")
	if e.question() == nil {
		t.Fatal("no key wait after the file error")
	}
}

// gobox's pool pick with no draft up reaches the selected member (bug class 2).
func TestPlayPoolPickMember(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	if _, err := e.doNow(play.SetChoice{Field: "iconbody", Text: "pool-body-7"}); err != nil {
		t.Fatal(err)
	}
	if got := e.state().Party[e.sel].PoolIDs.IconBodyID; got != "pool-body-7" {
		t.Errorf("selected member's icon body %q, want pool-body-7", got)
	}
}

// The party menu and the creation sheet as play actions: a new character
// rolled, a field set, kept under a name, added, dropped.
func TestPlayParty(t *testing.T) {
	e := testEngine(t)
	e.viewBit0 = 1
	e.Init()
	e.Start(e.Run)
	for i := 0; i < 5; i++ { // past the title
		if sc, ok := e.build().(play.Scene); !ok || sc.Kind != play.SceneTitle {
			break
		}
		if _, err := e.doNow(play.Continue{}); err != nil {
			t.Fatal("Continue:", err)
		}
	}
	if _, ok := e.build().(play.Assembling); !ok {
		t.Fatalf("state %T, want Assembling", e.build())
	}
	if _, err := e.doNow(play.NewCharacter{}); err != nil {
		t.Fatal("NewCharacter:", err)
	}
	as := e.build().(play.Assembling)
	if as.Draft == nil || len(as.Draft.Fields) == 0 {
		t.Fatal("no draft after NewCharacter")
	}
	if _, err := e.doNow(play.Reroll{}); err != nil {
		t.Fatal("Reroll:", err)
	}
	f := as.Draft.Fields[0]
	if _, err := e.doNow(play.SetChoice{Field: f.Name, N: len(f.Options) - 1}); err != nil {
		t.Fatal("SetChoice:", err)
	}
	if _, err := e.doNow(play.SetChoice{Field: "iconhead", Text: "pool-head-3"}); err != nil {
		t.Fatal("SetChoice pool id:", err)
	}
	if e.member(e.creation.c).PoolIDs.IconHeadID != "pool-head-3" {
		t.Error("the draft's pool pick not taken")
	}
	e.rename("TESTER")
	if _, err := e.doNow(play.KeepCharacter{}); err != nil {
		t.Fatal("KeepCharacter:", err)
	}
	if e.creation != nil {
		t.Fatal("draft still up after Keep")
	}
	roster := e.build().(play.Assembling).Roster
	i := slices.Index(roster, "TESTER")
	if i < 0 {
		t.Fatalf("TESTER not in the roster %v", roster)
	}
	if _, err := e.doNow(play.OpenPicker{Kind: PickAdd}); err != nil || e.picker == nil {
		t.Fatal("OpenPicker:", err)
	}
	if _, err := e.doNow(play.ClosePicker{}); err != nil || e.picker != nil {
		t.Fatal("ClosePicker:", err)
	}
	if _, err := e.doNow(play.AddCharacter{Roster: i}); err != nil {
		t.Fatal("AddCharacter:", err)
	}
	if n := len(e.state().Party); n != 1 {
		t.Fatalf("%d members after Add, want 1", n)
	}
	size := e.state().Party[0].IconSize
	for _, a := range []play.Action{play.EditIcon{Member: 0}, play.IconSize{}, play.EndIcon{Keep: true}} {
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("%T: %v", a, err)
		}
	}
	if e.iconEd != nil || e.state().Party[0].IconSize == size {
		t.Errorf("icon editor up %v, size %d (was %d)", e.iconEd != nil, e.state().Party[0].IconSize, size)
	}
	if _, err := e.doNow(play.DropMember{Member: 0}); err != nil {
		t.Fatal("DropMember:", err)
	}
	if n := len(e.state().Party); n != 0 {
		t.Errorf("%d members after Drop, want 0", n)
	}
}

// devPlay is an exploring engine with dev scene name up.
func devPlay(t *testing.T, name string) *Engine { return devPlayOn(t, testEngine(t), name) }

// devPlayOn is devPlay on the engine e (any game's: advEngine).
func devPlayOn(t *testing.T, e *Engine, name string) *Engine {
	e.Explore = true
	e.DevAction = e.DevScene(name)
	seed := uint32(1)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	return e
}

// Do answers a script's typed question with Reply and leaves the shop
// with Leave itself, not through the adapter's Answer or Cancel.
func TestPlayReplyLeave(t *testing.T) {
	e := devPlay(t, "password")
	if e.question() == nil {
		t.Fatalf("no question up in the password scene (%T)", e.build())
	}
	if _, err := e.doNow(play.Reply{Text: "WORD"}); err != nil {
		t.Fatal("Reply:", err)
	}
	if q := e.question(); q != nil && q.Typed {
		t.Error("the typed question still up")
	}
	e = devPlay(t, "shop")
	if _, ok := e.build().(play.Trading); !ok {
		t.Fatalf("shop scene: %T", e.build())
	}
	if _, err := e.doNow(play.Leave{}); err != nil {
		t.Fatal("Leave:", err)
	}
	if _, ok := e.build().(play.Trading); ok {
		t.Error("still trading after Leave")
	}
}

// toTurn runs e on to a member's turn in the fight.
func toTurn(t *testing.T, e *Engine) {
	t.Helper()
	for range 400 {
		if e.cs.turn != nil && e.cs.actor != core.NoChar && e.cs.aim == nil {
			return
		}
		if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	t.Fatal("no member's turn")
}

// A fight's question (Flee:) is pending data: Do asks it, the Reply takes
// it, Leave on another one says No; a melee Attack is taken.
func TestPlayFightQuestion(t *testing.T) {
	e := devPlay(t, "fleefight")
	toTurn(t, e)
	if _, err := e.doNow(play.Flee{}); err != nil {
		t.Fatal("Flee:", err)
	}
	f, ok := e.build().(play.Fighting)
	if !ok || f.Ask == nil || !f.Ask.YesNo || e.ext.pq == nil {
		t.Fatalf("after Flee %T, ask %+v: want the Flee question", e.build(), f.Ask)
	}
	if _, err := e.doNow(play.MoveTo{To: image.Pt(0, 0)}); err == nil {
		t.Error("a move with the question up: no refusal")
	}
	actor := e.member(e.cs.actor)
	if _, err := e.doNow(play.Reply{N: 1}); err != nil {
		t.Fatal("Reply yes:", err)
	}
	if e.ext.pq != nil || e.member(e.cs.actor) == actor && e.cs.turn != nil {
		t.Error("the Flee not taken")
	}

	e = devPlay(t, "meleefight")
	toTurn(t, e)
	f = e.build().(play.Fighting)
	var at play.Action
	for _, x := range f.Allowed.Plain {
		if _, ok := x.(play.Attack); ok {
			at = x
			break
		}
	}
	if at == nil {
		t.Fatal("no Attack allowed in the melee")
	}
	actor = e.member(e.cs.actor)
	if _, err := e.doNow(at); err != nil {
		t.Fatal("Attack:", err)
	}
	if e.member(e.cs.actor) == actor && e.cs.aim != nil {
		t.Error("the attack left the aim up")
	}
}

// The treasure's Leave while treasure is left asks (O072, ovr06:10B8) as
// pending data: Yes goes back to it, No leaves.
func TestPlayLootLeave(t *testing.T) {
	e := devPlay(t, "loot")
	for i := 0; i < 20; i++ {
		if _, ok := e.build().(play.Looting); ok {
			break
		}
		press(e)
	}
	if _, ok := e.build().(play.Looting); !ok {
		t.Skipf("loot scene: %T", e.build())
	}
	seeLoot(t, e)
	if _, err := e.doNow(play.Leave{}); err != nil {
		t.Fatal("Leave:", err)
	}
	if e.ext.pq == nil {
		if _, ok := e.build().(play.Looting); ok {
			t.Fatal("Leave neither asked nor left")
		}
		return // nothing left to ask about
	}
	t.Log("asked:", e.ext.pq.q.Text)
	if _, err := e.doNow(play.Reply{N: 1}); err != nil { // Yes: back to the treasure
		t.Fatal("Reply yes:", err)
	}
	if _, ok := e.build().(play.Looting); !ok {
		t.Fatalf("Yes left the treasure: %T", e.build())
	}
	e.doNow(play.Leave{})
	if _, err := e.doNow(play.Reply{N: 0}); err != nil { // No: leave it
		t.Fatal("Reply no:", err)
	}
	if _, ok := e.build().(play.Looting); ok {
		t.Error("No did not leave the treasure")
	}
}

// A step as gobox walks it through Do: RunEvent{Before}, Arrive, RunEvent.
func TestPlayWalk(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	for _, d := range []int{0, 2, 4, 6} {
		x := e.build().(play.Exploring)
		from := image.Pt(x.Where.X, x.Where.Y)
		if _, err := e.doNow(play.RunEvent{Before: true, Side: d}); err != nil {
			t.Fatal("RunEvent Before:", err)
		}
		x, ok := e.build().(play.Exploring)
		if !ok {
			continue
		}
		var arr play.Action
		for _, a := range x.Allowed.Plain {
			if v, ok := a.(play.Arrive); ok && v.To == from.Add(sideSteps[d]) {
				arr = v
			}
		}
		if arr == nil {
			continue // a wall that way
		}
		out, err := e.doNow(arr)
		if err != nil {
			t.Fatal("Arrive:", err)
		}
		if w := e.build().(play.Exploring).Where; image.Pt(w.X, w.Y) != from.Add(sideSteps[d]) {
			t.Fatalf("at %d,%d after Arrive, want %v", w.X, w.Y, from.Add(sideSteps[d]))
		}
		if len(out.Events) == 0 {
			t.Error("Arrive: no Moved event")
		}
		if _, err := e.doNow(play.RunEvent{}); err != nil {
			t.Fatal("RunEvent:", err)
		}
		return
	}
	t.Fatal("no way to step")
}

// commonOf is what every play state carries.
func commonOf(s play.State) play.Common {
	switch s := s.(type) {
	case play.Assembling:
		return s.Common
	case play.Exploring:
		return s.Common
	case play.Camping:
		return s.Common
	case play.Fighting:
		return s.Common
	case play.Scene:
		return s.Common
	case play.Trading:
		return s.Common
	case play.Looting:
		return s.Common
	case play.Over:
		return s.Common
	}
	return play.Common{}
}

// playAllows: the state allows a (by value).
func playAllows(s play.State, a play.Action) bool {
	return slices.ContainsFunc(commonOf(s).Allowed.Plain, func(x play.Action) bool { return reflect.DeepEqual(x, a) })
}

// playOn answers what the game waits on (a Continue, else a question's
// first reply) until done holds or n actions are taken; false: it never held.
func playOn(e *Engine, n int, done func(play.State) bool) bool {
	for range n {
		s := e.build()
		if done(s) {
			return true
		}
		switch {
		case playAllows(s, play.Continue{}):
			e.doNow(play.Continue{})
		case commonOf(s).Ask != nil:
			if q := commonOf(s).Ask; q.Typed || q.Number {
				e.doNow(play.Reply{})
			} else {
				e.doNow(play.Reply{N: 0})
			}
		default:
			return done(e.build())
		}
	}
	return done(e.build())
}

// press is the player's key to what waits, through Do: a page's or a
// pause's Continue, a lone "Press" reply; with nothing to continue the
// engine runs on.
func press(e *Engine) {
	var a play.Action = play.Continue{}
	if q := e.question(); q != nil && len(q.Replies) == 1 {
		a = play.Reply{N: 0}
	}
	if !allows(commonOf(e.build()).Allowed, a) {
		e.advance()
		return
	}
	if _, err := e.doNow(a); err != nil {
		panic(err) // allowed, so taken
	}
}

// TestPlayEndGame: EndGame drops the game in play for the party menu; it was
// Quit's call, refused outside the party menu.
func TestPlayEndGame(t *testing.T) {
	e := quickStarted(t)
	e.Explore = true
	e.Start(e.Play)
	if !e.exploring() {
		t.Fatal("not exploring")
	}
	if _, err := e.doNow(play.EndGame{}); err != nil {
		t.Fatal(err)
	}
	if !e.menuUp() {
		t.Fatalf("party menu not up: top %v", e.top())
	}
}

// TestPlayAdvanceTime: AdvanceTimeBy rests, as Fix does: it told of time passed
// while a call camp ignored did nothing.
func TestPlayAdvanceTime(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	c := e.state().Party[0]
	c.HP = 1
	if _, err := e.doNow(play.AdvanceTimeBy{D: time.Hour}); err != nil {
		t.Fatal(err)
	}
	if c.HP <= 1 {
		t.Fatalf("no rest: %d hit points", c.HP)
	}
}

// TestPlayCampCast: Cast out of a fight reaches the camp's spell list (it named
// a power none has, so it was always refused).
func TestPlayCampCast(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	e.state().Party[0].Memorized[0] = rules1e.SpCureLightWounds
	if _, err := e.doNow(play.Cast{Member: 0, Spell: rules1e.SpCureLightWounds}); err != nil {
		t.Fatal(err)
	}
}

// doNow is Do on the engine as it is now: these tests change it outside Do
// (its internals), so State is built anew first.
func (e *Engine) doNow(a play.Action) (play.Outcome, error) {
	e.rebuild()
	return e.Do(a)
}
