package engine

import (
	"fmt"
	"gobox/game"
	"gobox/rules"
	"slices"
)

// GameLoop = ovr04:32F8: load the current area, run its start events, then
// open exploring (ext.explore) until a script quits. The original's loop
// is command: each command that ends input goes on in afterCommand. What
// follows each script run is a cont on the engine's stack (cont.go).
func (e *Engine) GameLoop() {
	r := &e.run
	e.ext.inGame = true
	r.walldefSet, r.geoSet = 0, 0
	r.restoreSel = false
	r.scriptLoad = 0
	r.needFrame = 1
	e.setWorldWord(rules.WordBeginFlag, 4)
	r.areaChanged = false
	e.gs.areaKind = byte(e.world.AreaKind)
	if e.gs.areaKind == 0 { // no last block: the game's first (POR's is 0)
		e.gs.areaKind = byte(e.game().FirstBlock)
	}
	fresh := !e.loadedGame // extra: a new adventure, not a loaded game
	if k := e.gs.areaKind; (k == 0x19 || k == 0x1A || k == 0x1B) && e.worldWord(rules.WordIn3D) == 0 {
		e.setWorldWord(rules.WordBeginFlag, 3)
	}
	// the area's text script, also for a loaded game (whose save holds
	// the byte code)
	e.LoadArea(int(e.gs.areaKind))
	e.InitAreaScript()
	if e.Explore { // -explore: no guided tour of Phlan either
		e.Explore = false
		if r, ok := e.layoutField(tourDone); ok && fresh && e.gs.areaKind == 0 {
			e.host.Set(r, 1)
		}
	}
	e.push(cont{kind: kGameBegan, flag: fresh})
	e.RunECL(e.es.events[4])
}

// gameBegan: the area's start events are done; exploring opens.
func (e *Engine) gameBegan(fresh bool) {
	r := &e.run
	if r.areaChanged { // the start script switched areas (sotsb's opening, area 17's NEW ECL 16): the new one's start events before any input (sotsb-re, cotab-re, pod-re GameLoop: NewEclDone)
		e.push(cont{kind: kGameBegan, flag: fresh})
		e.AreaChange()
		return
	}
	if e.worldWord(rules.WordBeginFlag) != 3 {
		if r.needFrame == 1 {
			e.DrawMainScreen()
			r.needFrame = 0
		}
		e.Redraw()
	}
	if fresh {
		e.Music(4) // extra: the party first enters Phlan (the Amiga never plays song 4)
	}
	e.world.AreaKind = uint16(e.gs.areaKind)
	e.looking, e.over = false, false
	e.ext.explore, e.ext.exploreSeq = true, e.opens()
	e.commandBegin()
}

// afterCommand is GameLoop's loop after a command ended input (encamp:
// it was camp): the look and camp loop, else the step and its scripts;
// then the next command, or the game is over.
func (e *Engine) afterCommand(encamp bool) {
	r := &e.run
	r.savedSel = e.selected()
	if !e.looking {
		e.world.AreaKind = uint16(e.gs.areaKind)
	}
	if e.ext.interacting { // extra: Space on a square whose event waits (interact.go)
		e.ext.interacting = false
		if p := e.ext.pre; p != nil && !p.ran {
			e.preStep(p)
			return
		}
		if e.ext.owed != nil { // gobox's RunEvent: the rest of the step Arrive moved
			e.ext.owed = nil
			e.es.picShown, e.es.picAnimating = 0, 1
			e.push(cont{kind: kStepEnd}, cont{kind: kAreaChanged})
			e.stepEvents()
			return
		}
		e.interactNow()
		return
	}
	// Look / encamp: repeat while the last command was one of them.
	if (e.world.Search > 1 || encamp) && !r.quit {
		e.looking = true
		if encamp {
			e.push(cont{kind: kLookDone})
			e.Encamp()
			return
		}
		r.searchSave = byte(int(e.world.Search) & 1)
		e.world.Search = 1
		e.Redraw()
		e.push(cont{kind: kSearched, n: [3]int{e.ext.opened, e.posX, e.posY}})
		e.RunECL(e.es.events[1])
		return
	}
	e.looking = false
	e.stepDone()
}

