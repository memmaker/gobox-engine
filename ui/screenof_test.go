package ui

import "testing"

// The input that opens a screen never acts on it: held through the change
// it is ignored until let go, then works again.
func TestLatchScreens(t *testing.T) {
	var l Latch
	l.See(Screens(), nil)
	if !l.Down("left", true) {
		t.Fatal("a press on the same screen should count")
	}
	l.See("inventory", []any{"left"}) // the click opened the inventory
	if l.Down("left", true) {
		t.Fatal("the click that opened the screen acted on it")
	}
	l.See("inventory", nil) // let go
	if !l.Down("left", true) {
		t.Fatal("a new press on the new screen should count")
	}
	l.See("inventory", []any{"left"}) // the same screen: nothing turns stale
	if !l.Down("left", true) {
		t.Fatal("a held input on an unchanged screen was dropped")
	}
}

// A held input is ignored across a switch between the party menu and the
// creation sheet.
func TestLatchScreenSwitch(t *testing.T) {
	var l Latch
	menu := Screens(ScreenUp{Name: ScreenMainMenu, On: true})
	sheet := Screens(ScreenUp{Name: ScreenCreation, On: true}) // the presenter names its own screens
	if menu == sheet {
		t.Fatal("menu and creation sheet share a screen name")
	}
	l.See(menu, nil)
	l.See(sheet, []any{"enter"})
	if l.Down("enter", true) {
		t.Fatal("the key held through the switch acted on the creation sheet")
	}
}
