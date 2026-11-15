package ui

import (
	"image"
	"image/color"

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
const iconRow = IconRow

// menuTheme is the menu look: green on black, hotkeys white, rows off
// grey (EGA 8), the selection a blue EGA 1 ground; the colour spans gold,
// cyan and red, the footer gold.
var menuTheme MenuTheme // applyTheme

// drawMenuIcons is drawMenu with an icon before each row of icons that is
// not nil (ui.MenuCanvas.Draw), in the theme and the presenter's menu state.
func (p *Presenter) drawMenuIcons(dst *gpu.Image, glyph func(byte) [8]byte, title string, items []string, keys []byte, off []bool, icons []image.Image, sel int, k float64, cx, top int) {
	if drawnAt != nil {
		drawnAt(drawnMenu, dst, k)
	}
	st := MenuStyle{Theme: menuTheme, Tiles: p.border.val.tiles, Edges: edgesOf(p.border.val.tiles), Glyph: glyph, RowGlyph: MenuGlyph, Tight: p.menuTight, Tip: p.tipping}
	if p.logGrey {
		st.Theme.Text = col(inkGrey) // the battle Log's light grey
	}
	if a := p.inkAsk; a != nil && a.NoPick { // no highlight: only the hotkeys answer
		sel = -1
	}
	if a := p.inkAsk; a != nil { // the question's own colours (Ask's roles)
		for c, r := range map[*color.RGBA]Role{&st.Theme.Title: a.PromptInk, &st.Theme.Text: a.Ink, &st.Theme.Hot: a.HotInk} {
			if i, ok := roleInk(r); ok {
				*c = col(i)
			}
		}
	}
	if p.menuGrey {
		st.Tiles = p.border.val.grey
	}
	p.asksDrawn = true
	p.menuRects = p.mc.Draw(dst, st, title, items, keys, off, icons, sel, k, cx, top, p.menuRects[:0])
}

// roleInk is role r's palette index (false: RoleNone, the presenter's own).
func roleInk(r Role) (int, bool) {
	q := fontPalette
	i, ok := map[Role]int{RoleText: q.Text, RoleHeading: q.Heading, RoleSelected: q.Selected, RoleOff: q.Off,
		RoleWarn: q.Warn, RoleMark: q.Mark, RoleGold: q.Gold, RoleHotkey: q.Hotkey, RoleLog: q.Log,
		RoleBackground: q.Background}[r]
	return i, ok && i >= 0
}

// asking draws with v's question's colours (Ask's roles) while draw runs.
func (p *Presenter) asking(draw func()) {
	p.inkAsk = p.askUp()
	defer func() { p.inkAsk = nil }()
	draw()
}

// drawnKind is what drawnAt saw drawn: a menu (every menu, pop-up,
// question and tooltip: at K) or a box (a text pane, a picture or view
// window: its ribbon at a whole multiple of K) or a cell page (the
// credits: its glyphs and ribbon tiles at its whole scale).
type drawnKind int

const (
	drawnMenu drawnKind = iota
	drawnBox
	drawnPage
)

// drawnAt, when set (TestOneCellSize), sees the scale every menu and box
// is drawn at on dst.
var drawnAt func(what drawnKind, dst *gpu.Image, k float64)
