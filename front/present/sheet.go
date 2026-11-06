package present

import (
	"gobox/front"
	"image"
	"image/color"
	"image/draw"
	"slices"
	"strconv"
	"strings"

	"gobox/gpu"
	"gobox/ui"

	"gobox/input"
)

// The character sheet (View.Sheet): the page is the engine's; its View
// menu is the presenter's.
func (p *Presenter) sheetItems(v front.View) []ui.MenuRow {
	var it []ui.MenuRow // the powers, then the reactions' policies (each pick turns to the next): Tab is the inventory, Esc leaves
	for _, pw := range v.Sheet.Powers {
		it = append(it, ui.MenuRow{Text: pw.Name, Key: pw.Name[0], Do: func() { p.send(pw.Kind, -1) }})
	}
	for _, r := range v.Sheet.Reactions {
		it = append(it, ui.MenuRow{Text: r.Name + " " + r.Policy.String(), Key: r.Name[0], Do: func() {
			p.call(func(e front.Game) { e.SetPolicy(r.ID, r.Policy+1) })
		}})
	}
	return it
}

// sheetUp: the sheet's menu is the page's question.
func sheetUp(v front.View) bool { return v.Sheet != nil && askMenu(v) == nil && v.Inventory == nil }

// sheetKey takes key k on the sheet: the menu's keys, the member keys another member.
func (p *Presenter) sheetKey(k uint16, v front.View) bool {
	items := p.sheetItems(v)
	if m, ok := memberKey(k, selectedIndex(v), len(v.Party), true); ok {
		p.send(front.ActMember, m)
		return true
	}
	if actOf(k) == input.Minimap { // Tab, pad Y: the inventory (they toggle)
		p.send(front.ActInventory, -1)
		return true
	}
	i, back, ok := ui.MenuNav(actOf(k), k, &p.sel, ui.RowKeys(items))
	switch {
	case back:
		return false // ActClose (keyAct)
	case i >= 0:
		items[i].Do()
	}
	return ok
}

// sheetCanvas is the sheet as the presenter draws it: at the inventory's
// scale and cells, the member on the left (name, classes, portrait, the
// HP and experience bars), the sheet's figures on the right, the powers
// on the bottom row (the toast in their place while it shows).
type sheetCanvas struct {
	img    *gpu.Image
	px     *image.RGBA
	w, h   int
	faceAt image.Rectangle // the portrait on the window (Shift-click)
}

