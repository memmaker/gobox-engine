package importsotsb

import "rules1e"

// CreationStart is Silver Blades' per-adventure creation start for rules1e
// (rules.json creationStart, Felix's R3 rulings 3, 7, 8): 2d6 gold for a
// thief class, the experience by class count trained silently,
// mageStartSpells. (The unreachable cleric-without-experience branch goes.)
func CreationStart() rules1e.StartData {
	none := []int16{0, 0}
	s := rules1e.StartData{
		Money:      rules1e.StartMoney{Times: 1, Dice: [][]int16{none, none, none, none, none, none, {2, 6}, none}}, // rows: cleric druid fighter paladin ranger mage thief monk
		Experience: []int32{startExperience, startExperience2, startExperience3},
	}
	for _, id := range mageStartSpells {
		s.MageSpells.Known = append(s.MageSpells.Known, rules1e.SpellID(id))
	}
	return s
}
