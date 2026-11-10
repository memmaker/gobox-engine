package script

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// The operand types of the byte code.
const (
	ArgByte    byte = 0
	ArgVar     byte = 1
	ArgWord    byte = 2
	ArgVarWord byte = 3
	ArgLong    byte = 4    // a 32-bit immediate
	ArgVarLong byte = 5    // a 32-bit variable
	ArgText    byte = 0x80 // packed text in the code
	ArgTextVar byte = 0x81 // text kept in variables
)

// Line is one decoded command of Walk.
type Line struct {
	Byte    byte  // the byte-code command
	Cmd     Cmd   // what it is
	Args    []Arg // the operands as encoded
	Targets []int // the addresses it jumps to
	Next    int   // the address after it
}

// Known reports whether the dialect has the command.
func (l Line) Known() bool { return l.Cmd.Known }

// reader decodes operands of a code image.
type reader struct {
	code []byte
	base int
	pc   int
}

func (r *reader) at(a int) byte {
	if i := a - r.base; i >= 0 && i < len(r.code) {
		return r.code[i]
	}
	return 0
}

// operands reads n operands after pc; pc ends on the last byte read.
func (r *reader) operands(n int) []Arg {
	var args []Arg
	for range n {
		t, lo := r.at(r.pc+1), r.at(r.pc+2)
		r.pc += 2
		a := Arg{Type: t, N: int(lo)}
		switch t {
		case ArgVar, ArgWord, ArgVarWord, ArgVarLong, ArgTextVar:
			r.pc++
			a.N |= int(r.at(r.pc)) << 8
		case ArgLong:
			a.N |= int(r.at(r.pc+1))<<8 | int(r.at(r.pc+2))<<16 | int(int8(r.at(r.pc+3)))<<24
			r.pc += 3
		case ArgText:
			a.N = int(lo)
			if lo > 0 {
				a.S = r.packed(int(lo))
			}
		}
		args = append(args, a)
	}
	return args
}

func unpackChar(c byte) byte {
	if c > 31 {
		return c
	}
	return c + 64
}

// packed reads n 6-bit-packed bytes after pc (ovr08:1153).
func (r *reader) packed(n int) string {
	var out []byte
	count, state := 0, 1
	var prev, cur byte
	for count < n {
		if state < 4 {
			r.pc++
			prev, cur = cur, r.at(r.pc)
			count++
		}
		var c byte
		switch state {
		case 1:
			c = cur >> 2
		case 2:
			c = (prev&3)<<4 | cur>>4
		case 3:
			c = (prev&0xF)<<2 | cur>>6
		case 4:
			c = cur & 0x3F
		}
		if state < 4 {
			state++
		} else {
			state = 1
		}
		if c != 0 && len(out) < 255 {
			out = append(out, unpackChar(c))
		}
	}
	if state == 4 {
		if c := cur & 0x3F; c != 0 && len(out) < 255 {
			out = append(out, unpackChar(c))
		}
	}
	return string(out)
}

// decode decodes the command at pc.
func (v *Version) decode(r *reader, pc int) Line {
	r.pc = pc
	b := r.at(pc)
	c := v.Cmds[b]
	l := Line{Byte: b, Cmd: c}
	l.Args = r.operands(c.Fixed)
	if c.Count > 0 {
		n := 0
		if a := l.Args[c.Count-1]; a.Type == ArgByte {
			n = a.N
		}
		first := len(l.Args)
		l.Args = append(l.Args, r.operands(n)...)
		if c.Op == OpOnGoto || c.Op == OpOnGosub {
			for _, a := range l.Args[first:] {
				l.Targets = append(l.Targets, a.N)
			}
		}
	}
	if c.Op == OpGoto || c.Op == OpGosub {
		l.Targets = append(l.Targets, l.Args[0].N)
	}
	l.Next = r.pc + 1
	return l
}

// endsFlow: nothing runs on after the command.
func endsFlow(c Cmd) bool {
	return !c.Known || c.Op == OpGoto || c.Op == OpExit || c.Op == OpReturn || c.Op == OpNewECL || c.Op == OpClearExit
}

