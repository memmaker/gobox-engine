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
	Shared    bool   // dev: a rulesxxvc game on the shared engine (hosts) instead of its own
	Seed      uint64 // the dice seed (resolveSeed)
}
