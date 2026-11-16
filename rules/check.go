package rules

import "strings"

// Catalog is what a ruleset handles.
type Catalog struct {
	Facts        []Fact
	Spells       []SpellID
	Conditions   []Condition
	Traits       []Trait
	Races        []Race
	Alignments   []Alignment
	Classes      []Class
	ClassSets    []ClassSet // the class combinations a character can have, in the ruleset's own order (1e: the record's class byte; XXVc: one career each)
	Skills       []Skill
	Abilities    []Ability               // the abilities a sheet lists, in the ruleset's order (1e: STR INT WIS DEX CON CHA; XXVc: STR DEX CON INT WIS CHA TCH)
	AbilityNames map[Ability]AbilityName // their words; one without a name is not listed
	Saves        []Save
	SaveNames    map[Save]string // the saves a sheet lists, by the words it wears; one without a name is not listed (XXVc's original names none)
	Features     []Feature
	Monsters     []MonsterID
	Items        []ItemID
	Ships        []ShipID
	// ScriptFields are the member fields of a script layout this ruleset answers, by the layout's name (a name is a
	// layout's word, its meaning the Fact: 1e's SelMageLevel is the level in class mage). The engine binds each as a
	// read of the selected member; a layout name not here and not the engine's own is bound to nothing.
	ScriptFields map[string]ScriptField
}

// ScriptField is what a ruleset answers for one layout member field: Fact is what it reads of the selected member.
// Census: opCheckParty may average it over the party (por's SelFindTraps). InSave: the field's cell is part of a
// save (the save format as it was when the engine bound these names one by one; a save keeps no meaning in it).
type ScriptField struct {
	Fact   Fact
	Census bool
	InSave bool
}

// AbilityName is how a sheet words an ability: the abbreviation ("STR") and the name written out ("Strength").
type AbilityName struct{ Short, Long string }

// Key is the ability's key in the front's state (game.Member.Stats): its abbreviation in lower case.
func (n AbilityName) Key() string { return strings.ToLower(n.Short) }
