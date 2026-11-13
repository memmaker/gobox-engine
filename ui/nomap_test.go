package ui

import (
	"testing"

	"gobox/game"
)

func TestNoMapZones(t *testing.T) {
	old := Host.NoMapZones
	defer func() { Host.NoMapZones = old }()
	v := game.Automap{NoMap: true}
	Host.NoMapZones = true
	if !mapBarred(v) {
		t.Fatal("option on, area marked: the maps must be barred")
	}
	Host.NoMapZones = false
	if mapBarred(v) {
		t.Fatal("option off: the maps are always available")
	}
	Host.NoMapZones = true
	if mapBarred(game.Automap{}) {
		t.Fatal("unmarked area: the maps are available")
	}
	if !DefaultOptions.NoMapZones {
		t.Fatal("No Map Zones is on by default")
	}
}
