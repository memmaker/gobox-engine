package rules

// CreationHost is what the creation needs from the engine: the roster
// files, the picture table and the adventure's quick party.
type CreationHost interface {
	Selected() int                 // the member the sheet works as
	Keep(CharID) error             // a new character to the roster (SaveCharacter); a failed write is the error
	Join(RosterID) (CharID, error) // a roster character into the party
	Portrait(Pics) (head, body PicID)
	IconPics(Pics) (head, body PicID)
	SetPics(d Draft, p *Pics) // d's portrait and icon pictures into the numbers and pool picks
	QuickParty() []Template
}
