package front

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
}
