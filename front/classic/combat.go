package classic

import (
	"fmt"
	"image"
	"slices"

	"gobox/front"
	"gobox/game"
	"gobox/sprites"
)

// turnAct is the combat menu's action a Decision does (false: not a menu entry).
func turnAct(d game.Decision) (front.ActKind, bool) {
	switch d := d.(type) {
	case game.Move: // the turn's Move: the tabletop's walk (a square picked), never Move{} as it stands
		return front.ActWalk, true
	case game.Attack:
		return front.ActAttack, true
	case game.Ranged:
		return front.ActRanged, true
	case game.Guard:
		return front.ActGuard, true
	case game.Items:
		return front.ActItems, true
	case game.Cast:
		return front.ActCast, true
	case game.Maneuver:
		return front.ActManeuver, true
	case game.TurnUndead:
		return front.ActTurn, true
	case game.Bandage:
		return front.ActBandage, true
	case game.Wait:
		return front.ActWait, true
	case game.EndTurn:
		return front.ActEnd, true
	case game.Quick:
		if d.All { // its own entry: the bar sends the option itself
			return front.ActOption, true
		}
		return front.ActQuick, true
	case game.Flee:
		return front.ActFlee, true
	}
	return 0, false
}

// ownIconSize is the footprint a member's own PNG icon (Unit.Icon) is drawn
// at, whatever its pixels: a combat icon's square, as the composed ones.
var ownIconSize = image.Pt(24, 24)

// Battle is b (the game's) with what s.Combat gives in its place: the
// combat menu (Turn), each combatant's name, hit points, armor class,
// damage and status, the squares a move reaches, the foes an attack or aim
// takes and the aim (marks, area, why not). b's map, icons and effects stay.
func Battle(b *front.Battle, s game.State) *front.Battle {
	c := s.Combat
	if b == nil || c == nil {
		return b
	}
	o := *b
	o.Units = slices.Clone(b.Units)
	cloned := false
	for i, u := range o.Units {
		if k := slices.IndexFunc(c.Units, func(g game.Unit) bool { return g.ID == u.ID }); k >= 0 {
			g := c.Units[k]
			o.Units[i].Party, o.Units[i].Active, o.Units[i].Name = !g.Foe, g.Active, g.Name
			o.Units[i].HP, o.Units[i].MaxHP, o.Units[i].AC = g.HP, g.MaxHP, g.AC
			o.Units[i].Hit, o.Units[i].Dmg, o.Units[i].Status = g.Hit, g.Dmg, g.Status
			if g.Icon != nil {
				o.Units[i].Icon, o.Units[i].IconSize = g.Icon, ownIconSize
			} else if m := g.Member; !g.Foe && m >= 0 && m < len(s.Party) {
				if pic, size, ok := memberPicture(s.Party[m], g, 0); ok {
					if !cloned { // the game's IconPix by unit, so far
						o.IconPix, cloned = append(make([]front.Picture, 0, len(o.Units)), b.IconPix...), true
						o.IconPix = append(o.IconPix, make([]front.Picture, max(0, len(o.Units)-len(o.IconPix)))...)
					}
					o.Units[i].Icon, o.IconPix[i] = nil, pic
					o.Units[i].IconSize = size
				}
			}
		}
	}
	gone := func(i int) bool { // a unit the combat has off the map is not drawn
		k := slices.IndexFunc(c.Units, func(g game.Unit) bool { return g.ID == o.Units[i].ID })
		return k >= 0 && c.Units[k].Gone
	}
	for i := len(o.Units) - 1; i >= 0; i-- {
		if gone(i) {
			if i < len(o.IconPix) {
				o.IconPix = slices.Delete(slices.Clone(o.IconPix), i, i+1)
			}
			o.Units = slices.Delete(o.Units, i, i+1)
		}
	}
	o.Reach = nil
	for _, r := range c.Reach {
		o.Reach = append(o.Reach, r.At)
	}
	o.Targets, o.Panel, o.Marks, o.Area = slices.Clone(c.Targets), c.Panel, slices.Clone(c.Marks), c.Area
	o.AreaAt, o.Skipped, o.AimWhy = slices.Clone(c.AreaAt), slices.Clone(c.Skipped), c.Why
	o.Turn = nil
	if t := c.Turn; t != nil {
		o.Turn = &front.Turn{Unit: t.Unit, Aim: t.Aim, Casting: t.Spell, Moves: t.Moves}
		turnExtras(o.Turn, t)
		o.Turn.Can, o.Turn.Options = turnOptions(t)
		for _, sp := range c.Spells {
			o.Turn.Spells = append(o.Turn.Spells, front.Spell{ID: sp.ID, Name: sp.Name})
		}
	}
	return &o
}

