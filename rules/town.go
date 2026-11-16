package rules

// The town: training hall, temple, shop and the purse.

// Pay takes Gold from the purse; refused when short (Refused{NoMoney}, ErrNotNow).
type Pay struct{ Gold uint64 }

func (Pay) action() {}

// SetCredits is a script's write of the party's credits (BR ecl.go setChar OffMoney): Total is what the
// party should hold. A lower Total is paid by the ruleset across the members, a higher one is added to Who;
// which member pays first is its rule.
type SetCredits struct {
	Who   CharID
	Total Copper
}

func (SetCredits) action() {}

// SetPurseGold sets the party's gold pieces (the tavern game's wager).
type SetPurseGold struct{ Gold uint64 }

func (SetPurseGold) action() {}

// TrainFee: the gold the hall asks (Value); Yes: it asks none, and its screen says nothing of a fee.
type TrainFee struct{}

// BundlePrices: a shop's bundle of items costs its count times the price (Yes).
type BundlePrices struct{}

func (BundlePrices) check() {}

// HallOpen: the training hall is up (Yes); Value 1 when it waits for a spell pick.
type HallOpen struct{}

func (TrainFee) check() {}

func (HallOpen) check() {}

// TrainFact is what the hall offers Who (TrainCard).
type TrainFact struct{ Hall HallTerms }

// TempleCureFact is cure Cure of the temple, as Who needs it (TempleCure); Who may be NoChar.
type TempleCureFact struct{ Cure int }

// PurseFact is the party's coins, or the loot's (Purse).
type PurseFact struct{ Loot bool }

func (TrainFact) fact() {}

func (TempleCureFact) fact() {}

func (PurseFact) fact() {}

// TrainCard is the hall's offer: Why not ("" it trains), what it becomes, the experience held at.
type TrainCard struct {
	Why     string
	Becomes []string
	Best    int32
	Rises   bool // the hall would raise a class now (no Why, a class to raise)
}

// TempleCure is a service of the temple: its name and price in gold (0: free), whether
// Who needs it, and what the temple says of one it would not help (Not; "": it always helps).
type TempleCure struct {
	Name     string
	Needed   bool
	Price    int32
	Not      string
	Done     string // what the temple says of a member it helped ("": "is cured.")
	Declines bool   // it only says Not of one it would not help, and does not offer to cast anyway
	Only     bool   // it cannot be chosen for one who does not need it (the wound cures)
}

func (TrainCard) value() {}

func (TempleCure) value() {}

func (Purse) value() {}

// GiveTempleCure does the temple's cure Cure on Who, paid for (por's temple cures, ovr05).
type GiveTempleCure struct {
	Who  CharID
	Cure int32
}

func (GiveTempleCure) action() {}

// OpenHall opens the training hall for Who (the party menu's Train).
type OpenHall struct {
	Who  CharID
	Hall HallTerms
}

// CloseHall leaves it.
type CloseHall struct{}

// StockShop names the wares for no one and prices a free one at a piece,
// and sounds the door. A ruleset whose shops sell a numbered list (BR's
// EclShop) stocks List's wares, priced by Terms; the Pool's script put its
// wares in the treasure before (List 0).
type StockShop struct {
	List  ShopListID
	Terms ShopTerms
}

// ShopListID numbers a shop's item list as the script asks for it (0: none).
type ShopListID int32

func (OpenHall) action() {}

func (CloseHall) action() {}

func (StockShop) action() {}
