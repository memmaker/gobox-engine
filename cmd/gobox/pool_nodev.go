//go:build nodev

package main

import (
	"errors"
	"gobox/engine"
)

// devViews refuses the dev-view flags: a `-tags nodev` build has none.
func devViews(_ *engine.Engine, s start) error {
	dev := s.Capture.Dev
	if s.Dev != "" || s.Battle || s.Capture.PNG != "" || dev.Fight > 0 || dev.At != nil || dev.ECLTrace {
		return errors.New("-dev, -battle, -png, -fight, -at and -ecltrace are not in a -tags nodev build")
	}
	return nil
}
