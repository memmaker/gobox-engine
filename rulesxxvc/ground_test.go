package rulesxxvc

import (
	"slices"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// groundFight is a 12x10 board of open ground with a grenadier m at (0,5)
// throwing projectile kind, party members a at (5,5) and b at (6,5).
func groundFight(t *testing.T, kind int32, set map[rules.BoardPos]rules.TerrainID) (*Rules, rules.Fight, rules.CharID, rules.CharID, rules.CharID) {
	t.Helper()
	r := New()
	r.tables = xxvc.Tables{
		ItemTypes: []xxvc.ItemType{{}, {Slot: xxvc.SlotWeapon, WeaponClass: xxvc.Grenade, Damage: rules.Dice{N: 1, Sides: 6}, Range: 5, Consumable: true}},
		Items:     map[rules.ItemID]xxvc.ItemTemplate{1: {Type: 1, Special: [3]int32{0, kind, 1}}},
	}
	m := r.Add(Character{Name: "m", Side: rules.SideMonsters, Enabled: true, HP: 20, MaxHP: 20, BaseCombat: 100, AttackBonus: 100,
		Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1, Count: 2}, Readied: true}}})
	hurt := Character{Enabled: true, HP: 20, MaxHP: 20, Saves: [8]int{30, 30, 30, 30, 30, 30, 30, 30}}
	a, b := r.Add(hurt), r.Add(hurt)
	f := rules.Fight{Party: rules.Party{a, b}, Actor: m, Board: board(12, 10, openTerrain, set)}
	f.Board.Units = []rules.BoardUnit{{ID: m, At: rules.BoardPos{Y: 5}}, {ID: a, At: rules.BoardPos{X: 5, Y: 5}}, {ID: b, At: rules.BoardPos{X: 6, Y: 5}}}
	r.BeginFight(f, nil)
	scratch := rules.RNG{State: 9}
	r.BeginRound(f, &scratch)
	return r, f, m, a, b
}

// AttackGroundTarget by hand (cbtacts.pas CombatLaunchProjectile): an
// explosive grenade (kind 6, radius 1, a save halves) from a thrower whose
// scatter roll (100 + d20) is over 50 lands on a at (5,5); the draws are the
// damage die, the scatter d20, the rubble rolls of the one cell whose
// blastResult is 250 (the wall north of a: rubble, maybe more on a's cell), the burst's particles, then a save each for a
// (distance 0) and b (2) in that order; the grenade spends one.
func TestGroundAttack(t *testing.T) {
	var rubble rules.TerrainID
	for i, row := range combatTerrain {
		if row.blast == 250 {
			rubble = rules.TerrainID(i + 1)
			break
		}
	}
	at, wall := rules.BoardPos{X: 5, Y: 5}, rules.BoardPos{X: 5, Y: 4}
	for seed := uint32(1); seed < 40; seed++ {
		r, f, m, a, b := groundFight(t, 6, map[rules.BoardPos]rules.TerrainID{wall: rubble})
		rng, x := rules.RNG{State: seed}, rules.RNG{State: seed}
		o, err := r.Resolve(f, rules.Attack{At: &at}, &rng)
		if err != nil {
			t.Fatal(err)
		}
		dmg := rules.RNGDice(&x, 1, 6)
		rules.RNGDice(&x, 1, 20)
		want, below := rules.TerrainID(152), rules.TerrainID(0)
		if PercentileRoll(&x, 25) {
			want, below = 146, 153
			if PercentileRoll(&x, 25) {
				below = 147
			}
		}
		animateDraws(6, 1, at, &x)
		for _, id := range []rules.CharID{a, b} {
			_, made, _ := SaveDie(&x, 0)
			hp := 20 - dmg
			if made {
				hp = 20 - dmg/2
			}
			if r.chars[id].HP != hp {
				t.Fatalf("seed %d: %d has %d HP, want %d", seed, id, r.chars[id].HP, hp)
			}
		}
		if rng.State != x.State {
			t.Fatalf("seed %d: the draws differ from the original's order", seed)
		}
		got := r.board(f.Board)
		if terrainAt(got, wall) != want || below != 0 && terrainAt(got, at) != below {
			t.Fatalf("seed %d: rubble %d below %d, want %d %d", seed, terrainAt(got, wall), terrainAt(got, at), want, below)
		}
		if !slices.Contains(o.Effects, rules.Effect(rules.TerrainSet{At: wall, Terrain: want})) {
			t.Fatalf("seed %d: no TerrainSet for the rubble: %v", seed, o.Effects)
		}
		if c := r.chars[m]; c.Items[0].Ref.Count != 1 || r.fight[m].initiative != 0 {
			t.Fatalf("seed %d: grenades %d, initiative %d", seed, c.Items[0].Ref.Count, r.fight[m].initiative)
		}
	}
}

