package core

// Additions of sweep Ecombat6 (fold E-R6): por's combat end.

// FightCleared: the fight is over and its rule state goes (por combatEnd:
// the computer's loadouts restored, fight state, clouds and combat map reset).
type FightCleared struct{}

func (FightCleared) action() {}

// DetectLoot: the actor casts its memorized Detect Magic on the fight's
// treasure (por LootDetect); ErrNotNow when there is no treasure or no such spell.
type DetectLoot struct{}

func (DetectLoot) action() {}

// IconFact is a fighter's combat icon slot (por Character.Order); its
// Value is a Number.
type IconFact struct{}

func (IconFact) fact() {}
