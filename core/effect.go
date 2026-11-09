package core

import "image"

// Effect is one fact of an Outcome, in the order it happened (RNG order is
// the rules'). The engine applies the board ones and shows them all. The set
// is the rule-made part of play.Event.
type Effect interface{ effect() }

// Outcome is what an action did. Confirm, when set, means nothing happened:
// the rules ask this yes/no first and take Confirmed{action} on yes (Flee?,
// Attack Ally?).
type Outcome struct {
	Effects []Effect
	Confirm string
	Created CharID // KeepDraft: the character made, else NoChar
}

type (
	Narrated struct{ Text string }
	Sounded  struct{ Cue SoundCue }
	Walked   struct {
		Unit CharID
		Path []image.Point
	}
	Turned struct {
		Unit   CharID
		Facing Facing
	}
	Attacked struct {
		From, To CharID
		Hit      bool
		Damage   int32
	}
	Shot struct {
		From CharID
		To   image.Point
	}
	Casting struct {
		From  CharID
		Spell SpellID
		Ats   []image.Point
	}
	Struck struct {
		Unit   CharID
		Damage int32
	}
	Healed struct {
		Unit CharID
		HP   int32
	}
	Fell struct{ Unit CharID }
	Fled struct{ Unit CharID }
	// Placed puts a unit on the board: a summons, a Dimension Door.
	Placed struct {
		Unit CharID
		At   image.Point
	}
	// CloudOn and CloudOff are a lasting spell on a square (play Clouds).
	CloudOn struct {
		Spell SpellID
		At    image.Point
	}
	CloudOff struct {
		Spell SpellID
		At    image.Point
	}
	Gained struct {
		Who  CharID
		What string
	}
	Lost struct {
		Who  CharID
		What string
	}
)

// SoundCue is a sound a rule makes, named as the adventure's asset names it.
type SoundCue string

// Facing is 0 N, 2 E, 4 S, 6 W, with the diagonals between (game.Where.Facing).
type Facing int8

func (Narrated) effect() {}
func (Sounded) effect()  {}
func (Walked) effect()   {}
func (Turned) effect()   {}
func (Attacked) effect() {}
func (Shot) effect()     {}
func (Casting) effect()  {}
func (Struck) effect()   {}
func (Healed) effect()   {}
func (Fell) effect()     {}
func (Fled) effect()     {}
func (Placed) effect()   {}
func (CloudOn) effect()  {}
func (CloudOff) effect() {}
func (Gained) effect()   {}
func (Lost) effect()     {}
