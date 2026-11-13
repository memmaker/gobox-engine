package engine

// Advance runs the engine on to its next wait, as Do does after an action.
func (e *Engine) Advance() { e.advance() }
