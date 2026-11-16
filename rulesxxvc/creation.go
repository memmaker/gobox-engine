package rulesxxvc

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// XXVc's character creation: chargen.pas CreateCharacterMenu (race, gender, career, the stat roll with its training
// from level 0, the name), as rules.CreationRules. Not ported: Skills.EditSkills(80, 40, 30) (the skill point
// screen), ViewChar.SpecializationMenu (a warrior's weapon picks), CustomizeIcon, the age (1d6 + 18: the record keeps
// none), and ModifyCharacterMenu.

// The creation's fields, in the order the original's menus ask them (race, gender, career).
const (
	fieldRace rules.FieldID = iota
	fieldGender
	fieldCareer
	fieldSpecialization // a warrior's weapon (viewchar.pas SpecializationMenu(True)); the choice fields keep their ID as their place
)

// pickGenders are TGender's order (GenderMale, GenderFemale).
var pickGenders = brGenders[:]

// pickCareers are the careers some offered race may take, in the original's career order (Scout: none does).
var pickCareers = func() []rules.Class {
	var out []rules.Class
	for _, c := range xxvc.Careers {
		for _, r := range pickRaces {
			if slices.Contains(raceCareers[r], c) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}()

// Errors of the creation.
var (
	ErrNoRollKind    = errors.New("rulesxxvc: the original rolls its abilities one way only")
	ErrNoMember      = errors.New("rulesxxvc: no such member to modify")
	ErrNoSuchAbility = errors.New("rulesxxvc: not one of the seven abilities")
)

// creation is XXVc's rules.CreationRules; host keeps a new character on the roster and joins it.
type creation struct {
	r    *Rules
	host rules.CreationHost
}

var _ rules.CreationRules = creation{}

func (c creation) Fields(d rules.Draft) []rules.FieldDef {
	races := make([]string, len(pickRaces))
	for i, r := range pickRaces {
		races[i] = xxvc.RaceNames[r]
	}
	genders := make([]string, len(pickGenders))
	for i, g := range pickGenders {
		genders[i] = xxvc.GenderNames[g]
	}
	careers := make([]string, len(pickCareers))
	for i, c := range pickCareers {
		careers[i] = xxvc.CareerNames[c]
	}
	out := []rules.FieldDef{
		{ID: fieldRace, Label: "Race", Options: races},
		{ID: fieldGender, Label: "Gender", Options: genders},
		{ID: fieldCareer, Label: "Class", Options: careers},
	}
	if d.Of != rules.NoChar { // race, gender and career are not changed on a member
		return out
	}
	if opts := c.specializable(); careerOfDraft(d) == xxvc.Warrior && len(opts) > 0 { // CreateCharacterMenu: if career = CareerWarrior
		names := make([]string, len(opts))
		for i, id := range opts {
			names[i] = c.r.itemName(rules.ItemRef{ID: id})
		}
		out = append(out, rules.FieldDef{ID: fieldSpecialization, Label: "Weapon Specialization", Options: names})
	}
	return append(out, skillFields(c.r.tables.Creation, d)...)
}

// specializable are the weapons the specialization menu offers: itemTemplates 1..15 that are not empty (the tables
// list no empty template).
func (c creation) specializable() []rules.ItemID {
	var out []rules.ItemID
	for id := rules.ItemID(1); id <= lastSpecializableTemplate; id++ {
		if _, ok := c.r.tables.Items[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

// Fits: a career is legal when the race may take it (raceCareers); the other fields have no restriction.
func (c creation) Fits(d rules.Draft, f rules.FieldID, o int) bool {
	if d.Of != rules.NoChar { // a member keeps its race, gender and career
		return o == d.Picks[f]
	}
	switch f {
	case fieldRace:
		return o >= 0 && o < len(pickRaces)
	case fieldGender:
		return o >= 0 && o < len(pickGenders)
	case fieldCareer:
		return o >= 0 && o < len(pickCareers) && slices.Contains(raceCareers[pickRaces[d.Picks[fieldRace]]], pickCareers[o])
	case fieldSpecialization:
		return o >= 0 && o < len(c.specializable())
	case fieldCareerSkills, fieldGeneralSkills: // a point may be spent on the option
		_, err := c.Spend(d, f, o, 1)
		return err == nil
	}
	return false
}

// Enforce sets f to o; a race that does not allow the picked career moves it to the race's first career (the class
// menu offers only the race's). A new race or career rolls afresh (the original rolls after the career), a gender
// keeps the scores.
func (c creation) Enforce(d rules.Draft, f rules.FieldID, o int, rng *rules.RNG) rules.Draft {
	if d.Of != rules.NoChar {
		return d
	}
	d.Picks = maps.Clone(d.Picks)
	if d.Picks == nil {
		d.Picks = map[rules.FieldID]int{}
	}
	d.Picks[f] = o
	if f == fieldSpecialization { // the weapon changes nothing else
		return d
	}
	if f != fieldGender {
		delete(d.Picks, fieldSpecialization) // a new race or career asks the weapon afresh
	}
	if !c.Fits(d, fieldCareer, d.Picks[fieldCareer]) {
		d.Picks[fieldCareer] = slices.Index(pickCareers, raceCareers[pickRaces[d.Picks[fieldRace]]][0])
	}
	rng = c.r.stream(rng) // nil: the ruleset's own stream (the engine's seed)
	d.Sheet.Race, d.Sheet.Sex = pickRaces[d.Picks[fieldRace]], pickGenders[d.Picks[fieldGender]]
	career := pickCareers[d.Picks[fieldCareer]]
	d.Sheet.Classes = rules.ClassSet{career}
	if f != fieldGender || d.Sheet.Scores == nil {
		d.Sheet.Age = int32(rules.RNGDice(rng, 1, ageDie) + ageBase) // age := RollDice(1, 6) + 18, before the abilities are rolled
		rollAbilities(c.r.tables.Creation, &d, rng)
		l := c.r.startingSkills(career)
		d.Sheet.Skills = l.skillMap() // the skill screen begins afresh: InitializeAllSkills
		d.Points, d.Left = nil, nil
	}
	return d
}

// rollAbilities is CreateCharacterMenu's roll pass: per ability the best of six 3d6, plus the race's adjustment, at
// least 1 and at least the career's minimum; then the free training from level 0 (the hit points of each level).
func rollAbilities(cr xxvc.Creation, d *rules.Draft, rng *rules.RNG) {
	race, career := d.Sheet.Race, d.Sheet.Classes[0]
	d.Sheet.Scores = map[rules.Ability]int32{}
	for i, a := range brStatOrder {
		best := 0
		for range 6 {
			best = max(best, rules.RNGDice(rng, 3, 6))
		}
		v := max(best+raceStatAdjust[race][i], 1, careerStatMinimum[career][i])
		d.Sheet.Scores[a] = int32(v)
	}
	d.Sheet.Levels = map[rules.Class]int32{career: cr.StartLevel}
	d.Sheet.Exp = cr.Experience(career, cr.StartLevel)
	d.Sheet.HP, d.Sheet.MaxHP = 0, 0
	for level := int32(1); level <= cr.StartLevel; level++ { // CharGen.TrainingMenu until the experience buys no more
		gain := int32(levelHitPoints(cr, career, int(level), int(d.Sheet.Scores[rules.Constitution]), rng))
		d.Sheet.HP, d.Sheet.MaxHP = d.Sheet.HP+gain, d.Sheet.MaxHP+gain
	}
	d.MinHP, d.MaxHP = hpMinimum(career, int(cr.StartLevel), int(d.Sheet.Scores[rules.Constitution])), hpMaximum(career, int(cr.StartLevel), int(d.Sheet.Scores[rules.Constitution]))
}

// levelHitPoints is chargen.pas AddLevelHitPoints for a character reaching level: below the game's fixed gain level the better of two
// career dice plus the constitution bonus (at most 2 unless a warrior), at least 1; from level 9 a fixed 2, 3 for a warrior.
func levelHitPoints(cr xxvc.Creation, career rules.Class, level, con int, rng *rules.RNG) int {
	if level >= int(cr.FixedGainLevel) {
		if career == xxvc.Warrior {
			return 3
		}
		return 2
	}
	sides := 6
	switch career {
	case xxvc.Warrior:
		sides = 10
	case xxvc.Engineer:
		sides = 8
	}
	first := max(rules.RNGDice(rng, 1, sides), rules.RNGDice(rng, 1, sides))
	return max(first+conBonus(career, con), 1)
}

// conBonus is constitutionLevelBonus, at most 2 for a non-warrior.
func conBonus(career rules.Class, con int) int {
	b := constitutionLevelBonus[con-1]
	if career != xxvc.Warrior {
		b = min(b, 2)
	}
	return b
}

// hpMinimum and hpMaximum are chargen.pas HpMinimum and HpMaximum at level: what Modify allows.
func hpMinimum(career rules.Class, level, con int) int32 {
	v := max(constitutionLevelBonus[con-1]+1, 1)
	if v > 3 && career != xxvc.Warrior {
		v = 3
	}
	return int32(v * level)
}

func hpMaximum(career rules.Class, level, con int) int32 {
	v := conBonus(career, con)
	switch career {
	case xxvc.Warrior:
		v += 10
	case xxvc.Scout, xxvc.Engineer:
		v += 8
	default:
		v += 6
	}
	return int32(v * level)
}

// Rolls: the original rolls one way (chargen.pas CreateCharacterMenu has no minimums and no bought abilities), and
// not when it modifies a member.
func (creation) Rolls(d rules.Draft) []rules.RollKind {
	if d.Of != rules.NoChar {
		return nil
	}
	return []rules.RollKind{rules.RollFull}
}

// Roll rolls afresh; the original has no other way (no minimums, no point buy).
func (c creation) Roll(d *rules.Draft, k rules.RollKind, rng *rules.RNG) error {
	if k != rules.RollFull || d.Of != rules.NoChar { // a member modified is not rolled again
		return ErrNoRollKind
	}
	rollAbilities(c.r.tables.Creation, d, c.r.stream(rng))
	return nil
}

// Minimums are the career's lowest abilities (careerStatMinimum).
func (creation) Minimums(d rules.Draft) map[rules.Ability]int32 {
	out := map[rules.Ability]int32{}
	for i, a := range brStatOrder {
		out[a] = int32(careerStatMinimum[d.Sheet.Classes[0]][i])
	}
	return out
}

func (creation) QuickParty() []rules.Template { return nil }
func (creation) QuickRerolls() int            { return 0 }

// SetScore keeps ability a at v as ModifyCharacterMenu's Add and Subtract do: at most its race's adjustment + 18, at
// least + 3 and at least the career's minimum. Constitution brings the hit points into the range it allows.
func (creation) SetScore(d rules.Draft, a rules.Ability, v int32) (rules.Draft, error) {
	if a == rules.StrengthPercent && v == 0 { // the engine sets the exceptional strength beside a typed STR: none here
		return d, nil
	}
	i := slices.Index(brStatOrder[:], a)
	if i < 0 {
		return d, fmt.Errorf("%s: %w", a, ErrNoSuchAbility)
	}
	adj := int32(raceStatAdjust[d.Sheet.Race][i])
	v = min(v, adj+maxStartScore)
	v = max(v, adj+minStartScore, int32(careerStatMinimum[d.Sheet.Classes[0]][i]))
	d.Sheet.Scores = maps.Clone(d.Sheet.Scores)
	d.Sheet.Scores[a] = v
	if a == rules.Constitution {
		d = limitHitPoints(d)
	}
	return d, nil
}

// Make keeps a new character (chargen.pas: Shop.BuildAmmoItems, then SaveChar) on the roster; it is not in the party.
func (c creation) Make(d rules.Draft) (rules.CharID, rules.RosterID, error) {
	if d.Of != rules.NoChar { // Keep of ModifyCharacterMenu
		return d.Of, rules.NoRoster, c.r.modifyMember(d)
	}
	id, err := c.r.insertDraft(d, true)
	if err != nil {
		return rules.NoChar, rules.NoRoster, err
	}
	defer c.r.Dismiss(id)
	if err := c.host.Keep(id); err != nil {
		return rules.NoChar, rules.NoRoster, err
	}
	return rules.NoChar, c.r.chars[id].Roster, nil
}

func (c creation) Join(id rules.RosterID) (rules.CharID, error) { return c.host.Join(id) }

// Draft is the start of a new character: the first race, gender and career the menus offer; of a member, the member
// as it stands (ModifyCharacterMenu).
func (c creation) Draft(of rules.CharID) (rules.Draft, error) {
	if of != rules.NoChar {
		return c.r.draftOf(of)
	}
	return rules.Draft{Picks: map[rules.FieldID]int{fieldRace: 0, fieldGender: 0, fieldCareer: 0}}, nil
}

// insertDraft makes the character d describes and holds it outside the roster order and the purse (a sheet's draft,
// or the character Make keeps); kit: the two free ammunition items.
func (r *Rules) insertDraft(d rules.Draft, kit bool) (rules.CharID, error) {
	c, err := r.draftCharacter(d, kit)
	if err != nil {
		return rules.NoChar, err
	}
	id := r.next
	r.next++
	r.chars[id] = &c
	r.handOut(&c)
	return id, nil
}

// The skills a career starts favoured in (skills.pas StartingSkillLevel; c_skill.pas numbers).
const (
	favoredStart     = 10 // FAVORED_START
	skillPilotRocket = 51
	skillDriveJetcar = 59
)

var rogueFavored = []SkillNo{7, 8, 62, 63, 61, 65, 75, 82, 83, 76} // BYPASS_SECURITY, OPEN_LOCK, MOVE_SILENTLY, PICK_POCKETS, HIDE_IN_SHADOWS, CLIMB, ANIMAL_TRAINING, READ_LIPS, NOTICE, FAST_TALK

// startingSkillLevel is skills.pas StartingSkillLevel: 10 for a skill the career favours when its prerequisites are met.
func startingSkillLevel(l *SkillLevels, career rules.Class, n SkillNo) int8 {
	if !MeetsPrerequisite(l, career == xxvc.Medic, n) {
		return 0
	}
	if career == xxvc.RocketJock && n >= skillPilotRocket && n <= skillDriveJetcar ||
		career == xxvc.Rogue && slices.Contains(rogueFavored, n) {
		return favoredStart
	}
	return 0
}

// draftCharacter is the character a draft makes: chargen.pas CreateCharacterMenu's fixed fields, saves, skills (the
// starting levels), and RecalcStats over the kit.
func (r *Rules) draftCharacter(d rules.Draft, kit bool) (Character, error) {
	if d.Sheet.Classes[0] == "" || d.Sheet.Scores == nil {
		return Character{}, errors.New("rulesxxvc: a draft with no career or abilities")
	}
	race, career := d.Sheet.Race, d.Sheet.Classes[0]
	c := Character{
		Name: d.Name, Sex: d.Sheet.Sex, Race: race, Career: career, Level: int(r.tables.Creation.StartLevel), Age: int(d.Sheet.Age),
		Scores: map[rules.Ability]int{}, HP: int(d.Sheet.HP), MaxHP: int(d.Sheet.MaxHP),
		BaseCombat: int(r.tables.Creation.Combat(career, r.tables.Creation.StartLevel)), IconDimensions: startIconDimensions,
		Coins: startCredits, Enabled: true, Status: StOkay, Side: rules.SideParty, Exp: r.tables.Creation.Experience(career, r.tables.Creation.StartLevel),
		BaseArmorClass: startBaseArmorClass, BaseMovement: raceMovement[race], Movement: raceMovement[race],
		BaseScale: 2, Unarmed: [2]rules.Dice{{N: 1, Sides: 2}},
	}
	for a, v := range d.Sheet.Scores {
		c.Scores[a] = int(v)
	}
	if opts := (creation{r: r}).specializable(); career == xxvc.Warrior && len(opts) > 0 { // SpecializationMenu(True): one training in the picked weapon
		c.Specializations[0] = r.tables.Items[opts[min(max(d.Picks[fieldSpecialization], 0), len(opts)-1)]].Type
	}
	for i := range c.Saves {
		c.Saves[i] = baseSavingThrows[i] - raceSaveAdjust[race][i] // levels 1 and 2 take none off (every third level does)
	}
	if race == xxvc.DesertRunner || career == xxvc.Warrior {
		c.Unarmed[0].Sides = 3
	}
	if race == xxvc.DesertRunner {
		c.Unarmed[0].Bonus = 1
	}
	if career == xxvc.Warrior {
		c.BaseScale = 4
	}
	c.Skills = r.startingSkills(career) // InitializeAllSkills
	if d.Sheet.Skills != nil {          // then what EditSkills left
		c.Skills = levelsOfDraft(d)
	}
	if kit { // the character kept: SetDefaultIcon after the skills, then BuildAmmoItems
		c.Icon = defaultIcon(race, d.Sheet.Sex, career, r.stream(nil))
		for _, a := range freeAmmo { // each at the head of the list
			c.Items = slices.Insert(c.Items, 0, Carried{Ref: rules.ItemRef{ID: a.Item, Count: a.Quantity}, Readied: true})
		}
		for i := range c.Items {
			c.Items[i].Slot = rules.ItemSlot(i + 1)
		}
	}
	r.RecalcStats(&c)
	return c, nil
}

// resolveSheet answers the creation sheet's Actions that a new character needs: the draft as a character outside the
// party, seated after it while the sheet is up. A member modified is drawn in place: HoldCopy keeps it as it was,
// RestoreCopy puts it back.
func (r *Rules) resolveSheet(a rules.Action) (rules.Outcome, bool, error) {
	switch a := a.(type) {
	case rules.SheetDraft:
		if a.Draft.Of != rules.NoChar { // the member modified, drawn with the draft's name, abilities and hit points
			return rules.Outcome{Created: a.Draft.Of}, true, r.modifyMember(a.Draft)
		}
		id, err := r.insertDraft(a.Draft, false)
		return rules.Outcome{Created: id}, true, err
	case rules.HoldCopy:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoMember
		}
		if r.held == nil {
			r.held = map[rules.CharID]Character{}
		}
		r.held[a.Who] = c.clone()
	case rules.RestoreCopy:
		c, was := r.chars[a.Who], r.held[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoMember
		}
		*c = was
		delete(r.held, a.Who)
	case rules.ReplaceMember:
		if r.chars[a.Who] == nil || a.At < 0 || int(a.At) >= len(r.party) {
			return rules.Outcome{}, true, errNoMember
		}
		r.party[a.At] = a.Who
	case rules.ApplyItems: // the sheet draws a character's figures fresh (00BA:0BB8): RecalcStats
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoMember
		}
		r.RecalcStats(c)
	case rules.SeatTemp:
		if r.chars[a.Who] == nil {
			return rules.Outcome{}, true, errNoMember
		}
		r.party = append(r.party, a.Who)
	case rules.TrimParty:
		r.party = r.party[:a.Size]
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}
