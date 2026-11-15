package rules1e

import (
	"gobox/rules"
	"slices"
)

// Spells: the book, aims, the spell menu, Knock (core/spells.go).

// checkSpells answers the Checks of the spells topic.
func (r *Rules) checkSpells(k rules.Check) (rules.Answer, bool) {
	switch k := k.(type) {
	case rules.SpellNumber:
		for n, id := range SpellIDs {
			if id == k.Spell {
				return rules.Answer{Yes: true, Value: int32(n)}, true
			}
		}
		return rules.Answer{}, true
	case rules.Memorized:
		for _, c := range r.party(k.Party) {
			for slot, sp := range c.Memorized {
				if r.S.T().hasSpell(sp) && SpellIDs[sp] == k.Spell {
					return rules.Answer{Yes: true, Who: r.ID(c), Value: int32(slot)}, true
				}
			}
		}
		return rules.Answer{}, true
	default:
		return rules.Answer{}, false
	}
}

// resolveSpells answers the Actions of the spells topic.
func (r *Rules) resolveSpells(sit rules.Situation, a rules.Action) (rules.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case rules.OpenCast: // camp's Cast opens
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		s.SpellCaster = nil // DS:6DB5
		s.Who = s.PartyIndex(c)
		if !s.CanStudy(1, a.NoMagic) {
			return rules.Outcome{}, true, rules.ErrNotNow
		}
	case rules.CampCast:
		c := r.chars[a.Who]
		n := slices.Index(SpellIDs[:], a.Spell)
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		if n <= 0 {
			return rules.Outcome{}, true, errBadValue
		}
		s.Who = s.PartyIndex(c)
		var quit bool
		s.CastSpell(SpellID(n), 0, 1, CastReply{Fight: s.Fight.On, Yes: a.Sure, On: r.chars[a.Target], In3D: a.In3D}, &quit)
	case rules.Learn:
		return rules.Outcome{}, true, r.learn(a)
	case rules.Cast:
		o, err := r.cast(sit, a)
		return o, true, err
	case rules.Maneuver:
		o, err := r.maneuver(sit, a)
		return o, true, err
	case rules.SpendKnock:
		if c := s.SpendKnock(); c != nil {
			return rules.Outcome{Effects: []rules.Effect{rules.KnockCast{Who: r.ID(c)}}}, true, nil
		}
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}

// factSpells answers the Facts of the spells topic.
func (r *Rules) factSpells(id rules.CharID, f rules.Fact) (rules.Value, bool, error) {
	s, c := r.S, r.chars[id]
	switch f := f.(type) {
	case rules.SpellByNumber:
		if f.Game && f.N > 0 && f.N <= 0xFF {
			id, err := s.T().spellID(byte(f.N))
			if err != nil {
				return nil, true, err
			}
			f.N = int(id)
		}
		if f.N < 0 || f.N > LastSpellID || !s.T().hasSpell(SpellID(f.N)) && f.N != 0 {
			return rules.SpellID(""), true, nil
		}
		return SpellIDs[f.N], true, nil
	case rules.CastCardFact: // no holder
		card := rules.CastCard{Name: s.SpellName(f.Number)}
		switch s.SpellRec(SpellID(f.Number), 7) {
		case 0:
			card.Aim = rules.CastAimCombatOnly
		case 2:
			card.Aim = rules.CastAimMember
		}
		return card, true, nil
	case rules.StudyFact:
		if c == nil {
			return nil, true, errNoChar
		}
		book, scrolls := s.SpellLists(c)
		return rules.Study{Book: int32(len(book)), Scrolls: int32(len(scrolls))}, true, nil
	case rules.LearnableFact:
		if c == nil {
			return nil, true, errNoChar
		}
		var l rules.SpellCards
		for _, sp := range s.Learnable(c) {
			l = append(l, rules.SpellCard{Spell: SpellIDs[sp], Number: int32(sp), Name: s.SpellName(int(sp)), Level: int32(s.SpellRec(SpellID(sp), 1))})
		}
		return l, true, nil
	case rules.BookFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return r.book(c), true, nil
	case rules.AreaFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return r.areaPreview(c, f), true, nil
	case rules.SpellAimFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return r.spellAim(f.Spell), true, nil
	case rules.SpellMenuFact:
		if c == nil {
			return nil, true, errNoChar
		}
		return r.spellMenu(c, f.Kind), true, nil
	default:
		return nil, false, nil
	}
}

// bolts is the reach and bounce boltPath gets from each spell whose effect
// is a bolt (fxLightning, fxBolt).
var bolts = map[int][2]int{SpLightningBolt: {8, 1}, 60: {3, 0}}