// drawSheet draws the sheet (sheetUp) over the whole window.
func (p *Presenter) drawSheet(dst *gpu.Image, note Note, v front.View) {
	sp, i := v.SheetPage, selectedIndex(v)
	if i < 0 {
		return
	}
	m, sc := v.Party[i], &p.sheet
	p.setup(note, v)
	p.pageMenu = true
	b := dst.Bounds()
	k := max(1, min(b.Dx()/invW, b.Dy()/invH)) // the inventory's scale and cells
	cw, ch := max(invW, b.Dx()/k)/8, max(invH, b.Dy()/k)/8
	if w, h := cw*8, ch*8; gpu.Sized(&sc.img, w, h) {
		sc.w, sc.h, sc.px = w, h, image.NewRGBA(image.Rect(0, 0, w, h))
	}
	draw.Draw(sc.px, sc.px.Rect, image.Black, image.Point{}, draw.Src)
	text := func(s string, x, y int, c color.RGBA) { ui.DrawText(sc.px, note.Glyph, s, x, y, c) }

	// the member: the pane the inventory shares, then HP and experience under the portrait
	lx, lw := 8, (leftCells-2)*8
	vr := lx + lw - 8 // the values' right edge: a cell clear of the pane's border (a thick one, COTAB's, overhangs it)
	face, png := p.memberPane(sc.px, note.Glyph, m, (ch-2)*8)
	y := face.Max.Y + 6
	hpY, expY := y+10, y+36
	hpc := green
	if m.HP < m.MaxHP {
		hpc = col(inkYellow)
	}
	text("HP", lx, y, white)
	hp := strconv.Itoa(m.HP) + "/" + strconv.Itoa(m.MaxHP)
	text(hp, vr-8*len(hp), y, hpc)
	if !sp.Creating && (sp.Exp > 0 || sp.Next > 0) { // a game without experience leaves both zero
		text("Exp", lx, y+26, white)
		e := strconv.Itoa(sp.Exp)
		if sp.Next > 0 {
			e += "/" + strconv.Itoa(sp.Next)
		}
		text(e, vr-8*len(e), y+26, green)
	}
	ly := y + 52 // under the bars: a key left, its value right
	if !sp.Creating && sp.Levels != "" {
		text("Level", lx, ly, white)
		text(sp.Levels, vr-8*len(sp.Levels), ly, green)
	}
	if sp.Status != "" { // the pane's last line, level with the AC/THAC0/Damage values
		sy := (ch-2)*8 - 10
		text("Status", lx, sy, white)
		text(sp.Status, vr-8*len(sp.Status), sy, green)
	}

	// the figures: who, a blank line, the abilities, Move, the rule set's
	// figures (a 40x25 page fits them all above AC/THAC0/Damage); names written out on the left, the numbers
	// right-aligned on one column (ones under ones), the exceptional
	// strength and the rule set's modifier after it. Each line has a
	// tooltip (tips).
	rx, rr := leftCells*8, (cw-1)*8
	last := 10 + len(sp.Rules) // the last figure's row: Movement, then the rule set's (BR2: tech and 8 career skills)
	rh := max(9, min(12, ((ch-2)*8-10-14-10-12)/max(last, 1))) // rows close up for a long list, so it all fits above AC/THAC0/Damage
	row := func(r int) int { return 12 + rh*r }
	num := rx + 19*8 // the numbers' right edge ("Healing Surges" and a 5-wide value clear), further right for a longer name (BR2's "Treat Light Wounds")
	for _, st := range sp.Rules {
		num = max(num, rx+8*(len(st.Name)+1+max(len(st.Value), 2)))
	}
	num = min(num, rr-8*len(" (+0)")-8*len("(00)")) // the modifiers' column still inside the border
	type tip struct {
		r     image.Rectangle
		title string
		help  []string
	}
	var tips []tip
	line := func(name, v string, r int, help []string) {
		shown := name
		if n := (num - 8*len(v) - 8 - rx) / 8; len(shown) > n { // a name too long for the column is cut, never written under its value
			shown = shown[:max(n, 0)]
		}
		text(shown, rx, row(r), white)
		text(v, num-8*len(v), row(r), green)
		tips = append(tips, tip{image.Rect(rx, row(r)-(rh-8)/2, rr, row(r)+8+(rh-8)/2), name, help})
	}
	text(sp.Sex+" "+sp.Race+", age "+strconv.Itoa(sp.Age), rx, row(0), white)
	text(sp.Alignment, rx, row(1), white)
	for j, l := range abilityNames {
		help := abilityHelp[j]
		if sp.Rules != nil {
			help = append(slices.Clone(help), "", "Yellow: the 4e modifier,", "(score - 10) / 2, rounded down.")
		}
		line(l, strconv.Itoa(sp.Stats[j]), 3+j, help)
		x := num + 8*len("(00)") // the modifiers on one column, past any exceptional strength
		if j == 0 {
			text(sp.Str00, num, row(3), green)
		}
		if sp.Rules != nil {
			m := strconv.Itoa(sp.Mods[j])
			if sp.Mods[j] >= 0 {
				m = "+" + m
			}
			text(" ("+m+")", x, row(3+j), col(inkYellow))
		}
	}
	line("Movement", strconv.Itoa(sp.Move), 10, // a blank line after the abilities
		[]string{"Squares walked in one", "combat round.", "", "Set by race; heavy armour", "and load slow it down."})
	vy := (ch-2)*8 - 10                                    // the last row inside the border
	fits := func(r int) bool { return row(r)+10 <= vy-14 } // a row clear of the AC/THAC0/Damage labels
	r := 11
	for j, st := range sp.Rules {
		if !fits(r+1) && j < len(sp.Rules)-1 { // a long list (BR2's skills) stops above the footer, the rest counted
			text("+"+strconv.Itoa(len(sp.Rules)-j)+" more", rx, row(r), white)
			r++
			break
		}
		line(st.Name, st.Value, r, st.Help)
		r++
	}
	for _, f := range extraFields(sp) { // the game's extra fields (POD fills them), after a blank line
		if f[1] != "" && fits(r+1) {
			text(f[0], rx, row(r+1), white)
			text(f[1], rr-8*len(f[1]), row(r+1), green)
			r++
		}
	}
	if sp.Effects != "" && fits(r+2) {
		text("Effects", rx, row(r+1), white)
		for j, l := range ui.WrapWords(sp.Effects, cw-leftCells-2) {
			if fits(r + 2 + j) {
				text(l, rx, row(r+2+j), green)
			}
		}
	}

	// AC, THAC0 and Damage on the pane's bottom: three columns, each label
	// over its value, centred in its column
	colW := (rr - rx) / 3
	for j, f := range [][2]string{{"AC", strconv.Itoa(sp.AC)}, {"THAC0", strconv.Itoa(sp.THAC0)}, {"Damage", sp.Damage}} {
		if f[1] == "" {
			continue // the game leaves it empty: no label
		}
		mid := rx + colW*j + colW/2
		text(f[0], mid-4*len(f[0]), vy-12, white)
		text(f[1], mid-4*len(f[1]), vy, green)
		tips = append(tips, tip{image.Rect(mid-4*8, vy-14, mid+4*8, vy+10), f[0], bottomHelp[j]})
	}

	// the powers on the bottom row; the toast takes it while it shows
	items := p.sheetItems(v)
	rowY := (ch - 1) * 8
	p.winRects = p.winRects[:0]
	var words []image.Rectangle
	if note.Show && note.Text != "" {
		noteLine(sc.px, note, rowY)
	} else {
		n := 0
		for _, it := range items {
			n += 8 * (len(it.Text) + 1)
		}
		x := (sc.w - n + 8) / 2
		p.sel = min(max(p.sel, 0), len(items)-1)
		for j, it := range items {
			r := image.Rect(x-4, rowY-2, x+8*len(it.Text)+4, rowY+10)
			if j == p.sel {
				if inkBlue >= 0 {
					draw.Draw(sc.px, r, image.NewUniform(col(inkBlue)), image.Point{}, draw.Src)
				}
			}
			text(it.Text[:1], x, rowY, white)
			text(it.Text[1:], x+8, rowY, green)
			words = append(words, r)
			x += 8 * (len(it.Text) + 1)
		}
	}

	sc.img.WritePixels(sc.px.Pix)
	t := p.border.tiles
	window(sc.img, t, gpu.GeoM{}, leftCells, ch-1, nil)
	var g gpu.GeoM
	g.Translate(float64(leftCells-1)*8, 0)
	window(sc.img, t, g, cw-leftCells+1, ch-1, nil)
	paneName(sc.img, note.Glyph, m.Name)
	fill := greenFill
	if m.HP*4 < m.MaxHP {
		fill = redFill
	}
	p.bar(sc.img, lx-5, hpY, lw+10, m.HP, m.MaxHP, fill)
	if !sp.Creating && m.NextExp > 0 {
		p.bar(sc.img, lx-5, expY, lw+10, m.Exp-m.LevelExp, m.NextExp-m.LevelExp, goldFill)
	}
	kx, ky := p.fit(dst, sc.img, sc.w, sc.h)
	sc.faceAt = image.Rect(int(float64(face.Min.X)*kx), int(float64(face.Min.Y)*ky), int(float64(face.Max.X)*kx), int(float64(face.Max.Y)*ky))
	if png != nil { // the portrait at the window's resolution, nearest neighbour (hiFace)
		op := &gpu.DrawImageOptions{}
		op.GeoM.Translate(float64(face.Min.X)*kx, float64(face.Min.Y)*ky)
		dst.DrawImage(p.hiFace(png, int(float64(face.Dx())*kx)), op)
	}
	if x, y := gpu.CursorPosition(); askMenu(v) == nil { // the tooltip of the line under the pointer
		pt := image.Pt(int(float64(x)/kx), int(float64(y)/ky))
		for _, t := range tips {
			if pt.In(t.r) && len(t.help) > 0 {
				w := image.Rect(int(float64(t.r.Min.X)*kx), int(float64(t.r.Min.Y)*ky), int(float64(t.r.Max.X)*kx), int(float64(t.r.Max.Y)*ky))
				p.drawTipOver(dst, t.title, t.help, k, w)
			}
		}
	}
	for _, r := range words { // the clicks, on the window
		p.winRects = append(p.winRects, image.Rect(int(float64(r.Min.X)*kx), int(float64(r.Min.Y)*ky), int(float64(r.Max.X)*kx), int(float64(r.Max.Y)*ky)))
	}
}

