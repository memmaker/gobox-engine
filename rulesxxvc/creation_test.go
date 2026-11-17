package rulesxxvc

import (
	"errors"
	"maps"
	"slices"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// newDraft is a draft of race, career picks (indexes into pickRaces, pickCareers) rolled from seed.
func newDraft(t *testing.T, c creation, race, gender, career int, seed uint32) rules.Draft {
	t.Helper()
	rng := &rules.RNG{State: seed}
	d, _ := c.Draft(rules.NoChar)
	for _, p := range []struct {
		f rules.FieldID
		o int
	}{{fieldRace, race}, {fieldGender, gender}, {fieldCareer, career}} {
		d = c.Enforce(d, p.f, p.o, rng)
	}
	return d
}

// TestCreationRules: chargen.pas CreateCharacterMenu's rules over many rolls: every ability meets the career's minimum
// and is at least 1; the hit points of two levels stay in HpMinimum..HpMaximum; a race that cannot take the career moves it.
func TestCreationRules(t *testing.T) {
	c := creation{r: newBR1()}
	tinker := newDraft(t, c, 4, 0, 4, 1) // a Tinker asks for a Rogue (Terran's list has it, Tinker's has not)
	if got := tinker.Sheet.Classes[0]; got != xxvc.Medic {
		t.Fatalf("tinker rogue: %s, want the race's first career, medic", got)
	}
	if c.Fits(tinker, fieldCareer, 2) { // warrior
		t.Error("a tinker warrior fits")
	}
	for seed := uint32(1); seed <= 300; seed++ {
		for race := range pickRaces {
			d := newDraft(t, c, race, 1, 0, seed)
			career := d.Sheet.Classes[0]
			for i, a := range brStatOrder {
				if v := int(d.Sheet.Scores[a]); v < 1 || v < careerStatMinimum[career][i] {
					t.Fatalf("seed %d %s %s: %s %d under the minimum %d", seed, pickRaces[race], career, a, v, careerStatMinimum[career][i])
				}
			}
			con := int(d.Sheet.Scores[rules.Constitution])
			if d.Sheet.MaxHP < hpMinimum(career, int(br1Creation.StartLevel), con) || d.Sheet.MaxHP > hpMaximum(career, int(br1Creation.StartLevel), con) {
				t.Fatalf("seed %d %s %s: hit points %d outside %d..%d", seed, pickRaces[race], career, d.Sheet.MaxHP, hpMinimum(career, int(br1Creation.StartLevel), con), hpMaximum(career, int(br1Creation.StartLevel), con))
			}
		}
	}
	if err := c.Roll(&tinker, rules.RollMinimal, nil); !errors.Is(err, ErrNoRollKind) {
		t.Errorf("a minimal roll: %v", err)
	}
	if d, err := c.SetScore(tinker, rules.Strength, 17); err != nil || d.Sheet.Scores[rules.Strength] != 16 { // Tinker STR -2: at most 16 (Modify's Add stops there)
		t.Errorf("STR 17: %v", err)
	}
	if d, err := c.SetScore(tinker, rules.Strength, 16); err != nil || d.Sheet.Scores[rules.Strength] != 16 {
		t.Errorf("STR 16: %v", err)
	}
}

// TestDraftCharacter: the fields the original sets on a new character: level 2 with the career's experience, saves
// from the base less the race's adjustment, unarmed dice, movement, 1000 credits, the favoured skills at 10.
func TestDraftCharacter(t *testing.T) {
	r := newBR1()
	c := creation{r: r}
	d := newDraft(t, c, 5, 0, 2, 3) // Desert Runner, male, Warrior
	d.Name = "Rex"
	ch, err := r.draftCharacter(d, false)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Level != 2 || ch.Exp != 2000 || ch.BaseCombat != 41 || ch.Coins != 1000 || ch.BaseMovement != 15 || ch.BaseScale != 4 ||
		ch.Unarmed[0] != (rules.Dice{N: 1, Sides: 3, Bonus: 1}) || ch.Saves != [8]int{13, 14, 13, 12, 11, 14, 16, 9} || !ch.Enabled || ch.Status != StOkay {
		t.Errorf("warrior: %+v", ch)
	}
	rj, err := r.draftCharacter(newDraft(t, c, 0, 0, 0, 3), false) // Terran Rocket Jock
	if err != nil {
		t.Fatal(err)
	}
	if rj.Skills.Level(skillPilotRocket) != favoredStart || rj.Skills.Level(skillDriveJetcar) != favoredStart {
		t.Errorf("rocket jock skills: pilot rocket %d", rj.Skills.Level(skillPilotRocket))
	}
}

// TestRollsOffered: the rolls the ruleset offers are the ones Roll takes: the original (CreateCharacterMenu) rolls one way.
func TestRollsOffered(t *testing.T) {
	c := creation{r: newBR1()}
	d := newDraft(t, c, 0, 0, 0, 7)
	offered := c.Rolls(d)
	for _, k := range []rules.RollKind{rules.RollFull, rules.RollMinimal, rules.RollPointBuy} {
		err := c.Roll(&d, k, &rules.RNG{State: 3})
		if slices.Contains(offered, k) != (err == nil) {
			t.Errorf("kind %d: offered %v, error %v", k, slices.Contains(offered, k), err)
		}
	}
	if !slices.Equal(offered, []rules.RollKind{rules.RollFull}) {
		t.Errorf("offered %v", offered)
	}
}

// TestCreationAge: chargen.pas age := RollDice(1, 6) + 18, rolled with the career (a gender change keeps it), kept on
// the character and the sheet.
func TestCreationAge(t *testing.T) {
	c := creation{r: newBR1()}
	seen := map[int32]bool{}
	for seed := uint32(1); seed <= 400; seed++ {
		d := newDraft(t, c, 0, 0, 0, seed)
		if d.Sheet.Age < 19 || d.Sheet.Age > 24 {
			t.Fatalf("seed %d: age %d, want 1d6+18", seed, d.Sheet.Age)
		}
		seen[d.Sheet.Age] = true
		if got := c.Enforce(d, fieldGender, 1, &rules.RNG{State: seed + 1000}).Sheet.Age; got != d.Sheet.Age {
			t.Fatalf("a gender change rolled the age again: %d to %d", d.Sheet.Age, got)
		}
		ch, err := c.r.draftCharacter(d, false)
		if err != nil || int32(ch.Age) != d.Sheet.Age {
			t.Fatalf("character age %d, draft %d (%v)", ch.Age, d.Sheet.Age, err)
		}
	}
	if len(seen) != 6 {
		t.Errorf("ages seen %v, want all of 19..24", seen)
	}
}

// specTables is a Rules with three item templates: two within itemTemplates 1..15, one past it.
func specTables() *Rules {
	r := newBR1()
	r.tables.ItemTypes = make([]xxvc.ItemType, 12)
	r.tables.ItemWords = []string{"", "Dagger", "Sword", "Laser"}
	r.tables.Items = map[rules.ItemID]xxvc.ItemTemplate{
		1: {Type: 5, Words: [3]int32{1}}, 4: {Type: 7, Words: [3]int32{2}}, 16: {Type: 8, Words: [3]int32{3}},
	}
	return r
}

// TestWarriorSpecialization: viewchar.pas SpecializationMenu(True) in CreateCharacterMenu: a warrior (only) picks one
// of the weapons itemTemplates 1..15 hold and the character is trained once in it.
func TestWarriorSpecialization(t *testing.T) {
	c := creation{r: specTables()}
	warrior := newDraft(t, c, 0, 0, slices.Index(pickCareers, xxvc.Warrior), 3)
	var spec rules.FieldDef
	for _, f := range c.Fields(warrior) {
		if f.ID == fieldSpecialization {
			spec = f
		}
	}
	if !slices.Equal(spec.Options, []string{"Dagger", "Sword"}) {
		t.Fatalf("options %v: itemTemplates 1..15 only", spec.Options)
	}
	medic := newDraft(t, c, 0, 0, slices.Index(pickCareers, xxvc.Medic), 3)
	for _, f := range c.Fields(medic) {
		if f.ID == fieldSpecialization {
			t.Fatal("a medic is asked a weapon")
		}
	}
	scores := warrior.Sheet.Scores
	w := c.Enforce(warrior, fieldSpecialization, 1, &rules.RNG{State: 11})
	if !maps.Equal(w.Sheet.Scores, scores) || w.Sheet.Age != warrior.Sheet.Age || w.Picks[fieldSpecialization] != 1 {
		t.Errorf("the weapon changed more than itself: %v", w.Sheet.Scores)
	}
	ch, err := c.r.draftCharacter(w, false)
	if err != nil || ch.Specialized(7) != 1 || ch.Specialized(5) != 0 {
		t.Fatalf("specializations %v (%v)", ch.Specializations, err)
	}
	if g := c.Enforce(w, fieldGender, 1, &rules.RNG{State: 12}); g.Picks[fieldSpecialization] != 1 {
		t.Error("a gender change lost the weapon")
	}
	if r := c.Enforce(w, fieldCareer, slices.Index(pickCareers, xxvc.Engineer), &rules.RNG{State: 12}); len(r.Picks) != 3 {
		t.Errorf("a new career kept the weapon: %v", r.Picks)
	}
	if ch, _ := c.r.draftCharacter(medic, false); ch.Specialized(5)+ch.Specialized(7) != 0 {
		t.Error("a medic is trained in a weapon")
	}
}

// TestDefaultIcon: chargen.pas SetDefaultIcon: 17 for a Tinker, one of 16, 23, 24, 31 for a Desert Runner, else
// characterIcons[gender, career]; a character kept has it (not the draft drawn on the sheet: no dice for it).
func TestDefaultIcon(t *testing.T) {
	rng := &rules.RNG{State: 5}
	for sex, row := range map[rules.Sex]map[rules.Class]int{
		rules.Male: {xxvc.RocketJock: 32, xxvc.Medic: 0, xxvc.Warrior: 3, xxvc.Engineer: 7, xxvc.Rogue: 21}, rules.Female: {xxvc.RocketJock: 33, xxvc.Medic: 1, xxvc.Warrior: 6, xxvc.Engineer: 15, xxvc.Rogue: 25}} {
		for career, want := range row {
			if got := defaultIcon(xxvc.Terran, sex, career, rng); got != want {
				t.Errorf("%s %s: icon %d, want %d", sex, career, got, want)
			}
		}
	}
	if defaultIcon(xxvc.Tinker, rules.Male, xxvc.Medic, rng) != 17 {
		t.Error("a tinker's icon")
	}
	seen := map[int]bool{}
	for range 200 {
		seen[defaultIcon(xxvc.DesertRunner, rules.Female, xxvc.Warrior, rng)] = true
	}
	if len(seen) != 4 || !seen[16] || !seen[23] || !seen[24] || !seen[31] {
		t.Errorf("a desert runner's icons %v", seen)
	}
	c := creation{r: newBR1()}
	d := newDraft(t, c, 0, 0, slices.Index(pickCareers, xxvc.Warrior), 2)
	sheet, _ := c.r.draftCharacter(d, false)
	kept, _ := c.r.draftCharacter(d, true)
	if sheet.Icon != 3 || kept.Icon != 3 { // a draft naming none: SetDefaultIcon's, on the sheet as kept
		t.Errorf("icons: sheet %d, kept %d", sheet.Icon, kept.Icon)
	}
}
