package present

import (
	"fmt"
	"gobox/front"
	"image"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gobox/combat"
	"gobox/ega"
	"gobox/gpu"
	"gobox/ui"

	"gobox/input"
)

// The combat screen's own rows (extra): the original's two rows under the
// map (the panel and the message line) drawn from View.Battle in its font,
// colours and border; the spell list Cast opens and the battle Log.

const fightRows = 1 // cells: the status line

const logRows = 20 // the battle Log's lines shown at a time

// drawFight draws the rows, the bar, whose turn it is, the spell list and the Log over the 3D map.
func (p *Presenter) drawFight(dst *gpu.Image, note Note, v front.View) {
	glyph := note.Glyph
	b, k := askedBattle(v), scale(dst)
	p.mapBottom = dst.Bounds().Dy() - int(math.Ceil(k*8*fightRows)) // the line is drawn at k, whole or not
	if b.Turn != nil {
		for _, m := range v.Party { // whose turn: the portrait card, bottom left
			if m.Selected {
				if m.Portrait == nil && m.Head == nil { // no portrait: the member's combat icon, else no box
					for _, u := range b.Units {
						if u.ID == b.Turn.Unit {
							m.Portrait = u.Icon
						}
					}
					if m.Portrait == nil {
						continue
					}
				}
				ck := max(1, int(k)/2) // the portrait card: its own scale (cards are not K)
				y := p.mapBottom - (cardH+cardGap)*ck + cardFoot*ck
				p.fightCard = image.Rect(cardGap*ck, y, cardGap*ck+cardW*ck, y+cardH*ck)
				p.drawParty(dst, []front.Member{m}, func(int, int) image.Rectangle { return p.fightCard })
			}
		}
	}
	p.battle.Icons, p.battle.IconKey = nil, ""
	if b.Turn != nil { // the Attack button shows the weapon in hand, Ranged what it shoots
		p.inv.load()
		for kind, it := range map[front.ActKind]*front.InvItem{front.ActAttack: b.Turn.Weapon, front.ActRanged: b.Turn.Missile} {
			if it == nil {
				continue
			}
			if ic := p.inv.iconImage(it); ic != nil {
				n := p.inv.icon(it)
				if p.battle.Icons == nil {
					p.battle.Icons = map[combat.Kind]image.Image{}
				}
				p.battle.Icons[combatKinds[kind]] = ic
				p.battle.IconKey += strconv.Itoa(int(kind)) + ":" + strconv.Itoa(n) + " "
			}
		}
	}
	p.syncBattle()
	p.battle.Pointer = p.pointer
	p.battle.DrawBar(dst, glyph, toBoard(b), int(k), p.mapBottom) // the bar and its words at K
	if ck, r, ok := p.battle.Hovered(); ok && b.Turn != nil {
		kind := engineKinds[ck] // the readied weapon over Attack and Ranged: what it does, its rules
		if it := map[front.ActKind]*front.InvItem{front.ActAttack: b.Turn.Weapon, front.ActRanged: b.Turn.Missile}[kind]; it != nil {
			p.drawTipOver(dst, it.Name, append(slices.Clone(it.Look), it.Rules...), int(k), r)
		}
		if kind == front.ActBandage {
			p.drawTipOver(dst, "Bandage", []string{
				"Binds the wounds of the first",
				"party member who is Dying.",
				"They become Unconscious and",
				"stop bleeding; the Dying die",
				"after 10 rounds unbandaged.",
				"Uses this character's turn.",
			}, int(k), r)
		}
	}
	p.drawFightRows(dst, note, b, k)
	cx := dst.Bounds().Dx() / 2
	p.fightAskAt = nil
	if m := askMenu(v); m != nil && len(m.Commands) > 0 {
		lines, keys := make([]string, len(m.Commands)), make([]byte, len(m.Commands))
		for i, it := range m.Commands {
			lines[i], keys[i] = it.Text, it.Key
		}
		if id := menuID(m); id != p.menuFor {
			p.menuFor, p.sel = id, min(askDefault(v), len(lines)-1)
		}
		p.keyWindow(dst, m.Prompt, lines, keys, nil, p.sel, false)
		p.fightAskAt = slices.Clone(p.winRects)
	}
	switch {
	case p.battle.Picking && b.Turn != nil && p.pick == nil: // the bar's Cast list (View.Ask (AskSpell) is drawn by Draw)
		p.drawPickList(dst, glyph, k, cx, int(k)*16)
	case p.fightLog:
		var lines []string
		for _, m := range b.Log {
			for _, l := range ui.WrapWords(m, 36) {
				lines = append(lines, logInk(l))
			}
		}
		if len(lines) == 0 {
			lines = []string{"Nothing has happened yet."}
		}
		n := len(lines)
		p.logTop = min(max(p.logTop, 0), max(0, n-logRows))
		lines = lines[p.logTop:min(n, p.logTop+logRows)]
		p.logGrey = true
		p.drawMenu(dst, glyph, "Battle Log", lines, make([]byte, len(lines)), nil, -1, k, cx, int(k)*16)
		p.logGrey = false
		p.menuBar(dst, p.menuRects, n, len(lines), &p.logTop, nil, k)
		p.menuRects = p.menuRects[:0]
	}
}

