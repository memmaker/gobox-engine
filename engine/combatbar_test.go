package engine

import (
	"gobox/rules"
	"gobox/rules1e"
	"slices"
	"strings"
	"testing"

	"gobox/game"
)

// partyTurnScene: the combat dev scene up to its first party turn, Magic
// Missile memorized.
func partyTurnScene(t *testing.T) *Engine { t.Helper(); return partyTurnSceneOn(t, testEngine(t)) }

// partyTurnSceneOn is partyTurnScene on the engine e (any game's: advEngine).
func partyTurnSceneOn(t *testing.T, e *Engine) *Engine {
	t.Helper()
	e.Explore = true
	sc := e.DevScene("combat")
	e.DevAction = func() {
		sc()
		for _, c := range e.state().Party {
			c.Memorized[0] = rules1e.SpMagicMissile
		}
	}
	seed := uint32(2) // Felix's R3 rulings 1-2 (3d6+1 abilities, best-of-two creation HP) moved the Quick Start party's dice: seed 1's first turn has no foe in reach
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
	if !myTurn(e.build()) {
		t.Skip("no party turn")
	}
	return e
}

// fighting is the fight's state (the zero one when there is none).
func fighting(e *Engine) game.Fighting {
	f, _ := e.build().(game.Fighting)
	return f
}

// said is every Narrated text of o.
func said(o game.Outcome) []string {
	var l []string
	for _, ev := range o.Events {
		if n, ok := ev.(game.Narrated); ok {
			l = append(l, n.Text)
		}
	}
	return l
}

// canCast: a Cast is offered, plain or aimed.
func canCast(f game.Fighting, spell int) bool {
	for _, a := range f.Allowed.Plain {
		if c, ok := a.(game.Cast); ok && (spell == 0 || c.Spell == spell) {
			return true
		}
	}
	for _, a := range f.Allowed.Aimed {
		if c, ok := a.Action.(game.Cast); ok && (spell == 0 || c.Spell == spell) {
			return true
		}
	}
	return false
}

// TestCombatVerbs: the turn offers the original's verbs and no extras
// (combatMenu, ovr09:069E-07D4; its bar and letters are gobox's from
// Data.VerbOrder): the spells memorized to cast, Done's menu's Guard, Delay
// and its Quit (EndTurn, ovr09:0E85-0EE0), which ends the turn.
func TestCombatVerbs(t *testing.T) {
	e := partyTurnScene(t)
	f := fighting(e)
	for _, a := range f.Allowed.Plain {
		if _, ok := a.(game.DoManeuver); ok {
			t.Errorf("an extra offered: %T", a)
		}
	}
	if !canCast(f, rules1e.SpMagicMissile) {
		t.Errorf("Magic Missile not offered: %+v", f.Allowed)
	}
	for _, a := range []game.Action{game.Guard{}, game.Delay{}, game.EndTurn{}, game.Quick{}} {
		if !playAllows(f, a) {
			t.Errorf("%T not offered", a)
		}
	}
	unit := f.Field.Turn
	if _, err := e.doNow(game.EndTurn{}); err != nil {
		t.Fatal("EndTurn:", err)
	}
	if g := fighting(e); myTurn(g) && g.Field.Turn == unit && e.cs.aim == nil {
		t.Error("EndTurn did not end the turn")
	}
}

// TestFleeOffEdge: the actor walks (MoveTo, over its turns) to the map's
// edge; Flee asks the original's "Flee:" (ovr09:0C17) and Yes flees.
func TestFleeOffEdge(t *testing.T) {
	e := partyTurnScene(t)
	for range 40 {
		s := e.build()
		f, ok := s.(game.Fighting)
		if !ok {
			t.Skip("the fight ended before the edge")
		}
		if f.Ask != nil || !myTurn(s) {
			e.doNow(yes(f.Ask))
			continue
		}
		a := e.member(e.cs.actor)
		at := rules.BoardPos{X: e.state().CombatX(a), Y: e.state().CombatY(a)}
		for dir := 0; dir < 8; dir++ {
			if who, g := e.state().GroundAt(a, dir); who == 0 && g == 0 {
				if _, err := e.doNow(game.Flee{}); err != nil {
					t.Fatal("Flee:", err)
				}
				q := fightQ(e.build())
				if q == nil || q.Text != "Flee:" {
					t.Fatalf("a step off the map asks %+v, want Flee:", q)
				}
				o, err := e.doNow(yes(q))
				if err != nil {
					t.Fatal(err)
				}
				blocked := slices.Contains(said(o), "Escape is blocked")
				if note, _ := e.Note(); note == "Escape is blocked" {
					blocked = true
				}
				if a.Status != rules1e.StatusRunning && !blocked { // tryFlee: away, or blocked by a faster foe (the original's dice)
					t.Errorf("Yes neither fled nor was blocked: status %d, said %q", a.Status, said(o))
				}
				return
			}
		}
		b := f.Field
		edge := func(p rules.BoardPos) int { return min(p.X, p.Y, b.Size.W-1-p.X, b.Size.H-1-p.Y) }
		best := at
		for _, q := range f.Allowed.Moves {
			if edge(q.At) < edge(best) {
				best = q.At
			}
		}
		if best == at {
			e.doNow(game.EndTurn{})
			continue
		}
		e.doNow(game.MoveTo{To: best})
	}
	t.Skip("never reached the edge")
}

