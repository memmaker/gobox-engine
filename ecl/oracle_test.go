package ecl

import (
	"math/rand"
	"reflect"
	"testing"
)

// The cotab versions (engine/eclcmds.go, eclvm.go, eclrun.go skipCommand),
// copied with e.sc.X -> o.X, e.getVar -> o.h.GetVar and so on.

func (s *oracle) arg(k int) uint16 {
	lo, hi := s.argLo[k], s.argHi[k]
	switch s.argType[k] {
	case 0:
		return uint16(lo)
	case 1, 3, 128:
		return s.h.GetVar(uint16(hi)<<8 | uint16(lo))
	case 2, 129:
		return uint16(hi)<<8 | uint16(lo)
	}
	return 0
}

func (s *oracle) addr(k int) uint16 { return uint16(s.argHi[k])<<8 | uint16(s.argLo[k]) }

func (s *oracle) cmp(a, b uint16) {
	s.flags = [6]bool{a == b, a != b, a < b, a > b, a <= b, a >= b}
}
func (s *oracle) cmpT(a, b string) {
	s.flags = [6]bool{a == b, a != b, a < b, a > b, a <= b, a >= b}
}
func (s *oracle) scriptCall(k int) {
	s.calls = append(s.calls, s.pc)
	s.pc = s.addr(k)
}
func (s *oracle) skipCommand() {
	s.command = s.codeByte(s.pc)
	s.readArgs(int(s.skip[s.command]))
}

func (s *oracle) run() {
	s.command = s.codeByte(s.pc)
	switch c := s.command; {
	case c == 1:
		s.readArgs(1)
		s.pc = s.addr(1)
	case c == 2:
		s.readArgs(1)
		s.scriptCall(1)
	case c == 3:
		s.readArgs(2)
		if s.argType[1] >= 128 || s.argType[2] >= 128 {
			s.cmpT(s.text[1], s.text[2])
			return
		}
		s.cmp(s.arg(1), s.arg(2))
	case c >= 4 && c <= 7:
		s.readArgs(3)
		a, b := s.arg(1), s.arg(2)
		dst := s.addr(3)
		var r uint16
		switch c {
		case 4:
			r = a + b
		case 5:
			r = b - a
		case 6:
			if b != 0 {
				r = a / b
				s.h.SetVar(remAddr, a%b)
			}
		case 7:
			r = a * b
		}
		s.h.SetVar(dst, r)
	case c == 8:
		s.readArgs(2)
		n := byte(s.arg(1))
		if n < 255 {
			n++
		}
		s.h.SetVar(s.addr(2), uint16(s.h.Random(int(n))))
	case c == 9:
		s.readArgs(2)
		dst := s.addr(2)
		if s.argType[1] < 128 {
			s.h.SetVar(dst, s.arg(1))
			return
		}
		s.h.SetTextVar(dst, s.text[1])
	case c == 19:
		s.pc++
		if n := len(s.calls); n > 0 {
			s.pc = s.calls[n-1]
			s.calls = s.calls[:n-1]
			return
		}
		s.h.Exit(nil)
	case c == 20:
		s.flags = [6]bool{}
		s.readArgs(4)
		a, b, c2, d := s.arg(1), s.arg(2), s.arg(3), s.arg(4)
		if a == b && c2 == d {
			s.flags[0] = true
			return
		}
		s.flags[1] = true
	case c >= 22 && c <= 27:
		s.pc++
		if !s.flags[c-22] {
			s.skipCommand()
		}
	case c == 37 || c == 38:
		s.readArgs(2)
		n := byte(s.arg(1))
		count := byte(s.arg(2))
		s.pc--
		s.readArgs(int(count))
		for k := byte(0); k < count; k++ {
			if k == n {
				if s.command == 37 {
					s.pc = uint16(s.argHi[k+1])<<8 | uint16(s.argLo[k+1])
				} else {
					s.scriptCall(int(k + 1))
				}
				return
			}
		}
	case c == 42:
		s.readArgs(3)
		base := s.addr(1)
		i := byte(s.arg(2))
		s.h.SetVar(s.addr(3), s.h.GetVar(uint16(i)+base))
	case c == 47 || c == 48:
		s.readArgs(3)
		a, b := s.arg(1), s.arg(2)
		dst := s.addr(3)
		var r byte
		if c == 47 {
			r = byte(a & b)
		} else {
			r = byte(a | b)
		}
		s.cmp(0, uint16(r))
		s.h.SetVar(dst, uint16(r))
	case c == 53:
		s.readArgs(3)
		v := s.arg(1)
		base := s.addr(2)
		s.h.SetVar(base+s.arg(3), v)
	}
}

