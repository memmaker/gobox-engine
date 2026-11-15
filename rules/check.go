package rules

// Catalog is what a ruleset handles.
type Catalog struct {
	Facts      []Fact
	Spells     []SpellID
	Conditions []Condition
	Traits     []Trait
	Races      []Race
	Alignments []Alignment
	Classes    []Class
	Skills     []Skill
	Saves      []Save
	Features   []Feature
	Monsters   []MonsterID
	Items      []ItemID
	Ships      []ShipID
}
