// Package ecl is the shared core of the ECL area-script interpreter: operand
// decoding, the opcode table, and the opcodes both games run identically.
// The memory map (which addresses are area, aux, area2 or code words) and
// every game-specific opcode stay in the game, which reaches them through
// Host and by overriding Ops entries. Synchronous: no goroutines. How a game
// suspends for player input (cotab: replay from a checkpoint; sotsb: `then`
// callbacks) is the game's; the seam is VM.Step, one command per call.
package ecl

import (
	"fmt"

	"gobox/ecl/script"
)

// Layout is the game's address map, as data. Addresses outside every range
// are "special" values the Host handles.
const (
	AddrArea = iota
	AddrArea2
	AddrAux
	AddrCode
	AddrSpecial
)

// Range is an inclusive address range.
type Range struct{ First, Last uint16 }

func (r Range) has(a uint16) bool { return a >= r.First && a <= r.Last }

// Layout names the ranges. Cotab and sotsb: Area 19200..20223,
// Aux 31232..31743, Area2 31744..32767, Code 32768..40447.
type Layout struct{ Area, Area2, Aux, Code Range }

// KindOf is the address kind (AddrArea ...) of a.
func (l Layout) KindOf(a uint16) int {
	switch {
	case l.Area.has(a):
		return AddrArea
	case l.Area2.has(a):
		return AddrArea2
	case l.Aux.has(a):
		return AddrAux
	case l.Code.has(a):
		return AddrCode
	}
	return AddrSpecial
}

// Operand types (the byte before each operand).
const (
	ArgImmediate = 0
	ArgVarByte   = 1 // 1..3: a variable address (two bytes)
	ArgAddress   = 2
	ArgVarWord   = 3
	ArgText      = 128 // packed text follows
	ArgTextVar   = 129 // text read from a variable address
	MaxArgs      = 256 // operands are 1-based and an ON GOTO count is a byte (1..255)
)

// Host is what the interpreter asks of the game: the state a resolved
// operand (script.Ref, a field of the running script's Layout) names.
type Host interface {
	Get(r script.Ref) uint16
	Set(r script.Ref, v uint16)
	Text(r script.Ref) string // text kept in r (ArgTextVar operands, via ReadArgs)
	SetText(r script.Ref, s string)
	Random(n int) int // 0 .. n-1
	Exit(vm *VM)      // the game's EXIT, used when RETURN has nowhere to go
}

// Op runs one command; vm.PC is on the command byte.
type Op func(vm *VM, h Host)

// VM is the script machine: code, PC and the decoded operands of the
// current command.
type VM struct {
	Host   Host   // answers TextVar for text-variable operands
	Code   []byte // the script bytes; Code[a-Layout.Code.First] is address a
	Layout Layout
	PC     uint16

	ArgType [MaxArgs]byte
	ArgN    [MaxArgs]int        // an operand's number: an immediate, a label's command, a routine, a text's length; 0 for a variable
	ArgRef  [MaxArgs]script.Ref // a variable operand's field (the zero Ref for any other operand)
	Text    [MaxArgs]string // text operands of the current command, 1-based
	Flags   [6]bool         // equal, not equal, less, greater, less-equal, greater-equal
	Calls   []uint16        // return addresses

	Command, LastCommand byte
	Stop                 bool

	// Args[cmd] is the number of operands Skip steps over for cmd (a table
	// the game fills; commands that take none stay 0).
	Args [256]byte
	// Ops is the opcode table; nil entries do nothing. New starts with Default.
	Ops [256]Op
	// Remainder is the field DIVIDE leaves its remainder in; the zero Ref:
	// its dialect's DIVIDE keeps none.
	Remainder script.Ref
	// Trace, if set, is called after each readArgs with the decoded operands.
	Trace func(vm *VM, n int)

	// Script is the running area's decompiled commands; PC is a command
	// number in it. Code holds its tables and the area's text variables.
	Script *script.Script
	// Raw is the game's command number of a canonical op: Ops and Command
	// keep the game's own numbers.
	Raw    [256]byte
	cur    int // the command being run
	argPos int // its next operand
}

// SetVersion gives the VM its dialect's command numbers. Where several
// numbers mean one op the lowest stands for them.
func (vm *VM) SetVersion(v *script.Version) {
	for b := 255; b >= 0; b-- {
		if c := v.Cmds[b]; c.Known {
			vm.Raw[c.Op] = byte(b)
		}
	}
}

// Fetch starts the command at PC: Command, and its operands from the first.
func (vm *VM) Fetch() {
	vm.cur, vm.argPos = int(vm.PC), 0
	vm.Command = 0
	if vm.cur < len(vm.Script.Code) {
		vm.Command = vm.Raw[vm.Script.Code[vm.cur].Op]
	}
}

// Cur is the number of the command being run (PC has moved on once its
// operands are read).
func (vm *VM) Cur() int { return vm.cur }