// TestVerbOrder: the fight bar shows the original's verbs without View and
// Aim (removed at Felix's request); Aim keeps its A key.
func TestVerbOrder(t *testing.T) {
	d := testEngine(t).gameData
	want := []string{"Move", "Use", "Cast", "Turn", "Quick", "Done"}
	if !slices.Equal(d.VerbOrder, want) {
		t.Errorf("VerbOrder %v, want %v", d.VerbOrder, want)
	}
	for _, v := range turnBar { // every shown bar verb has its place
		if v.o.Label != "Aim" && !slices.Contains(want, v.o.Label) {
			t.Errorf("%s not in VerbOrder", v.o.Label)
		}
	}
}

// TestBandage: with a member dying, Done's menu offers Bandage
// (ovr09:0EE0); taking it binds the dying one (unconscious, no longer
// dying), as the original's bandage does, and no Bandage without one.
func TestBandage(t *testing.T) {
	e := partyTurnScene(t)
	bandage := func() bool {
		return slices.ContainsFunc(fighting(e).Allowed.Plain, func(a game.Action) bool { _, ok := a.(game.Bandage); return ok })
	}
	if bandage() {
		t.Fatal("Bandage with nobody dying")
	}
	var d *rules1e.Character
	for _, c := range e.state().Party {
		if c.Side == rules1e.SideParty && c != e.member(e.cs.actor) && c.Combat.Extra == 0 {
			d = c
			break
		}
	}
	if d == nil {
		t.Fatal("no other member")
	}
	d.HP, d.Status = 0, rules1e.StatusDying
	e.turnChoices(e.charID(e.member(e.cs.actor))) // fallen before this turn: its options built anew
	if !bandage() {
		t.Fatalf("no Bandage with %s dying: %+v", d.Name(), fighting(e).Allowed.Plain)
	}
	o, err := e.doNow(game.Bandage{Member: max(e.sel, 0)})
	if err != nil {
		t.Fatalf("Bandage: %v", err)
	}
	if d.Status != rules1e.StatusUnconscious || d.Combat.Dying != 0 {
		t.Errorf("%s after Bandage: status %d, dying %d", d.Name(), d.Status, d.Combat.Dying)
	}
	for _, c := range e.state().Party {
		if c.Side == rules1e.SideParty && c.Status == rules1e.StatusDying {
			t.Errorf("%s still dying after %s's Bandage", c.Name(), d.Name())
		}
	}
	if !slices.ContainsFunc(said(o), func(l string) bool { return strings.Contains(l, d.Name()+" is bandaged") }) {
		t.Errorf("no '%s is bandaged' in %q", d.Name(), said(o))
	}
	// the round's end (ovr09:0880) says "Your Teammate is Dying" only while one is
	round := e.state().CombatRound
	var lines []string
	for i := 0; i < 200 && e.state().CombatRound == round; i++ {
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			break
		}
		var o game.Outcome
		if q := commonOf(s).Ask; q != nil {
			o, _ = e.doNow(yes(q))
		} else {
			o, _ = e.doNow(game.EndTurn{})
		}
		lines = append(lines, said(o)...)
	}
	if e.state().CombatRound == round {
		t.Fatal("the round never ended")
	}
	dying := slices.ContainsFunc(e.state().Party, func(c *rules1e.Character) bool { return c.Side == rules1e.SideParty && c.Status == rules1e.StatusDying })
	if slices.Contains(lines, "Your Teammate is Dying") && !dying {
		t.Error("'Your Teammate is Dying' at the round's end with nobody dying")
	}
}

