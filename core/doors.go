package core

// What a try at a door (BashDoor, PickLock at AtDoor) did.
type (
	// DoorGaveWay: the door opened to Who's try.
	DoorGaveWay struct{ Who CharID }
	// DoorTriesSpent: this way of trying is used up at this door (por
	// bashDoor: a member too weak for a magically locked door).
	DoorTriesSpent struct{}
)

func (DoorGaveWay) effect()    {}
func (DoorTriesSpent) effect() {}
