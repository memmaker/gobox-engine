package script

import (
	"strings"
	"testing"
)

func TestLayout(t *testing.T) {
	l, err := ParseLayout(strings.NewReader(`# a test layout
6B17 21 byte member Memorized
6DD2 1  byte game   RestEvery
6DD3 1  byte game   RestChance
C04B 1  byte fact   PartyX readonly
`))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		a     int
		name  string
		index int
	}{{0x6B17, "Memorized", 0}, {0x6B2B, "Memorized", 20}, {0x6DD3, "RestChance", 0}, {0xC04B, "PartyX", 0}} {
		r, err := l.Resolve(c.a)
		if err != nil || l.Field(r).Name != c.name || r.Index != c.index {
			t.Errorf("%04X: %v %+v", c.a, err, r)
		}
	}
	if !l.Field(must(l.Resolve(0xC04B))).ReadOnly {
		t.Error("PartyX not read-only")
	}
	for _, a := range []int{0x6B16, 0x6B2C, 0x6DD4, 0} {
		if _, err := l.Resolve(a); err == nil {
			t.Errorf("%04X resolved", a)
		}
	}
	for _, bad := range []string{"6B17 2 byte member A\n6B18 1 byte member B\n", "6B17 1 byte member A\n6C00 1 byte game A\n",
		"6B17 1 long member A\n", "6B17 1 byte party A\n", "zz 1 byte game A\n", "6B17 0 byte game A\n", "6B17 1 byte game A ro\n"} {
		if _, err := ParseLayout(strings.NewReader(bad)); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
}

func must(r Ref, err error) Ref {
	if err != nil {
		panic(err)
	}
	return r
}

func TestLayoutWith(t *testing.T) {
	g, _ := ParseLayout(strings.NewReader("6DD2 1 byte game RestEvery\nC018 1 byte call LookAhead\n"))
	a, _ := ParseLayout(strings.NewReader("4A00 4 byte local Seen\n"))
	m, err := g.With(a)
	if err != nil || len(m.Fields) != 3 || m.Fields[0].Name != "Seen" {
		t.Fatalf("%v %+v", err, m)
	}
	for _, bad := range []string{"6DD2 1 byte local X\n", "4A00 1 byte local RestEvery\n"} {
		b, _ := ParseLayout(strings.NewReader(bad))
		if _, err := g.With(b); err == nil {
			t.Errorf("%q merged", bad)
		}
	}
}
