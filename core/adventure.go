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
	// Treasure is a list of placed items. por opTreasure (eclops2.go:34), shop goods.
	Treasure(TreasureRef) ([]ItemRef, error)

	// Assets, decoded: on demand and by typed id.
	Map(MapID) (Walls, error)           // por geo.go:75
	Picture(PictureID) (Picture, error) // por pic.go:31, anim.go:36, pickgrid.go:67
	Sound(SoundCue) (Sound, error)      // a rule's cue or a script's
}

// TreasureRef names a list of placed items (por opTreasure: ITEM<set>, block).
type TreasureRef struct {
	Set   DataSet
	Block int32
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

// VarName names a word for the decompiled text (script.Version.Vars).
type VarName struct {
	Addr ScriptAddr
	Name string
}

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
	// MonsterTerm: what a script's LoadMonster index means in a data set.
	// Variable indices (POR 24, POD 69 sites) go through the same table at run
	// time; an index with no row is an error.
	MonsterTerm struct {
		Set     DataSet
		Index   int32
		Monster MonsterID
	}
)

// ScriptSet is how to read a module's scripts.
type ScriptSet struct {
	Dialect    Dialect
	Exits      []ExitRow
	Vars       []VarName
	Chars      []CharBinding
	World      []WorldBinding
	Conditions []ConditionTerm
	Spells     []SpellTerm
	Items      []ItemTerm
	Monsters   []MonsterTerm
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
