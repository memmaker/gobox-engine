package importer

// BRItem is an item as both originals keep it (X62), in memory and in .STF
// files: a string[41] name, the list pointer (42), then the 16 bytes an
// ITEM0.DAX template gives (46..61): 46 type (an ITEMS row), 47..49 name
// parts, 50 the bonus, 52 flags (128 readied, 64 not loot, 32 counted in
// its name), 54 weight, 56 count or shots, 57 value, 60/61 the weapon
// special.
type BRItem [62]byte

// The BRItem byte offsets.
const (
	BRItType, BRItBonus, BRItFlags, BRItWeight, BRItCount = 46, 50, 52, 54, 56
)

func (it *BRItem) Type() int     { return int(it[BRItType]) }
func (it *BRItem) Readied() bool { return it[BRItFlags]&128 != 0 }

// Name is the name it carries (bytes 0..41).
func (it *BRItem) Name() string { return string(it[1 : 1+min(int(it[0]), 41)]) }

// BRRecord is a character as both originals keep it (CHRDATAn.SAV, MON0CHA):
// brCharacterSize bytes.
type BRRecord [brCharacterSize]byte

// BROffLevel is the level byte of a BRRecord.
const BROffLevel = 41

// Name is the character's name (string[15]).
func (r *BRRecord) Name() string { return string(r[1 : 1+int(min(r[0], 15))]) }
