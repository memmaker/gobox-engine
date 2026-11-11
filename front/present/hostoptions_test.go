package present

import (
	"slices"
	"testing"
)

// TestHostOptionRows: the rows and their built-in neighbours read, in order, as the Esc menu of gobox
// b8cbba9 (blobber/blobber.go menuOptions) did, with the Group C options Quick search (U2,
// 3be40ab) and Day sky by clock (U3, 36339c3) after Path search.
func TestHostOptionRows(t *testing.T) {
	built := map[string][]string{
		"Graphics": {"Fullscreen", "Dithered darkness", "Torch intensity", "Font"},
		"Gameplay": {"Smooth movement", "Fluid movement", "No Grid", "Walk speed"},
	}
	want := map[string][]string{
		"Graphics": {"Fullscreen", "Dithered darkness", "Torch intensity", "Portraits", "Party layout", "Minimap", "Minimap opacity", "Minimap size", "Picture box", "Damage Numbers", "Font", "Artwork"},
		"Gameplay": {"Smooth movement", "Fluid movement", "No Grid", "Explicit Interactions", "Walk speed", "Deskew combat maps", "4e Combat", "Path search", "Quick search", "Day sky by clock", "Full intro", "Zoom speed", "No Map Zones"},
		"Audio":    {"Voice-over", "Music", "DASP Sfx", "Step sounds", "Music volume", "Effects volume", "Voice volume"},
	}
	h := DefaultHostOptions
	for page, rows := range HostRows(&h.Options, HostOptionRows(&h, func(bool) bool { return true })...) {
		var got []string
		b := built[page]
		for i := 0; i <= len(b); i++ {
			for _, r := range rows {
				if r.At == i || page == "Audio" && i == 0 {
					got = append(got, r.Label)
				}
			}
			if i < len(b) {
				got = append(got, b[i])
			}
		}
		if page == "Audio" {
			got = got[:len(rows)]
		}
		if !slices.Equal(got, want[page]) {
			t.Errorf("%s: %q", page, got)
		}
	}
}
