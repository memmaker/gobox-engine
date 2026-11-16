package engine

import (
	"fmt"
	"strconv"

	"gobox/adventure"
	"gobox/game"
	"gobox/rules"
)

// The rocket flies on the system map (BR view kind 2; buckroger-re explore.pas MoveRocket and MapMainMenu's orbit-map arm,
// orbitmap.pas). The engine owns where the party is: the ship's square, its heading, the move counter and the orbit step are
// rules.World words; the bodies' squares are derived from the orbit step (adventure.SpaceMap.Body) and read by the scripts through
// the layout's BodySquares. The ship, its fuel and a move's time are the ruleset's (fuel is shown only: the Pascal spends none).

// rocketMenuText is the orbit-map menu's Move (explore.pas MapMainMenu 'Move Rocket Change Save'); Rocket, Change and Save are
// not ported: like the rest menu, it offers only what runs.
const rocketMenuText = "~Move ~Rocket ~Save"

// The orbit-map menu's places.
const (
	rocketMenuMove = iota
	rocketMenuRocket
	rocketMenuSave
)

// rocketPrompt is MoveRocket's prompt before the fuel (explore.pas: 'En route. Fuel = ').
const rocketPrompt = "En route. Fuel = "

// rocketDone is MoveRocket's Done: the answer that makes no move.
const (
	rocketDoneText = "Done"
	headingDone    = -1
)

// headings is how many ways a heading points (game.Heading's Dir 0..7).
const headings = 8

// fieldBodySquares is the layout's array of the bodies' squares (x, y per body).
const fieldBodySquares = "BodySquares"

// spaceMap is the game's system map; nil: the game has none.
func (e *Engine) spaceMap() *adventure.SpaceMap { return e.game().Space }

// spaceStart puts the ship where a new game starts it (gameinit.pas InitShipAndOrbit: orbit 2's body at phase 0, heading 0, the
// counters cleared). It is done once, at a new game, not at each entry to the map.
func (e *Engine) spaceStart() {
	s := e.spaceMap()
	if s == nil {
		return
	}
	q := s.Squares[s.Start]
	w := &e.world
	w.RocketX, w.RocketY, w.RocketHeading, w.OrbitTick, w.OrbitStep = uint16(q.X), uint16(q.Y), 0, 0, 0
}

// startFlying is view kind 2: the party flies the rocket.
func (e *Engine) startFlying() {
	if e.spaceMap() == nil {
		panic("ecl: view kind 2 (the rocket) in a game with no system map")
	}
	e.world.In3D, e.world.Scripted, e.world.Flying = 0, 0, 1
}

// spaceFields binds the layout's BodySquares: the bodies' squares at the orbit step, read only.
func (e *Engine) spaceFields() map[string]engineField {
	return map[string]engineField{fieldBodySquares: {get: func(i int) (int64, bool) {
		s := e.spaceMap()
		if s == nil || i < 0 || i >= 2*s.Bodies() {
			return 0, false
		}
		at := s.Body(i/2, int(e.world.OrbitStep)).At
		return int64([]int{at.X, at.Y}[i%2]), true
	}}}
}

// spaceHeading is the other numbering of a heading: the map's (SpaceMap.Steps: counterclockwise from north) and the front's
// (Where.Facing, game.Heading: clockwise); the same turn in both directions.
func spaceHeading(h int) int { return -h & 7 }

// flyBegin is the flying party's command: the orbit-map menu, or (after a move) the next move's heading, as the Pascal's move mode.
func (e *Engine) flyBegin() {
	e.commandEnd()
	if e.run.rocketMoving {
		e.askHeading()
	} else {
		e.askRocketMenu()
	}
	e.openAsk()
}

// askRocketMenu asks the orbit-map menu.
func (e *Engine) askRocketMenu() {
	e.askChoice("", rocketMenuText, false, false, askDone{kind: dRocketMenu})
}

// askHeading asks MoveRocket's heading, with Done: the fuel as the ship has it.
func (e *Engine) askHeading() {
	fuel := e.fact(rules.NoChar, shipCells["ShipFuel"]).(rules.Number)
	q := &game.Question{Text: fmt.Sprintf("%s%d ", rocketPrompt, int(fuel)), Heading: true, Replies: []string{rocketDoneText}, Keys: []rune{'D'},
		Esc: game.EscTo(0)}
	e.eclWait = &eclWait{ask: q, shape: shHeading, done: askDone{kind: dRocketMove}}
}

