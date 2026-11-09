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

// Host is what the interpreter asks of the game.
type Host interface {
	GetVar(a uint16) uint16
	SetVar(a, v uint16)
	TextVar(a uint16) string // text kept at a (ArgTextVar operands, via readArgs)
	SetTextVar(a uint16, s string)
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
	ArgLo   [MaxArgs]byte
	ArgHi   [MaxArgs]byte
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
	// RemainderAddr: the address DIVIDE leaves its remainder in.
	RemainderAddr uint16
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

// fetch starts the command at PC: Command, and its operands from the first.
func (vm *VM) fetch() {
	vm.cur, vm.argPos = int(vm.PC), 0
	vm.Command = 0
	if vm.cur < len(vm.Script.Code) {
		vm.Command = vm.Raw[vm.Script.Code[vm.cur].Op]
	}
}

// readScriptArgs is ReadArgs over the decoded command: the next n operands
// as the byte code had them; PC goes to the next command.
func (vm *VM) readScriptArgs(n int) {
	var args []script.Arg
	if vm.cur < len(vm.Script.Code) {
		args = vm.Script.Code[vm.cur].Args
	}
	texts := 0
	for k := 1; k <= n; k++ {
		var a script.Arg // a missing operand is a 0 byte
		if vm.argPos < len(args) {
			a = args[vm.argPos]
		}
		vm.argPos++
		vm.ArgType[k], vm.ArgLo[k] = a.Type, byte(a.N)
		switch a.Type {
		case ArgVarByte, ArgAddress, ArgVarWord:
			vm.ArgHi[k] = byte(a.N >> 8)
		case ArgText:
			texts++
			vm.ArgLo[k] = byte(len(a.S))
			vm.Text[texts] = a.S
		case ArgTextVar:
			texts++
			vm.ArgHi[k] = byte(a.N >> 8)
			vm.Text[texts] = vm.Host.TextVar(vm.Addr(k))
		}
	}
	vm.PC = uint16(vm.cur + 1)
	if vm.Trace != nil {
		vm.Trace(vm, n)
	}
}

// New returns a VM (reading text variables through h) over code with the shared opcodes installed.
func New(h Host, l Layout, code []byte, remainder uint16) *VM {
	vm := &VM{Host: h, Layout: l, Code: code, RemainderAddr: remainder}
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

// VarAddr joins an operand's two bytes into an address.
func VarAddr(lo, hi byte) uint16 { return uint16(hi)<<8 | uint16(lo) }

// Addr is the address held by operand k.
func (vm *VM) Addr(k int) uint16 { return VarAddr(vm.ArgLo[k], vm.ArgHi[k]) }

// Step runs the command at PC. Returns false when vm.Stop is set afterwards.
func (vm *VM) Step(h Host) bool {
	vm.LastCommand = vm.Command
	if vm.Script != nil {
		vm.fetch()
	} else {
		vm.Command = vm.CodeByte(vm.PC)
	}
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

// ReadArgs decodes the n operands following PC and steps PC past them.
func (vm *VM) ReadArgs(n int) {
	if vm.Script != nil {
		vm.readScriptArgs(n)
		return
	}
	texts := 0
	for k := 1; k <= n; k++ {
		vm.ArgType[k] = vm.CodeByte(vm.PC + 1)
		vm.ArgLo[k] = vm.CodeByte(vm.PC + 2)
		vm.PC += 2
		switch vm.ArgType[k] {
		case ArgVarByte, ArgAddress, ArgVarWord:
			vm.PC++
			vm.ArgHi[k] = vm.CodeByte(vm.PC)
		case ArgText:
			texts++
			if vm.ArgLo[k] > 0 {
				vm.Text[texts] = vm.unpackText(int(vm.ArgLo[k]))
			} else {
				vm.Text[texts] = ""
			}
		case ArgTextVar:
			texts++
			vm.PC++
			vm.ArgHi[k] = vm.CodeByte(vm.PC)
			vm.Text[texts] = vm.Host.TextVar(vm.Addr(k))
		}
	}
	vm.PC++
	if vm.Trace != nil {
		vm.Trace(vm, n)
	}
}

// Arg is the value of operand k.
func (vm *VM) Arg(k int, h Host) uint16 {
	lo, hi := vm.ArgLo[k], vm.ArgHi[k]
	switch vm.ArgType[k] {
	case ArgImmediate:
		return uint16(lo)
	case ArgVarByte, ArgVarWord, ArgText:
		return h.GetVar(VarAddr(lo, hi))
	case ArgAddress, ArgTextVar:
		return VarAddr(lo, hi)
	}
	return 0
}

// ArgsTrace is a ready Trace printing operands the way the games' scriptTrace does.
func ArgsTrace(h Host, print func(string)) func(*VM, int) {
	return func(vm *VM, n int) {
		for k := 1; k <= n; k++ {
			print(fmt.Sprintf("    arg%d type %d %d/%d = %d\n", k, vm.ArgType[k], vm.ArgLo[k], vm.ArgHi[k], vm.Arg(k, h)))
		}
	}
}

// Skip steps over the command at PC using the Args table.
func (vm *VM) Skip(h Host) {
	if vm.Script != nil {
		vm.fetch()
	} else {
		vm.Command = vm.CodeByte(vm.PC)
	}
	vm.ReadArgs(int(vm.Args[vm.Command]))
}

// Call pushes PC and jumps to the address in operand k.
func (vm *VM) Call(k int) {
	vm.Calls = append(vm.Calls, vm.PC)
	vm.PC = vm.Addr(k)
}

// SetFlags sets the six comparison flags.
func (vm *VM) SetFlags(a, b uint16) {
	vm.Flags = [6]bool{a == b, a != b, a < b, a > b, a <= b, a >= b}
}

// SetTextFlags is SetFlags for strings.
func (vm *VM) SetTextFlags(a, b string) {
	vm.Flags = [6]bool{a == b, a != b, a < b, a > b, a <= b, a >= b}
}

func letterOf(c byte) byte {
	if c > 31 {
		return c
	}
	return c + 64
}

// unpackText reads n bytes after PC, 6 bits per character (PC ends on the last).
func (vm *VM) unpackText(n int) string {
	var out []byte
	var cur, prev, c byte
	phase := 1
	for read := 0; read < n; {
		if phase < 4 {
			vm.PC++
			prev = cur
			cur = vm.CodeByte(vm.PC)
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
			out = append(out, letterOf(c))
		}
	}
	if phase == 4 {
		if c = cur & 63; c != 0 {
			out = append(out, letterOf(c))
		}
	}
	return string(out)
}
