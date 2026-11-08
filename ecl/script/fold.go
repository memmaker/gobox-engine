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
	switch in := code[i]; {
	case in.Op == OpTreasure && i+4 < len(code) && code[i+4].Op == OpCombat:
		return "shop", 5
	case in.Op == OpSave && len(in.Args) == 2 && in.Args[0] == (Arg{N: 1}) && names.Vars[in.Args[1].N] == varTemple && i+1 < len(code) && code[i+1].Op == OpCombat:
		return "temple", 2
	case in.Op == OpProgram && len(in.Args) == 1 && in.Args[0] == (Arg{N: programParty}):
		return "training", 1
	}
	return "", 0
}

// setUpOp may stand between a gate and its vendor; the fold still runs it.
func setUpOp(op byte) bool {
	return op == OpClearMonsters || op == OpPrint || op == OpPrintClear
}

// yesNoGate: instruction i asks Yes/No (a gosub of a section that is only that
// menu, or the menu itself) and i+1 branches on it; the two branches.
func yesNoGate(code []Instr, i int) (yes, no int, ok bool) {
	in := code[i]
	menu := in
	if in.Op == OpGosub && len(in.Args) == 1 && in.Args[0].Label {
		menu = code[in.Args[0].N]
	}
	if !slices.Equal(MenuLabels(menu), []string{"YES", "NO"}) || i+1 >= len(code) {
		return 0, 0, false
	}
	b := code[i+1]
	if b.Op != OpOnGoto || len(b.Args) != 4 || !b.Args[2].Label || !b.Args[3].Label {
		return 0, 0, false
	}
	return b.Args[2].N, b.Args[3].N, true
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
		yes, no, ok := yesNoGate(s.Code, i)
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
		end := v + n
		for end < len(s.Code) && s.Code[end].Op != OpExit {
			end++
		}
		tail := true // what follows the vendor up to its exit is replaced: nothing may jump in
		for k := v + n + 1; k <= end && k < len(s.Code); k++ {
			tail = tail && !targets[k]
		}
		if !tail || end >= len(s.Code) {
			continue
		}
		goTo := func(to int) Instr { return Instr{Op: OpGoto, Args: []Arg{{Type: 1, N: to, Label: true}}} }
		s.Code[i], s.Code[i+1] = goTo(yes), Instr{Op: OpExit}
		drop[i+1] = true
		s.Code[v+n] = goTo(no)
		for k := v + n + 1; k <= end; k++ {
			drop[k] = true
		}
		folds = append(folds, FoldedGate{Area: area, Gate: i, Vendor: name, Yes: yes, No: no})
	}
	s.removeInstrs(drop)
	s.DropNextGotos()
	return folds
}
