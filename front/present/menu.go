package present

import (
	"gobox/blobber"
	"gobox/ega"
	"gobox/ui"
	"image"

	"gobox/gpu"
)

// drawMenu draws a vertical menu in the original's look: a box with the
// text pane's red border, the items in the game's font, green on black with
// their hotkeys white, the rows off[i] (nothing to pick) grey (EGA 8), the
// selected one on the text selection (a blue EGA 1 ground). It goes at scale k,
// centred on cx, its top at top.
func (p *Presenter) drawMenu(dst *gpu.Image, glyph func(byte) [8]byte, title string, items []string, keys []byte, off []bool, sel int, k float64, cx, top int) {
	p.drawMenuIcons(dst, glyph, title, items, keys, off, nil, sel, k, cx, top)
}

// iconRow is the cells an item row with its atlas icon takes (ui.IconRow).
const iconRow = ui.IconRow

// menuTheme is the menu look: green on black, hotkeys white, rows off
// grey (EGA 8), the selection a blue EGA 1 ground; the colour spans gold,
// cyan and red, the footer gold.
var menuTheme = ui.MenuTheme{
	Text: ega.RGBA(10), Hot: ega.RGBA(15), Off: ega.RGBA(8), Sel: ega.RGBA(1),
	Gold: ega.RGBA(14), Head: ega.RGBA(11), Red: ega.RGBA(12), Foot: ega.RGBA(14),
}

// drawMenuIcons is drawMenu with an icon before each row of icons that is
// not nil (ui.MenuCanvas.Draw), in the theme and the presenter's menu state.
func (p *Presenter) drawMenuIcons(dst *gpu.Image, glyph func(byte) [8]byte, title string, items []string, keys []byte, off []bool, icons []image.Image, sel int, k float64, cx, top int) {
	st := ui.MenuStyle{Theme: menuTheme, Tiles: p.border.tiles, Glyph: glyph, RowGlyph: blobber.MenuGlyph, Tight: p.menuTight, Tip: p.tipping}
	if p.logGrey {
		st.Theme.Text = ega.RGBA(7) // the battle Log's light grey
	}
	if p.menuGrey {
		st.Tiles = p.border.grey
	}
	p.menuRects = p.mc.Draw(dst, st, title, items, keys, off, icons, sel, k, cx, top, p.menuRects[:0])
}