// rocketMenuAnswered: Move starts the move mode, Rocket shows the own ship's record until its key, Save is SaveQuit.
func (e *Engine) rocketMenuAnswered(n int) {
	switch n {
	case rocketMenuMove:
		e.run.rocketMoving = true
		e.askHeading()
	case rocketMenuRocket:
		sh, ok := e.ShipSheetState(true, true)
		if !ok {
			panic("rocket menu: the own ship has no record")
		}
		e.ext.screen = &game.Screen{Kind: game.ScreenShipSheet, ShipSheet: &sh}
		e.eclWait = &eclWait{ask: &game.Question{Continue: true}, shape: shPause, done: askDone{kind: dRocketSheet}}
	case rocketMenuSave:
		e.saveQuit(dRocketSaveQuit)
	}
}

// rocketMoved takes the heading asked (headingDone: none, the move mode ends), makes the move and runs the area's step entry
// (the Pascal's loop: the menu, the move, then the step events).
func (e *Engine) rocketMoved(dir int) {
	if dir == headingDone {
		e.run.rocketMoving = false
	} else {
		e.rocketStep(spaceHeading(dir))
	}
	e.push(cont{kind: kStepEnd}, cont{kind: kAreaChanged})
	e.runMoveEntry()
}

// rocketStep is MoveRocket from the heading on: the ship points that way and steps, stopping at the map's edge and never onto the
// sun; the team rests the move's days and, each time the counter wraps, the bodies move on an orbit step.
func (e *Engine) rocketStep(heading int) {
	s, w := e.spaceMap(), &e.world
	w.RocketHeading = uint16(heading)
	at := rules.MapPos{X: int(w.RocketX), Y: int(w.RocketY)}
	step := s.Steps[heading]
	to := rules.MapPos{X: min(max(at.X+step.X, 0), s.Size.W-1), Y: min(max(at.Y+step.Y, 0), s.Size.H-1)}
	if to == s.Sun {
		to = rules.MapPos{X: to.X - step.X, Y: to.Y - step.Y}
	}
	w.RocketX, w.RocketY = uint16(to.X), uint16(to.Y)
	w.OrbitTick += uint16(s.Move.Ticks)
	e.restDays(s.Move.Days)
	if int(w.OrbitTick) >= s.Move.TickCycle {
		w.OrbitTick -= uint16(s.Move.TickCycle)
		if w.OrbitStep == 0 {
			w.OrbitStep = uint16(s.Cycle)
		}
		w.OrbitStep--
	}
}

// spaceMapState is the system map as the front draws it: the sun and the bodies as terrain, the ship's square and heading as
// where the party is.
func (e *Engine) spaceMapState(w *game.Where) game.Automap {
	s := e.spaceMap()
	w.Mode, w.X, w.Y, w.Square = "space", int(e.world.RocketX), int(e.world.RocketY), nil
	w.Facing = spaceHeading(int(e.world.RocketHeading))
	m := game.Automap{Area: e.mapName(), Size: s.Size, Marker: "ship", Terrain: make([]int, s.Size.W*s.Size.H)}
	for y := range s.Size.H {
		for x := range s.Size.W {
			m.Known = append(m.Known, rules.MapPos{X: x, Y: y})
		}
	}
	m.Terrain[s.Sun.Y*s.Size.W+s.Sun.X] = s.SunTile
	for i := range s.Bodies() {
		b := s.Body(i, int(e.world.OrbitStep))
		m.Terrain[b.At.Y*s.Size.W+b.At.X] = b.Tile
	}
	return m
}

// answerHeading answers the open heading question with the front's heading.
func (e *Engine) answerHeading(h game.Heading) error {
	if q := e.question(); q == nil || !q.Heading {
		return rules.ErrNotNow
	}
	e.answer(h)
	return nil
}

// spaceArt adds the system map's pictures to art: the sun's and the bodies' tiles as terrain/<id> (a body's Tile, frame id - SunTile of
// the symbol set) and the ship as ship/<facing> (the front's clockwise heading, Where.Facing), the map's marker.
func (e *Engine) spaceArt(art map[string]game.Picture) {
	s := e.spaceMap()
	if s == nil {
		return
	}
	im := e.LoadPicMasked(s.Tiles.File, s.Tiles.Block, symbolMask)
	if im == nil {
		panic(fmt.Sprintf("the system map's tiles %s-%d are not in the art", s.Tiles.File, s.Tiles.Block))
	}
	for f := range im.Count {
		art["terrain/"+strconv.Itoa(s.SunTile+f)] = gamePicture(e.picPix(im, f, false, nil))
	}
	for dir := range headings {
		art["ship/"+strconv.Itoa(dir)] = gamePicture(e.picPix(im, s.ShipFrame+spaceHeading(dir), false, nil))
	}
}