// areaPreview = por aoeMarks: who the area aimed at f.At would hit; the
// near, target and aim squares are kept.
func (r *Rules) areaPreview(c *Character, f rules.AreaFact) rules.AreaPreview {
	s := r.S
	x, y, aoe := f.At.X, f.At.Y, f.Area
	n := slices.Index(SpellIDs[:], f.Spell)
	if f.Spell == "" {
		n = 0
	}
	var hit []int // combatant numbers
	var out rules.AreaPreview
	if b := bolts[n]; b[0] > 0 {
		hit, out.Area = r.boltPreview(x, y, b[0], b[1])
	} else {
		saved := s.Near
		s.Near = nil
		s.BuildNear(x, y, aoe-1, 0xFF, 1)
		for _, n := range s.Near {
			hit = append(hit, int(n.Index))
		}
		s.Near = saved
		// the squares buildNear's line test reaches round (x,y), for views that draw the area
		for sy := max(0, y-aoe+1); sy <= min(24, y+aoe-1); sy++ {
			for sx := max(0, x-aoe+1); sx <= min(49, x+aoe-1); sx++ {
				rr, ax, ay := aoe-1, sx, sy
				if s.LineOfSight(&rr, &ax, &ay, x, y) {
					out.Area = append(out.Area, rules.BoardPos{X: sx, Y: sy})
				}
			}
		}
	}
	for _, i := range hit {
		p := s.Cpos[i]
		at := rules.BoardPos{X: int(p.X), Y: int(p.Y)}
		if t := s.Combatant(byte(i)); t != nil && t.Side == c.Side && n == SpBless && r.blessSkips(t) {
			out.Skipped = append(out.Skipped, at) // left out by the cast: shown apart, not counted
			continue
		}
		out.Hit = append(out.Hit, at)
	}
	return out
}

// boltPreview walks a bolt from (x,y) as its cast would (the target
// square, then boltPath(reach, ..., bounce)) without striking: who it hits
// (each once) and the squares it crosses. ponytail: the occupants stay
// put, so a foe the cast would kill still stops the line.
func (r *Rules) boltPreview(x, y, reach, bounce int) (hit []int, line rules.Squares) {
	s := r.S
	tx, ty := s.TgtX, s.TgtY
	defer func() { s.TgtX, s.TgtY = tx, ty }()
	s.TgtX, s.TgtY = x, y
	seen := map[int]bool{}
	strike := func(x, y int) {
		if who, _ := s.MapAt(x, y); who > 0 && !seen[int(who)] {
			seen[int(who)] = true
			hit = append(hit, int(who))
		}
	}
	strike(x, y)
	line = append(line, rules.BoardPos{X: x, Y: y})
	s.BoltWalk(reach, bounce, func(x, y int) { line = append(line, rules.BoardPos{X: x, Y: y}) }, func(_, _, x, y int) bool {
		strike(x, y)
		_, ground := s.MapAt(x, y)
		return ground > 0 && s.GroundInfo(Ground(ground))[0] == 0xFF
	})
	return hit, line
}

// spellAim is how spell aims in a fight (por aimSpell, aimOne).
func (r *Rules) spellAim(id rules.SpellID) rules.SpellAim {
	s := r.S
	n := slices.Index(SpellIDs[:], id)
	if n <= 0 {
		return rules.SpellAim{}
	}
	kind := s.SpellData(n, 0) & 0xF
	a := rules.SpellAim{Name: s.SpellName(n), Range: s.SpellDuration(n), FoesOnly: s.T().Spells[n&0xFF][7] == 0, Now: s.SpellData(n, 6)/3 == 0}
	switch {
	case !s.Fight.On || kind == 0:
		a.Kind = rules.SpellAimNone
	case kind == 0xF:
		a.Kind = rules.SpellAimOne
	case kind >= 8 && kind <= 0xE:
		a.Kind, a.Area = rules.SpellAimArea, int(kind&7)+1
	default: // up to (kind&3)+1 different targets, as chooseSpellTargets' loop
		a.Kind, a.Targets = rules.SpellAimTargets, int(kind&3)+1
	}
	if SpellID(n) == spDelayedBlastFireball && s.Fight.On {
		a.Choose, a.Choices = "Blast Delay in rounds?", delayedBlastChoices
	}
	if SpellID(n) == spFireShield && s.Fight.On {
		a.Choose, a.Choices = "flame type?", fireShieldChoices
	}
	return a
}

