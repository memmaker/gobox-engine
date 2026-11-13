package ui

// The pure text helpers live in gobox/ui/text (no gpu or input dependency).
type (
	Cell     = TextCell
	Frame    = TextFrame
	UITheme  = TextTheme
	Console  = TextConsole
	MenuItem = TextMenuItem
	List     = TextList
)

var (
	NewConsole = TextNewConsole
	IsHotkey   = TextIsHotkey
	MenuItems  = TextMenuItems
	Punct      = TextPunct
	Flow       = TextFlow
	Wrap       = TextWrap
	PadLeft    = TextPadLeft
	PriceLine  = TextPriceLine
	Commas     = TextCommas
)
