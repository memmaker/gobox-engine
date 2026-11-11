package core

// Additions of sweep Ecombat6 (fold E-R6): por's combat end.

// FightCleared: the fight is over and its rule state goes (por combatEnd:
// the computer's loadouts restored, fight state, clouds and combat map reset).
type FightCleared struct{}

func (FightCleared) action() {}
