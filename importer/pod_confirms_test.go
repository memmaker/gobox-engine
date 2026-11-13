package importer

import "testing"

// TestConfirms: GameData's Confirms are the original's DS strings.
func TestConfirms(t *testing.T) {
	want := map[string][]string{
		"Drop":         {"Lose it? "},
		"DropScribe":   {"is it Okay to lose it? "},
		"Sell":         {"Will you sell it? "},
		"SellMany":     {"Will you sell them? "},
		"SellFree":     {"Shall I dispose of it for you? "},
		"SellFreeMany": {"Would you like to discard them? "},
		"DropMember":   {"Drop %s forever? ", "Are you sure? "},
	}
	c := podGameData("").Confirms
	if len(c) != len(want) || c["DropMember"].Yes || !c["Sell"].Yes || !c["SellMany"].Yes || !c["SellFree"].Yes || !c["SellFreeMany"].Yes {
		t.Fatalf("confirms %v", c)
	}
	for act, asks := range want {
		got := c[act].Asks
		if len(got) != len(asks) {
			t.Fatalf("%s: %q", act, got)
		}
		for i := range asks {
			if got[i] != asks[i] {
				t.Errorf("%s[%d] = %q, want %q", act, i, got[i], asks[i])
			}
		}
	}
}
