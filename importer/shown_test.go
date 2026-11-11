package importer

import (
	"reflect"
	"testing"
)

// TestShown: an Amiga step is dropped without the Amiga art, kept with it,
// and Short's indexes follow the steps.
func TestShown(t *testing.T) {
	s := Sequence{Steps: []SequenceStep{{Page: 0}, {Page: 1, Edition: EditionAmiga}, {Page: 2}}, Short: []int{0, 2}}
	dos := s.Shown(false)
	if len(dos.Steps) != 2 || dos.Steps[1].Page != 2 || !reflect.DeepEqual(dos.Short, []int{0, 1}) {
		t.Errorf("DOS: %+v", dos)
	}
	am := s.Shown(true)
	if len(am.Steps) != 3 || am.Steps[1].Edition != EditionAny || !reflect.DeepEqual(am.Short, []int{0, 2}) {
		t.Errorf("Amiga: %+v", am)
	}
	if only := (Sequence{Steps: []SequenceStep{{Edition: EditionAmiga}}, Short: []int{0}}).Shown(false); only.Short == nil || len(only.Short) != 0 {
		t.Errorf("empty Short must stay non-nil: %+v", only)
	}
}
