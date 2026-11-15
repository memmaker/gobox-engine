package rules

// MonsterKnown: the data set in play holds monster ID (Yes).
type MonsterKnown struct{ ID MonsterID }

// FirstMonster: the data set in play holds a monster (Yes), the lowest id of them (Value).
type FirstMonster struct{}

// AddTreasureRecords puts the game's item file records Records at the head
// of the treasure, as AddTreasureBlock does a data set's block (the vault's
// items, read back from its file).
type AddTreasureRecords struct{ Records ItemRecords }

func (MonsterKnown) check()        {}
func (FirstMonster) check()        {}
func (AddTreasureRecords) action() {}
