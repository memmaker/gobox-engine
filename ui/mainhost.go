package ui

import (
	"image"

	"gobox/front"
	"gobox/gpu"
)

// MainHost is what the main menu (front.MainMenu) needs of the front end
// drawing it.
type MainHost interface {
	Send(k front.ActKind, n int) // an action
	// DrawMenu draws a menu of lines (keys the hotkeys) at K k, centred
	// on cx from top; MenuRects are its rows.
	DrawMenu(dst *gpu.Image, lines []string, keys []byte, sel int, k float64, cx, top int)
	MenuRects() []image.Rectangle
	// KeyWindow pops a question and its answers up; WinRects are its rows.
	KeyWindow(dst *gpu.Image, title string, lines []string, keys []byte, sel int)
	WinRects() []image.Rectangle
	DrawCards(dst *gpu.Image, exp bool) []image.Rectangle // the party's cards (exp: the way to the next level under each) and the tip
	PopUp() (up *bool, sel *int)                          // a pop-up question is up, its highlight
	Point(x, y int) (moved bool)                          // the pointer is at x, y: it moved there
}