// Walk decodes the code reachable from the five event entry points: the
// entries, the commands by address and the jump targets. Flow never enters
// the tables (Tables), which a plain walk would decode as code where a
// command runs on into one.
func (v *Version) Walk(code []byte) ([]int, map[int]Line, map[int]bool) {
	tables := map[int]bool{}
	for {
		events, lines, jumps := v.walkFrom(code, tables)
		more := v.tables(lines, v.Base+len(code))
		if len(more) == len(tables) {
			return events, lines, jumps
		}
		tables = more
	}
}

// tables are the table addresses inside the code image: the first operand
// of GetTable and Protection.
func (v *Version) tables(lines map[int]Line, end int) map[int]bool {
	t := map[int]bool{}
	for _, l := range lines {
		if l.Cmd.Op != OpGetTable && l.Cmd.Op != OpProtection || len(l.Args) == 0 || l.Args[0].Type != ArgVar {
			continue
		}
		if a := l.Args[0].N; a >= v.Base && a < end {
			t[a] = true
		}
	}
	return t
}

func (v *Version) walkFrom(code []byte, tables map[int]bool) ([]int, map[int]Line, map[int]bool) {
	r := &reader{code: code, base: v.Base, pc: v.Base}
	end := v.Base + len(code)
	var events []int
	for range 5 {
		events = append(events, r.operands(1)[0].N)
		r.pc++
	}
	lines := map[int]Line{}
	jumps := map[int]bool{}
	work := append([]int(nil), events...)
	for len(work) > 0 {
		pc := work[len(work)-1]
		work = work[:len(work)-1]
		if _, seen := lines[pc]; seen || pc < v.Base || pc >= end || tables[pc] {
			continue
		}
		l := v.decode(r, pc)
		lines[pc] = l
		for _, t := range l.Targets {
			jumps[t] = true
		}
		work = append(work, l.Targets...)
		switch {
		case endsFlow(l.Cmd):
		case IsIf(l.Cmd.Op):
			work = append(work, l.Next)
			if n := v.decode(r, l.Next); n.Cmd.Known && !tables[l.Next] {
				work = append(work, n.Next)
			}
		default:
			work = append(work, l.Next)
		}
	}
	return events, lines, jumps
}

// SortedPCs are the lines' addresses in order.
func SortedPCs(lines map[int]Line) []int {
	pcs := make([]int, 0, len(lines))
	for pc := range lines {
		pcs = append(pcs, pc)
	}
	sort.Ints(pcs)
	return pcs
}

// List writes a listing of one ECL block (the bytes after its 2-byte
// header). A count held in a variable is listed as 0 operands. A
// development aid.
func (v *Version) List(code []byte, w io.Writer) {
	events, lines, _ := v.Walk(code)
	for i, pc := range events {
		fmt.Fprintf(w, "event %d -> %04X\n", i, pc)
	}
	for _, pc := range SortedPCs(lines) {
		l := lines[pc]
		var as []string
		for _, a := range l.Args {
			switch a.Type {
			case ArgByte:
				as = append(as, fmt.Sprint(a.N))
			case ArgWord:
				as = append(as, fmt.Sprintf("#%04X", a.N))
			case ArgVarWord:
				as = append(as, fmt.Sprintf("&[%04X]", a.N))
			case ArgVar:
				as = append(as, fmt.Sprintf("[%04X]", a.N))
			case ArgText:
				as = append(as, fmt.Sprintf("%q", a.S))
			case ArgTextVar:
				as = append(as, fmt.Sprintf("$[%04X]", a.N))
			default:
				as = append(as, fmt.Sprintf("?%02X:%X", a.Type, a.N))
			}
		}
		fmt.Fprintf(w, "%04X  %02X %s %s\n", pc, l.Byte, OpName(l.Cmd.Op), strings.Join(as, " "))
	}
}

// DecodeBlock decodes one block and adds the hints and folds; the text form

// Folding is whether an import folds the vendor gates (FoldVendorGates).
type Folding byte

const (
	FoldGates Folding = iota // the game's form
	KeepGates                // the byte code's gates, for a test comparing the two
)

// the import stores.
func DecodeBlock(code []byte, area int, names *Names, fold Folding) (string, []FoldedGate, error) {
	s, err := Decode(code, names)
	if err != nil {
		return "", nil, err
	}
	MarkExits(s, area, names.Version)
	var folds []FoldedGate
	if fold == FoldGates {
		folds = FoldVendorGates(s, area, names)
	}
	text, err := Write(s, names)
	return text, folds, err
}
