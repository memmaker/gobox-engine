package present

import (
	"go/ast"
	"go/parser"
	"go/token"
	"gobox/front"
	"gobox/game/play"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"gobox/gpu"
	"gobox/ui"

	"gobox/blobber"
	"gobox/input"
)

// Esc closes the command menu (Enter opened it).
func TestEscClosesCommandMenu(t *testing.T) {
	g, _ := playGame(play.Exploring{})
	p := &Presenter{active: true, view: &blobber.View{}}
	p.SetEngine(g)
	v := g.View()
	if !p.key(13, v) || !p.menu {
		t.Fatal("Enter did not open the menu")
	}
	if !p.key(27, v) || p.menu {
		t.Fatal("Esc did not close the menu")
	}
	p.key(13, v)
	p.Mouse(0, 0, false, false, true, 0, v)
	if p.menu {
		t.Fatal("a right click did not close the menu")
	}
}

// menuNav: Esc is back and never picks; a hotkey picks in either case; a
// hotkey letter not in the text is shown in front.
func TestMenuNav(t *testing.T) {
	sel := 0
	keys := []byte{'C', 'E'}
	if i, back, _ := ui.MenuNav(actOf(27), 27, &sel, keys); i != -1 || !back {
		t.Fatal("Esc picked a row")
	}
	if i, _, _ := ui.MenuNav(actOf('e'), 'e', &sel, keys); i != 1 || sel != 1 {
		t.Fatal("hotkey e did not pick row 1")
	}
	if ui.MenuNav(actOf(0x5000), 0x5000, &sel, keys); sel != 0 {
		t.Fatal("Down did not wrap")
	}
	if s, at := ui.HotText("Camp", 'E'); s != "[E] Camp" || at != 1 {
		t.Fatal(s)
	}
	if s, _ := ui.HotText("Cure Blindness  1000g", '1'); s != "[1] Cure Blindness  1000g" { // a digit in a price is not the key
		t.Fatal(s)
	}
}

// digitKeys: rows with no letter and that can be picked get 1-9, skipping
// digits already taken, then none.
func TestDigitKeys(t *testing.T) {
	keys := ui.DigitKeys([]byte{'B', 0, '2', 0, 0}, func(i int) bool { return i != 3 })
	if string(keys) != "B12\x003" {
		t.Fatalf("%q", keys)
	}
	many := ui.DigitKeys(make([]byte, 12), func(int) bool { return true })
	if many[8] != '9' || many[9] != 0 {
		t.Fatalf("%q", many)
	}
}

// editKey: the cursor moves, Backspace and Delete edit at it, typing
// inserts at it; caretLine shows it.
func TestEditCursor(t *testing.T) {
	s, c := "abd", 3
	for _, k := range []uint16{0x4B00, 'c', 0x4700, 0x5300, 'A', 0x4F00, 8} {
		editKey(k, &s, &c, 10, false)
	}
	if s != "Abc" || c != 3 {
		t.Fatalf("%q %d", s, c)
	}
	if editKey(0x4B00, &s, &c, 10, false); c != 2 {
		t.Fatal(c)
	}
	if l := caretLine(s, 1, 4, true); l != "A_c  " {
		t.Fatalf("%q", l)
	}
}

func TestPrice(t *testing.T) {
	for m, want := range map[front.Money]string{{}: "0g", {Gold: 12}: "12g", {Gold: 12, Copper: 5}: "12g 5c", {Copper: 50}: "0g 50c"} {
		if got := price(m); got != want {
			t.Errorf("price(%v) = %q, want %q", m, got, want)
		}
	}
	if got := withGold("Heal", 7); got != "Heal\x03 Party 7g " {
		t.Errorf("withGold = %q", got)
	}
}

