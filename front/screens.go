package front

import (
	"fmt"
	"strings"
)

// Screen is the name of a front screen up; the input latch tells screens
// apart by the names up (Screens), so an input held across a change never
// acts on the screen it opened.
type Screen string

// The front's screens.
const (
	ScreenMainMenu   Screen = "main menu"
	ScreenCreation   Screen = "creation"
	ScreenPicker     Screen = "picker"
	ScreenIconEditor Screen = "icon editor"
	ScreenSlots      Screen = "slots"
	ScreenQuestion   Screen = "question"
	ScreenTitle      Screen = "title"
)

// Screens joins the names of the screens up (on: up) in the order given.
func Screens(up ...ScreenUp) string {
	var b strings.Builder
	for _, s := range up {
		if s.On {
			b.WriteString(string(s.Name))
			b.WriteString(s.State)
		}
		b.WriteByte(';')
	}
	return b.String()
}

// ScreenUp is a screen and whether it is up; State tells its steps apart
// where each is a screen of its own to the latch (the title's phases).
type ScreenUp struct {
	Name  Screen
	On    bool
	State string
}

// ScreenOf names the screen v shows, with own the presenter's own screens
// (Presenter.Screens): it changes whenever a screen opens or closes, or a
// question pops up (the latch's Screen).
func ScreenOf(v View, own string) string {
	ask := -1
	if v.Ask != nil {
		ask = int(v.Ask.Kind)
	}
	return fmt.Sprint(own, v.Inventory != nil, v.Sheet != nil, v.SheetPage != nil, v.Camp != nil, v.Temple != nil,
		v.Shop != nil, v.Training != nil, v.Powers != nil,
		v.Read != nil, v.Results != nil, v.Battle != nil, v.Credits, ask, Screens(
			ScreenUp{Name: ScreenCreation, On: v.Creation != nil},
			ScreenUp{Name: ScreenPicker, On: v.Picker != nil},
			ScreenUp{Name: ScreenIconEditor, On: v.IconEditor != nil},
			ScreenUp{Name: ScreenMainMenu, On: v.MainMenu != nil}))
}
