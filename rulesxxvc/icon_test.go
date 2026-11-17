package rulesxxvc

import (
	"testing"

	"gobox/rules"
)

// TestIconPicsRoundTrip: the icon designer's undo (SetIconPics with the
// IconPicsFact it read) resolves; EndIcon(false) panicked with ErrNotPorted.
func TestIconPicsRoundTrip(t *testing.T) {
	r, who := trainer(t)
	p, err := r.Fact(who, rules.IconPicsFact{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(nil, rules.SetIconPics{Who: who, Pics: p.(rules.IconPics)}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(nil, rules.ResetIcon{Who: who}, nil); err != nil {
		t.Fatal(err)
	}
}
