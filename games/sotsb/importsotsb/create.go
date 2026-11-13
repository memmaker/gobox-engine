package importsotsb

// A new character's fixed starting values.
const (
	startBaseAC          = 50
	startTHAC0           = 40
	startIconID          = 10
	startExperience      = 200000 // one class
	startExperience2     = 100000 // two classes
	startExperience3     = 66667  // three classes
	startClericExpLawful = 2000   // a cleric still without experience (never happens)
	startClericExpOther  = 1500
	startBaseMove        = 12
	startClericSlots     = 2
	exceptionalStrDice   = 100
	strengthMaximum      = 18
	multiClericWisdom    = 13
	maxNameLength        = 15
	abilityRolls         = 6
	allClasses           = 255 // a class mask with every training bit
)

// The spells a new magic-user knows.
var mageStartSpells = []int{spellDetectMagicMU, spellReadMagic, spellShield, spellSleep}
