package rules

// SourceTable is one of the adventure's tables a ruleset reads.
type SourceTable int

const (
	SourceMonsters SourceTable = iota // the monster archive of the set in play
	SourceTreasure                    // the treasure archive of the set in play
	SourceItems                       // the base item table
)

// Source is where a ruleset reads its own tables from: the adventure's data
// folder, and what of the game in play picks among it (the set).
type Source interface {
	RulesFile() string              // the adventure's rules.json ("": it has none)
	Name(t SourceTable) string      // the table's file now ("MON1"): what a reader caches by
	Read(t SourceTable, v any) bool // that file decoded (JSON) into v; false: the data has none. A broken file panics: run the import again
}

// Builder builds a ruleset on its source; host is the engine's side of the creation.
type Builder func(src Source, host CreationHost) (Rules, CreationRules)
