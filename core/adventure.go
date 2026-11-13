package core

// ShopID names a shop of the adventure. Factor is the script's price word at
// the moment of asking, where one exists (cotab word 877).
type ShopID int32

type (
	ShopQuery struct {
		Shop   ShopID
		Item   ItemRef
		Factor int32
	}
	ShopPrices struct{ Buy, Sell Copper }
)

// Campaign is what an adventure says of the rules' choices. Spells nil: all in
// the registry may be learned. The rules hold class progressions and race caps;
// MaxLevel stops training and never discards experience.
type Campaign struct {
	MaxLevel int32
	Spells   []SpellID
	Tables   Tables // nil: the ruleset holds all its numbers
}

// Tables are a ruleset's numbers that differ by adventure (BR1's and BR2's
// START.EXE data), written by the importer and passed at Begin. Each ruleset's
// concepts package has the one implementation (core/xxvc.Tables).
type Tables interface{ RuleTables() }

// Options are the player's. Off by default.
type Options struct {
	CriticalHits   bool // POD's rule, all games
	PartyInventory bool // no encumbrance, one shared slotted inventory: POR's today
	// SharedPurse: the party's coins are one purse. Off (the original): each
	// member's own; turning it off gives the purse to the first member.
	SharedPurse bool
	// Difficulty is the player's setting (BR's Novice..Ace); the zero value is
	// each game's default.
	Difficulty Difficulty
}

// Difficulty is how hard the player wants fights. DifficultyStandard, the
// zero value, is the original's default (BR Trooper); each step away is one
// of the original's levels.
type Difficulty int32

const (
	DifficultyStandard Difficulty = iota // BR Trooper
	DifficultyEasiest                    // BR Novice
	DifficultyEasier                     // BR Cadet
	DifficultyHarder                     // BR Veteran
	DifficultyHardest                    // BR Ace
)

// Dialect names a script byte-code dialect (script.VersionID), for the machine's
// operand widths and memory layout: data, not behaviour flags.
type Dialect int32

const (
	DialectPOR   Dialect = 1 // script.V1
	DialectCOTAB Dialect = 2 // script.V2, also SotSB
	DialectPOD   Dialect = 3 // script.V3
	DialectBR    Dialect = 4 // script.V4
)
