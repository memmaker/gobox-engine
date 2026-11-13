package importer

// Spell table fields used in combat.
const (
	spellTargetsField  = 6  // low nibble: how targets are chosen; see spellTargets
	spellAffectField   = 10 // the affect the spell lays
	spellCombatField   = 11 // 0: camp only
	spellCastTimeField = 12 // segments; a third of it is rounds of delay
	spellAimField      = 14 // 0: the caster himself
)

// Spell target kinds (the low nibble of spellTargetsField).
const (
	spellTargetsCaster    = 0
	spellTargetsHitDice   = 5  // creatures up to a rolled number of hit dice
	spellTargetsAreaFirst = 8  // 8..14: an area round a square, of radius kind&7
	spellTargetsAreaLast  = 14 // ...
	spellTargetsOneOrArea = 15 // one creature, or the area round an empty square
	spellTargetsCountMask = 3  // otherwise: (kind & 3) + 1 creatures
)
