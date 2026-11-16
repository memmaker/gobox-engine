package ui

import (
	"cmp"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"slices"

	"gobox/asset"
	"gobox/game"
	"gobox/ui/combat"
)

// turnAct is the combat menu's action a Decision does (false: not a menu entry).
func turnAct(d game.Action) (ActKind, bool) {
	switch d := d.(type) {
	case game.MoveTo: // the turn's Move: the tabletop's walk (a square picked), never Move{} as it stands
		return ActWalk, true
	case game.Attack:
		return ActAttack, true
	case game.Shoot:
		return ActRanged, true
	case game.Guard:
		return ActGuard, true
	case game.OpenItems:
		return ActItems, true
	case game.Cast:
		return ActCast, true
	case game.DoManeuver:
		return ActManeuver, true
	case game.TurnUndead:
		return ActTurn, true
	case game.Bandage:
		return ActBandage, true
	case game.Delay:
		return ActWait, true
	case game.EndTurn:
		return ActEnd, true
	case game.Quick:
		if d.All { // its own entry: the bar sends the option itself
			return ActOption, true
		}
		return ActQuick, true
	case game.Flee:
		return ActFlee, true
	case game.ViewMember:
		return ActView, d.Member == viewSelected
	}
	return 0, false
}

// ownIconSize is the footprint a member's own PNG icon (Unit.Icon) is drawn
// at, whatever its pixels: a combat icon's square, as the composed ones.
var ownIconSize = image.Pt(24, 24)

// viewSelected is ViewMember's Member for the member whose turn it is (the
// selected one): the bar's own View, which the series' VerbKeys add.
const viewSelected = -1

// turnOptions are c's turn's options as the bar offers them: the game's,
// with the series' View (Data.VerbKeys) at its VerbOrder place when the game
// offers none (nil: no turn).
func turnOptions(c *game.Combat, d game.Data) []game.Option {
	if c == nil || c.Turn == nil {
		return nil
	}
	ops := c.Turn.Options
	k, ok := d.VerbKeys["View"]
	if !ok || len(ops) == 0 || c.Turn.Aim || slices.ContainsFunc(ops, func(o game.Option) bool {
		switch o.Do.(type) {
		case game.ViewMember, game.ViewTarget:
			return true
		}
		return false
	}) {
		return ops
	}
	at := len(ops)
	if vi := slices.Index(d.VerbOrder, "View"); vi >= 0 {
		if i := slices.IndexFunc(ops, func(o game.Option) bool { return slices.Index(d.VerbOrder, o.Label) > vi }); i >= 0 {
			at = i
		}
	}
	return slices.Insert(slices.Clone(ops), at, game.Option{Label: cmp.Or(d.Texts["View"], "View"), Key: k, Do: game.ViewMember{Member: viewSelected}})
}

// tilePicture is the tile name's art a as a picture to make, the series
// sprite of that name when a has no picture.
func tilePicture(name string, a game.Picture) Picture {
	pc := pix(a)
	if im := asset.SpriteImage(Series, "combat", name); a.File == "" && pc.Pix == nil && im != nil { // a series sprite
		pc.Img = im
	}
	return pc
}

