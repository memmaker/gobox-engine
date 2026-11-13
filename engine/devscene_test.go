package engine

import (
	"gobox/rules"
	"gobox/rules1e"
	"image"
	"testing"

	"gobox/game/play"
)

var devSeed = uint32(2) // Felix's R3 rulings 1-2 (3d6+1 abilities, best-of-two creation HP) moved the Quick Start party's dice: seed 1's castfight actor takes no timed spell

// devFirstTurn starts -dev name and plays to the party's first turn.
func devFirstTurn(t *testing.T, name string, deskew bool) *Engine {
	t.Helper()
	return devFirstTurnOn(t, testEngine(t), name, deskew)
}

// devFirstTurnOn is devFirstTurn on the engine e (any game's: advEngine).
func devFirstTurnOn(t *testing.T, e *Engine, name string, deskew bool) *Engine {
	t.Helper()
	e.Explore = true
	e.Deskew.Store(deskew)
	e.DevAction = e.DevScene(name)
	seed := devSeed
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	for range 400 {
		if e.cs.turn != nil && e.cs.actor != rules.NoChar {
			break
		}
		if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	if e.cs.turn == nil || e.cs.actor == rules.NoChar {
		t.Fatalf("%s: no party turn", name)
	}
	return e
}

// myTurn: a member's turn is up, its verbs offered.
func myTurn(s play.State) bool { return playAllows(s, play.EndTurn{}) }

// fightQ is the question up in the fight, nil none.
func fightQ(s play.State) *play.Question {
	if f, ok := s.(play.Fighting); ok {
		return f.Ask
	}
	return nil
}

// yes is the Reply that answers q yes (play's yes/no: no, yes).
func yes(q *play.Question) play.Reply {
	if q != nil && q.YesNo {
		return play.Reply{N: 1}
	}
	return play.Reply{}
}

// no is the Reply that answers q no.
func no(q *play.Question) play.Reply {
	if q != nil && q.YesNo {
		return play.Reply{}
	}
	return play.Reply{N: 1}
}

// TestDevMeleeAndFlee: on the first party turn of -dev meleefight the
// actor has a foe next to it, and an Attack at its square (gobox's bump) attacks; on that of
// -dev fleefight Flee (a step off the map) asks Flee:. Deskewed or not.
func TestDevMeleeAndFlee(t *testing.T) {
	for _, deskew := range []bool{false, true} {
		for _, name := range []string{"meleefight", "fleefight"} {
			e := devFirstTurn(t, name, deskew)
			a := e.member(e.cs.actor)
			found := false
			for dir := 0; dir < 8 && !found; dir++ {
				who, g := e.state().GroundAt(a, dir)
				to := image.Pt(e.state().CombatX(a)+rules1e.DirDX[dir], e.state().CombatY(a)+rules1e.DirDY[dir])
				switch {
				case name == "meleefight" && who > 0 && e.state().Combatant(who).Side != a.Side:
					found = true
					if _, err := e.doNow(play.Attack{At: &to}); err != nil { // gobox's bump
						t.Errorf("%s deskew %v: bump %v", name, deskew, err)
					}
				case name == "fleefight" && who == 0 && g == 0:
					found = true
					if _, err := e.doNow(play.Flee{}); err != nil {
						t.Errorf("%s deskew %v: Flee %v", name, deskew, err)
					}
					if q := fightQ(e.build()); q == nil || q.Text != "Flee:" {
						t.Errorf("%s deskew %v: Flee asks %+v, want Flee:", name, deskew, q)
					}
				}
			}
			if !found {
				t.Errorf("%s deskew %v: nothing next to the first actor %s at %d,%d", name, deskew, a.Name(), e.state().CombatX(a), e.state().CombatY(a))
			}
		}
	}
}

// TestDevDrop: -dev drop gives the selected member something to drop, and
// Drop drops it from the explore state (ovr20's Drop It? is gobox's).
func TestDevDrop(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("drop")
	e.Init()
	e.Start(e.Play)
	for range 50 {
		if e.snap.invOn {
			break
		}
		press(e)
	}
	cID := e.selected()
	items := e.listedItems(cID)
	if !e.snap.invOn || len(items) == 0 {
		t.Fatal("drop: nothing held")
	}
	c, h := e.member(cID), items[0]
	m := e.state().PartyIndex(c)
	e.CloseItems(true) // the items screen is gobox's: Drop opens what it needs
	n := len(e.state().Carried(c))
	if _, err := e.doNow(play.Drop{Member: m, Item: h}); err != nil || len(e.state().Carried(c)) != n-1 {
		t.Errorf("drop: %v, %d items, want %d (%T)", err, len(e.state().Carried(c)), n-1, e.build())
	}
}

// TestDevMissileShoots: in -dev missile a member with a launcher reaches
// 3+ squares, and Shoot at a foe in range fires a missile (Shot) that
// carries its sprite.
func TestDevMissileShoots(t *testing.T) {
	e := devFirstTurn(t, "missile", false)
	for range 400 {
		s := e.build()
		if _, ok := s.(play.Fighting); !ok {
			break
		}
		if q := fightQ(s); q != nil || !myTurn(s) {
			if _, err := e.doNow(yes(q)); err != nil {
				t.Fatalf("%T: %v", s, err)
			}
			continue
		}
		a := e.member(e.cs.actor)
		rng := e.state().WeaponRange(a)
		var foe *rules1e.Character
		for i := 1; i <= e.state().Ncomb; i++ {
			if f := e.state().CombatantSlots[i-1]; f != nil && f.Side != a.Side && e.state().CanTarget(a, f, rng, 1) {
				foe = f
			}
		}
		if !e.state().RangedWeapon(a) || foe == nil {
			if rng < 3 && e.state().RangedWeapon(a) {
				t.Fatalf("%s's launcher reaches %d squares", a.Name(), rng)
			}
			e.doNow(play.Delay{}) // the foes come closer
			continue
		}
		if rng < 3 {
			t.Fatalf("%s's launcher reaches %d squares", a.Name(), rng)
		}
		out, err := e.doNow(play.Shoot{Ats: []image.Point{image.Pt(e.state().CombatX(foe), e.state().CombatY(foe))}})
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range out.Events {
			if s, ok := ev.(play.Shot); ok {
				if s.Art.Pixels == nil { // BUG-CLASSES 44: no sprite, nothing flies
					t.Fatalf("a shot without its sprite: %+v", s)
				}
				return
			}
		}
		t.Fatalf("%s shot at %s: no Shot in %d events", a.Name(), foe.Name(), len(out.Events))
	}
	t.Fatal("no member with a launcher and a foe in range")
}

// TestFightAsksKeepCombat (BUG-CLASSES 46): every question a fight puts
// (Flee:, Attack Ally?) leaves the state the fight: Fighting
// with its Ask.
func TestFightAsksKeepCombat(t *testing.T) {
	inFight := func(what string, e *Engine, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", what, err)
		}
		if fightQ(e.build()) == nil {
			t.Fatalf("%s: %T with no question in the fight", what, e.build())
		}
	}
	e := devFirstTurn(t, "fleefight", false)
	_, err := e.doNow(play.Flee{})
	inFight("Flee:", e, err)
	e = devFirstTurn(t, "meleefight", false)
	a := e.member(e.cs.actor)
	for dir := 0; dir < 8; dir++ {
		if who, _ := e.state().GroundAt(a, dir); who > 0 && e.state().Combatant(who).Side == a.Side {
			at := image.Pt(e.state().CombatX(a)+rules1e.DirDX[dir], e.state().CombatY(a)+rules1e.DirDY[dir])
			_, err := e.doNow(play.Attack{At: &at})
			inFight("Attack Ally?", e, err)
			e.doNow(no(fightQ(e.build())))
			break
		}
	}
}

