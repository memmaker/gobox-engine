package core

// Additions of sweep Spool (fold E-R6): the paladin's powers, the door menu's
// questions and the area map's terrain.

type (
	// PaladinHeal is the paladin Who laying hands on Target: 2 hp per paladin
	// level (CotAB PaladinHeal), and the day's use spent. The Outcome holds
	// Healed when Target was healed, none when it is unaffected.
	PaladinHeal struct{ Who, Target CharID }
	// PaladinCure is the paladin Who curing Target of the temple's diseases
	// (CotAB PaladinCureDisease), and the day's use spent.
	PaladinCure struct{ Who, Target CharID }
)

func (PaladinHeal) action() {}
func (PaladinCure) action() {}

type (
	// Diseased asks whether Who has one of the diseases PaladinCure cures (Yes).
	Diseased struct{ Who CharID }
	// PartyHasSkill asks whether a party member holds thief skill number N
	// above 0 (ovr15:0236); Yes.
	PartyHasSkill struct{ N int }
	// MapTerrain asks the area map's terrain number at map square (X, Y) (Value).
	MapTerrain struct{ X, Y int }
)

func (Diseased) check()      {}
func (PartyHasSkill) check() {}
func (MapTerrain) check()    {}

type (
	// ItemViewFact is one held item as the member shows and pictures it
	// (Fact of the holder: an InvItem; its Index, Price and the flags of a
	// list are left unset).
	ItemViewFact struct{ Item ItemHandle }
	// InventoryFact is the member's inventory screen (Fact: an InventoryView);
	// Powers is the engine's to add.
	InventoryFact struct{}
)

// InventoryView is the inventory screen's rule part: the readied items by
// doll slot (nil: empty), the party inventory in its list order, and what
// can be done here.
type InventoryView struct {
	Member int // the member's place in the party
	Cap    int // the party inventory's slots
	Doll   []*InvItem
	Pool   []InvItem
	Use    bool // items can be used here
	Sell   bool // a shop buys them
	Switch bool // another member can be shown (not in combat)
}

func (ItemViewFact) fact()   {}
func (InventoryFact) fact()  {}
func (InvItem) value()       {}
func (InventoryView) value() {}

// The powers a member picks and keeps (rules4e: at-will and encounter powers).
type (
	// OwedFact is the pick the member still owes (Fact: PowerOwed; no Options: none).
	OwedFact struct{}
	// PowersFact is the member's picked powers and daily uses (Fact: PowerSet).
	PowersFact struct{}
)

// PowerOwed is one choice a member owes: one of Options.
type PowerOwed struct {
	What    string // "an at-will power"
	Options ManeuverCards
}

// PowerSet is a member's picked powers (their ids), the daily uses spent
// (Surges, Laid), the Toughness kept in its hit points and its reactions'
// policies.
type PowerSet struct {
	Powers       []int
	Surges, Laid int
	Tough        int
	Policies     map[int]Policy
}

type (
	// Toughen keeps the ruleset's Toughness in Who's maximum hit points.
	Toughen struct{ Who CharID }
	// TakePower adds Power to Who's picked powers.
	TakePower struct {
		Who   CharID
		Power int
	}
	// SetPowers replaces Who's powers and daily uses (a saved member loaded).
	SetPowers struct {
		Who CharID
		Set PowerSet
	}
)

func (OwedFact) fact()    {}
func (PowersFact) fact()  {}
func (PowerOwed) value()  {}
func (PowerSet) value()   {}
func (Toughen) action()   {}
func (TakePower) action() {}
func (SetPowers) action() {}
