package present

import (
	"testing"

	"gobox/front"
)

func TestNoMapZones(t *testing.T) {
	old := Host.NoMapZones
	defer func() { Host.NoMapZones = old }()
	v := front.View{NoMap: true}
	Host.NoMapZones = true
	if !mapBarred(v) {
		t.Fatal("option on, area marked: the maps must be barred")
	}
	Host.NoMapZones = false
	if mapBarred(v) {
		t.Fatal("option off: the maps are always available")
	}
	Host.NoMapZones = true
	if mapBarred(front.View{}) {
		t.Fatal("unmarked area: the maps are available")
	}
	if !DefaultOptions.NoMapZones {
		t.Fatal("No Map Zones is on by default")
	}
}
