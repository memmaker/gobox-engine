package script

import (
	"slices"
)

// The fold pass is the decompiler's (cmd/eclscript): a Yes/No gate whose Yes
// branch reaches a vendor (shop, temple, training hall) is cut, so the
// player chooses at the vendor's own menu at once. The vendor's exit runs the
// cut No branch. A gate with anything but plain set-up before the vendor is
// left alone: that work must not run ahead of the player's choice.

const (
	OpClearMonsters byte = 0x1C
	OpProgram       byte = 0x38
	programParty         = 0 // Program's party menu (ovr04:2E10), a hall's training
)

// FoldedGate is one gate the pass cut.
type FoldedGate struct {
	Area    int
	Gate    int    // the gate's instruction number before the pass
	Vendor  string // shop, temple or training
	Yes, No int    // the branches' instruction numbers before the pass
}

// vendorAt is the vendor starting at instruction i: its name and instruction count.
func vendorAt(names *Names, code []Instr, i int) (string, int) {
	kind := ""
	for k := i; k < len(code); k++ {
		in := code[k]
		switch {
		case in.Op == OpShop:
			return "shop", k - i + 1
		case in.Op == OpProgram && k == i && len(in.Args) == 1 && in.Args[0] == (Arg{N: programParty}):
			return "training", 1
		case in.Op == OpCombat:
			return kind, k - i + 1 // no flag set: a fight, no vendor
		case in.Op == OpSave && len(in.Args) == 2 && in.Args[0] == (Arg{N: 1}) && names.Vars[in.Args[1].N] == varShop:
			kind = "shop"
		case in.Op == OpSave && len(in.Args) == 2 && in.Args[0] == (Arg{N: 1}) && names.Vars[in.Args[1].N] == varTemple:
			kind = "temple"
		case in.Op == OpSave, in.Op == OpTreasure, in.Op == OpTreasureOf, setUpOp(in.Op):
		default:
			return "", 0
		}
	}
	return "", 0
}

// setUpOp may stand between a gate and its vendor; the fold still runs it.
func setUpOp(op byte) bool {
	return op == OpClearMonsters || op == OpPrint || op == OpPrintClear
}

// yesNoGate: instruction i asks Yes/No and what follows branches on it; the
// two branches and the gate's length. Either a menu (or a gosub of a section
// that is only that menu) and an OnGoto, or a gosub of a section returning
// the comparison with one row, an If and a Goto.
func yesNoGate(s *Script, i int) (yes, no, length int, ok bool) {
	code := s.Code
	in := code[i]
	menu := in
	if in.Op == OpGosub && len(in.Args) == 1 && in.Args[0].Label {
		menu = code[in.Args[0].N]
	}
	if slices.Equal(MenuLabels(menu), yesNoRows) && i+1 < len(code) {
		b := code[i+1]
		if b.Op == OpOnGoto && len(b.Args) == 4 && b.Args[2].Label && b.Args[3].Label {
			return b.Args[2].N, b.Args[3].N, 2, true
		}
	}
	if in.Op != OpGosub || len(in.Args) != 1 || !in.Args[0].Label || i+3 >= len(code) {
		return 0, 0, 0, false
	}
	_, items, k, found := s.menuOf(in.Args[0].N)
	cond, jump := code[i+1].Op, code[i+2]
	if !found || !slices.Equal(items, yesNoRows) || cond != OpIfEq && cond != OpIfNe || jump.Op != OpGoto || !jump.Args[0].Label {
		return 0, 0, 0, false
	}
	jumpsOn := k // the row whose answer takes the Goto
	if cond == OpIfNe {
		jumpsOn = 1 - k
	}
	if jumpsOn == noRow {
		return i + 3, jump.Args[0].N, 3, true
	}
	return jump.Args[0].N, i + 3, 3, true
}

// yesNoRows are a gate's rows.
var yesNoRows = []string{"YES", "NO"}

const noRow = 1

// jumpTo is a Goto to command to.
func jumpTo(to int) Instr {
	return Instr{Op: OpGoto, Args: []Arg{{Type: 1, N: to, Label: true}}}
}

// gateClear is the text clear of a gate's asking section, run after the
// answer on both branches, if it has one.
func gateClear(s *Script, i int) (Instr, bool) {
	in := s.Code[i]
	if in.Op != OpGosub || len(in.Args) != 1 || !in.Args[0].Label {
		return Instr{}, false
	}
	asked := false
	for _, c := range s.Code[in.Args[0].N:] {
		switch {
		case c.Op == OpHMenu:
			asked = true
		case asked && c.Op == OpPrintClear:
			return c, true
		case isTerminator(c.Op):
			return Instr{}, false
		}
	}
	return Instr{}, false
}

// cutGate replaces the gate at i, length commands long, with its asking
// section's clear (a gate is never shorter than the two commands this
// takes) and a jump to the Yes branch.
func cutGate(s *Script, i, length, yes int, drop map[int]bool) {
	next := i
	if c, ok := gateClear(s, i); ok {
		s.Code[next] = c
		next++
	}
	s.Code[next] = jumpTo(yes)
	for k := next + 1; k < i+length; k++ {
		drop[k] = true
	}
}

// FoldVendorGates cuts area's vendor gates and returns them.
func FoldVendorGates(s *Script, area int, names *Names) []FoldedGate {
	targets := map[int]bool{}
	for _, in := range s.Code {
		for _, a := range in.Args {
			if a.Label {
				targets[a.N] = true
			}
		}
	}
	for _, i := range s.Labels {
		targets[i] = true
	}
	for _, i := range s.Events {
		targets[i] = true
	}
	var folds []FoldedGate
	drop := map[int]bool{}
	for i := range s.Code {
		yes, no, gate, ok := yesNoGate(s, i)
		if !ok {
			continue
		}
		v := yes
		for v < len(s.Code) && setUpOp(s.Code[v].Op) {
			v++
		}
		if v >= len(s.Code) {
			continue
		}
		name, n := vendorAt(names, s.Code, v)
		if name == "" {
			continue
		}
		if a := v + n; a == no || a < len(s.Code) && s.Code[a].Op == OpGoto && s.Code[a].Args[0].Label && s.Code[a].Args[0].N == no {
			// the vendor already goes on to the No branch: only the gate goes
			inside := false
			for k := i + 1; k < i+gate; k++ {
				inside = inside || targets[k]
			}
			if inside {
				continue
			}
			cutGate(s, i, gate, yes, drop)
			folds = append(folds, FoldedGate{Area: area, Gate: i, Vendor: name, Yes: yes, No: no})
			continue
		}
		end := v + n
		for end < len(s.Code) && s.Code[end].Op != OpExit {
			end++
		}
		if no > v+n && no <= end {
			continue // cutting the tail would cut the No branch: left for the report
		}
		tail := true // what follows the vendor up to its exit is replaced: nothing may jump in
		for k := v + n + 1; k <= end && k < len(s.Code); k++ {
			tail = tail && !targets[k]
		}
		if !tail || end >= len(s.Code) {
			continue
		}
		inside := false // nothing may jump into the cut gate
		for k := i + 1; k < i+gate; k++ {
			inside = inside || targets[k]
		}
		if inside {
			continue
		}
		cutGate(s, i, gate, yes, drop)
		s.Code[v+n] = jumpTo(no)
		for k := v + n + 1; k <= end; k++ {
			drop[k] = true
		}
		folds = append(folds, FoldedGate{Area: area, Gate: i, Vendor: name, Yes: yes, No: no})
	}
	s.removeInstrs(drop)
	s.DropNextGotos()
	return folds
}