func TestMemberKey(t *testing.T) {
	for _, c := range []struct {
		k      uint16
		cur, n int
		lr     bool
		want   int
		wantOK bool
	}{
		{KeyF1, 2, 4, false, 0, true},
		{KeyF1 + 3<<8, 0, 4, false, 3, true},
		{KeyF1 + 5<<8, 0, 4, false, 5, false}, // no sixth member
		{0x4B00, 0, 4, true, 3, true},         // Left wraps
		{0x4D00, 3, 4, true, 0, true},         // Right wraps
		{0x4D00, -1, 4, true, 0, true},        // none selected: the first
		{0x4B00, 1, 4, false, -1, false},      // Left elsewhere is no member key
		{KeyPadPrev, 1, 4, false, 0, true},    // the stick buttons always are
		{KeyPadNext, 1, 4, false, 2, true},
		{9, 1, 4, true, -1, false}, // Tab is never one (minimap, tabs, targets)
	} {
		if got, ok := memberKey(c.k, c.cur, c.n, c.lr); got != c.want || ok != c.wantOK {
			t.Errorf("memberKey(%#x, %d, %d, %v) = %d, %v; want %d, %v", c.k, c.cur, c.n, c.lr, got, ok, c.want, c.wantOK)
		}
	}
}

// The engine's yes/no question: Y, N, Esc, Enter (on No first) answer it
// with ActYes or ActNo; the arrows move to Yes.
func TestConfirmKeys(t *testing.T) {
	v := front.View{Ask: &front.Ask{Kind: front.AskConfirm, Menu: front.Menu{Prompt: "Flee?"}}}
	for _, c := range []struct {
		keys []uint16
		want front.ActKind
	}{{[]uint16{'y'}, front.ActYes}, {[]uint16{'n'}, front.ActNo}, {[]uint16{27}, front.ActNo},
		{[]uint16{13}, front.ActNo}, {[]uint16{0x4800, KeyPadSelect}, front.ActYes}} {
		p := &Presenter{confirmSel: 1}
		for _, k := range c.keys {
			p.key(k, v)
		}
		if a := p.Actions(); len(a) != 1 || a[0].Kind != c.want {
			t.Errorf("keys %x: %v", c.keys, a)
		}
	}
}

// The pad reads the binding table: XCOM 2's combat buttons in combat, the
// adventuring ones elsewhere.
func TestPadKeys(t *testing.T) {
	keys := func(combat bool) map[gpu.StandardGamepadButton]uint16 {
		m := map[gpu.StandardGamepadButton]uint16{}
		for _, k := range PadKeys(combat) {
			m[k.B] = k.Key
		}
		return m
	}
	for b, want := range map[gpu.StandardGamepadButton]uint16{dUp: 0x4800, dLeft: 0x4b00, padA: KeyPadSelect, padB: 27, padLB: 0x4900, padRB: 0x5100,
		padLT: 'q', padRT: 'e', padR3: 0x4700, padY: 8, padX: 'g', padStart: KeyPadStart, padBack: 'l'} {
		if got := keys(true)[b]; got != want {
			t.Errorf("combat button %d sends %#x, want %#x", b, got, want)
		}
	}
	for b, want := range map[gpu.StandardGamepadButton]uint16{dUp: 0x4800, padA: KeyPadSelect, padY: KeyPadMinimap, padX: KeyPadOverview, padStart: 13, padBack: 14, padLT: 0x4700, padR3: KeyPadNext} {
		if got := keys(false)[b]; got != want {
			t.Errorf("button %d sends %#x, want %#x", b, got, want)
		}
	}
}

