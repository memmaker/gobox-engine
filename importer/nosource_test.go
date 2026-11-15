package importer

import "gobox/rules"

// noSource is a source with no rules file and no tables.
type noSource struct{}

func (noSource) RulesFile() string                { return "" }
func (noSource) Name(rules.SourceTable) string    { return "" }
func (noSource) Read(rules.SourceTable, any) bool { return false }