// drawTipOver draws a tooltip (title, lines) at scale k above r, kept on screen.
func (p *Presenter) drawTipOver(dst *gpu.Image, title string, lines []string, k int, r image.Rectangle) {
	rows := append(ui.WrapWords(title, 24), "")
	rows = append(rows, lines...)
	w := 0
	for _, l := range rows {
		w = max(w, len(l))
	}
	w, h := (w+4)*8*k, (len(rows)+2)*8*k
	cx := r.Min.X + r.Dx()/2
	top := r.Min.Y - h - 4*k
	if top < 0 { // no room above: under r (tipMenu keeps it on the window)
		top = r.Max.Y + 4*k
	}
	p.tipMenu(dst, title, lines, make([]byte, len(lines)), -1, float64(k), cx, top)
}

// drawFightRows draws the status line along the bottom: the question
// asked, the aiming help, else the panel (the unit under the pointer, else
// the one the game names); between the player's turns the last message.
// All of them are in the Log (L).
func (p *Presenter) drawFightRows(dst *gpu.Image, note Note, b *front.Battle, k float64) {
	w, h := max(40, int(float64(dst.Bounds().Dx())/k)/8)*8, fightRows*8 // as many cells as the window has
	glyph := note.Glyph
	if p.rowsImg == nil || p.rowsImg.Bounds().Dx() != w {
		p.rowsKey = ""
		p.rowsImg = gpu.NewImage(w, h)
	}
	var puts []struct { // the line's pieces: drawn only when they change (a texture upload a frame otherwise)
		s string
		c int
	}
	x := 0
	put := func(s string, c int, y int) {
		s = s[:min(len(s), max(0, (w-x)/8))]
		puts = append(puts, struct {
			s string
			c int
		}{s, c})
		x += len(s)*8 + 8
	}
	panel := b.Panel
	if b.Turn != nil && panel == 0 {
		panel = b.Turn.Unit
	}
	hover := false
	if mx, my, ok := p.battle.Pick(gpu.CursorPosition()); ok && my < p.mapBottom {
		for _, u := range b.Units {
			for _, s := range u.Squares {
				if s == image.Pt(mx, my) {
					panel, hover = u.ID, true
				}
			}
		}
	}
	switch {
	case note.Show: // a refusal (why the attack is not possible ...): the toast, as every view's (noteLine)
	case b.Turn != nil && b.Turn.Aim && b.AimWhy != "":
		put(b.AimWhy, inkRed, 0)
	case b.Turn != nil && b.Turn.Aim && b.Turn.Casting != "":
		put(b.Turn.Casting+": W A S D, Tab; Enter, click; Esc", inkGreen, 0)
	case b.Turn != nil && b.Turn.Aim:
		put("Aim: W A S D, Tab; Enter, click; Esc", inkGreen, 0)
	case p.blowLine != "" && !hover && now().Sub(p.blowAt) < blowLineFor: // a blow just played: what it said
		put(p.blowLine, inkGreen, 0)
	case b.Turn == nil && b.Panel == 0 && !hover && len(b.Log) > 0: // between the player's turns: what happens
		put(b.Log[len(b.Log)-1], inkGreen, 0)
	default:
		for _, u := range b.Units {
			if u.ID != panel {
				continue
			}
			name := inkGreen
			if !u.Party {
				name = inkRed
			}
			put(u.Name, name, 0)
			x -= 8
			put(" HP", inkWhite, 0)
			x -= 8
			put(" "+strconv.Itoa(u.HP), inkGreen, 0)
			put("AC", inkWhite, 0)
			x -= 8
			put(" "+strconv.Itoa(u.AC), inkGreen, 0)
			if u.Status != "" { // before the weapon, as BR2's O099:0560: GIZMO HP 43 AC 2 OKAY BOLT GUN
				put(u.Status, inkWhite, 0)
			}
			if u.Weapon != "" { // the readied weapon's name, as the originals' panel
				put(u.Weapon, inkGreen, 0)
			} else if u.Dmg != "" { // no label for an empty value
				put("Dmg", inkWhite, 0)
				x -= 8
				put(" "+u.Dmg, inkGreen, 0)
			}
			if u.Hit > 0 {
				put("CtH", inkWhite, 0)
				x -= 8
				put(" "+strconv.Itoa(u.Hit), inkGreen, 0)
			}
		}
	}
	if key := fmt.Sprint(note.Show, note.Text, puts); key != p.rowsKey {
		p.rowsKey = key
		im := image.NewRGBA(image.Rect(0, 0, w, h))
		for i := 3; i < len(im.Pix); i += 4 {
			im.Pix[i] = 255
		}
		x = 0
		if note.Show {
			noteLine(im, note, 0)
		}
		for _, t := range puts {
			ui.DrawText(im, glyph, t.s, x, 0, ega.RGBA(t.c))
			x += len(t.s)*8 + 8
		}
		p.rowsImg.WritePixels(im.Pix)
	}
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate((float64(dst.Bounds().Dx())-k*float64(w))/2, float64(p.mapBottom))
	dst.DrawImage(p.rowsImg, op)
}

