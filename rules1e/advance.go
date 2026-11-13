package rules1e

import (
	"gobox/core"
)

// classRow, raceRow, saveColumn and skillColumn say where a concept sits in
// the tables, so that callers name it and never depend on the order.
var classRow = map[core.Class]int{
	CoreClassCleric: 0, CoreClassDruid: 1, CoreClassFighter: 2, CoreClassPaladin: 3,
	CoreClassRanger: 4, CoreClassMage: 5, CoreClassThief: 6, CoreClassMonk: 7,
}

var raceRow = map[core.Race]int{
	CoreRaceMonster: 0, CoreRaceDwarf: 1, CoreRaceElf: 2, CoreRaceGnome: 3,
	CoreRaceHalfElf: 4, CoreRaceHalfling: 5, CoreRaceHalfOrc: 6, CoreRaceHuman: 7,
}

var saveColumn = map[core.Save]int{
	SaveParalysis: 0, SavePetrify: 1, SaveRod: 2, SaveBreath: 3, SaveSpell: 4,
}

var skillColumn = map[core.Skill]int{
	PickPockets: 0, OpenLocks: 1, FindTraps: 2, MoveQuietly: 3,
	Hide: 4, Hear: 5, Climb: 6, ReadLanguages: 7,
}

var thiefSkillOrder = [8]core.Skill{PickPockets, OpenLocks, FindTraps, MoveQuietly, Hide, Hear, Climb, ReadLanguages}

// xpNone is an experience entry no experience reaches: the class stops at the level before.
const xpNone = -1

// xpTop is the first level no class reaches (POD O102:12D1, Felix 20:05: all games).
const xpTop = 41

// XPForLevel is the experience level takes in class; xpNone where it can not
// be reached, 0 from xpTop. Past its table a class reads POD's table
// (XPPast), then adds its XPSteps entry per level (Felix: all games); a
// class whose table ends at xpNone stays there.
func (t *Tables) XPForLevel(class core.Class, level int) int32 {
	k := classRow[class]
	row := t.XP[k]
	if level < 0 || level >= xpTop {
		return 0
	}
	if level < len(row) {
		return row[level]
	}
	if row[len(row)-1] == xpNone {
		return xpNone
	}
	past := t.XPPast[k]
	if level < len(past) {
		return past[level]
	}
	last := len(past) - 1
	if past[last] == xpNone {
		return xpNone
	}
	return past[last] + int32(level-last)*t.XPSteps[k]
}

// Thac0At is the stored THAC0 base (thac0Tbl) of class at level. Level -1 (no
// level in the class) gives 0 and levels past the last column use the last
// column: fixed, as the Pascal reference (por-re BUGS.md, RecalcClassStats).
func (t *Tables) Thac0At(class core.Class, level int) int {
	if level < 0 || classRow[class] >= len(t.Thac0) { // a game without the class (SotSB's monk)
		return 0
	}
	return int(t.Thac0[classRow[class]][min(level, len(t.Thac0[0])-1)])
}

// TrainingDie is the dice a level of class gains.
func (t *Tables) TrainingDie(class core.Class) (count, sides int) {
	r := classRow[class]
	return int(t.HitDieCount[r]), int(t.HitDieSides[r])
}

// SaveAt is the saving throw of class at level. POR's paladin rows are COTAB's.
func (t *Tables) SaveAt(class core.Class, level int, save core.Save) int {
	if class == CoreClassPaladin && level < len(t.PaladinSaves) {
		return int(t.PaladinSaves[level][saveColumn[save]])
	}
	return int(t.Saves[classRow[class]*t.SaveRows+level][saveColumn[save]])
}

// ThiefBase is a thief's skill at level before race and dexterity.
func (t *Tables) ThiefBase(level int, skill core.Skill) int {
	return int(t.ThiefByLevel[level][skillColumn[skill]])
}

// ThiefRace is race's adjustment to skill.
func (t *Tables) ThiefRace(race core.Race, skill core.Skill) int {
	return int(t.ThiefByRace[raceRow[race]][skillColumn[skill]])
}

// ThiefDex is dexterity's adjustment to skill: only the first five skills have one.
func (t *Tables) ThiefDex(dex int, skill core.Skill) int {
	c := skillColumn[skill]
	if c >= len(t.ThiefByDex[dex]) {
		return 0
	}
	return int(t.ThiefByDex[dex][c])
}

// Items of thievery lift the first two skills to these levels, or add the bonus;
// free action lifts the level to freeActionLevel, or adds its bonus to every skill.
const (
	thieveryFirstLevel  = 5
	thieverySecondLevel = 7
	thieveryBonus       = 5
	freeActionLevel     = 4
	freeActionBonus     = 10
)

// ThiefSkills is ClassCal.RecalcThiefSkills with COTAB's items (O0FE:0AEA):
// each skill at level with race's and dexterity's adjustments, 0 when a
// race penalty exceeds the base and bonus.
func (t *Tables) ThiefSkills(level int, race core.Race, dex int, p ItemPowers) [8]int {
	var out [8]int
	freeAction := p.FreeAction
	if level < freeActionLevel && freeAction {
		level, freeAction = freeActionLevel, false
	}
	bonus := 0 // COTAB's quirk: not reset per skill, so skills 3..8 keep skill 2's
	for i, skill := range thiefSkillOrder {
		lvl := level
		if p.Thievery && i < 2 {
			if lift := [2]int{thieveryFirstLevel, thieverySecondLevel}[i]; lvl < lift {
				lvl = lift
			} else {
				bonus = thieveryBonus
			}
		}
		r, b := t.ThiefRace(race, skill), t.ThiefBase(lvl, skill)
		if !(r < 0 && -r > b-bonus) {
			out[i] = int(uint8(b + r + bonus + t.ThiefDex(dex, skill)))
		}
		if freeAction {
			out[i] = int(uint8(out[i] + freeActionBonus))
		}
	}
	return out
}

// ConHP is constitution's adjustment to hit points per die.
func (t *Tables) ConHP(con int) int { return int(t.ConHPByScore[con]) }

// ClericSlots and MageSlots are the spells of each level at a class level.
func (t *Tables) ClericSlots(level int) LevelSlots { return slots(t.ClericSpells[level]) }
func (t *Tables) MageSlots(level int) LevelSlots   { return slots(t.MageSpells[level]) }

func slots(r []int16) LevelSlots {
	var s LevelSlots
	for i := range min(len(r), len(s)) {
		s[i] = int(r[i])
	}
	return s
}
