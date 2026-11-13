package ui

import "gobox/ui/text"

// The pure text helpers live in gobox/ui/text (no gpu or input dependency).
type (
	Cell     = text.TextCell
	Frame    = text.TextFrame
	UITheme    = text.TextTheme
	Console  = text.TextConsole
	MenuItem = text.TextMenuItem
	List     = text.TextList
)

var (
	NewConsole = text.TextNewConsole
	IsHotkey   = text.TextIsHotkey
	MenuItems  = text.TextMenuItems
	Punct      = text.TextPunct
	Flow       = text.TextFlow
	Wrap       = text.TextWrap
	PadLeft    = text.TextPadLeft
	PriceLine  = text.TextPriceLine
	Commas     = text.TextCommas
)
