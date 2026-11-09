package ecl

import (
	"testing"

	"gobox/ecl/script"
)

var testLayout = Layout{Area: Range{19200, 20223}, Aux: Range{31232, 31743}, Area2: Range{31744, 32767}, Code: Range{32768, 40447}}

type memory map[uint16]uint16

func (m memory) GetVar(a uint16) uint16    { return m[a] }
func (m memory) SetVar(a, v uint16)        { m[a] = v }
func (m memory) TextVar(uint16) string     { return "" }
func (m memory) SetTextVar(uint16, string) {}
func (m memory) Random(int) int            { return 0 }
func (m memory) Exit(vm *VM)               { vm.Stop = true }

func newTestVM(m memory, code []byte) *VM {
	vm := New(m, testLayout, make([]byte, 7680), 31744+831)
	vm.SetVersion(script.V2)
	vm.Script = script.Assemble(script.V2, code)
	return vm
}

// A command's operands, the arithmetic and the flags run over its decoded
// form; PC counts commands.
func TestShared(t *testing.T) {
	m := memory{19200: 7}
	vm := newTestVM(m, nil)
	vm.Script = &script.Script{Code: []script.Instr{
		{Op: script.OpAdd, Args: []script.Arg{{Type: 1, N: 19200}, {N: 3}, {Type: 1, N: 19201}}},
		{Op: script.OpCompare, Args: []script.Arg{{Type: 1, N: 19201}, {N: 10}}},
		{Op: script.OpIfEq},
		{Op: script.OpGoto, Args: []script.Arg{{Type: 1, N: 5, Label: true}}},
		{Op: script.OpSave, Args: []script.Arg{{N: 1}, {Type: 1, N: 19202}}},
		{Op: script.OpSave, Args: []script.Arg{{N: 2}, {Type: 1, N: 19203}}},
	}}
	for range 4 {
		vm.Step(m)
	}
	vm.Step(m)
	if m[19201] != 10 || !vm.Flags[0] || m[19202] != 0 || m[19203] != 2 || vm.PC != 6 {
		t.Fatalf("mem %v flags %v pc %d", m, vm.Flags, vm.PC)
	}
}
