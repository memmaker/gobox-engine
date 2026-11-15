package engine

import (
	"cmp"
	"math"
	"slices"

	"gobox/game"
	"gobox/rules"
)

// turnDecisions are the combat menu's actions as the Decisions they do,
// with their labels.
// turnBar is combatMenu's bar (ovr09:069E-07D4 "Move View Aim Use Cast
// Turn Quick Done") with the original's letters, the verbs the turn can (an
// empty kind: always there). View is gobox's (its sheet). Done opens its
// own menu (doneMenu). Ranged, Maneuver and the Flee button are extras the
// bar does not carry: Aim shoots the missile weapon readied, a Move off the
// map flees.
var turnBar = []struct {
	can TurnVerb
	o   game.Option
}{
	{VerbNone, game.Option{Label: "Move", Key: 'M', Do: game.MoveTo{}}},
	{VerbAttack, game.Option{Label: "Aim", Key: 'A', Do: game.Attack{}}},
	{VerbItems, game.Option{Label: "Use", Key: 'U', Do: game.OpenItems{Member: -1}}},
	{VerbCast, game.Option{Label: "Cast", Key: 'C', Do: game.Cast{}}},
	{VerbTurn, game.Option{Label: "Turn", Key: 'T', Do: game.TurnUndead{}}},
	{VerbQuick, game.Option{Label: "Quick", Key: 'Q', Do: game.Quick{}}},
	{VerbEnd, game.Option{Label: "Done", Key: 'D', Do: game.EndTurn{}}},
}

// fightState is the fight as the View shows it: the combatants with their
// numbers, the turn's menu and spells (with the aims they take), the squares
// a move reaches, the foes an attack takes and the aim.
func (e *Engine) fightState() *game.Combat {
	if !e.battleUp() {
		return nil
	}
	b := e.battle()
	c := &game.Combat{Size: rules.Size{W: len(b.Ground[0]), H: len(b.Ground)}, Round: int(e.ruleset().Check(rules.FightRound{}, nil).Value), Dungeon: b.Dungeon,
		Targets: slices.Clone(b.Targets), Panel: b.Panel, Marks: slices.Clone(b.Marks), Area: b.Area,
		AreaAt: slices.Clone(b.AreaAt), Skipped: slices.Clone(b.Skipped), Why: b.AimWhy,
		Strike: game.Strike{Seq: b.Strike.Seq, From: b.Strike.From, To: b.Strike.To, Hit: b.Strike.Hit, Damage: b.Strike.Damage},
		Log:    slices.Clone(b.Log)}
	for y := range b.Ground {
		for x, q := range b.Ground[y] {
			c.Squares = append(c.Squares, game.Ground{At: rules.BoardPos{X: x, Y: y}, Block: q.Wall, Height: q.Height, Tile: q.Tile, Crown: q.Crown})
		}
	}
	for i, p := range b.TilePix { // their pictures are Art()'s (overlandArt)
		n := ""
		if p.Pix != nil {
			n = combatTileArt(i)
		}
		c.Tiles = append(c.Tiles, n)
	}
	for i, f := range b.Effects {
		c.Effects = append(c.Effects, game.Effect{X: int(math.Round(f.X * 3)), Y: int(math.Round(f.Y * 3)), Art: gamePicture(b.EffectPix[i])})
	}
	for i, u := range b.Units {
		g := game.Unit{ID: u.ID, Name: u.Name, At: rules.BoardPos{X: u.X, Y: u.Y}, Facing: u.Facing, Foe: !u.Party, HP: u.HP, MaxHP: u.MaxHP,
			Size: u.Size, Squares: slices.Clone(u.Squares), AC: u.AC, Status: u.Status, Dmg: u.Dmg, Member: -1, Active: u.Active, Hit: u.Hit}
		if !u.Party { // a member's icon is the front's to compose from its icon numbers
			g.Art = gamePicture(b.IconPix[i])
		}
		if u.Party {
			id := e.unitID(u.ID)
			g.Member = slices.Index(e.partyIDs(), id) // the rules' PartyIndex: past the end when not in it
			if g.Member < 0 {
				g.Member = e.partySize()
			}
			g.IconPNG = e.iconPics(id).PNG // a dropped PNG: the front cuts the pose the blow shows, facing as the unit
		}
		if len(u.Squares) > 0 && !u.Party { // a foe's squares carry the walk an attack on it takes (combatTurn's walks)
			for _, st := range b.Paths[u.Squares[0]] {
				g.Approach = append(g.Approach, game.Step{At: st.At, Free: st.Free})
			}
		}
		c.Units = append(c.Units, g)
	}
	for _, r := range b.Reach {
		rc := game.Reach{At: r, Steps: len(b.Paths[r])}
		for _, st := range b.Paths[r] {
			rc.Path = append(rc.Path, game.Step{At: st.At, Free: st.Free})
		}
		c.Reach = append(c.Reach, rc)
	}
	slices.SortFunc(c.Reach, func(a, b game.Reach) int { // the engine's map order differs between runs
		return cmp.Or(cmp.Compare(a.At.Y, b.At.Y), cmp.Compare(a.At.X, b.At.X))
	})
	if t := b.Turn; t != nil {
		c.Turn = &game.Turn{Unit: t.Unit, Moves: t.Moves, Aim: t.Aim, Spell: t.Casting, Weapon: gameWeapon(t.Weapon), Missile: gameWeapon(t.Missile)}
		for _, m := range t.Maneuvers {
			c.Turn.Maneuvers = append(c.Turn.Maneuvers, game.TurnManeuver{ID: int(m.ID), Name: m.Name, Tip: m.Tip, Aim: int(m.Aim), Range: m.Range, Area: m.Area, Shift: m.Move})
		}
		for _, v := range turnBar {
			if v.can == VerbNone || slices.Contains(t.Can, v.can) {
				c.Turn.Options = append(c.Turn.Options, v.o)
			}
		}
		ch := e.unitID(t.Unit)
		for _, sp := range t.Spells {
			ref := game.SpellRef{ID: sp.ID, Name: sp.Name}
			c.Spells = append(c.Spells, ref)
			if ch == rules.NoChar {
				continue
			}
			e.as(e.sel)
			aim := e.fact(ch, rules.SpellAimFact{Spell: e.spellID(sp.ID)}).(rules.SpellAim)
			if aim.Kind != rules.SpellAimNone { // the spells that take a square
				ref.Targets = 1
				switch aim.Kind {
				case rules.SpellAimArea:
					ref.Area = aim.Area - 1
				case rules.SpellAimTargets:
					ref.Targets = aim.Targets
				}
				ref.Range = aim.Range
				c.Aims = append(c.Aims, ref)
			}
		}
	}
	return c
}

// gameWeapon is a Turn's weapon in the State (nil: none).
func gameWeapon(w *rules.ItemCard) *game.Weapon {
	if w == nil {
		return nil
	}
	return &game.Weapon{Name: w.Name, Type: int(w.Type), Base: w.Kind, Words: w.Words, Look: w.Look, Rules: w.Rules}
}
