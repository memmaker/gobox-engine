package core

import "strings"

// Additions of sweep Spresent (fold E-R6): what por's presentation, temple and
// appraisal code ask of the rules.

// The name words of a valued gem and jewel (por's item word 2).
const (
	GemNameWord   = 0x65
	JewelNameWord = 0xD6
)

type (
	// Pay takes Gold from the purse; refused with the rules' "Not enough money." when short.
	Pay struct{ Gold uint64 }
	// Valuate takes a gem (or a jewel) from the purse, rolls its value and holds
	// the valued item (ovr22:19B2); refused when the purse has none.
	Valuate struct{ Jewel bool }
	// KeepValued keeps the valued item for Who when Keep and there is room,
	// else sells it; nothing is held after.
	KeepValued struct {
		Who  CharID
		Keep bool
	}
)

// ValuedItem asks the valued item held: Yes it is a jewel, Value its
// ItemHandle (NoItem none), High its value in gold pieces.
type ValuedItem struct{}

func (Pay) action()        {}
func (Valuate) action()    {}
func (KeepValued) action() {}
func (ValuedItem) check()  {}

// Pending is the ruleset's effects queued outside Resolve; taking them empties
// the queue (interim, until every rule call returns an Outcome).
type Pending interface{ TakePending() []Effect }

// LogJoin is a battle log line's next part after s: a space between, msg trimmed.
func LogJoin(s, msg string) string {
	msg = strings.TrimSpace(msg)
	if s != "" && msg != "" {
		return s + " " + msg
	}
	return s + msg
}

type (
	// SpellNumber asks the spell's number in the original's list (Value; Yes: it has one).
	SpellNumber struct{ Spell SpellID }
	// ArcDir asks the first direction (Value, 0..7) whose arc from Start holds To on the board.
	ArcDir struct{ Start, To BoardPos }
	// UnitBody asks about a fallen combat Unit: Yes its body is shown on the
	// ground, Value that ground's tile, Low and High the unit's square.
	UnitBody struct{ Unit CharID }
)

func (SpellNumber) check() {}
func (ArcDir) check()      {}
func (UnitBody) check()    {}
