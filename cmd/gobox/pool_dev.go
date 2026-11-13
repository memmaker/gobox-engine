//go:build devviews

// The pool engine's dev views (-dev, -battle, -fight, -at, -ecltrace, -png).
// Built only with `-tags devviews`; pool_nodev.go refuses these flags otherwise.
package main

import (
	"bytes"
	"fmt"
	"gobox/engine"
	"image/png"
	"os"
)

func devViews(eng *engine.Engine, s start) error {
	capt := s.Capture
	if capt.PNG != "" { // the -png picture the portrait dev scenes show
		b, err := os.ReadFile(capt.PNG)
		if err == nil {
			_, err = png.Decode(bytes.NewReader(b))
		}
		if err != nil {
			return fmt.Errorf("-png: %w", err)
		}
		eng.DevPortrait = b
	}
	if s.Dev != "" {
		eng.DevAction = eng.DevScene(s.Dev)
	}
	if s.Battle {
		eng.DevAction = eng.DevScene("autocombat")
	}
	dev := capt.Dev
	if dev.Fight > 0 {
		eng.DevMonster = dev.Fight
		eng.DevAction = eng.DevScene("combat")
	}
	if a := dev.At; a != nil {
		eng.DevAction = func() { eng.DevAt(a.Block, a.X, a.Y) }
	}
	if dev.ECLTrace {
		eng.TraceScripts()
	}
	return nil
}
