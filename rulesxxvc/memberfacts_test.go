package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// TestMemberFacts: each fact a script's member field is bound to reads what was written (BR ecl.go charVar/setChar).
func TestMemberFacts(t *testing.T) {
	r := New()
	id := r.Add(Character{Name: "Buck", Sex: rules.Female, Career: xxvc.Warrior, HP: 5, MaxHP: 12, Enabled: true, Status: StOkay,
		Side: rules.SideParty})
	for _, c := range []struct {
		name string
		f    rules.Fact
		set  rules.Value // written, then read back
		want rules.Value
	}{
		{"name", rules.FactName, rules.Text("Rogers"), rules.Text("Rogers")},
		{"hp", rules.FactHP, rules.Number(12), rules.Number(12)},
		{"hp max", rules.FactMaxHP, rules.Number(20), rules.Number(20)},
		{"companion", rules.FactCompanion, rules.Yes(true), rules.Yes(true)},
		{"companion off", rules.FactCompanion, rules.Yes(false), rules.Yes(false)},
		{"side auto", rules.FactSide, rules.SideAuto, rules.SideAuto},
		{"side hostile", rules.FactSide, rules.SideMonsters, rules.SideMonsters},
		{"flag 0", rules.FlagFact{Slot: 0}, rules.Number(7), rules.Number(7)},
		{"flag 2", rules.FlagFact{Slot: 2}, rules.Number(1), rules.Number(1)},
		{"dead", rules.FactStatus, rules.StatusDead, rules.StatusDead},
		{"dying", rules.FactStatus, rules.StatusDying, rules.StatusDying},
		{"okay", rules.FactStatus, rules.StatusOkay, rules.StatusOkay},
	} {
		if err := r.SetFact(id, c.f, c.set); err != nil {
			t.Errorf("%s: set: %v", c.name, err)
			continue
		}
		if got, err := r.Fact(id, c.f); err != nil || got != c.want {
			t.Errorf("%s: read %v, %v; want %v", c.name, got, err, c.want)
		}
	}
	if v, _ := r.Fact(id, rules.FactSex); v != rules.Female {
		t.Errorf("sex %v", v)
	}
	if v, _ := r.Fact(id, rules.FactClasses); v != (rules.ClassSet{xxvc.Warrior}) {
		t.Errorf("classes %v", v)
	}
	if v, _ := r.Fact(id, rules.FlagFact{Slot: 1}); v != rules.Number(0) {
		t.Errorf("flag 1 %v: another slot moved", v)
	}
	if err := r.SetFact(id, rules.FlagFact{Slot: ScriptFlagSlots}, rules.Number(1)); err == nil {
		t.Error("a flag past the last slot was set")
	}
	if _, err := r.Fact(id, rules.FlagFact{Slot: -1}); err == nil {
		t.Error("flag -1 read")
	}
}

// TestStatusFact: a status write keeps isEnabled beside it (setChar r[213] = 1 only for StOkay), and the
// statuses with no rules name read as the rules' nearest (coreStatus).
func TestStatusFact(t *testing.T) {
	r := New()
	id := r.Add(Character{Status: StDead})
	for _, c := range []struct {
		st      rules.Status
		enabled bool
	}{{rules.StatusOkay, true}, {rules.StatusDead, false}, {rules.StatusUnconscious, false}, {rules.StatusOkay, true}} {
		if err := r.SetFact(id, rules.FactStatus, c.st); err != nil {
			t.Fatal(err)
		}
		if got := r.chars[id].Enabled; got != c.enabled {
			t.Errorf("%s: enabled %v", c.st, got)
		}
	}
	if err := r.SetFact(id, rules.FactStatus, rules.StatusStoned); err == nil {
		t.Error("stoned has no BR status and was set")
	}
	r.chars[id].Status = StComatose
	if v, _ := r.Fact(id, rules.FactStatus); v != rules.StatusUnconscious {
		t.Errorf("comatose reads %v", v)
	}
}

// TestMemberFactsSurvive: sex, companion and the flags go through the character file and the BR import.
func TestMemberFactsSurvive(t *testing.T) {
	r := New()
	c, err := ConvertBRCharacter(xxvc.Tables{}, BRSave{Gender: 1, ScriptFlags: [ScriptFlagSlots]int{3, 0, 1}})
	if err != nil {
		t.Fatal(err)
	}
	id := r.Add(c)
	file, err := r.Export(id)
	if err != nil {
		t.Fatal(err)
	}
	q := New()
	got, err := q.Import(file)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := q.Fact(got, rules.FactSex); v != rules.Female {
		t.Errorf("sex %v", v)
	}
	if v, _ := q.Fact(got, rules.FlagFact{Slot: 2}); v != rules.Number(1) {
		t.Errorf("flag 2 %v", v)
	}
	if _, err := ConvertBRCharacter(xxvc.Tables{}, BRSave{Gender: 2}); err == nil {
		t.Error("gender 2 converted")
	}
}
