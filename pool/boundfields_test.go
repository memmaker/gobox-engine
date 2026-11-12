package pool

import (
	"os"
	"slices"
	"testing"
)

// TestBoundFieldsPOR: POR's layouts name every field the engine binds but the VM words POR
// keeps in the ruleset.
func TestBoundFieldsPOR(t *testing.T) {
	got, err := UnboundFields(os.DirFS("../../por-go/data/assets/scripts"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{fieldRemainder, fieldLastBlock}; !slices.Equal(got, want) {
		t.Errorf("unbound %v, want %v", got, want)
	}
}
