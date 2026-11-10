package ecl

import "gobox/ecl/script"

// Command numbers (the first byte of each command), the same in both games.
const (
	CmdExit       = 0
	CmdGoto       = 1
	CmdGosub      = 2
	CmdCompare    = 3
	CmdAdd        = 4
	CmdSubtract   = 5
	CmdDivide     = 6
	CmdMultiply   = 7
	CmdRandom     = 8
	CmdSave       = 9
	CmdReturn     = 19
	CmdCompareAnd = 20
	CmdIfEqual    = 22 // .. 27: skip the next command unless the flag is set
	CmdIfLast     = 27
	CmdOnGoto     = 37
	CmdOnGosub    = 38
	CmdGetTable   = 42
	CmdAnd        = 47
	CmdOr         = 48
	CmdSaveTable  = 53
)

// Default is the opcode set that cotab and sotsb run identically. The game
// overrides or adds entries (vm.Ops[n] = ...). Not included: everything that
// touches game state (party, monsters, pictures, menus, files ...), EXIT and
// PRINT (game-side effects) and LOAD CHARACTER.
func Default() (t [256]Op) {
	t[CmdGoto] = opGoto
	t[CmdGosub] = opGosub
	t[CmdCompare] = opCompare
	for k := CmdAdd; k <= CmdMultiply; k++ {
		t[k] = opArithmetic
	}
	t[CmdRandom] = opRandom
	t[CmdSave] = opSave
	t[CmdReturn] = opReturn
	t[CmdCompareAnd] = opCompareAnd
	for k := CmdIfEqual; k <= CmdIfLast; k++ {
		t[k] = opIf
	}
	t[CmdOnGoto] = opOnGotoGosub
	t[CmdOnGosub] = opOnGotoGosub
	t[CmdGetTable] = opGetTable
	t[CmdAnd] = opAndOr
	t[CmdOr] = opAndOr
	t[CmdSaveTable] = opSaveTable
	return
}

func opGoto(vm *VM, h Host) {
	vm.ReadArgs(1)
	vm.PC = vm.Label(1)
}

func opGosub(vm *VM, h Host) {
	vm.ReadArgs(1)
	vm.Call(1)
}

// opCompare compares texts when either operand is text.
func opCompare(vm *VM, h Host) {
	vm.ReadArgs(2)
	if vm.ArgType[1] >= ArgText || vm.ArgType[2] >= ArgText {
		vm.SetTextFlags(vm.Text[1], vm.Text[2])
		return
	}
	vm.SetFlags(vm.Arg(1, h), vm.Arg(2, h))
}

// opArithmetic: add, subtract (b-a), divide (remainder to vm.Remainder; by
// zero stores 0), multiply.
func opArithmetic(vm *VM, h Host) {
	vm.ReadArgs(3)
	a, b := vm.Arg(1, h), vm.Arg(2, h)
	dst := vm.Ref(3)
	var r int64
	switch vm.Command {
	case CmdAdd:
		r = a + b
	case CmdSubtract:
		r = b - a
	case CmdDivide:
		if b != 0 {
			r = a / b
			if vm.Remainder != (script.Ref{}) {
				h.Set(vm.Remainder, a%b)
			}
		}
	case CmdMultiply:
		r = a * b
	}
	h.Set(dst, r)
}

// opRandom: 0 .. operand.
func opRandom(vm *VM, h Host) {
	vm.ReadArgs(2)
	n := byte(vm.Arg(1, h))
	if n < 255 {
		n++
	}
	h.Set(vm.Ref(2), int64(h.Random(int(n))))
}

// opSave stores a value or text.
func opSave(vm *VM, h Host) {
	vm.ReadArgs(2)
	dst := vm.Ref(2)
	if vm.ArgType[1] < ArgText {
		h.Set(dst, vm.Arg(1, h))
		return
	}
	h.SetText(dst, vm.Text[1])
}

// opReturn: back from a GOSUB, or the host's EXIT when there is none.
func opReturn(vm *VM, h Host) {
	vm.PC++
	if n := len(vm.Calls); n > 0 {
		vm.PC = vm.Calls[n-1]
		vm.Calls = vm.Calls[:n-1]
		return
	}
	h.Exit(vm)
}

// opCompareAnd: equal when both pairs are.
func opCompareAnd(vm *VM, h Host) {
	vm.Flags = [6]bool{}
	vm.ReadArgs(4)
	a, b, c, d := vm.Arg(1, h), vm.Arg(2, h), vm.Arg(3, h), vm.Arg(4, h)
	if a == b && c == d {
		vm.Flags[0] = true
		return
	}
	vm.Flags[1] = true
}

// opIf skips the next command unless the flag holds.
func opIf(vm *VM, h Host) {
	vm.PC++
	if !vm.Flags[vm.Command-CmdIfEqual] {
		vm.Skip(h)
	}
}

func opOnGotoGosub(vm *VM, h Host) {
	vm.ReadArgs(2)
	n := byte(vm.Arg(1, h))
	count := byte(vm.Arg(2, h))
	vm.PC--
	vm.ReadArgs(int(count))
	for k := byte(0); k < count; k++ {
		if k == n {
			if vm.Command == CmdOnGoto {
				vm.PC = vm.Label(int(k + 1))
			} else {
				vm.Call(int(k + 1))
			}
			return
		}
	}
}

func opAndOr(vm *VM, h Host) {
	vm.ReadArgs(3)
	a, b := vm.Arg(1, h), vm.Arg(2, h)
	dst := vm.Ref(3)
	var r byte
	if vm.Command == CmdAnd {
		r = byte(a & b)
	} else {
		r = byte(a | b)
	}
	vm.SetFlags(0, int64(r))
	h.Set(dst, int64(r))
}

// opGetTable: dst := [base + index].
func opGetTable(vm *VM, h Host) {
	vm.ReadArgs(3)
	i := byte(vm.Arg(2, h))
	h.Set(vm.Ref(3), h.Get(vm.Ref(1).Plus(int(i))))
}

// opSaveTable: [base + index] := value.
func opSaveTable(vm *VM, h Host) {
	vm.ReadArgs(3)
	v := vm.Arg(1, h)
	h.Set(vm.Ref(2).Plus(int(vm.Arg(3, h))), v)
}
