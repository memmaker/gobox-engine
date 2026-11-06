package present

import (
	"gobox/front"
	"image"
	"slices"

	"gobox/combat"
	"gobox/gpu"
	"gobox/sprites"
)

// The adapter between the engine's combat snapshot and gobox/combat.

var combatKinds = map[front.ActKind]combat.Kind{
	front.ActWalk: combat.ActWalk, front.ActAttack: combat.ActAttack, front.ActAim: combat.ActAim, front.ActPoint: combat.ActPoint,
	front.ActClose: combat.ActClose, front.ActYes: combat.ActYes, front.ActNo: combat.ActNo, front.ActWait: combat.ActWait,
	front.ActEnd: combat.ActEnd, front.ActQuick: combat.ActQuick, front.ActGuard: combat.ActGuard, front.ActBandage: combat.ActBandage,
	front.ActFlee: combat.ActFlee, front.ActTurn: combat.ActTurn, front.ActCast: combat.ActCast, front.ActItems: combat.ActItems,
	front.ActRanged: combat.ActRanged, front.ActManeuver: combat.ActManeuver,
	front.ActStep: combat.ActStep, front.ActView: combat.ActView, front.ActOption: combat.ActOption,
}

var engineKinds = func() map[combat.Kind]front.ActKind {
	m := map[combat.Kind]front.ActKind{}
	for e, c := range combatKinds {
		m[c] = e
	}
	return m
}()

// syncBattle gives the tabletop its theme and Shift before it draws or takes a key.
func (p *Presenter) syncBattle() {
	p.battle.Theme, p.battle.Shift, p.battle.Alt = combat.DefaultTheme, gpu.IsKeyPressed(gpu.KeyShift), gpu.IsKeyPressed(gpu.KeyAlt)
	p.battle.Theme.Series = series
	p.battle.Meshes, p.battle.Props = sprites.Meshes(series), sprites.Combat(series)
}

// toAction is what the player chose as the engine's action.
func toAction(d combat.Decision) front.Action {
	return front.Action{Kind: engineKinds[d.Kind], N: d.N, X: d.X, Y: d.Y}
}

// toBoard is the engine's snapshot as the tabletop's.
func toBoard(b *front.Battle) *combat.Board {
	if b == nil {
		return nil
	}
	o := &combat.Board{
		Tiles: b.Tiles, TileNames: b.TileNames, FloorTile: b.FloorImg, WallTiles: dungeonTiles, StoneTile: 3, Floor: b.Floor, Dungeon: b.Dungeon,
		Cursor: b.Cursor, CursorOn: b.CursorOn, CursorAt: b.CursorAt, Marks: b.Marks, Reach: b.Reach,
		Targets: b.Targets, Panel: b.Panel, Log: b.Log, Area: b.Area, AreaAt: b.AreaAt, AimArea: b.AimArea, Skipped: b.Skipped,
		AimWhy: b.AimWhy, View: b.View, Strike: combat.Strike(b.Strike),
	}
	if b.CursorOn && len(b.CursorAt) == 0 { // a game that sets only Cursor: the square (a large unit's footprint)
		o.CursorAt = []image.Point{b.Cursor}
		for _, u := range b.Units {
			if slices.Contains(u.Squares, b.Cursor) {
				o.CursorAt = u.Squares
			}
		}
	}
	for y := range b.Ground {
		for x, s := range b.Ground[y] {
			o.Ground[y][x] = combat.Square(s)
		}
	}
	for _, u := range b.Units {
		o.Units = append(o.Units, combat.Unit(u))
	}
	if b.Paths != nil {
		o.Paths = map[image.Point][]combat.Step{}
		for k, p := range b.Paths {
			for _, s := range p {
				o.Paths[k] = append(o.Paths[k], combat.Step(s))
			}
		}
	}
	for _, f := range b.Effects {
		o.Effects = append(o.Effects, combat.Effect(f))
	}
	if t := b.Turn; t != nil {
		o.Turn = &combat.Turn{Unit: t.Unit, Aim: t.Aim, Casting: t.Casting, Moves: t.Moves, CastDirect: len(t.Spells) == 0}
		for _, k := range t.Can {
			if c, ok := combatKinds[k]; ok {
				o.Turn.Can = append(o.Turn.Can, c)
			}
		}
		for _, op := range t.Options { // the game's own: its acts as the tabletop's, the rest by index
			k, ok := combatKinds[op.Kind]
			if !ok || op.Kind == 0 {
				k = combat.ActOption
			}
			o.Turn.Options = append(o.Turn.Options, combat.Option{Label: op.Label, Key: op.Key, Kind: k, Chords: op.Chords})
		}
	}
	return o
}

