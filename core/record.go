package core

// RecordLayout names a DOS character record layout by its size in bytes.
// An adventure's manifest names the layout of its saves; gobox/char holds
// each layout's one codec.
type RecordLayout uint16

// The Pool records. NoRecord: the adventure has none (or predates the field).
const (
	NoRecord  RecordLayout = 0
	Layout285 RecordLayout = 0x11D // Pool of Radiance
	Layout422 RecordLayout = 422   // Curse of the Azure Bonds
	Layout439 RecordLayout = 439   // Secret of the Silver Blades
	Layout510 RecordLayout = 510   // Pools of Darkness
)

// RecordImage is a character record as read from its file, in its layout;
// only gobox/char codes it (the rules carry it for the fields they do not model).
type RecordImage []byte