// The leaper: an enabled UNKILLABLE fighter in the blast walks (Felix: no
// jump) to the free cell outside it whose line from it is shortest, and is
// not hit.
func TestLeapsForCover(t *testing.T) {
	r, f, _, a, _ := groundFight(t, 6, nil)
	r.chars[a].Effects = []Effect{{ID: EffUnkillable}}
	at := rules.BoardPos{X: 5, Y: 5}
	rng := rules.RNG{State: 3}
	o, err := r.Resolve(f, rules.Attack{At: &at}, &rng)
	if err != nil {
		t.Fatal(err)
	}
	i := slices.IndexFunc(o.Effects, func(e rules.Effect) bool { w, ok := e.(rules.Walked); return ok && w.Unit == unitOf(a) })
	if i < 0 || r.chars[a].HP != 20 {
		t.Fatalf("no leap: %v, HP %d", o.Effects, r.chars[a].HP)
	}
	path := o.Effects[i].(rules.Walked).Path
	if end := path[len(path)-1]; distance(end, at) < 2 || len(path) != 2 {
		t.Fatalf("leapt to %v along %v", end, path)
	}
}

// A gas grenade (kind 7) lays its marker on the 3x3 block for 4 rounds; the
// minutes are its damage roll; then the air clears back to open ground.
func TestGasCloud(t *testing.T) {
	r, f, _, a, _ := groundFight(t, 7, nil)
	at := rules.BoardPos{X: 5, Y: 5}
	rng, x := rules.RNG{State: 5}, rules.RNG{State: 5}
	o, err := r.Resolve(f, rules.Attack{At: &at}, &rng)
	if err != nil {
		t.Fatal(err)
	}
	minutes := rules.RNGDice(&x, 1, 6)
	rules.RNGDice(&x, 1, 20)
	animateDraws(7, 1, at, &x)
	if _, made, _ := SaveDie(&x, 0); !made && (len(r.chars[a].Effects) != 1 || r.chars[a].Effects[0] != Effect{ID: EffHelpless, Minutes: minutes}) {
		t.Fatalf("a: %v, want helpless %d", r.chars[a].Effects, minutes)
	}
	sets := 0
	for _, e := range o.Effects {
		if e, ok := e.(rules.TerrainSet); ok && e.Terrain == gasTerrain {
			sets++
		}
	}
	if sets != 9 {
		t.Fatalf("%d gas cells", sets)
	}
	for round := 1; round <= 4; round++ {
		end := r.EndRound(f, &rng)
		if gas := terrainAt(r.board(f.Board), at) == gasTerrain; gas != (round < 4) || (len(end.Effects) > 0) != (round == 4) {
			t.Fatalf("round %d: gas %v, %v", round, gas, end.Effects)
		}
	}
}

// ScatterOffset by hand: d4 div 2 further out, a d2 way at 0 and past 8.
func TestScatterOffset(t *testing.T) {
	for seed := uint32(1); seed < 20; seed++ {
		for _, v := range []int{-3, 0, 4, 9} {
			rng, x := rules.RNG{State: seed}, rules.RNG{State: seed}
			amount, way := rules.RNGDice(&x, 1, 4)>>1, rules.RNGDice(&x, 1, 2)*2-3
			want := map[int]int{-3: -3 - amount, 0: way, 4: 4 + amount, 9: 9 + amount*way}[v]
			if got := scatterOffset(v, &rng); got != want || rng.State != x.State {
				t.Fatalf("seed %d: %d to %d, want %d", seed, v, got, want)
			}
		}
	}
}

// ConsiderTerrainEffects by hand: on gas, save 1 or helpless for 1d6+1 and
// the actions end; breathing protection or open ground draws nothing.
func TestConsiderTerrain(t *testing.T) {
	r := New()
	at := rules.BoardPos{X: 1, Y: 1}
	b := board(3, 3, openTerrain, map[rules.BoardPos]rules.TerrainID{at: gasTerrain})
	for seed := uint32(1); seed < 20; seed++ {
		id := r.Add(Character{Enabled: true, Saves: [8]int{30, 30, 30, 30, 30, 30, 30, 30}})
		b.Units = []rules.BoardUnit{{ID: id, At: at}}
		c, s := r.chars[id], &combatState{initiative: 5}
		rng, x := rules.RNG{State: seed}, rules.RNG{State: seed}
		r.considerTerrain(b, id, c, s, &rng)
		if _, made, _ := SaveDie(&x, 0); !made {
			if want := rules.RNGDice(&x, 1, 6) + 1; !helpless(c) || c.Effects[0].Minutes != want || s.initiative != 0 {
				t.Fatalf("seed %d: %v %+v", seed, c.Effects, s)
			}
		}
		if rng.State != x.State {
			t.Fatalf("seed %d: draws", seed)
		}
		c.Effects, c.Armor.Breathing = nil, true
		if r.considerTerrain(b, id, c, s, &rng); rng.State != x.State || helpless(c) {
			t.Fatal("breathing protection drew a save")
		}
	}
}

