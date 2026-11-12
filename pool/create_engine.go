package pool

import (
	"errors"
	"gobox/core"
	gcreation "gobox/creation"
)

// tempSel is a selection selectTemp replaced: the selected member and the party's size.
type tempSel struct{ sel, size int }

// selectTemp makes c the selected character (DS:5D92 = c) although it is not
// in the party; unselectTemp restores the old selection.
func (e *Engine) selectTemp(cID core.CharID) {
	e.temps = append(e.temps, tempSel{sel: e.sel, size: e.partySize()})
	e.do(core.SeatTemp{Who: cID})
	e.sel = e.partySize() - 1
	e.tempSel = cID
}

func (e *Engine) unselectTemp() {
	t := e.temps[len(e.temps)-1]
	e.temps = e.temps[:len(e.temps)-1]
	e.do(core.TrimParty{Size: int32(t.size)})
	e.sel = t.sel
	e.tempSel = core.NoChar
}

// creationSheet draws the sheet for the creation: without level, experience,
// party items and status, with the combat icon under the portrait. A PNG
// dropped on the icon becomes the icon, as in the icon editor (one dropped
// on the portrait the portrait, as on the sheet in play: portraitPNG).
func (e *Engine) creationSheet(cID core.CharID) {
	ic := e.fact(cID, core.IconPicsFact{}).(core.IconPics)
	e.CharacterSheet()
	e.snap.iconBox = sheetIconBox()
	e.snap.iconDrop = func(b []byte) {
		e.creationEdit(func(cr *creation) error {
			if e.iconOK(b) {
				cr.d.D.Icon.PNG = b
			}
			return nil
		})
	}
	for p := 0; p < 2; p++ {
		r := iconRect(sheetIconCol(p), sheetIconRow())
		e.erase(r)
		if ic.PNG != nil { // a PNG: the presenter's, over the page at the window's resolution
			if sp := e.snap.sheet; sp != nil {
				sp.Icons[p] = iconPose(e.pngImage(ic.PNG), p, false)
			}
		} else if im := e.icons[e.fact(cID, core.IconFact{}).(core.Number)][p]; im != nil {
			e.picture(sheetIconCol(p)*8, sheetIconRow()*8, im.Width*8, im.Height, im.Frame(0), im.Mask, im, 0)
		}
	}
}

// CreateCharacter = ovr17:056C, redone as one screen (see the top of the
// file): the sheet of a sample character whose parts are picked in place;
// Keep saves it, Leave abandons it.
func (e *Engine) CreateCharacter() error {
	if e.creation != nil {
		return core.ErrNotNow
	}
	d, err := gcreation.New(e.Creation(), nil) // nil: the Engine's own seed draws
	if err != nil {
		return err
	}
	c := e.sheetDraft(d.D)
	e.selectTemp(c)
	e.pushCreation(d, c, "Abandon this character?")
	e.creation.unnamed = true
	return nil
}

// pushCreation (extra) puts the creation sheet up on draft d, drawn as c.
func (e *Engine) pushCreation(d *gcreation.Draft, cID core.CharID, abandon string) {
	cr := &creation{d: d, c: cID, abandon: abandon}
	e.creation = cr
	e.ext.creating = true
	e.pushRedraw(rdCreation)
	e.creationSheet(cID)
}

// sheetDraft is the character d draws as on the sheet (core.SheetDraft).
func (e *Engine) sheetDraft(d core.Draft) core.CharID {
	o, err := e.ruleset().Resolve(nil, core.SheetDraft{Draft: d}, nil)
	if err != nil {
		panic(err)
	}
	return o.Created
}

// creationReplaced: the creation's character c was replaced by a fresh
// one, which takes its place.
func (e *Engine) creationReplaced(cr *creation, cID core.CharID) {
	if cr.modify {
		e.do(core.ReplaceMember{At: int32(cr.idx), Who: cID})
		cr.d.D.Of = cID // the draft names the member now in the slot (ids name characters, not positions)
		return
	}
	e.unselectTemp()
	e.selectTemp(cID)
}

// creationDone ends a creation: Keep makes the draft (porCreation.Make),
// else it is abandoned.
func (e *Engine) creationDone(cr *creation, kept bool) error {
	if kept {
		_, _, err := cr.d.Keep()
		if !cr.modify {
			e.unselectTemp()
			e.ruleset().Dismiss(cr.c)
		}
		return err
	}
	if cr.modify {
		e.do(core.RestoreCopy{Who: cr.orig})
		e.do(core.ReplaceMember{At: int32(cr.idx), Who: cr.orig})
		e.applyItems(cr.orig) // 00BA:0BB8
		return nil
	}
	e.unselectTemp()
	e.ruleset().Dismiss(cr.c)
	return nil
}

