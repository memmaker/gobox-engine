package ui

// Boot is what cmd/gobox hands a game's start: the adventure folder and the
// command line's switches a game reads.
type Boot struct {
	Adventure string // the adventure folder (importer.Folder.Dir)
	Data      string // the host's data directory for the game (the player's music)
	Saves     string // -saves: the save directory ("": the game's)
	Seed      int64  // -seed: the same dice every run; -1: the clock
	Scripted  bool   // -keys: a scripted run (one seed, so each prefix's shot is the same game)
	Fight     int    // -fight: a fight with this monster at the first adventuring prompt (0: none)
}

// Launch is a started game as cmd/gobox runs it: its screens, its window
// and its Esc menu.
type Launch struct {
	Front *Front
	Title string
	W, H  int // the window's size at scale 1, and its shape

	SettingsDir string           // the folder of the settings file (SETTINGS.JSON)
	Extra       any              // a pointer to the game's own options struct, kept beside Settings (its fields at the file's top level)
	Rows        map[string][]Row // ExtraRows (HostRows)
	Apply       func()           // the options changed (once after loading too): the game takes them up

	SaveLoad, MainMenu func() // OnSaveLoad, OnMainMenu (nil: none)

	Art      *Arts  // two arts, cross-faded (nil: the Front's own, if any)
	FontPrev uint16 // the code Ctrl-Shift-F sends (0: none)
	TPS      int    // ticks a second; 0: 60
}

// Arts is a game with two arts (e.g. two ports' pictures), cross-faded when
// the one asked for changes (F12 toggles it).
type Arts struct {
	Want   func() bool // the art asked for
	Toggle func()      // F12: ask for the other
	Set    func(bool)  // the game now shows this art
}
