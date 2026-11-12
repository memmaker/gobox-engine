package pool

// Pace is the loop that goes on now, for the outside tests' settle.
func (e *Engine) Pace() Pace { return e.pace() }

// Advance runs the engine on to its next wait, as Do does after an action.
func (e *Engine) Advance() { e.advance() }
