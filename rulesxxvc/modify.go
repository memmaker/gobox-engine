package rulesxxvc

import (
	"maps"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// ModifyCharacterMenu (chargen.pas): a member whose experience is still the career's starting experience
// (experienceNeeded[career, the start level]) may have its abilities, hit points and name changed; each ability stays between
// its race's adjustment + 3 and + 18 and at least the career's minimum, the hit points between HpMinimum and
// HpMaximum. Race, gender, career and skills do not change. Keep sets the hit points to their maximum.

// modifiable is "can't be modified" turned round: the member's experience is the starting one.
func (r *Rules) modifiable(c *Character) bool {
	i := slices.Index(xxvc.Careers, c.Career)
	return i >= 0 && c.Exp == r.tables.Creation.Experience(c.Career, r.tables.Creation.StartLevel)
}

// clone is c with its own maps and lists, for HoldCopy.
func (c Character) clone() Character {
	c.Scores = maps.Clone(c.Scores)
	c.Items = slices.Clone(c.Items)
	c.Stashed = slices.Clone(c.Stashed)
	c.Effects = slices.Clone(c.Effects)
	return c
}

// draftOf is member id as a draft: its picks, its scores, its hit points and the range they may take.
func (r *Rules) draftOf(id rules.CharID) (rules.Draft, error) {
	c := r.chars[id]
	if c == nil {
		return rules.Draft{}, ErrNoMember
	}
	d := rules.Draft{Of: id, Name: c.Name, Picks: map[rules.FieldID]int{
		fieldRace:   slices.Index(pickRaces, c.Race),
		fieldGender: slices.Index(pickGenders, c.Sex),
		fieldCareer: slices.Index(pickCareers, c.Career),
	}}
	d.Sheet = r.Sheet(id)
	d.Sheet.Sex = c.Sex
	d.MinHP = hpMinimum(c.Career, c.Level, c.Scores[rules.Constitution])
	d.MaxHP = hpMaximum(c.Career, c.Level, c.Scores[rules.Constitution])
	return d, nil
}

// limitHitPoints brings d's maximum hit points into HpMinimum..HpMaximum for its constitution and sets the hit
// points to it (Modify's Add on constitution).
func limitHitPoints(d rules.Draft) rules.Draft {
	career, con := d.Sheet.Classes[0], int(d.Sheet.Scores[rules.Constitution])
	level := int(d.Sheet.Levels[career])
	d.MinHP, d.MaxHP = hpMinimum(career, level, con), hpMaximum(career, level, con)
	d.Sheet.MaxHP = min(max(d.Sheet.MaxHP, d.MinHP), d.MaxHP)
	d.Sheet.HP = d.Sheet.MaxHP
	return d
}

// modifyMember puts draft d's name, abilities and hit points on the member it modifies (d.Of): Keep.
func (r *Rules) modifyMember(d rules.Draft) error {
	c := r.chars[d.Of]
	if c == nil {
		return ErrNoMember
	}
	c.Name = d.Name
	for _, a := range brStatOrder {
		c.Scores[a] = int(d.Sheet.Scores[a])
	}
	c.MaxHP = int(d.Sheet.MaxHP)
	c.HP = c.MaxHP
	r.RecalcStats(c)
	return nil
}
