package ecl

import (
	"fmt"
	"sort"
)

// Decode decodes an area's ECL byte code (an ECL block without its
// two-byte header) into a Script: the commands reachable from the five
// events, their labels (named by names.Vars where it names an address,
// else L_XXXX) and the tables. cmd/eclscript writes the result with
// WriteScript.
func Decode(code []byte, names *Names) (*Script, error) {
	v := names.Version
	events, lines, jumps := v.Walk(code)
	end := v.Base + len(code)
	tables := v.tables(lines, end)
	isCode := map[int]bool{}
	for pc, l := range lines {
		for a := pc; a < l.Next; a++ {
			if isCode[a] {
				return nil, fmt.Errorf("instructions overlap at %04X", a)
			}
			isCode[a] = true
		}
	}
	for pc := range jumps {
		if _, ok := lines[pc]; !ok {
			return nil, fmt.Errorf("a jump to %04X, which is not code", pc)
		}
	}
	for pc := range tables {
		if isCode[pc] {
			return nil, fmt.Errorf("the table at %04X is code", pc)
		}
	}
	s := &Script{Labels: map[string]int{}, At: map[int]string{}, Tables: map[string]int{}}
	name := func(pc int, def string) string {
		if n := names.Vars[pc]; n != "" {
			return n
		}
		return def
	}
	// the code, in address order; a label where a jump lands, an event
	// starts or the names name the address
	pcs := SortedPCs(lines)
	index := map[int]int{} // instruction number by address
	for _, pc := range pcs {
		index[pc] = len(s.Code)
		l := lines[pc]
		s.Code = append(s.Code, Instr{Op: l.Cmd.Op})
		if jumps[pc] || names.Vars[pc] != "" {
			s.label(name(pc, fmt.Sprintf("L_%04X", pc)), index[pc])
		}
		if !endsFlow(l.Cmd) || !l.Cmd.Known {
			if _, ok := lines[l.Next]; !ok {
				// the original runs on into what is not code
				s.Code = append(s.Code, Instr{Op: OpExit, Comment: fmt.Sprintf("the original runs on into data at %04X", l.Next)})
			}
		}
	}
	for i, pc := range events {
		s.Events[i] = index[pc]
		s.label(fmt.Sprintf("event%d", i), index[pc])
	}
	// the tables, each to the next code or table, laid out one after the
	// other from the code base as the parser lays them (loadScript copies Data
	// there): their references move with them
	var tpcs []int
	for pc := range tables {
		tpcs = append(tpcs, pc)
	}
	sort.Ints(tpcs)
	moved := map[int]int{}
	for _, pc := range tpcs {
		moved[pc] = v.Base + len(s.Data)
		s.Tables[name(pc, fmt.Sprintf("T_%04X", pc))] = moved[pc]
		a := pc
		for a < end && !isCode[a] && (a == pc || !tables[a]) {
			a++
		}
		s.Data = append(s.Data, code[pc-v.Base:a-v.Base]...)
	}
	// the operands
	for _, pc := range pcs {
		l := lines[pc]
		i := index[pc]
		for _, a := range l.Args {
			arg, err := decodeArg(a, l, index)
			if err != nil {
				return nil, fmt.Errorf("%04X: %v", pc, err)
			}
			if at, ok := moved[arg.N]; ok && !arg.Label && (arg.Type == 1 || arg.Type == 3) {
				arg.N = at
			}
			s.Code[i].Args = append(s.Code[i].Args, arg)
		}
	}
	return s, nil
}

// label names instruction i.
func (s *Script) label(name string, i int) {
	s.Labels[name] = i
	if _, ok := s.At[i]; !ok {
		s.At[i] = name
	}
}

// decodeArg is an encoded operand as an Arg: a jump target as its
// instruction number, packed text with its encoded length.
func decodeArg(a Arg, l Line, index map[int]int) (Arg, error) {
	switch a.Type {
	case ArgByte, ArgWord, ArgVarWord, ArgLong, ArgVarLong, ArgTextVar:
		return a, nil
	case ArgText:
		a.N = (len(a.S)*6 + 7) / 8
		return a, nil
	case ArgVar:
		for _, t := range l.Targets {
			if t == a.N {
				return Arg{Type: ArgVar, N: index[a.N], Label: true}, nil
			}
		}
		return a, nil
	}
	return Arg{}, fmt.Errorf("operand type %02X", a.Type)
}
