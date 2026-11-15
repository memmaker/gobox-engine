package rulesxxvc

import (
	"fmt"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// Builder is the XXVc ruleset for the shared engine (rules.Builder): its tables
// are read from the source (rules.SourceRuleTables, the import's tables file) and
// handed to Begin with the default options. The adventure has no tables
// without its import: that panics, as a broken data folder does (run the import again).
func Builder(src rules.Source, _ rules.CreationHost) (rules.Rules, rules.CreationRules) {
	var t xxvc.Tables
	if !src.Read(rules.SourceRuleTables, &t) {
		panic("rulesxxvc: the adventure has no tables file (run the import again)")
	}
	r := New()
	r.Begin(rules.Campaign{Tables: t}, rules.Options{})
	return r, creation{}
}

// creation is XXVc's creation: not ported (BR1's is in the old engine). Each
// method that can say so answers ErrNotPorted, the others their zero value, so
// that the engine reads an empty sheet and no quick party and never panics.
type creation struct{}

var _ rules.CreationRules = creation{}

func (creation) Fields(rules.Draft) []rules.FieldDef                                     { return nil }
func (creation) Fits(rules.Draft, rules.FieldID, int) bool                               { return false }
func (creation) Enforce(d rules.Draft, _ rules.FieldID, _ int, _ *rules.RNG) rules.Draft { return d }
func (creation) Roll(*rules.Draft, rules.RollKind, *rules.RNG) error                     { return ErrNotPorted }
func (creation) Minimums(rules.Draft) map[rules.Ability]int32                            { return nil }
func (creation) QuickParty() []rules.Template                                            { return nil }
func (creation) QuickRerolls() int                                                       { return 0 }
func (creation) SetScore(d rules.Draft, _ rules.Ability, _ int32) (rules.Draft, error) {
	return d, ErrNotPorted
}
func (creation) Make(rules.Draft) (rules.CharID, rules.RosterID, error) {
	return rules.NoChar, rules.NoRoster, fmt.Errorf("Make: %w", ErrNotPorted)
}
func (creation) Join(rules.RosterID) (rules.CharID, error) {
	return rules.NoChar, fmt.Errorf("Join: %w", ErrNotPorted)
}
func (creation) Draft(rules.CharID) (rules.Draft, error) {
	return rules.Draft{}, fmt.Errorf("Draft: %w", ErrNotPorted)
}
