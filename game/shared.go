package game

import "gobox/rules"

// The questions, turns and results the rules and the classic front share:
// what carries an Action lives here, beside Action.

// Option is a rule verb the player may pick, with its original hotkey.
type Option struct {
	Label  string
	Key    rune
	Do     Action
	Chords string // more keys that pick it (alt+q,space; Data.VerbChords)
	Func   string // the function's gobox name when Label is the series' (Data.Texts: "Drop Character" shown as "Remove Character")
}

// Ask is a domain question. The answer is an Answer decision. Options on an
// Ask of another kind are extra moves the player may make instead of an
// answer (a move question's Done): gobox applies the Option's Do as is.
type Ask struct {
	Kind    AskKind
	Prompt  string
	Text    string   // the question's body text
	Options []Option // Choice: labels with the original hotkeys
	Default int      // default option or number
	Yes     bool     `json:",omitempty"` // YesNo: the highlight starts on Yes (false: on No)
	NoEsc   bool     `json:",omitempty"` // Esc does not answer it: the question stays up
	NoPick  bool     `json:",omitempty"` // no highlight, Enter does not answer it (only the hotkeys; YesNo: Y and N)
	Max     int      // Name length, Number upper bound
	Lines   []Line   // PickItem: the list's rows
	Typed   string   // Name, Number: what is typed so far
	Ms      int      // Continue: a timed pause of Ms milliseconds (0: wait for a key)
	Break   int      `json:",omitempty"` // Choice: >0 puts Options[Break:] on a second bar row (the original's '@'); the bar takes rows 23-24 and Up/Down switch rows
	View    bool     // while set, the player may open the View menu on the highlighted option's member (ViewMember{N}) and the ask stays up; used by member picks
}

// Refusal is an Err that carries the reason and, if any, the question to put instead.
type Refusal struct {
	Text string
	Ask  *Ask
}

// Turn is whose turn it is and what they may do.
type Turn struct {
	Unit      int
	Moves     int
	Casting   int    // the spell the unit is casting over rounds, 0 none
	Aim       bool   // aiming: Aim takes a square (Cast: Ats), Cancel gives up
	Spell     string // aiming a spell, its name
	Options   []Option
	Weapon    *Weapon        // the weapon in hand (its icon on Attack), nil none
	Missile   *Weapon        // the ranged weapon readied (its icon on Ranged), nil none
	Maneuvers []TurnManeuver // the rule set's maneuvers (Maneuver picks one); none: no Maneuver
}

// Combat is a fight. Squares hold art refs, not cells.
type Combat struct {
	Size    rules.Size
	Squares []Ground
	Units   []Unit
	Turn    *Turn
	Round   int
	Dungeon bool
	Reach   []Reach    // squares the turn's unit may walk to
	Spells  []SpellRef // castable by the turn's unit
	Items   []Item     // the turn's member's items it may use now (Allowed UseItem)
	Aims    []SpellRef // spells that take a square, with their Range, Targets and Area
	Targets []int      // unit IDs the turn's Attack takes; while aiming, the foes Tab cycles through
	Panel   int        // the unit ID the status row shows (who acts, or is being hit), 0 none
	// Marks are the squares framed for a choice: with AimArea, a spell's area
	// (AreaAt the squares it covers round the aimed square, Skipped the allies
	// in it that it passes over), else the units an aim would hit.
	Marks, AreaAt, Skipped []rules.BoardPos
	AimArea                int    // aiming an area: its radius+1 round the aimed square (0: none)
	Why                    string // aiming: why the aimed square cannot be taken, "" when it can

	// The ground as a view draws it: Squares' Tile numbers index Tiles, the
	// Data.Art names of the tile set (read live; "" a tile with no picture).
	Tiles     []string
	FloorTile string   `json:",omitempty"` // the Tiles name (a series combat sprite) of the plain floor, tiled on every floor square; "": bare ground
	Effects   []Effect // sprites over the map
	Strike    Strike   // the last weapon attack
	Log       []string // the battle's messages, oldest first
}

// Report is something that happened.
type Report struct {
	Kind         EventKind
	Combat       *Combat  // EvEngaged: the fight's board before its first turn
	Party        []Member // EvEngaged: the party as the fight began (its figures; a fight lost leaves none after)
	Text         string
	Role         string // EvSaid: "narrator", an NPC name, ...
	N            int
	Wait         bool
	Clear        bool
	Turn         bool    // EvMove: a turn in place (N the facing), no step (To unset)
	Picture      Picture // EvDown: the body the unit leaves on its square (a ground Overlay), for the fight's last blow, when no Combat follows
	Sound        string  // sound ref for EvSound / song ref for EvSong
	Unit, Target int
	At, To       rules.BoardPos
	Sprite       []Picture
	StepMs       int        // EvMissile with Sprite: the flight's time per third of a square (0: Ms shared out)
	Ms           int        // >0: the moment holds this many milliseconds (with Screen: a timed moment outside a fight)
	Pose         int        // >0 with Ms in a fight: Unit's icon pose (1 action) while this moment shows
	Hit          bool       // EvAttack: the blow hit
	Screen       *Screen    // with Ms: the typed screen as it stood for this moment (a blink, a blast frame), shown for Ms milliseconds
	Door         *DoorTried // EvDoor
}

// Result is what one Apply or Step did. Err refuses the decision; nothing changed.
type Result struct {
	Events []Report
	Err    error
}

func (r *Refusal) Error() string { return r.Text }

// The actions only the front answers: a game may name them as an Option's Do
// or a Question's Do; Do never receives them.
type (
	Cancel    struct{}                    // Esc: the question or screen up is given up
	Aim       struct{ At rules.BoardPos } // the free cursor, or the square picked
	OpenItems struct{ Member int }        // the inventory screen (-1: the selected or acting member)
	SpeedMenu struct{}                    // the game speed menu
)

func (Cancel) action()    {}
func (Aim) action()       {}
func (OpenItems) action() {}
func (SpeedMenu) action() {}

// CommandEntries are cs's party menu entries, for Data.Commands.
func CommandEntries(cs []Option) []PartyCommand {
	out := make([]PartyCommand, len(cs))
	for i, c := range cs {
		out[i] = PartyCommand{Label: c.Label, Key: c.Key}
		if a, ok := c.Do.(AddCharacter); ok {
			out[i].From = a.From
		}
	}
	return out
}
