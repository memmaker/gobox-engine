package core

// What the engine's screens ask of the rules about the party, its items and
// the training hall (sweep Sviews, fold E-R6): additive Checks, Facts and
// Actions a ruleset answers; none of them has a side effect but the Actions.

type (
	// PartySize: the members of the party (Value).
	PartySize struct{}
	// HiredCount: the hired NPCs that joined so far (Value).
	HiredCount struct{}
	// TrainFee: the gold the hall asks (Value).
	TrainFee struct{}
	// PoolCap: the party inventory's room (Value).
	PoolCap struct{}
	// CarryCount: what Who can reach, the party inventory after its own (Value).
	CarryCount struct{ Who CharID }
	// HallOpen: the training hall is up (Yes); Value 1 when it waits for a spell pick.
	HallOpen struct{}
	// CanJoin: Who may join the party (Yes); the Refusal says why not is the rules' own.
	CanJoin struct{ Who CharID }
	// InParty: a character of Who's name and random id is in the party (Yes).
	// Value: the members the player controls (hired ones not) up to it, or all.
	InParty struct{ Who CharID }
)

func (PartySize) check()  {}
func (HiredCount) check() {}
func (CarryCount) check() {}
func (PoolCap) check()    {}
func (TrainFee) check()   {}
func (HallOpen) check()   {}
func (CanJoin) check()    {}
func (InParty) check()    {}

type (
	// ListedFact is the items Who sees listed as the inventory orders them
	// (Handles); Goods: the shop's wares, named for no one.
	ListedFact struct{ Goods bool }
	// IdentifiableFact is the items the shop's Id offers Who (Handles).
	IdentifiableFact struct{}
	// ItemCardFact is Item as Who sees it (ItemCard); Who may be NoChar.
	ItemCardFact struct{ Item ItemHandle }
	// TrainFact is what the hall offers Who (TrainCard).
	TrainFact struct{}
	// LearnableFact is the spells Who may learn in the hall (SpellCards).
	LearnableFact struct{}
	// DamageFact is the dice Who's blow rolls, as the sheet words them ("1d8+1": Text).
	DamageFact struct{}
	// BookFact is Who's spell book (Book).
	BookFact struct{}
	// TempleCureFact is cure Cure of the temple, as Who needs it (TempleCure); Who may be NoChar.
	TempleCureFact struct{ Cure int }
	// PurseFact is the party's coins, or the loot's (Purse).
	PurseFact struct{ Loot bool }
)

func (ListedFact) fact()       {}
func (IdentifiableFact) fact() {}
func (ItemCardFact) fact()     {}
func (TrainFact) fact()        {}
func (LearnableFact) fact()    {}
func (TempleCureFact) fact()   {}
func (BookFact) fact()         {}
func (DamageFact) fact()       {}
func (PurseFact) fact()        {}

type (
	// Handles are items in an order.
	Handles []ItemHandle
	// ItemCard is an item as a character sees it.
	ItemCard struct {
		Name                string
		Type, Count, Price  int32
		Weight              int32
		Readied, Identified bool
		Usable              bool // it has a Use
		CannotUse           bool // the class may not ready it
		Tab, Slot           string
		Words               []string
		Hidden              string   // its real name while unidentified
		Look, LookPlain     []string // against the item in its slot; alone
		Sale                Copper
		Scribing            bool // someone was to scribe from the scroll
		Scribe              bool // Who may scribe its first spell
		Scroll              []ScrollSpell
	}
	// ScrollSpell is one spell on a scroll.
	ScrollSpell struct {
		Spell    SpellID
		Number   int32 // as the game numbers it
		Scribing bool
	}
	// TrainCard is the hall's offer: Why not ("" it trains), what it becomes, the experience held at.
	TrainCard struct {
		Why     string
		Becomes []string
		Best    int32
	}
	// Book is a character's spells for the roster cards: the book's, those in
	// the head, the slots a class has per level (class 0 cleric, 1 mage; level 1..3
	// at index 0..2) and the level of each class column.
	Book struct {
		Known, Memorized []BookSpell
		Slots            [2][3]int32
		Levels           []int32
	}
	// BookSpell is a spell of a book as the engine numbers it (Number, the game's
	// own spell numbers) with its class and level, how many are planned and
	// whether it is still to be memorized (Memorized lists only).
	BookSpell struct {
		Spell        SpellID
		Number       int32
		Name         string
		Class, Level int32
		Planned      int32
	}
	// SpellCard is a spell to pick: its name and level.
	SpellCard struct {
		Spell  SpellID
		Number int32 // as the game numbers it
		Name   string
		Level  int32
	}
	// SpellCards are spells in an order.
	SpellCards []SpellCard
	// TempleCure is a cure of the temple and whether Who needs it.
	TempleCure struct {
		Name   string
		Needed bool
	}
)

func (Handles) value()    {}
func (ItemCard) value()   {}
func (TrainCard) value()  {}
func (SpellCards) value() {}
func (TempleCure) value() {}
func (Purse) value()      {}
func (Book) value()       {}

type (
	// OpenHall opens the training hall for Who (the party menu's Train).
	OpenHall struct{ Who CharID }
	// CloseHall leaves it.
	CloseHall struct{}
	// IdentifyItem pays the shop's price for Item's hidden words to show (the
	// Outcome has an ItemIdentified; none: the party could not pay).
	IdentifyItem struct {
		Who  CharID
		Item ItemHandle
	}
	// EnterShop makes the party's screen the shop's.
	EnterShop struct{}
	// StockShop names the wares for no one and prices a free one at a piece,
	// and sounds the door.
	StockShop struct{}
	// ClearHired forgets the hired NPCs of the roster (an empty party starts again).
	ClearHired struct{}
)

func (OpenHall) action()     {}
func (CloseHall) action()    {}
func (ClearHired) action()   {}
func (EnterShop) action()    {}
func (IdentifyItem) action() {}
func (StockShop) action()    {}

// RefusedWhy is a refusal with the words the ruleset gives it.
type RefusedWhy struct{ Why string }

func (r *RefusedWhy) Error() string { return r.Why }

// ItemIdentified is the shop's Id done: Revealed is false when the item had
// nothing new to tell.
type ItemIdentified struct {
	Item     ItemHandle
	Revealed bool
}

func (ItemIdentified) effect() {}
