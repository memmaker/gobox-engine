package rulesxxvc

import (
	"slices"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// trainer is a warrior of level 2 with the experience for any level, in BR1's numbers and two weapons to specialize in.
func trainer(t *testing.T) (*Rules, rules.CharID) {
	t.Helper()
	r := newBR1()
	r.tables.Items = map[rules.ItemID]xxvc.ItemTemplate{1: {Type: 11}, 2: {Type: 12}}
	id := r.Add(Character{Name: "Hal", Career: xxvc.Warrior, Level: 2, Exp: 1 << 30, Status: StOkay, HP: 10, MaxHP: 10,
		Scores: map[rules.Ability]int{rules.Constitution: 12}, Saves: [8]int{10, 10, 10, 10, 10, 10, 10, 10}})
	return r, id
}

// TestTrainRefusals: chargen.pas TrainingMenu's three refusals, in its order.
func TestTrainRefusals(t *testing.T) {
	r, id := trainer(t)
	c := r.chars[id]
	if card := r.trainCard(c); card.Why != "" || !card.Rises || !slices.Equal(card.Becomes, []string{"a level 3 Warrior"}) {
		t.Errorf("a trainable member: %+v", card)
	}
	c.Exp = 3999 // experienceNeeded[Warrior, 3] is 4000
	if why := r.trainWhy(c); why != hallNoExp {
		t.Errorf("short of experience: %q", why)
	}
	c.Exp, c.Level = 1<<30, int(br1Creation.MaxLevel)
	if why := r.trainWhy(c); why != hallMaxLevel {
		t.Errorf("at the maximum level: %q", why)
	}
	c.Status = StDead
	if why := r.trainWhy(c); why != hallUnconscious {
		t.Errorf("unconscious: %q", why)
	}
}

// TestTrainLevel: a level takes the table's base combat, AddLevelHitPoints, holds the experience under the next level,
// takes one off each saving throw every third level, and owes the picks: the weapon of a warrior's odd level, then the
// level's 40 career and 20 general points, 15 at most a skill.
func TestTrainLevel(t *testing.T) {
	r, id := trainer(t)
	r.hall = true
	o, err := r.Resolve(nil, rules.Train{Who: id}, &rules.RNG{State: 5})
	if err != nil {
		t.Fatal(err)
	}
	c := r.chars[id]
	if c.Level != 3 || c.BaseCombat != int(br1Creation.Combat(xxvc.Warrior, 3)) || c.MaxHP <= 10 || c.HP != c.MaxHP {
		t.Errorf("level %d combat %d hp %d/%d", c.Level, c.BaseCombat, c.HP, c.MaxHP)
	}
	if want := br1Creation.Experience(xxvc.Warrior, 4) - 1; c.Exp != want {
		t.Errorf("experience %d, want %d (just under level 4)", c.Exp, want)
	}
	if c.Saves[0] != 9 || r.hall || !slices.Contains(o.Effects, rules.Effect(rules.PicksOwed{Who: id})) {
		t.Errorf("saves %v hall %v effects %v", c.Saves, r.hall, o.Effects)
	}
	cr := creation{r: r}
	d, _ := cr.Draft(id)
	d.Rise = true
	fields := cr.Fields(d)
	if len(fields) != 3 || fields[0].ID != fieldSpecialization || fields[1].Budget != 40 || fields[2].Budget != 20 || fields[1].Max != 15 {
		t.Fatalf("the picks of level 3: %+v", fields)
	}
	if cr.Rolls(d) != nil {
		t.Error("a level's picks are not rolled")
	}
	d.Picks[fieldSpecialization] = 1
	career := fields[1]
	var name string
	for i, row := range career.Rows {
		if row.Open && row.Value == 0 {
			name = career.Options[i]
			break
		}
	}
	d = spend(t, cr, d, fieldCareerSkills, name, 3)
	if _, _, err := cr.Make(d); err != nil {
		t.Fatal(err)
	}
	n, _ := skillByName(name)
	if c.Skills.Level(n) != 3 || c.Specialized(12) != 1 {
		t.Errorf("skill %s %d, specialized %d", name, c.Skills.Level(n), c.Specialized(12))
	}
	if r.hall {
		t.Error("the hall is still up")
	}
}

// TestTrainWeaponAtMost: a weapon is trained at most MAX_SPECIALIZATION (3) times (viewchar.pas SpecializationMenu).
func TestTrainWeaponAtMost(t *testing.T) {
	r, id := trainer(t)
	r.chars[id].Specializations[0], r.chars[id].Specializations[1], r.chars[id].Specializations[2] = 11, 11, 11
	r.chars[id].Level = 3
	cr := creation{r: r}
	d, _ := cr.Draft(id)
	d.Rise = true
	if cr.Fits(d, fieldSpecialization, 0) || !cr.Fits(d, fieldSpecialization, 1) {
		t.Error("a weapon trained three times is still offered, or the other is not")
	}
}

// TestTrainMaxGeneral: after the start level a member may begin one more general skill than it knew (skills.pas
// EditSkills: maxGeneral := knownGeneralCount + 1).
func TestTrainMaxGeneral(t *testing.T) {
	r, id := trainer(t)
	c := r.chars[id]
	c.Level = 3
	var general []SkillNo
	_, all := skillLists(xxvc.Warrior)
	for _, n := range all {
		if MeetsPrerequisite(&SkillLevels{}, false, n) && len(general) < 9 {
			general = append(general, n)
		}
	}
	for _, n := range general[:8] { // eight general skills known: more than the start level's seven
		c.Skills.Set(n, 5)
	}
	cr := creation{r: r}
	d, _ := cr.Draft(id)
	d.Rise = true
	d = spend(t, cr, d, fieldGeneralSkills, skillName(general[8]), 2) // the ninth: one more than the eight known, and a second point on it
	for _, n := range all {
		if !slices.Contains(general, n) && MeetsPrerequisite(&SkillLevels{}, false, n) {
			if got := refusal(t, cr, d, fieldGeneralSkills, skillName(n)); got != rules.SpendNoNewItem {
				t.Errorf("a tenth skill: refusal %v, want no new skill", got)
			}
			return
		}
	}
}
