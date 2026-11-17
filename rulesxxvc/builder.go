package rulesxxvc

import (
	"gobox/rules"
	"gobox/rules/xxvc"
)

// Builder is the XXVc ruleset for the shared engine (rules.Builder): its tables
// are read from the source (rules.SourceRuleTables, the import's tables file) and
// handed to Begin with the default options. The adventure has no tables
// without its import: that panics, as a broken data folder does (run the import again).
func Builder(src rules.Source, host rules.CreationHost) (rules.Rules, rules.CreationRules) {
	var t xxvc.Tables
	if !src.Read(rules.SourceRuleTables, &t) {
		panic("rulesxxvc: the adventure has no tables file (run the import again)")
	}
	r := New()
	r.Begin(rules.Campaign{Tables: t}, rules.Options{})
	r.pics = host
	return r, creation{r: r, host: host}
}
