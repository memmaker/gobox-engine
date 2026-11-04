package present

import (
	"image"
	"image/color"

	"gobox/front"
	"gobox/ui"
)

// Theme is a game's look in the presenter: the font, the text colours by
// role and the ribbon border.
type Theme struct {
	Font    string // the ui/fonts pool name of the default font
	Palette Palette
	Border  Border

	HeadTitles bool // menu titles and the slot page's column heads in the Heading colour (else Text, the titles of the menus, and the Hotkey colour)
}

// Palette is the presenter's colours by role, as EGA indices.
type Palette struct {
	Background int // the page's black, a cell's ground
	Selected   int // the selected row's ground, a text selection (-1: none, the row marked by the Hotkey colour alone); any role -1: none
	Log        int // the battle log's text, light grey
	Off        int // a row that is off; greys behind poses and swatches
	Text       int // body text, values
	Heading    int // headings, info lines
	Warn       int // refusals, can't-afford prices, the note
	Mark       int // bullets
	Gold       int // prices, prompts, the footer, the cursor
	Hotkey     int // hotkeys, titles, labels on the borders
}

// Border is the ribbon border's 8x8 pieces as symbols: frame Pieces[p]
// of block Block of DAX file File for each place p (ui.TopLeft ..
// ui.Right), decoded in the EGA colours (a frame below 0: no piece). A
// game that sets View.Border hands its pieces in instead.
type Border struct {
	File   string
	Block  int
	Pieces [ui.Pieces]int
	Caps   []int  // frames of the end caps (ui.LeftTopCap ..), 4 entries; absent or below 0: none, the side piece
	Mask   *uint8 // the colour index left transparent (as Picture.Mask; nil: none)

	// Pics, when set, give the pieces as DAX picture refs (Picture.File, Block,
	// Frame, Mask) by place instead of File's symbols; a ref with File "": no piece.
	Pics    [ui.Pieces]front.Picture
	CapPics [4]front.Picture // the caps' refs, as Pics

	// Edges, when set, give an edge's cells one by one as frames of File's
	// symbols, corners included: "top" and "bottom" (the screen's top and
	// bottom rows), "rule" (any other horizontal edge), "left", "right". An
	// edge of another length keeps the first and last entries at its ends and
	// repeats the middle ones (the original's cracked borders); a missing
	// edge draws the Pieces.
	Edges map[string][]int
}

// DefaultTheme is por's look.
var DefaultTheme = Theme{
	Font:    "Dark_Queen_of_Krynn",
	Palette: Palette{Background: 0, Selected: 1, Log: 7, Off: 8, Text: 10, Heading: 11, Warn: 12, Mark: 13, Gold: 14, Hotkey: 15},
	Border:  Border{Pieces: [ui.Pieces]int{0x114, 0x114, 0x114, 0x114, 0x116, 0x116, 0x115, 0x115}}, // por hands its pieces in (View.Border): its symbols are in the game's palette
}

var theme Theme

func init() { applyTheme(DefaultTheme) }

// applyTheme makes t the presenter's look.
func applyTheme(t Theme) {
	theme = t
	q := t.Palette
	inkBlack, inkBlue, inkGrey, inkDark, inkGreen, inkCyan, inkRed, inkMagenta, inkYellow, inkWhite = q.Background, q.Selected, q.Log, q.Off, q.Text, q.Heading, q.Warn, q.Mark, q.Gold, q.Hotkey
	green, white, cyan, gray, dark, red = col(inkGreen), col(inkWhite), col(inkCyan), col(inkGrey), col(inkDark), col(inkRed)
	menuTheme = ui.MenuTheme{
		Text: col(inkGreen), Hot: col(inkWhite), Off: col(inkDark), Sel: col(inkBlue),
		Gold: col(inkYellow), Head: col(inkCyan), Red: col(inkRed), Foot: col(inkYellow),
	}
	if t.HeadTitles {
		menuTheme.Title = col(inkCyan)
	}
}

// borderEdges is Theme.Border.Edges decoded, by edge (nil: none).
func borderEdges() map[string][]image.Image {
	b := theme.Border
	if b.File == "" || len(b.Edges) == 0 {
		return nil
	}
	mask := -1
	if b.Mask != nil {
		mask = int(*b.Mask)
	}
	out := map[string][]image.Image{}
	for k, fs := range b.Edges {
		for _, f := range fs {
			out[k] = append(out[k], daxMasked(b.File, b.Block, f, mask))
		}
	}
	return out
}

// edgeCell is the entry of an m-entry edge sequence for cell i of n: the
// ends fixed, the middle entries repeated.
func edgeCell(i, n, m int) int {
	switch {
	case m <= 2 || i == 0:
		return min(i, m-1)
	case i == n-1:
		return m - 1
	}
	return (i-1)%(m-2) + 1
}

// borderPieces is v's border pieces, or Theme.Border's decoded when the
// game hands none in.
func borderPieces(v front.View) [ui.Places]image.Image {
	if v.Border != ([8]image.Image{}) {
		var t [ui.Places]image.Image
		copy(t[:], v.Border[:])
		copy(t[ui.LeftTopCap:], v.BorderCaps[:])
		return t
	}
	var t [ui.Places]image.Image
	pics := append(theme.Border.Pics[:], theme.Border.CapPics[:]...)
	for i, pc := range pics {
		if pc.File != "" {
			if m := picImage(pc, [16]color.RGBA{}); m != nil {
				t[i] = m
			}
		}
	}
	if theme.Border.File == "" {
		return t // the zero value with no Pics: no pieces, as v.Border was
	}
	frames := append(theme.Border.Pieces[:], theme.Border.Caps...)
	for i, f := range frames {
		if f >= 0 && t[i] == nil {
			mask := -1
			if theme.Border.Mask != nil {
				mask = int(*theme.Border.Mask)
			}
			if m := daxMasked(theme.Border.File, theme.Border.Block, f, mask); m != nil {
				t[i] = m
			}
		}
	}
	return t
}
