package rulesxxvc

import "testing"

// Cases by hand from charhlp.pas MeetsSkillPrerequisite and c_skill.pas.
func TestMeetsPrerequisite(t *testing.T) {
	var none, skilled SkillLevels
	skilled.Set(RepairElectrical, 30)
	skilled.Set(RepairMechanical, 24)
	skilled.Set(Literature, 20)
	skilled.Set(TreatLight, 30)
	for _, c := range []struct {
		l     *SkillLevels
		medic bool
		s     SkillNo
		want  bool
	}{
		{&none, false, 7, true},                     // unlisted
		{&none, false, RepairEngine, false},         // mechanical 10
		{&skilled, false, RepairEngine, true},       // 24 >= 10
		{&skilled, false, RepairLifeSupport, false}, // mechanical 24 < 25
		{&skilled, false, JuryRig, true},            // 30, 24 >= 10, 10
		{&skilled, false, RepairComputer, false},    // electrical 30 < 40
		{&skilled, false, Composition, true},        // literature 20, any one
		{&none, false, Composition, false},
		{&none, false, 16, false},          // medical, not a medic
		{&none, true, 16, true},            // medical, unlisted
		{&none, true, TreatSerious, false}, // treat light 30
		{&skilled, true, TreatSerious, true},
		{&skilled, false, TreatSerious, false}, // a medic's only
		{&none, false, Diagnose, false},        // 23 is medical
	} {
		if got := MeetsPrerequisite(c.l, c.medic, c.s); got != c.want {
			t.Errorf("skill %d medic %v: %v, want %v", c.s, c.medic, got, c.want)
		}
	}
}
