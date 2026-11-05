package front

import "image"

// CellRect is a rectangle of 8x8 text cells, both corners included. For a
// box it is the inside, as DrawBorder takes it: the border is the ring of
// cells around it (Framed).
type CellRect struct{ X1, Y1, X2, Y2 int }

// Pixels is the rectangle in screen pixels.
func (r CellRect) Pixels() image.Rectangle {
	return image.Rect(r.X1*8, r.Y1*8, r.X2*8+8, r.Y2*8+8)
}

// Framed is the box with its border.
func (r CellRect) Framed() CellRect { return CellRect{r.X1 - 1, r.Y1 - 1, r.X2 + 1, r.Y2 + 1} }

// Creation is the creation sheet's parts: the drop-down fields, and the
// cells a click on the name, an ability, the portrait or the icon takes.
type Creation struct {
	Fields    []Field // the sheet's fields in the game's order (ActSetField N): adnd1 FieldSex, FieldRace, FieldAlign, FieldClass
	Powers    []Field // the rule set's power picks (ActSetField N len(Fields)+i), off the page: the bar's Powers; Cur -1: none picked yet
	Name      CellRect
	Stats     []CellRect // the abilities' cells (ActStat N), STR .. CHA or the game's own
	StatNames []string   // their labels
	Head      CellRect   // the portrait's head (PickHead) ...
	Body      CellRect   // ... and body (PickBody)
	Icon      CellRect   // the combat icon (ActIconEdit)
	Scale     int        // the portrait's and the icon's whole scale: the engine draws them at 1x from Head's and Icon's corners
	Abandon   string     // what Esc asks before ActClose
	Minimums  string     // the minimums Minimal last rolled to (ActRoll N 1)
	PointBuy  bool       // the game takes bought abilities (Game.PointBuy, game.Data.PointBuy): a new character is asked Roll or Point Buy
}

// StatCol is the column an ability's value starts at after its label
// ("STR 18", "TECH 12").
func StatCol(label string) int { return max(5, len(label)+2) }

// The adnd1 creation's fields (ActSetField N).
const (
	FieldSex = iota
	FieldRace
	FieldAlign
	FieldClass
)

// Field is a drop-down of the creation sheet.
type Field struct {
	Name    string
	At      CellRect // its text on the sheet
	Options []Option
	Cur     int // the option it holds
}

// Option is one of a Field's values; one that does not Fit changes the
// fields depending on it.
type Option struct {
	Text string
	Fits bool
}

// Intro is the title waiting to be played (TitlePicture, TitleCredits,
// TitleFlash, TitleDone).
type Intro struct {
	Full      bool  // picture 2 and the credits follow picture 1 (when Pictures is nil)
	Pictures  []int // N > 0 picture N, -N credits page N-1 (see game.Intro) // the picture steps, each the n of TitlePicture(n) (nil: 1, with Full 1 and 2); the game starts a step's tune there
	Lightning bool  // picture 2 flashes (EGA graphics: palette entry 1)
}
