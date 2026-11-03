package front

import "image"

// Game is the rule calls the party menu and the creation sheet make; the
// game implements it over its engine and keeps all state. A call that the
// rules refuse returns the refusal as an error (nil: done).
type Game interface {
	// Party menu.
	Begin() error
	Lead(n int) error
	RemoveMember(n int) error
	DropMember(n int) error
	NewClass(n, class int) error // refused with a class pick: the error says which classes
	QuickStart()
	QuitGame() error

	// Creation sheet (CreationView nil: none up).
	CreateCharacter()
	ModifyCharacter(n int)
	Roll(minimal bool, minimums string)
	Rename(name string)
	SetStat(i int, typed string)
	SetField(field, option int)
	EndCreation(keep bool)
	CreationView() *Creation

	// Title (Intro).
	TitlePicture(n int)
	TitleCredits()
	TitleFlash(c int)
	TitleDone()

	// Camp.
	LeaveCamp() error
	PlanSpell(n, spell, count int) error
	Scribe(n, spell int) error
	CampRest(kind int) error
	StopRest() error

	// Powers out of combat: refused with Confirm, PickMember or PickSpell.
	UsePower(n int, k ActKind, spell, target int, sure bool) error
	EffectsList(n int) []ListLine
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

	// Save slots.
	SlotsView(loadOnly bool) *Slots
	Save(name string)
	DeleteSlot(name string)
	Load(name string, loadOnly bool) bool

	// Paced steps (View.Pace).
	Animate()
	CombatStep() error
	RestStep() error
	StoryStep() error
	WalkStep() error
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

	View() View      // the view published last
	Refuse(s string) // the game says no with s
	Resets() int     // counts the loads and returns to the main menu: the screens over the old game close
}

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
	Door                                  *DoorOpts
	Resting                               bool // camp's rest goes on
	Party                                 int  // the members
}