// fightKey takes key k on the combat screen: the spell list's, L for the
// Log, else the 3D map's and bar's.
// CameraKey is whether k turns the fight's camera (it acts while the fight's events play).
func (p *Presenter) CameraKey(k uint16) bool {
	return p.fight && combat.CameraAct(combatAct(k))
}

func (p *Presenter) fightKey(k uint16) bool {
	if m := askMenu(p.shownView); m != nil && len(m.Commands) > 0 { // the game's own menu (Continue Battle: Yes/No ...), over the board
		keys := make([]byte, len(m.Commands))
		for i, it := range m.Commands {
			keys[i] = it.Key
		}
		switch i, back, ok := ui.MenuNav(actOf(k), k, &p.sel, keys); {
		case back:
			p.send(front.ActClose, 0)
		case i >= 0:
			p.send(front.ActChoose, i)
		case !ok:
			if k > ' ' && k < 128 { // a letter or digit the menu has no line for: the bar's hotkey (Quick) under it
				p.syncBattle()
				if p.battle.Key(combatAct(k), k) {
					return true
				}
			}
			return !globalKey(k)
		}
		return true
	}
	if p.pickKey(k) {
		return true
	}
	a := combatAct(k)
	if a == input.Log || p.fightLog && a == input.Esc {
		p.fightLog = !p.fightLog && a != input.Esc
		p.logTop = math.MaxInt32 // it opens on the latest lines
		return true
	}
	if d, ok := map[input.Act]int{input.Up: -1, input.Down: 1, input.PgUp: -logRows, input.PgDn: logRows, input.Home: -math.MaxInt32, input.End: math.MaxInt32}[actOf(k)]; ok && p.fightLog { // Up/Down, PgUp/PgDn, Home/End scroll the Log, as every list
		p.logTop = max(0, min(p.logTop, math.MaxInt32-logRows)+d) // clamped when drawn
		return true
	}
	if a == input.Options && p.view != nil {
		p.view.OpenOptions()
		return true
	}
	if a == input.PadA { // pad A: as Enter
		a = input.Enter
	}
	if v := p.shownView; askedTurn(v) && v.Ask.Kind == front.AskMove { // a step at a time, as the original's move mode: the direction keys step, Enter is done, Esc undoes
		if d, ok := keypadDirs[actOf(k)]; ok {
			p.send(front.ActDir, d)
			return true
		}
		switch a {
		case input.Enter:
			p.send(front.ActOK, 0)
			return true
		case input.Esc:
			p.send(front.ActClose, 0)
			return true
		}
	}
	p.syncBattle()
	if p.battle.Key(combatAct(k), k) {
		return true
	}
	if a == input.Esc && p.view != nil { // Esc with nothing to give up: the options
		p.view.OpenOptions()
		return true
	}
	return false
}

