package games

import (
	"gobox/rules1e"
)

// CreationStart is POD's per-adventure creation start for rules1e
// (rules.json creationStart, Felix's R3 rulings 3, 7, 8, 11; ruling 4: no extra years), from the
// imported adventure folder adv: 200 platinum, the experience by class
// count trained silently, the magic-user spells
// (startingMageSpells) and the class kits from ITEM0 (startingEquipment;
// the ItemCount=3 quirk dropped).
func PodCreationStart(adv string) rules1e.StartData {
	s := rules1e.StartData{
		Money:      rules1e.StartMoney{Fixed: podStartPlatinum * goldPerPlatinum}, // platinum by value (Felix 2026-11-12: gold and copper only),
		Experience: []int32{podStartExperience, podStartExperience2, podStartExperience3},
		MageSpells: rules1e.StartSpells{AllBelow: 4, AllAt: allMageSpellsLevel, BySlot: map[int32][]rules1e.SpellID{}},
	}
	for lvl, ids := range startingMageSpellsByLevel {
		for _, id := range ids {
			sid, ok := rules1e.PodSpellIDs[byte(id)] // BySlot holds unified ids, startingMageSpellsByLevel the game's numbers
			if !ok {
				panic("startingMageSpellsByLevel: spell number not in rules1e.PodSpellIDs")
			}
			s.MageSpells.BySlot[int32(lvl)] = append(s.MageSpells.BySlot[int32(lvl)], sid)
		}
	}
	library := ItemLibrary(adv)
	for kit, ids := range startingKits {
		for slot, id := range ids {
			if id == 0 || int(id) > len(library) || library[id-1] == nil {
				continue
			}
			it := *library[id-1]
			it.Hidden &^= hiddenWordBits
			if (slot == 4 && kit == kitFighter) || (slot == kitUnreadiedSlot && kit == kitOther) {
				it.Count = kitMissileCount
			}
			if slot != kitUnreadiedSlot {
				it.Readied = 1
			}
			s.Kit[kit] = append(s.Kit[kit], it)
		}
	}
	return s
}

// ItemLibrary is the imported adventure adv's ITEM0 library as rules1e
// items, by id-1 (rules.json itemLibrary); nil: an empty entry.
func ItemLibrary(adv string) []*rules1e.Item {
	library := OpenPodFolder(adv).Data.Items("ITEM0", 1)
	out := make([]*rules1e.Item, len(library))
	for i, it := range library {
		if it.Type == libraryAbsent { // the library record's first byte, the item's type
			continue
		}
		out[i] = &it
	}
	return out
}

// RandomScroll is randomScroll's making for rules1e (rules.json randomScroll).
func RandomScroll() rules1e.RandomScroll {
	s := rules1e.RandomScroll{Type: scrollType, Plus: scrollPlus, Weight: scrollWeight, Price: scrollPrice, Last: spellLast}
	for k, w := range scrollWords {
		s.Words[k] = rules1e.WordID(w)
	}
	for n := 1; n <= spellLast; n++ {
		if podSpellTable[n].class() == podSpellClassMage {
			s.Mage = append(s.Mage, rules1e.SpellID(n))
		}
	}
	return s
}