// endTurn ends the turn (play's EndTurn is Done's end: no menu between).
func endTurn(t *testing.T, e *Engine) play.Outcome {
	t.Helper()
	o, err := e.doNow(play.EndTurn{})
	if err != nil {
		t.Fatal("EndTurn:", err)
	}
	return o
}

// TestContinueBattle: the foes gone and the party standing, the round's end
// asks ovr09's "Continue Battle:" (0971) inside the fight (Fighting, its
// Ask a strict yes/no): Yes plays on and asks again, No ends the fight.
func TestContinueBattle(t *testing.T) {
	e := devFirstTurn(t, "meleefight", false)
	for i := 1; i <= e.state().Ncomb; i++ {
		if c := e.state().CombatantSlots[i-1]; c != nil && c.Side != 0 {
			c.Okay, c.Status = 0, rules1e.StatusDead
		}
	}
	ask := func() *play.Question {
		t.Helper()
		for range 60 {
			s := e.build()
			if q := commonOf(s).Ask; q != nil {
				if _, ok := s.(play.Fighting); q.Text != "Continue Battle:" || !ok {
					t.Fatalf("asked %q in %T", q.Text, s)
				}
				if !q.YesNo || !q.NoEsc || !q.NoPick { // ovr26:10D1 loops on Y/N
					t.Fatalf("Continue Battle: %+v", q)
				}
				return q
			}
			if !myTurn(s) {
				t.Fatalf("at rest in %T with nothing to do", s)
			}
			endTurn(t, e)
		}
		t.Fatal("no Continue Battle:")
		return nil
	}
	e.doNow(yes(ask()))
	if _, ok := e.build().(play.Fighting); !ok {
		sc, _ := e.build().(play.Scene)
		t.Fatalf("Yes ended the fight: %T %q %q phase %v flag %d", e.build(), sc.Text, commonOf(sc).Text, e.cs.phase, e.state().BeginFlag)
	}
	e.doNow(no(ask()))
	if _, ok := e.build().(play.Fighting); ok {
		t.Fatal("No left the fight up")
	}
}

