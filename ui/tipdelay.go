package ui

import (
	"image"
	"time"

	"gobox/gpu"
)

// tipRest is where the pointer came to rest and since when (tipShown).
var tipRest struct {
	at    image.Point
	since time.Time
}

// tipShown reports whether a mouse tooltip may show: the pointer has rested
// Host.TipDelay seconds (the Gameplay option); any move starts the wait again.
func tipShown() bool {
	x, y := gpu.CursorPosition()
	if pt := image.Pt(x, y); pt != tipRest.at || tipRest.since.IsZero() {
		tipRest.at, tipRest.since = pt, time.Now()
	}
	return time.Since(tipRest.since) >= time.Duration(Host.TipDelay*float64(time.Second))
}
