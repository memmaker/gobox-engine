package core

import (
	"fmt"
	"slices"
)

// Catalog is what a ruleset handles.
type Catalog struct {
	Facts      []Fact
	Spells     []SpellID
	Conditions []Condition
	Traits     []Trait
	Races      []Race
	Classes    []Class
	Skills     []Skill
	Saves      []Save
	Features   []Feature
}

// Mismatches is every need the catalog lacks, as text for the load error: the
// engine refuses to start the adventure on this ruleset if there is any.
func Mismatches(n Needs, c Catalog) []string {
	var out []string
	out = append(out, lacking("fact", n.Facts, c.Facts)...)
	out = append(out, lacking("spell", n.Spells, c.Spells)...)
	out = append(out, lacking("condition", n.Conditions, c.Conditions)...)
	out = append(out, lacking("trait", n.Traits, c.Traits)...)
	out = append(out, lacking("race", n.Races, c.Races)...)
	out = append(out, lacking("class", n.Classes, c.Classes)...)
	out = append(out, lacking("skill", n.Skills, c.Skills)...)
	out = append(out, lacking("save", n.Saves, c.Saves)...)
	out = append(out, lacking("feature", n.Features, c.Features)...)
	return out
}

func lacking[T comparable](kind string, need, have []T) []string {
	var out []string
	for _, x := range need {
		if !slices.Contains(have, x) {
			out = append(out, fmt.Sprintf("the adventure needs %s %v, the ruleset has none", kind, x))
		}
	}
	return out
}
