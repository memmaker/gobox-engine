package present

import (
	"image"

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
	Mask   *uint8 // the colour index left transparent (as Picture.Mask; nil: none)
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

// borderPieces is v's border pieces, or Theme.Border's decoded when the
// game hands none in.
func borderPieces(v front.View) [ui.Pieces]image.Image {
	if v.Border != ([ui.Pieces]image.Image{}) || theme.Border.File == "" {
		return v.Border
	}
	var t [ui.Pieces]image.Image
	for i, f := range theme.Border.Pieces {
		if f >= 0 {
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
