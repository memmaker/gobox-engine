package rulesxxvc

import (
	"testing"

	"gobox/rules/xxvc"
)

// Against skillDefinitions (gdconsts.pas) by hand: rows 23 Diagnose, 24
// General Knowledge, 50 Disguise, 51 Pilot Rocket, 69 Play Instrumnt, 70
// Hypnosis, 79 Etiquette, 80 Tracking.
func TestGroupsAndRoster(t *testing.T) {
	for s, g := range map[SkillNo]SkillGroup{1: GroupTechnical, 23: GroupTechnical, 24: GroupIntelligence,
		50: GroupIntelligence, 51: GroupDexterity, 69: GroupDexterity, 70: GroupCharisma, 79: GroupCharisma,
		80: GroupWisdom, 84: GroupWisdom} {
		if Group(s) != g {
			t.Errorf("Group(%d) = %d, want %d", s, Group(s), g)
		}
	}
	if Offered(12) || !Offered(13) || Offered(82) || !Offered(83) {
		t.Error("Offered: Cook and Read Lips are not, First Aid and Notice are")
	}
	if len(xxvc.Skills) != SkillCount || xxvc.Skills[BattleTacticsNo-1] != xxvc.BattleTactics ||
		xxvc.Skills[PlanetarySurvival-1] != "planetary-survival" || xxvc.Skills[Composition-1] != "composition" {
		t.Errorf("xxvc.Skills does not match skillDefinitions: %d", len(xxvc.Skills))
	}
	if CareerSkills[CareerMedic][7] != 16 || CareerSkills[CareerScout][0] != 83 {
		t.Error("careerSkills")
	}
}
