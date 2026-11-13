package rules1e

import "gobox/rules"

// Career is a character's class or multi-class, POR's C_Class number: saves hold it.
type Career int

// Multi-classes stand past the eight single classes.
const (
	CareerCleric Career = iota
	CareerDruid
	CareerFighter
	CareerPaladin
	CareerRanger
	CareerMage
	CareerThief
	CareerMonk
)

// Ability is one of the six scores, in the order of Character.Stats and
// the rows of the creation tables.
type Ability int

const (
	Strength Ability = iota
	Intelligence
	Wisdom
	Dexterity
	Constitution
	Charisma
)

// raceLimit is a race's limits. Str00Max caps exceptional strength for fighters
// with 18. Strength limits used to differ by gender; everyone gets the men's.
type raceLimit struct {
	StrMin, StrMax int16
	Str00Max       int16
	Min, Max       [5]int16 // intelligence, wisdom, dexterity, constitution, charisma
}

// ageDice is a starting age: Base + Count dice of Sides.
type ageDice struct {
	Base         int32
	Count, Sides int16
}

// AbilityRange is race's lowest and highest ability a.
func (t *Tables) AbilityRange(race rules.Race, a Ability) (lo, hi int) {
	r := t.RaceLimits[raceRow[race]]
	if a == Strength {
		return int(r.StrMin), int(r.StrMax)
	}
	return int(r.Min[a-1]), int(r.Max[a-1])
}

// ExceptionalStrengthCap is the most percent race's fighter has at Str 18.
func (t *Tables) ExceptionalStrengthCap(race rules.Race) int {
	return int(t.RaceLimits[raceRow[race]].Str00Max)
}

// RaceCareers lists what race may be, in table order.
func (t *Tables) RaceCareers(race rules.Race) []Career {
	row := t.RaceClasses[raceRow[race]]
	out := make([]Career, row[0])
	for i := range out {
		out[i] = Career(row[i+1])
	}
	return out
}

// AlignmentNumber is C_Align's number of an alignment: saves hold it.
type AlignmentNumber int

// CareerAlignments lists the alignments career allows, in table order.
func (t *Tables) CareerAlignments(c Career) []AlignmentNumber {
	row := t.ClassAlignments[c]
	out := make([]AlignmentNumber, row[0])
	for i := range out {
		out[i] = AlignmentNumber(row[i+1])
	}
	return out
}

// CareerMinimum is the least ability a career takes.
func (t *Tables) CareerMinimum(c Career, a Ability) int { return int(t.ClassMinStats[c][a]) }

// StartingAge is the base and dice of race's starting age for a single class.
func (t *Tables) StartingAge(race rules.Race, c Career) (base, count, sides int) {
	a := t.StartAge[raceRow[race]][c]
	return int(a.Base), int(a.Count), int(a.Sides)
}

// AgeStep is the age race reaches its category step (1..5) at.
func (t *Tables) AgeStep(race rules.Race, step int) int {
	return int(t.AgeSteps[raceRow[race]][step-1])
}
