package importer

// A new character's fixed starting values.
const (
	podStartBaseAC        = 50
	podStartTHAC0         = 40
	podStartIconID        = 10
	podStartExperience    = 1500001 // one class
	podStartExperience2   = 750000  // two classes
	podStartExperience3   = 500000  // three classes
	startAgeExtra         = 10      // added to a single class's age
	podStartBaseMove      = 12
	podStartClericSlots   = 2
	podStartPlatinum      = 200
	podExceptionalStrDice = 100
	podStrengthMaximum    = 18
	podMultiClericWisdom  = 13
	podMaxNameLength      = 15
	podAbilityRolls       = 6
	podAllClasses         = 255 // a class mask with every training bit
)

// The magic-user spells O075:21EE grants by the highest spell level the
// new character can cast; the ninth level grants every magic-user spell.
var startingMageSpellsByLevel = map[int][]int{
	4: {82, 84, 88, 81, 83, 85, 87},
	5: {91, 92, 118, 119},
	6: {113, 114, 112, 111},
	7: {115, 117},
}

const allMageSpellsLevel = 9

const (
	kitFighter, kitThief, kitCleric, kitOther = 0, 1, 2, 3
	kitUnreadiedSlot                          = 3  // the fourth entry is carried, not readied
	kitMissileCount                           = 50 // the fighter's fifth and the rest's fourth entry
	libraryItemSize                           = 17
	libraryAbsent                             = 255
	startingItemCount                         = 3 // what O075:5478 leaves in the record (sic)
)
