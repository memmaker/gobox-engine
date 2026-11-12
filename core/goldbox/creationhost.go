package goldbox

import "gobox/core"

// CreationHost is what the creation needs from the engine: the roster
// files, the picture table and the adventure's quick party.
type CreationHost interface {
	Selected() int                           // the member the sheet works as
	Keep(core.CharID) error                  // a new character to the roster (SaveCharacter); a failed write is the error
	Join(core.RosterID) (core.CharID, error) // a roster character into the party
	Portrait(Pics) (head, body core.PicID)
	IconPics(Pics) (head, body core.PicID)
	SetPics(d core.Draft, p *Pics) // d's portrait and icon pictures into the numbers and pool picks
	QuickParty() []core.Template
}
