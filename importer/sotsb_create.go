package importer

// A new character's fixed starting values.
const (
	sotsbStartBaseAC        = 50
	sotsbStartTHAC0         = 40
	sotsbStartIconID        = 10
	sotsbStartExperience    = 200000 // one class
	sotsbStartExperience2   = 100000 // two classes
	sotsbStartExperience3   = 66667  // three classes
	startClericExpLawful    = 2000   // a cleric still without experience (never happens)
	startClericExpOther     = 1500
	sotsbStartBaseMove      = 12
	sotsbStartClericSlots   = 2
	sotsbExceptionalStrDice = 100
	sotsbStrengthMaximum    = 18
	sotsbMultiClericWisdom  = 13
	sotsbMaxNameLength      = 15
	sotsbAbilityRolls       = 6
	sotsbAllClasses         = 255 // a class mask with every training bit
)

// The spells a new magic-user knows.
var mageStartSpells = []int{spellDetectMagicMU, spellReadMagic, spellShield, spellSleep}
