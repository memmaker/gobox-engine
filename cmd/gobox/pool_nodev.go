//go:build !devviews

package main

import (
	"errors"
	"gobox/engine"
)

// devViews refuses the dev-view flags: they exist only in a `-tags devviews` build.
func devViews(_ *engine.Engine, s start) error {
	dev := s.Capture.Dev
	if s.Dev != "" || s.Battle || s.Capture.PNG != "" || dev.Fight > 0 || dev.At != nil || dev.ECLTrace {
		return errors.New("-dev, -battle, -png, -fight, -at and -ecltrace need a build with -tags devviews")
	}
	return nil
}
