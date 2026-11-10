package core

// Adventure is a module: scripts, walls, pictures, sounds, names. It knows no
// Rules and no engine, and states its data in core's types: every number of its
// monsters and items is the ruleset's. Canonical files come from the importer;
// reading them, or the originals, is its own business.
type Adventure interface {
	// Identity names the module for saves and the window. game.Data.ID, Series.
	Identity() Identity
	// Needs is every id its data uses, for Mismatches at load.
	Needs() Needs

	// Scripts is how to run its scripts: dialect, bindings, id tables.
	Scripts() ScriptSet
	// Script is an area's canonical script text (the checked-in area text),
	// patched at import where originals did native work. script.Library.
	Script(AreaID) ([]byte, error)
	// Area is an area's names and set. names/areas.txt; play.Common.Area.
	Area(AreaID) (AreaInfo, error)
	// Treasure is a list of placed items. por opTreasure (eclops2.go:34).
	Treasure(TreasureID) ([]ItemRef, error)
	// ShopPrice answers when asked, from a table or computed at run time.
	ShopPrice(ShopQuery) (ShopPrices, error)
	// Campaign is what this module's world allows.
	Campaign() Campaign

	// Assets, decoded: on demand and by typed id.
	Map(MapID) (Walls, error)           // por geo.go:75
	Picture(PictureID) (Picture, error) // por pic.go:31, anim.go:36, pickgrid.go:67
	Sound(SoundCue) (Sound, error)      // a rule's cue or a script's
}

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
}

// Options are the player's. Off by default.
type Options struct {
	CriticalHits   bool // POD's rule, all games
	PartyInventory bool // no encumbrance, one shared slotted inventory: POR's today
}

// Identity names a module.
type Identity struct {
	ID     string // game.Data.ID
	Series string // game.Data.Series
	Title  string
}

// Dialect names a script byte-code dialect (script.VersionID), for the machine's
// operand widths and memory layout: data, not behaviour flags.
type Dialect int32

const (
	DialectPOR   Dialect = 1 // script.V1
	DialectCOTAB Dialect = 2 // script.V2, also SotSB
	DialectPOD   Dialect = 3 // script.V3
	DialectBR    Dialect = 4 // script.V4
)

// ExitRow corrects the exit heuristic for one menu (script.ExitOverrides).
// Row is NoExitRow when every row is a real choice.
type ExitRow struct {
	Area, Instr, Row int32
}

// NoExitRow is script.NoExitRow.
const NoExitRow int32 = -1

// ConditionTerm: the number a script's FindAffect or CheckParty gives. SpellTerm likewise.
type (
	ConditionTerm struct {
		Raw       int32
		Condition Condition
	}
	SpellTerm struct {
		Raw   int32
		Spell SpellID
	}
	ItemTerm struct {
		Raw  int32
		Item ItemID
	}
)

// ScriptSet is how to read a module's scripts.
type ScriptSet struct {
	Dialect    Dialect
	Exits      []ExitRow
	Chars      []CharBinding
	World      []WorldBinding
	Conditions []ConditionTerm
	Spells     []SpellTerm
	Items      []ItemTerm
}

// AreaInfo is what the names files say of an area.
type AreaInfo struct {
	ID   AreaID
	Set  DataSet
	Name string
}

// Needs is the neutral ids an adventure's data uses.
type Needs struct {
	Facts      []Fact
	Spells     []SpellID
	Conditions []Condition
	Traits     []Trait
	Races      []Race
	Classes    []Class
	Skills     []Skill
	Saves      []Save
	Features   []Feature
	Monsters   []MonsterID
	Items      []ItemID
	Ships      []ShipID
}