// keypadSteps are the keypad's directions (ActDir N) by square offset.
var keypadSteps = map[image.Point]int{{0, -1}: 8, {1, -1}: 9, {1, 0}: 6, {1, 1}: 3, {0, 1}: 2, {-1, 1}: 1, {-1, 0}: 4, {-1, -1}: 7}

// askedTurn: v's question is a turn the tabletop answers (AskSquare: an
// aim, whatever Turn the rules publish; AskMove: a step at a time, where they
// publish none).
func askedTurn(v front.View) bool {
	return v.Battle != nil && v.Ask != nil && (v.Ask.Kind == front.AskSquare || v.Ask.Kind == front.AskMove && v.Battle.Turn == nil)
}

// askedBattle is v's battle, with the active unit's turn its question asks
// for (askedTurn): an aim, or a move whose reach is the open squares next
// to the unit and whose End ends it.
func askedBattle(v front.View) *front.Battle {
	if !askedTurn(v) {
		return v.Battle
	}
	b := *v.Battle
	t := &front.Turn{Aim: v.Ask.Kind == front.AskSquare, Can: []front.ActKind{front.ActAim}}
	if pt := v.Battle.Turn; pt != nil { // the game's turn, now aiming (a spell's square): its unit and casting, the aim's acts only
		t.Casting, t.Weapon, t.Missile = pt.Casting, pt.Weapon, pt.Missile
	}
	var at image.Point
	for _, u := range b.Units {
		if u.Active {
			t.Unit, at = u.ID, image.Pt(u.X, u.Y)
		}
	}
	if !t.Aim {
		t.Can, b.Reach, b.Paths = []front.ActKind{front.ActEnd}, nil, nil
		for d := range keypadSteps {
			if q := at.Add(d); q.In(image.Rect(0, 0, len(b.Ground[0]), len(b.Ground))) && !b.Ground[q.Y][q.X].Wall {
				b.Reach = append(b.Reach, q)
			}
		}
	}
	b.Turn = t
	return &b
}

// askAction is the tabletop's decision d as the answer to v's question:
// on a move (askedTurn) a square walked to is that step (ActDir), End is
// done (ActOK); otherwise toAction's.
func askAction(v front.View, d combat.Decision) front.Action {
	if !askedTurn(v) || v.Ask.Kind != front.AskMove {
		return toAction(d)
	}
	switch d.Kind {
	case combat.ActEnd:
		return front.Action{Kind: front.ActOK}
	case combat.ActStep:
		return front.Action{Kind: front.ActDir, N: d.N}
	case combat.ActWalk:
		for _, u := range v.Battle.Units {
			if n, ok := keypadSteps[image.Pt(d.X-u.X, d.Y-u.Y)]; u.Active && ok {
				return front.Action{Kind: front.ActDir, N: n}
			}
		}
		return front.Action{}
	}
	return toAction(d)
}

// keepFallen is b with the units of the board drawn before the events now
// playing that b has dropped, each kept until its own fall (EvDied) plays:
// the blows land on figures, not on empty squares. Outside the replay b is
// the board kept for the next one.
func (p *Presenter) keepFallen(b *front.Battle) *front.Battle {
	if b == nil || !p.Playing() {
		p.lastBattle = b
		return b
	}
	last := p.lastBattle
	if last == nil {
		return b
	}
	o := *b
	for i, u := range last.Units {
		if slices.ContainsFunc(o.Units, func(x front.Unit) bool { return x.ID == u.ID }) ||
			!slices.ContainsFunc(p.replay.q[1:], func(e front.CombatEvent) bool { return e.Kind == front.EvDied && e.Unit == u.ID }) { // the fall showing now (q[0]) is played: the unit is down
			continue
		}
		if len(o.IconPix) == len(o.Units) && i < len(last.IconPix) {
			o.IconPix = append(slices.Clone(o.IconPix), last.IconPix[i])
		}
		o.Units = append(slices.Clone(o.Units), u)
	}
	return &o
}
