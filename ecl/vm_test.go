package ecl

import (
	"strings"
	"testing"

	"gobox/ecl/script"
)

var testLayout = Layout{Area: Range{19200, 20223}, Aux: Range{31232, 31743}, Area2: Range{31744, 32767}, Code: Range{32768, 40447}}

// memory is a host over the fields' addresses.
type memory struct {
	l *script.Layout
	m map[int]uint16
}

func (m memory) Get(r script.Ref) uint16    { return m.m[m.l.Address(r)] }
func (m memory) Set(r script.Ref, v uint16) { m.m[m.l.Address(r)] = v }
func (m memory) Text(script.Ref) string     { return "" }
func (m memory) SetText(script.Ref, string) {}
func (m memory) Random(int) int             { return 0 }
func (m memory) Exit(vm *VM)                { vm.Stop = true }

// A command's operands, the arithmetic and the flags run over its decoded
// form and its resolved fields; PC counts commands.
func TestShared(t *testing.T) {
	l, err := script.ParseLayout(strings.NewReader("4B00 4 byte game A\n"))
	if err != nil {
		t.Fatal(err)
	}
	m := memory{l, map[int]uint16{19200: 7}}
	vm := New(m, testLayout, make([]byte, 7680))
	vm.SetVersion(script.V2)
	vm.Script = &script.Script{Code: []script.Instr{
		{Op: script.OpAdd, Args: []script.Arg{{Type: 1, N: 19200}, {N: 3}, {Type: 1, N: 19201}}},
		{Op: script.OpCompare, Args: []script.Arg{{Type: 1, N: 19201}, {N: 10}}},
		{Op: script.OpIfEq},
		{Op: script.OpGoto, Args: []script.Arg{{Type: 1, N: 5, Label: true}}},
		{Op: script.OpSave, Args: []script.Arg{{N: 1}, {Type: 1, N: 19202}}},
		{Op: script.OpSave, Args: []script.Arg{{N: 2}, {Type: 1, N: 19203}}},
	}}
	if err := script.ResolveRefs(vm.Script, l); err != nil {
		t.Fatal(err)
	}
	for range 5 {
		vm.Step(m)
	}
	if m.m[19201] != 10 || !vm.Flags[0] || m.m[19202] != 0 || m.m[19203] != 2 || vm.PC != 6 {
		t.Fatalf("mem %v flags %v pc %d", m.m, vm.Flags, vm.PC)
	}
}

// An operand naming state with no field is a script not resolved: the VM
// refuses it rather than reading some address.
func TestUnresolvedOperand(t *testing.T) {
	vm := New(memory{}, testLayout, nil)
	vm.SetVersion(script.V2)
	vm.Script = &script.Script{Code: []script.Instr{{Op: script.OpSave, Args: []script.Arg{{N: 1}, {Type: 1, N: 19202}}}}}
	defer func() {
		if recover() == nil {
			t.Fatal("no panic")
		}
	}()
	vm.Step(memory{})
}

// CALL's operand is a routine number: the VM keeps it (ArgN) for the game's
// command, though its operand type is a variable's.
func TestCallKeepsRoutine(t *testing.T) {
	l, err := script.ParseLayout(strings.NewReader("2E10 1 byte call CallRedraw\n"))
	if err != nil {
		t.Fatal(err)
	}
	m := memory{l, map[int]uint16{}}
	vm := New(m, testLayout, nil)
	vm.SetVersion(script.V2)
	vm.Script = &script.Script{Code: []script.Instr{{Op: script.OpCall, Args: []script.Arg{{Type: 1, N: 0x2E10}}}}}
	if err := script.ResolveRefs(vm.Script, l); err != nil {
		t.Fatal(err)
	}
	got := -1
	vm.Ops[vm.Raw[script.OpCall]] = func(vm *VM, h Host) { vm.ReadArgs(1); got = vm.ArgN[1] }
	vm.Step(m)
	if got != 0x2E10 {
		t.Fatalf("routine %04X", got)
	}
}
