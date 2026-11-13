//go:build nodev

package engine

// A `-tags nodev` build has no dev scenes (devviews_engine.go), so no
// continuation or answer of theirs can be pending.
func (e *Engine) devCont(c cont) { panic("dev scene step in a -tags nodev build") }

func (e *Engine) devAnswered(d askDone, v ansv) {
	panic("dev scene answer in a -tags nodev build")
}
