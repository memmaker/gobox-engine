package pool

import (
	"gobox/core"
	"gobox/rules1e"
)

// itemHandle is it's handle in the ruleset, heldOf the item of h (tests reach into the rules' records).
func (e *Engine) itemHandle(it *rules1e.Item) core.ItemHandle { return e.r1e().Handle(it) }
func (e *Engine) heldOf(h core.ItemHandle) *rules1e.Item      { return e.r1e().Held(h) }
