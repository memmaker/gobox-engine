package pool

import (
	"image"

	"gobox/game"
)

// Map is the area the party walks (gobox moves it, Arrive and Face):
// nil unless the party explores with nothing owed.
func (e *Engine) walkMap() game.Map {
	if !e.exploring() || e.powersView() != nil || e.question() != nil {
		return nil
	}
	return walkMap{e}
}

type walkMap struct{ e *Engine }

func (m walkMap) Name() string { return m.e.mapName() }

func (m walkMap) Size() image.Point {
	if m.e.viewMode != 1 {
		return image.Pt(44, 36) // the overland (mapState)
	}
	return image.Pt(16, 16)
}

// Edge: DoorAt's 0 a wall, 1 open, 2 locked, 3 wizard locked (Move tries
// those). The area map and overland are open every way.
func (m walkMap) Edge(at image.Point, side int) game.Edge {
	if m.e.viewMode != 1 {
		return game.Edge{Kind: game.EdgeOpen}
	}
	switch m.e.DoorAt(at.X, at.Y, side) {
	case 0:
		return game.Edge{Kind: game.EdgeWall}
	case 1:
		return game.Edge{Kind: game.EdgeOpen}
	case 2:
		return game.Edge{Kind: game.EdgeDoor, State: game.DoorLocked}
	}
	return game.Edge{Kind: game.EdgeDoor, State: game.DoorWizardLocked}
}

// Departure: the area's pre-step script runs ahead of every step
// (RunEvent{Before}).
func (m walkMap) Departure(at image.Point, side int) bool { return true }

// Event: the party's square, where the engine keeps the waiting event
// (interactHere) or the step's owed events (stepOwed) by PosX, PosY. On the
// overland at is the map's square (overlandSquare), not PosX, PosY.
func (m walkMap) Event(at image.Point) bool {
	here := image.Pt(m.e.posX, m.e.posY)
	if m.e.viewMode != 1 {
		here = m.e.overlandSquare()
	}
	return at == here && (m.e.interactHere() || m.e.stepOwed())
}
