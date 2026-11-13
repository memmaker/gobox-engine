package engine

import (
	"fmt"
	"image"
	"strconv"

	"gobox/core"
	"gobox/game"
	"gobox/game/play"
)

// The exploring verbs as gobox walks: RunEvent{Before} runs the pre-step
// script, Arrive the step itself, RunEvent its events; Face turns.

// runEvent runs, with before, the pre-step script for a step to side
// (0 2 4 6); else the step's events Arrive left owed, or the square's
// waiting one.
func (e *Engine) runEvent(before bool, side int) error {
	if before {
		if !e.exploring() || side&1 != 0 {
			return core.ErrNotNow
		}
		e.ext.pre = &preStep{x: e.posX, y: e.posY, side: side & 6}
	}
	return e.exploreVerb(e.interactCmd)
}

// arrive is gobox's move of the party one square from from to to, facing:
// the move (the door, the clock); RunEvent then runs its events.
func (e *Engine) arrive(from, to image.Point, facing int) error {
	e.ext.owed = nil
	way := dirOf(to.Sub(from))
	if p := e.ext.pre; p == nil || !p.ran || p.x != e.posX || p.y != e.posY || p.side != way {
		return core.ErrNotNow // no RunEvent{Before} ran the pre-step script for this step: Arrive runs none
	}
	if e.viewMode != 1 { // the area map and overland: a step that way
		e.ext.arriving = true
		err := e.exploreVerb(func() (bool, bool) { return e.goCmd(way / 2) })
		e.ext.arriving = false
		return err
	}
	if err := e.faceTo(facing); err != nil {
		return err
	}
	rel := (way - facing + 8) % 8
	if rel%2 != 0 {
		return &play.Refusal{Text: "not a step"}
	}
	e.ext.arriving = true
	err := e.exploreVerb(func() (bool, bool) { return e.stepCmd(rel) })
	e.ext.arriving = false
	return err
}

// faceTo turns the party to f (0 2 4 6) the original's way: right or left
// turns, twice for about.
func (e *Engine) faceTo(f int) error {
	switch (f - e.facing + 8) % 8 {
	case 0:
		return nil
	case 2:
		return e.turnBy(2)
	case 6:
		return e.turnBy(6)
	case 4:
		if err := e.turnBy(2); err != nil {
			return err
		}
		return e.turnBy(2)
	}
	return &play.Refusal{Text: "not a turn"}
}

// turnBy is one turn of the party (turnCmd).
func (e *Engine) turnBy(d int) error {
	return e.exploreVerb(func() (bool, bool) { return e.turnCmd(d) })
}

// adjustStat moves the draft's stat (a game.StatKeys index) by by.
func (e *Engine) adjustStat(stat, by int) error {
	cr := e.creation
	if cr == nil {
		return core.ErrNotNow
	}
	if stat < 0 || stat >= len(game.StatKeys) {
		return &play.Refusal{Text: "no such stat"}
	}
	if game.StatKeys[stat] == "hp" { // ponytail: POR's sheet rolls hit points, never sets them; add with a hit point step on the sheet
		return &play.Refusal{Text: "not in this game"}
	}
	if game.StatKeys[stat] == "str00" { // typed as strength 18/NN, 18/00 for 100
		sc := e.ruleset().Sheet(cr.c).Scores
		n := int(sc[core.StrengthPercent]) + by
		if sc[core.Strength] != 18 || n < 1 || n > 100 {
			return &play.Refusal{Text: "no exceptional strength"}
		}
		return e.setStat(0, fmt.Sprintf("18/%02d", n%100))
	}
	return e.setStat(stat, strconv.Itoa(int(e.ruleset().Sheet(cr.c).Scores[abilities[stat]])+by))
}
