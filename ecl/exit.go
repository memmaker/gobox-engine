package ecl

import (
	"strings"

	"gobox/game"
)

// exitLabel marks a menu's exit row in script text: Menu("ATTACK", exit="LEAVE").
const exitLabel = "exit"

// The exit heuristic is the decompiler's alone (cmd/eclscript, MarkExits): it
// reads the row labels once and writes the typed hint. The engine reads only
// Instr.Exit and never compares a label.

// exitPriority are the labels that mean "back out", best first: a label that
// is one of them, or starts with one followed by a space where it ends in a
// space. NO comes after LEAVE so YES, NO, ATTACK, LEAVE exits on LEAVE.
var exitPriority = []string{
	"EXIT", "LEAVE", "DONE", "CANCEL", "NONE", "NOTHING", "GO BACK", "BACK OFF",
	"TRY TO LEAVE", "SEND HIM AWAY", "REJECT IT", "REFUSE", "REFUSE TO GIVE",
	"LEAVE ", "DON'T ", "DO NOT ", "IGNORE ", "NO", "NON",
}

// guessExit is the row of rows that backs out, by the labels.
func guessExit(rows []string) (int, bool) {
	for _, want := range exitPriority {
		for i, r := range rows {
			r = strings.TrimSpace(r)
			if r == want || strings.HasSuffix(want, " ") && strings.HasPrefix(r, want) {
				return i, true
			}
		}
	}
	return 0, false
}

// MenuLabels are the labels of menu instruction in, nil for any other command
// or a one-row "press a key".
func MenuLabels(in Instr) []string {
	var a []Arg
	switch {
	case in.Op == OpHMenu && len(in.Args) > 3:
		a = in.Args[2:]
	case in.Op == OpVMenu && len(in.Args) > 4:
		a = in.Args[3:]
	}
	var rows []string
	for _, x := range a {
		rows = append(rows, x.S)
	}
	return rows
}

// MarkExits writes the exit hints on the menus of area's decoded script.
func MarkExits(s *Script, area int, v *Version) {
	for i := range s.Code {
		rows := MenuLabels(s.Code[i])
		if rows == nil {
			continue
		}
		row, ok := guessExit(rows)
		if o, over := v.ExitOverrides[ExitPlace{area, i}]; over {
			row, ok = o, o != NoExitRow
		}
		if ok {
			s.Code[i].Exit = game.EscTo(row)
		}
	}
}
