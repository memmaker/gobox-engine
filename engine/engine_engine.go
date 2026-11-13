package engine

import (
	"gobox/ega"
	"gobox/rules"
	"time"
)

// Run is the program body (START.EXE main), Start(e.Run): it puts the
// title up; the screens then take the decisions (Apply).
func (e *Engine) Run() {
	e.Init()           // ovr12:002D
	e.ext.intro = true // ovr02: title pictures, lightning, credits: the presenter's (TitleDone)
	// Copy protection (ovr03:00FD code wheel) deliberately skipped.
	// With the title gone, Apply puts the party menu up.
}

// Play is the rest of main after the title (Start(e.Play) after Init):
// Apply puts the party menu up, then the adventure.
func (e *Engine) Play() {
	if e.ruleset().Check(rules.ItemTableMissing{}, nil).Yes {
		e.PromptRow24(ega.Yellow, "ITEMS missing: import the game again") // extra (readItems): shown until a key
	}
}

// Randomize = 0AF8:1247: RandSeed = DOS time (CX = hour:min low word,
// DX = sec:hundredths high word).
func (e *Engine) Randomize() {
	var seed uint32
	rules.Randomize(&seed, e.FixedSeed)
	e.do(rules.SetSeed{Seed: seed})
}

// Ticks = 0709:0C3F, time in hundredths of a second.
func (e *Engine) Ticks() int64 {
	return int64(time.Since(e.start) / (10 * time.Millisecond))
}
