package core

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
	// Narrated is a line of the rules' text; More: it goes on the line before
	// (the original's message on the same row: "... Hitting for 3 points of damage").
	Narrated struct {
		Text string
		More bool
	}
	Sounded struct{ Cue SoundCue }
	Walked  struct {
		Unit CharID
		Path []BoardPos
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
	// Shot is something flying from Start to To: a missile (Item) or a
	// spell or power (Spell); From is who stands at Start (NoChar: a bolt
	// going on from a square).
	Shot struct {
		From      CharID
		Start, To BoardPos
		Item      ItemID
		Spell     SpellID
	}
	Casting struct {
		From  CharID
		Spell SpellID
		Ats   []BoardPos
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
		At   BoardPos
	}
	// CloudOn and CloudOff are a lasting spell on a square (play Clouds).
	CloudOn struct {
		Spell SpellID
		At    BoardPos
	}
	CloudOff struct {
		Spell SpellID
		At    BoardPos
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