// searched: the look's events are done (opened, x, y: as they began).
func (e *Engine) searched(opened, x, y int) {
	if e.ext.opened == opened && e.posX == x && e.posY == y && !e.run.quit { // extra: say that the search happened
		e.say("Nothing found")
	}
	e.push(cont{kind: kSearchRestore})
	e.areaChanged()
}

// lookDone ends one pass of the look and camp loop.
func (e *Engine) lookDone() {
	if !e.run.quit {
		e.commandBegin()
		return
	}
	e.looking = false
	e.stepDone()
}

// stepDone is the step part of GameLoop's loop, then the next command.
func (e *Engine) stepDone() {
	r := &e.run
	split := e.ext.arriving // gobox's Arrive: the move only; RunEvent runs the step's events
	e.ext.arriving = false
	p := e.ext.pre
	e.ext.pre = nil
	if split && p != nil && p.ran && p.x == e.posX && p.y == e.posY && p.side == e.moveDir() {
		e.world.HoldMove = p.hold // RunEvent{Before} ran the pre-step script for this step
		e.stepMove(split)
		return
	}
	e.push(cont{kind: kStepPre, flag: split})
	if r.quit {
		e.ret()
		return
	}
	e.RunECL(e.es.events[0])
}

// stepPre: the pre-step script is done.
func (e *Engine) stepPre(split bool) {
	switch r := &e.run; {
	case r.areaChanged:
		e.unback() // a backward step the scripts sent elsewhere
		e.push(cont{kind: kStepEnd})
		e.AreaChange()
	case r.quit:
		e.stepEnd()
	default:
		e.stepMove(split)
	}
}

// stepMove is the step's move, then its events.
func (e *Engine) stepMove(split bool) {
	e.world.PrevX = uint16(int16(int8(e.posX)))
	e.world.PrevY = uint16(int16(int8(e.posY)))
	e.push(cont{kind: kMoved, flag: split})
	e.Move()
}

// moved: the step is made.
func (e *Engine) moved(split bool) {
	e.Redraw()
	if int(e.world.PrevX) != e.posX || int(e.world.PrevY) != e.posY {
		e.cue(rules.CueStep)
	}
	if split {
		e.ext.owed = &[2]int{e.posX, e.posY}
		e.stepEnd()
		return
	}
	e.es.picShown, e.es.picAnimating = 0, 1
	e.push(cont{kind: kStepEnd}, cont{kind: kAreaChanged})
	e.stepEvents()
}

// preStep runs the pre-step script (events[0]) for a step through p.side
// ahead of the move (gobox's RunEvent{Before}); the step's Arrive then
// skips it. A hold (HoldMove) is kept for that Arrive.
func (e *Engine) preStep(p *preStep) {
	r := &e.run
	p.ran = true
	kept := e.facing
	if e.viewMode == 1 {
		e.ext.back = (p.side - e.facing + 8) % 8
	} else {
		e.facing = p.side // goCmd's: the script reads the way; Arrive sets it again
	}
	e.es.wallAhead = e.WallAt(e.posX, e.posY, e.moveDir()) // the scripts read it
	if e.viewMode == 1 {
		e.checkEdge() // as the step command does ahead of events[0]: the area exits read LeftMap (a stale one sent the next step out any way)
	}
	e.push(cont{kind: kPreEnd, pre: p, n: [3]int{kept}})
	if r.quit {
		e.ret()
		return
	}
	e.RunECL(e.es.events[0])
}

// preEnd: the pre-step script is done (kept: the facing before it).
func (e *Engine) preEnd(p *preStep, kept int) {
	switch r := &e.run; {
	case r.areaChanged:
		e.ext.pre = nil
		e.unback()
		e.push(cont{kind: kStepEnd})
		e.AreaChange()
	case r.quit:
		e.ext.pre = nil
		e.stepEnd()
	default:
		p.hold = uint16(e.world.HoldMove)
		if e.viewMode != 1 && e.facing == p.side {
			e.facing = kept // the facing shows the step's way only once Arrive takes it
		}
		e.stepEnd()
	}
}

