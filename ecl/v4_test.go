package ecl

import (
	"strings"
	"testing"
)

// cells is a Host over the fields' addresses, 64 bits wide.
type cells struct {
	l *Layout
	m map[int]int64
}

func (c cells) Get(r Ref) int64     { return c.m[c.l.Address(r)] }
func (c cells) Set(r Ref, v int64)  { c.m[c.l.Address(r)] = v }
func (c cells) Text(Ref) string     { return "" }
func (c cells) SetText(Ref, string) {}
func (c cells) Random(int) int      { return 0 }
func (c cells) Exit(vm *VM)         { vm.PC = 0 } // a Pool host's Exit: the script continues

const v4Layout = "4B00 1 word game A\n4B01 1 word game B\n4B02 1 word game C\n4B03 1 word game LoopCounter\n4B04 1 word game LoopLimit\n"

// v4VM assembles code under dialect v and sets up a VM over the layout.
func v4VM(t *testing.T, v *Version, code []byte) (*VM, cells) {
	t.Helper()
	l, err := ParseLayout(strings.NewReader(v4Layout))
	if err != nil {
		t.Fatal(err)
	}
	s, err := Assemble(v, l, code)
	if err != nil {
		t.Fatal(err)
	}
	h := cells{l, map[int]int64{}}
	vm := New(h, testLayout, nil)
	vm.SetVersion(v)
	vm.Script = s
	return vm, h
}

var (
	cmdSave    = []byte{CmdSave}
	varA       = []byte{ArgVar, 0x00, 0x4B}
	varB       = []byte{ArgVar, 0x01, 0x4B}
	varC       = []byte{ArgVar, 0x02, 0x4B}
	varLongA   = []byte{ArgVarLong, 0x00, 0x4B}
	loopCount  = Ref{Field: 4}
	loopLimit  = Ref{Field: 5}
	cmdExitV4  = []byte{CmdExit}
	cmdFor     = []byte{CmdFor}
	cmdNext    = []byte{CmdNext}
	cmdCompare = []byte{CmdCompare}
)

func cat(parts ...[]byte) (b []byte) {
	for _, p := range parts {
		b = append(b, p...)
	}
	return
}

func long(v int32) []byte {
	u := uint32(v)
	return []byte{ArgLong, byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}
}

func byteArg(n byte) []byte { return []byte{ArgByte, n} }

func run(vm *VM, h Host, n int) {
	for range n {
		vm.Step(h)
	}
}

// Long immediates (signed) and long variables are read; no Pool dialect has a
// command whose operand is one, so nothing of theirs changes.
func TestLongOperands(t *testing.T) {
	code := cat(
		cmdSave, long(100000), varA, // A := 100000
		cmdSave, long(-5), varB, // B := -5
		cmdSave, varLongA, varC, // C := A (long variable)
		cmdCompare, varLongA, long(100000), // equal
	)
	vm, h := v4VM(t, V4, code)
	run(vm, h, 4)
	if h.m[0x4B00] != 100000 || h.m[0x4B01] != -5 || h.m[0x4B02] != 100000 || !vm.Flags[0] {
		t.Fatalf("mem %v flags %v", h.m, vm.Flags)
	}
}

// FOR 1 TO 3 with BODY adding 1 to a cell runs the body three times.
func TestForNext(t *testing.T) {
	add := cat([]byte{CmdAdd}, varA, byteArg(1), varA)
	code := cat(cmdFor, byteArg(1), byteArg(3), add, cmdNext, cmdExitV4)
	vm, h := v4VM(t, V4, code)
	vm.LoopCounter, vm.LoopLimit = loopCount, loopLimit
	run(vm, h, 1+3*2)
	if h.m[0x4B00] != 3 || h.m[0x4B03] != 4 || h.m[0x4B04] != 3 || vm.PC != 3 {
		t.Fatalf("mem %v pc %d", h.m, vm.PC)
	}
}

func TestForNextWithoutRefs(t *testing.T) {
	for _, code := range [][]byte{cat(cmdFor, byteArg(1), byteArg(3)), cmdNext} {
		vm, h := v4VM(t, V4, code)
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("no error without loop fields")
				}
			}()
			vm.Step(h)
		}()
	}
}

// Skip steps over each V4 command by BR's skip() counts, not by Fixed.
func TestV4SkipDiffers(t *testing.T) {
	for _, c := range []struct {
		cmd byte
		n   int
	}{{12, 3}, {31, 2}, {52, 1}, {54, 1}, {21, 0}, {37, 0}, {38, 0}, {39, 0}, {43, 0}, {49, 0}} {
		code := cat([]byte{c.cmd}, make([]byte, 2*c.n))
		vm, h := v4VM(t, V4, cat(code, cmdExitV4))
		vm.Skip(h)
		if vm.PC != 1 {
			t.Errorf("command %d: PC %d after Skip", c.cmd, vm.PC)
		}
		if int(V4.Skip[c.cmd]) != c.n {
			t.Errorf("command %d: Skip %d, want %d", c.cmd, V4.Skip[c.cmd], c.n)
		}
	}
	// 59, 60 and 68 are one op (OpNothing): skipped by none, as BR's executor runs them.
	if V4.Skip[59] != 0 || V4.Skip[60] != 0 || V4.Skip[68] != 0 {
		t.Errorf("Skip 59: %d, 60: %d, 68: %d", V4.Skip[59], V4.Skip[60], V4.Skip[68])
	}
}

// RETURN with no caller ends the script under V4, continues under V1.
func TestReturnEmptyStack(t *testing.T) {
	for _, c := range []struct {
		v    *Version
		stop bool
	}{{V4, true}, {V1, false}} {
		vm, h := v4VM(t, c.v, []byte{CmdReturn})
		vm.Step(h)
		if vm.Stop != c.stop {
			t.Errorf("V%d: Stop %v", c.v.ID, vm.Stop)
		}
	}
}

// V4's CALL table names BR's routines.
func TestV4Calls(t *testing.T) {
	want := map[uint16]CallFn{11723: CallRedrawView, 49182: CallStepForward, 49176: CallWallAhead, 45571: CallSound, 26627: CallPictureAnimation}
	for k, f := range want {
		if V4.Calls[k] != f {
			t.Errorf("call %d: %v", k, V4.Calls[k])
		}
	}
}

func TestRefAsAndBytes(t *testing.T) {
	vm, _ := v4VM(t, V4, append(append(append([]byte{}, cmdSave...), varA...), varLongA...))
	vm.Fetch()
	vm.ReadArgs(2)
	for k, w := range []Width{WidthByte, WidthDword} {
		if got := vm.Ref(k + 1).As; got != w {
			t.Errorf("operand %d As %d, want %d", k+1, got, w)
		}
	}
	if vm.Ref(1).Cell() != (Ref{Field: 1}) {
		t.Errorf("Cell keeps As: %+v", vm.Ref(1).Cell())
	}
	l, err := ParseLayout(strings.NewReader("4C00 32 bytes local Scratch\n"))
	if err != nil {
		t.Fatal(err)
	}
	if r, err := l.Resolve(0x4C05); err != nil || r != (Ref{Field: 1, Index: 5}) {
		t.Errorf("resolve %+v %v", r, err)
	}
	if _, err := ParseLayout(strings.NewReader("4C00 4 bytes const T\n")); err == nil {
		t.Error("const bytes accepted")
	}
}
