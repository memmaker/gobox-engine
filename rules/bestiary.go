package rules

// BeastKindFact asks the bestiary kind of a monster name ("Goblin" for
// "GOBLIN LEADER"). The Fact's own id is NoChar.
type BeastKindFact struct{ Monster string }

// BeastKind is BeastKindFact's answer; Name "": the monster has no kind.
type BeastKind struct{ Name string }

func (BeastKindFact) fact() {}

func (BeastKind) value() {}