// stepEnd ends a step: the next command, or the game is over.
func (e *Engine) stepEnd() {
	r := &e.run
	if r.quit {
		r.quit = false
		e.over = true // the explore screen goes: endGame
		return
	}
	e.commandBegin() // ovr04:34FC jumps back to 3395: the write follows the input
}

// areaChanged runs AreaChange when a script switched areas, else goes on.
func (e *Engine) areaChanged() {
	if e.run.areaChanged {
		e.AreaChange()
		return
	}
	e.ret()
}

// AreaChange = ovr04:31B4: after a script switched areas, run the new
// area's start events and settle the view mode; then what follows.
func (e *Engine) AreaChange() {
	e.es.cell = e.MapCell(e.posX, e.posY)
	e.world.LeftMap = 0
	e.ruleset()         // charID
	sel := e.selected() // DS:5D92 kept at bp-4 (ovr04:31D5)
	e.prevViewMode = e.viewMode
	e.viewMode = 1
	e.push(cont{kind: kAreaStart, sel: sel})
	e.RunECL(e.es.events[4])
}

// areaStart: the new area's start events are done.
func (e *Engine) areaStart(sel rules.CharID) {
	r := &e.run
	e.setWorldWord(rules.WordBeginFlag, 4)
	if e.worldWord(rules.WordIn3D) == 0 {
		switch e.gs.areaKind {
		case 0x19:
			e.viewMode = 2
			e.setWorldWord(rules.WordBeginFlag, 3)
		case 0x1A:
			e.viewMode = 3
			e.setWorldWord(rules.WordBeginFlag, 3)
			if e.worldWord(rules.WordBoss) == 0xFE && e.world.RiverOpen == 0 {
				e.EndgameHook(1)
			}
		case 0x1B:
			e.viewMode = 4
			e.setWorldWord(rules.WordBeginFlag, 3)
		}
	}
	if e.prevViewMode > 1 && e.viewMode == 1 {
		e.clearView() // the picture frame's text
	}
	if (e.prevViewMode > 1 || e.viewMode == 1) && r.scriptLoad != 0 ||
		e.prevViewMode == 1 && e.viewMode == 1 {
		e.Redraw()
	}
	r.areaChanged = false
	e.push(cont{kind: kAreaStep, sel: sel})
	e.RunECL(e.es.events[0])
}

// areaStep: the new area's step events are done.
func (e *Engine) areaStep(sel rules.CharID) {
	if e.run.areaChanged {
		e.AreaChange()
		return
	}
	e.push(cont{kind: kAreaSettled, sel: sel})
	e.stepEvents()
}

// areaSettled: the area's square events are done.
func (e *Engine) areaSettled(sel rules.CharID) {
	e.selectChar(sel)
	if e.run.areaChanged {
		e.AreaChange()
		return
	}
	e.ret()
}

// Encamp = ovr04:2DD3: the camp event, the camp menu, then the area's
// after-camp event; then what follows.
func (e *Engine) Encamp() {
	e.push(cont{kind: kCamped})
	e.RunECL(e.es.events[2])
}

// campEnded: camp is left (interrupted: a rest met an encounter).
func (e *Engine) campEnded(interrupted bool) {
	if !interrupted {
		e.campOver()
		return
	}
	e.DrawMainScreen()
	e.push(cont{kind: kCampOver})
	e.RunECL(e.es.events[3])
}

// campOver: the camp's events are done.
func (e *Engine) campOver() {
	e.Redraw()
	e.ret()
}

// commandBegin = the start of CommandInput (ovr15:0989): the explore
// screen then waits for the adventuring decisions until one ends input
// (camp, search here, a step: commandEnd).
func (e *Engine) commandBegin() {
	e.world.HoldMove = 0
	e.areaMusic()
	e.unback() // a backward step the scripts ended
	if e.viewMode >= 2 && e.viewMode <= 4 && e.prevViewMode < 2 {
		e.loadAreaIcon(0, 9)
		e.loadAreaIcon(1, 0xA)
		e.loadAreaIcon(2, 0xB)
		e.areaMapIcons() // 00BA:2B5E
	}
	if e.viewMode >= 1 && e.viewMode <= 4 {
		e.awaitCommand()
		return
	}
	e.commandEnd()
	e.afterCommand(false)
}

