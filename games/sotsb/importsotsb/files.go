package importsotsb

// itemRecord is the 63-byte DOS item record, in its file layout
// (encoding/binary).
type itemRecord struct {
	PName   [41]byte // 0 shown name, Pascal string (built by ItemName)
	Pad41   byte     // 41
	_       [4]byte  // 42 next pointer
	Type    byte     // 46 base item type (index into the item type records)
	Words   [3]byte  // 47 name word ids into itemWords
	Plus    byte     // 50 signed
	Plus2   byte     // 51
	Readied byte     // 52
	Hidden  byte     // 53 bit 3-i: word i not known yet
	Cursed  byte     // 54
	Weight  uint16   // 55
	Count   byte     // 57 a bag: the items in it
	Price   uint16   // 58
	Spells  [3]byte  // 60 a scroll's spells, +128 = being scribed
}

// affectRecord is the 9-byte DOS affect node in its file layout.
type affectRecord struct {
	Type byte   // affect number, index into the ovr13 handler table
	Time uint16 // remaining duration, 0 = permanent
	Data byte
	Call byte    // non-zero: the handler runs when the affect is removed
	_    [4]byte // next pointer
}
