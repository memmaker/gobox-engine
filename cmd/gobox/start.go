package main

import "gobox/ui"

// start is how main starts an adventure: its folder, the host's data
// directory (the port's own assets, the saves) and the command line's host
// and debug switches.
type start struct {
	Adventure string // the adventure folder (importer.Folder.Dir)
	Data      string // the host's data directory
	Settings  ui.WindowMode
	Capture   ui.Capture
	Param     string // the original's command-line parameter (POR: "STING")
	Explore   bool   // debug: no intro, a Quick Start party when there is none, begin adventuring at once
	Battle    bool   // debug: a fight the computer plays out for both sides at the first prompt
	Dev       string // debug: a dev scene at the first adventuring prompt
	Open      string // debug: open the remastered map there, map:NAME:X:Y:D
	Seed      int64  // debug: the same dice every run; -1: the clock
}

// boot is the part of s a game outside gobox reads (ui.Boot).
func (s start) boot() ui.Boot {
	return ui.Boot{
		Adventure: s.Adventure,
		Data:      s.Data,
		Saves:     s.Capture.Dev.Saves,
		Seed:      s.Seed,
		Scripted:  len(s.Capture.Keys) > 0,
		Fight:     s.Capture.Dev.Fight,
	}
}
