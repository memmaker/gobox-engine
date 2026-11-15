package engine

import (
	"gobox/rules"
	"gobox/rules1e"
	"slices"
	"testing"

	"gobox/game"
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
	e.attackUnit(e.charID(c), 0, nil, false)
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
	j := e.state().AddItemCopy(c, e.state().RandomItem(weaponFor(t, e, c, rules1e.ItemJavelin, true))) // readied in the Missiles slot (thrown only) or in hand (melee too), Ranged throws it: an aim, nobody in range needed
	e.state().ReadyItem(j)
	e.state().ApplyItems(c)
	if c.Slots[rules1e.SlotMissiles] != j && c.Slots[rules1e.SlotWeapon] != j {
		t.Fatalf("thrown weapon readied %d, not in the Missiles slot or in hand", j.Readied)
	}
	e.attackUnit(e.charID(c), 0, nil, true)
	if e.cs.aim == nil {
		t.Fatal("Ranged with a thrown weapon opened no aim")
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
	at := rules.BoardPos{X: e.state().CombatX(ally), Y: e.state().CombatY(ally)}
	ask := func() *game.Question {
		t.Helper()
		if _, err := e.doNow(game.Attack{At: &at}); err != nil {
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

// TestShootSquaresAreTargets: the Ranged action's squares are the ones its
// aim takes, each with a figure on it (no bare ground, not the whole board,
// not the shooter's own): what gobox highlights, and none is nothing in reach.
func TestShootSquaresAreTargets(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	for _, h := range []int{rules1e.SlotWeapon, rules1e.SlotShield, rules1e.SlotMissiles, rules1e.SlotArrows, rules1e.SlotBolts} {
		if it := c.Slots[h]; it != nil {
			e.state().ReadyItem(it)
		}
	}
	e.state().ApplyItems(c)
	j := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemDart)) // thrown only: readied in the Missiles slot, the hand empty
	if _, err := e.doNow(game.Ready{Member: e.state().PartyIndex(c), Item: e.itemHandle(j)}); err != nil {
		t.Fatal(err)
	}
	f := fighting(e)
	i := slices.IndexFunc(f.Allowed.Aimed, func(a game.Aimed) bool { _, ok := a.Action.(game.Shoot); return ok })
	if i < 0 {
		t.Fatalf("no Shoot with a thrown weapon readied: %+v", f.Allowed.Aimed)
	}
	sq := f.Allowed.Aimed[i].Squares
	if len(sq) == 0 || len(sq) >= f.Field.Size.W*f.Field.Size.H {
		t.Fatalf("%d squares to shoot at on a %v board", len(sq), f.Field.Size)
	}
	me := e.at(e.cs.actor)
	if !slices.ContainsFunc(sq, func(at rules.BoardPos) bool { d := at.Sub(me); return max(d.X, -d.X, d.Y, -d.Y) > 1 }) {
		t.Errorf("shoot offers only the squares next to the thrower: %v from %v (the reach is the dart's, not the empty hand's)", sq, me)
	}
	if c.Slots[rules1e.SlotWeapon] == j {
		t.Error("the state left the dart drawn in the weapon hand")
	}
	for _, at := range sq {
		on := slices.IndexFunc(f.Field.Units, func(u game.Unit) bool {
			return !u.Dead && !u.Gone && (u.At == at || slices.Contains(u.Squares, at))
		})
		if on < 0 || f.Field.Units[on].ID == f.Field.Turn {
			t.Errorf("shoot offers %v: no figure to shoot there", at)
		}
	}
}

// TestTurnWeaponChoice: a turn fights with the weapon in hand or with its
// Launcher or Missiles slot, one choice (rules1e ChooseWeapon, the
// computer's and the player's alike): it sets the turn's attacks by the
// weapon's rate, and once the turn has struck it stands, the other weapon
// refused and no longer offered.
func TestTurnWeaponChoice(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	s := e.state()
	for _, h := range []int{rules1e.SlotWeapon, rules1e.SlotShield, rules1e.SlotMissiles, rules1e.SlotArrows, rules1e.SlotBolts} {
		if it := c.Slots[h]; it != nil {
			s.ReadyItem(it)
		}
	}
	s.ApplyItems(c)
	darts := s.AddItemCopy(c, s.RandomItem(rules1e.ItemDart))
	darts.Count = 9
	if _, err := e.doNow(game.Ready{Member: s.PartyIndex(c), Item: e.itemHandle(darts)}); err != nil {
		t.Fatal(err)
	}
	melee := c.Attacks[0]
	if !s.ChooseWeapon(c, true) || !c.Combat.Ranged || s.TurnWield(c).Weapon != darts {
		t.Fatalf("ranged not chosen: ranged %v", c.Combat.Ranged)
	}
	if c.Attacks[0] <= melee {
		t.Errorf("the darts give %d attacks, the hand %d: want the darts' rate", c.Attacks[0], melee)
	}
	if !s.ChooseWeapon(c, false) || c.Attacks[0] != melee { // not struck yet: the choice may still change
		t.Fatalf("back to the hand before a blow: ranged %v, attacks %d (hand %d)", c.Combat.Ranged, c.Attacks[0], melee)
	}
	s.ChooseWeapon(c, true)
	c.Combat.Attacked = 1 // the turn has struck with the darts
	if s.ChooseWeapon(c, false) || !c.Combat.Ranged {
		t.Fatal("the hand taken after the turn struck with the darts")
	}
	e.rebuild()
	f := fighting(e)
	if slices.ContainsFunc(f.Allowed.Plain, func(a game.Action) bool { _, ok := a.(game.Attack); return ok }) {
		t.Error("a melee Attack is still offered after the turn struck with the darts")
	}
	if e.attackUnit(e.charID(c), 0, nil, false); e.cs.aim != nil {
		t.Error("the melee aim opened after the turn struck with the darts")
	}
}
