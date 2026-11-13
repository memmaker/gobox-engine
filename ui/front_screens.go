package ui

import "strings"

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