// EndCreation ends the creation sheet: keep keeps the character (Keep),
// else it is abandoned; the party menu under it draws itself again.
func (e *Engine) EndCreation(keep bool) error {
	cr := e.creation
	if cr == nil {
		return core.ErrNotNow
	}
	err := e.creationDone(cr, keep) // the creation still up: charOf takes its record's DAX pictures
	e.popRedraw()
	e.creation, e.ext.creating = nil, false
	e.snap.iconDrop, e.snap.iconBox = nil, CellRect{}
	if e.menu.on {
		e.partyRedraw()
	}
	return err
}

// Roll rolls the creation's abilities, to the minimums typed (mins:
// "STR INT WIS DEX CON CHA") when minimal.
func (e *Engine) Roll(minimal bool, mins string) error {
	return e.creationEdit(func(cr *creation) error {
		if !minimal {
			return cr.d.Roll(core.RollFull)
		}
		e.do(core.SetMinimums{Text: mins})
		cr.d.D.Lows = e.Creation().Minimums(cr.d.D)
		return cr.d.Roll(core.RollMinimal)
	})
}

// PointBuy (extra: gobox's Point Buy) gives the created character the
// abilities bought, STR .. CHA, then the post-roll rules and hit points.
func (e *Engine) PointBuy(scores []int) error {
	if e.creation == nil {
		return errors.New("Point buy is for a new character")
	}
	if len(scores) != 6 {
		return errors.New("Point buy needs 6 scores")
	}
	bases := map[core.Ability]int32{}
	for i, a := range abilities { // scores are in Creation.StatNames' order: STR .. CHA
		bases[a] = int32(scores[i])
	}
	return e.creationEdit(func(cr *creation) error { return cr.d.PointBuy(bases) })
}

// Rename names the character created (0709:09E3, extra: case kept).
func (e *Engine) Rename(name string) error {
	return e.creationEdit(func(cr *creation) error {
		if cr.d.Rename(name); cr.d.D.Name != "" {
			cr.unnamed = false
		}
		return nil
	})
}

// SetStat sets the created character's ability i (STR .. CHA) as typed (STR as 18/xx).
func (e *Engine) SetStat(i int, typed string) error {
	if i < 0 || i >= len(abilities) {
		return nil
	}
	return e.creationEdit(func(cr *creation) error {
		if err := cr.d.SetScore(abilities[i], typed); !errors.Is(err, gcreation.ErrScore) {
			return err
		}
		return nil // baseline setStatTyped: what is not a value keeps the score, no message
	})
}

// SetField gives creation field f (FieldSex ..) its option x.
func (e *Engine) SetField(f, x int) error {
	return e.creationEdit(func(cr *creation) error {
		cr.d.SetField(core.FieldID(f), x)
		return nil
	})
}

// creationEdit changes the draft of the creation up (none: nothing); the
// sheet's record is made anew from it and drawn; a refusal comes back with
// it redrawn.
func (e *Engine) creationEdit(edit func(cr *creation) error) error {
	cr := e.creation
	if cr == nil {
		return nil
	}
	err := edit(cr)
	c := e.sheetDraft(cr.d.D)
	e.creationReplaced(cr, c)
	cr.c = c
	e.creationSheet(c)
	return err
}

// sheetScale is the creation sheet's portrait and combat icon scale: the
// 40x25 page's sizes times what the right of the page leaves room for (the
// combat stats reach column 37, the prompt row stays free).
func sheetScale() int { return max(1, min((winRight-37)/11, (winBottom-3)/14)) }

// sheetPortraitBox is the creation sheet's portrait box, against the right
// border: 11x11 cells at 1x (the head 5 rows, the body 6), times sheetScale.
func sheetPortraitBox() CellRect {
	s := sheetScale()
	return CellRect{X1: winRight - 11*s + 1, Y1: 1, X2: winRight, Y2: 11 * s}
}

// Where the sheet shows the combat icon's poses: under the portrait box.
func sheetIconRow() int { return 2 + 11*sheetScale() }

// sheetIconCol is the column of pose p's icon under the portrait box, as
// the engine draws it (at 1x: the presenter scales it by sheetScale from
// the icon box's corner).
func sheetIconCol(p int) int { return winRight - 10*sheetScale() + 1 + 5*p }

// sheetIconBox is the icons' cells at sheetScale.
func sheetIconBox() CellRect {
	s := sheetScale()
	x, y := sheetIconCol(0), sheetIconRow()
	return CellRect{X1: x, Y1: y, X2: x + 8*s - 1, Y2: y + 3*s - 1}
}

// creation is the creation sheet up: c, a character of kind k.
type creation struct {
	d       *gcreation.Draft // the draft (gobox's one flow); c is drawn from it (charOf)
	c       core.CharID
	abandon string
	asking  bool // Esc or Exit asked abandon (Pending); Yes abandons, No back to the sheet
	unnamed bool // a new character no name was typed for: Keep asks "Character name:" first (ovr17's name prompt, BUG-CLASSES 59)
	naming  bool // Keep asked the name (Pending); the typed name keeps it, Esc back to the sheet
	modify  bool // ModifyCharacter's: the party member idx (was: as it was), not a new one
	idx     int
	orig    core.CharID // the member modified, as it was held (core.HoldCopy)
}
