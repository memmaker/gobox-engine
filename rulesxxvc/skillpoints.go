package rulesxxvc

import (
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// The skill point screen of a new character (skills.pas EditSkills(80, 40, 30), chargen.pas CreateCharacterMenu):
// the career's eight skills take 80 points, the general skills 40, none more than 30 this time. It is two
// FieldSkills fields of the creation. The draft keeps the levels in Sheet.Skills, the points put on each option
// since the screen began in Points (a level less the level the screen began with: oldLevels) and the points still
// to spend in Left. The original asks the career points first, then the general ones; here both are open at once.

const (
	careerSkillPoints  = 80 // EditSkills's careerPoints
	generalSkillPoints = 40 // EditSkills's generalPoints
	maxPointsPerSkill  = 30 // EditSkills's maxPerSkill
	maxGeneralLevel    = 80 // skills.pas MAX_GENERAL_LEVEL: a general skill stops there
	startGeneralSkills = 7  // maxGeneral at level 2: no more general skills known than that
)

// The two skill fields, after the three choices.
const (
	fieldCareerSkills rules.FieldID = iota + fieldSpecialization + 1
	fieldGeneralSkills
)

// The words of the refusals (skills.pas EditSkills's PrintBottomLine).
var (
	refuseNoCareerPoints  = &rules.SpendRefusal{Reason: rules.SpendNoPoints, Text: "No career skill points left"}
	refuseNoGeneralPoints = &rules.SpendRefusal{Reason: rules.SpendNoPoints, Text: "No general skill points left"}
	refuseWants           = &rules.SpendRefusal{Reason: rules.SpendWants, Text: "Prerequisites not satisfied"}
	refuseAtMostThisLevel = &rules.SpendRefusal{Reason: rules.SpendAtMost, Text: "Skill at maximum for this level"}
	refuseAtMostGeneral   = &rules.SpendRefusal{Reason: rules.SpendAtMost, Text: "Skill at maximum for general skills"}
	refuseNoNewSkill      = &rules.SpendRefusal{Reason: rules.SpendNoNewItem, Text: "No more new skills allowed this level"}
	refuseNoOption        = &rules.SpendRefusal{Reason: rules.SpendNoOption, Text: "No such skill or step"}
)

// groupNames are gdconsts.pas skillGroupNames, by the group's heading on the screen.
var groupNames = map[SkillGroup]string{
	GroupCareer: "Career Skills", GroupDexterity: "Dexterity Skills", GroupConstitution: "Constitution Skills",
	GroupIntelligence: "Intelligence Skills", GroupWisdom: "Wisdom Skills", GroupCharisma: "Charisma Skills",
	GroupTechnical: "technical Skills",
}

// careerOfDraft is the career the draft's picks name.
func careerOfDraft(d rules.Draft) rules.Class { return pickCareers[d.Picks[fieldCareer]] }

// skillLists are the skills of the two fields for a career (skills.pas BuildSkillLists): its eight, then every
// other skill the general list offers except the medical ones, by number.
func skillLists(career rules.Class) (careerList, general []SkillNo) {
	if i := slices.Index(xxvc.Careers, career); i >= 0 {
		careerList = CareerSkills[i][:]
	}
	for n := SkillNo(1); n <= SkillCount; n++ {
		if !slices.Contains(careerList, n) && (n < firstMedical || n > lastMedical) && Offered(n) {
			general = append(general, n)
		}
	}
	return careerList, general
}

func skillName(n SkillNo) string { return xxvc.SkillNames[xxvc.Skills[n-1]] }

// levelsOfDraft is the draft's skill levels as the rules count them.
func levelsOfDraft(d rules.Draft) SkillLevels {
	var l SkillLevels
	for n := SkillNo(1); n <= SkillCount; n++ {
		l.Set(n, int8(d.Sheet.Skills[xxvc.Skills[n-1]]))
	}
	return l
}

// skillMap is the levels as a draft keeps them: the skills above 0.
func (l *SkillLevels) skillMap() map[rules.Skill]int32 {
	m := map[rules.Skill]int32{}
	for n := SkillNo(1); n <= SkillCount; n++ {
		if v := l.Level(n); v != 0 {
			m[xxvc.Skills[n-1]] = int32(v)
		}
	}
	return m
}

// startingSkills is skills.pas InitializeAllSkills: each skill at its starting level, in order, a later skill's
// prerequisite seeing the earlier.
func startingSkills(career rules.Class) SkillLevels {
	var l SkillLevels
	for n := SkillNo(1); n <= SkillCount; n++ {
		l.Set(n, startingSkillLevel(&l, career, n))
	}
	return l
}

// skillFields are the two skill fields of draft d (EditSkills's career then general screens).
func skillFields(d rules.Draft) []rules.FieldDef {
	career := careerOfDraft(d)
	l := levelsOfDraft(d)
	if d.Sheet.Skills == nil {
		l = startingSkills(career)
	}
	careerList, general := skillLists(career)
	field := func(id rules.FieldID, label string, list []SkillNo, budget int32) rules.FieldDef {
		f := rules.FieldDef{ID: id, Kind: rules.FieldSkills, Label: label, Budget: budget, Max: maxPointsPerSkill}
		for _, n := range list {
			group := GroupCareer
			if id == fieldGeneralSkills {
				group = Group(n)
			}
			f.Options = append(f.Options, skillName(n))
			f.Rows = append(f.Rows, skillRow(&l, d, career, n, groupNames[group]))
		}
		return f
	}
	return []rules.FieldDef{
		field(fieldCareerSkills, "Career skill points", careerList, careerSkillPoints),
		field(fieldGeneralSkills, "General skill points", general, generalSkillPoints),
	}
}

// skillRow is skill n as the skill screen shows it (DrawSkillRow): the level, the bonus of its group's ability when
// the prerequisite is met, the total, and the prerequisites still unmet (PrintPrerequisite).
func skillRow(l *SkillLevels, d rules.Draft, career rules.Class, n SkillNo, group string) rules.FieldRow {
	met := MeetsPrerequisite(l, career == xxvc.Medic, n)
	row := rules.FieldRow{Group: group, Value: int32(l.Level(n)), Open: met}
	if met {
		if a, ok := groupAbility[Group(n)]; ok {
			row.Bonus = d.Sheet.Scores[a]
		}
	}
	if row.Value > 0 {
		row.Total = row.Value + row.Bonus
	}
	if p, ok := prerequisites[n]; ok && !p.anyOf {
		for _, need := range p.needs {
			if int32(l.Level(need.Skill)) < int32(need.Level) {
				row.Wants = append(row.Wants, rules.Want{Name: skillName(need.Skill), Level: int32(need.Level)})
			}
		}
	}
	return row
}

// dependent is a skill whose prerequisite is a level of another: it is reset when that level is reached or lost
// (skills.pas AdjustPrerequisites). Composition is left out: it is not offered, so it stays at 0 and its reset
// changes nothing.
type dependent struct {
	skill, on SkillNo
	level     int8
}

var dependents = func() map[SkillNo][]dependent {
	m := map[SkillNo][]dependent{}
	for s, p := range prerequisites {
		if p.anyOf {
			continue
		}
		for _, n := range p.needs {
			m[n.Skill] = append(m[n.Skill], dependent{skill: s, on: n.Skill, level: n.Level})
		}
	}
	return m
}()

// skillEdit is the skill screen's state while a point is spent: copies of the draft's levels, points and left.
type skillEdit struct {
	d      rules.Draft
	career rules.Class
	levels SkillLevels
	lists  map[rules.FieldID][]SkillNo
	field  rules.FieldID // the group on the screen: where a refund goes
}

func newSkillEdit(d rules.Draft, f rules.FieldID) *skillEdit {
	e := &skillEdit{d: d, career: careerOfDraft(d), field: f}
	if d.Sheet.Skills == nil {
		e.levels = startingSkills(e.career)
	} else {
		e.levels = levelsOfDraft(d)
	}
	cl, gl := skillLists(e.career)
	e.lists = map[rules.FieldID][]SkillNo{fieldCareerSkills: cl, fieldGeneralSkills: gl}
	e.d.Points = map[rules.FieldID][]int32{}
	for id, list := range e.lists {
		e.d.Points[id] = make([]int32, len(list))
		copy(e.d.Points[id], d.Points[id])
	}
	e.d.Left = map[rules.FieldID]int32{}
	for _, f := range skillFields(d) {
		e.d.Left[f.ID] = d.PointsLeft(f)
	}
	return e
}

// where is the field and option of skill n (false: on no list).
func (e *skillEdit) where(n SkillNo) (rules.FieldID, int, bool) {
	for id, list := range e.lists {
		if i := slices.Index(list, n); i >= 0 {
			return id, i, true
		}
	}
	return 0, 0, false
}

// started is the level skill n had when the screen began, or when a reset gave it its base (oldLevels).
func (e *skillEdit) started(n SkillNo) int8 {
	if id, i, ok := e.where(n); ok {
		return e.levels.Level(n) - int8(e.d.Points[id][i])
	}
	return e.levels.Level(n)
}

// knownGeneral is knownGeneralCount: the general skills with a level.
func (e *skillEdit) knownGeneral() int {
	known := 0
	for _, n := range e.lists[fieldGeneralSkills] {
		if e.levels.Level(n) > 0 {
			known++
		}
	}
	return known
}

// reset is ResetDependent: when changed has just reached the level dep needs, dep goes back to its base; when it
// has just fallen below, dep's levels above its base are given back and dep is 0.
func (e *skillEdit) reset(changed SkillNo, dep dependent, increasing bool) {
	base := startingSkillLevel(&e.levels, e.career, dep.skill)
	id, i, listed := e.where(dep.skill)
	switch {
	case increasing && e.levels.Level(changed) == dep.level:
		e.levels.Set(dep.skill, base)
	case !increasing && e.levels.Level(changed) == dep.level-1:
		if refund := e.levels.Level(dep.skill) - base; refund > 0 {
			e.d.Left[e.field] += int32(refund)
		}
		e.levels.Set(dep.skill, 0)
	default:
		return
	}
	if listed {
		e.d.Points[id][i] = 0 // its start is its level now
	}
}

// Spend is EditSkills's Add (by 1) and Subtract (by -1) on option o of field f.
func (c creation) Spend(d rules.Draft, f rules.FieldID, o int, by int32) (rules.Draft, error) {
	if f != fieldCareerSkills && f != fieldGeneralSkills || by != 1 && by != -1 {
		return d, refuseNoOption
	}
	e := newSkillEdit(d, f)
	list := e.lists[f]
	if o < 0 || o >= len(list) {
		return d, refuseNoOption
	}
	n := list[o]
	if by > 0 {
		if err := e.add(f, n); err != nil {
			return d, err
		}
		e.d.Points[f][o]++
	} else {
		if e.d.Points[f][o] <= 0 { // at the level the screen began with: nothing to take back
			return d, nil
		}
		e.d.Points[f][o]--
		e.d.Left[f]++
		e.levels.Set(n, e.levels.Level(n)-1)
	}
	for _, dep := range dependents[n] {
		e.reset(n, dep, by > 0)
	}
	e.d.Sheet.Skills = e.levels.skillMap()
	return e.d, nil
}

// add checks Add's refusals in the order EditSkills makes them and takes the point.
func (e *skillEdit) add(f rules.FieldID, n SkillNo) error {
	level := e.levels.Level(n)
	meets := MeetsPrerequisite(&e.levels, e.career == xxvc.Medic, n)
	over := int(level) >= int(e.started(n))+maxPointsPerSkill
	if f == fieldCareerSkills {
		switch {
		case e.d.Left[f] <= 0:
			return refuseNoCareerPoints
		case !meets:
			return refuseWants
		case over:
			return refuseAtMostThisLevel
		}
	} else {
		switch {
		case e.d.Left[f] <= 0:
			return refuseNoGeneralPoints
		case level >= maxGeneralLevel:
			return refuseAtMostGeneral
		case over:
			return refuseAtMostThisLevel
		case !meets:
			return refuseWants
		case level == startingSkillLevel(&e.levels, e.career, n) && e.knownGeneral() >= startGeneralSkills:
			return refuseNoNewSkill
		}
	}
	e.d.Left[f]--
	e.levels.Set(n, level+1)
	return nil
}
