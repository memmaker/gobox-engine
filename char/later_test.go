package char

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"testing"

	"rules1e"
)

// LaterRoundTrip checks f through PackLater and UnpackLater: the record
// byte-equal (pointers included), the items and affects too.
func LaterRoundTrip[R Later](t *testing.T, f LaterFiles) rules1e.CharacterJSON {
	t.Helper()
	b, err := PackLater[R](f)
	if err != nil {
		t.Fatal(err)
	}
	g, err := UnpackLater[R](b, f.ItemSize)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(g.Record, f.Record) || !bytes.Equal(g.Items, f.Items) || !bytes.Equal(g.Affects, f.Affects) {
		for i := range f.Record {
			if i < len(g.Record) && g.Record[i] != f.Record[i] {
				t.Fatalf("%T: record byte %d: %d want %d", *new(R), i, g.Record[i], f.Record[i])
			}
		}
		t.Fatalf("%T: round trip differs (items or affects, or length %d/%d)", *new(R), len(g.Record), len(f.Record))
	}
	var c rules1e.CharacterJSON
	json.Unmarshal(b, &c)
	return c
}

func TestLaterRoundTrip(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	fill := func(n int) []byte {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte(r.Uint32())
		}
		b[0] %= 16 // a name: up to 15 ASCII letters (JSON strings are UTF-8)
		for i := 1; i < 16; i++ {
			b[i] &= 0x7F
		}
		return b
	}
	for range 50 {
		LaterRoundTrip[Record422](t, LaterFiles{Record: fill(int(Layout422)), ItemSize: ItemSize})
		LaterRoundTrip[Record439](t, LaterFiles{Record: fill(int(Layout439)), ItemSize: 67})
		LaterRoundTrip[Record510](t, LaterFiles{Record: fill(int(Layout510)), ItemSize: ItemSize})
	}
}

// A changed modelled field reaches the record; the rest stays.
func TestLaterWritesSheet(t *testing.T) {
	var r Record510
	r.PName[0], r.PName[1] = 1, 'A'
	r.HP, r.Exp, r.SpellSlots[18+8], r.Coins[0] = 5, 100, 2, 7
	b, _ := PackLater[Record510](LaterFiles{Record: Encode(r), ItemSize: ItemSize})
	var c rules1e.CharacterJSON
	json.Unmarshal(b, &c)
	if c.Sheet.SpellSlots[rules1e.SpellClassMage][8] != 2 || c.Sheet.Coins[4] != 7 || c.Sheet.CharName != "A" {
		t.Fatalf("sheet %+v", c.Sheet)
	}
	c.Sheet.HP, c.Sheet.CharName, c.CanTrain = 9, "Bo", true
	b, _ = json.Marshal(c)
	f, err := UnpackLater[Record510](b, ItemSize)
	if err != nil {
		t.Fatal(err)
	}
	g := Decode[Record510](f.Record)
	if g.HP != 9 || g.Exp != 100 || g.CanTrain != 1 || string(g.PName[:3]) != "\x02Bo" || g.SpellSlots[26] != 2 {
		t.Fatalf("record %+v", g)
	}
}