// TestTurnWeapon (BUG-CLASSES 39): a member with a weapon readied
// publishes it on its turn (Field.Weapon.Name, the item's list name), as
// the original's panel shows it.
func TestTurnWeapon(t *testing.T) { testTurnWeapon(t, testEngine) }

func testTurnWeapon(t *testing.T, mk func(*testing.T) *Engine) {
	e := partyTurnSceneOn(t, mk(t))
	c := e.member(e.cs.actor)
	w := c.Slots[rules1e.SlotWeapon]
	if w == nil { // readied before the turn: its options built anew
		w = e.state().AddItemCopy(c, e.state().RandomItem(weaponFor(t, e, c, rules1e.ItemLongSword, false)))
		e.state().ReadyItem(w)
		e.state().ApplyItems(c)
		e.turnChoices(e.charID(c))
	}
	if c.Slots[rules1e.SlotWeapon] != w {
		t.Fatalf("%s readied no weapon", c.Name())
	}
	tw := fighting(e).Field.Weapon
	if tw == nil || tw.Name == "" || tw.Name != e.invItem(e.charID(c), e.itemHandle(w)).Name {
		t.Errorf("%s's turn: weapon %+v, want %q", c.Name(), tw, e.invItem(e.charID(c), e.itemHandle(w)).Name)
	}
}

// TestBarConditions: the turn offers each verb on the original's condition
// (combatMenu ovr09:069E-07D4; Use only with an item to use now, Felix's
// ruling: the original offers it with any item carried, +0xC7), Cast
// with any of the 20 memorized (+0x17, may cast, no no-magic area), Turn
// for a cleric (+0x96) who has not turned (+0x11), undead about or not.
func TestBarConditions(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	type verbs struct{ use, cast, turn bool }
	offered := func() verbs {
		e.turnChoices(e.charID(c))
		f := fighting(e)
		return verbs{
			use:  slices.ContainsFunc(f.Allowed.Plain, func(a game.Action) bool { _, ok := a.(game.UseItem); return ok }),
			cast: canCast(f, 0),
			turn: playAllows(f, game.TurnUndead{}),
		}
	}
	c.Memorized = [len(c.Memorized)]rules1e.SpellID{}
	c.Items = nil
	e.state().Pool = nil
	e.state().ApplyItems(c)
	c.Levels[rules1e.ClsCleric] = 0
	if v := offered(); v.use || v.cast || v.turn {
		t.Errorf("nothing carried, memorized, no cleric: %+v", v)
	}
	e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemLongSword)) // not usable, but carried
	c.Memorized[19] = rules1e.SpMagicMissile                              // the 20th slot
	c.Combat.CanCast = 1
	c.Levels[rules1e.ClsCleric] = 1
	c.Combat.TurnUsed = 0
	if v := offered(); v.use || !v.cast || v.turn {
		t.Errorf("an item with no use, a spell, a cleric with no undead about: %+v", v)
	}
	e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemWand))
	if v := offered(); !v.use {
		t.Errorf("a wand carried: %+v", v)
	}
	var foes []*rules1e.Character
	for _, o := range e.state().Party {
		if o.Side != c.Side {
			foes = append(foes, o)
		}
	}
	for _, o := range foes {
		o.Undead = 1 // the weakest row of the turning table
	}
	if v := offered(); !v.turn {
		t.Errorf("a cleric with undead about: %+v", v)
	}
	c.Combat.TurnUsed = 1
	if v := offered(); v.turn {
		t.Errorf("Turn once used: %+v", v)
	}
}

// TestWaitAndRanged: Wait only with someone still to act this round (the
// last one would come straight back), Ranged only with a launcher and its
// ammunition, or a missile, readied.
func TestWaitAndRanged(t *testing.T) {
	e := partyTurnScene(t)
	c := e.member(e.cs.actor)
	offered := func() (wait, ranged bool) {
		e.turnChoices(e.charID(c))
		f := fighting(e)
		return playAllows(f, game.Delay{}), slices.ContainsFunc(f.Allowed.Aimed, func(a game.Aimed) bool { _, ok := a.Action.(game.Shoot); return ok })
	}
	c.Slots[rules1e.SlotLauncher], c.Slots[rules1e.SlotMissiles] = nil, nil
	if wait, ranged := offered(); !wait || ranged {
		t.Errorf("others to act, nothing to shoot: wait %v ranged %v", wait, ranged)
	}
	for _, o := range e.state().Party {
		if o != c {
			o.Combat.Delay = 0
		}
	}
	bow := e.state().RandomItem(rules1e.ItemShortBow)
	c.Slots[rules1e.SlotLauncher], c.Slots[rules1e.SlotArrows] = bow, nil
	if wait, ranged := offered(); wait || ranged {
		t.Errorf("the last to act, a bow without arrows: wait %v ranged %v", wait, ranged)
	}
	c.Slots[rules1e.SlotArrows] = e.state().RandomItem(rules1e.ItemArrow)
	if _, ranged := offered(); !ranged {
		t.Error("a bow with arrows: no Ranged")
	}
}

