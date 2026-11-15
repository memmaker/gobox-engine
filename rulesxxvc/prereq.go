package rulesxxvc

// SkillNo is a skill's number in the original (c_skill.pas, 1..84): the
// record's skill levels are indexed by it. The roster's names (core.Skill)
// map to it once, at import.
type SkillNo int

// The skills c_skill.pas names.
const (
	RepairElectrical  SkillNo = 1
	RepairMechanical  SkillNo = 2
	RepairEngine      SkillNo = 3
	RepairLifeSupport SkillNo = 4
	RepairHull        SkillNo = 5
	JuryRig           SkillNo = 6
	RepairComputer    SkillNo = 15
	TreatLight        SkillNo = 17
	TreatSerious      SkillNo = 18
	TreatCritical     SkillNo = 19
	TreatPoisoning    SkillNo = 20
	TreatDisease      SkillNo = 22
	Diagnose          SkillNo = 23
	Composition       SkillNo = 26
	BattleTacticsNo   SkillNo = 31
	Mathematics       SkillNo = 27
	Planetology       SkillNo = 30
	Astronomy         SkillNo = 36
	Literature        SkillNo = 38
	Programming       SkillNo = 43
	Astrogation       SkillNo = 46
	Navigation        SkillNo = 47
	PlayInstrument    SkillNo = 68
	BefriendAnimal    SkillNo = 74
	AnimalTraining    SkillNo = 75
	Singing           SkillNo = 77
	PlanetarySurvival SkillNo = 84

	firstMedical SkillNo = 16 // FIRST_MEDICAL_SKILL..LAST_MEDICAL_SKILL: medics only
	lastMedical  SkillNo = 23
	SkillCount           = 84
)

// SkillLevels are a character's skill levels (c_skill.pas skills[1..84]).
// Read and write them only through Level and Set, which take the SkillNo;
// Levels is exported for the JSON save alone (Levels[0] is skill 1).
type SkillLevels struct{ Levels [SkillCount]int8 }

// Level is skill s's level; 0 for a number outside 1..SkillCount.
func (l *SkillLevels) Level(s SkillNo) int8 {
	if s < 1 || s > SkillCount {
		return 0
	}
	return l.Levels[s-1]
}

// Set sets skill s's level; a number outside 1..SkillCount is ignored.
func (l *SkillLevels) Set(s SkillNo, v int8) {
	if s >= 1 && s <= SkillCount {
		l.Levels[s-1] = v
	}
}

// Need is one prerequisite: a level in another skill.
type Need struct {
	Skill SkillNo
	Level int8
}

// prerequisite is one skill's needs: all of them, or any one when anyOf.
type prerequisite struct {
	needs []Need
	anyOf bool
}

// prerequisites is charhlp.pas MeetsSkillPrerequisite's table with
// c_skill.pas's *_NEEDS_* levels (and charhlp.pas COMPOSITION_NEEDS_ARTS).
var prerequisites = map[SkillNo]prerequisite{
	RepairEngine:      {needs: []Need{{RepairMechanical, 10}}},
	RepairLifeSupport: {needs: []Need{{RepairElectrical, 30}, {RepairMechanical, 25}}},
	RepairHull:        {needs: []Need{{RepairMechanical, 10}}},
	JuryRig:           {needs: []Need{{RepairElectrical, 10}, {RepairMechanical, 10}}},
	RepairComputer:    {needs: []Need{{RepairElectrical, 40}}},
	TreatSerious:      {needs: []Need{{TreatLight, 30}}},
	TreatCritical:     {needs: []Need{{TreatSerious, 40}}},
	TreatPoisoning:    {needs: []Need{{Diagnose, 15}}},
	TreatDisease:      {needs: []Need{{Diagnose, 25}}},
	Composition:       {needs: []Need{{Singing, 20}, {PlayInstrument, 20}, {Literature, 20}}, anyOf: true},
	Programming:       {needs: []Need{{Mathematics, 10}}},
	Astrogation:       {needs: []Need{{Astronomy, 20}, {Mathematics, 25}}},
	Navigation:        {needs: []Need{{Astronomy, 15}, {Mathematics, 25}}},
	AnimalTraining:    {needs: []Need{{BefriendAnimal, 20}}},
	PlanetarySurvival: {needs: []Need{{Planetology, 10}}},
}

// MeetsPrerequisite is charhlp.pas MeetsSkillPrerequisite: the medical
// skills are a medic's only, and a listed skill needs its levels.
func MeetsPrerequisite(l *SkillLevels, medic bool, s SkillNo) bool {
	if s >= firstMedical && s <= lastMedical && !medic {
		return false
	}
	p, ok := prerequisites[s]
	if !ok {
		return true
	}
	for _, n := range p.needs {
		met := l.Level(n.Skill) >= n.Level
		if met == p.anyOf {
			return met
		}
	}
	return !p.anyOf
}