// ManualAimCursor with a loaded area weapon accepts ground cells a WalkLine
// reaches within range (distance div 2); an enemy's cell stays its Attack.
// A downed fighter covers no cell (CreatureDowned: size 0).
func TestGroundAims(t *testing.T) {
	r, f, m, a, _ := groundFight(t, 6, nil)
	ground := map[rules.BoardPos]bool{}
	for _, x := range r.allowedFight(f).Aimed {
		if at, ok := x.Action.(rules.Attack); ok && at.At != nil {
			ground[*at.At] = true
		}
	}
	if !ground[rules.BoardPos{X: 4, Y: 5}] || ground[rules.BoardPos{X: 7, Y: 5}] || ground[rules.BoardPos{X: 5, Y: 5}] {
		t.Errorf("ground aims %v", ground)
	}
	if _, ok := r.occupant(f.Board, rules.BoardPos{X: 5, Y: 5}); !ok {
		t.Fatal("a not on (5,5)")
	}
	r.chars[a].Enabled = false
	if _, ok := r.occupant(f.Board, rules.BoardPos{X: 5, Y: 5}); ok {
		t.Error("a downed still covers (5,5)")
	}
	r.chars[m].Status = StSprinting
	for _, x := range r.allowedFight(f).Aimed {
		if at, ok := x.Action.(rules.Attack); ok && at.At != nil {
			t.Fatal("ground aim while sprinting")
		}
	}
}

// TestBurstOffset pins AnimateProjectileArea's offsets on the fixed 50x25
// arena (cbtacts.pas: SizeOf(cells[0]) = 50 columns, 25 rows), values
// worked by hand from the Pascal.
func TestBurstOffset(t *testing.T) {
	for _, c := range []struct{ v, size, radius, want int }{
		{0, arenaCols, 1, -3}, {0, arenaCols, 3, -9}, {2, arenaCols, 3, -3}, {3, arenaCols, 3, 0},
		{25, arenaCols, 3, 0}, {46, arenaCols, 3, 0}, {47, arenaCols, 3, 3}, {49, arenaCols, 3, 9}, {49, arenaCols, 1, 3},
		{21, arenaRows, 3, 0}, {22, arenaRows, 3, 3}, {24, arenaRows, 3, 9}, {24, arenaRows, 1, 3}, {23, arenaRows, 1, 0},
	} {
		if got := burstOffset(c.v, c.size, c.radius); got != c.want {
			t.Errorf("burstOffset(%d, %d, %d) = %d, want %d", c.v, c.size, c.radius, got, c.want)
		}
	}
	if arenaCols != 50 || arenaRows != 25 {
		t.Fatalf("arena %dx%d, want TArena's 50x25", arenaCols, arenaRows)
	}
}

// A cloud's or corpse's cell publishes its marker (the terrain value, cbtmap.pas DrawArenaCell draws rows 33..35 over it)
// with the ground beneath as its tile; plain ground has none.
func TestGroundMarks(t *testing.T) {
	r := New()
	r.ground = board(arenaCols, arenaRows, openTerrain, nil)
	r.arena.cells = map[rules.BoardPos]rules.TerrainID{{X: 2, Y: 3}: chaffTerrain, {X: 4, Y: 1}: corpseTerrain, {X: 6, Y: 6}: gasTerrain}
	r.arena.corpses = []corpse{{at: rules.BoardPos{X: 4, Y: 1}, terrain: 21}}
	g := r.groundFact().Squares
	for _, c := range []struct {
		x, y int
		mark rules.SquareMark
	}{{2, 3, rules.SquareMark(chaffTerrain)}, {4, 1, rules.SquareMark(corpseTerrain)}, {6, 6, rules.SquareMark(gasTerrain)}, {0, 0, 0}} {
		if g[c.y][c.x].Mark != c.mark {
			t.Errorf("(%d,%d): mark %d, want %d", c.x, c.y, g[c.y][c.x].Mark, c.mark)
		}
	}
	if g[1][4].Tile != 21 {
		t.Errorf("the corpse's ground is tile %d, want the terrain beneath (21)", g[1][4].Tile)
	}
}
