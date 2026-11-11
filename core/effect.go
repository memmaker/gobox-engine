package core

// Effect is one fact of an Outcome, in the order it happened (RNG order is
// the rules'). The engine applies the board ones and shows them all. The set
// is the rule-made part of play.Event.
type Effect interface{ effect() }

// Outcome is what an action did. Confirm, when set, means nothing happened:
// the rules ask this yes/no first and take Confirmed{action} on yes (Flee?,
// Attack Ally?).
type Outcome struct {
	Effects  []Effect
	Confirm  string
	Created  CharID // KeepDraft: the character made, else NoChar
	TurnOver bool   // in a fight: the actor's turn is used (por turnChoice's done)
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
		Pace Pace
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
	// TerrainSet is a square's new terrain number (Board.Terrain): a blast's
	// rubble, a cloud's marker laid or cleared, a corpse (BR cbtacts.pas).
	TerrainSet struct {
		At      BoardPos
		Terrain TerrainID
	}
	Gained struct {
		Who  CharID
		What string
	}
	Lost struct {
		Who  CharID
		What string
	}
	// Took is a loot item gone from the pile to Who.
	Took struct {
		Who  CharID
		Item CarriedID
	}
	// PartyHealed: nobody is wounded any more (a rest's healing, por healTick).
	PartyHealed struct{}
	// RestInterrupted: a monster broke the rest off (por stepRest's roll).
	RestInterrupted struct{}
	// Hurt: a script's damage (ScriptDamage) hit Unit for Damage, already
	// taken; Dies: it killed outright (the "dies." message).
	Hurt struct {
		Unit   CharID
		Damage int32
		Dies   bool
	}
	// Refused is an action the rules turned down, and why.
	Refused struct {
		Who CharID
		Why Refusal
	}
)

// Pace is how fast a Walked unit goes: a walk, or fast (BR's leap for cover).
type Pace int8

const (
	PaceWalk Pace = iota
	PaceFast
)

// Refusal is why the rules turned an action down.
type Refusal int8

const (
	OverLoaded Refusal = iota + 1 // no room for another item
)

// SoundCue is a sound a rule makes, named as the adventure's asset names it.
type SoundCue string

// Facing is 0 N, 2 E, 4 S, 6 W, with the diagonals between (game.Where.Facing).
type Facing int8

func (Narrated) effect()        {}
func (Hurt) effect()            {}
func (Sounded) effect()         {}
func (Walked) effect()          {}
func (Turned) effect()          {}
func (Attacked) effect()        {}
func (Shot) effect()            {}
func (Casting) effect()         {}
func (Struck) effect()          {}
func (Healed) effect()          {}
func (PartyHealed) effect()     {}
func (RestInterrupted) effect() {}
func (Fell) effect()            {}
func (Fled) effect()            {}
func (Placed) effect()          {}
func (CloudOn) effect()         {}
func (CloudOff) effect()        {}
func (TerrainSet) effect()      {}
func (Gained) effect()          {}
func (Lost) effect()            {}
func (Took) effect()            {}
func (Refused) effect()         {}
