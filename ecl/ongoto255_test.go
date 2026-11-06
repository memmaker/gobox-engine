package ecl

import "testing"

// An ON GOTO count is a byte: SOTSB's town exit (ecl 35004) has 64 targets,
// and operands are 1-based, so the arrays need room for index 255.
func TestOnGotoManyTargets(t *testing.T) {
	for _, count := range []int{64, 255} {
		b := []byte{CmdOnGoto, 0, byte(count - 1), 0, byte(count)}
		for k := 0; k < count; k++ {
			b = append(b, ArgAddress, byte(k), 0x80)
		}
		vm := run(newFake(), 1, b...)
		if want := uint16(0x8000 | (count - 1)); vm.PC != want {
			t.Fatalf("count %d: pc %x, want %x", count, vm.PC, want)
		}
	}
}
