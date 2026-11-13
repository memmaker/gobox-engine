package rules1e

import (
	"gobox/core"
	"gobox/core/dnd"
	"gobox/core/goldbox"
)

// Who is what ClassStats reads: levels, race, abilities, and the slots the
// character holds now, which stay where no row applies.
type Who struct {
	Levels      map[core.Class]int
	Race        core.Race
	Dex, Wis    int
	Int         int // magic-user spells of levels 6..9 need it (IntelligenceLimit)
	Con         int // the constitution save bonus (COTAB O0FE:07FB) reads it
	ClericSlots LevelSlots
	MageSlots   LevelSlots
	DruidSlots  LevelSlots
	OldLevels   map[core.Class]int // a dual-classed character's levels in the class left
	Powers      ItemPowers
}

// ItemPowers are the readied items' powers ClassStats reads (COTAB's; POR's items have none).
type ItemPowers struct {
	Thievery, FreeAction, Wizardry bool
	SaveBonus                      bool // COTAB's item power 6: the dwarf's and halfling's constitution save bonus
}

// Stats is what ClassStats derives. Skills are zero without thief levels.
type Stats struct {
	Thac0Base   int
	Highest     int
	Attacks     int
	ClericSlots LevelSlots
	MageSlots   LevelSlots
	DruidSlots  LevelSlots // a ranger's, from RangerSpells
	Saves       map[core.Save]int
	Skills      [8]int
}

// LevelSlots are one spell class's spells by spell level 1..9.
type LevelSlots [MaxSpellLevel]int

// classOrder is the order later classes overwrite earlier ones' attacks.
var classOrder = [8]core.Class{dnd.Cleric, dnd.Druid, dnd.Fighter, dnd.Paladin, dnd.Ranger, dnd.Mage, dnd.Thief, dnd.Monk}

// ClassStats is ClassCal.RecalcClassStats over POR's tables.
func ClassStats(w Who) Stats { return POR.ClassStats(w) }

// ClassStats is ClassCal.RecalcClassStats with COTAB's superset (O0FE:000F
// calcSpellSlots, O0FE:07FB savingThrows, O0FE:0AEA thiefSkills): paladin
// and ranger spells, the monk's old-level saves, items of thievery, free
// action and wizardry. Differences are in gobox-engine/CHANGES-FROM-ORIGINAL.md.
func (t *Tables) ClassStats(w Who) Stats {
	s := Stats{ClericSlots: w.ClericSlots, MageSlots: w.MageSlots, DruidSlots: w.DruidSlots, Saves: map[core.Save]int{}}
	fresh := false // the magic-user slots were set from a row
	for _, class := range classOrder {
		lvl := w.Levels[class]
		s.Thac0Base = max(s.Thac0Base, t.Thac0At(class, lvl))
		s.Highest = max(s.Highest, lvl)
		switch class {
		case dnd.Fighter:
			s.Attacks = 2
			if lvl > 6 {
				s.Attacks = 3
			}
		case dnd.Paladin:
			if lvl > 6 {
				s.Attacks = 3
			}
			if add := rowSum(t.PaladinSpells, lvl); add != ([5]int{}) {
				if w.Levels[dnd.Cleric] <= 0 {
					s.ClericSlots = LevelSlots{}
				}
				for i := range add {
					s.ClericSlots[i] += add[i]
				}
			}
		case dnd.Ranger:
			if lvl > 7 {
				s.Attacks = 3
			}
			if add := rowSum(t.RangerSpells, lvl); add != ([5]int{}) {
				if w.Levels[dnd.Mage] <= 0 {
					s.MageSlots = LevelSlots{}
				}
				fresh = true
				s.DruidSlots = LevelSlots{add[0], add[1], add[2]}
				s.MageSlots[0] += add[3]
				s.MageSlots[1] += add[4]
			}
		case dnd.Cleric:
			if lvl = t.slotRow(lvl); lvl > 0 && lvl <= t.SlotLevels {
				s.ClericSlots = slotsAt(t.ClericSlots, lvl)
				s.ClericSlots = WisdomBonus(s.ClericSlots, w.Wis)
			}
		case dnd.Mage:
			if lvl = t.slotRow(lvl); lvl > 0 && lvl <= t.SlotLevels {
				s.MageSlots = IntelligenceLimit(slotsAt(t.MageSlots, lvl), w.Int)
				fresh = true
			}
		}
	}
	if w.Powers.Wizardry && fresh { // held slots were doubled when they were set
		for i := range s.MageSlots {
			s.MageSlots[i] *= 2
		}
	}
	for _, save := range []core.Save{dnd.SaveParalysis, dnd.SavePetrify, dnd.SaveRod, dnd.SaveBreath, dnd.SaveSpell} {
		s.Saves[save] = 20
		for _, class := range classOrder {
			if lvl := w.Levels[class]; lvl > 0 {
				s.Saves[save] = min(s.Saves[save], t.SaveAt(class, lvl, save))
			}
		}
		if old := w.OldLevels[dnd.Monk]; old < w.Levels[dnd.Monk] && old < len(t.MonkOldSaves) {
			s.Saves[save] = min(s.Saves[save], int(t.MonkOldSaves[old][saveColumn[save]]))
		}
	}
	// the constitution bonus lowers the stored paralysis save (Felix: COTAB's
	// bonus, subtracted so it helps; POR's roll-time affect 0x5A is not given
	// and dropped from loaded records, so it counts once).
	con := byte(w.Con)
	if w.Race == dnd.Dwarf || w.Race == dnd.Halfling || w.Powers.SaveBonus {
		s.Saves[dnd.SaveParalysis] -= int(goldbox.ConSaveLow(con))
	}
	s.Saves[dnd.SaveParalysis] -= int(goldbox.ConSaveHigh(con))
	if lvl := w.Levels[dnd.Thief]; lvl > 0 {
		s.Skills = t.ThiefSkills(lvl, w.Race, w.Dex, w.Powers)
	}
	return s
}