// awaitCommand: the explore screen waits for the next command.
func (e *Engine) awaitCommand() {
	if f := e.DevAction; f != nil { // headless checks (-dev)
		e.DevAction = nil
		f()
	}
}

// commandEnd = the end of CommandInput.
func (e *Engine) commandEnd() {
	e.prevViewMode = e.viewMode
	if e.run.textCleared == 0 {
		e.clearPane()
		e.run.textCleared = 1
	}
}

// exploreVerb makes one adventuring decision: verb says whether it ended input
// (the step or search is then made by GameLoop) and whether it encamps. The
// game over (a script quit), exploring ends.
func (e *Engine) exploreVerb(verb func() (ends, encamp bool)) error {
	if !e.exploring() {
		return rules.ErrNotNow
	}
	if ends, encamp := verb(); ends {
		e.commandEnd()
		e.afterCommand(encamp)
	} else {
		e.awaitCommand()
	}
	if e.over && e.top() == onExplore {
		e.endExplore()
	}
	return nil
}

// endExplore: the game is over; back to the party menu.
func (e *Engine) endExplore() {
	e.ext.explore = false
	e.endGame()
}

// stepCmd makes a step way (0 ahead; extra: 4 back, 6 or 2 aside) from the
// facing, which never changes for those (the original turns round): Move
// goes that way (moveDir).
func (e *Engine) stepCmd(way int) (ends, encamp bool) {
	if e.viewMode != 1 {
		return false, false
	}
	if way != 0 {
		e.ext.back = way
		e.es.wallAhead = e.WallAt(e.posX, e.posY, e.moveDir()) // the step scripts read it (City Hall: 0xC04E 11 is the locked door)
	}
	e.checkEdge()
	return true, false
}

// turnCmd turns the party d (2 right, 6 left) in the 3D view.
func (e *Engine) turnCmd(d int) (ends, encamp bool) {
	if e.viewMode == 1 {
		e.turn(d, true)
	}
	return false, false
}

// goCmd is a step way n (0..3) on the area map and overland.
func (e *Engine) goCmd(n int) (ends, encamp bool) {
	if e.viewMode == 1 || n < 0 || n >= 4 {
		return false, false
	}
	e.facing = n * 2
	return true, false
}

// interactCmd runs the event waiting here (interact.go), or the step's owed.
func (e *Engine) interactCmd() (ends, encamp bool) {
	if e.interactHere() || e.stepOwed() || e.ext.pre != nil && !e.ext.pre.ran {
		e.ext.interacting = true
		return true, false
	}
	return false, false
}

// lookCmd is the original's Look: search this square once.
func (e *Engine) lookCmd() (ends, encamp bool) {
	e.world.Search |= 2
	e.searchSecrets()
	e.searchTime()
	e.vm.PC = uint16(e.es.events[1])
	return true, false
}

// sweepCmd is the original's Search: the slow, searching pace on or off.
func (e *Engine) sweepCmd() (ends, encamp bool) {
	e.world.Search ^= 1
	e.say(map[bool]string{true: "Sweeping", false: "Sweeping off"}[e.world.Search&1 != 0]) // extra: the toggle shows
	return false, false
}

// turn = the turn keys of ovr15:0989: face dir+d, look again.
func (e *Engine) turn(d int, sound bool) {
	e.facing = (e.facing + d) % 8
	if sound {
		e.cue(rules.CueStep)
	}
	e.es.wallAhead = e.WallAt(e.posX, e.posY, e.facing)
	e.show3D() // ovr30:05FD Draw3DView: the remaster draws the view
	e.viewDrawn()
}

// checkEdge = ovr15:068D: game+0x5AA is set when stepping through an open
// side would leave the map.
func (e *Engine) checkEdge() {
	x, y := e.posX, e.posY
	e.world.LeftMap = 0
	if e.DoorAt(x, y, e.moveDir()) == 0 {
		return
	}
	x += rules.StepX[e.moveDir()]
	y += rules.StepY[e.moveDir()]
	if x > 15 {
		e.posX = 15
		e.world.LeftMap = 1
	}
	if x < 0 {
		e.posX = 0
		e.world.LeftMap = 1
	}
	if y > 15 {
		e.posY = 15
		e.world.LeftMap = 1
	}
	if y < 0 {
		e.posY = 0
		e.world.LeftMap = 1
	}
}

