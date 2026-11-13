package ui

import (
	"image"
	"image/color"
	"strconv"

	"gobox/game"
)

// Theme is a game's look in the presenter: the font, the text colours by
// role and the ribbon border.
type Theme struct {
	Font   string // the embedded font (asset.Get) of a game without one pooled
	Border Border
}

// fontPalette is every game's font colours, gobox's (no game overrides them).
var fontPalette = Palette{Background: 0, Selected: 1, Log: 7, Off: 8, Text: 10, Heading: 11, Warn: 12, Mark: 13, Gold: 14, Hotkey: 15}

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
// game that has art "border/<place>" hands its pieces in instead.
type Border struct {
	File   string
	Block  int
	Pieces [Pieces]int
	Caps   []int  // frames of the end caps (ui.LeftTopCap ..), 4 entries; absent or below 0: none, the side piece
	Mask   *uint8 // the colour index left transparent (as Picture.Mask; nil: none)

	// Pics, when set, give the pieces as DAX picture refs (Picture.File, Block,
	// Frame, Mask) by place instead of File's symbols; a ref with File "": no piece.
	Pics    [Pieces]Picture
	CapPics [4]Picture // the caps' refs, as Pics

	// Edges, when set, give an edge's cells one by one as frames of File's
	// symbols, corners included: "top" and "bottom" (the screen's top and
	// bottom rows), "rule" (any other horizontal edge), "left", "right",
	// "divider" (a vertical edge between two boxes or windows). An
	// edge of another length keeps the first and last entries at its ends and
	// repeats the middle ones (the original's cracked borders); a missing
	// edge draws the Pieces.
	Edges map[string][]int

	// Sets are the frame sets' blocks by the game's symbol set
	// (game.Border.Sets; nil: Block); Amiga the frame under Amiga art
	// (game.Border.Amiga; nil: this one).
	Sets  []int
	Amiga *game.AmigaFrame
}

// DefaultTheme is por's look.
var DefaultTheme = Theme{
	Font:   "Dark_Queen_of_Krynn",
	Border: Border{Pieces: [Pieces]int{0x114, 0x114, 0x114, 0x114, 0x116, 0x116, 0x115, 0x115}}, // por hands its pieces in (Border): its symbols are in the game's palette
}

var theme Theme

func init() { applyTheme(DefaultTheme) }

// applyTheme makes t the presenter's look.
func applyTheme(t Theme) {
	theme = t
	clear(amigaScreens) // the last theme's frame
	q := fontPalette
	inkBlack, inkBlue, inkGrey, inkDark, inkGreen, inkCyan, inkRed, inkMagenta, inkYellow, inkWhite = q.Background, q.Selected, q.Log, q.Off, q.Text, q.Heading, q.Warn, q.Mark, q.Gold, q.Hotkey
	green, white, cyan, gray, dark, red = col(inkGreen), col(inkWhite), col(inkCyan), col(inkGrey), col(inkDark), col(inkRed)
	menuTheme = MenuTheme{
		Text: col(inkGreen), Hot: col(inkWhite), Off: col(inkDark), Sel: col(inkBlue),
		Gold: col(inkYellow), Head: col(inkCyan), Red: col(inkRed), Foot: col(inkYellow),
	}
}

// borderEdges is Theme.Border.Edges decoded for symbol set set, by edge
// (nil: none): the Amiga frame's cells under Amiga art (borderAmiga).
func borderEdges(set int) map[string][]image.Image {
	b := theme.Border
	if a := b.Amiga; a != nil && amigaShown() {
		_, edges := amigaFrame(a, set)
		return edges
	}
	if b.File == "" || len(b.Edges) == 0 {
		return nil
	}
	mask := -1
	if b.Mask != nil {
		mask = int(*b.Mask)
	}
	block := setBlock(b.Sets, b.Block, set)
	out := map[string][]image.Image{}
	for k, fs := range b.Edges {
		for _, f := range fs {
			out[k] = append(out[k], daxMasked(b.File, block, f, mask))
		}
	}
	return out
}

// setBlock is the block of frame set set (sets: the blocks by set; nil or
// a set past them: block).
func setBlock(sets []int, block, set int) int {
	if set >= 0 && set < len(sets) {
		return sets[set]
	}
	return block
}

// windowBorder is the window border's pieces by place, and whether the
// game handed them in (else they are Theme.Border's).
type windowBorder struct {
	pieces [Places]image.Image
	own    bool
	edges  map[string][]image.Image // Theme.Border's edges (borderEdges; nil: none)
}

// windowBorder is the border the state's derived reads hold (derived); with
// no game, Theme.Border's.
func (p *Presenter) windowBorder() windowBorder {
	if p.eng == nil {
		return borderOf(nil, 0)
	}
	return p.eng.r.derived().border
}

// borderOf is the game's border pieces from its art ("border/<place>",
// 0..7, ui.TopLeft ..; "border/cap/<n>", 0..3, ui.LeftTopCap ..; in the
// EGA's colours), or Theme.Border's decoded for symbol set set when the
// art has no first piece.
func borderOf(art map[string]game.Picture, set int) windowBorder {
	if pix(art["border/0"]).Pix != nil {
		var t [Places]image.Image
		for i := range LeftTopCap {
			t[i] = pixImage(pix(art["border/"+strconv.Itoa(i)]), egaInk)
		}
		for i := range 4 {
			t[LeftTopCap+i] = pixImage(pix(art["border/cap/"+strconv.Itoa(i)]), egaInk)
		}
		return windowBorder{pieces: t, own: true}
	}
	return windowBorder{pieces: themeBorder(set), edges: borderEdges(set)}
}

// themeBorder is Theme.Border's pieces of symbol set set, decoded: the
// Amiga frame's cells under Amiga art (amigaFrame).
func themeBorder(set int) [Places]image.Image {
	var t [Places]image.Image
	if a := theme.Border.Amiga; a != nil && amigaShown() {
		pieces, _ := amigaFrame(a, set)
		copy(t[:], pieces[:])
		return t
	}
	pics := append(theme.Border.Pics[:], theme.Border.CapPics[:]...)
	for i, pc := range pics {
		if pc.File != "" {
			if m := picImage(pc, [16]color.RGBA{}); m != nil {
				t[i] = m
			}
		}
	}
	if theme.Border.File == "" {
		return t // the zero value with no Pics: no pieces
	}
	frames := append(theme.Border.Pieces[:], theme.Border.Caps...)
	for i, f := range frames {
		if f >= 0 && t[i] == nil {
			mask := -1
			if theme.Border.Mask != nil {
				mask = int(*theme.Border.Mask)
			}
			if m := daxMasked(theme.Border.File, setBlock(theme.Border.Sets, theme.Border.Block, set), f, mask); m != nil {
				t[i] = m
			}
		}
	}
	return t
}