// Every binding has a keyboard, a mouse and a gamepad way, or says which
// is missing (gap); the help is made from the table, one line each.
func TestBindings(t *testing.T) {
	help := HelpLines()
	for _, b := range bindings {
		if (b.Keys == "" || b.Mouse == "" || b.Pad == "") && b.Gap == "" {
			t.Errorf("%s %q: keyboard %q, mouse %q, pad %q and no gap", b.Ctx, b.Desc, b.Keys, b.Mouse, b.Pad)
		}
		if !slices.Contains(help, [2]string{b.Keys, b.Desc}) {
			t.Errorf("%s %q has no help line", b.Ctx, b.Desc)
		}
		for _, s := range []string{b.Keys, b.Mouse, b.Pad} {
			if len(s) > helpKeyW || len(b.Desc) > helpDescW {
				t.Errorf("%q / %q too wide for the help", s, b.Desc)
			}
		}
	}
}

// The keyboard dispatch is the table's: every key code a row binds
// resolves to that row's Act in its context (no row shadows another), each
// such Act is one a screen switches on, and no screen names a key code.
func TestKeyDispatch(t *testing.T) {
	names := map[input.Act]string{} // the Acts by name, from the package's source
	f, err := parser.ParseFile(token.NewFileSet(), "../../../gobox-engine/input/input.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.CONST && len(g.Specs) > 1 {
			for i, sp := range g.Specs {
				names[input.Act(i)] = sp.(*ast.ValueSpec).Names[0].Name
			}
		}
	}
	var src strings.Builder
	files, _ := filepath.Glob("*.go")
	battle, _ := filepath.Glob("../../../gobox-engine/combat/*.go")
	scan := regexp.MustCompile(`0x(3[B-F]|4[0-9A-F]|5[0-3]|E[0-5]|FD)00\b|KeyPad|KeyF[19]|KeyFontPrev|\bk ?[=!]= ?(8|9|13|27)\b|case (8|9|13|27)\b`)
	for _, name := range append(files, battle...) {
		if strings.HasSuffix(name, "_test.go") || name == "bindings.go" {
			continue
		}
		b, _ := os.ReadFile(name)
		for i, l := range strings.Split(string(b), "\n") {
			if scan.MatchString(l) {
				t.Errorf("%s:%d names a key code; dispatch through the table (actOf, combatAct): %s", name, i+1, strings.TrimSpace(l))
			}
		}
		src.Write(b)
	}
	for _, b := range bindings {
		for _, c := range b.Codes {
			look := actOf
			if b.Ctx == ctxCombat {
				look = combatAct
			}
			if a := look(c.K); a != c.A {
				t.Errorf("%s %q: key %#x resolves to %s, not %s", b.Ctx, b.Desc, c.K, names[a], names[c.A])
			}
			if n := names[c.A]; c.A < input.Member1 || c.A >= input.Member1+input.Members {
				if !strings.Contains(src.String(), "input."+n) {
					t.Errorf("%s %q: no screen handles input.%s", b.Ctx, b.Desc, n)
				}
			}
		}
	}
}

// A script's unseen "press Enter" wait (area 18's PressToContinue before
// "YOUR COVER IS BLOWN."): a card's click continues it, as a click elsewhere,
// instead of asking for the inventory the engine refuses.
func TestCardClickContinues(t *testing.T) {
	p := &Presenter{shownView: front.View{Ask: &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Commands: []front.Command{{Text: "PRESS <ENTER>"}}}}}}
	p.cardClick(2)
	if a := p.Actions(); len(a) != 1 || a[0].Kind != front.ActOK {
		t.Errorf("card click on a press-Enter wait: %v", a)
	}
}

// A member pick while exploring (a spell's target): a card's click is the
// pick, not the command bar's select and context menu.
func TestCardClickPicksMember(t *testing.T) {
	p := &Presenter{command: true, shownView: front.View{Exploring: true, Ask: &front.Ask{Kind: front.AskMember}}}
	p.cardClick(6)
	if a := p.Actions(); len(a) != 1 || a[0].Kind != front.ActMember || a[0].N != 6 || p.ctx != 0 {
		t.Errorf("card click on a member pick: %v, ctx %d", a, p.ctx)
	}
}