// slotRow is the slot row a class level reads: the last one past the table's end when the table clamps.
func (t *Tables) slotRow(lvl int) int {
	if t.SlotsClamp {
		return min(lvl, t.SlotLevels)
	}
	return lvl
}

// rowSum is the slots a paladin's or ranger's rows add up to at lvl (each row is a level's gain).
func rowSum(rows [][]int16, lvl int) [5]int {
	var sum [5]int
	for l := 1; l <= min(lvl, len(rows)-1); l++ {
		for i, n := range rows[l] {
			sum[i] += int(n)
		}
	}
	return sum
}

// slotsAt: level 1 is the creation's one spell; its row is unlimited.
func slotsAt(row func(int) LevelSlots, lvl int) LevelSlots {
	if lvl == 1 {
		return LevelSlots{1}
	}
	return row(lvl)
}

// WisdomBonus is ClassCal.AwardWisdomBonus: extra cleric spells for wisdom 13..18,
// only at a spell level that already has slots; sixth- and seventh-level
// spells need wisdom 17 and 18. The later games' superset (COTAB O0FE:000F,
// SotSB O164 clericSlots, POD O102:0594; ruling 18:10 #4); POR's rows have
// no fourth level, so POR plays as before.
func WisdomBonus(slots LevelSlots, wis int) LevelSlots {
	for _, b := range []struct{ above, level int }{{12, 0}, {13, 0}, {14, 1}, {15, 1}, {16, 2}, {17, 3}} {
		if wis > b.above && slots[b.level] > 0 {
			slots[b.level]++
		}
	}
	for _, m := range []struct{ level, need int }{{6, 17}, {7, 18}} {
		if wis < m.need {
			slots[m.level-1] = 0
		}
	}
	return slots
}

// IntelligenceLimit is SotSB O164:07BE / POD intelligenceSlots: magic-user
// spells of levels 6..9 need intelligence 12, 14, 16 and 18 (POR's rows end
// at level 3, so POR plays as before).
func IntelligenceLimit(slots LevelSlots, intel int) LevelSlots {
	for _, m := range []struct{ level, need int }{{6, 12}, {7, 14}, {8, 16}, {9, 18}} {
		if intel < m.need {
			slots[m.level-1] = 0
		}
	}
	return slots
}

// RollDice is FxEngine.RollDice: the sum in a byte, as the Pascal keeps it.
func RollDice(rng *core.RNG, n, sides int) int {
	var total uint8
	for range n {
		total += uint8(rng.N(sides) + 1)
	}
	return int(total)
}
