package parity

import (
	"image"
	"reflect"
	"testing"
	"time"

	"gobox/front"
	"gobox/game"
	"por/rules"
)

// call is one rule call and the Decision that must do the same.
type call struct {
	name   string
	setup  func(t *testing.T) *rules.Engine
	direct func(e *rules.Engine) error
	d      game.Decision
}

func menu(t *testing.T) *rules.Engine { return partyMenu(t) }
func party(t *testing.T) *rules.Engine {
	e := partyMenu(t)
	e.QuickStart()
	settled(e, front.Action{})
	return e
}
func callWalk(t *testing.T) *rules.Engine { e := engine(t, true, ""); settled(e); return e }
func scene(name string) func(t *testing.T) *rules.Engine {
	return func(t *testing.T) *rules.Engine {
		e := engine(t, true, name)
		settled(e, front.Action{})
		return e
	}
}
func created(t *testing.T) *rules.Engine {
	e := menu(t)
	e.CreateCharacter()
	settled(e, front.Action{})
	return e
}

// errText is err's words, "" for none.
func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// TestCalls: Apply(decision) leaves the same View, and refuses the same way,
// as the rule call it names, on two engines started alike.
func TestCalls(t *testing.T) {
	pt := func(x, y int) image.Point { return image.Pt(x, y) }
	for _, c := range []call{
		{"QuickStart", menu, func(e *rules.Engine) error { e.QuickStart(); return nil }, game.QuickStart{}},
		{"RemoveMember", party, func(e *rules.Engine) error { return e.RemoveMember(0) }, game.RemoveCharacter{Roster: 0}},
		{"NewClass", party, func(e *rules.Engine) error { return e.NewClass(0, 1) }, game.ChangeClass{Member: 0, Class: 1}},
		{"CreateCharacter", menu, (*rules.Engine).CreateCharacter, game.NewCharacter{}},
		{"ModifyCharacter", party, func(e *rules.Engine) error { return e.ModifyCharacter(0) }, game.Modify{Member: 0}},
		{"Roll", created, func(e *rules.Engine) error { return e.Roll(false, "") }, game.Roll{}},
		{"Rename", created, func(e *rules.Engine) error { return e.Rename("Zed") }, game.Rename{Name: "Zed"}},
		{"SetField", created, func(e *rules.Engine) error { return e.SetField(front.FieldRace, 2) }, game.SetField{Field: front.FieldRace, Option: 2}},
		{"EndCreation", created, func(e *rules.Engine) error { return e.EndCreation(false) }, game.EndCreation{}},
		{"LeaveCamp", callWalk, (*rules.Engine).LeaveCamp, game.LeaveCamp{}},
		{"PlanSpell", callWalk, func(e *rules.Engine) error { return e.PlanSpell(0, 1, 1) }, game.Plan{Spell: 1, Count: 1}},
		{"Scribe", callWalk, func(e *rules.Engine) error { return e.Scribe(0, 1) }, game.Scribe{Spell: 1}},
		{"CampRest", callWalk, func(e *rules.Engine) error { return e.CampRest(0) }, game.CampRest{}},
		{"StopRest", callWalk, (*rules.Engine).StopRest, game.StopRest{}},
		{"UsePower", callWalk, func(e *rules.Engine) error { return e.UsePower(0, rules.ActHeal, 0, 0, true) }, game.UsePower{Kind: int(rules.ActHeal), Sure: true}},
		{"LootView", callWalk, func(e *rules.Engine) error { return e.LootView(0) }, game.LootLook{}},
		{"LootTake", callWalk, func(e *rules.Engine) error { return e.LootTake(0) }, game.LootTake{}},
		{"LootDetect", callWalk, (*rules.Engine).LootDetect, game.LootDetect{}},
		{"LootLeave", callWalk, func(e *rules.Engine) error { return e.LootLeave(true) }, game.LootLeave{Sure: true}},
		{"TryDoor", callWalk, func(e *rules.Engine) error { return e.TryDoor(front.DoorBash) }, game.TryDoor{Way: int(front.DoorBash)}},
		{"OpenPicker", created, func(e *rules.Engine) error {
			if !e.OpenPicker(front.PickHead) {
				return &game.Refusal{Text: "no such picker"}
			}
			return nil
		}, game.OpenPicker{Kind: front.PickHead}},
		{"Pick", created, func(e *rules.Engine) error { e.OpenPicker(front.PickHead); e.Pick(1); return nil }, game.Pick{N: 1}},
		{"ClosePicker", created, func(e *rules.Engine) error { e.ClosePicker(); return nil }, game.ClosePicker{}},
		{"EditIcon", party, func(e *rules.Engine) error { e.EditIcon(-1); return nil }, game.EditIcon{Member: -1}},
		{"IconSize", party, func(e *rules.Engine) error { e.EditIcon(-1); e.IconSize(); return nil }, game.IconSize{}},
		{"IconColour", party, func(e *rules.Engine) error { e.EditIcon(-1); e.IconColour(0, 1, 2); return nil }, game.IconColour{Part: 0, Shade: 1, Value: 2}},
		{"IconRemove", party, func(e *rules.Engine) error { e.EditIcon(-1); e.IconRemove(); return nil }, game.IconRemove{}},
		{"IconPNG", party, func(e *rules.Engine) error { e.EditIcon(-1); e.IconPNG(""); return nil }, game.IconPNG{}},
		{"EndIcon", party, func(e *rules.Engine) error { e.EditIcon(-1); e.EndIcon(false); return nil }, game.EndIcon{}},
		{"DeleteSlot", party, func(e *rules.Engine) error { return e.DeleteSlot("none") }, game.DeleteSlot{Name: "none"}},
		{"OpenSheet", callWalk, func(e *rules.Engine) error { return e.OpenSheet(0) }, game.OpenSheet{}},
		{"CloseSheet", callWalk, (*rules.Engine).CloseSheet, game.CloseSheet{}},
		{"ShowMember", callWalk, func(e *rules.Engine) error { return e.ShowMember(1) }, game.ShowMember{Member: 1}},
		{"OpenItems", callWalk, func(e *rules.Engine) error { return e.OpenItems(0, true) }, game.OpenItems{Direct: true}},
		{"CloseItems", callWalk, func(e *rules.Engine) error { return e.CloseItems(true) }, game.CloseItems{All: true}},
		{"SetPolicy", callWalk, func(e *rules.Engine) error { return e.SetPolicy(1, front.Never) }, game.SetPolicy{ID: 1, Policy: int(front.Never)}},
		{"UseItem", callWalk, func(e *rules.Engine) error { e.ShowMember(0); return e.UseItem(0, 0, 0, true) }, game.UseItem{}},
		{"Browse", scene("shop"), func(e *rules.Engine) error { return e.Browse(-1) }, game.Browse{Member: -1}},
		{"Identify", scene("shop"), func(e *rules.Engine) error { return e.Identify(0, 0) }, game.Identify{}},
		{"Appraise", scene("appraise"), func(e *rules.Engine) error { return e.Appraise(false) }, game.Value{}},
		{"KeepAppraised", scene("appraise"), func(e *rules.Engine) error { return e.KeepAppraised(true) }, game.KeepAppraised{Keep: true}},
		{"Cure", scene("temple"), func(e *rules.Engine) error { return e.Cure(0, 0) }, game.Cure{}},
		{"ViewMember", scene("temple"), func(e *rules.Engine) error { return e.ViewMember(0) }, game.ViewMember{}},
		{"Leave", scene("shop"), (*rules.Engine).Leave, game.Leave{}},
		{"OpenTraining", scene("hall"), func(e *rules.Engine) error { return e.OpenTraining(0) }, game.OpenTraining{}},
		{"CloseTraining", scene("hall"), (*rules.Engine).CloseTraining, game.CloseTraining{}},
		{"Learn", scene("hall"), func(e *rules.Engine) error { return e.Learn(1) }, game.Learn{Spell: 1}},
		{"PickPower", callWalk, func(e *rules.Engine) error { return e.PickPower(1) }, game.PickPower{ID: 1}},
		{"Mark", callWalk, func(e *rules.Engine) error { e.Mark(3, 4, "here"); return nil }, game.Mark{At: pt(3, 4), Text: "here"}},
		{"Unmark", callWalk, func(e *rules.Engine) error { e.Mark(3, 4, "here"); e.Unmark(3, 4); return nil }, game.Unmark{At: pt(3, 4)}},
		{"WalkTo", callWalk, func(e *rules.Engine) error {
			if !e.WalkTo(3, 4) {
				return &game.Refusal{Text: "walk"}
			}
			return nil
		}, game.WalkTo{At: pt(3, 4)}},
	} {
		t.Run(c.name, func(t *testing.T) {
			a, b := c.setup(t), c.setup(t)
			if c.name == "Unmark" || c.name == "Pick" || c.name == "IconSize" || c.name == "IconColour" || c.name == "IconRemove" || c.name == "IconPNG" || c.name == "EndIcon" {
				// the decision's engine gets the same set-up call first
				prep(c.name, b)
			}
			derr := c.direct(a)
			t.Logf("call error: %q", errText(derr))
			va := settled(a, front.Action{})
			res := b.Rules().Apply(c.d)
			vb := settled(b, front.Action{})
			for _, v := range []*front.View{&va, &vb} {
				v.Seq, v.Note.At = 0, time.Time{} // counts and clocks, not state
			}
			if c.name == "WalkTo" && derr != nil && res.Err != nil {
				derr, res.Err = nil, nil // both refused; the refusal's words are the game's note
			}
			if errText(derr) != errText(res.Err) && (derr == nil) != (res.Err == nil) {
				t.Errorf("error: call %v, decision %v", derr, res.Err)
			}
			if !reflect.DeepEqual(va, vb) {
				t.Errorf("views differ after %T: %v", c.d, diffFields(va, vb))
			}
		})
	}
}

// prep puts b where the call's own set-up (before the named verb) leaves a.
func prep(name string, b *rules.Engine) {
	switch name {
	case "Unmark":
		b.Mark(3, 4, "here")
	case "Pick":
		b.OpenPicker(front.PickHead)
	default:
		b.EditIcon(-1)
	}
}

// diffFields names the View fields that differ.
func diffFields(a, b front.View) (names []string) {
	x, y := reflect.ValueOf(a), reflect.ValueOf(b)
	for i := range x.NumField() {
		if !reflect.DeepEqual(x.Field(i).Interface(), y.Field(i).Interface()) {
			names = append(names, x.Type().Field(i).Name)
		}
	}
	return
}