// TestReadyInFight (BUG-CLASSES 40): mid-fight the acting member's Ready
// of its own weapon is offered and taken, a second one unreadies it (the
// fight's item menu toggles), and Field.Weapon follows; another member's
// items are not its to ready.
func TestReadyInFight(t *testing.T) { testReadyInFight(t, testEngine) }

func testReadyInFight(t *testing.T, mk func(*testing.T) *Engine) {
	e := partyTurnSceneOn(t, mk(t))
	c := e.member(e.cs.actor)
	m := e.state().PartyIndex(c)
	if m < 0 {
		t.Fatal("the actor is not in the party")
	}
	w := e.state().AddItemCopy(c, e.state().RandomItem(weaponFor(t, e, c, rules1e.ItemLongSword, false)))
	if h := c.Slots[rules1e.SlotWeapon]; h != nil {
		e.state().ReadyItem(h)
	}
	e.state().ApplyItems(c)
	i := slices.Index(e.state().Carried(c), w)
	if i < 0 {
		t.Fatal("the new sword is not carried")
	}
	ready := game.Ready{Member: m, Item: e.itemHandle(w)}
	if !playAllows(e.build(), ready) {
		t.Errorf("no Ready for the acting member: %+v", fighting(e).Allowed.Plain)
	}
	weapon := func() string {
		if tw := fighting(e).Field.Weapon; tw != nil {
			return tw.Name
		}
		return ""
	}
	if _, err := e.doNow(ready); err != nil || w.Readied == 0 {
		t.Fatalf("Ready: %v, readied %d", err, w.Readied)
	}
	if n := weapon(); n != e.invItem(e.charID(c), e.itemHandle(w)).Name {
		t.Errorf("Field.Weapon %q after Ready, want %q", n, e.invItem(e.charID(c), e.itemHandle(w)).Name)
	}
	if _, err := e.doNow(game.Unready{Member: m, Item: ready.Item}); err != nil || w.Readied != 0 { // what the front sends for a readied item
		t.Fatalf("Unready: %v, readied %d", err, w.Readied)
	}
	if n := weapon(); n != "" {
		t.Errorf("Field.Weapon %q after unreadying", n)
	}
	for o, oc := range e.state().Party {
		if oc != c && oc.Side == rules1e.SideParty {
			if _, err := e.doNow(game.Ready{Member: o, Item: e.listedItems(e.partyIDs()[o])[0]}); err == nil {
				t.Errorf("Ready of %s's items on %s's turn taken", oc.Name(), c.Name())
			}
			break
		}
	}
}

