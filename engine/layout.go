package engine

// The text screen: the original's 40x25 (textCols, textRows and the rows
// and columns derived from them). The 3D screen's boxes (ViewWindow,
// TextPane) are the original's: gobox lays that screen out itself.
const (
	minCols  = 40
	minRows  = 25
	noteCols = 36 // a note (say) fits the bottom line's box with a blank each side
	msgTop   = 17 // first row of the message pane under the portrait
	// The 3D screen's text pane and party panel keep the original's edges
	// whatever the page's size: the remaster shows them as a pop-up and
	// cards.
	paneRight  = 0x26
	paneBottom = 0x16
)

// The page: the original's 40x25 text screen.
const (
	textCols   = minCols // characters per line
	textRows   = minRows
	lastCol    = textCols - 1
	winRight   = textCols - 2 // inner right edge of the framed text window
	winBottom  = textRows - 3 // its inner bottom row
	messageRow = textRows - 2 // the bottom message line
	promptRow  = textRows - 1 // the menu and prompt line
)

// The insides of the original's main screen boxes, at their
// 40x25 places whatever the page's size: the remaster lays that screen
// out itself (the text pane as a pop-up, the party as cards).
var (
	ViewWindow  = CellRect{X1: 1, Y1: 1, X2: 0xF, Y2: 0xF}
	InnerWindow = CellRect{X1: 3, Y1: 3, X2: 0xD, Y2: 0xD} // the picture frame
	TextPane    = CellRect{X1: 1, Y1: msgTop, X2: paneRight, Y2: paneBottom}
)

// paneRows is the text pane's rows.
const paneRows = paneBottom - msgTop + 1
