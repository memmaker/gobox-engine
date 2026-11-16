package importer

import (
	"slices"
	"testing"

	"gobox/asset/orig"
	"gobox/rules"
	"gobox/rules/xxvc"
)

// TestCareerCreation: the career tables the import reads from each game's data segment are gdconsts.pas' literals
// (buckroger-re / buckrogers2-re): experienceNeeded, careerBaseCombat, and BR2's careerInitialSkills; BR1 has no initial skills.
func TestCareerCreation(t *testing.T) {
	for _, g := range []struct {
		id   string
		exe  BREXE
		exp  map[rules.Class][]int64 // the first and the last entries of experienceNeeded
		base map[rules.Class][]int32
		init map[rules.Class][]xxvc.InitialSkill
	}{
		{"br1", BR1EXE,
			map[rules.Class][]int64{xxvc.RocketJock: {1250, 110000}, xxvc.Medic: {1500, 192000}, xxvc.Scout: {2250, 300000}},
			map[rules.Class][]int32{xxvc.RocketJock: {40, 44}, xxvc.Warrior: {40, 48}},
			map[rules.Class][]xxvc.InitialSkill{}},
		{"br2", BR2EXE,
			map[rules.Class][]int64{xxvc.RocketJock: {1250, 112640000}, xxvc.Warrior: {2000, 256000000}, xxvc.Scout: {2250, 384000000}},
			map[rules.Class][]int32{xxvc.RocketJock: {40, 49}, xxvc.Medic: {40, 52}, xxvc.Warrior: {40, 59}},
			map[rules.Class][]xxvc.InitialSkill{
				xxvc.RocketJock: {{Skill: "drive-ground-car", Value: 40}, {Skill: "maneuver-in-0g", Value: 85}},
				xxvc.Medic:      {{Skill: "diagnose", Value: 15}, {Skill: "treat-critical-wounds", Value: 75}},
			}},
	} {
		im, err := orig.Load("../data/"+g.id+"/sources/dos/START.EXE", g.exe.DS)
		if err != nil {
			t.Fatal(err)
		}
		c := CareerCreation(im.Mem[im.DS:], g.exe)
		for career, want := range g.exp {
			row := c.ExperienceNeeded[career]
			if len(row) != g.exe.CareerTables.ExperienceLevels || row[0] != want[0] || row[len(row)-1] != want[1] {
				t.Errorf("%s %s experience %v, want %v..%v", g.id, career, row, want[0], want[1])
			}
		}
		for career, want := range g.base {
			row := c.BaseCombat[career]
			if len(row) != g.exe.CareerTables.BaseCombatLevels || row[0] != want[0] || row[len(row)-1] != want[1] {
				t.Errorf("%s %s base combat %v, want %v..%v", g.id, career, row, want[0], want[1])
			}
		}
		for career, want := range g.init {
			if got := c.InitialSkills[career]; !slices.Equal(got[:len(want)], want) {
				t.Errorf("%s %s initial skills %v, want %v first", g.id, career, got, want)
			}
		}
		if len(g.init) == 0 && len(c.InitialSkills) != 0 {
			t.Errorf("%s has initial skills: %v", g.id, c.InitialSkills)
		}
		if got := c.Experience(xxvc.Warrior, c.StartLevel); got == 0 {
			t.Errorf("%s: no experience for the start level %d", g.id, c.StartLevel)
		}
	}
}