// TestDevWinFight: -dev winfight is winnable by fighting: Quick for the
// party (the computer attacks) beats its one 1-hp foe, and the round's end
// asks "Continue Battle:" (ovr09:0971), live-reachable.
func TestDevWinFight(t *testing.T) {
	e := devFirstTurn(t, "winfight", false)
	for range 300 {
		s := e.build()
		if q := commonOf(s).Ask; q != nil && q.Text == "Continue Battle:" {
			return
		}
		if !myTurn(s) {
			t.Fatalf("at rest in %T with nothing to do", s)
		}
		e.doNow(play.Quick{})
	}
	t.Fatal("winfight: no Continue Battle:")
}

// TestPartyWipeEnds: a lost fight ("The monsters rejoice", ovr06:1527 with
// combatResult's quit) answered on reaches the game's end, the party menu,
// not explore with a dead party (BUG-CLASSES 53).
func TestPartyWipeEnds(t *testing.T) { testPartyWipeEnds(t, testEngine) }

func testPartyWipeEnds(t *testing.T, mk func(*testing.T) *Engine) {
	e := devFirstTurnOn(t, mk(t), "losefight", false)
	for range 300 {
		s := e.build()
		switch q := commonOf(s).Ask; {
		case isAssembling(s):
			if e.run.quit {
				t.Fatal("party menu with quit set")
			}
			return
		case q != nil:
			e.doNow(yes(q))
		case myTurn(s):
			endTurn(t, e)
		case playAllows(s, play.Continue{}):
			e.doNow(play.Continue{})
		default:
			if _, ok := s.(play.Exploring); ok {
				t.Fatal("a lost party left exploring")
			}
			t.Fatalf("at rest in %T with nothing to do", s)
		}
	}
	t.Fatal("a lost fight never reached the party menu")
}

func isAssembling(s play.State) bool { _, ok := s.(play.Assembling); return ok }

