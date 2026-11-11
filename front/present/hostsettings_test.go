package present

import (
	"encoding/json"
	"testing"
)

// TestHostOptionsSettingsJSON: a SETTINGS.JSON written by por's cmd/por
// HostOptions (before the move, por-go b2e395e^ cmd/por/host.go) loads
// unchanged into HostOptions: same keys, same values, nothing dropped.
func TestHostOptionsSettingsJSON(t *testing.T) {
	old := `{
 "MapOff": true, "PartyRow": true, "MapAlpha": 0.6, "Cards": 2, "Map": 3, "Pic": 4,
 "Zoom": 1.5, "DamageNumbers": false, "NoMapZones": false, "SIDMusic": true, "PixelWalls": true,
 "Deskew": false, "Music": 0.25, "Effects": 0.5, "Voice": 0.75,
 "Dasp": false, "Soundtrack": true, "Steps": false, "Combat4e": true,
 "PathSearch": true, "QuickSearch": true, "DaySky": true, "Explicit": true, "DOSArt": true, "FullIntro": true
}`
	want := HostOptions{
		Options: Options{MapOff: true, PartyRow: true, MapAlpha: 0.6, Cards: 2, Map: 3, Pic: 4,
			Zoom: 1.5, SIDMusic: true, PixelWalls: true},
		Music: 0.25, Effects: 0.5, Voice: 0.75, Soundtrack: true, Combat4e: true,
		PathSearch: true, QuickSearch: true, DaySky: true, Explicit: true, DOSArt: true, FullIntro: true,
	}
	got := DefaultHostOptions
	if err := json.Unmarshal([]byte(old), &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("loaded %+v\nwant   %+v", got, want)
	}
	// The round trip writes the same keys back: no field lost or renamed.
	b, _ := json.Marshal(got)
	var keys, oldKeys map[string]any
	json.Unmarshal(b, &keys)
	json.Unmarshal([]byte(old), &oldKeys)
	if len(keys) != len(oldKeys) {
		t.Fatalf("keys %d, old file %d", len(keys), len(oldKeys))
	}
	for k := range oldKeys {
		if _, ok := keys[k]; !ok {
			t.Errorf("key %s not written back", k)
		}
	}
	// The defaults are cmd/por's.
	if d := (HostOptions{Options: DefaultOptions, Deskew: true, Music: 1, Effects: 1, Voice: 1, Dasp: true, Steps: true}); DefaultHostOptions != d {
		t.Errorf("defaults %+v", DefaultHostOptions)
	}
}