// fightClick takes a click at (x, y) on the combat screen.
func (p *Presenter) fightClick(x, y int) bool {
	if askMenu(p.shownView) != nil && len(askMenu(p.shownView).Commands) > 0 {
		if i := slices.IndexFunc(p.fightAskAt, image.Pt(x, y).In); i >= 0 {
			p.send(front.ActChoose, i)
		}
		return true
	}
	if p.fightLog {
		p.fightLog = false
		return true
	}
	if p.pickClick(x, y) {
		return true
	}
	if image.Pt(x, y).In(p.fightCard) && p.fightView != nil && p.fightView.Turn != nil { // the active member's sheet (View has no button: Felix's ruling 2026-11-13)
		p.send(front.ActView, -1)
		return true
	}
	return p.battle.Click(x, y)
}

// replay is the events still to play (Play): one held for Ms (or until a
// key: Key) shows the fight or the page (Scene) it carries; the engine waits until all are played (no
// ActTick, no decisions: Playing), then the live View is drawn again.
type replay struct {
	q    []front.CombatEvent
	from time.Time // the shown event began
	skip bool      // a key cut the shown event short
	fast bool      // Space: the rest of the queued timed events pass at once (the computer's turn sped up)
}

// lastBoard: the shown event is the last held one queued (the board the
// rest leave): Space never skips it, so the fight reaches its end state.
func (r *replay) lastBoard() bool {
	for _, e := range r.q[1:] {
		if (e.Key || e.Ms > 0) && (e.Battle != nil || e.Scene != nil) {
			return false
		}
	}
	return true
}

// Play queues the fight's events Apply handed out.
// Their snapshots' pictures are made here, as View.Battle's are (pixView).
func (p *Presenter) Play(evs []front.CombatEvent) {
	for _, ev := range evs {
		ev.Battle = pixBattle(ev.Battle, egaInk)
		p.replay.q = append(p.replay.q, ev)
	}
}

// Playing: the fight's events are still being played.
func (p *Presenter) Playing() bool { return len(p.replay.q) > 0 }

