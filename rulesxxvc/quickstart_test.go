package rulesxxvc

import (
	"testing"

	"gobox/engine"
	"gobox/rules"
	"gobox/rules/xxvc"
)

// keepHost is a creation host that keeps a copy of each character Make keeps.
type keepHost struct {
	seamHost
	r     *Rules
	kinds []rules.PickKind
	kept  []Character
}

func (h *keepHost) Keep(id rules.CharID) error { h.kept = append(h.kept, *h.r.chars[id]); return nil }
func (h *keepHost) QuickParty() []rules.Template {
	var out []rules.Template
	for _, k := range h.kinds {
		o, err := h.r.pickKind(k)
		if err != nil {
			panic(err)
		}
		p := o.Effects[0].(rules.Picked)
		out = append(out, rules.Template{Picks: p.Picks, Spend: p.Spend})
	}
	return out
}

// TestQuickStartSpends: Quick Start's warrior takes its weapon and every member spends its skill points, all career
// points on the career skills, the general ones on the skill named up to its most (buckrogers-go quickCreate: quickWeapon, quickSkills); they were left
// unspent and the weapon the first on the list.
func TestQuickStartSpends(t *testing.T) {
	r := newBR1()
	r.tables.Items = map[rules.ItemID]xxvc.ItemTemplate{}
	for id := rules.ItemID(1); id <= lastSpecializableTemplate; id++ {
		r.tables.Items[id] = xxvc.ItemTemplate{Type: int32(id)}
	}
	h := &keepHost{r: r, kinds: []rules.PickKind{
		{Race: "Desert Runner", Gender: "Male", Class: "Warrior", Weapon: xxvc.ItemIDLaserRifle, Skill: xxvc.FirstAid},
		{Race: "Tinker", Gender: "Female", Class: "Engineer", Skill: xxvc.RepairWeapon},
	}}
	if _, err := engine.QuickParty(creation{r: r, host: h}, &rules.RNG{State: 3}); err != nil {
		t.Fatal(err)
	}
	if len(h.kept) != 2 {
		t.Fatalf("kept %d", len(h.kept))
	}
	if got := h.kept[0].Specializations[0]; got != int32(xxvc.ItemIDLaserRifle) {
		t.Errorf("the warrior's weapon type %d, want the Laser Rifle's %d", got, xxvc.ItemIDLaserRifle)
	}
	for i, want := range []rules.Skill{xxvc.FirstAid, xxvc.RepairWeapon} {
		c := h.kept[i]
		careerList, _ := skillLists(c.Career)
		spent, start := 0, startingSkills(r.tables.Creation, c.Career)
		for _, n := range careerList {
			spent += int(c.Skills.Level(n)) - int(start.Level(n))
		}
		if spent != int(r.tables.Creation.CareerPoints) {
			t.Errorf("%s: %d career points spent, want all %d", c.Career, spent, r.tables.Creation.CareerPoints)
		}
		if got := int32(c.Skills.Level(skillNumber(want))); got != r.tables.Creation.MaxPerSkill { // the rest stays, as the reference's Done leaves it
			t.Errorf("%s: %s at %d, want the most a skill takes, %d", c.Career, want, got, r.tables.Creation.MaxPerSkill)
		}
	}
}
