package games

import "testing"

// TestTown: the town economy the import writes: 10 services, the hall's fee 0.
func TestSotsbTown(t *testing.T) {
	town := SotsbTown()
	if len(town.Temple) != 10 || town.TrainPrice != 0 {
		t.Fatalf("temple of %d, hall %d", len(town.Temple), town.TrainPrice)
	}
	for _, s := range town.Temple {
		if !s.Declines || s.Done == "" || s.Not == "" { // the temple says what it did and why not, never asks to cast anyway
			t.Errorf("service %+v lacks its words", s)
		}
		if s.Name == "" || s.Kind == "" {
			t.Errorf("service %+v", s)
		}
	}
}
