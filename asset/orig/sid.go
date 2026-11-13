package orig

import "encoding/binary"

// The Innovation SSI-2001 (MOS 6581 SID) half of SOUND.COM, device 4 (pod-go
// docs/MUSIC.md section 9, tools/music/podmusic.py SidDriver): registers relative to
// the card's port base, voice v at 7*v.

var sidFreq = [...]int{274, 291, 308, 326, 346, 366, 388, 411, 435, 461, 489, 518} // DS:019C, a 1 MHz C scale

type sidDev struct {
	inst  [8][8]byte
	mask  int
	voice [6]int
	ins   [6]int
	ctrl  [4]byte // per voice, then the filter routing (0x19B)
}

// newSIDDev takes the 8-byte Innovation instruments, GLBR parts 16..23; nil if missing.
func newSIDDev(b []byte, cnt int) *sidDev {
	if cnt < 24 {
		return nil
	}
	d := &sidDev{voice: [6]int{0, 1, 2, 3, 4, 5}}
	off := 12 + 2*cnt
	for i := 0; i < 24; i++ {
		n := int(binary.BigEndian.Uint16(b[12+2*i:]))
		if i >= 16 {
			if off+8 > len(b) {
				return nil
			}
			copy(d.inst[i-16][:], b[off:off+8])
		}
		off += n
	}
	d.mask = int(d.inst[7][7]) // 0x17DB copies 8 bytes at a stride of 7: the last record's byte 7 lands on the voice mask
	return d
}

// init = 0x16F5 then 0x1723.
func (d *sidDev) init(w func(int, byte)) {
	d.silence(w)
	for v := 0; v < 3; v++ {
		w(7*v, 0)
		w(7*v+1, 0)
		w(7*v+2, 0)
		w(7*v+3, 8)
		d.ctrl[v] = 0x10
		w(7*v+4, 0x10)
		w(7*v+5, 0)
		w(7*v+6, 0)
	}
	w(0x15, 7)
	w(0x16, 0xBF)
	w(0x17, 0)
	w(0x18, 0x1F)
}

// setInst = 0x19F7: slot t takes the lowest free voice and instrument i (slots past 3 take 1).
func (d *sidDev) setInst(w func(int, byte), t, i int) {
	if i >= 4 {
		i = 1
	}
	d.mask &^= 1 << d.voice[t]
	v := 0
	for d.mask&(1<<v) != 0 && v <= 2 {
		v++
	}
	d.mask |= 1 << v
	d.voice[t], d.ins[t] = v, i
	if v > 2 { // no free voice: the driver's error, nothing sounds
		return
	}
	r := d.inst[i]
	d.ctrl[v] = r[4] << 4
	w(7*v+4, d.ctrl[v])
	w(7*v+5, r[0]<<4+r[1])
	w(7*v+6, r[2]<<4+r[3])
	if r[6] != 0 {
		d.ctrl[3] |= 1 << v
	} else {
		d.ctrl[3] &^= 1 << v
	}
	w(0x17, d.ctrl[3])
}

// note = 0x18C0: gate off, then for a sound (vel >= 0) the frequency, sustain and gate on.
// The driver plays a minor third above the AdLib (note+3).
func (d *sidDev) note(w func(int, byte), t, note, vel int) {
	v := d.voice[t]
	if v > 2 {
		return
	}
	r := d.inst[d.ins[t]]
	note += 3
	w(7*v+4, d.ctrl[v])
	if vel < 0 {
		return
	}
	oc := min(max(note/12-1, 1), 8)
	f := sidFreq[note%12] << oc & 0xFFFF
	w(7*v, byte(f))
	w(7*v+1, byte(f>>8))
	vel = min(max(vel-(oc-4)*int(r[5]), 0), 0x7F)
	w(7*v+6, byte((int(r[2])*vel>>7+1)<<4)+r[3])
	w(7*v+4, d.ctrl[v]|1)
}

// silence = 0x16F5: every voice's gate off.
func (d *sidDev) silence(w func(int, byte)) {
	for v := 0; v < 3; v++ {
		w(7*v+4, d.ctrl[v])
	}
}
