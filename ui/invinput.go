// invinput.go holds the inventory screen's input: its keys and its mouse.

package ui

import (
	"gobox/gpu"
	"gobox/rules"
	"gobox/ui/input"
	"image"
	"slices"
	"strings"
)

// inventoryKey handles a key on the inventory screen: Up/Down, Home/End the
// items, the member keys the member, Tab (pad Y) the tab, PgUp/PgDn (the
// shoulders) a page, Enter (pad A) the focused word or the
// item; false leaves it to the game (F10, ?, Ctrl-F).
func (p *Presenter) inventoryKey(k uint16) bool {
	inv := &p.inv
	if inv.ask != nil {
		switch i, back, _ := MenuNav(actOf(k), k, &inv.askSel, []byte{'Y', 'N'}); {
		case i == 0:
			inv.do(inv.ask.Kind)
		case i == 1 || back:
			inv.ask = nil
		}
		return true
	}
	order := inv.order()
	step := func(d int) {
		if len(order) == 0 {
			return
		}
		i := slices.Index(order, pick{inv.ware != rules.NoItem, max(inv.ware, inv.sel)})
		if i < 0 && d < 0 {
			i = 0
		}
		inv.choose(order[(i+d+len(order))%len(order)].ware, order[(i+d+len(order))%len(order)].i)
	}
	a := actOf(k)
	switch a {
	case input.PadA, input.Space: // pad A, Space: as Enter
		a = input.Enter
	}
	m, member := memberKey(k, inv.member, inv.count, true)
	switch {
	case a == input.Esc && inv.carry != nil: // the stack goes back where it was
		inv.carry = nil
	case a == input.Esc:
		inv.do(ActClose)
	case a == input.Minimap && inv.up.kind == invOwn: // Tab, pad Y: the sheet (they toggle)
		inv.do(ActSheet)
	case a == input.Minimap: // the shop and the loot: the next tab of the highlighted item's list
		pn := &inv.list
		if inv.ware != rules.NoItem {
			pn = &inv.wares
		}
		pn.tab, pn.top = (pn.tab+1)%len(tabNames), 0
	case (a == input.InputLeft || a == input.InputRight) && inv.up.kind != invOwn: // the shop: Left the goods, Right the member's items
		for _, o := range order {
			if o.ware == (a == input.InputLeft) {
				inv.choose(o.ware, o.i)
				break
			}
		}
	case a == input.Up:
		step(-1)
	case a == input.Down:
		step(1)
	case a == input.Home && len(order) > 0: // Home, End: the first and last item
		inv.choose(order[0].ware, order[0].i)
	case a == input.End && len(order) > 0:
		inv.choose(order[len(order)-1].ware, order[len(order)-1].i)
	case (a == input.PgUp || a == input.PgDn) && len(order) > 0: // PgUp/PgDn, the pad's shoulders: a page of the highlighted pane
		pn := &inv.list
		if inv.ware != rules.NoItem {
			pn = &inv.wares
		}
		sign := 1
		if a == input.PgUp {
			sign = -1
		}
		d := sign * max(1, pn.rows)
		i := max(0, slices.Index(order, pick{inv.ware != rules.NoItem, max(inv.ware, inv.sel)}))
		j := min(max(i+d, 0), len(order)-1)
		for order[j].ware != order[i].ware { // within the pane
			j -= sign
		}
		inv.choose(order[j].ware, order[j].i)
	case a == input.Enter && inv.focus > 0:
		if w := inv.words[inv.focus-1]; w.on {
			inv.do(w.act)
		}
	case a == input.Enter && inv.ware != rules.NoItem:
		inv.do(ActBuy)
	case a == input.Enter && inv.up.kind != invOwn && inv.up.sell:
		inv.do(ActSell)
	case a == input.Enter && inv.up.kind == invOwn:
		inv.do(ActReady)
	case member:
		if inv.up.switchable {
			inv.acts = append(inv.acts, Action{Kind: ActMember, N: m})
		}
	case k > ' ' && k < 128:
		for _, w := range inv.menu() {
			if w.on && strings.EqualFold(w.text[:1], string(rune(k))) {
				inv.do(w.act)
			}
		}
	default:
		return false
	}
	return true
}

