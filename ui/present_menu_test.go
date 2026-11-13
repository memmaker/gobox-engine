package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"gobox/game"
	"gobox/game/play"
	"gobox/ui/blobber"
	"gobox/ui/input"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"gobox/gpu"
)

// Esc closes the command menu (Enter opened it).
func TestEscClosesCommandMenu(t *testing.T) {
	g, _ := playGame(play.Exploring{})
	p := &Presenter{active: true, view: &blobber.View{}}
	p.SetEngine(g)
	if !p.key(13) || !p.menu {
		t.Fatal("Enter did not open the menu")
	}
	if !p.key(27) || p.menu {
		t.Fatal("Esc did not close the menu")
	}
	p.key(13)
	p.Mouse(0, 0, false, false, true, 0)
	if p.menu {
		t.Fatal("a right click did not close the menu")
	}
}

// menuNav: Esc is back and never picks; a hotkey picks in either case; a
// hotkey letter not in the text is shown in front.
func TestMenuNav(t *testing.T) {
	sel := 0
	keys := []byte{'C', 'E'}
	if i, back, _ := MenuNav(actOf(27), 27, &sel, keys); i != -1 || !back {
		t.Fatal("Esc picked a row")
	}
	if i, _, _ := MenuNav(actOf('e'), 'e', &sel, keys); i != 1 || sel != 1 {
		t.Fatal("hotkey e did not pick row 1")
	}
	if MenuNav(actOf(0x5000), 0x5000, &sel, keys); sel != 0 {
		t.Fatal("Down did not wrap")
	}
	if s, at := HotText("Camp", 'E'); s != "[E] Camp" || at != 1 {
		t.Fatal(s)
	}
	if s, _ := HotText("Cure Blindness  1000g", '1'); s != "[1] Cure Blindness  1000g" { // a digit in a price is not the key
		t.Fatal(s)
	}
}

// digitKeys: rows with no letter and that can be picked get 1-9, skipping
// digits already taken, then none.
func TestDigitKeys(t *testing.T) {
	keys := DigitKeys([]byte{'B', 0, '2', 0, 0}, func(i int) bool { return i != 3 })
	if string(keys) != "B12\x003" {
		t.Fatalf("%q", keys)
	}
	many := DigitKeys(make([]byte, 12), func(int) bool { return true })
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
	for m, want := range map[Money]string{{}: "0g", {Gold: 12}: "12g", {Gold: 12, Copper: 5}: "12g 5c", {Copper: 50}: "0g 50c"} {
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

// askedOn is a Presenter with the game's question a up (as a refusal's:
// over any screen).
func askedOn(a *play.Ask) *Presenter {
	p, _ := presenterOn(play.Scene{})
	p.eng.r.refusalAsk = a
	return p
}

// The engine's yes/no question: Y, N, Esc, Enter (on No first) answer it
// with ActYes or ActNo; the arrows move to Yes.
func TestConfirmKeys(t *testing.T) {
	for _, c := range []struct {
		keys []uint16
		want ActKind
	}{{[]uint16{'y'}, ActYes}, {[]uint16{'n'}, ActNo}, {[]uint16{27}, ActNo},
		{[]uint16{13}, ActNo}, {[]uint16{0x4800, KeyPadSelect}, ActYes}} {
		p := askedOn(&play.Ask{Kind: game.YesNo, Prompt: "Flee?"})
		p.confirmSel = 1
		for _, k := range c.keys {
			p.key(k)
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
	f, err := parser.ParseFile(token.NewFileSet(), "input/act.go", nil, 0)
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
	// The package holds the screens, the key tables and the key definitions; only the screens are scanned.
	tables := map[string]bool{
		"present_bindings.go": true, // the key table (it was present's bindings.go)
		"bioskeys.go":         true, // the BIOS key codes of the host's keys
		"vace.go":             true, // VACE opcodes, not keys
	}
	used := map[string]bool{} // the identifiers the screens use
	files, _ := filepath.Glob("*.go")
	sub, _ := filepath.Glob("combat/*.go") // the board takes Acts too
	files = append(files, sub...)
	scan := regexp.MustCompile(`0x(3[B-F]|4[0-9A-F]|5[0-3]|E[0-5]|FD)00\b|KeyPad|KeyF[19]|KeyFontPrev|\bk ?[=!]= ?(8|9|13|27)\b|case (8|9|13|27)\b`)
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || tables[name] {
			continue
		}
		b, _ := os.ReadFile(name)
		for i, l := range strings.Split(string(b), "\n") {
			if scan.MatchString(l) {
				t.Errorf("%s:%d names a key code; dispatch through the table (actOf, combatAct): %s", name, i+1, strings.TrimSpace(l))
			}
		}
		sf, err := parser.ParseFile(token.NewFileSet(), name, b, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(sf, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				used[id.Name] = true
			}
			return true
		})
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
				if !used[n] {
					t.Errorf("%s %q: no screen handles %s", b.Ctx, b.Desc, n)
				}
			}
		}
	}
}

