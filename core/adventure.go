package core

// Adventure is a module's real data, supplied on demand: it answers what the
// engine asks and decodes its own files into neutral types. It knows nothing
// of any Rules.
type Adventure interface {
	// Identity names the module for saves and the window. game.Data.ID, Series.
	Identity() Identity
	// Needs is every neutral id its data uses, for Mismatches at load.
	Needs() Needs

	// Scripts is how to take its scripts: dialect, bindings, fixes.
	// script.Library{Version}; Version.Vars, ExitOverrides.
	Scripts() ScriptSet
	// ScriptBlock is a raw script block with its header, false if absent.
	// script.Source.Block, DAXSource (library.go).
	ScriptBlock(disk, block int32) ([]byte, bool)
	// ScriptFile is a checked-in file beside them: names/*, patches/*.
	// script.ReadNames, patchPath (import.go); por data/assets/scripts.
	ScriptFile(ScriptFilePath) ([]byte, bool)
	// TreasureNote is the comment a decompile writes for a treasure block.
	// script.Source.Treasure; por TreasureItems (scriptvalues.go:76).
	TreasureNote(ItemSetRef) string

	// Area is an area's names and set. names/areas.txt, script.ReadAreas; play.Common.Area.
	Area(AreaID) (AreaInfo, error)
	// Map is the raw walls of a map. por geo.go:75, gobox/geo decodes.
	Map(MapRef) ([]byte, error)
	// Creature is a monster or NPC. por LoadMonster (files.go:306); BR monster.go.
	Creature(CreatureRef) (Creature, error)
	// Items is a list of items. por opTreasure (eclops2.go:34), shop goods.
	Items(ItemSetRef) ([]Item, error)
	// Asset is a picture, sound or other file by name and block. Every
	// L.BlockShared of por pic.go:31, anim.go:36, bestiary.go:122, pickgrid.go:67.
	Asset(AssetRef) ([]byte, error)
}

// Identity names a module.
type Identity struct {
	ID     string // game.Data.ID
	Series string // game.Data.Series
	Title  string
}

// Dialect names a script byte-code dialect (script.VersionID): the engine
// maps it to the decompiler's Version and to its machine.
type Dialect int32

const (
	DialectPOR   Dialect = 1 // script.V1
	DialectCOTAB Dialect = 2 // script.V2, also SotSB
	DialectPOD   Dialect = 3 // script.V3
	DialectBR    Dialect = 4 // script.V4, eclbr's machine: longint operands
)

// Quirk is a way a module's scripts need the machine to behave, by name. Each
// is one of today's overrides.
type Quirk int32

const (
	QuirkCompareDaySky   Quirk = iota + 1 // por opCompare (eclops.go)
	QuirkGetTableEffects                  // por opGetTable (eclops.go:427)
	QuirkReturnGoesOn                     // por eclHost.Exit is empty
	QuirkNoRemainder                      // por ecl.NoRemainder
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
	ItemTermRow struct {
		Raw  int32
		Term ItemTerm
	}
)

// ScriptSet is how to read a module's scripts.
type ScriptSet struct {
	Dialect    Dialect
	Quirks     []Quirk
	Exits      []ExitRow
	Vars       []VarName
	Chars      []CharBinding
	World      []WorldBinding
	Conditions []ConditionTerm
	Spells     []SpellTerm
	Items      []ItemTermRow
}

// ScriptFilePath is a file's path under the module's scripts folder.
type ScriptFilePath string

// AreaInfo is what the names files say of an area.
type AreaInfo struct {
	ID   AreaID
	Set  DataSet
	Name string
}

// AssetRef is a block of a named file of the module ("HEAD2", 3); Block -1: the whole file.
type AssetRef struct {
	File  string
	Block int32
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
}
