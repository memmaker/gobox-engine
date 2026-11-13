//go:build devviews

package engine

import (
	"errors"
	"gobox/game"
	"gobox/rules1e"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// The creation's rules: a grey pick forces the dependants to their first
// valid value.
func TestKindEnforce(t *testing.T) {
	k := rules1e.Kind{Race: rules1e.RaceElf, Sex: 0, Class: rules1e.ClsFighterThief, Align: 7}
	if k = (rules1e.Kind{Race: rules1e.RaceElf, Sex: 0, Class: rules1e.ClsPaladin, Align: 7}).Enforce(rules1e.FClass); k.Race != rules1e.RaceHuman || k.Align != 0 {
		t.Errorf("elf paladin: %+v, want a lawful good human", k)
	}
	if k = (rules1e.Kind{Race: rules1e.RaceHuman, Sex: 0, Class: rules1e.ClsPaladin, Align: 8}).Enforce(rules1e.FAlign); k.Class != rules1e.ClsCleric || k.Align != 8 {
		t.Errorf("chaotic evil paladin: %+v, want a cleric", k)
	}
	if k = (rules1e.Kind{Race: rules1e.RaceDwarf, Sex: 0, Class: rules1e.ClsMage, Align: 3}).Enforce(rules1e.FRace); k.Class != rules1e.ClsFighter || k.Align != 3 {
		t.Errorf("dwarf magic-user: %+v, want a neutral good fighter", k)
	}
	if k = (rules1e.Kind{Race: rules1e.RaceDwarf, Sex: 1, Class: rules1e.ClsFighter, Align: 3}).Enforce(rules1e.FSex); k != (rules1e.Kind{Race: rules1e.RaceDwarf, Sex: 1, Class: rules1e.ClsFighter, Align: 3}) {
		t.Errorf("gender: %+v changed", k)
	}
	if cl := rules1e.PickClasses(); slices.Contains(cl, rules1e.ClsDruid) || slices.Contains(cl, rules1e.ClsMonk) || !slices.Contains(cl, rules1e.ClsClericRanger) {
		t.Errorf("classes offered: %v", cl)
	}
}

// The creation sheet takes decisions: a race that does not fit the class
// forces it, a stat typed, the head picker paged, dropped PNGs, Keep
// saving the character under its default name; then Add, Modify, Forget,
// the icon editor and the party menu.
func TestCreateSheet(t *testing.T) {
	e := testEngine(t)
	e.viewBit0 = 1 // the sheet's portrait, as init sets it
	draft := func() *game.Draft { a, _ := e.build().(game.Assembling); return a.Draft }
	cm := func() game.Common { return commonOf(e.build()) }
	option := func(f int, text string) int {
		i := slices.Index(draft().Fields[f].Options, text)
		if i < 0 {
			t.Fatalf("no %q in field %d", text, f)
		}
		return i
	}
	e.createCharacter()
	if d := draft(); d == nil || len(d.Fields) != rules1e.FieldClass+1 {
		t.Fatalf("creation %+v", d)
	}
	c := e.member(e.selected())
	if c.Name() != "No-One" || c.Race != rules1e.RaceHuman || c.Class != rules1e.ClsFighter || c.Stats[0] == 0 || c.MaxHP == 0 {
		t.Fatalf("sample: %s %s %d stats %v hp %d", c.Name(), c.Race, c.Class, c.Stats, c.MaxHP)
	}
	e.SetField(rules1e.FieldClass, option(rules1e.FieldClass, rules1e.ClassNames[rules1e.ClsMage]))
	if c = e.member(e.selected()); c.Class != rules1e.ClsMage {
		t.Fatalf("class %d, want a magic-user", c.Class)
	}
	// Dwarf does not fit a magic-user: it forces the class to fighter.
	dw := option(rules1e.FieldRace, rules1e.RaceName(rules1e.RaceDwarf))
	if draft().Fields[rules1e.FieldRace].Fits[dw] {
		t.Error("a dwarf magic-user fits")
	}
	e.SetField(rules1e.FieldRace, dw)
	if c = e.member(e.selected()); c.Race != rules1e.RaceDwarf || c.Class != rules1e.ClsFighter || c.IconSize != 1 {
		t.Fatalf("race %s class %d size %d, want a dwarf fighter", c.Race, c.Class, c.IconSize)
	}
	e.setStat(rules1e.StatCha, "9")
	if c = e.member(e.selected()); c.Stats[rules1e.StatCha] != 9 {
		t.Errorf("charisma %d, want 9", c.Stats[rules1e.StatCha])
	}
	var ref *game.Refusal // typed text that is no number is refused, never kept silently
	if err := e.setStat(rules1e.StatCha, "x"); !errors.As(err, &ref) || ref.Text != "not a number" {
		t.Errorf("SetStat x: %v, want the refusal not a number", err)
	}
	e.rename("Brok")
	e.roll(true, "3 3 3 3 3 3")
	if c = e.member(e.selected()); c.Name() != "Brok" || !e.state().MeetsMinStats(c) {
		t.Errorf("name %q, minimums %v", c.Name(), e.state().MinStats)
	}
	if d := draft().Member; d.AC == 0 || d.THAC0 == 0 || d.Move == 0 || d.Age == 0 { // BUG-CLASSES 19: a fresh roll's figures
		t.Errorf("fresh roll's draft: AC %d THAC0 %d Move %d Age %d", d.AC, d.THAC0, d.Move, d.Age)
	}
	// A head from the pool (gobox's picker answers SetChoice; here the DAX head 2 directly).
	e.creationEdit(func(cr *creation) error { cr.d.D.Head = e.portPic("HEAD", 2); return nil })
	c = e.member(e.selected())
	// The icon editor: Size, then Exit puts the old icon back.
	if e.EditIcon(-1); cm().IconEditor == nil {
		t.Fatal("no icon editor")
	}
	e.IconSize()
	if e.EndIcon(false); c.IconSize != 1 || draft() == nil {
		t.Fatalf("icon size %d, creation %v", c.IconSize, draft() != nil)
	}
	// Dropped PNGs: one on the icon's cells (IconBox) is the combat icon,
	// one on the portrait box the portrait; the sheet shows both at once.
	dir := t.TempDir()
	write := func(name string, w, h int) string {
		p := filepath.Join(dir, name)
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, image.NewRGBA(image.Rect(0, 0, w, h)))
		f.Close()
		return p
	}
	if ib, pb := gameRect(e.snap.iconBox), gameRect(e.snap.portrait); ib.Empty() || pb.Empty() {
		t.Fatalf("drop targets: icon %+v portrait %+v", ib, pb)
	}
	e.IconPNG(write("icon.png", 24, 24))
	e.portraitPNG(write("face.png", 64, 64))
	if c = e.member(e.selected()); c.Icon == nil || c.Portrait == nil {
		t.Errorf("dropped: icon %v portrait %v, want both on the sheet", c.Icon != nil, c.Portrait != nil)
	}
	if e.EndCreation(true); draft() != nil || e.top() > onCamp {
		t.Fatalf("after Keep: creation %v screens %d", draft() != nil, e.top())
	}
	id, ok := e.LoadCharacter("Brok", false)
	d := e.member(id)
	if !ok || d.Race != rules1e.RaceDwarf || d.PortHead != 2 || d.Icon == nil || d.Portrait == nil {
		t.Errorf("saved: %v race %s head %d icon %v portrait %v", d.Name(), d.Race, d.PortHead, d.Icon != nil, d.Portrait != nil)
	}
	// Add: the saved characters as cards (the portrait PNG the presenter's);
	// a pick adds one, the picker stays until verbClose.
	e.OpenPicker(PickAdd)
	if p := cm().Picker; p == nil || len(p.Pics) != 1 || len(p.Portraits) != 1 || p.Portraits[0] == nil || p.Classes[0] != "Fighter" {
		t.Fatalf("add: picker %+v", p)
	}
	e.Pick(0)
	if len(e.state().Party) != 1 || e.state().Party[0].Name() != "Brok" || cm().Picker.Labels[0] != "* Brok" {
		t.Fatalf("added: party %d", len(e.state().Party))
	}
	e.ClosePicker()
	// Modify: the creation sheet on a member with no experience of its
	// own (Felix's R3 ruling 16), Keep keeps the changes.
	e.state().Party[0].Exp = 500
	if err := e.modifyCharacter(-1); err == nil || draft() != nil {
		t.Fatal("modify: an experienced member was modifiable")
	}
	e.state().Party[0].Exp = 0
	e.modifyCharacter(-1)
	if draft() == nil || e.creation.abandon != "Discard the changes?" {
		t.Fatalf("modify: %+v", draft())
	}
	e.setStat(rules1e.StatCha, "7")
	e.EndCreation(true)
	if m := e.state().Party[0]; m.Stats[rules1e.StatCha] != 7 || m.Share != 1 {
		t.Errorf("modified: charisma %d share %d", m.Stats[rules1e.StatCha], m.Share)
	}
	// The party menu takes decisions: View opens the sheet over it, the
	// sheet's Exit brings the menu back, Begin starts play.
	e.menu.on = true
	if m := e.partyFunctions(); !slices.Contains(m, partyView) || !slices.Contains(m, partyBegin) {
		t.Fatalf("party menu %v", m)
	}
	e.OpenSheet(-1)
	if !e.sheetUp() || e.menuUp() {
		t.Fatalf("view: sheet %v menu %v", e.sheetUp(), e.menuUp())
	}
	if e.CloseSheet(); !e.menuUp() {
		t.Fatal("back from the sheet: no party menu")
	}
	e.begin()
	if e.state().BeginFlag != 4 || e.menu.on {
		t.Errorf("begin flag %d, screens %d", e.state().BeginFlag, e.top())
	}
	// Forget: the picker asks (the presenter's question), a pick deletes.
	e.SaveCharacter(e.charID(e.state().Party[0]), "") // as Remove
	e.state().Party = nil
	e.OpenPicker(PickForget)
	if p := cm().Picker; p == nil || p.Ask == "" {
		t.Fatalf("forget: %+v", p)
	}
	e.Pick(0)
	if len(e.BuildCharList()) != 0 || e.top() > onCamp {
		t.Errorf("forgotten: still there, screens %d", e.top())
	}
}