func panics(f func()) (p bool) {
	defer func() { p = recover() != nil }()
	f()
	return
}

func TestOpsMatchCotabOracle(t *testing.T) {
	ops := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 19, 20, 22, 23, 24, 25, 26, 27, 37, 38, 42, 47, 48, 53}
	skip := [256]byte{}
	for _, c := range []byte{1, 2, 10, 14, 17, 18, 29, 32, 45} {
		skip[c] = 1
	}
	for _, c := range []byte{3, 8, 9, 15, 16, 31, 34} {
		skip[c] = 2
	}
	for _, c := range []byte{4, 5, 6, 7, 11, 12, 33, 40, 42, 47, 48, 53, 55, 59} {
		skip[c] = 3
	}
	r := rand.New(rand.NewSource(7))
	types := []byte{0, 1, 2, 3, 128, 129}
	mk := func() *fake {
		f := newFake()
		return f
	}
	for i := 0; i < 20000; i++ {
		seed := r.Int63()
		rr := rand.New(rand.NewSource(seed))
		var code [7680]byte
		op := ops[rr.Intn(len(ops))]
		code[0] = op
		for j := 1; j < 120; j++ {
			code[j] = byte(rr.Intn(8)) // small values: addresses collide with memory
			if j%3 == 1 && rr.Intn(2) == 0 {
				code[j] = types[rr.Intn(len(types))]
			}
		}
		if rr.Intn(4) == 0 {
			code[1+rr.Intn(8)] = 128 // text operands
		}
		// next command for IF's skip: a Compare.
		fo, fv := mk(), mk()
		for a := uint16(0); a < 8*256+8; a++ {
			if rr.Intn(3) == 0 {
				v := uint16(rr.Intn(300))
				fo.mem[a], fv.mem[a] = v, v
			}
		}
		for _, a := range []uint16{0x0100, 0x0000, 0x0302, 0x0201} {
			fo.texts[a], fv.texts[a] = "ab", "ab"
		}
		fo.rnd = rr.Intn(1000)
		fv.rnd = fo.rnd
		flags := [6]bool{rr.Intn(2) == 0, rr.Intn(2) == 0, rr.Intn(2) == 0, rr.Intn(2) == 0, rr.Intn(2) == 0, rr.Intn(2) == 0}
		o := &oracle{h: fo, pc: 32768, code: code, flags: flags, skip: &skip}
		vm := newVM(fv)
		copy(vm.Code, code[:])
		vm.Args = skip
		vm.Flags = flags
		po, pv := panics(o.run), panics(func() { vm.Step(fv) })
		if po != pv {
			t.Fatalf("op %d seed %d: panic %v/%v", op, seed, po, pv)
		}
		if po {
			continue // both overran the 64 operand slots (ON GOTO with a huge count)
		}
		if o.pc != vm.PC || o.flags != vm.Flags || !reflect.DeepEqual(fo.mem, fv.mem) ||
			!reflect.DeepEqual(fo.texts, fv.texts) || fo.exits != fv.exits ||
			len(o.calls) != len(vm.Calls) || (len(o.calls) > 0 && !reflect.DeepEqual(o.calls, vm.Calls)) {
			t.Fatalf("op %d seed %d: pc %d/%d flags %v/%v exits %d/%d", op, seed, o.pc, vm.PC, o.flags, vm.Flags, fo.exits, fv.exits)
		}
	}
}