// moveDir (extra) is the way the next step goes: the facing, or back or
// aside of it for a step backwards or aside (the facing stays).
func (e *Engine) moveDir() int {
	return (e.facing + e.ext.back) % 8
}

// unback (extra): the step backwards or aside is made; the next goes ahead.
func (e *Engine) unback() {
	if e.ext.back != 0 {
		e.ext.back = 0
		e.es.wallAhead = e.WallAt(e.posX, e.posY, e.facing)
	}
}

// Move = ovr15:0D6D: step forward in the 3D view (through a door if it can
// be opened) or across the area map; then what follows.
func (e *Engine) Move() {
	switch {
	case e.viewMode == 1:
		if e.world.HoldMove >= 0xFF {
			e.world.HoldMove = 0
			break
		}
		switch e.DoorAt(e.posX, e.posY, e.moveDir()) {
		case 1:
			if e.WallAt(e.posX, e.posY, e.moveDir()) != 0 {
				e.emit(rules.Sounded{Cue: rules.SoundCue("opendoor")})
			}
			e.moveStep(true)
		case 2, 3:
			e.push(cont{kind: kMoveStep})
			e.lockedDoor(e.DoorAt(e.posX, e.posY, e.moveDir()) == 3)
		default:
			e.moveStep(false)
		}
		return
	case e.viewMode >= 2 && e.viewMode <= 4:
		if e.world.HoldMove < 0xFF {
			e.mapMove()
		} else {
			e.world.HoldMove = 0
		}
	}
	e.moveEnd()
}

// moveStep is the step through the way ahead (moved: the party may go
// through it).
func (e *Engine) moveStep(moved bool) {
	// extra: a way off the map no script caught bumps (the original wrapped)
	if moved && rules.OnMap(e.posX+rules.StepX[e.moveDir()], e.posY+rules.StepY[e.moveDir()]) {
		e.stepAhead()
	} else {
		e.cue(rules.CueBump)
	}
	e.unback()
	e.moveEnd()
}

// moveEnd ends the move.
func (e *Engine) moveEnd() {
	e.FreeAnimPic()
	e.port = portrait{headID: 0xFF, bodyID: 0xFF} // 0464:2595 x2
	e.unback()
	e.ret()
}

// lockedDoor = the door menu of ovr15:0D6D, the wait PendDoor
// (Door) until TryDoor. A plain locked door (state 2) offers Bash,
// Pick and Knock; a wizard-locked one (3) only lets Bash work. Whether the
// party goes through is ext.doorThrough, then what follows.
func (e *Engine) lockedDoor(wizard bool) {
	var d DoorOpts
	d.Bash = e.run.doorTry[0] != 0
	d.Pick = e.run.doorTry[1] != 0 && e.ruleset().Check(rules.PartyHasSkill{N: 6}, nil).Yes
	knock := false
	if wizard {
		knock = e.ruleset().Check(rules.PartyHasSkill{N: 5}, nil).Yes
	} else {
		knock = e.ruleset().Check(rules.HasMemorized{Party: e.partyIDs(), Spell: rules.SpellKnock}, nil).Yes
	}
	d.Knock = e.run.doorTry[2] != 0 && knock
	if d == (DoorOpts{}) {
		e.setDoor(e.posX, e.posY, e.moveDir(), 0)
		e.ext.doorThrough = false
		e.ret()
		return
	}
	w := &doorWait{wizard: wizard, opts: d}
	e.ext.door = w
	e.wait(PendDoor)
}

