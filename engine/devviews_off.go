//go:build !devviews

package engine

// The dev scenes are built only with `-tags devviews` (devviews_engine.go);
// without it no continuation or answer of theirs can be pending.
func (e *Engine) devCont(c cont) { panic("dev scene step in a build without -tags devviews") }

func (e *Engine) devAnswered(d askDone, v ansv) {
	panic("dev scene answer in a build without -tags devviews")
}
