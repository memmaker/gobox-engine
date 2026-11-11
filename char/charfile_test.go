package char

import (
	"rules1e"
	"testing"
)

// NAME.pow keeps the powers and the daily uses; an older one has only powers.
func TestPowerBytes(t *testing.T) {
	c := rules1e.CharacterFile{Powers: []int{1, 100}, Surges: 3, Laid: 1, Tough: 5, Policies: map[int]rules1e.Policy{44: rules1e.Never, 9: rules1e.OnHit}}
	var d rules1e.CharacterFile
	readPowers(&d, powerBytes(c))
	if len(d.Powers) != 2 || d.Powers[1] != 100 || d.Surges != 3 || d.Laid != 1 || d.Tough != 5 || d.Policies[44] != rules1e.Never || d.Policies[9] != rules1e.OnHit || len(d.Policies) != 2 {
		t.Errorf("read back %v %d %d %v", d.Powers, d.Surges, d.Laid, d.Policies)
	}
	d = rules1e.CharacterFile{}
	readPowers(&d, []byte{1, 2})
	if len(d.Powers) != 2 || d.Surges != 0 || d.Policies != nil {
		t.Errorf("an older file: %v %d", d.Powers, d.Surges)
	}
}

// A character's files survive Pack and Unpack (the pointer bytes aside).
func TestPackRoundTrip(t *testing.T) {
	var s rules1e.Sheet
	s.CharName, s.HP, s.Levels[2] = "Brok", 9, 3
	rec := append(SheetBytes(s), PoolTrailer(rules1e.PoolIDs{IconHeadID: "x", Roster: 4})...)
	it := WriteRecords([]*Item{{Type: 3, Count: 2, Price: 0x7A79}})
	f := Files{Record: rec, Items: it, Plan: []byte{5, 6}, Dual: []byte{1}, Powers: []byte{7, 0, 1, 2, 3}, Portrait: []byte("png")}
	b, err := Pack285(f)
	if err != nil {
		t.Fatal(err)
	}
	g, err := Unpack285(b, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(g.Record) != string(f.Record) || string(g.Items) != string(f.Items) || string(g.Plan) != string(f.Plan) ||
		string(g.Dual) != string(f.Dual) || string(g.Powers) != string(f.Powers) || string(g.Portrait) != string(f.Portrait) || g.Affects != nil {
		t.Errorf("round trip: %+v\nwant %+v", g, f)
	}
}
