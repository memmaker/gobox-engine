package ui

import "gobox/ui/text"

// The pure text helpers live in gobox/ui/text (no gpu or input dependency).
type (
	Cell     = text.Cell
	Frame    = text.Frame
	Theme    = text.Theme
	Console  = text.Console
	MenuItem = text.MenuItem
	List     = text.List
)

var (
	NewConsole = text.NewConsole
	IsHotkey   = text.IsHotkey
	MenuItems  = text.MenuItems
	Punct      = text.Punct
	Flow       = text.Flow
	Wrap       = text.Wrap
	PadLeft    = text.PadLeft
	PriceLine  = text.PriceLine
	Commas     = text.Commas
)
