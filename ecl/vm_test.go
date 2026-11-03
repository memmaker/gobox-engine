package ecl

import (
	"math/rand"
	"os"
	"reflect"
	"testing"

	"gobox/dax"
)

var layout = Layout{Area: Range{19200, 20223}, Aux: Range{31232, 31743}, Area2: Range{31744, 32767}, Code: Range{32768, 40447}}

const remAddr = 31744 + 831

// fake is a flat-memory host; text variables live in texts.
type fake struct {
	mem   map[uint16]uint16
	texts map[uint16]string
	rnd   int
	exits int
}

func newFake() *fake { return &fake{mem: map[uint16]uint16{}, texts: map[uint16]string{}} }

func (f *fake) GetVar(a uint16) uint16        { return f.mem[a] }
func (f *fake) SetVar(a, v uint16)            { f.mem[a] = v }
func (f *fake) TextVar(a uint16) string       { return f.texts[a] }
func (f *fake) SetTextVar(a uint16, s string) { f.texts[a] = s }
func (f *fake) Random(n int) int              { return f.rnd % n }
func (f *fake) Exit(vm *VM) {
	f.exits++
	if vm != nil {
		vm.Stop = true
	}
}

func newVM(h Host, b ...byte) *VM {
	code := make([]byte, 7680)
	copy(code, b)
	vm := New(h, layout, code, remAddr)
	vm.PC = layout.Code.First
	return vm
}

// oracle: cotab's readArgs/unpackText/arg, copied (engine/eclvm.go) onto a bare struct.
type oracle struct {
	code    [7680]byte
	pc      uint16
	argType [MaxArgs]byte
	argLo   [MaxArgs]byte
	argHi   [MaxArgs]byte
	text    [MaxArgs]string
	h       Host
	flags   [6]bool
	calls   []uint16
	command byte
	skip    *[256]byte
	stopped bool
}

func (s *oracle) codeByte(a uint16) byte { return s.code[uint16(a-32768)%7680] }

func (s *oracle) readArgs(n int) {
	texts := 0
	for k := 1; k <= n; k++ {
		s.argType[k] = s.codeByte(s.pc + 1)
		s.argLo[k] = s.codeByte(s.pc + 2)
		s.pc += 2
		switch s.argType[k] {
		case 1, 2, 3:
			s.pc++
			s.argHi[k] = s.codeByte(s.pc)
		case 128:
			texts++
			if s.argLo[k] > 0 {
				s.text[texts] = s.unpackText(int(s.argLo[k]))
			} else {
				s.text[texts] = ""
			}
		case 129:
			texts++
			s.pc++
			s.argHi[k] = s.codeByte(s.pc)
			s.text[texts] = s.h.TextVar(uint16(s.argHi[k])<<8 | uint16(s.argLo[k]))
		}
	}
	s.pc++
}

func (s *oracle) unpackText(n int) string {
	var out []byte
	var cur, prev, c byte
	phase := 1
	letter := func(c byte) byte {
		if c > 31 {
			return c
		}
		return c + 64
	}
	for read := 0; read < n; {
		if phase < 4 {
			s.pc++
			prev = cur
			cur = s.codeByte(s.pc)
			read++
		}
		switch phase {
		case 1:
			c = cur >> 2
		case 2:
			c = (prev<<6 + cur>>2) >> 2
		case 3:
			c = (prev<<4 + cur>>4) >> 2
		case 4:
			c = cur & 63
		}
		if phase < 4 {
			phase++
		} else {
			phase = 1
		}
		if c != 0 {
			out = append(out, letter(c))
		}
	}
	if phase == 4 {
		if c = cur & 63; c != 0 {
			out = append(out, letter(c))
		}
	}
	return string(out)
}

