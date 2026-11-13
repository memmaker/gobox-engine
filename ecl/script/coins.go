package script

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

// CoinUnit is money a script may keep: copper or gold, never platinum,
// electrum or silver (Felix 2026-11-12). The import rewrites every other coin
// to one of these, by value (GoldCoins).
type CoinUnit uint8

const (
	NoCoin CoinUnit = iota
	CoinCopper
	CoinGold
)

// CoinCell is one element of a coin field (layout coins=): worth Times of
// Unit. Times 1 is the unit's own cell; the zero CoinCell is no coin.
type CoinCell struct {
	Unit  CoinUnit
	Times int
}

// parseCoins reads a layout coins= list, one entry an element: "-" no coin,
// "c" copper, "g" gold, "<n>c" or "<n>g" a coin worth n of them.
func parseCoins(s string, n int) ([]CoinCell, error) {
	parts := strings.Split(s, ",")
	if len(parts) != n {
		return nil, fmt.Errorf("coins=%s: %d entries, the field has %d", s, len(parts), n)
	}
	cells := make([]CoinCell, n)
	for i, p := range parts {
		if p == "-" {
			continue
		}
		var u CoinUnit
		switch {
		case strings.HasSuffix(p, "c"):
			u = CoinCopper
		case strings.HasSuffix(p, "g"):
			u = CoinGold
		default:
			return nil, fmt.Errorf("coins=%s: entry %q", s, p)
		}
		times := 1
		if num := p[:len(p)-1]; num != "" {
			t, err := strconv.Atoi(num)
			if err != nil || t < 2 {
				return nil, fmt.Errorf("coins=%s: entry %q", s, p)
			}
			times = t
		}
		cells[i] = CoinCell{Unit: u, Times: times}
	}
	return cells, nil
}

// coinMove is where the import moves element i and by how much its value
// grows: the element of its unit worth 1, or i itself when the field has none
// (the cell then holds that unit). ok: i is a coin worth more than one unit.
func (f *Field) coinMove(i int) (to, times int, ok bool) {
	if i < 0 || i >= len(f.Coins) || f.Coins[i].Times < 2 {
		return 0, 0, false
	}
	c := f.Coins[i]
	for k, o := range f.Coins {
		if o == (CoinCell{Unit: c.Unit, Times: 1}) {
			return k, c.Times, true
		}
	}
	return i, c.Times, true
}

// Moved: element i is a coin the import rewrites to another element; no
// imported script names it (TestScriptsGoldOnly in each game).
func (f *Field) Moved(i int) bool {
	to, _, ok := f.coinMove(i)
	return ok && to != i
}

// scalable: the operands of op are one quantity, so a literal among them
// takes the coin's value (compare, add, subtract, copy).
func scalable(op byte) bool { return op == OpCompare || op == OpAdd || op == OpSub || op == OpSave }

// compareRound is what a compare's literal L (times L in the unit) gains so
// that the tests reading its flags (the If commands after it) hold for the
// unit's value exactly as for the coin's: coin >= L and coin < L are unit >=
// times*L and unit < times*L; coin > L and coin <= L are unit > times*L+times-1
// and unit <= times*L+times-1. coinFirst: the coin is the compare's first
// operand. Not exact: an equality test, tests that want both, or no test
// after the compare (its flags read further on).
func compareRound(next []Instr, coinFirst bool, times int) (round int, exact bool) {
	up, down, n := false, false, 0
	for j := 0; j < len(next) && IsIf(next[j].Op); j++ {
		in := next[j]
		if j+1 < len(next) && !IsIf(next[j+1].Op) {
			j++ // the command the test guards; a test after it reads the same flags
		}
		n++
		op := in.Op
		if !coinFirst { // L op coin: the coin's relation is the mirror
			switch op {
			case OpIfLt:
				op = OpIfGt
			case OpIfGt:
				op = OpIfLt
			case OpIfLe:
				op = OpIfGe
			case OpIfGe:
				op = OpIfLe
			}
		}
		switch op {
		case OpIfGt, OpIfLe:
			up = true
		case OpIfLt, OpIfGe:
			down = true
		default:
			return 0, false
		}
	}
	if n == 0 || up && down {
		return 0, false
	}
	if up {
		return times - 1, true
	}
	return 0, true
}

