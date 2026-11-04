// fullmap.go is the presenter's full map: M, the command or a click on the
// minimap opens it; its cursor, the cell's menu (Walk, Mark, Remove) and
// the label being typed are the presenter's, the engine's FullMapView
// draws it and WalkTo, Mark and Unmark act.

package present

import (
	"gobox/front"
	"image"

	"gobox/input"
	"gobox/ui"
)

// fullMapUI is the full map up.
type fullMapUI struct {
	cursor image.Point
	menu   bool    // the cursor cell's menu is open (drawMapMenu)
	label  *string // the label being typed (nil: none)
	cur    int     // ... its caret
}

// openMap puts the full map up over the 3D view, the cursor on the
// party; where it is not allowed the engine says "Not Here".
func (p *Presenter) openMap(v front.View) {
	if p.eng == nil || !v.Exploring || mapBarred(v) {
		return
	}
	if p.eng.FullMapView(image.Pt(v.X, v.Y), nil) == nil {
		p.call(func(e front.Game) { e.Refuse("Not Here") })
		return
	}
	p.fullMap = &fullMapUI{cursor: image.Pt(v.X, v.Y)}
}

// fm is the full map as the engine draws it; nil with none up.
func (p *Presenter) fm() *front.FullMap {
	if p.fullMap == nil || p.eng == nil {
		return nil
	}
	return layFullMap(p.eng.FullMapView(p.fullMap.cursor, p.fullMap.label), p.shownView.Cells)
}

// Full map geometry: 16x16 squares as large as the page's height under
// the title row and above the message row allows, leaving fmLegendW text
// columns to the right for the legend (11 pixels a square on 40x25).
const (
	fmX, fmY  = 0, 8
	fmLegendW = 18
)

// layFullMap places fm on a page of cells (40x25 at least): the map's
// corner and square size, the legend's place and the lines that fit.
func layFullMap(fm *front.FullMap, cells image.Point) *front.FullMap {
	if fm == nil {
		return nil
	}
	cols, rows := max(cells.X, 40), max(cells.Y, 25)
	fm.Cell = max(1, min((rows-3)*8, (cols-fmLegendW)*8)/16)
	fm.At = image.Pt(fmX, fmY)
	fm.LegendAt = image.Pt((16*fm.Cell+7)/8, 2)
	fm.Legend = fm.Legend[:min(len(fm.Legend), rows-4)] // legend lines: rows 2..winBottom
	return fm
}

// mapItems is the cursor cell's menu: Walk, Mark where it has no label,
// Remove where the label is the player's.
func (p *Presenter) mapItems() *front.Menu {
	m := &front.Menu{Items: []front.MenuItem{{Text: "Walk", Key: 'W'}}}
	if fm := p.fm(); fm != nil && !fm.Labelled {
		m.Items = append(m.Items, front.MenuItem{Text: "Mark", Key: 'M'})
	}
	if fm := p.fm(); fm != nil && fm.Mine {
		m.Items = append(m.Items, front.MenuItem{Text: "Remove", Key: 'R'})
	}
	return m
}

// mapPick does item i of the cell's menu.
func (p *Presenter) mapPick(i int) {
	f := p.fullMap
	f.menu = false
	m := p.mapItems()
	if i < 0 || i >= len(m.Items) {
		return
	}
	c := f.cursor
	switch m.Items[i].Key {
	case 'W':
		p.call(func(e front.Game) {
			if e.WalkTo(c.X, c.Y) {
				p.fullMap = nil
			}
		})
	case 'M':
		s := ""
		f.label, f.cur = &s, 0
	case 'R':
		p.call(func(e front.Game) { e.Unmark(c.X, c.Y) })
	}
}

// mapMove moves the cursor to square c, on the map.
func (p *Presenter) mapMove(c image.Point) {
	p.fullMap.cursor = image.Pt(min(max(c.X, 0), 15), min(max(c.Y, 0), 15))
}

// mapKey takes key k on the full map: a label being typed takes it, then
// the cell's menu; else the arrows move the cursor, Enter opens the menu,
// Esc or M closes the map.
func (p *Presenter) mapKey(k uint16) {
	f := p.fullMap
	if f.label != nil {
		switch done, cancel, _ := editKey(k, f.label, &f.cur, 0x20, false); {
		case done:
			c, s := f.cursor, *f.label
			p.call(func(e front.Game) { e.Mark(c.X, c.Y, s) })
			f.label = nil
		case cancel:
			f.label = nil
		}
		return
	}
	a := actOf(k)
	if kp := keypad(k); kp != input.None {
		a = kp
	}
	if f.menu {
		switch i, back, _ := ui.MenuNav(actOf(k), k, &p.sel, ui.RowKeys(itemRows(p.mapItems()))); {
		case back:
			f.menu = false
		case i >= 0:
			p.mapPick(i)
		}
		return
	}
	d := map[input.Act]image.Point{input.Up: {0, -1}, input.Down: {0, 1}, input.Left: {-1, 0}, input.Right: {1, 0}}[a]
	switch {
	case d != (image.Point{}):
		p.mapMove(f.cursor.Add(d))
	case a == input.Enter || a == input.Space || a == input.PadA:
		f.menu, p.sel = true, 0
	case a == input.Esc || a == input.Map:
		p.fullMap = nil
	}
}

// mapMouse takes the pointer on text cell (cx, cy) of the page: over a
// square the cursor follows it; a left click on a square opens its menu,
// on a legend line goes to its first square, elsewhere closes the menu or
// the map; the right button closes the menu or the map.
func (p *Presenter) mapMouse(cx, cy int, in, left, right bool) {
	f, fm := p.fullMap, p.fm()
	if fm == nil || f.label != nil {
		return
	}
	px, py := cx*8+4-fm.At.X, cy*8+4-fm.At.Y
	sq := image.Pt(px/fm.Cell, py/fm.Cell)
	onSq := in && px >= 0 && py >= 0 && sq.In(image.Rect(0, 0, 16, 16))
	switch i := cy - fm.LegendAt.Y; {
	case right && f.menu:
		f.menu = false
	case right:
		p.fullMap = nil
	case !left && onSq && !f.menu && image.Pt(cx, cy) != p.mapPointer: // the pointer moved over a square: the cursor follows (the arrows keep it while it rests)
		p.mapPointer = image.Pt(cx, cy)
		f.cursor = sq
	case !left || !in:
	case onSq:
		f.cursor, f.menu, p.sel = sq, true, 0
	case cx >= fm.LegendAt.X && i >= 0 && i < len(fm.Legend):
		f.cursor, f.menu = fm.Legend[i].First, false
	case f.menu:
		f.menu = false
	default:
		p.fullMap = nil
	}
}

// mapBarred: v's area is marked unmappable and the No Map Zones option is on.
func mapBarred(v front.View) bool { return v.NoMap && Host.NoMapZones }