// ReadArgs gives the command's next n operands as the byte code had them;
// PC goes to the next command.
func (vm *VM) ReadArgs(n int) {
	var args []script.Arg
	routine := false // CALL's operand is a routine number (ArgN), its Ref a routine field
	if vm.cur < len(vm.Script.Code) {
		args = vm.Script.Code[vm.cur].Args
		routine = vm.Script.Code[vm.cur].Op == script.OpCall
	}
	texts := 0
	for k := 1; k <= n; k++ {
		var a script.Arg // a missing operand is a 0 byte
		if vm.argPos < len(args) {
			a = args[vm.argPos]
		}
		vm.argPos++
		vm.ArgType[k], vm.ArgN[k], vm.ArgRef[k] = a.Type, a.N, a.Ref
		if isVar(a) && !routine {
			if a.Ref == (script.Ref{}) {
				panic(fmt.Sprintf("ecl: command %d operand %d (%04X) has no field: the script was not resolved in its layout", vm.cur, k, a.N))
			}
			vm.ArgN[k] = 0 // state is named by its field, never by an address
		}
		switch a.Type {
		case ArgText:
			texts++
			vm.ArgN[k] = len(a.S)
			vm.Text[texts] = a.S
		case ArgTextVar:
			texts++
			vm.Text[texts] = vm.Host.Text(a.Ref)
		}
	}
	vm.PC = uint16(vm.cur + 1)
	if vm.Trace != nil {
		vm.Trace(vm, n)
	}
}

// isVar: a names state, so it must carry its field (script.ResolveRefs). A
// label and a CALL's routine are numbers.
func isVar(a script.Arg) bool {
	return !a.Label && (a.Type == ArgVarByte || a.Type == ArgVarWord || a.Type == script.ArgVarLong || a.Type == ArgTextVar)
}

// New returns a VM (reading text variables through h) over code with the shared opcodes installed.
func New(h Host, l Layout, code []byte) *VM {
	vm := &VM{Host: h, Layout: l, Code: code}
	vm.Ops = Default()
	return vm
}

// CodeByte reads the script byte at address a (wrapping inside the code range).
func (vm *VM) CodeByte(a uint16) byte {
	n := int(vm.Layout.Code.Last) - int(vm.Layout.Code.First) + 1
	return vm.Code[int(uint16(a-vm.Layout.Code.First))%n]
}

// SetCodeByte writes the script byte at address a.
func (vm *VM) SetCodeByte(a uint16, v byte) {
	n := int(vm.Layout.Code.Last) - int(vm.Layout.Code.First) + 1
	vm.Code[int(uint16(a-vm.Layout.Code.First))%n] = v
}

// Ref is the field operand k names.
func (vm *VM) Ref(k int) script.Ref { return vm.ArgRef[k] }

// Label is the command operand k jumps to.
func (vm *VM) Label(k int) uint16 { return uint16(vm.ArgN[k]) }

// Step runs the command at PC. Returns false when vm.Stop is set afterwards.
func (vm *VM) Step(h Host) bool {
	vm.LastCommand = vm.Command
	vm.Fetch()
	if op := vm.Ops[vm.Command]; op != nil {
		op(vm, h)
	}
	return !vm.Stop
}

// Run steps until Stop or halt() reports true; Stop is cleared on return.
func (vm *VM) Run(h Host, halt func() bool) {
	for !vm.Stop && (halt == nil || !halt()) {
		vm.Step(h)
	}
	vm.Stop = false
}

// Arg is the value of operand k.
func (vm *VM) Arg(k int, h Host) uint16 {
	switch vm.ArgType[k] {
	case ArgImmediate:
		return uint16(byte(vm.ArgN[k]))
	case ArgVarByte, ArgVarWord:
		return h.Get(vm.ArgRef[k])
	case ArgAddress:
		return uint16(vm.ArgN[k])
	}
	// A text or text variable has no number. It cannot be asked for: in every
	// shipped script of POR, COTAB, SotSB and POD a text operand stands only
	// where the -re Pascal reads it as text (Compare, Save and Print test the
	// type for $80 first; the menus', Who's, InputString's and EncounterMenu's
	// GetEclVar operands are never texts).
	return 0
}

// ArgsTrace is a ready Trace printing operands the way the games' scriptTrace does.
func ArgsTrace(h Host, print func(string)) func(*VM, int) {
	return func(vm *VM, n int) {
		for k := 1; k <= n; k++ {
			print(fmt.Sprintf("    arg%d type %d %d %v = %d\n", k, vm.ArgType[k], vm.ArgN[k], vm.ArgRef[k], vm.Arg(k, h)))
		}
	}
}

// Skip steps over the command at PC using the Args table.
func (vm *VM) Skip(h Host) {
	vm.Fetch()
	vm.ReadArgs(int(vm.Args[vm.Command]))
}

// Call pushes PC and jumps to the address in operand k.
func (vm *VM) Call(k int) {
	vm.Calls = append(vm.Calls, vm.PC)
	vm.PC = vm.Label(k)
}

// SetFlags sets the six comparison flags.
func (vm *VM) SetFlags(a, b uint16) {
	vm.Flags = [6]bool{a == b, a != b, a < b, a > b, a <= b, a >= b}
}

// SetTextFlags is SetFlags for strings.
func (vm *VM) SetTextFlags(a, b string) {
	vm.Flags = [6]bool{a == b, a != b, a < b, a > b, a <= b, a >= b}
}
