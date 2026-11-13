package main

import (
	"testing"
	"time"

	"buckrogers/presentdata"
	"buckrogers/rules"
	"gobox/ui"
)

// Start-up as main does it, headless: each step and the first frames must be quick.
func TestBR1StartupTiming(t *testing.T) {
	start, t0 := time.Now(), time.Now()
	defer func() {
		if d := time.Since(start); d > time.Second {
			t.Errorf("start-up to the fifth frame took %v, want well under a second", d)
		}
	}()
	lap := func(what string) { t.Logf("%-16s %v", what, time.Since(t0)); t0 = time.Now() }
	e, err := rules.New(br1Adv(), t.TempDir(), 1)
	if err != nil {
		t.Skip(err)
	}
	d := presentdata.Data(br1Adv())
	lap("rules")
	f := ui.NewFront(ui.NewGame(e, d), d)
	lap("NewFront")
	f.Start("")
	for i := range 5 {
		f.Frame(1280, 800)
		lap("frame " + string(rune('0'+i)))
	}
}
