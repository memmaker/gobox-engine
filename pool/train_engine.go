package pool

import (
	"gobox/asset"
	"gobox/core"
	"gobox/core/goldbox"
	gcreation "gobox/creation"
	"gobox/ega"
	"slices"
)

// ModifyCharacter = ovr17:209E, redone (extra) as the creation sheet on
// the selected character with no experience of its own (as the original,
// Felix's R3 ruling 16; ErrNotNow else): Roll, Minimal, the fields
// and the pictures change it as at the creation; Keep keeps it, Esc puts
// the old character back. (The original raised and lowered the abilities
// and hit points within the race and class limits, and only for a
// character without experience.)
func (e *Engine) ModifyCharacter(n int) error {
	if e.partySize() == 0 || e.creation != nil {
		return core.ErrNotNow
	}
	c, i := e.actor(n), e.sel
	if !e.ruleset().Check(core.Modifiable{Who: c}, nil).Yes { // no experience of its own only (Felix's R3 ruling 16)
		return core.ErrNotNow
	}
	e.do(core.HoldCopy{Who: c})
	d, err := gcreation.Modify(e.Creation(), nil, c)
	if err != nil {
		return err
	}
	e.pushCreation(d, c, "Discard the changes?")
	cr := e.creation
	cr.modify, cr.idx, cr.orig = true, i, c
	return nil
}

// OpenTraining = ovr17:29AC: the training hall's offer for member n (n <
// 0: the selected one), from the party menu, until CloseTraining (or the
// training is over).
func (e *Engine) OpenTraining(n int) error {
	if open, _ := e.hallOpen(); open || !e.menu.on || !slices.Contains(e.partyFunctions(), partyTraining) {
		return core.ErrNotNow
	}
	_, err := e.ruleset().Resolve(nil, core.OpenHall{Who: e.actor(n)}, nil)
	return err
}

// Train trains the member: a level in each class it may rise in; a mage
// level gained leaves the hall open for the new spell (Learn).
func (e *Engine) Train() error {
	if open, learning := e.hallOpen(); !open || learning {
		return core.ErrNotNow
	}
	o, err := e.ruleset().Resolve(nil, core.Train{Who: e.selected()}, nil)
	for _, f := range o.Effects {
		if n, ok := f.(core.Narrated); ok { // the hall's words for a Refused{NotTrainable}
			e.say(n.Text)
		}
	}
	if err != nil {
		return err
	}
	if open, _ := e.hallOpen(); !open {
		e.hallClosed()
	}
	return nil
}

// Learn adds spell id, one of the learnable, to the book; the hall closes.
func (e *Engine) Learn(id int) error {
	if open, learning := e.hallOpen(); !open || !learning {
		return core.ErrNotNow
	}
	if _, err := e.ruleset().Resolve(nil, core.Learn{Who: e.selected(), Spell: e.spellID(id)}, nil); err != nil {
		return err
	}
	return e.CloseHall()
}

// CloseTraining leaves the hall (not while the new spell is owed).
func (e *Engine) CloseTraining() error {
	if open, learning := e.hallOpen(); !open || learning {
		return core.ErrNotNow
	}
	return e.CloseHall()
}

// CloseHall leaves the hall for good.
func (e *Engine) CloseHall() error {
	if _, err := e.ruleset().Resolve(nil, core.CloseHall{}, nil); err != nil {
		return err
	}
	e.hallClosed()
	return nil
}

// hallClosed: the party menu under the hall draws itself again.
func (e *Engine) hallClosed() {
	if e.menu.on {
		e.partyRedraw()
	}
}

// addSaved = ovr17:306D's pick: saved character i of the Add picker
// joins the party (at most 6 characters and 8 with hired NPCs); chosen
// entries get a "* " in the list (extra: the list is the picker's cards).
// Done when the party is full.
func (e *Engine) addSaved(p *picker, i int) bool {
	names, count := p.labels, 0  // uninitialised until the first add in the original
	defer e.WriteCharList(names) // ovr18:01B8 (extra: on each pick)
	if id := p.chars[i]; id != core.NoChar && (names[i] == "" || names[i][0] != '*') {
		names[i] = goldbox.Cut("* "+names[i], textCols)
		var refused bool
		if refused, count = e.joinPicked(id); refused {
			names[i] = names[i][2:] // unmarked again, as in CotAB
		}
	}
	return count > 5 || e.worldWord(core.WordPartyCount) > 7
}

// joinPicked puts the picked character id in the party (ovr17:306D): the first into an empty
// party, the others unless already in, refused by the party's rules, or past the six
// players or eight members. refused is the rules' refusal; count the players now (0: none yet).
func (e *Engine) joinPicked(id core.CharID) (refused bool, count int) {
	if e.partySize() == 0 {
		e.must(core.EmptyParty{})
		e.join(id) // ovr18:1446
		return false, 0
	}
	dup := e.ruleset().Check(core.InParty{Who: id}, nil).Yes
	count = int(e.ruleset().Check(core.ControlledBefore{Who: id}, nil).Value)
	npc := e.fact(id, core.FactControl).(core.Number) >= 0x80
	refused = !dup && !e.ruleset().Check(core.CanJoin{Who: id}, nil).Yes
	if !dup && !refused && (!npc && count < 6 || npc && e.worldWord(core.WordPartyCount) < 8) {
		e.join(id)
		if !npc {
			count++
		}
	}
	return refused, count
}

// recolourIcon = 0464:1E5C (identity, to, 0, 0, icon): colour i becomes
// to[i] for i = 0..15 in turn, so a pixel already moved to a later colour
// moves again (3 -> 15 -> to[15]). Mask pixels are left alone. Amiga
// pixels (extra) go through amigaIconColour, and once only: the Amiga's
// routine (0x63240) looks each pixel's colour up a single time.
func recolourIcon(im *asset.Image, to [16]byte) {
	if im == nil {
		return
	}
	var f, am [16]byte
	for c := range f {
		x := byte(c)
		for i := x; i < 16; i++ {
			if x == i {
				x = to[i]
			}
		}
		f[c] = x
		am[amigaIconColour[c]] = amigaIconColour[to[c]]
	}
	for k, v := range im.Data {
		switch {
		case v < 16 && int(v) != im.Mask:
			im.Data[k] = f[v]
		case v >= ega.AmigaBase:
			im.Data[k] = v&^0xF | am[v&0xF]
		}
	}
}

// amigaIconColour (extra): the Amiga palette slot for each EGA colour
// number a character's icon colours hold. The Amiga icon editor (its slot
// table at 0x8B10E, against DS:3CF5) colours the parts in slots 9/1,
// 10/2, 11/3, 6/7, 5/8 and 15/14 where DOS uses 1/9 .. 7/15; the four
// left over pair by shade. So the DOS default colours give the Amiga
// default icon. The Amiga editor (0x1F11A) steps a part's colour through
// slots 1..15 only, never slot 0 (a second black): the designer offers
// those 15 in slot order on Amiga art.
var amigaIconColour = [16]byte{0, 9, 10, 11, 6, 4, 5, 15, 12, 1, 2, 3, 7, 13, 8, 14}
