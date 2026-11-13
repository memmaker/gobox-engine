package games

import "testing"

// TestTown: the town economy the import writes: 11 services (no Restoration), the hall's fee 0.
func TestPodTown(t *testing.T) {
	town := PodTown()
	if len(town.Temple) != 11 || town.TrainPrice != 0 {
		t.Fatalf("temple of %d, hall %d", len(town.Temple), town.TrainPrice)
	}
	if !town.Bundles { // shopPrice (O052:0484): a bundle costs its count times the price
		t.Error("pod's shops sell bundles at their count times the price")
	}
	for _, s := range town.Temple {
		if s.Name == "" || s.Kind == "" {
			t.Errorf("service %+v", s)
		}
	}
}
