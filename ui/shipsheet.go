package ui

// The Buck Rogers ship sheet (game.ShipSheet): BR1 pO0F6_3346 with its
// figures pO0F6_1CF8 and weapons pO0F6_23F4(ship, 1, -1), BR2 pO099_3472
// with pO099_1E1D and pO099_2508(ship, 1, 255). Both are the same cells.

import (
	"strconv"
	"strings"

	"gobox/game"
)

var shipRanges = [...]string{game.RangeRamming: "ramming", game.RangeShort: "short", game.RangeMedium: "medium", game.RangeLong: "long", game.RangeExtreme: "extreme"}

// shipSheet draws s on (1,1)-(24,17), bg 8.
func (c *cellPage) shipSheet(s *game.ShipSheet) {
	c.fillCells(1, 1, 24, 17, 8)
	lab := toneInk[game.ToneWarn]
	if s.Own {
		lab = toneInk[game.ToneOwn]
	}
	loud := toneInk[game.ToneLoud]
	// fig is 1CF8 / 1E1D: n right-justified 5 wide plus its unit, fg15 on
	// bg8; a hidden figure (the foe's, unknown) is '?' in fg12.
	fig := func(x, y, n int, unit string, hidden bool) {
		t, fg := strconv.Itoa(n), loud
		if hidden && !s.Own && !s.Known {
			t, fg = "?", toneInk[game.ToneBad]
		}
		c.put(x, y, fg, 8, strings.Repeat(" ", max(0, 5-len(t)))+t+unit)
	}
	for i, r := range [...]struct{ l, v string }{{"name:", s.Name}, {"type:", s.Type}, {"race:", s.Race}} {
		c.put(1, 1+i, lab, 8, r.l)
		c.put(10, 1+i, loud, 8, r.v)
	}
	for i, r := range [...]struct {
		l, unit string
		n       int
	}{{"tonnage:", " tons", s.Tonnage}, {"length:", " feet", s.Length}, {"width:", " feet", s.Width}, {"cargo:", " tons", s.Cargo}} {
		c.put(1, 4+i, lab, 8, r.l)
		fig(10, 4+i, r.n, r.unit, false)
	}
	c.put(1, 9, lab, 8, "AC:")
	fig(7, 9, s.AC, "", true)
	c.put(13, 9, lab, 8, "crew:")
	fig(20, 9, s.Crew, "", true)
	c.put(1, 10, lab, 8, "speed:")
	fig(7, 10, s.Speed, "", true)
	c.put(13, 10, lab, 8, "react:")
	fig(20, 10, s.React, "", true)
	if s.Own {
		c.put(1, 11, lab, 8, "medical supplies:")
		fig(20, 11, s.MedSupplies, "", true)
	} else {
		c.put(1, 11, lab, 8, "range:")
		fg := loud
		if s.Range == game.RangeRamming {
			fg = toneInk[game.ToneWarn]
		}
		if int(s.Range) >= 0 && int(s.Range) < len(shipRanges) {
			c.put(10, 11, fg, 8, shipRanges[s.Range])
		}
	}
	// 23F4 / 2508 with a1 = 1: mounts 0..3 from row 14
	text := toneInk[game.ToneText]
	c.put(1, 13, text, 8, "Weapons:")
	c.put(13, 13, text, 8, "Ammo")
	c.put(18, 13, text, 8, "Reloads")
	for i, w := range s.Weapons[:min(4, len(s.Weapons))] {
		if w.Mounts <= 0 {
			continue
		}
		fg := text
		if w.Working != w.Mounts {
			fg = toneInk[game.ToneWarn]
		}
		fig(11, 14+i, w.Ammo, "", true)
		fig(17, 14+i, w.Reloads, "", true)
		c.put(1, 14+i, fg, 8, short8(w.Short))
		c.put(9, 14+i, fg, 8, rjust(w.Working, 2))
	}
}
