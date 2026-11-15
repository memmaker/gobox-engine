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
	ClassSets  []ClassSet // the class combinations a character can have, in the ruleset's own order (1e: the record's class byte; XXVc: one career each)
	Skills     []Skill
	Saves      []Save
	Features   []Feature
	Monsters   []MonsterID
	Items      []ItemID
	Ships      []ShipID
}
