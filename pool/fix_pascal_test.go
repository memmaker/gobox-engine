package pool

import (
	"bufio"
	"fmt"
	"os"
	"rules1e"
	"slices"
	"strings"
	"testing"
)

// TestFixesAgainstPascal compares ApplyDamage with CharHlp.InflictDamage out of
// combat, and setSpeed with CbtActs.RecalcAttacksLeft for a melee creature
// (rules1e testdata/por-fix.txt, por-re vecfix): the fixed hit on a body and
// attacks left that never grow once the turn has attacked. Also the spell lost by
// the one hit (Damage), spitAcid with no target, charmGaze's 12 squares and
// restored curSpell, wildRoad's pair at the last column, isHardier's hit points, Constitution's save bonus,
// an item spell's magic resistance (ruling 15) and Hold Person's save vs spell (ruling 16).
func TestFixesAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-fix.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := quickStarted(t)
	c := e.state().Party[0]
	var m *Engine // magres's engine, started once
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		fs := strings.Fields(key)
		k := atois(t, fs[1:])
		var got string
		switch fs[0] {
		case "damage":
			e.state().BeginFlag = 0
			c.HP, c.Status, c.Okay, c.Affects = 4, byte(k[0]), 1, nil
			e.state().ApplyDamage(c, k[1])
			got = fmt.Sprintf("%d %d", c.Status, c.HP)
		case "spelllost":
			got = spellLostCase(t, k[0])
		case "acidnil":
			a := quickStarted(t)
			m := fixChar(8, 1)
			a.state().RandSeed = uint32(int32(k[0]))
			a.state().SpitAcid(m, &rules1e.Affect{}, 0)
			got = fmt.Sprint(int32(a.state().RandSeed))
		case "charm":
			got = charmCase(t, k[0], k[1])
		case "hardier":
			h := quickStarted(t)
			m := fixChar(30, 0)
			m.MaxHP, m.HP, m.HPRoll = 30, 25, 20
			m.Stats[rules1e.StatCon] = rules1e.Score(byte(k[0]))
			m.Levels[rules1e.ClsCleric], m.Levels[rules1e.ClsFighter] = 2, rules1e.Level(k[1])
			h.state().IsHardier(m, &rules1e.Affect{}, 0)
			got = fmt.Sprintf("%d %d %d %d", m.MaxHP, m.HP, m.Stats[rules1e.StatCon], rules1e.B2i(len(m.Affects) > 0))
		case "drain":
			z := quickStarted(t)
			m := fixChar(30, 0)
			m.Levels = [len(m.Levels)]rules1e.Level{}
			m.Levels[rules1e.ClsCleric], m.MaxLevel, m.Stats[rules1e.StatWis] = 3, 3, rules1e.Score(k[0])
			z.state().RecomputeCharacter(m)
			m.Memorized = [len(m.Memorized)]rules1e.SpellID{}
			for i := range k[1] {
				m.Memorized[i] = rules1e.SpBless
			}
			z.state().DrainLevels(m, 1)
			held := 0
			for _, v := range m.Memorized {
				held += rules1e.B2i(v != 0)
			}
			got = fmt.Sprintf("%d %d %d", m.Levels[rules1e.ClsCleric], m.SpellSlots[rules1e.SpellClassCleric][0], held)
		case "ringsave":
			got = ringSaveCase(t, k[0], k[1], k[2])
		case "robring":
			z := quickStarted(t)
			m := fixChar(8, 0)
			ring := &rules1e.Item{Type: 0x5A, Readied: 1}
			ring.Spells[1], ring.Spells[2] = rules1e.SpellID(rules1e.AffFeatherFall), 130 // effect 27, handler 130 (grant the item's effect)
			m.Items = []*rules1e.Item{ring}
			z.state().AddAffect(m, rules1e.AffFeatherFall, 0, 12, 0)
			z.state().RandSeed = uint32(int32(k[1]))
			z.state().ItemSave(m, k[0])
			got = fmt.Sprintf("%d %d", rules1e.B2i(z.state().FindAffect(m, rules1e.AffFeatherFall) != nil), rules1e.B2i(len(m.Items) > 0))
		case "efftable":
			z := quickStarted(t)
			a, b := &rules1e.Character{}, &rules1e.Character{}
			a.Affects = []*rules1e.Affect{{AffectRecord: rules1e.AffectRecord{Type: 5}}}
			b.Affects = []*rules1e.Affect{{AffectRecord: rules1e.AffectRecord{Type: 9}}, {AffectRecord: rules1e.AffectRecord{Type: rules1e.AffFeatherFall}}}
			z.state().Party = []*rules1e.Character{a, b}
			v, ok := z.state().EffectTable(k[0])
			got = fmt.Sprintf("%d %d", rules1e.B2i(ok), v)
		case "z1write":
			z := quickStarted(t)
			m := z.member(z.selected())
			m.Memorized = [len(m.Memorized)]rules1e.SpellID{}
			if r, ok := refAt(t, z, gameVarBase+k[0]); ok { // no field there: nothing changes
				z.host.Set(r, 9)
			}
			got = fmt.Sprintf("%d %d %d", m.Memorized[0], m.Memorized[1], m.Memorized[2])
		case "consave":
			c.Stats[rules1e.StatCon] = rules1e.Score(byte(k[0]))
			e.state().Rv.SaveType, e.state().Rv.Save = byte(k[1]), 10
			e.state().SaveBonusVsType0(c, &rules1e.Affect{}, 0)
			e.state().SaveBonusVsType42(c, &rules1e.Affect{}, 0)
			got = fmt.Sprint(e.state().Rv.Save)
		case "magres":
			if m == nil {
				m = quickStarted(t)
			}
			m.member(m.selected()).Levels[rules1e.ClsMage] = rules1e.Level(byte(k[0]))
			m.state().CurSpell, m.state().UsingItem = 15, k[1] != 0
			m.state().Rv.AffVal, m.state().Rv.DmgType = 5, 0
			m.state().RandSeed = uint32(int32(k[2]))
			m.as(m.sel)
			m.state().Resist50(m.member(m.selected()), &rules1e.Affect{}, 0)
			got = fmt.Sprintf("%d %d", m.state().Rv.AffVal, int32(m.state().RandSeed))
		case "hold":
			got = holdCase(t, k[0], k[1], k[2])
		case "river":
			got = riverCase(t, k[0])
		case "attacks":
			c.Affects, c.Slots[rules1e.SlotWeapon] = nil, nil
			c.Combat = &rules1e.CombatInfo{Attacked: byte(k[3])}
			e.state().CombatRound = byte(k[0])
			c.BaseAttack[0], c.Attacks[0] = rules1e.Count(k[1]), rules1e.Count(k[2])
			e.state().SetSpeed(c)
			got = fmt.Sprint(c.Attacks[0])
		}
		if got != val {
			t.Errorf("%s: Go %s", line, got)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// spellLostCase: Damage in combat on t casting spell 5, the attacker (the selected member) holding it in the same
// slot; vecfix spelllost.
func spellLostCase(t *testing.T, slot int) string {
	e := quickStarted(t)
	a := e.state().Party[0]
	d := fixChar(40, 1)
	d.Combat.Spell = 5
	a.Memorized, d.Memorized = [21]rules1e.SpellID{}, [21]rules1e.SpellID{}
	a.Memorized[slot], d.Memorized[slot] = 5, 5
	fixFloor(e)
	fixPlace(e, a, 6, 4)
	fixPlace(e, d, 6, 6)
	e.state().BeginFlag = 5
	e.state().RandSeed = 1
	e.state().Damage(d, 3, 0, false)
	return fmt.Sprintf("%d %d %d", d.Memorized[slot], a.Memorized[slot], d.Combat.Spell)
}

// charmCase: a gazer at (6, 4) charms a target dist squares down the column; vecfix charm.
func charmCase(t *testing.T, dist, seed int) string {
	e := quickStarted(t)
	fixFloor(e)
	c := fixChar(8, 1)
	o := fixChar(8, 0)
	o.Saves[rules1e.SaveVsSpell] = 255
	c.Combat.Target = o
	fixPlace(e, c, 6, 4)
	fixPlace(e, o, 6, 4+dist)
	e.state().BeginFlag = 5
	e.state().CurSpell = 7
	e.state().RandSeed = uint32(int32(seed))
	e.state().CharmGaze(c, &rules1e.Affect{}, 0)
	return fmt.Sprintf("%d %d", rules1e.B2i(e.state().FindAffect(o, rules1e.AffCharmed) != nil), e.state().CurSpell)
}

// riverCase: wildRoad on floor at terrain 181, seeded; columns 48, 49 and 0 of every row; vecfix river.
func riverCase(t *testing.T, seed int) string {
	e := quickStarted(t)
	fixFloor(e)
	e.state().Terrain = 181
	e.state().RandSeed = uint32(int32(seed))
	e.r1e().WildRoad()
	var b strings.Builder
	for row := range e.state().Cmap.Ground {
		g := e.state().Cmap.Ground[row]
		fmt.Fprintf(&b, " %d.%d.%d", g[48], g[49], g[0])
	}
	return strings.TrimPrefix(b.String(), " ")
}

// fixChar is a creature of hp on side, okay, with a combat record, as TFix.AddCreature.
func fixChar(hp, side byte) *rules1e.Character {
	c := &rules1e.Character{Combat: &rules1e.CombatInfo{}}
	c.HP, c.Okay, c.Side = byte(hp), 1, side
	return c
}

// fixFloor gives e a combat map of plain floor, as TFix.FillArenaFloor.
func fixFloor(e *Engine) {
	e.state().Cmap = &rules1e.CombatMap{}
	for y := range e.state().Cmap.Ground {
		for x := range e.state().Cmap.Ground[y] {
			e.state().Cmap.Ground[y][x] = rules1e.GroundFloor
		}
	}
}

// fixPlace registers c as the next combatant at (x, y), as TFix.PlaceOnArena.
func fixPlace(e *Engine, c *rules1e.Character, x, y int) {
	e.state().CombatantSlots = append(e.state().CombatantSlots, c)
	e.state().Ncomb = len(e.state().CombatantSlots)
	e.state().Cpos[e.state().Ncomb] = rules1e.CombatPos{X: byte(x), Y: byte(y), Size: 1}
	e.state().Occ[y][x] = rules1e.Slot(byte(e.state().Ncomb))
}

// holdCase casts a cleric's Hold Person (level 5) at one target whose
// paralysis and spell saves are given: held, RandSeed.
func holdCase(t *testing.T, para, spell, seed int) string {
	e := quickStarted(t)
	sel := e.state().Acting()
	sel.Levels[rules1e.ClsCleric] = 5
	tg := fixChar(10, 1)
	tg.Saves[rules1e.SaveVsParalysis], tg.Saves[rules1e.SaveVsSpell] = rules1e.Score(para), rules1e.Score(spell)
	e.state().CurSpell, e.state().NSpellTargets = rules1e.SpHoldPersonCl, 1
	e.state().SetTgt(1, tg)
	e.state().RandSeed = uint32(int32(seed))
	e.state().SpellHold()
	return fmt.Sprintf("%d %d", rules1e.B2i(e.state().FindAffect(tg, rules1e.AffHeld) != nil), int32(e.state().RandSeed))
}

// ringSaveCase readies a Ring of Protection, a Cloak of Displacement (save
// bonus their plus, as fixItemRecord leaves them) and magical body armour of
// the given pluses (0: none) on bases as vecfix's: saveBonus, AC.
func ringSaveCase(t *testing.T, ring, cloak, armour int) string {
	e := quickStarted(t)
	const bodyArmour = 0x32
	e.state().Items = slices.Clone(e.state().Items)
	e.state().Items[rules1e.ItemProtectionRing] = rules1e.ItemBase{Slot: 9, Protection: rules1e.AcGiven}
	e.state().Items[rules1e.ItemCloakDisplacement] = rules1e.ItemBase{Slot: 4, Protection: rules1e.AcGiven}
	e.state().Items[bodyArmour] = rules1e.ItemBase{Slot: rules1e.SlotArmor, Protection: rules1e.AcGiven + 5}
	c := fixChar(8, 0)
	for _, w := range []struct {
		typ  rules1e.ItemType
		plus int
		save bool
	}{{rules1e.ItemProtectionRing, ring, true}, {rules1e.ItemCloakDisplacement, cloak, true}, {bodyArmour, armour, false}} {
		if w.plus > 0 {
			it := &rules1e.Item{Type: w.typ, Readied: 1, Plus: byte(w.plus)}
			if w.save {
				it.Plus2 = byte(w.plus)
			}
			c.Items = append(c.Items, it)
		}
	}
	e.state().ApplyItems(c)
	return fmt.Sprintf("%d %d", c.SaveBonus, c.AC[0])
}
