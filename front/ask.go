package front

// Questions: what a screen asks the player (Ask), and the refusals of a
// rule call that ask for something first (Confirm, PickMember, PickSpell,
// PickClass): the screen asks, then makes the call again with the answer.

import "errors"

// AskKind is what a question waits for.
type AskKind int

const (
	AskMenu    AskKind = iota + 1 // options (a menu bar, the scripts' HMENU, Parlay, a yes/no): ActChoose N; ActOK where Enter takes the first; no options: a message to acknowledge (ActOK)
	AskList                       // a list and its options (ListMenu, VMENU): ActChoose N with X the line (N -1: the line itself); ActClose takes a last LEAVE
	AskText                       // a typed line (Text, at most Max): ActText, then ActOK; ActClose gives up
	AskNumber                     // AskText for a number 0..65536
	AskMember                     // a party member (SelectCharacter, WHO): Party's Selected is the candidate; ActMember N takes one, ActLead N makes it the candidate, ActOK takes it
	AskPause                      // a pause (DELAY, MsgDelay, the title's waits): the presenter waits Ms, then sends ActOK (Enter cuts it short)
	AskKey                        // "press any key" (the pagers, PromptRow24): any key answers (ActOK, ActPage ...)
	AskSpell                      // a memorized spell (Spells): ActCast N, ActClose backs out
	AskConfirm                    // a yes/no question (Menu.Prompt, no Items): ActYes, or ActNo / ActClose; the presenter asks it, starting on No
	AskSquare                     // a combat map square (aiming a spell, a throw): ActAim X, Y; ActPoint follows the cursor; ActClose gives up
	AskMove                       // a step-by-step move: each step ActDir N (the keypad's direction); ActOK ends it, ActClose undoes
	AskModify                     // raise or lower a stat (List's lines): ActModify N with X +1 or -1; ActOK keeps, ActClose gives up
)

// Role is a colour of the presenter's palette by what it is for (the
// Theme's Palette); RoleNone: the presenter's own choice.
type Role int

const (
	RoleNone Role = iota
	RoleText
	RoleHeading
	RoleSelected
	RoleOff
	RoleWarn
	RoleMark
	RoleGold
	RoleHotkey
	RoleLog
	RoleBackground
)

// Ask is the question waiting for the player (View.Ask): every one is
// published the same way, whoever asks (the engine's prompts, a script
// instruction).
type Ask struct {
	Kind   AskKind
	Menu              // the prompt and the options (ActChoose N)
	List   []ListLine // AskList: the lines
	Text   string     // AskText, AskNumber: the line as typed so far
	Max    int        // AskText, AskNumber: the longest line
	Ms     int        // AskPause
	Spells []Spell    // AskSpell: the spells to pick from (ActCast ID)

	Default  int  // the option (or List line) highlighted first
	Dirs     bool // the direction keys answer too (ActDir N, the keypad's 1..9), as a one-option menu's keypad
	Reselect bool // ActLead N selects another member while the question stays up

	PromptInk, Ink, HotInk Role // the prompt's, the options' and the hotkeys' and highlight's colours (RoleNone: the presenter's)
}

// Confirm is a turn call's refusal: it needs the yes/no question asked
// (Flee?, Attack Ally?); answered Yes, the same call is made again sure.
type Confirm string

func (q Confirm) Error() string { return string(q) }

// ConfirmNo is Confirm whose No goes on: the game's question is asked and,
// answered No, the call is made again sure; Yes stays ("Do you want to go
// back and claim your treasure?").
type ConfirmNo string

func (q ConfirmNo) Error() string { return string(q) }

// AskSure asks err's question when it is a Confirm or a ConfirmNo, and
// calls again on the answer that goes on; false: err is neither.
func AskSure(h Host, err error, again func()) bool {
	var yes Confirm
	var no ConfirmNo
	switch {
	case errors.As(err, &yes):
		AskYes(h, string(yes), again)
	case errors.As(err, &no):
		h.Ask(&Question{Ask: Ask{Kind: AskConfirm, Menu: Menu{Prompt: string(no)}}, Answer: func(a Action) bool {
			switch a.Kind {
			case ActNo:
				again()
			case ActYes, ActClose:
			default:
				return false
			}
			return true
		}})
	default:
		return false
	}
	return true
}

// PickClass is NewClass's refusal: one of these classes (ID: the class)
// is to be picked; the call is made again with it.
type PickClass []Spell

func (q PickClass) Error() string { return "Pick New Class:" }

// PickMember is a rule call's refusal: a party member is to be picked
// (the prompt); the call is made again with the member.
type PickMember string

func (q PickMember) Error() string { return string(q) }

// PickSpell is a rule call's refusal: one of these spells is to be
// picked; the call is made again with its ID.
type PickSpell []Spell

func (q PickSpell) Error() string { return "pick a spell" }

// Menu is a question waiting for the player: a prompt and its options.
type Menu struct {
	Prompt string
	Items  []MenuItem // the options (ActChoose N); none: a message to acknowledge (ActOK)
}

// ListLine is a line of View.Ask (AskList).
type ListLine struct {
	Text    string
	Header  bool     // a heading
	Dim     bool     // greyed out
	Off     bool     // not chosen (headings, some greyed lines)
	Without []string // options (MenuItem.Text) left out while this line is chosen
	Spell   int      // the line's spell (its rules shown beside), 0 for none
}

// MenuItem is one option: its text and its hotkey.
type MenuItem struct {
	Text string
	Key  byte
}

// Spell is one memorised spell ActCast can cast.
type Spell struct {
	ID   int
	Name string
}

// Question is a question a screen asks: Answer takes each decision while
// it is up, true when it is answered (it goes); Cand is a member pick's
// candidate.
type Question struct {
	Ask    Ask
	Answer func(a Action) bool
	Cand   int
}
