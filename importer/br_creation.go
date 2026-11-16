package importer

import (
	"encoding/binary"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// The rows of the career tables: one per career in xxvc.Careers' order.
const (
	initialSkillRows   = 12 // gdconsts.pas careerInitialSkills[career, 1..12]
	initialSkillStride = 2 * initialSkillRows
	experienceSize     = 4 // LongInt
)

// CareerCreation is e.Creation with the career tables read from the data segment ds (GDConsts: experienceNeeded,
// careerBaseCombat, careerInitialSkills).
func CareerCreation(ds []byte, e BREXE) xxvc.Creation {
	c, t := e.Creation, e.CareerTables
	c.ExperienceNeeded = map[rules.Class][]int64{}
	c.BaseCombat = map[rules.Class][]int32{}
	c.InitialSkills = map[rules.Class][]xxvc.InitialSkill{}
	for i, career := range xxvc.Careers {
		at := t.Experience + i*t.ExperienceLevels*experienceSize
		for l := range t.ExperienceLevels {
			c.ExperienceNeeded[career] = append(c.ExperienceNeeded[career], int64(binary.LittleEndian.Uint32(ds[at+l*experienceSize:])))
		}
		at = t.BaseCombat + i*t.BaseCombatLevels
		for l := range t.BaseCombatLevels {
			c.BaseCombat[career] = append(c.BaseCombat[career], int32(ds[at+l]))
		}
		if t.InitialSkills == 0 {
			continue
		}
		at = t.InitialSkills + i*initialSkillStride
		for k := range initialSkillRows {
			if n := int(ds[at+2*k]); n != 0 { // skill 0 ends the row (ApplyInitialSkills)
				c.InitialSkills[career] = append(c.InitialSkills[career], xxvc.InitialSkill{Skill: xxvc.Skills[n-1], Value: int32(ds[at+2*k+1])})
				continue
			}
			break
		}
	}
	return c
}
