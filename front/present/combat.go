package present

import (
	"gobox/front"
	"image"

	"gobox/combat"
	"gobox/gpu"
)

// The adapter between the engine's combat snapshot and gobox/combat.

var combatKinds = map[front.ActKind]combat.Kind{
	front.ActWalk: combat.ActWalk, front.ActAttack: combat.ActAttack, front.ActAim: combat.ActAim, front.ActPoint: combat.ActPoint,
	front.ActClose: combat.ActClose, front.ActYes: combat.ActYes, front.ActNo: combat.ActNo, front.ActWait: combat.ActWait,
	front.ActEnd: combat.ActEnd, front.ActQuick: combat.ActQuick, front.ActGuard: combat.ActGuard, front.ActBandage: combat.ActBandage,
	front.ActFlee: combat.ActFlee, front.ActTurn: combat.ActTurn, front.ActCast: combat.ActCast, front.ActItems: combat.ActItems,
	front.ActRanged: combat.ActRanged, front.ActManeuver: combat.ActManeuver,
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
	p.battle.Theme, p.battle.Shift = combat.DefaultTheme, gpu.IsKeyPressed(gpu.KeyShift)
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
		Tiles: b.Tiles, WallTiles: dungeonTiles, StoneTile: 3, Floor: b.Floor, Dungeon: b.Dungeon,
		Cursor: b.Cursor, CursorOn: b.CursorOn, CursorAt: b.CursorAt, Marks: b.Marks, Reach: b.Reach,
		Targets: b.Targets, Panel: b.Panel, Log: b.Log, Area: b.Area, AreaAt: b.AreaAt, Skipped: b.Skipped,
		AimWhy: b.AimWhy, View: b.View, Strike: combat.Strike(b.Strike),
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
		o.Turn = &combat.Turn{Unit: t.Unit, Aim: t.Aim, Casting: t.Casting, Moves: t.Moves}
		for _, k := range t.Can {
			if c, ok := combatKinds[k]; ok {
				o.Turn.Can = append(o.Turn.Can, c)
			}
		}
	}
	return o
}