// TryDoor = the locked door's Bash, Pick, Knock or Exit: the party goes
// through or not, and the step goes on.
func (e *Engine) TryDoor(k game.DoorWay) error {
	w := e.ext.door
	if !e.waiting(PendDoor) || w == nil ||
		k == game.DoorBash && !w.opts.Bash || k == game.DoorPick && !w.opts.Pick || k == game.DoorKnock && !w.opts.Knock {
		return rules.ErrNotNow
	}
	switch k {
	case game.DoorBash:
		if w.through = e.bashDoor(); !w.through {
			e.emit(rules.Sounded{Cue: rules.SoundCue("hitwall")})
		} else {
			e.emit(rules.Sounded{Cue: rules.SoundCue("shutdoor")}) // a slam
		}
	case game.DoorPick:
		switch {
		case w.wizard:
			e.run.doorTry[1] = 0
		case !e.pickLock():
			e.emit(rules.Sounded{Cue: rules.SoundCue("lockpick_fail")})
		default:
			e.emit(rules.Sounded{Cue: rules.SoundCue("lockpick")})
			w.through = true
		}
	case game.DoorKnock:
		w.through = e.knock()
	}
	e.event(game.DoorTried{Way: k, Through: w.through}) // gobox flashes the outcome (the original only bumped)
	e.resume()
	return nil
}

// setDoor = ovr15:0000 (state 0) and ovr15:0109 (state 1): set the door
// bits of one side of a cell. 0109 ignores cells off the map.
func (e *Engine) setDoor(x, y, dir int, state byte) {
	if !rules.OnMap(x, y) {
		return
	}
	shift := uint(dir) // N 0, E 2, S 4, W 6
	i := 0x300 + y<<4 + x
	e.gs.geo.Plane[i] = e.gs.geo.Plane[i]&^(3<<shift) | state<<shift
}

// openDoor opens the side ahead from both cells (ovr15:0109 twice).
// Extra: the door stays open (forced, picked or knocked), kept with the
// save (DOORS.GOB) and put back whenever the map is loaded (reopenDoors).
func (e *Engine) openDoor() {
	x, y, d := e.posX, e.posY, e.moveDir()
	for _, s := range [2][3]int{{x, y, d}, {x + rules.StepX[d], y + rules.StepY[d], (d + 4) % 8}} {
		e.setDoor(s[0], s[1], s[2], 1)
		if !rules.OnMap(s[0], s[1]) {
			continue
		}
		if e.ext.doors == nil {
			e.ext.doors = map[int][]doorSide{}
		}
		k, ds := e.areaKey(), doorSide{X: s[0], Y: s[1], Dir: s[2]}
		if !slices.Contains(e.ext.doors[k], ds) {
			e.ext.doors[k] = append(e.ext.doors[k], ds)
		}
	}
}

// reopenDoors opens again the doors of the map once opened (openDoor).
func (e *Engine) reopenDoors() {
	for _, ds := range e.ext.doors[e.areaKey()] {
		e.setDoor(ds.X, ds.Y, ds.Dir, 1)
	}
}

// bashDoor = ovr15:0294: each member in turn tries to force the door.
func (e *Engine) bashDoor() bool {
	door := rules.Door{State: rules.DoorLocked}
	if e.DoorAt(e.posX, e.posY, e.moveDir()) == 3 {
		door.State = rules.DoorMagic
	}
	return e.tryDoor(door, func(id rules.CharID) rules.Action { return rules.BashDoor{Who: id} })
}

// pickLock = ovr15:04DF: d100 against each healthy member's +0x78 skill.
func (e *Engine) pickLock() bool {
	ok := e.tryDoor(rules.Door{State: rules.DoorLocked}, func(id rules.CharID) rules.Action { return rules.PickLock{Who: id} })
	e.run.doorTry[1] = 0
	return ok
}

// tryDoor has each member in turn try the door ahead until it gives way,
// and opens it then.
func (e *Engine) tryDoor(door rules.Door, try func(rules.CharID) rules.Action) bool {
	at := rules.AtDoor{Party: e.partyIDs(), Door: door}
	for _, id := range at.Party {
		o, err := e.ruleset().Resolve(at, try(id), nil)
		if err != nil {
			panic(err)
		}
		for _, f := range o.Effects {
			switch f.(type) {
			case rules.DoorTriesSpent:
				e.run.doorTry[0] = 0
			case rules.DoorGaveWay:
				e.openDoor()
				return true
			}
		}
	}
	return false
}

