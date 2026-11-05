// Package play is the typed game contract (draft, 2026-11-05). It replaces
// game.Rules once every game has moved over; see MIGRATION.md.
//
// The game is purely reactive. It never does anything by itself: gobox reads
// State, switches over its type to pick the screen and controls, turns the
// player's input into one of the Actions that state publishes, and hands it
// to Do. Do refuses an action it does not allow, or applies it and returns
// the Outcome, which gobox plays (animates, voices, paces) before it reads
// State again. Questions, menus, waits, paging and screens are all gobox's.
package play

import (
	"image"
	"time"

	"gobox/game"
)

// Game is what every game implements.
type Game interface {
	State() State               // what is true now; a copy
	Do(Action) (Outcome, error) // the only way anything changes; error: refused, nothing changed
}

// State is one situation the party is in. The set is closed: gobox handles
// every type, and a new situation means a new type here first.
type State interface{ state() }

// Action is one thing the player does. Each State lists the ones it allows
// now; Do still checks.
type Action interface{ action() }

// Outcome is what an action did, in order: gobox plays the events, then
// shows the new State.
type Outcome struct {
	Events []game.Event
}

// Choice is an allowed action as gobox offers it: the rules' words and the
// original's hotkey. gobox builds its menus and dialogs from these.
type Choice struct {
	Label string
	Key   rune
	Do    Action
}

// Common is what every state shows besides its own view.
type Common struct {
	Party    []game.Member
	Selected int
	Purse    map[string]int
	Where    game.Where
	Area     string      // the area's name as the player knows it
	Lore     []game.Lore // the journal, newest first
	Choices  []Choice    // the actions allowed now
}

// Exploring: the party walks a map (dungeon, town or wilderness).
type Exploring struct {
	Common
	Map ExploreMap
}

// ExploreMap is what gobox needs to draw and walk the 3D view and the
// automap. Whether a step goes through is the game's call (Do(Step)).
type ExploreMap struct {
	Size    image.Point
	Walls   *game.Walls
	Automap game.Automap
}

// Fighting: a battle is on.
type Fighting struct {
	Common
	Combat game.Combat // the field, the units and the turn's options
}

// Talking: a script shows text and pictures and waits for the player. The
// script's place is part of the game's state; Continue or a reply moves it.
type Talking struct {
	Common
	Speaker game.Picture
	Text    []game.TextRun
	Picture []game.Picture
}

// Reading: a handout's text, which gobox pages through. Continue leaves.
type Reading struct {
	Common
	Title string
	Lines []string // journal.Layout's
}

// Shopping: a shop, temple or trainer is open.
type Shopping struct {
	Common
	Shop game.Shop
}

// Looting: treasure lies before the party.
type Looting struct {
	Common
	Loot game.Loot
}

// Camping: the party is in camp (rest, memorize, order, ...).
type Camping struct{ Common }

// Creating: a character is being made.
type Creating struct {
	Common
	Sheet game.SheetPage
}

// Over: the game has ended (won or lost).
type Over struct {
	Common
	Won bool
}

func (Exploring) state() {}
func (Fighting) state()  {}
func (Talking) state()   {}
func (Reading) state()   {}
func (Shopping) state()  {}
func (Looting) state()   {}
func (Camping) state()   {}
func (Creating) state()  {}
func (Over) state()      {}

// The shared actions. Game-specific verbs stay game.Decision values until
// they move here (Legacy wraps them meanwhile).
type (
	Step          struct{ Facing int }      // walk one square that way; refused at a wall
	Face          struct{ Facing int }      // turn
	Continue      struct{}                  // go on past text or a handout
	Reply         struct{ N int }           // pick reply N of a conversation
	AdvanceTimeBy struct{ D time.Duration } // rest or wait; the outcome tells what happened meanwhile
	EndTurn       struct{ Unit int }        // a member's combat turn is over; the outcome carries the monsters' moves up to the next member's turn
	Legacy        struct{ D game.Decision } // an old decision, until it gets a type here
)

func (Step) action()          {}
func (Face) action()          {}
func (Continue) action()      {}
func (Reply) action()         {}
func (AdvanceTimeBy) action() {}
func (EndTurn) action()       {}
func (Legacy) action()        {}