// BattleOf is the combat map built from s.Combat alone: the ground with its
// tiles (Combat.Tiles, read from art), the units with their icons (a foe's
// Art, a member's composed from its pool ids or original icon numbers),
// effects, cursor, walks, aim, strike and log. nil outside a fight.
func BattleOf(s game.State, art map[string]game.Picture) *front.Battle {
	c := s.Combat
	if c == nil {
		return nil
	}
	b := &front.Battle{
		Floor: c.Floor, Dungeon: c.Dungeon, Cursor: c.Cursor, CursorOn: c.CursorOn, CursorAt: slices.Clone(c.CursorAt),
		Marks: slices.Clone(c.Marks), Targets: slices.Clone(c.Targets), Panel: c.Panel, Log: slices.Clone(c.Log),
		Area: c.Area, AreaAt: slices.Clone(c.AreaAt), Skipped: slices.Clone(c.Skipped), AimWhy: c.Why, View: c.View,
		Strike: front.Strike{Seq: c.Strike.Seq, From: c.Strike.From, To: c.Strike.To, Hit: c.Strike.Hit, Damage: c.Strike.Damage},
	}
	for _, g := range c.Squares {
		if g.At.In(image.Rect(0, 0, len(b.Ground[0]), len(b.Ground))) {
			b.Ground[g.At.Y][g.At.X] = front.Square{Tile: g.Tile, Wall: g.Block, Height: g.Height}
		}
	}
	b.TilePix = make([]front.Picture, 0x100)
	b.TileNames = slices.Clone(c.Tiles)
	if c.FloorTile != "" {
		b.FloorImg = sprites.Image(Series, "combat", c.FloorTile)
	}
	slot, slots := len(c.Tiles), map[string]int{} // a square's own Art: a tile slot of its own past the named tiles, one per picture
	slotOf := func(a game.Picture) (int, bool) {  // a's tile slot, made on first use
		if a.File == "" && a.Pixels == nil && a.Name == "" {
			return 0, false
		}
		if a.File == "" && a.Pixels == nil { // a name only: a Combat.Tiles tile is its own slot
			if i := slices.Index(c.Tiles, a.Name); i >= 0 {
				return i, true
			}
		}
		k := fmt.Sprint(a.File, a.Block, a.Frame, a.Name, len(a.Pixels))
		t, ok := slots[k]
		if !ok {
			for len(b.TilePix) <= slot { // past 0x100 (a game naming every tile): more slots, not none
				b.TilePix = append(b.TilePix, front.Picture{})
			}
			t, ok, slots[k] = slot, true, slot
			b.TilePix[slot] = pix(a)
			if im := sprites.Image(Series, "combat", a.Name); a.File == "" && a.Pixels == nil && im != nil { // a series sprite by name, as Combat.Tiles' are
				b.TilePix[slot].Img = im
			}
			slot++
		}
		return t, ok
	}
	for _, g := range c.Squares {
		if !g.At.In(image.Rect(0, 0, len(b.Ground[0]), len(b.Ground))) {
			continue
		}
		if g.Art.File != "" || g.Art.Pixels != nil {
			if t, ok := slotOf(g.Art); ok {
				b.Ground[g.At.Y][g.At.X].Tile = t
			}
		}
		if t, ok := slotOf(g.Overlay); ok {
			b.Ground[g.At.Y][g.At.X].Overlay = t + 1
		}
	}
	for i, n := range c.Tiles {
		if n != "" && i < len(b.TilePix) {
			b.TilePix[i] = pix(art[n])
			if im := sprites.Image(Series, "combat", n); art[n].File == "" && b.TilePix[i].Pix == nil && im != nil { // a series sprite
				b.TilePix[i].Img = im
			}
		}
	}
	for _, f := range c.Effects {
		b.Effects = append(b.Effects, front.Effect{X: float64(f.X) / 3, Y: float64(f.Y) / 3})
		b.EffectPix = append(b.EffectPix, pix(f.Art))
	}
	for _, r := range c.Reach {
		b.Reach = append(b.Reach, r.At)
		if b.Paths == nil {
			b.Paths = map[image.Point][]front.Step{}
		}
		var path []front.Step
		for _, st := range r.Path {
			path = append(path, front.Step{Point: st.At, Free: st.Free})
		}
		b.Paths[r.At] = path
	}
	for _, g := range c.Units {
		if g.Gone {
			continue
		}
		u := front.Unit{ID: g.ID, X: g.At.X, Y: g.At.Y, Size: g.Size, Squares: slices.Clone(g.Squares), Party: !g.Foe, Facing: g.Facing,
			Active: g.Active, Name: g.Name, HP: g.HP, MaxHP: g.MaxHP, AC: g.AC, Hit: g.Hit, Dmg: g.Dmg, Status: g.Status, Down: g.Dead}
		if g.Member >= 0 && g.Member < len(s.Party) {
			for _, it := range s.Party[g.Member].Items {
				if it.Readied && it.Kind == "weapon" {
					u.Weapon = it.Name
				}
			}
		}
		if t := c.Turn; t != nil && t.Weapon != nil && t.Weapon.Name != "" && (g.ID == t.Unit || g.Active) { // the turn's own say for the unit acting
			u.Weapon = t.Weapon.Name
		}
		icon := pix(g.Art)
		if g.Art.W > 0 { // a DAX ref's size is the decoded picture's (the presenter's nativeSize)
			u.IconSize = image.Pt(g.Art.W, g.Art.H)
		}
		if g.Icon != nil {
			u.Icon, u.IconSize = g.Icon, ownIconSize
		} else if m := g.Member; !g.Foe && m >= 0 && m < len(s.Party) {
			if pic, size, ok := memberPicture(s.Party[m], g, g.Pose); ok {
				icon, u.IconSize = pic, size
			}
		}
		b.Units = append(b.Units, u)
		b.IconPix = append(b.IconPix, icon)
	}
	if t := c.Turn; t != nil {
		b.Turn = &front.Turn{Unit: t.Unit, Aim: t.Aim, Casting: t.Spell, Moves: t.Moves}
		turnExtras(b.Turn, t)
		b.Turn.Can, b.Turn.Options = turnOptions(t)
		for _, sp := range c.Spells {
			b.Turn.Spells = append(b.Turn.Spells, front.Spell{ID: sp.ID, Name: sp.Name})
		}
	}
	return b
}

