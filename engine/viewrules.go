package engine

import (
	"fmt"
	"gobox/rules"
)

// word is the ruleset's sheet word for c's fact f (core.WordFact).
func (e *Engine) word(cID rules.CharID, f rules.PlainFact) string {
	v, err := e.ruleset().Fact(cID, rules.WordFact{Of: f})
	if err != nil {
		panic(fmt.Sprintf("sheet: word %v: %v", f, err))
	}
	return string(v.(rules.Text))
}
