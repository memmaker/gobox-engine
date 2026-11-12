package pool

import (
	"fmt"
	"gobox/core"
)

// word is the ruleset's sheet word for c's fact f (core.WordFact).
func (e *Engine) word(cID core.CharID, f core.PlainFact) string {
	v, err := e.ruleset().Fact(cID, core.WordFact{Of: f})
	if err != nil {
		panic(fmt.Sprintf("sheet: word %v: %v", f, err))
	}
	return string(v.(core.Text))
}