// playing is the event to show now (nil: none), the ones played dropped.
func (p *Presenter) playing() *front.CombatEvent {
	r := &p.replay
	for len(r.q) > 0 {
		if ev := &r.q[0]; r.from.IsZero() {
			switch {
			case ev.Kind == front.EvSound && p.Sound != nil:
				p.Sound(ev.Amount, ev.Text)
			case ev.Kind == front.EvSound && p.adLibEffect(ev.Amount):
			case ev.Kind == front.EvSound && p.Snd != nil:
				p.Snd.Effect(ev.Amount, false)
			case ev.Kind == front.EvSong && p.AdLib != nil && p.Snd != nil && p.Snd.Song(ev.Amount, 0, false): // the AdLib song, as the driver's soundtrack
			case ev.Kind == front.EvSong && p.Snd != nil:
				p.Snd.AmigaSong(ev.Amount)
			case ev.Kind == front.EvSaves && p.slots == nil:
				p.openSlots(p.shownView, false)
			}
		}
		if ev := &r.q[0]; r.from.IsZero() {
			if ev.Kind == front.EvMessage && strings.TrimSpace(ev.Text) != "" {
				p.blowLine, p.blowAt = ev.Text, now()
			}
			switch id, n, word := blowFloat(*ev); {
			case word != "":
				p.battle.FloatText(id, word)
			case n > 0 && Host.DamageNumbers:
				p.battle.Float(id, n, ev.Kind == front.EvHeal)
			}
		}
		if ev := &r.q[0]; (ev.Key || ev.Ms > 0 && (!r.fast || r.lastBoard())) && (ev.Battle != nil || ev.Scene != nil) && !r.skip {
			if r.from.IsZero() {
				r.from = now()
			}
			if ev.Key || now().Sub(r.from) < time.Duration(ev.Ms)*time.Millisecond {
				return ev
			}
		}
		if ev := &r.q[0]; traceOn && !r.from.IsZero() { // a timed moment ended
			fmt.Fprintf(os.Stderr, "trace %s anim end kind %d unit %d after %dms (Ms %d)\n", time.Now().Format("15:04:05.000"), ev.Kind, ev.Unit, now().Sub(r.from).Milliseconds(), ev.Ms)
		}
		r.q, r.from, r.skip = r.q[1:], time.Time{}, false
	}
	r.fast = false
	return nil
}

// CombatPad: the combat map is up with no menu over it (no question, Log
// or spell list): the gamepad takes the combat bindings (PadKeys).
func (p *Presenter) CombatPad() bool {
	return p.fight && confirm(p.shownView) == "" && !p.fightLog && p.pickList() == nil && !p.Options()
}

// PadStick hands over the pad's right stick (x right, y down, -1..1): in
// combat it pans the camera.
func (p *Presenter) PadStick(x, y float64) {
	dead := func(v float64) float64 { // a resting stick drifts a little
		if math.Abs(v) < 0.25 {
			return 0
		}
		return v
	}
	p.battle.Stick = [2]float64{dead(x), dead(y)}
}

// pickList is the spell list up for a choice: View.Ask (AskSpell) anywhere, else
// the bar's Cast in combat (Picking); nil when none.
func (p *Presenter) pickList() []front.Spell {
	switch {
	case p.pick != nil:
		return p.pick
	case p.fight && p.battle.Picking && p.fightView != nil && p.fightView.Turn != nil && p.battle.PickKind == combat.ActManeuver:
		out := []front.Spell{} // never nil: an empty list still takes Esc
		for _, m := range p.fightView.Turn.Maneuvers {
			out = append(out, front.Spell{ID: m.ID, Name: m.Name})
		}
		return out
	case p.fight && p.battle.Picking && p.fightView != nil && p.fightView.Turn != nil:
		return p.fightView.Turn.Spells
	}
	return nil
}

// maneuvering: the bar's Maneuver list is up (pickList holds maneuvers).
func (p *Presenter) maneuvering() bool {
	return p.pick == nil && p.fight && p.battle.Picking && p.battle.PickKind == combat.ActManeuver
}

// drawPickList draws the spell list (pickList) as a menu with hotkeys, its
// box centred on cx with its top at top.
func (p *Presenter) drawPickList(dst *gpu.Image, glyph func(byte) [8]byte, k float64, cx, top int) {
	lines := make([]string, len(p.pickList()))
	for i, s := range p.pickList() {
		lines[i] = s.Name
	}
	p.pickSel = min(max(p.pickSel, 0), len(lines)-1)
	title := "Cast"
	if p.maneuvering() {
		title = "Maneuver"
	}
	p.drawMenu(dst, glyph, title, lines, spellKeys(len(lines)), nil, p.pickSel, k, cx, top)
	switch {
	case p.pickSel < 0 || len(p.menuRects) == 0:
	case p.maneuvering():
		p.drawTip(dst, ui.WrapWords(p.fightView.Turn.Maneuvers[p.pickSel].Tip, 36), k, cx)
	default:
		p.drawSpellTip(dst, p.pickList()[p.pickSel].ID, k, cx)
	}
}

