package rulesxxvc

import (
	"image"
	"slices"
	"testing"

	"gobox/rules"
)

// board is a w x h board of terrain t with the given cells set.
func board(w, h int, t rules.TerrainID, set map[rules.BoardPos]rules.TerrainID) rules.Board {
	b := rules.Board{Size: image.Pt(w, h), Terrain: make([]rules.TerrainID, w*h)}
	for i := range b.Terrain {
		b.Terrain[i] = t
	}
	for p, v := range set {
		b.Terrain[p.Y*w+p.X] = v
	}
	return b
}

const (
	wallTerrain  rules.TerrainID = 8  // combatTerrain[8]: cost 255, LINE_STOPS
	roughTerrain rules.TerrainID = 46 // combatTerrain[46]: cost 2, LINE_BLOCKED
)

// WalkLine by hand (pathing.pas): (0,0) to (4,2) walks 2+1, 2, 2+1, 2 = 10;
// a wall on (2,1) stops it there after 5; rough ground past the two grace
// cells is seen as blocked; chaff stops it only with stopsAtChaff.
func TestWalkLine(t *testing.T) {
	if w, g := terrainRow(wallTerrain), terrainRow(roughTerrain); w.cost != impassable || w.line != lineStops || g.cost != 2 || g.line != lineBlocked {
		t.Fatal("table rows moved")
	}
	to := rules.BoardPos{X: 4, Y: 2}
	if end, d, ok, _ := walkLine(board(5, 3, openTerrain, nil), rules.BoardPos{}, to, 9, true, false); !ok || end != to || d != 10 {
		t.Fatal(end, d, ok)
	}
	wall := board(5, 3, openTerrain, map[rules.BoardPos]rules.TerrainID{{X: 2, Y: 1}: wallTerrain})
	if end, d, ok, _ := walkLine(wall, rules.BoardPos{}, to, 9, true, false); ok || end != (rules.BoardPos{X: 2, Y: 1}) || d != 5 {
		t.Fatal(end, d, ok)
	}
	rough := board(5, 3, openTerrain, map[rules.BoardPos]rules.TerrainID{{X: 3, Y: 2}: roughTerrain})
	if _, _, ok, seen := walkLine(rough, rules.BoardPos{}, to, 9, true, false); !ok || !seen.blocked {
		t.Fatal(ok, seen)
	}
	chaff := board(5, 3, openTerrain, map[rules.BoardPos]rules.TerrainID{{X: 1, Y: 1}: chaffTerrain})
	if _, _, ok, seen := walkLine(chaff, rules.BoardPos{}, to, 9, true, false); !ok || !seen.chaff {
		t.Fatal(ok, seen)
	}
	if _, _, ok, _ := walkLine(chaff, rules.BoardPos{}, to, 9, true, true); ok {
		t.Fatal("chaff did not stop the line")
	}
}

// CombatMoveCreature's cost: rough ground (cost 2) costs 4 straight, so
// movement 4 reaches it but not past; a wall is never entered.
func TestReachTerrain(t *testing.T) {
	r := New()
	f := rules.Fight{Board: board(3, 1, openTerrain, map[rules.BoardPos]rules.TerrainID{{X: 1, Y: 0}: roughTerrain})}
	if rs := r.reach(f, rules.NoChar, rules.BoardPos{}, 4); len(rs) != 1 || rs[0].At != (rules.BoardPos{X: 1}) {
		t.Fatalf("%+v", rs)
	}
	f.Board.Terrain[1] = wallTerrain
	if rs := r.reach(f, rules.NoChar, rules.BoardPos{}, 9); len(rs) != 0 {
		t.Fatalf("%+v", rs)
	}
}

// AiMoveStep walled in: no step three times (stuck 1, 2: the target dropped
// and found again, 3: the movement ends), then FinishRound guards.
func TestAiStuck(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5, BaseMovement: 6, Scale: 2, BaseScale: 2})
	b := board(3, 1, openTerrain, map[rules.BoardPos]rules.TerrainID{{X: 1}: wallTerrain})
	b.Units = []rules.BoardUnit{{ID: a, At: rules.BoardPos{X: 2}}, {ID: m}}
	f := rules.Fight{Party: rules.Party{a}, Board: b, Actor: m}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	r.fight[m].initiative = 5
	r.BeginTurn(f, m, &rng)
	s := r.fight[m]
	if act := r.Decide(f, m, &rng); act != (rules.Guard{}) || s.stuck != 3 || s.movement != 0 || s.target != a {
		t.Fatalf("%#v stuck %d movement %d target %d", act, s.stuck, s.movement, s.target)
	}
}

// GuardersAttack: a guarding monster beside the path attacks the walker
// once and stops guarding; the walk goes on (the walker survives).
func TestGuardersAttack(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	r.chars[a].HP, r.chars[a].MaxHP = 50, 50
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5, BaseScale: 2, Scale: 2, Unarmed: [2]rules.Dice{{N: 1, Sides: 1}}})
	b := board(5, 3, openTerrain, nil)
	b.Units = []rules.BoardUnit{{ID: a, At: rules.BoardPos{X: 0, Y: 1}}, {ID: m, At: rules.BoardPos{X: 2, Y: 0}}}
	f := rules.Fight{Party: rules.Party{a}, Board: b, Actor: a, Round: 1}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	r.fight[a].movement, r.fight[m].guarding = 20, true
	o, err := r.Resolve(f, rules.Move{To: rules.BoardPos{X: 3, Y: 1}}, &rng)
	if err != nil || r.fight[m].guarding || r.fight[a].movement != 14 {
		t.Fatalf("%v %+v guarding %v movement %d", err, o, r.fight[m].guarding, r.fight[a].movement)
	}
	attacked := 0
	for _, e := range o.Effects {
		if x, ok := e.(rules.Attacked); ok && x.From == unitOf(m) {
			attacked++
		}
	}
	if attacked == 0 {
		t.Fatalf("%+v", o.Effects)
	}
}

// A two-cell icon (size 3: cell and its east neighbour) steps east onto its own second cell, as ProbeDirection lets it
// (the unit's own cells do not block) and needs the square beyond: from x=0 it reaches x=1 and x=2 (the sweep's
// "Move ... not reachable" for a big monster).
func TestReachBigUnitOwnCells(t *testing.T) {
	r := New()
	id := r.Add(Character{Name: "big", Enabled: true, Status: StOkay})
	f := rules.Fight{Board: board(4, 1, openTerrain, nil)}
	f.Board.Units = []rules.BoardUnit{{ID: id, At: rules.BoardPos{}, Size: 3}}
	var got []int
	for _, x := range r.reach(f, id, rules.BoardPos{}, 99) {
		got = append(got, x.At.X)
	}
	if !slices.Equal(got, []int{1, 2}) {
		t.Errorf("reach %v, want 1 2 (x=3 would put the second cell off the board)", got)
	}
}