// knock = ovr15:0648: use up a memorized Knock on the door ahead.
func (e *Engine) knock() bool {
	out, err := e.ruleset().Resolve(nil, rules.SpendKnock{}, nil)
	if err != nil {
		panic(fmt.Sprintf("knock: %v", err)) // the action never refuses
	}
	if len(out.Effects) == 0 {
		return false
	}
	e.openDoor() // both sides, as Bash and Pick (the Pascal's fix: the original left it locked)
	return true
}

// stepAhead = ovr15:073B: walk one cell, look, and let time pass. extra:
// the original wrapped an off-map step to the far edge; Move blocks it.
func (e *Engine) stepAhead() {
	// the step sound plays once the loop sees the square change (ovr04:3395)
	e.foundSecret(e.posX, e.posY, e.moveDir())
	e.posX += rules.StepX[e.moveDir()]
	e.posY += rules.StepY[e.moveDir()]
	e.es.wallAhead = e.WallAt(e.posX, e.posY, e.facing)
	e.run.doorTry = [3]byte{1, 1, 1}
	e.es.cell = e.MapCell(e.posX, e.posY)
	if e.world.Search&1 != 0 {
		e.searchSecrets() // sweeping: every square stepped on is searched
		e.searchTime()
	} else {
		e.as(e.sel)
		e.do(rules.AdvanceClock{Count: 1, Unit: clockMinute})
	}
}

// mapMove = ovr15:0818: one step on the area map, unless the target square
// is impassable (DS:3648 map, 0x2C bytes a row).
func (e *Engine) mapMove() {
	x := min(max(int(int8(byte(e.worldWord(rules.WordMapX)+rules.StepX[e.facing]))), 0), 0xF)
	y := min(max(int(int8(byte(e.worldWord(rules.WordMapY)+rules.StepY[e.facing]))), 0), 0x23)
	col := int(int8(byte(int(ViewModeShift[e.viewMode]) + x)))
	t := e.ruleset().Check(rules.MapTerrain{X: col, Y: y}, nil).Value
	switch {
	case t == 0x4E, t >= 0x7E && t <= 0x80, t >= 0x82 && t <= 0x84,
		t >= 0xA8 && t <= 0xAC, t >= 0xAE && t <= 0xB3, t == 0xBC,
		t >= 0xDD && t <= 0xDE, t >= 0xE1 && t <= 0xE9, t == 0xF5:
		return
	case t >= 0x66 && t <= 0x6C, t == 0x72:
		if e.world.RiverOpen == 0 {
			e.DisplayMessageWait(game.InkYellow, "The river's stench drives you back.")
			return
		}
	}
	e.setWorldWord(rules.WordMapX, int(uint16(int16(x))))
	e.setWorldWord(rules.WordMapY, int(uint16(int16(y))))
	e.as(e.sel)
	e.do(rules.AdvanceClock{Count: 0xC, Unit: 3})
}

// searchTime is a searching step's or a Look's time: one turn, as the
// Pascal's MovePartyForward3D, or 3 minutes under the QuickSearch option.
func (e *Engine) searchTime() {
	if e.QuickSearch.Load() {
		e.as(e.sel)
		e.do(rules.AdvanceClock{Count: 3, Unit: clockMinute})
		return
	}
	e.as(e.sel)
	e.do(rules.AdvanceClock{Count: 1, Unit: clockTurn})
}

// exploring: the adventuring commands wait (nothing over them).
func (e *Engine) exploring() bool {
	return e.top() == onExplore && !e.over && !e.view.on && !e.snap.invOn
}

// tourDone is area 0's TourDone (layout/area0.txt): Rolf's tour runs while it is 0.
const tourDone = "TourDone"

// clockLimits = DS:363A: rounds per minute, minutes per ten, tens per
// hour, hours, days, months, years.
// clockMinute and clockTurn are AdvanceClock's units: a minute and a turn
// of ten minutes (C_Time.MINUTES, C_Time.TURNS).
const (
	clockMinute = 1
	clockTurn   = 2
)

// doorWait is the locked door waiting (PendDoor).
type doorWait struct {
	wizard  bool
	opts    DoorOpts
	through bool // the party goes through
}

// doorSide is one side of a cell whose door was opened for good.
type doorSide struct{ X, Y, Dir int }
