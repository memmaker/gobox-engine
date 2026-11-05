package classic

import (
	"image"
	"slices"

	"gobox/front"
	"gobox/game"
	"gobox/sprites"
)

// turnAct is the combat menu's action a Decision does (false: not a menu entry).
func turnAct(d game.Decision) (front.ActKind, bool) {
	switch d.(type) {
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
		return front.ActQuick, true
	case game.Flee:
		return front.ActFlee, true
	}
	return 0, false
}

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
			if m := g.Member; !g.Foe && m >= 0 && m < len(s.Party) {
				pm := defaultIcon(s.Party[m])
				head, body := memberIcon(pm)
				if pic, ok := poolIcon(head, body, pm.IconColors, 0, g.Facing > 3); ok {
					if !cloned { // the game's IconPix by unit, so far
						o.IconPix, cloned = append(make([]front.Picture, 0, len(o.Units)), b.IconPix...), true
						o.IconPix = append(o.IconPix, make([]front.Picture, max(0, len(o.Units)-len(o.IconPix)))...)
					}
					o.Units[i].Icon, o.IconPix[i] = nil, pic
					o.Units[i].IconSize = image.Pt(pic.PixW, len(pic.Pix)/pic.PixW)
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
		if b.Turn != nil { // the weapon icons and the rule set's maneuvers are the game's
			o.Turn.Weapon, o.Turn.Missile, o.Turn.Maneuvers = b.Turn.Weapon, b.Turn.Missile, b.Turn.Maneuvers
		}
		for _, op := range t.Options {
			if act, ok := turnAct(op.Do); ok {
				o.Turn.Can = append(o.Turn.Can, act)
			}
		}
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
			Active: g.Active, Name: g.Name, HP: g.HP, MaxHP: g.MaxHP, AC: g.AC, Hit: g.Hit, Dmg: g.Dmg, Status: g.Status}
		icon := pix(g.Art)
		u.IconSize = image.Pt(24, 24)
		if g.Art.W > 0 {
			u.IconSize = image.Pt(g.Art.W, g.Art.H)
		}
		if m := g.Member; !g.Foe && m >= 0 && m < len(s.Party) {
			head, body := memberIcon(s.Party[m])
			if pic, ok := poolIcon(head, body, s.Party[m].IconColors, g.Pose, g.Facing > 3); ok {
				icon, u.IconSize = pic, image.Pt(pic.PixW, len(pic.Pix)/pic.PixW)
			}
		}
		b.Units = append(b.Units, u)
		b.IconPix = append(b.IconPix, icon)
	}
	if t := c.Turn; t != nil {
		b.Turn = &front.Turn{Unit: t.Unit, Aim: t.Aim, Casting: t.Spell, Moves: t.Moves}
		for _, op := range t.Options {
			if act, ok := turnAct(op.Do); ok {
				b.Turn.Can = append(b.Turn.Can, act)
			}
		}
		for _, sp := range c.Spells {
			b.Turn.Spells = append(b.Turn.Spells, front.Spell{ID: sp.ID, Name: sp.Name})
		}
	}
	return b
}
