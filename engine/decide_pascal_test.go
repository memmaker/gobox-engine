package engine

import (
	"bufio"
	"fmt"
	"os"
	"gobox/rules1e"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// The vectors come from the pure Pascal (por-re vecai): one whole computer turn on open
// floor. The turn still lives here; a difference is a bug, or a row of CHANGES-FROM-ORIGINAL.md.
// aiVaried is what the morale, flight and wall turns set beyond a plain turn.
type aiVaried struct {
	control, aHP, aInt, dMove, turned, startY, wall byte
	memo, pending, itemEff, casterLvl, undead       byte
	quick                                           byte // the caster on quick fight (Auto)
	bow                                             byte // a readied bow: 1 plain, 2 cursed (vecai BowTurn)
	breath                                          *rules1e.Affect
	outStatus                                       byte // set: the defender is taken out of play instead // set: the breath runs in place of the computer turn
}

func aiTurnVector(e *Engine, v aiVaried, seed, dist, hp, lvl, flank int) string {
	const west = 6
	d, a := &rules1e.Character{}, &rules1e.Character{}
	d.Side, a.Side = rules1e.SideParty, rules1e.SideMonsters
	d.Okay, a.Okay = 1, 1
	d.HP, d.MaxHP = byte(hp), byte(hp)
	d.BaseAC, d.AC[0], d.BaseMove, d.Move = 50, 50, v.dMove, v.dMove
	d.Stats[rules1e.StatDex] = 10
	a.Stats[rules1e.StatStr], a.Stats[rules1e.StatDex] = 10, 10
	a.Levels[rules1e.ClsThief] = rules1e.Level(byte(lvl))
	a.BaseAttack[0], a.Attacks[0] = 4, 2
	a.Thac0Base, a.THAC0 = 50, 50
	a.BaseDice[0], a.BaseSides[0], a.BaseBonus[0] = 1, 6, 1
	a.Dice[0], a.Sides[0], a.Bonus[0] = 1, 6, 1
	a.HP, a.MaxHP, a.BaseMove, a.Move = v.aHP, 40, 12, 12
	a.Control = v.control
	a.Auto = v.quick
	d.Undead = v.undead
	e.state().Rv.Side = [2]int8{1, 1}
	a.Memorized[0] = rules1e.SpellOfByte(v.memo)
	a.Levels[rules1e.ClsCleric], a.Levels[rules1e.ClsMage] = rules1e.Level(v.casterLvl), rules1e.Level(v.casterLvl)
	if v.itemEff != 0 {
		it := &rules1e.Item{Readied: 1, Type: rules1e.ItemType(len(e.state().Items))} // the Pascal's base item 0: all zero
		it.Spells[1] = rules1e.ItemPower(v.itemEff)
		a.Items = []*rules1e.Item{it}
	}
	if v.bow != 0 {
		it := &rules1e.Item{Readied: 1, Type: rules1e.ItemType(len(e.state().Items) - 1), Cursed: v.bow - 1} // the bow appended to e.items
		a.Items = []*rules1e.Item{it}
		a.Slots[rules1e.SlotWeapon] = it
	}
	a.Stats[rules1e.StatInt] = rules1e.Score(v.aInt)
	d.Combat = &rules1e.CombatInfo{Attackers: byte(flank), Facing: west}
	a.Combat = &rules1e.CombatInfo{Delay: 4, Moves: 24, AttackSlot: 2, CanCast: 1, CanUse: 1, Turned: v.turned, Spell: rules1e.SpellID(v.pending)}
	e.state().Party, e.state().Who = []*rules1e.Character{d, a}, 1
	e.state().CombatantSlots, e.state().Ncomb = []*rules1e.Character{d, a}, 2
	e.state().Cmap = &rules1e.CombatMap{}
	for y := range e.state().Cmap.Ground {
		for x := range e.state().Cmap.Ground[y] {
			e.state().Cmap.Ground[y][x] = rules1e.PorGroundIDs.Floor
		}
	}
	if v.wall != 0 {
		for y := 0; y < 25; y++ {
			e.state().Cmap.Ground[y][12] = rules1e.Ground(v.wall)
		}
	}
	clear(e.state().Occ[:])
	e.state().Cpos[1] = rules1e.CombatPos{X: 10, Y: v.startY, Size: 1}
	e.state().Cpos[2] = rules1e.CombatPos{X: byte(10 + dist), Y: v.startY, Size: 1}
	e.state().Occ[v.startY][10], e.state().Occ[v.startY][10+dist] = 1, 2
	e.state().BeginFlag = 5
	e.state().Game.Morale, e.facing = 50, 0 // TFix: threshold 50, StartCombat(NORTH)
	e.state().RandSeed = uint32(int32(seed))
	if v.breath != nil {
		e.state().BreathWeapon(a, v.breath, 0)
	} else if v.outStatus != 0 {
		e.state().KillCharacter(d, v.outStatus, "")
	} else {
		e.state().ComputerTurn(a, rules1e.AiKeyNone)
	}
	return fmt.Sprintf("%d %d %d %d %d %d %d %d", d.HP, d.Status, e.state().Cpos[2].X, e.state().Cpos[2].Y, a.Attacks[0], int32(e.state().RandSeed), a.Status, a.Combat.Fleeing)
}

func TestDecideAgainstPascal(t *testing.T) {
	e := testEngine(t)
	e.Init()
	f, err := os.Open("../rules1e/testdata/por-ai.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cases, diffs := 0, 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "#") {
			continue
		}
		key, want, _ := strings.Cut(sc.Text(), " =")
		fs := strings.Fields(key)[1:]
		k := make([]int, len(fs))
		for i, s := range fs {
			if k[i], err = strconv.Atoi(s); err != nil {
				t.Fatal(err)
			}
		}
		cases++
		plain := aiVaried{aHP: 10, aInt: 10, dMove: 12, startY: 10}
		var got string
		switch strings.Fields(key)[0] {
		case "aiturn":
			got = aiTurnVector(e, plain, k[0], k[1], k[2], k[3], k[4])
		case "aimorale":
			seed, m, hp, intel, move, tu, pct, dist, y := k[0], k[1], k[2], k[3], k[4], k[5], k[6], k[7], k[8]
			e.state().FoesHealth = byte(pct)
			got = aiTurnVector(e, aiVaried{control: 0x80 + byte(m), aHP: byte(hp), aInt: byte(intel), dMove: byte(move),
				turned: byte(tu), startY: byte(y)}, seed, dist, 40, 0, 0)
		case "aicast":
			seed, sp, dist, pend, item := k[0], byte(k[1]), k[2], k[3], k[4]
			if rules1e.SpellIDs[sp] == "" { // a removed spell (56 Restoration, Felix 2026-11-12): no caster holds it
				cases--
				continue
			}
			v := aiVaried{control: 0x80 + 50, aHP: 40, aInt: 10, dMove: 12, startY: 10, casterLvl: 5}
			switch {
			case item != 0:
				v.itemEff = sp
				v.quick = byte(item / 2)
			case pend != 0:
				v.pending = sp
			default:
				v.memo = sp
			}
			e.state().FoesHealth = 100
			rest := aiTurnVector(e, v, seed, dist, 40, 0, 0)
			a := e.state().CombatantSlots[1]
			got = fmt.Sprintf("%d %d %d %s", a.Combat.Spell, a.Combat.Delay, a.Memorized[0].Byte(), rest)
		case "aibreath":
			b := &rules1e.Affect{AffectRecord: rules1e.AffectRecord{Type: rules1e.AffBreathWeapon, Data: byte(k[3])}}
			e.state().CombatRound = byte(k[2])
			rest := aiTurnVector(e, aiVaried{aHP: 40, aInt: 10, dMove: 12, startY: 10, breath: b}, k[0], k[1], 40, 0, 0)
			e.state().CombatRound = 0
			got = fmt.Sprintf("%d %d %d %s", e.state().TgtX, e.state().TgtY, b.Data, rest)
		case "aiout":
			v := aiVaried{aHP: 10, aInt: 10, dMove: 12, startY: 10, outStatus: byte(k[1])}
			aiTurnVector(e, v, 1, k[0], 40, 0, 0)
			d := e.state().CombatantSlots[0]
			got = fmt.Sprintf("%d %d %d %d", e.state().Occ[10][10], d.HP, d.Status, d.Okay)
		case "aihigh":
			v := aiVaried{control: 0x80 + 50, aHP: 40, aInt: 10, dMove: 12, startY: 10, casterLvl: 5, memo: 0x80 + byte(k[1])}
			e.state().FoesHealth = 100
			rest := aiTurnVector(e, v, k[0], k[2], 40, 0, 0)
			a := e.state().CombatantSlots[1]
			got = fmt.Sprintf("%d %d %d %s", a.Combat.Spell, a.Combat.Delay, a.Memorized[0].Byte(), rest)
		case "aiundead":
			v := aiVaried{control: 0x80 + 50, aHP: 40, aInt: 10, dMove: 12, startY: 10, casterLvl: byte(k[1]), undead: byte(k[2])}
			e.state().FoesHealth = 100
			rest := aiTurnVector(e, v, k[0], 1, 40, 0, 0)
			got = fmt.Sprintf("%d %d %d %s", e.state().Rv.Side[0], e.state().Rv.Side[1], e.state().CombatantSlots[0].Combat.Turned, rest)
		case "aisilence":
			c := &rules1e.Character{Combat: &rules1e.CombatInfo{CanUse: byte(k[0]), CanCast: byte(k[1])}}
			if k[2] == 1 {
				// ruled (Felix): coughing acts only in a fight (PoD's rule); the
				// Pascal vector's use/cast outcome is checked there
				e.state().BeginFlag = 5
				e.state().CoughingEnds(c, nil, 0)
			} else {
				e.state().IsSilenced(c, nil, 0)
			}
			got = fmt.Sprintf("%d %d", c.Combat.CanUse, c.Combat.CanCast)
		case "aialigned":
			c := &rules1e.Character{}
			c.HP, c.MaxHP, c.Okay = 40, 40, 1
			it := &rules1e.Item{Readied: 1}
			it.Spells[1] = 0x81
			e.state().BeginFlag = byte(k[0])
			c.Combat = &rules1e.CombatInfo{}
			e.state().CombatantSlots, e.state().Ncomb = []*rules1e.Character{c}, 1
			e.state().CursedAlignment(c, &rules1e.Affect{Item: it}, 0)
			got = fmt.Sprintf("%d %d", c.HP, it.Readied)
		case "aibow":
			items := e.state().Items
			e.state().Items = append(slices.Clone(items), rules1e.ItemBase{Range: 5, Usage: 9}) // vecai BOW_BASE: range 5, fires arrows
			rest := aiTurnVector(e, aiVaried{aHP: 10, aInt: 10, dMove: 12, startY: 10, bow: 1 + byte(k[1])}, k[0], 1, 40, 0, 0)
			e.state().Items = items
			got = fmt.Sprintf("%d %s", rules1e.B2i(e.state().CombatantSlots[1].Slots[rules1e.SlotWeapon] != nil), rest)
		case "aiwall":
			v := plain
			v.wall = byte(k[2])
			got = aiTurnVector(e, v, k[0], k[1], 40, 0, 0)
		}
		if w := strings.Join(strings.Fields(strings.ReplaceAll(want, " = ", " ")), " "); got != w {
			diffs++
			t.Errorf("%s: pascal %s, ours %s", key, w, got)
		}
	}
	t.Logf("%d cases, %d differences", cases, diffs)
}