func TestDecodeMatchesGameOracle(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 2000; i++ {
		f := newFake()
		f.texts[0x1234] = "HELLO"
		o := &oracle{pc: 32768, h: f}
		for j := range o.code {
			o.code[j] = byte(r.Intn(256))
			if r.Intn(3) == 0 {
				o.code[j] = []byte{0, 1, 2, 3, 128, 129}[r.Intn(6)]
			}
		}
		o.code[0] = 8
		vm := newVM(f)
		copy(vm.Code, o.code[:])
		vm.PC = 32768
		n := 1 + r.Intn(5)
		o.readArgs(n)
		vm.ReadArgs(n)
		if o.pc != vm.PC || o.argType != vm.ArgType || o.argLo != vm.ArgLo || o.argHi != vm.ArgHi || o.text != vm.Text {
			t.Fatalf("iter %d: decode differs: pc %d vs %d", i, o.pc, vm.PC)
		}
	}
}

func TestUnpackText(t *testing.T) {
	// "AB" packed 6 bits each: 000001 000010 -> 00000100 0010(0000)
	vm := newVM(newFake(), 8, ArgText, 2, 0x04, 0x20, 0)
	vm.ReadArgs(1)
	if vm.Text[1] != "AB" || vm.PC != 32768+5 {
		t.Fatalf("got %q pc %d", vm.Text[1], vm.PC)
	}
}

// run executes the program and returns the VM; one command per Step until n steps.
func run(h Host, steps int, b ...byte) *VM {
	vm := newVM(h, b...)
	for i := 0; i < steps; i++ {
		vm.Step(h)
	}
	return vm
}

func TestArithmetic(t *testing.T) {
	f := newFake()
	// ADD 7,5 -> 100 ; SUB 7,5 -> 101 (b-a = -2) ; DIV 17,5 -> 102 ; MUL 3,4 -> 103
	vm := run(f, 4,
		CmdAdd, 0, 7, 0, 5, 2, 100, 0,
		CmdSubtract, 0, 7, 0, 5, 2, 101, 0,
		CmdDivide, 0, 17, 0, 5, 2, 102, 0,
		CmdMultiply, 0, 3, 0, 4, 2, 103, 0)
	if f.mem[100] != 12 || f.mem[101] != 0xFFFE || f.mem[102] != 3 || f.mem[remAddr] != 2 || f.mem[103] != 12 {
		t.Fatalf("%v", f.mem)
	}
	if vm.PC != 32768+32 {
		t.Fatalf("pc %d", vm.PC)
	}
	f = newFake()
	run(f, 1, CmdDivide, 0, 1, 0, 0, 2, 102, 0)
	if v, ok := f.mem[102]; !ok || v != 0 {
		t.Fatal("divide by zero should store 0")
	}
}

func TestFlowAndIf(t *testing.T) {
	f := newFake()
	vm := newVM(f,
		CmdCompare, 0, 4, 0, 4, // flags: equal
		CmdIfEqual+1, // IF not-equal: false -> skips next
		CmdGoto, 2, 0xFF, 0xFF,
		CmdGosub, 2, 0x10, 0x80, // 0x8010 = 32784
	)
	vm.Args[CmdGoto] = 1
	vm.Step(f)
	vm.Step(f)
	if vm.PC != 32768+5+1+4 || !vm.Flags[0] || vm.Flags[1] {
		t.Fatalf("pc %d flags %v", vm.PC, vm.Flags)
	}
	vm.Step(f) // GOSUB
	if vm.PC != 32784 || len(vm.Calls) != 1 || vm.Calls[0] != 32768+14 {
		t.Fatalf("pc %d calls %v", vm.PC, vm.Calls)
	}
	vm.Code[16] = CmdReturn
	vm.Step(f)
	if vm.PC != 32768+14 || len(vm.Calls) != 0 {
		t.Fatalf("return: pc %d", vm.PC)
	}
	vm.PC = 32784
	vm.Step(f) // RETURN with an empty stack: the host's EXIT
	if f.exits != 1 || !vm.Stop {
		t.Fatal("exit not called")
	}
}