// memberPane draws the left pane's member part, the same on the sheet and
// the inventory so it stays put when Tab switches them: the classes under
// the top border (the name is on it: paneName), then the portrait as wide
// as the pane, or as tall as the room above the doll allows (bottom: the
// pane's inner bottom), aspect kept. The portrait (the PNG or the
// paper-doll head) is returned to be drawn at the window's resolution
// with hiFace: crisp pixels, never the canvas's stretch.
func (p *Presenter) memberPane(px *image.RGBA, glyph func(byte) [8]byte, m front.Member, bottom int) (face image.Rectangle, png image.Image) {
	p.inv.load()
	lx, lw := 8, (leftCells-2)*8
	y := 10
	for _, cl := range strings.Split(m.Class, "/") { // a class a line: three don't fit one
		ui.DrawText(px, glyph, cl, lx+(lw-8*len(cl))/2, y, green)
		y += 10
	}
	n := max(0, min(lw, bottom-y-4-p.inv.doll.Bounds().Dy()-8))
	face = image.Rect(lx+(lw-n)/2, y+4, lx+(lw+n)/2, y+4+n)
	png = m.Portrait
	if png == nil {
		png = p.portrait(m) // the paper-doll head
	}
	return face, png
}

// paneName writes the member's name on the left pane's top border, centred.
func paneName(dst *gpu.Image, glyph func(byte) [8]byte, name string) {
	lw := leftCells * 8
	onBorder(dst, glyph, (lw-8*(len(name)+2))/2, 0, true, []string{name}, []color.RGBA{white})
}

var abilityNames = [6]string{"Strength", "Intelligence", "Wisdom", "Dexterity", "Constitution", "Charisma"}

// abilityHelp and bottomHelp are the sheet's tooltips (drawSheet).
var abilityHelp = [6][]string{
	{"Physical power. Raises", "melee hits and damage and", "what you can carry.", "", "A fighter's 18 adds an", "exceptional (01-00) roll."},
	{"Reasoning and memory.", "Mages need it to learn", "and cast spells."},
	{"Insight and willpower.", "Clerics gain extra spells", "from a high score."},
	{"Agility and reflexes.", "Improves armour class and", "missile attacks."},
	{"Health and stamina. Adds", "hit points every level."},
	{"Presence and leadership.", "Paladins need 17 or more."},
}

var bottomHelp = [3][]string{
	{"Armour class: lower is", "better. Armour, shield,", "DEX and magic lower it."},
	{"To hit armour class 0: roll", "a d20 of THAC0 minus the", "target's AC or more.", "Lower is better."},
	{"Damage of the readied", "weapon, STR bonus included."},
}