// pick is an item on the inventory screen: a ware of the shop or the member's.
type pick struct {
	ware bool
	i    rules.ItemHandle
}

// order is the items as the screen shows them: the goods or the doll, then the list.
func (inv *inventory) order() []pick {
	var order []pick
	for _, it := range inv.wares.shown(inv.up.wares) {
		order = append(order, pick{true, it.Handle})
	}
	for _, it := range inv.up.doll {
		if it != nil {
			order = append(order, pick{false, it.Handle})
		}
	}
	for _, it := range inv.list.shown(inv.up.pool) {
		order = append(order, pick{false, it.Handle})
	}
	return order
}

// InventoryMouse handles the pointer at screen pixel (x, y) with the left
// button down or not, the right one just pressed and the wheel's movement.
// A click on a stack picks the whole of it up on the cursor, the next
// click drops it: onto the doll it is readied, into the party inventory it
// is unreadied, back where it came from it is put back; a right click
// picks up or drops one of a stack, with Shift half of it; Shift-click
// readies or unreadies the whole stack at once, as does dragging it onto
// the doll or off it; a right click elsewhere closes the screen. The arrows
// are the previous and next member, the menu row acts on the highlighted
// item; the wheel and the scroll bar scroll.
func (p *Presenter) InventoryMouse(x, y int, down, right bool, wheel float64) {
	inv := &p.inv
	if inv.k == 0 || inv.up == nil || p.q != nil { // a question over the items takes the keys, not the pointer
		return
	}
	moved := image.Pt(x, y) != inv.mouseAt // only a moved pointer takes the question's highlight
	inv.mouseAt = image.Pt(x, y)
	inv.hover = image.Pt(int(float64(x)/inv.kx), int(float64(y)/inv.ky))
	pressed, released := down && !inv.held, !down && inv.held
	inv.held = down

	// the wheel: a line a step, at most one a pause; the rest of a fling is dropped
	pn := &inv.list // the list under the pointer
	if inv.up.kind != invOwn && inv.hover.X < inv.list.at.Min.X {
		pn = &inv.wares
	}
	if inv.wheel += wheel; (inv.wheel >= 1 || inv.wheel <= -1) && now().Sub(inv.wheeled) >= wheelPause {
		if inv.wheel > 0 { // wheel up: up, as every list
			pn.top--
		} else {
			pn.top++
		}
		pn.clamp()
		inv.wheel, inv.wheeled = 0, now()
	} else if now().Sub(inv.wheeled) < wheelPause {
		inv.wheel = 0
	}

	if inv.ask != nil { // the question: its Yes and No (hover highlights), else any click is No
		inv.press, inv.dragging, inv.scrolling = nil, false, nil
		i := slices.IndexFunc(inv.askAt, image.Pt(x, y).In)
		if i >= 0 && moved {
			inv.askSel = i
		}
		switch {
		case pressed && i == 0:
			inv.do(inv.ask.Kind)
		case pressed || right:
			inv.ask = nil
		}
		return
	}
	if sc := inv.scrolling; sc != nil && down && sc.scroll.Dy() > 0 {
		sc.top = ScrollBar{R: sc.scroll, N: sc.n, Rows: sc.rows, Top: sc.top}.To(inv.hover.Y)
	}
	if inv.press != nil && down && !inv.dragging {
		d := inv.hover.Sub(inv.pressAt)
		inv.dragging = d.X*d.X+d.Y*d.Y > 9
	}
	shift := gpu.IsKeyPressed(gpu.KeyShift)
	if pressed && shift && inv.up.kind == invOwn && inv.hover.In(inv.face) { // Shift-click on the portrait: the next head
		inv.acts = append(inv.acts, Action{Kind: ActHead, N: -1})
		return
	}
	if pressed && inv.up.kind == invOwn && inv.carry == nil && inv.hover.In(inv.face) { // a click on the portrait: the sheet, as Tab
		inv.do(ActSheet)
		return
	}
	if right && inv.up.kind == invOwn { // a stack on the cursor: one of it (Shift: half) is dropped where the pointer is; else one of the stack under the pointer (Shift: half) is picked up; elsewhere the screen closes
		switch c := inv.carry; {
		case c != nil && shift:
			inv.put(max(1, c.n/2))
		case c != nil:
			inv.put(1)
		default:
			for _, l := range inv.lines {
				if inv.hover.In(l.r) {
					if shift {
						inv.pickUp(&l, max(1, max(l.item.Count, 1)/2))
					} else {
						inv.pickUp(&l, 1)
					}
					return
				}
			}
			inv.do(ActClose)
		}
		return
	}
	if right { // in the shop a right click buys a ware or sells an item
		if inv.up.kind != invOwn {
			for _, l := range inv.lines {
				if inv.hover.In(l.r) && !l.doll {
					inv.choose(l.ware, l.item.Handle)
					if l.ware && inv.up.kind == invLoot {
						inv.do(ActBuy)
					} else if l.ware {
						inv.ask, inv.askSel = &Action{Kind: ActBuy, Item: l.item.Handle}, 0
					} else if inv.up.sell {
						inv.do(ActSell)
					}
					return
				}
			}
		}
		inv.do(ActClose)
		return
	}
	if pressed {
		for _, w := range inv.words {
			if inv.hover.In(w.r) && w.on {
				inv.do(w.act)
				return
			}
		}
		for _, sc := range []*pane{&inv.list, &inv.wares} {
			if inv.hover.In(sc.scroll) {
				inv.scrolling = sc
			}
		}
		for _, pn := range []*pane{&inv.list, &inv.wares} {
			for i, r := range pn.tabs {
				if inv.hover.In(r) {
					pn.tab, pn.top = i, 0
					return
				}
			}
		}
		if inv.carry != nil && inv.up.kind == invOwn { // the stack on the cursor is dropped where the pointer is
			inv.put(inv.carry.n)
			return
		}
		for _, l := range inv.lines {
			if !inv.hover.In(l.r) {
				continue
			}
			if inv.up.kind != invOwn && !l.doll && shift { // Shift-click: a quick trade, no question
				inv.choose(l.ware, l.item.Handle)
				if l.ware {
					inv.do(ActBuy)
				} else if inv.up.sell {
					inv.remove(ActSell, l.item.Handle)
				}
				return
			}
			if inv.up.kind == invOwn && shift { // Shift-click: the whole stack is readied or unreadied
				inv.choose(false, l.item.Handle)
				inv.do(ActReady)
				return
			}
			inv.press, inv.pressAt = &l, inv.hover
		}
		return
	}
	if !released {
		return
	}
	h, drag := inv.press, inv.dragging
	inv.press, inv.dragging, inv.scrolling = nil, false, nil
	if h == nil {
		return
	}
	partySide := inv.hover.X >= inv.list.at.Min.X // dropped on the party's side of the divider
	if h.ware {                                   // a click highlights a ware, a second one or a drag over to the party's items buys it
		was := inv.ware
		inv.choose(true, h.item.Handle)
		switch {
		case drag && partySide && inv.up.kind != invLoot:
			inv.ask, inv.askSel = &Action{Kind: ActBuy, Item: h.item.Handle}, 0 // Yes first; a drag asks first (loot is free: no question)
		case drag && partySide, !drag && was == h.item.Handle:
			inv.do(ActBuy)
		}
		return
	}
	inv.choose(false, h.item.Handle)
	if inv.up.kind != invOwn { // the shop: a drag over to the goods sells
		if drag && !partySide && inv.up.sell {
			inv.do(ActSell)
		}
		return
	}
	inv.pickUp(h, 0) // a click picks the whole stack up on the cursor (the next click drops it)
	if drag {        // a drag is a click here and a click where it ends: one path
		inv.put(inv.carry.n)
	}
}
