package fixture

import (
	"encoding/binary"
	"fmt"

	"gobox/ecl/script"
)

// RawLayout describes a game's SAVGAMx.DAT byte layout; the game passes it (offsets are data).
// Area and Game are 0x800-byte regions; the offsets At... are inside them.
type RawLayout struct {
	Size                        int // the whole file
	Disk                        int // the disk byte
	Area, Game                  int // regions
	ECL, ECLSize                int
	Pos                         int // x, y, facing, 2 more
	Count                       int // party size byte
	Names                       int // Pascal strings of NameSize bytes
	NameSize                    int
	GameSymSet                  int // in Game: the symbol set byte
	AreaBlock, AreaGeo          int // in Area: u16 each
	AreaWallBlock, AreaWallSlot int // in Area: 3 u16 each
	Vars                        *script.Layout // the game's script fields (layout/global.txt): SetVar's names
}

// RawSave edits a SAVGAMA.DAT in place; it is a State.
type RawSave struct {
	sv []byte
	l  RawLayout
}

// OpenRawSave checks the size and wraps sv.
func OpenRawSave(sv []byte, l RawLayout) (*RawSave, error) {
	if len(sv) != l.Size {
		return nil, fmt.Errorf("SAVGAMA.DAT: %d bytes, want %d", len(sv), l.Size)
	}
	return &RawSave{sv, l}, nil
}

func (r *RawSave) Bytes() []byte { return r.sv }

func (r *RawSave) EnterArea(disk, block, geo int, ecl []byte, walls [3]int) {
	sv, l := r.sv, r.l
	area, game := sv[l.Area:l.Area+0x800], sv[l.Game:l.Game+0x800]
	sv[l.Disk] = byte(disk)
	game[l.GameSymSet] = byte(disk)
	binary.LittleEndian.PutUint16(area[l.AreaBlock:], uint16(block))
	binary.LittleEndian.PutUint16(area[l.AreaGeo:], uint16(geo))
	e := sv[l.ECL : l.ECL+l.ECLSize]
	clear(e)
	copy(e, ecl)
	for i, id := range walls { // SaveGame: block and slot per wall set (-1 none)
		blk, slot := id, i+1
		if id == 0xFF {
			blk, slot = -1, -1
		}
		binary.LittleEndian.PutUint16(area[l.AreaWallBlock+2*i:], uint16(blk))
		binary.LittleEndian.PutUint16(area[l.AreaWallSlot+2*i:], uint16(slot))
	}
	for v := 0x4A00; v <= 0x4A1F; v++ { // zeroed on entry (InitAreaScript)
		binary.LittleEndian.PutUint16(area[2*(v-0x4900):], 0)
	}
}

func (r *RawSave) SetPos(x, y, dir int) {
	for i, v := range []int{x, y, dir} {
		if v >= 0 {
			r.sv[r.l.Pos+i] = byte(v)
		}
	}
}

// SetVar sets the word of field name, which the original kept at its
// layout address in the area (0x4900..) or game (0x6B00..) record.
func (r *RawSave) SetVar(name string, n uint16) error {
	ref, err := r.l.Vars.Named(name)
	if err != nil {
		return err
	}
	a := r.l.Vars.Address(ref)
	area, game := r.sv[r.l.Area:r.l.Area+0x800], r.sv[r.l.Game:r.l.Game+0x800]
	switch {
	case a >= 0x4900 && a <= 0x4CFF:
		binary.LittleEndian.PutUint16(area[2*(a-0x4900):], n)
	case a >= 0x6B00 && a <= 0x6EFF:
		binary.LittleEndian.PutUint16(game[2*(a-0x6B00):], n)
	default:
		return fmt.Errorf("%s (%04X) is not an area or game variable", name, a)
	}
	return nil
}

func (r *RawSave) Party() (names []string) {
	for i := 0; i < int(r.sv[r.l.Count]); i++ {
		p := r.sv[r.l.Names+i*r.l.NameSize:]
		names = append(names, string(p[1:1+p[0]]))
	}
	return
}

// MakeStrong is the Strong callback for the Gold Box character record.
func MakeStrong(b []byte) []byte {
	b[0x2D] = 40                             // hit dice: THAC0 is recomputed from it
	b[0x32], b[0x11B] = 250, 250             // max HP, HP
	b[0xA9], b[0x111], b[0x112] = 80, 80, 80 // AC -20 (60 - AC)
	b[0xA3], b[0xA5] = 6, 12                 // unarmed damage 6d12
	return b
}