func TestTextCompareSaveTable(t *testing.T) {
	f := newFake()
	f.texts[0x1234] = "HELLO"
	run(f, 1, CmdCompare, ArgTextVar, 0x34, 0x12, ArgTextVar, 0x34, 0x12)
	vm := run(f, 1, CmdCompare, ArgTextVar, 0x34, 0x12, ArgText, 1, 0x04+0x10)
	if vm.Flags[0] {
		t.Fatal("texts differ")
	}
	run(f, 1, CmdSave, ArgTextVar, 0x34, 0x12, ArgAddress, 0x00, 0x30)
	if f.texts[0x3000] != "HELLO" {
		t.Fatal("text save")
	}
	f.mem[0x2010] = 99
	run(f, 1, CmdGetTable, ArgAddress, 0x00, 0x20, 0, 0x10, ArgAddress, 0x00, 0x40)
	if f.mem[0x4000] != 99 {
		t.Fatal("gettable")
	}
	run(f, 1, CmdSaveTable, 0, 77, ArgAddress, 0x00, 0x20, 0, 0x11)
	if f.mem[0x2011] != 77 {
		t.Fatal("savetable")
	}
}

func TestOnGotoCompareAndAndOr(t *testing.T) {
	f := newFake()
	// ON GOTO index 1 of 2 targets
	vm := run(f, 1, CmdOnGoto, 0, 1, 0, 2, ArgAddress, 0x10, 0x80, ArgAddress, 0x20, 0x80)
	if vm.PC != 0x8020 {
		t.Fatalf("pc %x", vm.PC)
	}
	vm = run(f, 1, CmdCompareAnd, 0, 1, 0, 1, 0, 2, 0, 2)
	if !vm.Flags[0] || vm.Flags[1] {
		t.Fatal("compareand")
	}
	run(f, 1, CmdAnd, 0, 0x0F, 0, 0x3C, 2, 50, 0)
	if f.mem[50] != 0x0C {
		t.Fatal("and")
	}
	f.rnd = 1000
	run(f, 1, CmdRandom, 0, 9, 2, 51, 0)
	if f.mem[51] != 1000%10 {
		t.Fatal("random")
	}
}

// Real scripts: the five entry points decode inside the code range and the
// start event walks (with sotsb's operand-count table) to an EXIT.
func TestRealScripts(t *testing.T) {
	const dir = "/Users/felix/Projects/sotsb-re/gog"
	if _, err := os.Stat(dir); err != nil {
		t.Skip("no data")
	}
	l, err := dax.NewLoader(dir)
	if err != nil {
		t.Skip(err)
	}
	args := map[int][]byte{1: {1, 2, 10, 14, 17, 18, 29, 32, 45, 50, 52, 54, 56, 57, 60, 63, 64},
		2: {3, 8, 9, 15, 16, 31, 34, 65}, 3: {4, 5, 6, 7, 11, 12, 33, 40, 42, 47, 48, 53, 55, 59},
		4: {20, 35}, 5: {46}, 6: {30, 44}, 8: {39}, 14: {41}}
	walked := 0
	for _, name := range []string{"ECL1", "ECL2", "ECL3"} {
		for id := 1; id < 60; id++ {
			b := l.Block(name, id)
			if len(b) < 20 {
				continue
			}
			f := newFake()
			vm := newVM(f)
			copy(vm.Code, b[2:])
			for n, cmds := range args {
				for _, c := range cmds {
					vm.Args[c] = byte(n)
				}
			}
			vm.PC = layout.Code.First
			var ev [5]uint16
			for k := range ev {
				vm.ReadArgs(1)
				ev[k] = vm.Addr(1)
			}
			if vm.PC != layout.Code.First+20 {
				t.Fatalf("%s %d: pc %d", name, id, vm.PC)
			}
			vm.PC = ev[4]
			ok := false
			for i := 0; i < 500; i++ {
				c := vm.CodeByte(vm.PC)
				if c == CmdExit {
					ok = true
					break
				}
				if c > 65 {
					break
				}
				vm.Skip(f)
			}
			if !ok {
				t.Errorf("%s %d: walk from %d lost the thread", name, id, ev[4])
			}
			walked++
		}
	}
	t.Logf("walked %d blocks", walked)
}

func TestDefaultTableShape(t *testing.T) {
	d := Default()
	if reflect.ValueOf(d[CmdExit]).IsValid() && d[CmdExit] != nil {
		t.Fatal("EXIT must stay the game's")
	}
}
