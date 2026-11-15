package adventure

import (
	"fmt"

	"gobox/rules"
)

// StatusWriteKind is what a script's write of the selected member's status does. The games differ in rule,
// not just in numbers (POR's write of 0 only marks the member for a later removal, BR's sets a status).
type StatusWriteKind int

const (
	StatusWriteSet         StatusWriteKind = iota // the member gets StatusWrite.Status (active only when Okay)
	StatusWriteInactive                           // out of the active members, the status kept
	StatusWriteMarkRemoval                        // a later LOAD CHARACTER with bit 7 removes the member; the status is not changed
	StatusWriteNothing                            // the original ignores the value
)

// StatusWrite is a script's write of the status, decoded.
type StatusWrite struct {
	Kind   StatusWriteKind
	Status rules.Status // StatusWriteSet's
}

// statusWriteCommands are the kinds with no status of their own, in the order of their canonical numbers:
// the canonical status write is a place in StatusKinds (a set) or, above them, len(StatusKinds)+place here.
var statusWriteCommands = []StatusWriteKind{StatusWriteInactive, StatusWriteMarkRemoval, StatusWriteNothing}

// CanonStatusWrite is w as the canonical number the field machinery carries to the binding.
func CanonStatusWrite(w StatusWrite) int64 {
	if w.Kind == StatusWriteSet {
		for n, st := range StatusKinds {
			if st == w.Status {
				return int64(n)
			}
		}
		panic(fmt.Sprintf("status write: no kind for %s", w.Status))
	}
	for n, k := range statusWriteCommands {
		if k == w.Kind {
			return int64(len(StatusKinds) + n)
		}
	}
	panic(fmt.Sprintf("status write: kind %d", w.Kind))
}

// DecodeStatusWrite is the one place a canonical number becomes the typed command.
func DecodeStatusWrite(c int64) StatusWrite {
	switch {
	case c >= 0 && c < int64(len(StatusKinds)):
		return StatusWrite{Kind: StatusWriteSet, Status: StatusKinds[c]}
	case c >= int64(len(StatusKinds)) && c < int64(len(StatusKinds)+len(statusWriteCommands)):
		return StatusWrite{Kind: statusWriteCommands[c-int64(len(StatusKinds))]}
	}
	panic(fmt.Sprintf("status write: canonical %d", c))
}

// PodStatusOf is the rules status of pod's C_Status number n; ok false for a number with no status.
func PodStatusOf(n int64) (rules.Status, bool) {
	for st, p := range PodStatus {
		if p == n {
			return st, true
		}
	}
	return "", false
}

// SelStatus writes of the games (Pascal: eclmenu.pas SpecialWriteZ1, see each).

// porStatusWrite is POR's and COTAB's SpecialWriteZ1 $100 (por-re, cotab-re eclmenu.pas): value >= $80 takes the
// member out of the active ones and $87 also stones it; 0 only sets dumpOnStatusClear (EclLoadCharacter removes
// the member later, its status stays); any other value does nothing.
func porStatusWrite(v int64) StatusWrite {
	const outOfPlay, stoned = 0x80, 0x87
	switch {
	case v == stoned:
		return StatusWrite{Kind: StatusWriteSet, Status: rules.StatusStoned}
	case v >= outOfPlay:
		return StatusWrite{Kind: StatusWriteInactive}
	case v == 0:
		return StatusWrite{Kind: StatusWriteMarkRemoval}
	}
	return StatusWrite{Kind: StatusWriteNothing}
}

// sotsbStatusWrite is sotsb's SpecialWriteZ1 $100 (sotsb-re eclmenu.pas): 1 only sets status OKAY and enabled 1.
func sotsbStatusWrite(v int64) StatusWrite {
	if v == SotsbStatus[rules.StatusOkay] {
		return StatusWrite{Kind: StatusWriteSet, Status: rules.StatusOkay}
	}
	return StatusWrite{Kind: StatusWriteNothing}
}

// podStatusWrite is pod's SpecialWriteZ1 $1ED (pod-re eclmenu.pas): status := Byte(value), in play when OKAY.
// A number with no rules status (2 and 10 up) is stored raw by the original; the rules cannot hold it: it fails.
func podStatusWrite(v int64) StatusWrite {
	st, ok := PodStatusOf(v)
	if !ok {
		panic(fmt.Sprintf("pod: SelStatus %d: no rules status", v))
	}
	return StatusWrite{Kind: StatusWriteSet, Status: st}
}

// brStatusWrite is BR's setChar (ecl.go): r[212] = v & 127, a status BR has a rules status for.
func brStatusWrite(v int64) StatusWrite {
	for st, b := range BRStatus {
		if b == v&BRStatusMask {
			return StatusWrite{Kind: StatusWriteSet, Status: st}
		}
	}
	panic(fmt.Sprintf("br: SelStatus %d: no rules status", v&BRStatusMask))
}
