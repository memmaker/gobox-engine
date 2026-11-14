package engine

import (
	"image"
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// partyTurn4e is partyTurnScene with "4e Combat" handed in.
func partyTurn4e(t *testing.T) *Engine {
	t.Helper()
	e := testEngine(t)
	e.HandIn(Settings{Combat4e: true})
	return partyTurnSceneOn(t, e)
}

// maneuversOffered are the DoManeuver IDs f offers, Plain and Aimed.
func maneuversOffered(f game.Fighting) (plain, aimed []int) {
	for _, a := range f.Allowed.Plain {
		if m, ok := a.(game.DoManeuver); ok {
			plain = append(plain, m.ID)
		}
	}
	for _, a := range f.Allowed.Aimed {
		if m, ok := a.Action.(game.DoManeuver); ok {
			aimed = append(aimed, m.ID)
		}
	}
	return
}

// TestManeuvers4e: under 4e every turn maneuver is offered (by ID), a self
// one and an aimed one go through Do; classic refuses DoManeuver.
func TestManeuvers4e(t *testing.T) {
	const secondWind = 1 // rules4e: every class's AimSelf
	e := partyTurn4e(t)
	f := fighting(e)
	plain, aimed := maneuversOffered(f)
	if len(f.Field.Maneuvers) == 0 || len(plain)+len(aimed) != len(f.Field.Maneuvers) {
		t.Fatalf("offered plain %v aimed %v of %+v", plain, aimed, f.Field.Maneuvers)
	}
	if _, err := e.Do(game.DoManeuver{ID: secondWind}); err != nil {
		t.Fatal("Second Wind:", err)
	}
	if g := fighting(e); myTurn(g) && g.Field.Turn == f.Field.Turn {
		if p, _ := maneuversOffered(g); slices.Contains(p, secondWind) {
			t.Error("Second Wind (once a fight) offered again")
		}
	}

	e = partyTurn4e(t)
	f = fighting(e)
	i := slices.IndexFunc(f.Field.Units, func(u game.Unit) bool { return u.Foe && !u.Dead && !u.Gone })
	if i < 0 {
		t.Skip("no foe to aim at")
	}
	foe := f.Field.Units[i].At
	plain, aimed = maneuversOffered(f)
	id := -1
	if len(aimed) > 0 {
		id = aimed[0]
	} else {
		for _, p := range plain {
			if p != secondWind {
				id = p
			}
		}
	}
	if id < 0 {
		t.Skip("no aimed maneuver")
	}
	if len(aimed) == 0 { // a weapon maneuver: its attack's aim comes up and is published
		if _, err := e.Do(game.DoManeuver{ID: id}); err != nil {
			t.Fatal("weapon maneuver:", err)
		}
		if e.cs.aim != nil && len(fighting(e).Allowed.Aimed) == 0 {
			t.Fatal("the maneuver's aim is up but not published")
		}
		if e.cs.aim == nil {
			return
		}
		if _, err := e.Do(game.Shoot{Ats: []image.Point{foe}}); err != nil {
			t.Fatal("its attack:", err)
		}
	} else if _, err := e.Do(game.DoManeuver{ID: id, Ats: []image.Point{aimAt(f, id, foe)}}); err != nil {
		t.Fatal("aimed maneuver:", err)
	}
	if e.cs.aim != nil {
		t.Errorf("maneuver %d: the aim (%d) is still up", id, e.cs.aim.then.kind)
	}
}

// TestWeaponManeuver4e: a weapon maneuver (Plain) puts its attack's aim
// up, published Aimed, and Shoot at a foe takes it.
func TestWeaponManeuver4e(t *testing.T) {
	e := partyTurn4e(t)
	f := fighting(e)
	id := -1
	for _, m := range f.Field.Maneuvers {
		if rules.ManeuverAim(m.Aim) == rules.AimWeapon {
			id = m.ID
		}
	}
	i := slices.IndexFunc(f.Field.Units, func(u game.Unit) bool { return u.Foe && !u.Dead && !u.Gone })
	if id < 0 || i < 0 {
		t.Skip("no weapon maneuver or foe")
	}
	if _, err := e.Do(game.DoManeuver{ID: id}); err != nil {
		t.Fatal(err)
	}
	if e.cs.aim == nil {
		t.Skip("the maneuver took no aim")
	}
	if len(fighting(e).Allowed.Aimed) == 0 {
		t.Fatal("the aim is up but not published")
	}
	if _, err := e.Do(game.Shoot{Ats: []image.Point{f.Field.Units[i].At}}); err != nil {
		t.Fatal("its attack:", err)
	}
	if e.cs.aim != nil {
		t.Errorf("maneuver %d: the aim (%d) is still up", id, e.cs.aim.then.kind)
	}
}

// TestManeuversClassic: the original offers none, and Do refuses one.
func TestManeuversClassic(t *testing.T) {
	e := partyTurnScene(t)
	if p, a := maneuversOffered(fighting(e)); len(p)+len(a) != 0 {
		t.Fatalf("offered %v %v", p, a)
	}
	if _, err := e.Do(game.DoManeuver{ID: 1}); err == nil {
		t.Error("DoManeuver not refused")
	}
}

// aimAt is where maneuver id of f's turn aims: an ally one at the actor
// itself, the rest at foe.
func aimAt(f game.Fighting, id int, foe image.Point) image.Point {
	for _, m := range f.Field.Maneuvers {
		if m.ID == id && rules.ManeuverAim(m.Aim) == rules.AimAlly {
			for _, u := range f.Field.Units {
				if u.ID == f.Field.Turn {
					return u.At
				}
			}
		}
	}
	return foe
}