// GoldCoins rewrites s's coin operands to copper and gold by value: an
// operand naming a coin worth n of its unit names the unit's cell instead,
// and each literal of the command is multiplied by n (1 platinum = 5 gold:
// "platinum >= 3" is "gold >= 15", "take 2 platinum" is "take 10 gold"). A
// command whose other operands are not literals, or that is no compare, add,
// subtract or copy, keeps them as they are and says so in its comment.
func GoldCoins(s *Script, l *Layout) {
	for i := range s.Code {
		in := &s.Code[i]
		times, coin := 0, map[int]bool{}
		for k := range in.Args {
			a := &in.Args[k]
			if a.Label || !isVarArg(a.Type) {
				continue
			}
			r, err := l.Resolve(a.N)
			if err != nil {
				continue // ResolveRefs reports it
			}
			f := l.Field(r)
			to, n, ok := f.coinMove(r.Index)
			if !ok {
				continue
			}
			a.N, coin[k] = f.From+to, true
			if times != 0 && times != n {
				times = -1 // two coins of other values: not one quantity
			} else if times == 0 {
				times = n
			}
		}
		if times == 0 {
			continue
		}
		kept := times < 0 || !scalable(in.Op)
		round, exact := 0, true
		if in.Op == OpCompare && !kept {
			round, exact = compareRound(s.Code[i+1:], coin[0], times)
		}
		for k := range in.Args {
			a := &in.Args[k]
			switch {
			case coin[k] || a.Label:
			case !kept && (a.Type == ArgByte || a.Type == ArgWord || a.Type == ArgLong):
				a.N = a.N*times + round
				if a.Type == ArgByte && a.N > 0xFF {
					a.Type = ArgWord
				}
			case a.Type != ArgText:
				kept = true
			}
		}
		if kept || !exact {
			note := "coins: rewritten to gold/copper; operands not converted by value"
			if !kept {
				note = "coins: rewritten to gold/copper; the compare is exact only for whole coins"
			}
			if in.Comment != "" {
				note = in.Comment + "; " + note
			}
			in.Comment = note
		}
	}
}

// MovedCoins lists the operands of the imported texts in dir that name a coin
// the import moves (Field.Moved): none, when every script deals in gold and
// copper only. Each game's TestScriptsGoldOnly runs it on its import.
func MovedCoins(v *Version, dir string, data fs.FS, fixes []Fix) ([]string, error) {
	var bad []string
	err := eachImportedArea(v, dir, data, fixes, func(area int, s *Script) {
		for i, in := range s.Code {
			for k, a := range in.Args {
				if a.Ref != (Ref{}) && s.Layout.Field(a.Ref).Moved(a.Ref.Index) {
					bad = append(bad, fmt.Sprintf("area %d command %d (%s) operand %d: %s element %d", area, i, OpName(in.Op), k, s.Layout.Field(a.Ref).Name, a.Ref.Index))
				}
			}
		}
	})
	return bad, err
}

// eachImportedArea runs fn on every imported area text in dir.
func eachImportedArea(v *Version, dir string, data fs.FS, fixes []Fix, fn func(area int, s *Script)) error {
	l := Library{Version: v, Dir: dir, Data: data, Fixes: fixes}
	areas := 0
	for area := range maxBlock {
		if _, err := os.Stat(AreaPath(dir, area)); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		s, err := l.Area(area)
		if err != nil {
			return err
		}
		areas++
		fn(area, s)
	}
	if areas == 0 {
		return fmt.Errorf("%s: no imported scripts", dir)
	}
	return nil
}
