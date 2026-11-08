package game

import "image"

// The rule calls a game takes besides the verbs of game.go: screens a game
// opens and closes, waits it answers, the creation sheet, the pickers and the
// icon designer, the map's labels. A member is a party index; -1 means the
// selected one where the call allows it.
type (
	// Party menu and creation.
	QuickStart   struct{}
	Quit         struct{}
	NewCharacter struct{}
	Modify       struct{ Member int }
	Roll         struct {
		Minimal  bool
		Minimums string
	}
	Rename      struct{ Name string }
	SetField    struct{ Field, Option int }
	EndCreation struct{ Keep bool }
	// Camp.
	Plan     struct{ Member, Spell, Count int }
	CampRest struct{ Kind int }
	StopRest struct{}
	UsePower struct {
		Member, Kind, Spell, Target int // Kind: a front.ActKind
		Sure                        bool
	}
	// Loot and doors.
	LootLook   struct{ Member int }
	LootTake   struct{ Member int }
	LootDetect struct{}
	LootLeave  struct{ Sure bool }
	LootShare  struct{} // divide the money among the party
	LootStatus struct{} // the treasure menu's Status (Loot.Status)
	TryDoor    struct{ Way int }
	// Pickers and the icon designer.
	OpenPicker  struct{ Kind int }
	Pick        struct{ N int }
	ClosePicker struct{}
	EditIcon    struct{ Member int }
	IconSize    struct{}
	IconColour  struct{ Part, Shade, Value int }
	IconRemove  struct{}
	IconPNG     struct{ Path string }
	EndIcon     struct{ Keep bool }
	// Slots.
	DeleteSlot struct{ Name string }
	// Sheet and inventory.
	OpenSheet  struct{ Member int }
	CloseSheet struct{}
	ShowMember struct{ Member int }
	OpenItems  struct {
		Member int
		Direct bool
	}
	CloseItems struct{ All bool }
	SetPolicy  struct{ ID, Policy int }
	// Shops, temples, training.
	Browse        struct{ Member int }
	Value         struct{ Jewelry bool } // appraise the gem (or jewel) offered
	KeepAppraised struct{ Keep bool }
	ViewMember    struct{ Member int }
	OpenTraining  struct{ Member int }
	CloseTraining struct{}
	Learn         struct{ Spell int }
	PickPower     struct{ ID int }
	// The page: the window shows Cols by Rows text cells (State.Cells follows).
	Resize struct{ Cols, Rows int }
	// Map.
	Mark struct {
		At   image.Point
		Text string
	}
	Unmark   struct{ At image.Point }
	WalkTo   struct{ At image.Point }
	StopWalk struct{}
	// Speak voices a journal page: its speech clips Keys, in order, in place
	// of what is being voiced (none: silence).
	Speak struct{ Keys []string }
	// The title sequence (State.Intro): picture step N, the credits, the
	// lightning's palette colour C, and its end.
	TitlePicture struct{ N int }
	TitleCredits struct{ Page int } // Page: the credits page (Credit.Page) to show
	TitleFlash   struct{ C int }
	TitleDone    struct{}
	// SequenceDone: the State.Sequence Name ended or the player skipped it.
	SequenceDone struct{ Name string }
)

func (QuickStart) decision()    {}
func (Quit) decision()          {}
func (NewCharacter) decision()  {}
func (Modify) decision()        {}
func (Roll) decision()          {}
func (Rename) decision()        {}
func (SetField) decision()      {}
func (EndCreation) decision()   {}
func (Plan) decision()          {}
func (CampRest) decision()      {}
func (StopRest) decision()      {}
func (UsePower) decision()      {}
func (LootLook) decision()      {}
func (LootTake) decision()      {}
func (LootDetect) decision()    {}
func (LootLeave) decision()     {}
func (LootShare) decision()     {}
func (LootStatus) decision()    {}
func (TryDoor) decision()       {}
func (OpenPicker) decision()    {}
func (Pick) decision()          {}
func (ClosePicker) decision()   {}
func (EditIcon) decision()      {}
func (IconSize) decision()      {}
func (IconColour) decision()    {}
func (IconRemove) decision()    {}
func (IconPNG) decision()       {}
func (EndIcon) decision()       {}
func (DeleteSlot) decision()    {}
func (OpenSheet) decision()     {}
func (CloseSheet) decision()    {}
func (ShowMember) decision()    {}
func (OpenItems) decision()     {}
func (CloseItems) decision()    {}
func (SetPolicy) decision()     {}
func (Browse) decision()        {}
func (Value) decision()         {}
func (KeepAppraised) decision() {}
func (ViewMember) decision()    {}
func (OpenTraining) decision()  {}
func (CloseTraining) decision() {}
func (Learn) decision()         {}
func (PickPower) decision()     {}
func (Resize) decision()        {}
func (Mark) decision()          {}
func (TitlePicture) decision()  {}
func (TitleCredits) decision()  {}
func (TitleFlash) decision()    {}
func (TitleDone) decision()     {}
func (SequenceDone) decision()  {}
func (Speak) decision()         {}
func (Unmark) decision()        {}
func (WalkTo) decision()        {}
func (StopWalk) decision()      {}

// PickHead .. PickForget are the pickers OpenPicker readies.
const (
	PickHead     = iota // the portrait's head (creation sheet)
	PickBody            // its body
	PickIconHead        // the combat icon's head (icon editor)
	PickWeapon          // its weapon: the icon's body (CBODY)
	PickAdd             // the saved characters to add to the party (party menu)
	PickForget          // the saved characters to forget (party menu)
)
