package ui

import "testing"

// TestScreens: a screen opening, another taking its place or a step of
// one changes the names.
func TestScreens(t *testing.T) {
	none := Screens(ScreenUp{ScreenCreation, false, ""}, ScreenUp{ScreenPicker, false, ""})
	sheet := Screens(ScreenUp{ScreenCreation, true, ""}, ScreenUp{ScreenPicker, false, ""})
	pick := Screens(ScreenUp{ScreenCreation, false, ""}, ScreenUp{ScreenPicker, true, ""})
	if none == sheet || sheet == pick || none == pick {
		t.Fatal(none, sheet, pick)
	}
	if Screens(ScreenUp{ScreenTitle, true, "1"}) == Screens(ScreenUp{ScreenTitle, true, "2"}) {
		t.Fatal("the title's steps share a name")
	}
}
