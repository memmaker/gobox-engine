package front

// Waits: the options of what a suspended script waits on that a screen
// puts up (the treasure after a fight, a locked door).

// Loot is the treasure screen's options besides View and Exit (View.Loot).
type Loot struct {
	Take   bool // there is treasure to take (LootTake)
	Detect bool // the selected member has Detect Magic memorized (LootDetect)
	Share  bool // there is money to divide among the party (LootShare)
}

// DoorOpts is a locked door's options besides Exit (View.Door).
type DoorOpts struct{ Bash, Pick, Knock bool }

// DoorWay is what TryDoor does with a locked door.
type DoorWay int

const (
	DoorExit  DoorWay = iota // leave it
	DoorBash                 // force it
	DoorPick                 // pick the lock
	DoorKnock                // cast Knock
)