// TestMoveStep: a MoveTo a free square of the turn's Moves walks there; an
// Attack at a foe's square next to it (gobox's bump, moveAttack ovr09:09DF)
// attacks it and stays. A wall is never among the Moves (ovr09:0C6B's
// "Blocked" is gobox's not to send).
func TestMoveStep(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("castfight")
	seed := uint32(5) // Felix's R3 rulings 1-2 moved the Quick Start party's dice: seed 1's first actor stands by no wall
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
	if !myTurn(e.build()) {
		t.Skip("no party turn")
	}
	a := e.member(e.cs.actor)
	at := rules.BoardPos{X: e.state().CombatX(a), Y: e.state().CombatY(a)}
	step := func(dir int) rules.BoardPos {
		return at.Add(rules.BoardPos{X: rules1e.DirDX[dir], Y: rules1e.DirDY[dir]})
	}
	inReach := func(p rules.BoardPos) bool { return playAllows(e.build(), game.MoveTo{To: p}) }
	free, walled := -1, -1
	for dir := 0; dir < 8; dir++ {
		who, g := e.state().GroundAt(a, dir)
		switch {
		case who == 0 && g != 0 && inReach(step(dir)) && free < 0:
			free = dir
		case who == 0 && g != 0 && !inReach(step(dir)) && walled < 0:
			walled = dir
		}
	}
	if free < 0 || walled < 0 {
		t.Fatalf("no free (%d) or walled (%d) square next to %s", free, walled, a.Name())
	}
	if _, err := e.doNow(game.MoveTo{To: step(walled)}); err == nil {
		t.Errorf("a move into the wall at %v taken", step(walled))
	}
	var foe *rules1e.Character // a foe on the free square: the bump attacks it
	for i := 1; i <= e.state().Ncomb; i++ {
		if c := e.state().Combatant(byte(i)); c != nil && c.Side != a.Side && c.Okay != 0 {
			foe = c
			break
		}
	}
	if foe == nil {
		t.Fatal("no foe")
	}
	fi := e.state().CombatIndex(foe)
	p := step(free)
	e.state().Occ[e.state().Cpos[fi].Y][e.state().Cpos[fi].X] = 0
	e.state().Cpos[fi].X, e.state().Cpos[fi].Y = byte(p.X), byte(p.Y)
	e.state().Occ[p.Y][p.X] = rules1e.Slot(byte(fi))
	e.turnChoices(e.charID(a))
	o, err := e.doNow(game.Attack{At: &p})
	if err != nil {
		t.Fatalf("bump into %s: %v", foe.Name(), err)
	}
	attacked := slices.ContainsFunc(o.Events, func(ev game.Event) bool { x, ok := ev.(game.Attacked); return ok && x.To == fi })
	if !attacked || e.state().CombatX(a) != at.X || e.state().CombatY(a) != at.Y {
		t.Errorf("bump into %s: attacked %v, now at %d,%d", foe.Name(), attacked, e.state().CombatX(a), e.state().CombatY(a))
	}
}

// TestAllyBump (BUG-CLASSES 41): an Attack at an ally's square asks the
// original's "Attack Ally?" (a Question, not a silent answer); No does
// nothing, Yes attacks and turns the party on its friend.
func TestAllyBump(t *testing.T) { testAllyBump(t, testEngine) }

func testAllyBump(t *testing.T, mk func(*testing.T) *Engine) {
	e := partyTurnSceneOn(t, mk(t))
	a := e.member(e.cs.actor)
	var to rules.BoardPos
	found := false
	for dir := 0; dir < 8 && !found; dir++ {
		if who, _ := e.state().GroundAt(a, dir); who > 0 && e.state().Combatant(who).Side == a.Side {
			to, found = rules.BoardPos{X: e.state().CombatX(a) + rules1e.DirDX[dir], Y: e.state().CombatY(a) + rules1e.DirDY[dir]}, true
		}
	}
	if !found {
		t.Skip("no ally next to the actor")
	}
	bump := func() *game.Question {
		t.Helper()
		if _, err := e.doNow(game.Attack{At: &to}); err != nil {
			t.Fatal(err)
		}
		q := fightQ(e.build())
		if q == nil || q.Text != "Attack Ally?" {
			t.Fatalf("step into an ally asks %+v, want Attack Ally?", q)
		}
		return q
	}
	if _, err := e.doNow(no(bump())); err != nil || e.state().AttackedAlly != 0 {
		t.Fatalf("No: %v, attacked %d", err, e.state().AttackedAlly)
	}
	if _, err := e.doNow(yes(bump())); err != nil || e.state().AttackedAlly == 0 { // asked again: No took nothing
		t.Fatalf("Yes: %v, attacked %d", err, e.state().AttackedAlly)
	}
}

// TestDropOnce (BUG-CLASSES 47): Drop drops the item at once; gobox asked
// ovr20's "Drop It?" already, so there is no second question.
func TestDropOnce(t *testing.T) { testDropOnce(t, testEngine) }

func testDropOnce(t *testing.T, mk func(*testing.T) *Engine) {
	e := partyTurnSceneOn(t, mk(t))
	c := e.member(e.cs.actor)
	m := e.state().PartyIndex(c)
	if m < 0 {
		t.Fatal("the actor is not in the party")
	}
	w := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemLongSword))
	i := slices.Index(e.state().Carried(c), w)
	if i < 0 {
		t.Fatal("the new sword is not carried")
	}
	if _, err := e.doNow(game.Drop{Member: m, Item: e.itemHandle(w)}); err != nil || slices.Contains(e.state().Carried(c), w) || commonOf(e.build()).Ask != nil {
		t.Fatalf("Drop: %v, still carried %v, ask %+v", err, slices.Contains(e.state().Carried(c), w), commonOf(e.build()).Ask)
	}
}