// spellKeys are a spell list's hotkeys, as every spell list's (camp,
// training): the digits 1-9 (digitKeys).
func spellKeys(n int) []byte { return ui.DigitKeys(make([]byte, n), func(int) bool { return true }) }

// pickKey takes key k while a spell is picked: the arrows, Enter, a hotkey, Esc.
func (p *Presenter) pickKey(k uint16) bool {
	spells := p.pickList()
	if spells == nil || globalKey(k) {
		return false
	}
	switch i, back, _ := ui.MenuNav(actOf(k), k, &p.pickSel, spellKeys(len(spells))); {
	case back:
		p.cast(-1)
	case i >= 0:
		p.cast(i)
	}
	return true
}

// pickClick takes a click while a spell is picked: a line casts, elsewhere gives up.
func (p *Presenter) pickClick(x, y int) bool {
	if p.pickList() == nil {
		return false
	}
	p.cast(slices.IndexFunc(p.menuRects, func(r image.Rectangle) bool { return image.Pt(x, y).In(r) }))
	return true
}

// cast casts spell i of the list (-1: none, giving up).
func (p *Presenter) cast(i int) {
	spells, fromBar := p.pickList(), p.pick == nil
	p.battle.Picking = false
	switch {
	case i >= 0 && i < len(spells) && fromBar && p.battle.PickKind == combat.ActManeuver:
		p.send(front.ActManeuver, spells[i].ID)
	case i >= 0 && i < len(spells):
		p.send(front.ActCast, spells[i].ID)
	case !fromBar: // the game waits for the pick: give it up
		p.send(front.ActClose, 0)
	}
}

// logSpecial are the battle Log's events drawn in gold: spells, backstabs, the fallen.
var logSpecial = regexp.MustCompile(`(?i)backstab|spell|casting|goes down|dying|killed|dies|slain`)

// logInk marks a battle Log line: its numbers (the damage) red, a special
// event's line gold, the rest in the log's grey.
func logInk(l string) string {
	gold := logSpecial.MatchString(l)
	var b strings.Builder
	mark := func() { // opens or closes the gold span
		if gold {
			b.WriteByte(ui.InkGold)
		}
	}
	mark()
	for i := 0; i < len(l); i++ {
		if l[i] >= '0' && l[i] <= '9' {
			j := i
			for j < len(l) && l[j] >= '0' && l[j] <= '9' {
				j++
			}
			mark()
			b.WriteByte(ui.InkRed)
			b.WriteString(l[i:j])
			b.WriteByte(ui.InkRed)
			mark()
			i = j - 1
			continue
		}
		b.WriteByte(l[i])
	}
	mark()
	return b.String()
}

// adLibEffect plays effect n on the AdLib player when it has it and the
// soundtrack is on (the sound setting); false leaves n to the driver.
// ponytail: Effect is not locked against the driver's Read; a lock hook in audio if it races.
func (p *Presenter) adLibEffect(n int) bool {
	if p.AdLib == nil || p.Snd == nil {
		return false
	}
	if on, _ := p.Snd.Track(); !on {
		return true // sound off: nothing plays, no fallback either
	}
	p.AdLib.Reset()
	return p.AdLib.Effect(n)
}

// blowLineFor is how long a blow's line stays on the pane.
const blowLineFor = 2500 * time.Millisecond

// blowFloat is what ev floats over unit id: number n or word (none: n 0,
// word ""). One blow floats once: only EvMiss says MISS (not the EvAttack
// swing), only EvDamage/EvHeal carry the number (not EvHit, not Strike).
func blowFloat(ev front.CombatEvent) (id, n int, word string) {
	switch ev.Kind {
	case front.EvMiss:
		return ev.Target, 0, "MISS"
	case front.EvDamage, front.EvHeal:
		return ev.Unit, ev.Amount, ""
	}
	return 0, 0, ""
}

// traceOn (GOBOX_TRACE=1): the fight's timed moments say when they end.
var traceOn = os.Getenv("GOBOX_TRACE") != ""
