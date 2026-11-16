package rules1e

import "gobox/rules"

// ScriptFields are the member fields the Pool games' layouts name that this ruleset answers (rules.Catalog.ScriptFields):
// the class levels cotab and sotsb read, the thief skills (por Thief[0..7], pod thiefSkills[]) and the save the
// scripts compare against. Census: the fields opCheckParty averages (por SelFindTraps 0x6BA7; cotab's census covers
// all eight thief skills, cvThiefFirst..Last, but the engine has always counted these three). InSave: the four the
// engine bound as stored cells.
var ScriptFields = map[string]rules.ScriptField{
	"SelMageLevel":     {Fact: rules.LevelFact{Class: ClassMage}},   // cotab $C9, sotsb $D0
	"SelRangerLevel":   {Fact: rules.LevelFact{Class: ClassRanger}}, // cotab $C9, sotsb $D0
	"SelSaveVsPetrify": {Fact: rules.SaveFact{Save: SavePetrify}, InSave: true},
	"SelPickPockets":   {Fact: rules.SkillFact{Skill: PickPockets}, Census: true, InSave: true},
	"SelOpenLocks":     {Fact: rules.SkillFact{Skill: OpenLocks}},
	"SelFindTraps":     {Fact: rules.SkillFact{Skill: FindTraps}, Census: true, InSave: true},
	"SelMoveSilently":  {Fact: rules.SkillFact{Skill: MoveQuietly}},
	"SelHideShadows":   {Fact: rules.SkillFact{Skill: Hide}},
	"SelHearNoise":     {Fact: rules.SkillFact{Skill: Hear}},
	"SelThiefSkill6":   {Fact: rules.SkillFact{Skill: Climb}, Census: true, InSave: true},
	"SelReadLanguages": {Fact: rules.SkillFact{Skill: ReadLanguages}},
}
