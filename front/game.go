package front

import (
	"image"
	"log"
	"sync"
)

// Game is the rule calls the party menu and the creation sheet make; the
// game implements it over its engine and keeps all state. A call that the
// rules refuse returns the refusal as an error (nil: done).
type Game interface {
	// Party menu.
	Begin() error
	Lead(n int) error
	RemoveMember(n int) error
	DropMember(n int) error
	NewClass(n, class int) error // class < 0: the game asks which
	QuickStart()
	QuitGame() error

	// Creation sheet (CreationView nil: none up). A refusal's error is
	// shown (Refuse).
	CreateCharacter() error
	ModifyCharacter(n int) error
	Roll(minimal bool, minimums string) error
	Rename(name string) error
	SetStat(i int, typed string) error
	PointBuy(scores []int) error // the abilities bought on points, in Creation.StatNames' order (play.PointBuy)
	SetField(field, option int) error
	EndCreation(keep bool) error
	CreationView() *Creation

	// Title (Intro).
	TitlePicture(n int)
	TitleCredits(page int)
	TitleDone()
	SequenceDone(name string) // the View.Sequence animation ended or was skipped

	// Camp.
	LeaveCamp() error
	PlanSpell(n, spell, count int) error
	Scribe(n, spell int) error
	CampRest(kind int) error
	StopRest() error

	// Powers out of combat: refused with Confirm, PickMember or PickSpell.
	UsePower(n int, k ActKind, spell, target int, sure bool) error
	EffectsList(n int) []Entry
	SelectedMember() int // the selected member (-1: none)

	// Waits (Loot, DoorOpts).
	LootView(n int) error
	LootTake(n int) error
	LootDetect() error
	LootLeave(sure bool) error // refused with Confirm while treasure is left
	TryDoor(k DoorWay) error

	// Grid pickers (PickHead ..) and the combat icon editor.
	OpenPicker(kind int) bool
	PickerView(page int) *Picker
	Pick(i int) (done bool)
	ClosePicker()
	EditIcon(n int)
	IconEditorView() *IconEditor
	IconSize()
	IconColour(part, shade, v int)
	IconRemove()
	IconPNG(path string)
	EndIcon(keep bool)

	// Save slots. A refusal's error is shown (Refuse).
	SlotsView(loadOnly bool) *Slots
	Save(name string) error
	DeleteSlot(name string) error
	Load(name string, loadOnly bool) bool

	// Paced steps (View.Pace).
	Animate()
	CombatStep() error
	RestStep() error
	StoryStep() error
	WalkStep() error
	StopWalk() error
	TurnChoice(a Action, sure bool) error

	// Sheet and inventory.
	OpenSheet(n int) error
	CloseSheet() error
	ShowMember(n int) error
	OpenItems(n int, direct bool) error
	CloseItems(all bool) error
	Ready(i, x int) error
	UseItem(i, spell, target int, sure bool) error
	Drop(i int) error
	Sell(i int) error
	Buy(w int) error
	TakeAll() error
	SetPolicy(id int, p Policy) error

	// Shops, temples, training (Pending, Training).
	Browse(n int) error
	Identify(n, x int) error
	Appraise(jewelry bool) error
	KeepAppraised(keep bool) error
	Cure(n, x int) error
	ViewMember(n int) error
	Leave() error
	OpenTraining(n int) error
	CloseTraining() error
	TrainingView() *Training
	Train() error
	Learn(id int) error
	PickPower(id int) error

	// The full map and the journal.
	FullMapView(cursor image.Point, label *string) *FullMap
	Mark(x, y int, text string)
	Unmark(x, y int)
	WalkTo(x, y int) bool
	JournalPage(title string, lines []string, p, off int) (*Pages, int)
	SpeakPage(lines []string, p int)
	ClueChapters() []string
	ClueRead(i int) *Read

	View() View // the view published last

	// Running: Open runs the game to its first wait (the title, the party
	// menu); slot is the save to start from ("" none; the shell's
	// Capture.Slot). Answer hands the player's decision a to what waits on
	// it (View.Ask, the page, exploring's commands). Settle, after the
	// answers of a tick (or none), lets the game go on to its next wait
	// and publish: the View it returns carries the fight's events since
	// the last Settle (View.Events), for the presenter to play.
	Open(slot string)
	Answer(a Action) error
	Settle() View
	// Refuse is the game saying no with s. The front calls it inside a
	// Call (between frames) to show a call's refusal; a game may call it
	// itself at any time. Either way the game puts s where its refusals
	// go and shows it in its next View. Where that is is the game's data:
	// View.Note (Text, and At the time it was said) is the status line's
	// toast, which the presenter draws on the bottom row (in place of the
	// prompt) for a second plus 80 ms a letter after At, in the Warn
	// colour; por puts its refusals there. A game with a message log may
	// append s to its log instead (in combat, Battle.Log), or do both.
	Refuse(s string)
	// Resets counts every load and return to the main menu, the game's own
	// too (a script's or death's load, a quick start): when it changes, the
	// screens over the old game close. The front cannot see a load the game
	// makes itself, so the game owns the count and bumps it on each one.
	Resets() int
}

// Applying actions: every Action the presenter's Actions returns is the
// game's to apply, in order (por: Engine.Apply), after the queued rule
// calls. They are the page's own decisions (ActOK, ActClose, ActChoose,
// ActText, ActYes, ActNo, ActLead, ActMember, the command bar's and the
// exploration keys' kinds) and the screen-wide passes (ActHead,
// ActPortrait, ActFont, ActSound, ActResize, ActMainMenu, ActNote,
// ActNoteRemove, ActIcon, ActMinimap). A kind the game does not apply it
// hands to Unhandled.

// Unhandled reports an action the game did not apply: a log line by
// default; a test may set it to collect them.
// A kind is logged once per run (a resize sends one every frame of a drag).
var Unhandled = func(a Action) {
	if _, seen := unhandledSeen.LoadOrStore(a.Kind, true); !seen {
		log.Printf("front: action %d (N %d) not applied by the game", a.Kind, a.N)
	}
}

var unhandledSeen sync.Map

// Host is what a screen needs of the front end that shows it: the rule
// calls it queues run on the game between frames.
type Host interface {
	Call(f func(Game))
	Edit(f func(Game)) // Call, after which the laid-out pages are laid out again
	Ask(q *Question)   // put q up (nil: none)
	Asking() *Question // the question up (nil: none)
	Shown() Shown      // what the last frame showed
}

// Shown is what a screen needs to know of the view on screen.
type Shown struct {
	Ask, Powers, Sheet, Inventory, Battle bool // up
	Camp, MainMenu                        bool // up
	Loot                                  *Loot
	Treasure                              *Loot // the loot inventory's (Inventory.Treasure)
	Door                                  *DoorOpts
	Resting                               bool // camp's rest goes on
	Party                                 int  // the members
}