// turnOptions are t's options as the front's acts (Can) and as the bar's
// buttons, every one the game offers: one without a front act is ActOption.
func turnOptions(t *game.Turn) (can []front.ActKind, ops []front.TurnOption) {
	for _, op := range t.Options {
		act, ok := turnAct(op.Do)
		if !ok { // a decision with no act of its own (an Answer by hotkey): the bar sends the option itself
			act = front.ActOption
		}
		can = append(can, act)
		key := byte(0)
		if op.Key > 0 && op.Key < 128 {
			key = byte(op.Key)
		}
		ops = append(ops, front.TurnOption{Label: op.Label, Key: key, Kind: act, Chords: op.Chords})
	}
	return can, ops
}

// turnExtras gives o the Turn's weapons and maneuvers.
func turnExtras(o *front.Turn, t *game.Turn) {
	o.Weapon, o.Missile, o.Maneuvers = weapon(t.Weapon), weapon(t.Missile), nil
	for _, m := range t.Maneuvers {
		o.Maneuvers = append(o.Maneuvers, front.Maneuver{ID: m.ID, Name: m.Name, Tip: m.Tip, Aim: front.Aim(m.Aim), Range: m.Range, Area: m.Area, Move: m.Shift})
	}
}

// weapon is w as the bar shows it (nil: none).
func weapon(w *game.Weapon) *front.InvItem {
	if w == nil {
		return nil
	}
	return &front.InvItem{Name: w.Name, Type: byte(w.Type), Words: w.Words, Look: w.Look, Rules: w.Rules}
}

// memberPicture is member m's figure as unit g in pose: g's Art for a member
// with no icon of its own (COTAB's NPC members, a ref without W: the
// presenter's native size), else its pool icon (defaultIcon's pick when it
// has neither pool ids nor icon numbers), mirrored when g faces west. The
// one way a member's figure is made: the live board and every replayed one.
func memberPicture(m game.Member, g game.Unit, pose int) (pic front.Picture, size image.Point, ok bool) {
	if own := m.IconHeadID != "" || m.Icon != nil; !own && hasPic(g.Art) {
		return pix(g.Art), image.Pt(g.Art.W, g.Art.H), true
	}
	pm := defaultIcon(m)
	head, body := memberIcon(pm)
	if pic, ok = poolIcon(head, body, pm.IconColors, pose, g.Facing > 3); ok {
		size = image.Pt(pic.PixW, len(pic.Pix)/max(1, pic.PixW))
	}
	return pic, size, ok
}
