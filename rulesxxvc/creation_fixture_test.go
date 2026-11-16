package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// br1Creation is BR1's xxvc.Creation as the importer reads it (buckroger-re chargen.pas, skills.pas, gdconsts.pas
// careerBaseCombat, experienceNeeded; the importer's test holds the EXE's bytes to these).
var br1Creation = xxvc.Creation{
	StartLevel: 2, FixedGainLevel: 9, CareerPoints: 80, GeneralPoints: 40, MaxPerSkill: 30, StartGeneralSkills: 7,
	ExperienceNeeded: map[rules.Class][]int64{
		xxvc.RocketJock: {1250, 2500, 5000, 10000, 20000, 40000, 70000, 110000},
		xxvc.Medic:      {1500, 3000, 6000, 12000, 24000, 48000, 96000, 192000},
		xxvc.Warrior:    {2000, 4000, 8000, 16000, 32000, 64000, 125000, 250000},
		xxvc.Engineer:   {1250, 2500, 5000, 10000, 20000, 40000, 70000, 110000},
		xxvc.Rogue:      {1250, 2500, 5000, 10000, 20000, 40000, 70000, 110000},
		xxvc.Scout:      {2250, 4500, 9000, 18000, 36000, 75000, 150000, 300000},
	},
	BaseCombat: map[rules.Class][]int32{
		xxvc.RocketJock: {40, 40, 40, 41, 41, 42, 42, 43, 43, 44},
		xxvc.Medic:      {40, 40, 40, 40, 42, 42, 42, 44, 44, 44},
		xxvc.Warrior:    {40, 40, 41, 42, 43, 44, 45, 46, 47, 48},
		xxvc.Engineer:   {40, 40, 40, 40, 42, 42, 42, 44, 44, 44},
		xxvc.Rogue:      {40, 40, 40, 41, 41, 42, 42, 43, 43, 44},
		xxvc.Scout:      {40, 40, 41, 42, 43, 44, 45, 46, 47, 48},
	},
}

// newBR1 is a ruleset holding BR1's creation numbers.
func newBR1() *Rules {
	r := New()
	r.tables.Creation = br1Creation
	return r
}
