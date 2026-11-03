package fixture

import (
	"encoding/binary"
	"fmt"
)

// SAVGAMx.DAT layout shared by cotab and sotsb (engine/files.go LoadGame).
const (
	svDisk  = 0x0000
	svArea  = 0x0001 // 0x800
	svGame  = 0x0801 // 0x800
	svECL   = 0x1401 // 0x1E00
	svPos   = 0x3201 // x, y, facing, 2 more
	svCount = 0x3208
	svNames = 0x3209 // 8 x 0x29 Pascal strings
	svSize  = 0x3351
)

// RawSave edits a cotab/sotsb SAVGAMA.DAT in place; it is a State.
type RawSave struct{ sv []byte }

// OpenRawSave checks the size and wraps sv.
func OpenRawSave(sv []byte) (*RawSave, error) {
	if len(sv) != svSize {
		return nil, fmt.Errorf("SAVGAMA.DAT: %d bytes, want %d", len(sv), svSize)
	}
	return &RawSave{sv}, nil
}

func (r *RawSave) Bytes() []byte { return r.sv }

func (r *RawSave) EnterArea(disk, block, geo int, ecl []byte, walls [3]int) {
	sv := r.sv
	area, game := sv[svArea:svArea+0x800], sv[svGame:svGame+0x800]
	sv[svDisk] = byte(disk)
	game[0x624] = byte(disk) // DS:6E12 symbol set
	binary.LittleEndian.PutUint16(area[0x1E4:], uint16(block))
	binary.LittleEndian.PutUint16(area[0x18A:], uint16(geo))
	e := sv[svECL : svECL+0x1E00]
	clear(e)
	copy(e, ecl)
	for i, id := range walls { // SaveGame: area 0x3F4+2i block, 0x3FA+2i slot (-1 none)
		blk, slot := id, i+1
		if id == 0xFF {
			blk, slot = -1, -1
		}
		binary.LittleEndian.PutUint16(area[0x3F4+2*i:], uint16(blk))
		binary.LittleEndian.PutUint16(area[0x3FA+2*i:], uint16(slot))
	}
	for v := 0x4A00; v <= 0x4A1F; v++ { // zeroed on entry (InitAreaScript)
		binary.LittleEndian.PutUint16(area[2*(v-0x4900):], 0)
	}
}

func (r *RawSave) SetPos(x, y, dir int) {
	for i, v := range []int{x, y, dir} {
		if v >= 0 {
			r.sv[svPos+i] = byte(v)
		}
	}
}

func (r *RawSave) SetVar(a int, n uint16) error {
	area, game := r.sv[svArea:svArea+0x800], r.sv[svGame:svGame+0x800]
	switch {
	case a >= 0x4900 && a <= 0x4CFF:
		binary.LittleEndian.PutUint16(area[2*(a-0x4900):], n)
	case a >= 0x6B00 && a <= 0x6EFF:
		binary.LittleEndian.PutUint16(game[2*(a-0x6B00):], n)
	default:
		return fmt.Errorf("not an area or game variable")
	}
	return nil
}

func (r *RawSave) Party() (names []string) {
	for i := 0; i < int(r.sv[svCount]); i++ {
		p := r.sv[svNames+i*0x29:]
		names = append(names, string(p[1:1+p[0]]))
	}
	return
}

// MakeStrong is cotab/sotsb's Strong callback.
func MakeStrong(b []byte) []byte {
	b[0x2D] = 40                             // hit dice: THAC0 is recomputed from it
	b[0x32], b[0x11B] = 250, 250             // max HP, HP
	b[0xA9], b[0x111], b[0x112] = 80, 80, 80 // AC -20 (60 - AC)
	b[0xA3], b[0xA5] = 6, 12                 // unarmed damage 6d12
	return b
}
