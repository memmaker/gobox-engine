package rules1e

import "gobox/rules"

// testSource is a source of a rules file only: no tables.
type testSource struct{ rules string }

func (s testSource) RulesFile() string              { return s.rules }
func (testSource) Name(rules.SourceTable) string    { return "" }
func (testSource) Read(rules.SourceTable, any) bool { return false }
