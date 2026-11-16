package rulesxxvc

import (
	"encoding/json"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

func TestConvertBRCharacter(t *testing.T) {
	b := BRSave{Name: "Buck", Abilities: [7]int{17, 16, 15, 14, 13, 12, 11}, Race: 1, Career: 3, Level: 2, Age: 27,
		CreditTotal: 500, HitPoints: 9, HitPointMaximum: 12, Status: StOkay, IsEnabled: true,
		ArmorFlags: armorBodyBit | armorECMBit,
		Items:      []BRSavedItem{{ItemType: 5, NameIDs: [3]int32{1, 2, 0}, Plus: 1, Quantity: 1, Readied: true}},
		Effects:    []Effect{{ID: EffStunned, Minutes: 30}}}
	b.SkillLevels[0] = 40
	tb := xxvc.Tables{Items: map[rules.ItemID]xxvc.ItemTemplate{9: {Type: 5, Words: [3]int32{1, 2, 0}}, 7: {Type: 5, Words: [3]int32{1, 2, 0}}}}
	c, err := ConvertBRCharacter(tb, b)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "Buck" || c.Age != 27 || c.Race != xxvc.Terran || c.Career != xxvc.Warrior || c.Scores[xxvc.Tech] != 11 ||
		c.Coins != 500 || c.HP != 9 || c.MaxHP != 12 || !c.Enabled || c.Skills.Level(1) != 40 || !c.Armor.Body || !c.Armor.ECM || c.Armor.Fields {
		t.Fatalf("character %+v", c)
	}
	if len(c.Items) != 1 || c.Items[0].Ref.ID != 7 || !c.Items[0].Readied || c.Items[0].Ref.Plus != 1 {
		t.Fatalf("items %+v", c.Items)
	}
	if _, err := ConvertBRCharacter(tb, BRSave{Items: []BRSavedItem{{ItemType: 99}}}); err == nil {
		t.Fatal("unknown item converted")
	}
}

// TestSkillIndexing reads intimidate both ways: the save's 0-based slot and
// the SkillNo the fight menu asks for (fightactions.go read it one low).
func TestSkillIndexing(t *testing.T) {
	var b BRSave
	b.SkillLevels[intimidateSkill-1] = 10
	c, err := ConvertBRCharacter(xxvc.Tables{}, b)
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := skillNo("intimidate"); n != intimidateSkill || c.Skills.Level(intimidateSkill) != 10 || c.Skills.Level(intimidateSkill-1) != 0 {
		t.Fatalf("intimidate %d: level %d", intimidateSkill, c.Skills.Level(intimidateSkill))
	}
	j, _ := json.Marshal(c.Skills)
	var back SkillLevels
	if json.Unmarshal(j, &back) != nil || back != c.Skills {
		t.Fatalf("json %s", j)
	}
}