// spellMenu = por buildSpellList (ovr23:059C) for c: the lines of kind.
func (r *Rules) spellMenu(c *Character, kind int) rules.SpellMenu {
	s := r.S
	var out rules.SpellMenu
	var ids []SpellID
	// addSpellLine = ovr23:028D: behind a class header when the group changes
	add := func(spell SpellID, pending bool, item *Item) {
		group := byte(0) // bp-6 is 0 for the first line
		if n := len(ids); n > 0 {
			group = s.SpellRec(ids[n-1], 1)
		}
		if g := s.SpellRec(spell, 1); g != group {
			out = append(out, rules.SpellLine{Header: true, Text: SpellGroups[g], Planned: -1, Slots: -1})
		}
		l := rules.SpellLine{Text: s.SpellName(int(spell)), Number: int(spell), Memo: pending}
		if item != nil {
			l.Item = r.Handle(item)
		}
		out = append(out, l)
		ids = append(ids, spell)
	}
	// addScrollSpells = ovr23:044E (all) for scroll it; a pure query: the
	// reader's uncovering (takeScroll) is UseItem's Resolve's
	scroll := func(it *Item, all byte) {
		if it == nil || it.Hidden != 0 && !s.readsScroll(c, it) {
			return
		}
		for _, v := range it.Spells {
			if all != 0 && v.Scribing() || all == 0 && v > 0 {
				add(v.Spell(), v&ScrollScribing != 0, it)
			}
		}
	}
	// addAllScrolls = ovr23:051C: every scroll carried (DS:6B80 walks the
	// item list and is nil at the end: left untouched here)
	scrolls := func(all byte) {
		for _, it := range s.Carried(c) {
			if s.IsScroll(it) {
				scroll(it, all)
			}
		}
	}
	switch kind {
	case 0:
		for i := 0; i <= 0x14; i++ {
			if v := c.Memorized[i]; v.Held() {
				add(v, false, nil)
			}
		}
	case 5:
		for i := 0; i <= 0x14; i++ {
			if v := c.Memorized[i]; v.Pending() {
				add(v.Spell(), true, nil)
			}
		}
	case 1:
		for i := 1; i <= len(c.SpellBook); i++ {
			if c.SpellBook[i-1] != 0 && s.inBook(c, SpellID(i)) {
				add(SpellID(i), false, nil)
				out[len(out)-1].Count = SpellCount(c.Plan[:], SpellID(i)) // extra: how often it is memorized on a rest
			}
		}
		for i := range out { // extra: the level headers carry the slots planned
			if out[i].Header && i+1 < len(out) {
				sp := SpellID(out[i+1].Number)
				lvl, cls := s.SpellLevel(sp), s.SpellClass(sp)
				out[i].Planned, out[i].Slots = s.Planned(c, lvl, cls), int(c.SpellSlot(cls, int(lvl)))
			}
		}
	case 2:
		scroll(s.UsedScroll, 0)
	case 3:
		scrolls(0)
	case 6:
		scrolls(1)
	case 4:
		for i := 1; i <= LastSpell; i++ {
			if c.SpellSlot(s.SpellClass(SpellID(i)), int(s.SpellRec(SpellID(i), 1))) > 0 && c.SpellBook[i-1] == 0 {
				add(SpellID(i), false, nil)
			}
		}
	}
	return out
}

// coreSpellClass is a SpellClass as core names it.
var coreSpellClass = [...]rules.SpellClass{SpellClassCleric: rules.SpellCleric, SpellClassMage: rules.SpellMage, SpellClassDruid: rules.SpellDruid, SpellClassPower: rules.SpellPower}

// book is c's spells (por the roster cards' known, planned and memorized).
func (r *Rules) book(c *Character) rules.Book {
	s := r.S
	var b rules.Book
	for i, n := range c.SpellBook { // the book: the spells c knows
		if id := i + 1; n != 0 && s.inBook(c, SpellID(id)) {
			cl, lv, planned := s.BookSpell(c, id)
			b.Known = append(b.Known, rules.BookSpell{Spell: SpellIDs[id], Number: int32(id), Name: s.SpellName(id), Class: coreSpellClass[cl], Level: int32(lv), Planned: int32(planned)})
		}
	}
	for cls, levels := range c.SpellSlots {
		for i, n := range levels {
			b.Slots[cls][i] = int32(n)
		}
	}
	for _, sp := range c.Memorized {
		if sp.Held() {
			b.Memorized = append(b.Memorized, rules.BookSpell{Spell: SpellIDs[sp], Number: int32(sp), Name: s.SpellName(int(sp))})
		}
	}
	for _, l := range c.Levels {
		b.Levels = append(b.Levels, int32(l))
	}
	return b
}
