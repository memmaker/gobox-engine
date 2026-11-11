package core

// Additions of sweep Secl (fold E-R6): what por's script command handlers
// (eclops*, story) ask of the rules.

type (
	// RollDice asks N dice of Sides from the rules' own generator (Answer.Value).
	RollDice struct{ N, Sides int32 }
	// RollRandom asks Random(N) from the rules' generator (Answer.Value), the script's RANDOM.
	RollRandom struct{ N int32 }
	// IsDay asks whether the clock is by day (Yes).
	IsDay struct{}
	// PartyMoveRange asks the fastest (Answer.High) and slowest (Answer.Low)
	// party member's movement, haste doubling and slow halving (por PartyHpRange).
	PartyMoveRange struct{}
)

type (
	// ClearLoot empties the fight's coins and items (por opClearMonsters).
	ClearLoot struct{}
	// SetLoot replaces the fight's coins by the seven kinds of coin (por opTreasure).
	SetLoot struct{ Coins [7]uint16 }
	// AddRandomTreasure puts N random items at the head of the treasure and
	// names each treasure item for Who (por opTreasure's random form).
	AddRandomTreasure struct {
		N   int32
		Who CharID
	}
	// Rob takes Pct percent of the purse times the robbed share of the party,
	// and each robbed member's items with chance ItemPct (por opRob).
	Rob struct {
		Robbed       Party
		Pct, ItemPct int32
	}
	// SetInactive makes Who unable to act (por DuelSetup: the other members sit out).
	SetInactive struct{ Who CharID }
	// SetOrder sets Who's place in its monster group (por addMonsters: the icon).
	SetOrder struct {
		Who   CharID
		Order int32
	}
)

func (RollDice) check()       {}
func (RollRandom) check()     {}
func (IsDay) check()          {}
func (PartyMoveRange) check() {}

func (ClearLoot) action()         {}
func (SetLoot) action()           {}
func (AddRandomTreasure) action() {}
func (Rob) action()               {}
func (SetInactive) action()       {}
func (SetOrder) action()          {}
