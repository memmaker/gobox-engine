package importcotab

import "gobox/rules1e"

// CreationStart is CotAB's per-adventure creation start for rules1e
// (rules.json creationStart, Felix's R3 rulings 3, 7, 8): 300 platinum, the
// experience by class count trained silently, four magic-user spells.
func CreationStart() rules1e.StartData {
	return rules1e.StartData{
		Money:      rules1e.StartMoney{Fixed: startPlatinum * goldPerPlatinum}, // platinum by value (Felix 2026-11-12: gold and copper only),
		Experience: []int32{startExperience, startExperience2, startExperience3},
		MageSpells: rules1e.StartSpells{Known: []rules1e.SpellID{spellDetectMagicMU, spellReadMagic, spellEnlarge, spellSleep}},
	}
}

// Layout1e is the shape of COTAB's slot tables in rules.json (fold T3b): five
// spell levels a row, and class levels 1..12 before the 0xFF rows and the
// experience longints. The import writes it as rules.json's "layout".
var Layout1e = rules1e.Layout{SlotStride: mageSlotLevels, SlotLevels: 12}
