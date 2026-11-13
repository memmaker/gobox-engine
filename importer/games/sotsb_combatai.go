package games

// Spell table fields read by the computer's spell choice.
const (
	spellSaveTypeField = 9  // the saving throw kind against it
	spellAIPowerField  = 13 // the highest pass (7 down to 1) at which the computer uses it
	spellAIRadiusField = 15 // area radius; 0: one target only
)
