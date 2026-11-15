package rulesxxvc

import (
	"gobox/rules"
)

// SkillValue is charhlp.pas SkillValue: the level plus the group's ability
// when the prerequisite is met (SkillAbilityScore), 0 untrained or disabled.
func SkillValue(level, ability int8, prerequisiteMet, enabled bool) int8 {
	if level <= 0 || !enabled {
		return 0
	}
	if !prerequisiteMet {
		ability = 0
	}
	return level + ability
}

// SkillCheck is charhlp.pas SelectSkillCharacter over the candidates' skill
// values (one character, or the whole party in order): each rolls a d100
// against its scaled value, the best margin wins. who is the index of the
// winner. As the original, the bands use the LAST candidate's scaled value,
// not the winner's. No candidates: core.FailedBadly, who -1.
func SkillCheck(values []int8, h rules.Hardness, r *rules.RNG) (g rules.Grade, who int) {
	best, who := int16(-10000), -1
	var adjusted int16
	for i, v := range values {
		adjusted = int16(v)
		switch h {
		case rules.VeryEasy:
			adjusted <<= 2
		case rules.Easy:
			adjusted <<= 1
		case rules.Hard:
			adjusted = int16(uint16(adjusted) / 2)
		case rules.VeryHard:
			adjusted = int16(uint16(adjusted) / 4)
		}
		if roll := adjusted - int16(rules.RNGDice(r, 1, 100)); roll > best {
			best, who = roll, i
		}
	}
	// Integer against Word compares as LongInt.
	band := int32(uint16(adjusted))
	if best >= 0 {
		if int32(best) < band/2 {
			return rules.Made, who
		}
		return rules.MadeWell, who
	}
	if -int32(best) > band {
		return rules.FailedBadly, who
	}
	return rules.Failed, who
}
