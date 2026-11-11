package script

import (
	"strings"
	"testing"
	"testing/fstest"
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
	r, _ := ParseLayout(strings.NewReader("6DD2 1 byte area RiverState replaces\n"))
	if m, err := g.With(r); err != nil || len(m.Fields) != 2 || m.Field(must(m.Resolve(0x6DD2))).Name != "RiverState" {
		t.Errorf("replace: %v %+v", err, m)
	}
	for _, bad := range []string{"6DD2 1 byte local X\n", "4A00 1 byte local RestEvery\n", "6DD3 1 byte area Y replaces\n",
		"6DD2 2 byte area Y replaces\n"} {
		b, _ := ParseLayout(strings.NewReader(bad))
		if _, err := g.With(b); err == nil {
			t.Errorf("%q merged", bad)
		}
	}
}

// ResolveRefs gives each variable operand its Ref, a table read its const
// field, and reports every address no field covers.
func TestResolveRefs(t *testing.T) {
	l, _ := ParseLayout(strings.NewReader("6DD2 1 byte game RestEvery\n"))
	s := &Script{Area: 3, Data: make([]byte, 8), Tables: map[string]int{"Odds": V1.Base + 4}, Code: []Instr{
		{Op: OpSave, Args: []Arg{{N: 1}, {Type: ArgVar, N: 0x6DD2}}},
		{Op: OpGetTable, Args: []Arg{{Type: ArgVar, N: V1.Base + 4}, {N: 2}, {Type: ArgVar, N: 0x6DD2}}},
		{Op: OpGoto, Args: []Arg{{Type: ArgVar, N: 0, Label: true}}},
	}}
	full, err := l.With(ConstLayout(s, V1))
	if err != nil {
		t.Fatal(err)
	}
	if err := ResolveRefs(s, full); err != nil {
		t.Fatal(err)
	}
	if r := s.Code[0].Args[1].Ref; r == (Ref{}) || full.Field(r).Name != "RestEvery" {
		t.Errorf("save: %+v", r)
	}
	if r := s.Code[1].Args[0].Ref; r == (Ref{}) || full.Field(r).Name != "Odds" || !full.Field(r).ReadOnly {
		t.Errorf("table: %+v", r)
	}
	if s.Code[2].Args[0].Ref != (Ref{}) || s.Code[0].Args[0].Ref != (Ref{}) {
		t.Error("a label or an immediate got a Ref")
	}
	s.Code = append(s.Code, Instr{Op: OpSave, Args: []Arg{{N: 1}, {Type: ArgVar, N: 0x6DD3}}}, Instr{Op: OpSave, Args: []Arg{{N: 1}, {Type: ArgVar, N: 0x6DD4}}})
	err = ResolveRefs(s, full)
	if err == nil || !strings.Contains(err.Error(), "6DD3") || !strings.Contains(err.Error(), "6DD4") {
		t.Errorf("unresolved: %v", err)
	}
}

// TestRoutineSpace: a CALL routine number may equal a memory address (POD's
// CALL 1 and its bank-0 cell 1); each resolves in its own space, and fields
// still may not overlap within one.
func TestRoutineSpace(t *testing.T) {
	l, err := ParseLayout(strings.NewReader("0001 1 byte game MonsterHitMod\n0001 1 byte call CallUpdateDisplay\n"))
	if err != nil {
		t.Fatal(err)
	}
	if l.Field(must(l.Resolve(1))).Name != "MonsterHitMod" || l.Field(must(l.ResolveCall(1))).Name != "CallUpdateDisplay" {
		t.Fatal("spaces mixed")
	}
	if _, err := l.ResolveCall(2); err == nil {
		t.Fatal("routine 2 resolved")
	}
	if _, err := ParseLayout(strings.NewReader("0001 2 byte call CallA\n0002 1 byte call CallB\n")); err == nil {
		t.Fatal("overlapping routines accepted")
	}
}

// TestPartialLayout: layout/partial.txt reads as a Partial layout; an
// operand it does not cover stays a raw address (no Ref, no error) while a
// covered one resolves, and Partial survives With.
func TestPartialLayout(t *testing.T) {
	data := fstest.MapFS{PartialLayout: {Data: []byte("7EE6 1 byte game RevivesParty\n")}}
	l, err := ReadLayout(data, 1)
	if err != nil || l == nil || !l.Partial {
		t.Fatalf("ReadLayout: %+v, %v", l, err)
	}
	s := &Script{Area: 1, Data: make([]byte, 8), Code: []Instr{
		{Op: OpSave, Args: []Arg{{N: 1}, {Type: ArgVar, N: 0x7EE6}}},
		{Op: OpSave, Args: []Arg{{N: 1}, {Type: ArgVar, N: 0x4BE6}}},
		{Op: OpCall, Args: []Arg{{Type: ArgVar, N: 0x1234}}},
	}}
	full, err := l.With(ConstLayout(s, V1))
	if err != nil || !full.Partial {
		t.Fatalf("With: %+v, %v", full, err)
	}
	if err := ResolveRefs(s, full); err != nil {
		t.Fatal(err)
	}
	if r := s.Code[0].Args[1].Ref; r == (Ref{}) || full.Field(r).Name != "RevivesParty" {
		t.Errorf("covered: %+v", r)
	}
	if s.Code[1].Args[1].Ref != (Ref{}) || s.Code[2].Args[0].Ref != (Ref{}) {
		t.Error("an uncovered operand got a Ref")
	}
}
