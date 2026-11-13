package ui

import "gobox/game"

// Launch is a started game as cmd/gobox runs it: its screens, its window
// and what it does with the shared engine settings (HostSettings).
type Launch struct {
	Front *Front
	Title string
	W, H  int // the window's size at scale 1, and its shape

	Apply func(HostOptions)      // the settings changed (once after loading too): the game takes them up
	Voice func(toggle bool) bool // the Audio page's voice-over (HostOptionRows; nil: none)

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

// Tune is the Apply of a game outside gobox: the settings it reads, handed
// to it if it takes them (game.Tuned).
func Tune(g game.Game) func(HostOptions) {
	t, ok := g.(game.Tuned)
	if !ok {
		return nil
	}
	return func(h HostOptions) { t.Tune(game.Tuning{Deskew: h.Deskew}) }
}