// Point buy (gobox's): refused outside a draft or with the wrong count;
// on a human fighter the scores bought become the stats, 18 strength
// gets its percentile, the age rules apply as after a roll, and hit points are rolled.
func TestPointBuy(t *testing.T) {
	e := testEngine(t)
	buy := []int{18, 16, 12, 8, 8, 8}
	if e.doNow(game.PointBuy{Scores: buy}); e.creation != nil {
		t.Fatal("a draft appeared")
	}
	if err := e.pointBuy(buy); err == nil {
		t.Error("point buy outside a draft")
	}
	e.createCharacter()
	if err := e.pointBuy(buy[:5]); err == nil {
		t.Error("five scores taken")
	}
	if err := e.pointBuy([]int{18, 18, 18, 8, 8, 8}); err == nil {
		t.Error("over budget taken")
	}
	if err := e.pointBuy(buy); err != nil {
		t.Fatal(err)
	}
	c := e.member(e.selected())
	for i, v := range buy { // the age rules apply as after a roll
		if want := rules1e.AdjustStat(c, i, rules1e.Score(byte(v))); c.Stats[i] != want {
			t.Fatalf("stat %d %d, want %d", i, c.Stats[i], want)
		}
	}
	if (c.Stats[0] == 18) != (c.Str00 != 0) || c.MaxHP == 0 {
		t.Errorf("str00 %d hp %d", c.Str00, c.MaxHP)
	}
}
