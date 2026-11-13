package engine

import (
	"gobox/rules"
	"gobox/rules1e"
	"image"
	"testing"

	"gobox/game/play"
)

// TestAimOffersAccepted (BUG-CLASSES 17): in a real fight, every foe the aim
// offers (its Targets, what Tab cycles and gobox shows) is a square the aim
// accepts: pointing at it gives no refusal.
func TestAimOffersAccepted(t *testing.T) { testAimOffersAccepted(t, testEngine) }

func testAimOffersAccepted(t *testing.T, mk func(*testing.T) *Engine) {
	e := mk(t)
	e.Explore = true
	e.DevAction = e.DevScene("combat")
	seed := uint32(1)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	for range 400 {
		if e.cs.turn != nil && e.cs.actor != rules.NoChar && e.cs.aim == nil {
			break
		}
		if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	c := e.member(e.cs.actor)
	if c == nil || e.cs.turn == nil {
		t.Skip("no party turn")
	}
	check := func(what string, rng, showRange, ground int) {
		e.playerAim(e.charID(c), rng, showRange, ground, aimThen{kind: atNothing})
		if e.cs.aim == nil {
			t.Fatalf("%s: no aim", what)
		}
		for _, id := range e.snap.targets {
			f := e.state().Combatant(byte(id))
			x, y := e.state().CombatX(f), e.state().CombatY(f)
			e.cs.aim.apply(TurnAct{Verb: VerbPoint, X: x, Y: y})
			if e.snap.aimWhy != "" {
				t.Errorf("%s: offered %s at %d,%d refused: %s", what, f.Name(), x, y, e.snap.aimWhy)
			}
		}
		e.aimApply(TurnAct{Verb: VerbClose})
	}
	check("attack", 0xFF, 1, 0) // as the original: allies too, with or without a foe at hand
	e.attackUnit(e.charID(c), 0, nil)
	if e.cs.aim == nil { // parity: no "No foe in reach" before the aim
		t.Fatal("Attack opened no aim")
	}
	e.aimApply(TurnAct{Verb: VerbClose})
	for _, h := range []int{rules1e.SlotWeapon, rules1e.SlotShield} { // empty hands (a game's starting gear may fill them)
		if it := c.Slots[h]; it != nil {
			e.state().ReadyItem(it)
		}
	}
	e.state().ApplyItems(c)
	j := e.state().AddItemCopy(c, e.state().RandomItem(weaponFor(t, e, c, rules1e.ItemJavelin, true))) // readied in hand, Attack throws it: an aim, nobody in range needed
	e.state().ReadyItem(j)
	e.state().ApplyItems(c)
	if c.Slots[rules1e.SlotWeapon] != j {
		t.Fatalf("javelin readied %d, not in hand", j.Readied)
	}
	e.attackUnit(e.charID(c), 0, nil)
	if e.cs.aim == nil {
		t.Fatal("Attack with a javelin opened no aim")
	}
	e.aimApply(TurnAct{Verb: VerbClose})
	check("spell range 3", 3, 0, 1)
}

// TestAllyAskKeepsAim (BUG-CLASSES 18): the aim's sub-question (Attack Ally?)
// is the fight's pending Ask: the fight stays up, No takes nothing and ends the aim (back to the turn's menu, as
// the original's aim loop: ovr14:3212), Yes attacks.
func TestAllyAskKeepsAim(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("combat")
	seed := uint32(5) // Felix's R3 rulings 1-2 (3d6+1 abilities, best-of-two creation HP) moved the dice: seed 1's first turn has no foe in reach, so no Attack
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	for range 400 {
		if e.cs.turn != nil && e.cs.actor != rules.NoChar && e.cs.aim == nil {
			break
		}
		if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	c := e.member(e.cs.actor)
	if c == nil || e.cs.turn == nil {
		t.Skip("no party turn")
	}
	var ally *rules1e.Character
	for _, m := range e.state().Party {
		if m != c && e.state().RangeTo(c, m) <= 1 {
			ally = m
		}
	}
	if ally == nil {
		t.Skip("no ally next to the actor")
	}
	at := image.Pt(e.state().CombatX(ally), e.state().CombatY(ally))
	ask := func() *play.Question {
		t.Helper()
		if _, err := e.doNow(play.Attack{At: &at}); err != nil {
			t.Fatalf("Attack at the ally: %v", err)
		}
		q := fightQ(e.build())
		if q == nil || q.Text != "Attack Ally?" || !q.NoEsc || !q.NoPick { // Y/N only (ovr26:10D1)
			t.Fatalf("Attack at the ally asks %+v, want Attack Ally?", q)
		}
		if e.state().AttackedAlly != 0 {
			t.Fatalf("attacked %d before the answer", e.state().AttackedAlly)
		}
		return q
	}
	q := ask()
	if _, err := e.doNow(no(q)); err != nil || e.cs.aim != nil || e.state().AttackedAlly != 0 || !myTurn(e.build()) { // No: the aim ends, the turn's menu again (ovr14:3212, the T/Enter left in the exit set 2DA7)
		t.Fatalf("No: %v, aim %v, attacked %d", err, e.cs.aim != nil, e.state().AttackedAlly)
	}
	q = ask()                                                                                    // aimed again: No took nothing
	if _, err := e.doNow(yes(q)); err != nil || e.cs.aim != nil || e.state().AttackedAlly == 0 { // Yes
		t.Fatalf("Yes: %v, aim %v, attacked %d", err, e.cs.aim != nil, e.state().AttackedAlly)
	}
}