// A question with Dirs takes the arrows as ActDir, not the command bar.
func TestDirsAskTakesArrows(t *testing.T) {
	p := &Presenter{active: true, view: &blobber.View{}, menu: true}
	v := front.View{Exploring: true, Ask: &front.Ask{Kind: front.AskMenu, Dirs: true}}
	if p.key(0x4800, v) {
		t.Fatal("the bar took an arrow the question answers")
	}
	p.keyAct(0x4800, v)
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActDir || p.inv.acts[0].N != 8 {
		t.Fatal(p.inv.acts)
	}
}

// A game's own menu (AskMenu "Continue Battle:") in a fight is answered by
// its keys and Esc, not left hanging.
func TestFightMenuKeys(t *testing.T) {
	v := front.View{Ask: &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "Continue Battle:",
		Commands: []front.Command{{Text: "Yes", Key: 'Y'}, {Text: "No", Key: 'N'}}}}}
	for _, c := range []struct {
		k    uint16
		want front.ActKind
	}{{'y', front.ActChoose}, {27, front.ActClose}} {
		p := &Presenter{fight: true, shownView: v}
		p.key(c.k, v)
		if a := p.Actions(); len(a) != 1 || a[0].Kind != c.want {
			t.Errorf("key %x: %v", c.k, a)
		}
	}
}

// TestConfirmStart: the yes/no starts on the game's default by its option,
// Yes drawn first whatever order the game lists them in.
func TestConfirmStart(t *testing.T) {
	yn := front.Menu{Commands: []front.Command{{Text: "Yes"}, {Text: "No"}}}
	ny := front.Menu{Commands: []front.Command{{Text: "No"}, {Text: "Yes"}}}
	for _, c := range []struct {
		a    front.Ask
		want int
	}{{front.Ask{Menu: yn}, 0}, {front.Ask{Menu: yn, Default: 1}, 1}, {front.Ask{Menu: ny}, 1}, {front.Ask{Menu: ny, Default: 1}, 0}, {front.Ask{Yes: true}, 0}, {front.Ask{}, 1}} {
		if got := confirmStart(&c.a); got != c.want {
			t.Errorf("%+v: row %d, want %d", c.a, got, c.want)
		}
	}
}

// TestConfirmNoEsc: Esc on a yes/no the game keeps up on Esc (Ask.NoEsc)
// answers nothing; on others it is No.
func TestConfirmNoEsc(t *testing.T) {
	for _, keep := range []bool{true, false} {
		v := front.View{Ask: &front.Ask{Kind: front.AskConfirm, Menu: front.Menu{Prompt: "Claim it?"}, NoEsc: keep}}
		p := &Presenter{confirmSel: 0, shownView: v}
		p.key(27, v)
		if a := p.Actions(); keep && len(a) != 0 || !keep && (len(a) != 1 || a[0].Kind != front.ActNo) {
			t.Errorf("NoEsc %v: %v", keep, a)
		}
	}
}

// TestConfirmNoPick: a yes/no with no highlight (Ask.NoPick, SotSB's place
// asks) lights no row; Enter and the arrows answer nothing, Y and N do,
// Esc as NoEsc says.
func TestConfirmNoPick(t *testing.T) {
	a := &front.Ask{Kind: front.AskConfirm, Menu: front.Menu{Prompt: "Pay?"}, NoPick: true, NoEsc: true}
	if confirmStart(a) != -1 {
		t.Error("a row is lit")
	}
	v := front.View{Ask: a}
	for _, c := range []struct {
		k    uint16
		want []front.ActKind
	}{{13, nil}, {0x4800, nil}, {27, nil}, {'y', []front.ActKind{front.ActYes}}, {'N', []front.ActKind{front.ActNo}}} {
		p := &Presenter{confirmSel: -1, shownView: v}
		p.key(c.k, v)
		var got []front.ActKind
		for _, x := range p.Actions() {
			got = append(got, x.Kind)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("key %#x: %v, want %v", c.k, got, c.want)
		}
	}
}
