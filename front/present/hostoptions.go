package present

// HostOptions are the settings the Esc menu of gobox's blobber does not keep
// itself (it keeps only its own view options): the host writes them to
// SETTINGS.JSON beside blobber.Options (app.Menu.Extra) and puts their rows
// back with HostOptionRows. The engine reads them through the host's apply.
type HostOptions struct {
	Options                       // the presenter's own: minimap, portraits, picture box, zoom, damage numbers
	Deskew                bool    // the combat maps on a square grid (walls straight) instead of the original's perspective
	Music, Effects, Voice float64 // the volumes, 0 to 1
	Dasp                  bool    // the Dubtrain sound pack effects in place of the original's
	Soundtrack            bool    // the music from the recordings instead of the original's
	Steps                 bool    // the step and turn sound
	Combat4e              bool    // the fights under the 4e rule set laid over the original's
	PathSearch            bool    // the computer's creatures walk by a path search instead of the original's step towards the target
	QuickSearch           bool    // a sweeping step or a Look takes 3 minutes instead of the original's turn
	DaySky                bool    // the scripts' day sky follows day 6:00-20:00 instead of their hour < 14
	Explicit              bool    // the harmless square events wait for Space instead of running on stepping
	DOSArt                bool    // the pictures, portraits and combat icons in the DOS art instead of the Amiga's (F12)
}

// DefaultHostOptions are the host settings before SETTINGS.JSON is read.
var DefaultHostOptions = HostOptions{Options: DefaultOptions, Deskew: true, Music: 1, Effects: 1, Voice: 1, Dasp: true, Steps: true}

// HostOptionRows are the Esc menu rows of h (HostRows puts them with the
// presenter's own): the Graphics and Gameplay rows go before the built-in
// row At, the Audio page is new. voice toggles the voice-over when its
// argument is true and reports whether it is on.
func HostOptionRows(h *HostOptions, voice func(toggle bool) bool) []HostRow {
	return []HostRow{
		{Page: "Graphics", Label: "Artwork", At: 4, Flag: &h.DOSArt, Names: [2]string{"Amiga", "DOS"}},
		{Page: "Gameplay", Label: "Explicit Interactions", At: 3, Flag: &h.Explicit},
		{Page: "Gameplay", Label: "Deskew combat maps", At: 4, Flag: &h.Deskew},
		{Page: "Gameplay", Label: "4e Combat", At: 4, Flag: &h.Combat4e},
		{Page: "Gameplay", Label: "Path search", At: 4, Flag: &h.PathSearch},
		{Page: "Gameplay", Label: "Quick search", At: 4, Flag: &h.QuickSearch},
		{Page: "Gameplay", Label: "Day sky by clock", At: 4, Flag: &h.DaySky},
		{Page: "Audio", Label: "Voice-over", Get: func() string { return OnOff(voice(false)) }, Change: func(int) { voice(true) }},
		{Page: "Audio", Label: "Music", Flag: &h.Soundtrack, Names: [2]string{"Original", "Soundtrack"}},
		{Page: "Audio", Label: "DASP Sfx", Flag: &h.Dasp},
		{Page: "Audio", Label: "Step sounds", Flag: &h.Steps},
		{Page: "Audio", Label: "Music volume", Level: &h.Music},
		{Page: "Audio", Label: "Effects volume", Level: &h.Effects},
		{Page: "Audio", Label: "Voice volume", Level: &h.Voice},
	}
}
