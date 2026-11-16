package xxvc

import "gobox/rules"

// InitialSkill is a career's starting skill and its level (types.pas TInitialSkill, gdconsts.pas careerInitialSkills).
type InitialSkill struct {
	Skill rules.Skill
	Value int32
}

// Creation is what chargen.pas, skills.pas and gdconsts.pas fix for the characters of one game: BR1 makes a level 2
// character with 80, 40 and 30 skill points to spend, BR2 a level 7 one whose career skills are given. The importer
// writes it per game (importer BREXE); the ruleset reads it and names no game.
type Creation struct {
	// StartLevel is the level free training stops at (chargen.pas: level := 2; BR2: 7); the character starts with the
	// experience ExperienceNeeded gives for it.
	StartLevel int32
	// FixedGainLevel is the first level whose hit points are a fixed gain (chargen.pas AddLevelHitPoints: 9; BR2: 10).
	FixedGainLevel int32
	// CareerPoints, GeneralPoints and MaxPerSkill are skills.pas EditSkills's arguments for a new character
	// (BR1 80, 40, 30; BR2 0, 0, 105: nothing to spend).
	CareerPoints, GeneralPoints, MaxPerSkill int32
	// StartGeneralSkills is EditSkills's maxGeneral at the start level: no more general skills known than that
	// (skills.pas: level 2 gives 7; BR2: NEW_CHARACTER_GENERAL_SKILLS 12).
	StartGeneralSkills int32
	// ExperienceNeeded is gdconsts.pas experienceNeeded by career: the experience for level 2, 3, ... (index level-2).
	ExperienceNeeded map[rules.Class][]int64
	// BaseCombat is gdconsts.pas careerBaseCombat by career and level (index level, from 0).
	BaseCombat map[rules.Class][]int32
	// InitialSkills is gdconsts.pas careerInitialSkills (BR2 only: Skills.ApplyInitialSkills), in the original's order.
	InitialSkills map[rules.Class][]InitialSkill
}

// Experience is the experience a career needs for level (2 and up); 0 outside the table.
func (c Creation) Experience(career rules.Class, level int32) int64 {
	row := c.ExperienceNeeded[career]
	if level < 2 || int(level-2) >= len(row) {
		return 0
	}
	return row[level-2]
}

// Combat is the career's base combat at level (0 and up); 0 outside the table.
func (c Creation) Combat(career rules.Class, level int32) int32 {
	row := c.BaseCombat[career]
	if level < 0 || int(level) >= len(row) {
		return 0
	}
	return row[level]
}
