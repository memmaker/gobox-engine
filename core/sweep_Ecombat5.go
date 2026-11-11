package core

// Additions of sweep Ecombat5 (fold E-R6): what por's combat screens read.

// AtFact is a fighter's square on the fight's board (por CombatX, CombatY);
// its Value is a BoardPos.
type AtFact struct{}

func (AtFact) fact()    {}
func (BoardPos) value() {}

// BankLoot: the fight's loot coins go into the party's purse (por afterShared:
// MoneySound, Purse.Add, Loot cleared).
type BankLoot struct{}

func (BankLoot) action() {}