// TestLastBlowBody: the blow that ends the fight (the party's last member
// falls) carries the body it leaves (ovr32:0E11's ground 0x1F, a party
// member's only: a monster vanishes) on Fell's Body, since no field
// follows to show it (BUG-CLASSES 55).
func TestLastBlowBody(t *testing.T) { testLastBlowBody(t, testEngine) }

func testLastBlowBody(t *testing.T, mk func(*testing.T) *Engine) {
	e := devFirstTurnOn(t, mk(t), "losefight", false)
	downs := 0
	for range 300 {
		var out play.Outcome
		switch s := e.build(); {
		case myTurn(s):
			out = endTurn(t, e)
		case downs > 0 || commonOf(s).Ask != nil || playAllows(s, play.Continue{}):
			if downs == 0 {
				t.Fatal("the fight ended with no Fell")
			}
			return
		default:
			t.Fatalf("at rest in %T with nothing to do", s)
		}
		for _, ev := range out.Events {
			if f, ok := ev.(play.Fell); ok {
				downs++
				if f.Body.Pixels == nil {
					t.Fatalf("a member's fall without its body: %+v", f)
				}
			}
		}
	}
	t.Fatal("no end")
}

// devLootOn is devLoot on the engine e (any game's: advEngine).
func devLootOn(t *testing.T, e *Engine) *Engine {
	t.Helper()
	e.Explore = true
	e.DevAction = e.DevScene("loot")
	seed := devSeed
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	for range 40 {
		if _, ok := e.build().(play.Looting); ok {
			seeLoot(t, e)
			return e
		}
		e.Step()
		press(e)
	}
	t.Fatal("no treasure screen")
	return nil
}

// TestLootEscLeaves: Leave on the treasure screen is the leave, as Exit
// (ovr06:1093: treasure left asks first), not "not now" while the screen's
// animation paces it (BUG-CLASSES 56). The scene's end gives the parent
// state back: exploring on the same square (BUG-CLASSES 60).
func TestLootEscLeaves(t *testing.T) { testLootEscLeaves(t, testEngine) }

func testLootEscLeaves(t *testing.T, mk func(*testing.T) *Engine) {
	e := devLootOn(t, mk(t))
	w := commonOf(e.build()).Where
	if _, err := e.doNow(play.Leave{}); err != nil {
		t.Fatal("Leave on the treasure:", err)
	}
	q := commonOf(e.build()).Ask
	if q == nil || !q.YesNo {
		t.Fatalf("Leave with treasure left: %+v, want a yes/no", q)
	}
	if _, err := e.doNow(no(q)); err != nil || e.waiting(PendLoot) { // No: leave the treasure
		t.Fatalf("leave: %v, still on the treasure %v", err, e.waiting(PendLoot))
	}
	x, ok := e.build().(play.Exploring)
	if !ok || x.Where.X != w.X || x.Where.Y != w.Y || x.Where.Area != w.Area || x.Ask != nil {
		t.Errorf("after the scene: %T at %+v ask %+v, want exploring at %+v", e.build(), x.Where, x.Ask, w)
	}
}

// TestTimedCastTaken: a Cast with its square for a spell cast over rounds
// (castfight's Cure Light Wounds "Begins Casting") is taken, not "not now"
// for the square: the turn is over; when the spell goes off its aim is
// offered (Aimed) and taken.
func TestTimedCastTaken(t *testing.T) {
	e := devFirstTurn(t, "castfight", false)
	a := e.member(e.cs.actor)
	at := image.Pt(e.state().CombatX(a), e.state().CombatY(a))
	if _, err := e.doNow(play.Cast{Member: -1, Spell: 3, Ats: []image.Point{at}}); err != nil || a.Combat.Spell == 0 {
		t.Fatalf("Cast: %v, casting %d", err, a.Combat.Spell)
	}
	for range 200 {
		s := e.build()
		f, ok := s.(play.Fighting)
		switch {
		case !ok:
			t.Fatalf("the fight ended (%T) before the spell went off", s)
		case e.cs.aim != nil:
			if len(f.Allowed.Aimed) == 0 {
				t.Fatalf("the spell's aim is up but nothing aimed is offered: %+v", f.Allowed)
			}
			act := f.Allowed.Aimed[0].Action
			if c, ok := act.(play.Cast); ok {
				c.Ats = []image.Point{at}
				act = c
			}
			if _, err := e.doNow(act); err != nil {
				t.Fatalf("%T at the aim: %v", act, err)
			}
			return
		case f.Ask != nil:
			e.doNow(yes(f.Ask))
		case myTurn(s):
			endTurn(t, e)
		default:
			t.Fatalf("at rest with nothing to do: %+v", f.Allowed)
		}
	}
	t.Fatal("the timed spell never asked its square")
}

