package pool

import (
	"bufio"
	"fmt"
	"gobox/char"
	"os"
	"rules1e"
	"strings"
	"testing"
)

// TestTreasureBlockAgainstPascal compares addTreasureBlock with EclVm.EclTreasure's
// item block (rules1e testdata/por-treasure.txt, por-re vectres): the records of a
// block, named by number, onto an empty list or one holding item 0.
func TestTreasureBlockAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-treasure.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := testEngine(t)
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "treasure ") {
			continue
		}
		key, val, _ := strings.Cut(line, " =")
		k := atois(t, strings.Fields(key)[1:])
		e.state().Treasure = nil
		if k[1] == 1 {
			e.state().Treasure = []*rules1e.Item{namedItem("0")}
		}
		var block []*rules1e.Item
		for i := 1; i <= k[0]; i++ {
			block = append(block, namedItem(string(rune('0'+i))))
		}
		e.state().AddTreasureBlock(values(block))
		var got []string
		for _, it := range e.state().Treasure {
			got = append(got, string(it.PName[1:1+it.PName[0]]))
		}
		if g, w := strings.Join(got, " "), strings.TrimSpace(val); g != w {
			t.Errorf("%s: Go %q", line, g)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// TestRandomTreasureAgainstPascal compares addRandomTreasure with EclTreasure's
// random items on a shared seed (por-treasure.txt "random" lines): each item's
// record bytes from the type to the end, head first, and the RNG state after.
func TestRandomTreasureAgainstPascal(t *testing.T) {
	f, err := os.Open("../../rules1e/testdata/por-treasure.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e := testEngine(t)
	sc := bufio.NewScanner(f)
	cases := 0
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "random ") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		e.state().Treasure = nil
		e.state().RandSeed = uint32(int32(k[0]))
		e.state().AddRandomTreasure(k[1])
		var got []string
		for _, it := range e.state().Treasure {
			rec, err := e.state().GameRecords([]rules1e.Item{*it})
			if err != nil {
				t.Fatal(err)
			}
			b := char.WriteItems(rec)[0x2E:rules1e.ItemRecSize]
			var parts []string
			for _, x := range b {
				parts = append(parts, fmt.Sprint(int8(x)))
			}
			got = append(got, strings.Join(parts, "."))
		}
		got = append(got, fmt.Sprint(int32(e.state().RandSeed)))
		if g := strings.Join(got, " "); g != val {
			t.Errorf("%s:\n Go %s", line, g)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// TestRandomItemAgainstPascal compares RandomItem with Money.GenerateRandomItem for the javelin, potion
// and wand kinds on 40 seeds (por-treasure.txt "genitem" lines), whose rolls reached leftover item
// data in the original: the record bytes from the type to the end, and the RNG state after.
func TestRandomItemAgainstPascal(t *testing.T) {
	data, err := os.ReadFile("../../rules1e/testdata/por-treasure.txt")
	if err != nil {
		t.Fatal(err)
	}
	e := testEngine(t)
	cases := 0
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "genitem ") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		k := atois(t, strings.Fields(key)[1:])
		e.state().RandSeed = uint32(int32(k[1]))
		it := e.state().RandomItem(rules1e.ItemType(k[0]))
		rec, err := e.state().GameRecords([]rules1e.Item{*it})
		if err != nil {
			t.Fatal(err)
		}
		var parts []string
		for _, x := range char.WriteItems(rec)[0x2E:rules1e.ItemRecSize] {
			parts = append(parts, fmt.Sprint(int8(x)))
		}
		if g := strings.Join(parts, ".") + " " + fmt.Sprint(int32(e.state().RandSeed)); g != val {
			t.Errorf("%s:\n Go %s", line, g)
		}
		cases++
	}
	if cases == 0 {
		t.Fatal("no cases")
	}
}

// TestFixItemRecordAgainstPascal compares the broken records addTreasureBlock mends (fixItemRecord) with
// CharHlp.FixItemRecord in EclTreasure (por-treasure.txt "fixblock" lines): the Flame Tongue, the Ring of
// Feather Falling, Bracers AC 2, the Fine Composite Long Bow, a Ring of
// Protection and a Cloak of Displacement with no save bonus, and a sound Flame Tongue left alone.
func TestFixItemRecordAgainstPascal(t *testing.T) {
	data, err := os.ReadFile("../../rules1e/testdata/por-treasure.txt")
	if err != nil {
		t.Fatal(err)
	}
	e := testEngine(t)
	e.state().Treasure = nil
	block := []*rules1e.Item{namedItem("1"), namedItem("2"), namedItem("3"), namedItem("4"), namedItem("5"), namedItem("6"), namedItem("7")}
	block[0].Type, block[0].Words[0], block[0].Spells[1] = rules1e.ItemLongSword, 0xF1, 0x16
	block[1].Type, block[1].Words[0] = rules1e.ItemFeatherFallRing, 0xEE
	block[2].Type, block[2].Words[1], block[2].Plus = rules1e.ItemBracers, 0xDC, 6
	*block[3] = *block[0]
	block[3].PName = namedItem("4").PName
	block[3].Spells[2] = 0x80
	block[4].Type, block[4].Words[1] = rules1e.ItemHeavyCrossbow, rules1e.WordID(rules1e.ItemCompositeLongBow)
	block[5].Type, block[5].Plus = rules1e.ItemProtectionRing, 3
	block[6].Type, block[6].Plus = rules1e.ItemCloakDisplacement, 2
	e.state().AddTreasureBlock(values(block))
	got := map[string]string{}
	for _, it := range e.state().Treasure {
		rec, err := e.state().GameRecords([]rules1e.Item{*it})
		if err != nil {
			t.Fatal(err)
		}
		var parts []string
		for _, x := range char.WriteItems(rec)[0x2E:rules1e.ItemRecSize] {
			parts = append(parts, fmt.Sprint(int8(x)))
		}
		got[string(it.PName[1:1+it.PName[0]])] = strings.Join(parts, ".")
	}
	cases := 0
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "fixblock ") {
			continue
		}
		key, val, _ := strings.Cut(line, " = ")
		if g := got[strings.Fields(key)[1]]; g != val {
			t.Errorf("%s:\n Go %s", line, g)
		}
		cases++
	}
	if cases != 7 {
		t.Fatalf("%d cases", cases)
	}
}

// namedItem is an empty item record named n, as vectres's Str(k, name).
func namedItem(n string) *rules1e.Item {
	it := &rules1e.Item{}
	it.PName[0] = rules1e.Letter(len(n))
	copy(it.PName[1:], []rules1e.Letter(n))
	return it
}

// values is the items of ps (a block's records, as the data import has them).
func values(ps []*rules1e.Item) []rules1e.Item {
	var out []rules1e.Item
	for _, p := range ps {
		out = append(out, *p)
	}
	return out
}
