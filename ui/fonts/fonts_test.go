package fonts

import "testing"

func TestFonts(t *testing.T) {
	all := All()
	if len(all) != 58 {
		t.Fatalf("%d fonts", len(all))
	}
	for _, f := range all {
		if len(f.Bits) != 2048 {
			t.Fatal(f.Name, len(f.Bits))
		}
	}
	if len(Get("IBM_BIOS")) != 2048 || Get("nope") != nil {
		t.Fatal("Get")
	}
	l := Load(t.TempDir())
	if len(l) != 58 || Find(l, "IBM_BIOS") < 0 || Find(l, "x") != -1 {
		t.Fatal("Load/Find")
	}
}
