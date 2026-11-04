package parity

import (
	"reflect"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"por/rules"
)

// sameBattle says classic's combat menu, combatant info and aim equal por's.
func sameBattle(t *testing.T, name string, e *rules.Engine) front.View {
	t.Helper()
	v, s := settled(e, front.Action{}), e.Rules().State()
	if v.Battle == nil || s.Combat == nil || s.Mode != game.ModeCombat {
		t.Fatalf("%s: no fight: %v %+v", name, s.Mode, s.Combat)
	}
	if c := classic.View(v, s); !reflect.DeepEqual(c.Battle, v.Battle) {
		t.Errorf("%s: battle\n got %+v\nwant %+v", name, c.Battle, v.Battle)
	}
	return v
}

func TestCombatParity(t *testing.T) {
	e := engine(t, true, "combat")
	v := sameBattle(t, "combat", e)
	s := e.Rules().State()
	if s.Combat.Turn == nil || len(s.Combat.Turn.Options) == 0 || len(s.Combat.Units) != len(v.Battle.Units) || len(s.Combat.Reach) == 0 || len(s.Combat.Targets) == 0 {
		t.Fatalf("combat state %+v", s.Combat)
	}
	if _, ok := s.Combat.Turn.Options[0].Do.(game.Attack); !ok {
		t.Errorf("first option %+v", s.Combat.Turn.Options[0])
	}
	for range 8 { // the turns go by: each menu and each unit's figures as por's
		next(t, e, game.EndTurn{})
		if sameBattle(t, "turn", e); e.Rules().State().Combat.Turn == nil {
			break
		}
	}
	next(t, e, game.Guard{})
	sameBattle(t, "guard", e)
}

// next applies d and lets the engine run on, passing a pause as the presenter does.
func next(t *testing.T, e *rules.Engine, d game.Decision) {
	t.Helper()
	if res := e.Rules().Apply(d); res.Err != nil {
		t.Fatalf("%T: %v", d, res.Err)
	}
	for v := settled(e, front.Action{}); v.Ask != nil && v.Ask.Kind == front.AskPause; v = settled(e, front.Action{Kind: front.ActOK}) {
	}
}

func TestCastParity(t *testing.T) {
	e := engineSeed(t, true, "castfight", 1) // the seed whose caster is aimed (rules TestCastAimed)
	sameBattle(t, "castfight", e)
	s := e.Rules().State()
	if len(s.Combat.Spells) == 0 || len(s.Combat.Aims) == 0 {
		t.Fatalf("spells %+v aims %+v", s.Combat.Spells, s.Combat.Aims)
	}
	a := s.Combat.Aims[0]
	if a.Targets < 1 || a.Range < 0 {
		t.Errorf("aim %+v", a)
	}
	next(t, e, game.Cast{Spell: a.ID})
	for range 20 { // the others' turns go by until the caster's spell goes off, and is aimed
		if c := e.Rules().State().Combat; c == nil || c.Turn == nil || c.Turn.Aim {
			break
		}
		next(t, e, game.EndTurn{})
	}
	sameBattle(t, "casting", e)
	s = e.Rules().State()
	if !s.Combat.Turn.Aim || s.Combat.Turn.Spell != a.Name {
		t.Fatalf("not aiming %q: %+v", a.Name, s.Combat.Turn)
	}
	self := s.Combat.Units[0].At
	for _, u := range s.Combat.Units {
		if u.Active {
			self = u.At
		}
	}
	next(t, e, game.Aim{At: self})
	sameBattle(t, "cast", e)
}

// TestBattleView: the adapter's Battle (built from State; gobox composes the
// icons) has an icon for every unit, a picture for every square's tile, and
// some party, foes and walls.
func TestBattleView(t *testing.T) {
	e := engine(t, true, "combat")
	g := classic.NewGame(e.Rules(), e.Data())
	g.Answer(front.Action{})
	b := g.Settle().Battle
	if b == nil {
		t.Fatal("no Battle in combat")
	}
	party, foes, walls := 0, 0, 0
	for i, u := range b.Units {
		if u.Icon == nil && b.IconPix[i].Pix == nil {
			t.Errorf("unit %+v has no icon", u)
		}
		if u.Party {
			party++
		} else {
			foes++
		}
	}
	for _, row := range b.Ground {
		for _, s := range row {
			if s.Wall {
				walls++
			}
			if b.TilePix[s.Tile].Pix == nil {
				t.Fatalf("square tile %d missing from Tiles", s.Tile)
			}
		}
	}
	if party == 0 || foes == 0 || walls == 0 {
		t.Errorf("party %d foes %d walls %d: want some of each", party, foes, walls)
	}
}