// TestOpeningStepsAlone: a fight's opening (the foes' first moves) runs on
// with no input: nothing is asked before the party's turn.
func TestOpeningStepsAlone(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("losefight")
	seed := devSeed
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	var out play.Outcome
	e.stepOn(&out, nil) // as a Do's run on
	s := e.build()
	if q := commonOf(s).Ask; q != nil {
		t.Fatalf("the opening waits on %q", q.Text)
	}
	if !myTurn(s) {
		t.Fatalf("no party turn on its own: %T", s)
	}
}

// TestMovesHaveSquares: every Walked is a step of one square from its
// unit's real square, so an ended-fight replay never puts a unit on (0,0)
// (BUG-CLASSES 57).
func TestMovesHaveSquares(t *testing.T) {
	e := devFirstTurn(t, "combat", false)
	n := 0
	at := map[int]image.Point{}
	for range 400 {
		s := e.build()
		f, ok := s.(play.Fighting)
		if !ok {
			break
		}
		for _, u := range f.Field.Units {
			at[u.ID] = u.At
		}
		var out play.Outcome
		var err error
		switch {
		case f.Ask != nil:
			out, err = e.doNow(no(f.Ask))
		case myTurn(s):
			out, err = e.doNow(play.Quick{})
		default:
			t.Fatalf("at rest with nothing to do: %+v", f.Allowed)
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range out.Events {
			w, ok := ev.(play.Walked)
			if !ok {
				continue
			}
			n++
			for _, st := range w.Path {
				from, known := at[w.Unit]
				d := st.At.Sub(from)
				if st.At == (image.Point{}) || known && (d.X < -1 || d.X > 1 || d.Y < -1 || d.Y > 1 || d == (image.Point{})) {
					t.Fatalf("unit %d walked %v -> %v", w.Unit, from, st.At)
				}
				at[w.Unit] = st.At
			}
		}
	}
	if n == 0 {
		t.Fatal("no Walked")
	}
}

// TestStepNoRepeat: the engine's note goes out once, not again with
// nothing between ("SWORDSMAN Attacks" twice, 2 ms apart, with one attack).
func TestStepNoRepeat(t *testing.T) {
	e := devFirstTurn(t, "losefight", false)
	notes := 0
	for range 80 {
		var out play.Outcome
		switch s := e.build(); {
		case !myTurn(s) && commonOf(s).Ask == nil:
			if notes == 0 {
				t.Fatal("no message")
			}
			return
		case myTurn(s):
			out = endTurn(t, e)
		default:
			out, _ = e.doNow(yes(commonOf(s).Ask))
		}
		for i, ev := range out.Events {
			n, ok := ev.(play.Narrated)
			if !ok {
				continue
			}
			notes++
			if i > 0 && out.Events[i-1] == ev {
				t.Fatalf("%q twice with nothing between", n.Text)
			}
		}
	}
}

// seeLoot passes the treasure's "You found some loot!" (Looting.Announce).
func seeLoot(t *testing.T, e *Engine) {
	t.Helper()
	if l, ok := e.build().(play.Looting); ok && l.Announce {
		if _, err := e.doNow(play.Continue{}); err != nil {
			t.Fatal("Continue on the found treasure:", err)
		}
	}
}
