package parity

import (
	"image"
	"reflect"
	"testing"
	"time"

	"gobox/front"
	"gobox/game"
	"por/rules"
	"por/rules4e"
)

// call is one rule call and the Decision that must do the same.
type call struct {
	name   string
	setup  func(t *testing.T) *rules.Engine
	direct func(e *rules.Engine) error
	d      game.Decision
}

// lates are the Decisions that depend on the set-up scene (an id it offers),
// made from the engine the decision goes to.
func lastItem(e *rules.Engine) int { return len(e.Rules().State().Party[0].Items) - 1 }

var lates = map[string]func(e *rules.Engine) game.Decision{
	"UseItem": func(e *rules.Engine) game.Decision { return game.UseItem{Item: lastItem(e)} },
	"Learn":   func(e *rules.Engine) game.Decision { return game.Learn{Spell: e.Rules().State().Training.Learn[0].ID} },
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

// raw is a dev scene as it is put up, with no pace run out.
func raw(name string) func(t *testing.T) *rules.Engine {
	return func(t *testing.T) *rules.Engine { return engine(t, true, name) }
}

// then is setup with the calls after it.
func then(setup func(t *testing.T) *rules.Engine, f func(e *rules.Engine)) func(t *testing.T) *rules.Engine {
	return func(t *testing.T) *rules.Engine {
		e := setup(t)
		f(e)
		settled(e, front.Action{})
		return e
	}
}

// fourE is exploring under the 4e rule set: the members owe power picks.
func fourE(t *testing.T) *rules.Engine {
	e := engine(t, true, "")
	e.SetRules(rules4e.New())
	settled(e, front.Action{})
	return e
}

// picked is fourE with every pick taken (the first option) and the mage, who has a reaction, selected.
func picked(t *testing.T) *rules.Engine {
	e := fourE(t)
	for p := e.Rules().State().PowerPick; p != nil; p = e.Rules().State().PowerPick {
		e.PickPower(p.Options[0].ID)
		settled(e, front.Action{})
	}
	e.Apply(rules.Action{Kind: rules.ActLead, N: 4})
	return e
}

func created(t *testing.T) *rules.Engine {
	e := menu(t)
	e.CreateCharacter()
	settled(e, front.Action{})
	return e
}

// run is e's State once published, after the engine's paced steps until it
// stops pacing (a count, not a clock: both engines take the same steps).
func run(e *rules.Engine) game.State {
	e.Apply() // publish what the call left
	for i := 0; i < 20000; i++ {
		if p := e.Rules().State().Pacing; p == nil || p.Pace == game.PaceNone || p.Pace == game.PaceAnim {
			break
		}
		e.Step()
		e.Apply()
	}
	return normal(e)
}

// normal is e's State without the counts and clocks.
func normal(e *rules.Engine) game.State {
	s := e.Rules().State()
	if s.Pacing != nil {
		p := *s.Pacing
		p.Seq, s.Pacing = 0, &p
	}
	s.Note.At = time.Time{}
	return still(s)
}

// errText is err's words, "" for none.
func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// TestCalls: Apply(decision) leaves the same State, and refuses the same way,
// as the rule call it names, on two engines started alike.
func TestCalls(t *testing.T) {
	pt := func(x, y int) image.Point { return image.Pt(x, y) }
	for _, c := range []call{
		{"QuickStart", menu, func(e *rules.Engine) error { e.QuickStart(); return nil }, game.QuickStart{}},
		{"RemoveMember", party, func(e *rules.Engine) error { return e.RemoveMember(0) }, game.RemoveCharacter{Roster: 0}},
		{"NewClass", scene("dual"), func(e *rules.Engine) error { return e.NewClass(0, 3) }, game.ChangeClass{Member: 0, Class: 3}},
		{"CreateCharacter", menu, (*rules.Engine).CreateCharacter, game.NewCharacter{}},
		{"ModifyCharacter", party, func(e *rules.Engine) error { return e.ModifyCharacter(0) }, game.Modify{Member: 0}},
		{"Roll", created, func(e *rules.Engine) error { return e.Roll(false, "") }, game.Roll{}},
		{"Rename", created, func(e *rules.Engine) error { return e.Rename("Zed") }, game.Rename{Name: "Zed"}},
		{"SetField", created, func(e *rules.Engine) error { return e.SetField(front.FieldRace, 2) }, game.SetField{Field: front.FieldRace, Option: 2}},
		{"EndCreation", created, func(e *rules.Engine) error { return e.EndCreation(false) }, game.EndCreation{}},
		{"LeaveCamp", scene("camp"), (*rules.Engine).LeaveCamp, game.LeaveCamp{}},
		{"PlanSpell", scene("camp"), func(e *rules.Engine) error { return e.PlanSpell(0, 1, 1) }, game.Plan{Spell: 1, Count: 1}},
		{"Scribe", scene("camp"), func(e *rules.Engine) error { return e.Scribe(0, 1) }, game.Scribe{Spell: 1}},
		{"CampRest", scene("camp"), func(e *rules.Engine) error { return e.CampRest(0) }, game.CampRest{}},
		{"StopRest", raw("resting"), (*rules.Engine).StopRest, game.StopRest{}},
		{"UsePower", scene("paladin"), func(e *rules.Engine) error { return e.UsePower(0, rules.ActHeal, 0, 0, true) }, game.UsePower{Kind: int(rules.ActHeal), Sure: true}},
		{"LootDetect", scene("lootdetect"), (*rules.Engine).LootDetect, game.LootDetect{}},
		{"LootLeave", scene("loot"), func(e *rules.Engine) error { return e.LootLeave(true) }, game.LootLeave{Sure: true}},
		{"TryDoor", then(scene("lockeddoor"), func(e *rules.Engine) { e.Apply(rules.Action{Kind: rules.ActForward}) }), func(e *rules.Engine) error { return e.TryDoor(rules.DoorBash) }, game.TryDoor{Way: int(rules.DoorBash)}},
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
		{"CloseSheet", scene("sheet"), (*rules.Engine).CloseSheet, game.CloseSheet{}},
		{"ShowMember", scene("sheet"), func(e *rules.Engine) error { return e.ShowMember(1) }, game.ShowMember{Member: 1}},
		{"OpenItems", scene("sheet"), func(e *rules.Engine) error { return e.OpenItems(0, true) }, game.OpenItems{Direct: true}},
		{"CloseItems", scene("items"), func(e *rules.Engine) error { return e.CloseItems(true) }, game.CloseItems{All: true}},
		{"SetPolicy", picked, func(e *rules.Engine) error { return e.SetPolicy(106, rules.Never) }, game.SetPolicy{ID: 106, Policy: int(rules.Never)}},
		{"UseItem", scene("usable"), func(e *rules.Engine) error { e.ShowMember(0); return e.UseItem(lastItem(e), 0, 0, true) }, game.UseItem{}},
		{"Browse", scene("shop"), func(e *rules.Engine) error { return e.Browse(-1) }, game.Browse{Member: -1}},
		{"Identify", scene("shop"), func(e *rules.Engine) error { return e.Identify(0, 0) }, game.Identify{}},
		{"Appraise", scene("appraise"), func(e *rules.Engine) error { return e.Appraise(false) }, game.Value{}},
		{"KeepAppraised", scene("appraise"), func(e *rules.Engine) error { return e.KeepAppraised(true) }, game.KeepAppraised{Keep: true}},
		{"Cure", scene("temple"), func(e *rules.Engine) error { return e.Cure(0, 0) }, game.Cure{}},
		{"ViewMember", scene("temple"), func(e *rules.Engine) error { return e.ViewMember(0) }, game.ViewMember{}},
		{"Leave", scene("shop"), (*rules.Engine).Leave, game.Leave{}},
		{"OpenTraining", scene("hall"), func(e *rules.Engine) error { return e.OpenTraining(0) }, game.OpenTraining{}},
		{"CloseTraining", then(scene("hall"), func(e *rules.Engine) { e.OpenTraining(0) }), (*rules.Engine).CloseTraining, game.CloseTraining{}},
		{"Learn", then(scene("learn"), func(e *rules.Engine) { e.OpenTraining(4); e.Train() }), func(e *rules.Engine) error { return e.Learn(e.Rules().State().Training.Learn[0].ID) }, nil},
		{"PickPower", fourE, func(e *rules.Engine) error { return e.PickPower(100) }, game.PickPower{ID: 100}},
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
			if late := lates[c.name]; late != nil {
				c.d = late(b)
			}
			derr := c.direct(a)
			t.Logf("call error: %q", errText(derr))
			va := run(a)
			res := b.Rules().Apply(c.d)
			vb := run(b)
			if c.name == "WalkTo" && derr != nil && res.Err != nil {
				derr, res.Err = nil, nil // both refused; the refusal's words are the game's note
			}
			if errText(derr) != errText(res.Err) && (derr == nil) != (res.Err == nil) {
				t.Errorf("error: call %v, decision %v", derr, res.Err)
			}
			if !reflect.DeepEqual(va, vb) {
				t.Errorf("states differ after %T: %v\n%+v\n%+v", c.d, diffFields(va, vb), va.Where, vb.Where)
			}
			atKey(t, "TestCalls/"+c.name, groups(vb))
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

// diffFields names the State fields that differ.
func diffFields(a, b game.State) (names []string) {
	x, y := reflect.ValueOf(a), reflect.ValueOf(b)
	for i := range x.NumField() {
		if !reflect.DeepEqual(x.Field(i).Interface(), y.Field(i).Interface()) {
			names = append(names, x.Type().Field(i).Name)
		}
	}
	return
}