// boardOf is the fight c as the tabletop draws it, its pictures made: the
// ground with its tiles (Combat.Tiles, read from the game's art), the units
// with their icons (a foe's Art, a member's composed from its pool ids or
// original icon numbers), effects (fly: a shot's sprite over the moment),
// the turn's bar, the cursor, the walks, the aim, the strike and the log.
// party are the figures the members stand as then. nil outside a fight.
func (p *Presenter) boardOf(c *game.Combat, party []game.Member, fly *flight) *combat.Board {
	if c == nil {
		return nil
	}
	var art map[string]game.Picture
	var d game.Data
	if p.eng != nil {
		art, d = p.eng.art(), p.eng.D
	}
	ink := egaInk
	b := &combat.Board{
		WallTiles: dungeonTiles, StoneTile: 3, TileNames: slices.Clone(c.Tiles),
		Floor: combatGround(), Dungeon: c.Dungeon,
		Marks: pts(c.Marks), Targets: slices.Clone(c.Targets), Panel: c.Panel, Log: slices.Clone(c.Log),
		AimArea: c.AimArea, AreaAt: pts(c.AreaAt), Skipped: pts(c.Skipped), AimWhy: c.Why,
		Strike: combat.Strike{Seq: c.Strike.Seq, From: c.Strike.From, To: c.Strike.To, Hit: c.Strike.Hit, Damage: c.Strike.Damage},
	}
	grid := image.Rect(0, 0, len(b.Ground[0]), len(b.Ground))
	for _, g := range c.Squares {
		if pt(g.At).In(grid) {
			b.Ground[g.At.Y][g.At.X] = combat.Square{Tile: g.Tile, Wall: g.Block, Height: g.Height, Crown: g.Crown}
		}
	}
	tiles := make([]Picture, 0x100)
	if c.FloorTile != "" {
		b.FloorTile = asset.SpriteImage(Series, "combat", c.FloorTile)
	}
	slot, slots := len(c.Tiles), map[string]int{} // a square's own Art: a tile slot of its own past the named tiles, one per picture
	slotOf := func(a game.Picture) (int, bool) {  // a's tile slot, made on first use
		if !hasPic(a) {
			return 0, false
		}
		if a.File == "" && a.Pixels == nil { // a name only: a Combat.Tiles tile is its own slot
			if i := slices.Index(c.Tiles, a.Name); i >= 0 {
				return i, true
			}
		}
		k := fmt.Sprint(a.File, a.Block, a.Frame, a.Name, len(a.Pixels), crc32.ChecksumIEEE(a.Pixels)) // pixels of one size differ: the checksum tells them apart
		t, ok := slots[k]
		if !ok {
			for len(tiles) <= slot { // past 0x100 (a game naming every tile): more slots, not none
				tiles = append(tiles, Picture{})
			}
			t, ok, slots[k] = slot, true, slot
			tiles[slot] = tilePicture(a.Name, a)
			slot++
		}
		return t, ok
	}
	for _, g := range c.Squares {
		if !pt(g.At).In(grid) {
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
		if n != "" && i < len(tiles) {
			tiles[i] = tilePicture(n, art[n])
		}
	}
	b.Tiles = make([]image.Image, len(tiles))
	for i, pc := range tiles {
		b.Tiles[i] = textured(picImage(pc, ink))
	}
	for _, f := range c.Effects {
		b.Effects = append(b.Effects, combat.Effect{X: float64(f.X) / 3, Y: float64(f.Y) / 3, Pic: picImage(pix(f.Art), ink)})
	}
	if fly != nil {
		b.Effects = append(b.Effects, combat.Effect{X: fly.X, Y: fly.Y, Pic: picImage(pix(fly.Art), ink)})
	}
	for _, r := range c.Reach {
		b.Reach = append(b.Reach, pt(r.At))
		if b.Paths == nil {
			b.Paths = map[image.Point][]combat.Step{}
		}
		var path []combat.Step
		for _, st := range r.Path {
			path = append(path, combat.Step{Point: pt(st.At), Free: st.Free})
		}
		b.Paths[pt(r.At)] = path
	}
	for _, g := range c.Units { // hovering a foe shows the walk an attack on it takes
		for _, sq := range g.Squares {
			if len(g.Approach) == 0 || g.Gone {
				break
			}
			if b.Paths == nil {
				b.Paths = map[image.Point][]combat.Step{}
			}
			for _, st := range g.Approach {
				b.Paths[pt(sq)] = append(b.Paths[pt(sq)], combat.Step{Point: pt(st.At), Free: st.Free})
			}
		}
	}
	for _, g := range c.Units {
		if g.Gone {
			continue
		}
		u := combat.Unit{ID: g.ID, X: g.At.X, Y: g.At.Y, Size: g.Size, Squares: pts(g.Squares), Party: !g.Foe, Facing: g.Facing,
			Active: g.Active, Name: g.Name, HP: g.HP, MaxHP: g.MaxHP, AC: g.AC, Hit: g.Hit, Dmg: g.Dmg, Status: g.Status, Down: g.Dead}
		if g.Member >= 0 && g.Member < len(party) {
			for _, it := range party[g.Member].Items {
				if it.Readied && it.Kind == "weapon" {
					u.Weapon = it.Name
				}
			}
		}
		if t := c.Turn; t != nil && t.Weapon != nil && t.Weapon.Name != "" && (g.ID == t.Unit || g.Active) { // the turn's own say for the unit acting
			u.Weapon = t.Weapon.Name
		}
		icon, size := figureOf(g, party)
		u.IconSize = size
		if im := ownPose(g.IconPNG, g.Pose&1, g.Facing > 3); im != nil {
			u.Icon, u.IconSize = im, ownIconSize
		}
		if u.Icon == nil {
			u.Icon = picImage(icon, ink)
			if u.Icon == nil && (icon.File != "" || icon.Name != "" || icon.Pix != nil) { // a picture that does not load (a DAX block the file lacks): gobox's default figure, never an empty square
				u.Icon = PickIcon(game.Member{}, "", "") // ponytail: one stand-in for every lost picture; a per-kind monster stand-in when the pool has one
			}
			if u.IconSize == (image.Point{}) {
				u.IconSize = nativeSize(icon)
			}
		}
		b.Units = append(b.Units, u)
	}
	if p.eng != nil && p.eng.r.aim != nil { // gobox's aim cursor (frontState): its square, a large unit's footprint
		b.Cursor, b.CursorOn = pt(p.eng.r.cursor), true
		b.CursorAt = []image.Point{b.Cursor}
		for _, u := range b.Units {
			if slices.Contains(u.Squares, b.Cursor) {
				b.CursorAt = u.Squares
			}
		}
	}
	if t := c.Turn; t != nil {
		b.Turn = &combat.Turn{Unit: t.Unit, Aim: t.Aim, Casting: t.Spell, Moves: t.Moves, CastDirect: len(c.Spells) == 0}
		for _, op := range turnOptions(c, d) {
			act, ok := turnAct(op.Do)
			if !ok { // a decision with no act of its own (an Answer by hotkey): the bar sends the option itself
				act = ActOption
			}
			k, known := combatKinds[act]
			if known {
				b.Turn.Can = append(b.Turn.Can, k)
			}
			if !known || act == 0 {
				k = combat.DecideOption
			}
			key := byte(0)
			if op.Key > 0 && op.Key < 128 {
				key = byte(op.Key)
			}
			b.Turn.Options = append(b.Turn.Options, combat.Option{Label: op.Label, Key: key, DecisionKind: k, Chords: op.Chords})
		}
	}
	return b
}

// flight is a shot's sprite at a square (in squares, a third's steps) over
// the board a moment holds.
type flight struct {
	X, Y float64
	Art  game.Picture
}

// weapon is w as the bar shows it (nil: none).
func weapon(w *game.Weapon) *InvItem {
	if w == nil {
		return nil
	}
	return &InvItem{Name: w.Name, Type: byte(w.Type), Base: w.Base, Words: w.Words, Look: w.Look, Rules: w.Rules}
}

// figureOf is the picture unit g shows, and its size in the original's
// pixels (zero: the decoded picture's): its Art, or a member's figure (memberPicture)
// when it has no icon of its own.
func figureOf(g game.Unit, party []game.Member) (icon Picture, size image.Point) {
	if g.Pose&1 == 1 && hasPic(g.Action) { // the unit's action pose, where the game has one
		g.Art = g.Action
	}
	icon = pix(g.Art)
	if a := shownPic(g.Art); a.W > 0 { // a DAX ref's size is the decoded picture's (the presenter's nativeSize)
		size = image.Pt(a.W, a.H)
	}
	if m := g.Member; g.IconPNG == nil && !g.Foe && m >= 0 && m < len(party) {
		if pic, sz, ok := memberPicture(party[m], g, g.Pose); ok {
			icon, size = pic, sz
		}
	}
	return icon, size
}

// memberPicture is member m's figure as unit g in pose: g's Art for a member
// with no icon of its own (COTAB's NPC members, a ref without W: the
// presenter's native size), else its pool icon (defaultIcon's pick when it
// has neither pool ids nor icon numbers), mirrored when g faces west. The
// one way a member's figure is made: the live board and every replayed one.
func memberPicture(m game.Member, g game.Unit, pose int) (pic Picture, size image.Point, ok bool) {
	if own := m.IconHeadID != "" || m.Icon != nil; !own && hasPic(g.Art) {
		a := shownPic(g.Art)
		return pix(a), image.Pt(a.W, a.H), true
	}
	pm := defaultIcon(m)
	head, body := memberIcon(pm)
	if pic, ok = poolIcon(head, body, pm.IconColors, pose, g.Facing > 3); ok {
		size = image.Pt(pic.PixW, len(pic.Pix)/max(1, pic.PixW))
	}
	return pic, size, ok
}

// combatGround is the bare ground's colour on a board with no floor tile:
// the shown art's (the look file's in the Amiga's; in the DOS art colour 0,
// which SetCombatPalette, cbtmap.pas, shows as dark grey in a fight).
func combatGround() color.RGBA {
	if amigaGround != nil && Amiga() {
		return *amigaGround
	}
	return asset.EGARGBA(asset.EGADarkGrey)
}
