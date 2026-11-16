package rulesxxvc

import (
	"errors"
	"slices"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// skillDraft is a Terran draft of career, its skill screen not begun.
func skillDraft(t *testing.T, career rules.Class) (creation, rules.Draft) {
	t.Helper()
	c := creation{r: newBR1()}
	return c, newDraft(t, c, 0, 0, slices.Index(pickCareers, career), 5)
}

// skillByName is the skill number of an option's name.
func skillByName(name string) (SkillNo, bool) {
	for n := SkillNo(1); n <= SkillCount; n++ {
		if skillName(n) == name {
			return n, true
		}
	}
	return 0, false
}

func fieldOf(c creation, d rules.Draft, id rules.FieldID) rules.FieldDef {
	for _, f := range c.Fields(d) {
		if f.ID == id {
			return f
		}
	}
	return rules.FieldDef{}
}

// spend puts n points on the option named name of field f, failing the test on a refusal.
func spend(t *testing.T, c creation, d rules.Draft, f rules.FieldID, name string, n int) rules.Draft {
	t.Helper()
	o := slices.Index(fieldOf(c, d, f).Options, name)
	if o < 0 {
		t.Fatalf("no option %q", name)
	}
	for range n {
		var err error
		if d, err = c.Spend(d, f, o, 1); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	return d
}

func refusal(t *testing.T, c creation, d rules.Draft, f rules.FieldID, name string) rules.SpendReason {
	t.Helper()
	o := slices.Index(fieldOf(c, d, f).Options, name)
	_, err := c.Spend(d, f, o, 1)
	var sr *rules.SpendRefusal
	if !errors.As(err, &sr) {
		t.Fatalf("%s: %v, want a refusal", name, err)
	}
	return sr.Reason
}

// TestCareerSkillPoints: skills.pas EditSkills(80, 40, 30) on the career's eight skills: 80 points, at most 30 on
// one skill, a skill with a prerequisite refused until the other skill has its level.
func TestCareerSkillPoints(t *testing.T) {
	c, d := skillDraft(t, xxvc.Medic)
	f := fieldOf(c, d, fieldCareerSkills)
	if len(f.Options) != 8 || f.Budget != 80 || f.Max != 30 || d.PointsLeft(f) != 80 {
		t.Fatalf("field %+v", f)
	}
	serious := slices.Index(f.Options, skillName(TreatSerious))
	if row := f.Rows[serious]; row.Open || len(row.Wants) != 1 || row.Wants[0].Name != skillName(TreatLight) || row.Wants[0].Level != 30 {
		t.Fatalf("serious: %+v", row)
	}
	if refusal(t, c, d, fieldCareerSkills, skillName(TreatSerious)) != rules.SpendWants {
		t.Error("treat serious was not refused for its prerequisite")
	}
	d = spend(t, c, d, fieldCareerSkills, skillName(TreatLight), 30)
	if refusal(t, c, d, fieldCareerSkills, skillName(TreatLight)) != rules.SpendAtMost {
		t.Error("31 points on a skill")
	}
	f = fieldOf(c, d, fieldCareerSkills)
	if row := f.Rows[slices.Index(f.Options, skillName(TreatLight))]; row.Value != 30 || row.Bonus != d.Sheet.Scores[xxvc.Tech] || row.Total != 30+row.Bonus {
		t.Errorf("light: %+v (tech %d)", row, d.Sheet.Scores[xxvc.Tech])
	}
	if row := f.Rows[serious]; !row.Open || len(row.Wants) != 0 {
		t.Errorf("serious after 30 light: %+v", row)
	}
	d = spend(t, c, d, fieldCareerSkills, skillName(TreatSerious), 30)
	d = spend(t, c, d, fieldCareerSkills, skillName(Diagnose), 20)
	if d.PointsLeft(fieldOf(c, d, fieldCareerSkills)) != 0 || refusal(t, c, d, fieldCareerSkills, skillName(Diagnose)) != rules.SpendNoPoints {
		t.Errorf("80 points spent: left %v", d.Left)
	}
	// Taking light below 30 resets serious: its 30 points come back.
	o := slices.Index(f.Options, skillName(TreatLight))
	d, err := c.Spend(d, fieldCareerSkills, o, -1)
	if err != nil {
		t.Fatal(err)
	}
	f = fieldOf(c, d, fieldCareerSkills)
	if left := d.PointsLeft(f); left != 31 || f.Rows[serious].Value != 0 || d.Points[fieldCareerSkills][serious] != 0 {
		t.Errorf("after taking one back: left %d, serious %+v", left, f.Rows[serious])
	}
	var sum int32
	for _, p := range d.Points[fieldCareerSkills] {
		sum += p
	}
	if sum+d.PointsLeft(f) != 80 { // 29 on light and 20 on diagnose, 31 left: the 30 on serious came back
		t.Errorf("points %v left %d", d.Points[fieldCareerSkills], d.PointsLeft(f))
	}
}

// TestGeneralSkillPoints: 40 points on the general skills (the career's skills and the medical ones are not
// among them), at most seven skills known at level 2, a prerequisite met first.
func TestGeneralSkillPoints(t *testing.T) {
	c, d := skillDraft(t, xxvc.Warrior)
	f := fieldOf(c, d, fieldGeneralSkills)
	if f.Budget != 40 || f.Max != 30 {
		t.Fatalf("field %+v", f)
	}
	for _, n := range []SkillNo{83, 31, firstMedical, lastMedical, 26} { // the career's, medical, not offered
		if slices.Contains(f.Options, skillName(n)) {
			t.Errorf("%s is on the general list", skillName(n))
		}
	}
	started := 0
	for i := 0; started < int(br1Creation.StartGeneralSkills); i++ {
		n, ok := skillByName(f.Options[i])
		if !ok || !MeetsPrerequisite(&SkillLevels{}, false, n) {
			continue
		}
		d = spend(t, c, d, fieldGeneralSkills, f.Options[i], 1)
		started++
	}
	var next string
	for _, o := range f.Options {
		if n, _ := skillByName(o); MeetsPrerequisite(&SkillLevels{}, false, n) && d.Points[fieldGeneralSkills][slices.Index(f.Options, o)] == 0 {
			next = o
			break
		}
	}
	if refusal(t, c, d, fieldGeneralSkills, next) != rules.SpendNoNewItem {
		t.Errorf("an eighth skill (%s) was started", next)
	}
	more := slices.IndexFunc(d.Points[fieldGeneralSkills], func(p int32) bool { return p > 0 })
	d = spend(t, c, d, fieldGeneralSkills, f.Options[more], 3) // more points on a known skill are fine
	if left := d.PointsLeft(f); left != 40-7-3 {
		t.Errorf("left %d", left)
	}
}

// Programming needs 10 points of Mathematics (CharHlp.PROGRAMMING_NEEDS_MATHEMATICS).
func TestGeneralPrerequisite(t *testing.T) {
	c, d := skillDraft(t, xxvc.Warrior)
	if refusal(t, c, d, fieldGeneralSkills, skillName(Programming)) != rules.SpendWants {
		t.Fatal("programming without mathematics")
	}
	d = spend(t, c, d, fieldGeneralSkills, skillName(Mathematics), 10)
	d = spend(t, c, d, fieldGeneralSkills, skillName(Programming), 1)
	if l := levelsOfDraft(d); l.Level(Programming) != 1 || l.Level(Mathematics) != 10 {
		t.Errorf("levels %d %d", l.Level(Programming), l.Level(Mathematics))
	}
}

// A rogue's favoured skill starts at 10 when its prerequisite is met (StartingSkillLevel) and is given back with
// that 10 when the prerequisite is lost again (ResetDependent's refund is the level above 0).
func TestRogueFavouredDependent(t *testing.T) {
	c, d := skillDraft(t, xxvc.Rogue)
	d = spend(t, c, d, fieldGeneralSkills, skillName(BefriendAnimal), 20)
	if l := levelsOfDraft(d); l.Level(AnimalTraining) != favoredStart {
		t.Fatalf("animal training %d, want the favoured start", l.Level(AnimalTraining))
	}
	o := slices.Index(fieldOf(c, d, fieldGeneralSkills).Options, skillName(BefriendAnimal))
	d, err := c.Spend(d, fieldGeneralSkills, o, -1)
	if err != nil {
		t.Fatal(err)
	}
	if l := levelsOfDraft(d); l.Level(AnimalTraining) != 0 || d.Left[fieldGeneralSkills] != 40-19+favoredStart {
		t.Errorf("animal training %d, left %d", l.Level(AnimalTraining), d.Left[fieldGeneralSkills])
	}
}

// The skill levels the screen leaves are the character's (draftCharacter), the starting levels where none were spent.
func TestSkillPointsReachTheCharacter(t *testing.T) {
	c, d := skillDraft(t, xxvc.Medic)
	d = spend(t, c, d, fieldCareerSkills, skillName(TreatLight), 12)
	d = spend(t, c, d, fieldGeneralSkills, skillName(Mathematics), 4)
	ch, err := c.r.draftCharacter(d, false)
	if err != nil || ch.Skills.Level(TreatLight) != 12 || ch.Skills.Level(Mathematics) != 4 || ch.Skills.Level(RepairElectrical) != 0 {
		t.Fatalf("%v: %d %d", err, ch.Skills.Level(TreatLight), ch.Skills.Level(Mathematics))
	}
	rg := skillDraftCharacter(t, xxvc.Rogue)
	if rg.Skills.Level(8) != favoredStart { // OPEN_LOCK
		t.Errorf("a rogue's open lock starts at %d", rg.Skills.Level(8))
	}
}

func skillDraftCharacter(t *testing.T, career rules.Class) Character {
	c, d := skillDraft(t, career)
	ch, err := c.r.draftCharacter(d, false)
	if err != nil {
		t.Fatal(err)
	}
	return ch
}

// A new race or career begins the screen afresh; a gender change keeps it.
func TestSkillScreenBeginsAfresh(t *testing.T) {
	c, d := skillDraft(t, xxvc.Medic)
	d = spend(t, c, d, fieldCareerSkills, skillName(TreatLight), 5)
	if g := c.Enforce(d, fieldGender, 1, &rules.RNG{State: 9}); g.Points[fieldCareerSkills][slices.Index(fieldOf(c, g, fieldCareerSkills).Options, skillName(TreatLight))] != 5 {
		t.Error("a gender change lost the points")
	}
	e := c.Enforce(d, fieldCareer, slices.Index(pickCareers, xxvc.Warrior), &rules.RNG{State: 9})
	if e.Points != nil || e.Left != nil || e.Sheet.Skills[xxvc.Skills[TreatLight-1]] != 0 {
		t.Errorf("a new career kept the screen: %v %v", e.Points, e.Left)
	}
}

// TestGameCreationNumbers: the skill screen's numbers are the game's xxvc.Creation: a BR2-like one (skills.pas
// NEW_CHARACTER_GENERAL_SKILLS 12, a start level 7, a career's given skills and nothing to spend) lets twelve general skills
// start, not a thirteenth; with no points the screen refuses any spending.
func TestGameCreationNumbers(t *testing.T) {
	c := creation{r: newBR1()}
	c.r.tables.Creation.GeneralPoints, c.r.tables.Creation.StartGeneralSkills = 100, 12
	d := newDraft(t, c, 0, 0, slices.Index(pickCareers, xxvc.Warrior), 5)
	f := fieldOf(c, d, fieldGeneralSkills)
	started := 0
	for i := 0; started < 12; i++ {
		if n, ok := skillByName(f.Options[i]); ok && MeetsPrerequisite(&SkillLevels{}, false, n) {
			d = spend(t, c, d, fieldGeneralSkills, f.Options[i], 1)
			started++
		}
	}
	var next string
	for _, o := range f.Options {
		if n, _ := skillByName(o); MeetsPrerequisite(&SkillLevels{}, false, n) && d.Points[fieldGeneralSkills][slices.Index(f.Options, o)] == 0 {
			next = o
			break
		}
	}
	if refusal(t, c, d, fieldGeneralSkills, next) != rules.SpendNoNewItem {
		t.Errorf("a thirteenth skill (%s) was started", next)
	}
	c.r.tables.Creation = xxvc.Creation{StartLevel: 7, CareerPoints: 0, GeneralPoints: 0, MaxPerSkill: 105,
		InitialSkills:    map[rules.Class][]xxvc.InitialSkill{xxvc.Warrior: {{Skill: "battle-tactics", Value: 50}}},
		ExperienceNeeded: br1Creation.ExperienceNeeded, BaseCombat: br1Creation.BaseCombat}
	if l := startingSkills(c.r.tables.Creation, xxvc.Warrior); l.Level(BattleTacticsNo) != 50 {
		t.Errorf("the career's given skill: %d", l.Level(BattleTacticsNo))
	}
	d = newDraft(t, c, 0, 0, slices.Index(pickCareers, xxvc.Warrior), 5)
	if refusal(t, c, d, fieldCareerSkills, skillName(BattleTacticsNo)) != rules.SpendNoPoints {
		t.Error("a point was spent with none to spend")
	}
}