// A script's unseen "press Enter" wait (area 18's PressToContinue before
// "YOUR COVER IS BLOWN."): a card's click continues it, as a click elsewhere,
// instead of asking for the inventory the engine refuses.
func TestCardClickContinues(t *testing.T) {
	p := askedOn(&play.Ask{Kind: game.Continue})
	p.cardClick(2)
	if a := p.Actions(); len(a) != 1 || a[0].Kind != ActOK {
		t.Errorf("card click on a press-Enter wait: %v", a)
	}
}

// A member pick while exploring (a spell's target): a card's click is the
// pick, not the command bar's select and context menu.
func TestCardClickPicksMember(t *testing.T) {
	p := askedOn(&play.Ask{Kind: game.PickMember})
	p.command = true
	p.cardClick(6)
	if a := p.Actions(); len(a) != 1 || a[0].Kind != ActMember || a[0].N != 6 || p.ctx != 0 {
		t.Errorf("card click on a member pick: %v, ctx %d", a, p.ctx)
	}
}

// A heading to pick (AskMove) takes the arrows as ActDir, not the command bar.
func TestDirsAskTakesArrows(t *testing.T) {
	p := askedOn(&play.Ask{Kind: game.Direction})
	p.active, p.view, p.menu = true, &blobber.View{}, true
	if p.key(0x4800) {
		t.Fatal("the bar took an arrow the question answers")
	}
	p.keyAct(0x4800)
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != ActDir || p.inv.acts[0].N != 8 {
		t.Fatal(p.inv.acts)
	}
}

// A game's own menu (AskMenu "Continue Battle:") in a fight is answered by
// its keys and Esc, not left hanging.
func TestFightMenuKeys(t *testing.T) {
	menu := &play.Ask{Kind: game.Choice, Prompt: "Continue Battle:", Options: []play.Option{{Label: "Yes", Key: 'Y'}, {Label: "No", Key: 'N'}}}
	for _, c := range []struct {
		k    uint16
		want ActKind
	}{{'y', ActChoose}, {27, ActClose}} {
		p := askedOn(menu)
		p.fight = true
		p.key(c.k)
		if a := p.Actions(); len(a) != 1 || a[0].Kind != c.want {
			t.Errorf("key %x: %v", c.k, a)
		}
	}
}

// TestConfirmStart: the yes/no starts on the game's default by its option,
// Yes drawn first whatever order the game lists them in.
func TestConfirmStart(t *testing.T) {
	yn := Menu{Commands: []Command{{Text: "Yes"}, {Text: "No"}}}
	ny := Menu{Commands: []Command{{Text: "No"}, {Text: "Yes"}}}
	for _, c := range []struct {
		a    FrontAsk
		want int
	}{{FrontAsk{Menu: yn}, 0}, {FrontAsk{Menu: yn, Default: 1}, 1}, {FrontAsk{Menu: ny}, 1}, {FrontAsk{Menu: ny, Default: 1}, 0}, {FrontAsk{Yes: true}, 0}, {FrontAsk{}, 1}} {
		if got := confirmStart(&c.a); got != c.want {
			t.Errorf("%+v: row %d, want %d", c.a, got, c.want)
		}
	}
}

// TestConfirmNoEsc: Esc on a yes/no the game keeps up on Esc (Ask.NoEsc)
// answers nothing; on others it is No.
func TestConfirmNoEsc(t *testing.T) {
	for _, keep := range []bool{true, false} {
		p := askedOn(&play.Ask{Kind: game.YesNo, Prompt: "Claim it?", NoEsc: keep})
		p.confirmSel = 0
		p.key(27)
		if a := p.Actions(); keep && len(a) != 0 || !keep && (len(a) != 1 || a[0].Kind != ActNo) {
			t.Errorf("NoEsc %v: %v", keep, a)
		}
	}
}

// TestConfirmNoPick: a yes/no with no highlight (Ask.NoPick, SotSB's place
// asks) lights no row; Enter and the arrows answer nothing, Y and N do,
// Esc as NoEsc says.
func TestConfirmNoPick(t *testing.T) {
	a := &FrontAsk{Kind: AskConfirm, Menu: Menu{Prompt: "Pay?"}, NoPick: true, NoEsc: true}
	if confirmStart(a) != -1 {
		t.Error("a row is lit")
	}
	for _, c := range []struct {
		k    uint16
		want []ActKind
	}{{13, nil}, {0x4800, nil}, {27, nil}, {'y', []ActKind{ActYes}}, {'N', []ActKind{ActNo}}} {
		p := askedOn(&play.Ask{Kind: game.YesNo, Prompt: "Pay?", NoPick: true, NoEsc: true})
		p.confirmSel = -1
		p.key(c.k)
		var got []ActKind
		for _, x := range p.Actions() {
			got = append(got, x.Kind)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("key %#x: %v, want %v", c.k, got, c.want)
		}
	}
}
