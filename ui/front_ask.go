package ui

// Questions: what a screen asks the player (Ask), and the refusals of a
// rule call that ask for something first (Confirm, PickMember, PickSpell):
// the screen asks, then makes the call again with the answer.

// AskKind is what a question waits for.
type AskKind int

const (
	AskMenu    AskKind = iota + 1 // options (a menu bar, the scripts' HMENU, Parlay, a yes/no): ActChoose N; ActOK where Enter takes the first; no options: a message to acknowledge (ActOK)
	AskList                       // entries and the commands acting on them (ListMenu, VMENU): ActChoose N with X the entry (N -1: the line itself); ActClose takes a last LEAVE
	AskText                       // a typed line (Text, at most Max): Enter sends it whole as one ActText; ActClose gives up
	AskNumber                     // AskText for a number 0..65536
	AskMember                     // a party member (SelectCharacter, WHO): Party's Selected is the candidate; ActMember N takes one, ActLead N makes it the candidate, ActOK takes it
	AskPause                      // a pause (DELAY, MsgDelay, the title's waits): the presenter waits Ms, then sends ActOK (Enter cuts it short)
	AskKey                        // "press any key" (the pagers, PromptRow24): any key answers (ActOK, ActPage ...)
	AskSpell                      // a memorized spell (Entries, each with its Spell): ActCast N, ActClose backs out
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

// FrontAsk is the question waiting for the player (the FrontAsk): every one is
// published the same way, whoever asks (the engine's prompts, a script
// instruction).
type FrontAsk struct {
	Kind    AskKind
	Menu                 // the prompt and the options (ActChoose N)
	Entries []FrontEntry // AskList, AskSpell: the entries the commands act on
	Text    string       // AskText, AskNumber: the line as typed so far
	NoEsc   bool         // Esc does not answer it: the question stays up (the game's say)
	NoPick  bool         // no highlight: Enter and the arrows do nothing, only the hotkeys (Y and N) answer
	Max     int          // AskText, AskNumber: the longest line
	Ms      int          // AskPause

	Default  int  // the option (or entry) highlighted first
	Break    int  // >0: Commands[Break:] go on the bar's second row (play.Ask.Break)
	Yes      bool // AskConfirm: the highlight starts on Yes (false: on No)
	Reselect bool // ActLead N selects another member while the question stays up
	View     bool // V opens the View menu on the member selected (ActView -1) while the question stays up

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
	Prompt   string
	Commands []Command // the options (ActChoose N); none: a message to acknowledge (ActOK)
}

// FrontEntry is a line of the Ask (AskList, AskSpell): what its commands act on.
type FrontEntry struct {
	Text    string
	Header  bool     // a heading
	Dim     bool     // greyed out
	Off     bool     // not chosen (headings, some greyed lines)
	Without []string // commands (Command.Text) left out while this entry is chosen
	Spell   int      // the line's spell (its rules shown beside), 0 for none
}

// Command is one option of a menu (acting on the chosen entry, if any): its text and its hotkey.
type Command struct {
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
	Ask    FrontAsk
	Answer func(a FrontAction) bool
	Cand   int
}
