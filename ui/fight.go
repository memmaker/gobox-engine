package ui

import (
	"fmt"
	"gobox/game"
	"gobox/rules"
	"gobox/ui/combat"
	"gobox/ui/input"
	"image"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gobox/asset"
	"gobox/gpu"
)

// The combat screen's own rows (extra): the original's two rows under the
// map (the panel and the message line) drawn from the fight in its font,
// colours and border; the spell list Cast opens and the battle Log.

const fightRows = 1 // cells: the status line

const logRows = 20 // the battle Log's lines shown at a time

// drawFight draws the rows, the bar, whose turn it is, the spell list and the Log over the 3D map.
func (p *Presenter) drawFight(dst *gpu.Image, note Note) {
	glyph := note.Glyph
	b, k := p.fightBoard, scale(dst)
	ck := max(1, int(k)/2)                                          // the portrait card and the action bar: their own scale (the portrait card's, as before ui.K)
	p.mapBottom = dst.Bounds().Dy() - int(math.Ceil(k*8*fightRows)) // the line is drawn at k, whole or not
	if b.Turn != nil {
		if i := p.selectedIndex(); i >= 0 { // whose turn: the portrait card, bottom left
			var icon image.Image
			bare := p.party()[i].Dropped == nil && headOf(p.party()[i]) == nil
			if bare { // no portrait: the member's combat icon, else no box
				for _, u := range b.Units {
					if u.ID == b.Turn.Unit {
						icon = u.Icon
					}
				}
			}
			if !bare || icon != nil {
				y := p.mapBottom - (cardH+cardGap)*ck + cardFoot*ck
				p.fightCard = image.Rect(cardGap*ck, y, cardGap*ck+cardW*ck, y+cardH*ck)
				p.drawParty(dst, p.party()[i:i+1], 0, func(int, int) image.Rectangle { return p.fightCard }, icon)
			}
		}
	}
	p.battle.Icons, p.battle.IconKey = nil, ""
	weap, shot := p.fightLists.weapon, p.fightLists.missile
	if b.Turn != nil { // the Attack button shows the weapon in hand, Ranged what it shoots
		p.inv.load()
		for kind, it := range map[ActKind]*InvItem{ActAttack: weap, ActRanged: shot} {
			if it == nil {
				continue
			}
			if ic := p.inv.iconImage(it.Base); ic != nil {
				n := p.inv.icon(it.Base)
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
	p.battle.DrawBar(dst, glyph, b, ck, p.mapBottom) // the bar and its words at the portrait card's scale
	if ck, r, ok := p.battle.Hovered(); ok && b.Turn != nil && tipShown() {
		kind := engineKinds[ck] // the readied weapon over Attack and Ranged: what it does, its rules
		if it := map[ActKind]*InvItem{ActAttack: weap, ActRanged: shot}[kind]; it != nil {
			p.drawTipOver(dst, it.Name, append(slices.Clone(it.Look), it.Rules...), int(k), r)
		}
		if kind == ActBandage {
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
	if p.askList() != nil { // a list over the board (the turn's Items): drawn as on every screen, its keys and clicks list.go's
		rows := min(len(p.askList()), listRows) + len(p.askMenu().Commands) + 6
		p.asking(func() { p.drawList(dst, glyph, k, cx, max(0, dst.Bounds().Dy()/2-int(k*float64(rows)*4))) })
	} else if m := p.askMenu(); m != nil && len(m.Commands) > 0 {
		lines, keys := make([]string, len(m.Commands)), make([]byte, len(m.Commands))
		for i, it := range m.Commands {
			lines[i], keys[i] = it.Text, it.Key
		}
		if id := menuID(m); id != p.menuFor {
			p.menuFor, p.sel = id, min(p.askDefault(), len(lines)-1)
		}
		p.keyWindow(dst, m.Prompt, lines, keys, nil, p.sel, false)
		p.fightAskAt = slices.Clone(p.winRects)
	}
	switch {
	case p.battle.Picking && b.Turn != nil && p.pick == nil: // the bar's Cast list (the Ask (AskSpell) is drawn by Draw)
		p.drawPickList(dst, glyph, k, cx, int(k)*16)
	case p.fightLog:
		var lines []string
		for _, m := range b.Log {
			for _, l := range WrapWords(m, 36) {
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
	rows := append(WrapWords(title, 24), "")
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
func (p *Presenter) drawFightRows(dst *gpu.Image, note Note, b *combat.Board, k float64) {
	w, h := max(40, int(float64(dst.Bounds().Dx())/k)/8)*8, fightRows*8 // as many cells as the window has
	glyph := note.Glyph
	if p.rowsImg == nil || p.rowsImg.Bounds().Dx() != w {
		p.rowsKey = rowsKey{}
		p.rowsImg = gpu.NewImage(w, h)
	}
	type piece struct {
		s    string
		c, x int // its colour and where it starts
	}
	var puts []piece // the line's pieces: drawn only when they change (a texture upload a frame otherwise)
	x := 0
	put := func(s string, c int, y int) {
		s = s[:min(len(s), max(0, (w-x)/8))]
		puts = append(puts, piece{s, c, x})
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
			x += 8
			pair := func(label, value string) { // label, one space, value, then two spaces before the next label
				put(label, inkWhite, 0)
				x -= 8
				put(" "+value, inkGreen, 0)
				x += 8
			}
			pair("HP", strconv.Itoa(u.HP))
			pair("AC", strconv.Itoa(u.AC))
			if u.Status != "" { // before the weapon, as BR2's O099:0560: GIZMO HP 43 AC 2 OKAY BOLT GUN
				put(u.Status, inkWhite, 0)
				x += 8
			}
			if u.Weapon != "" { // the readied weapon's name, as the originals' panel
				put(u.Weapon, inkGreen, 0)
				x += 8
			} else if u.Dmg != "" { // no label for an empty value
				pair("Dmg", u.Dmg)
			}
			if u.Hit > 0 {
				pair("CtH", strconv.Itoa(u.Hit))
			}
		}
	}
	if key := (rowsKey{fmt.Sprint(note.Show, note.Text, puts), p.glyphs()}); key != p.rowsKey {
		p.rowsKey = key
		im := image.NewRGBA(image.Rect(0, 0, w, h))
		for i := 3; i < len(im.Pix); i += 4 {
			im.Pix[i] = 255
		}
		if note.Show {
			noteLine(im, note, 0)
		}
		for _, t := range puts { // each where put placed it (pair's spacing)
			DrawText(im, glyph, t.s, t.x, 0, asset.EGARGBA(t.c))
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
	if m := p.askMenu(); m != nil && len(m.Commands) > 0 { // the game's own menu (Continue Battle: Yes/No ...), over the board
		keys := make([]byte, len(m.Commands))
		for i, it := range m.Commands {
			keys[i] = it.Key
		}
		switch i, back, ok := MenuNav(actOf(k), k, &p.sel, keys); {
		case back:
			p.send(ActClose, 0)
		case i >= 0:
			p.send(ActChoose, i)
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
	if a == input.InputOptions && p.view != nil {
		p.openOptions()
		return true
	}
	if a == input.PadA { // pad A: as Enter
		a = input.Enter
	}
	if p.askedTurn() && p.askUp().Kind == AskMove { // a step at a time, as the original's move mode: the direction keys step, Enter is done, Esc undoes
		if d, ok := keypadDirs[actOf(k)]; ok {
			p.send(ActDir, d)
			return true
		}
		switch a {
		case input.Enter:
			p.send(ActOK, 0)
			return true
		case input.Esc:
			p.send(ActClose, 0)
			return true
		}
	}
	p.syncBattle()
	if p.battle.Key(combatAct(k), k) {
		return true
	}
	if a == input.Esc && p.view != nil { // Esc with nothing to give up: the options
		p.openOptions()
		return true
	}
	return false
}

// fightClick takes a click at (x, y) on the combat screen.
func (p *Presenter) fightClick(x, y int) bool {
	if p.askMenu() != nil && len(p.askMenu().Commands) > 0 {
		if i := slices.IndexFunc(p.fightAskAt, image.Pt(x, y).In); i >= 0 {
			p.send(ActChoose, i)
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
	if image.Pt(x, y).In(p.fightCard) && p.fightBoard != nil && p.fightBoard.Turn != nil { // the active member's sheet (View has no button: Felix's ruling 2026-11-13)
		p.send(ActView, -1)
		return true
	}
	return p.battle.Click(x, y)
}

// replay is the moments still to play (Play): one held for Ms (or until a
// key: Key) shows the fight (Board) or the page or screen it carries; the
// engine waits until all are played (no ActTick, no decisions: Playing),
// then the live state is drawn again.
type replay struct {
	q    []frame
	from time.Time // the shown moment began
	skip bool      // a key cut the shown moment short
	fast bool      // Space: the rest of the queued timed moments pass at once (the computer's turn sped up)
}

// frame is a Moment queued to play; fly: a missile's sprite over its board,
// one square of its flight.
type frame struct {
	Moment
	fly *flight
}

// held is whether f holds the screen: a board or screen shown for Ms or
// until a key.
func (f *frame) held() bool {
	return f.Board != nil || f.Screen != nil
}

// lastBoard: the shown moment is the last held one queued (the board the
// rest leave): Space never skips it, so the fight reaches its end state.
func (r *replay) lastBoard() bool {
	for _, e := range r.q[1:] {
		if (e.Key || e.Ms > 0) && e.held() {
			return false
		}
	}
	return true
}

// Play queues the moments Settle handed out; a shot with a sprite plays as
// its flight over its board.
func (p *Presenter) Play(ms []Moment) {
	for _, m := range ms {
		if m.Kind == game.EvMissile && len(m.Sprite) > 0 && m.Board != nil {
			p.replay.q = append(p.replay.q, missileFlight(m, m.Sprite[0], m.StepMs)...)
			continue
		}
		p.replay.q = append(p.replay.q, frame{Moment: m})
	}
}

// missileFlight is the moment m (a missile held on its board) as the
// sprite's flight from m.Unit's square (else At) to To, a frame a third of
// a square, the moment's time shared among them.
func missileFlight(m Moment, art game.Picture, step int) []frame {
	from := m.At
	for _, u := range m.Board.Units {
		if u.ID == m.Unit {
			from = u.At
		}
	}
	d := m.To.Sub(from)
	n := 3 * max(d.X, -d.X, d.Y, -d.Y, 1)
	out := make([]frame, 0, n+1)
	for i := 0; i <= n; i++ {
		o := frame{Moment: m, fly: &flight{X: float64(from.X) + float64(d.X*i)/float64(n), Y: float64(from.Y) + float64(d.Y*i)/float64(n), Art: art}}
		o.At, o.Ms = from, max(m.Ms/(n+1), 1)
		if step > 0 { // the game's own pace, whatever the game speed
			o.Ms = step
		}
		out = append(out, o)
	}
	return out
}

// Playing: the fight's moments are still being played.
func (p *Presenter) Playing() bool { return len(p.replay.q) > 0 }

// soundNo is the number of the sound or song m starts: its N, else its
// Sound ref read as a number.
func soundNo(m Moment) int {
	if n, err := strconv.Atoi(m.Sound); err == nil && m.N == 0 {
		return n
	}
	return m.N
}

// playing is the moment to show now (nil: none), the ones played dropped.
func (p *Presenter) playing() *frame {
	r := &p.replay
	for len(r.q) > 0 {
		if ev := &r.q[0]; r.from.IsZero() {
			switch {
			case ev.Kind == game.EvSound && p.Sound != nil:
				p.Sound(soundNo(ev.Moment), ev.Sound) // the cue: the pack's event, or the sound's number
			case ev.Kind == game.EvSound && p.adLibEffect(soundNo(ev.Moment)):
			case ev.Kind == game.EvSound && p.Snd != nil:
				p.Snd.Effect(soundNo(ev.Moment), false)
			case ev.Kind == game.EvSong && p.AdLib != nil && p.Snd != nil && p.Snd.Song(soundNo(ev.Moment), 0, false): // the AdLib song, as the driver's soundtrack
			case ev.Kind == game.EvSong && p.Snd != nil:
				p.Snd.AmigaSong(soundNo(ev.Moment))
			case ev.Kind == game.EvSaves && p.slots == nil:
				p.openSlots(false)
			}
		}
		if ev := &r.q[0]; r.from.IsZero() {
			if ev.Kind == game.EvMessage && strings.TrimSpace(ev.Text) != "" {
				p.blowLine, p.blowAt = ev.Text, now()
			}
			switch id, n, word := blowFloat(ev.Moment); {
			case word != "":
				p.battle.FloatText(id, word)
			case n > 0 && Host.DamageNumbers:
				p.battle.Float(id, n)
			}
		}
		if ev := &r.q[0]; (ev.Key || ev.Ms > 0 && (!r.fast || r.lastBoard())) && ev.held() && !r.skip {
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
	return p.fight && p.confirm() == "" && !p.fightLog && p.pickList() == nil && !p.Options()
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

// pickList is the spell list up for a choice: the Ask (AskSpell) anywhere, else
// the bar's Cast in combat (Picking); nil when none.
func (p *Presenter) pickList() []Spell {
	switch {
	case p.pick != nil:
		return p.pick
	case p.fight && p.battle.Picking && p.fightView != nil && p.fightView.Turn != nil && p.battle.PickKind == combat.DecideManeuver:
		if p.fightLists.maneuvers == nil { // never nil: an empty list still takes Esc
			return []Spell{}
		}
		return p.fightLists.maneuvers
	case p.fight && p.battle.Picking && p.fightView != nil && p.fightView.Turn != nil:
		return p.fightLists.spells
	}
	return nil
}

// maneuvering: the bar's Maneuver list is up (pickList holds maneuvers).
func (p *Presenter) maneuvering() bool {
	return p.pick == nil && p.fight && p.battle.Picking && p.battle.PickKind == combat.DecideManeuver
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
		p.drawTip(dst, WrapWords(p.fightView.Turn.Maneuvers[p.pickSel].Tip, 36), k, cx)
	default:
		p.drawSpellTip(dst, p.pickList()[p.pickSel].ID, k, cx)
	}
}

// spellKeys are a spell list's hotkeys, as every spell list's (camp,
// training): the digits 1-9 (digitKeys).
func spellKeys(n int) []byte { return DigitKeys(make([]byte, n), func(int) bool { return true }) }

// pickKey takes key k while a spell is picked: the arrows, Enter, a hotkey, Esc.
func (p *Presenter) pickKey(k uint16) bool {
	spells := p.pickList()
	if spells == nil || globalKey(k) {
		return false
	}
	switch i, back, _ := MenuNav(actOf(k), k, &p.pickSel, spellKeys(len(spells))); {
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
	case i >= 0 && i < len(spells) && fromBar && p.battle.PickKind == combat.DecideManeuver:
		p.send(ActManeuver, spells[i].ID)
	case i >= 0 && i < len(spells):
		p.send(ActCast, spells[i].ID)
	case !fromBar: // the game waits for the pick: give it up
		p.send(ActClose, 0)
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
			b.WriteByte(InkGold)
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
			b.WriteByte(InkRed)
			b.WriteString(l[i:j])
			b.WriteByte(InkRed)
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
		return !TwoMusics() // sound off: nothing plays, no fallback either; the Amiga's music picked: its effects, the driver's
	}
	p.AdLib.Reset()
	return p.AdLib.Effect(n)
}

// blowLineFor is how long a blow's line stays on the pane.
const blowLineFor = 2500 * time.Millisecond

// blowFloat is what m floats over unit id: number n or word (none: n 0,
// word ""). One blow floats once: only EvMiss says MISS (not the EvAttack
// swing), only EvDamage carries the number (not EvHit, not Strike).
func blowFloat(m Moment) (id, n int, word string) {
	switch m.Kind {
	case game.EvMiss:
		return m.Target, 0, "MISS"
	case game.EvDamage:
		return m.Unit, m.N, ""
	}
	return 0, 0, ""
}

// traceOn (GOBOX_TRACE=1): the fight's timed moments say when they end.
var traceOn = os.Getenv("GOBOX_TRACE") != ""

// liveFight is the fight up in the game's state (nil: none, or no game) and
// the party's figures: read again only when the game's State was built anew
// or gobox's own part of the fight changed (fightSource), not every frame.
func (p *Presenter) liveFight() (*game.Combat, []game.Member) {
	if p.eng == nil {
		return nil, nil
	}
	c := p.lay.combat.get(p.fightSource(), func() *game.Combat { return p.eng.r.combatOf(p.state()) })
	return c, p.party()
}

// liveCombat is the fight up in the game's state (nil: none).
func (p *Presenter) liveCombat() *game.Combat {
	c, _ := p.liveFight()
	return c
}

// lists are the turn's weapons and pick lists as the bar and the menus show
// them, made with the board (drawnBoard).
type lists struct {
	weapon, missile   *InvItem // the weapon in hand and the one shot (nil: none)
	spells, maneuvers []Spell  // nil: none
}

// fightSource is what the live fight is read from: the game's State (its
// build number) and gobox's own part of it, the aim, its cursor and the log
// (rules.combat folds them in).
type fightSource struct {
	version stateBuild
	aim     *squareAim
	ats     int
	cursor  rules.BoardPos
	log     int
}

// fightSource is the source of the fight now.
func (p *Presenter) fightSource() fightSource {
	r := p.eng.r
	k := fightSource{version: p.eng.r.build(), aim: r.aim, cursor: r.cursor, log: len(r.log)}
	if r.aim != nil {
		k.ats = len(r.aim.ats)
	}
	return k
}

// boardKey is what the drawn fight board is made from: it is built again
// only when one of them changes (not every frame; the game is turn-based).
type boardKey struct {
	live   fightSource  // the game's State and gobox's part of the fight
	board  *game.Combat // the live fight, or the replayed moment's
	fly    *flight      // a missile's square
	queued int          // the moments still to play (keepFallen)
	ask    AskKind      // the question over the board (-1: none): an aim or a move
}

// drawnBoard makes fightView and fightBoard from p.board, when what they
// are made from has changed.
func (p *Presenter) drawnBoard() {
	k := boardKey{live: p.fightSource(), board: p.board, fly: p.fly, queued: len(p.replay.q), ask: -1}
	if q := p.askUp(); q != nil {
		k.ask = q.Kind
	}
	d := p.lay.board.get(k, func() (d fightDrawn) {
		d.view = p.keepFallen(p.board)
		p.fightView = d.view // aimPreview and askedBoard read it
		if t := d.view.Turn; t != nil {
			d.lists.weapon, d.lists.missile = weapon(t.Weapon), weapon(t.Missile)
			for _, m := range t.Maneuvers {
				d.lists.maneuvers = append(d.lists.maneuvers, Spell{ID: m.ID, Name: m.Name})
			}
			for _, sp := range d.view.Spells {
				d.lists.spells = append(d.lists.spells, Spell{ID: sp.ID, Name: sp.Name})
			}
		}
		d.board = p.askedBoard(p.aimPreview(p.moveMarks(p.boardOf(d.view, p.boardParty, p.fly))), d.view)
		return d
	})
	p.fightView, p.fightLists, p.fightBoard = d.view, d.lists, d.board
}

// fightDrawn is what the drawn fight is made of (drawnBoard).
type fightDrawn struct {
	view  *game.Combat  // the board with the fallen kept until their fall plays
	lists lists         // the turn's weapons, spells and maneuvers
	board *combat.Board // view as the tabletop draws it, aims put in
}

// moveMarks is b with the walk shown (CanMove: the reach, the hovered
// path) only while a party member has its turn and may walk: not on a foe's
// turn, and not while what happened before is still played (the turn that
// ended).
func (p *Presenter) moveMarks(b *combat.Board) *combat.Board {
	if b == nil {
		return nil
	}
	b.CanMove = false
	if p.Playing() || b.Turn == nil || len(b.Reach) == 0 {
		return b
	}
	for _, u := range b.Units {
		b.CanMove = b.CanMove || u.Active && u.Party
	}
	return b
}

// aimPreview is b with the aimed spell's area washed round the hovered
// square, while it has none of its own (the live board; a replayed one never
// aims).
func (p *Presenter) aimPreview(b *combat.Board) *combat.Board {
	if p.eng == nil || p.Playing() || p.eng.r.aim == nil || len(b.AreaAt) != 0 {
		return b
	}
	if _, ok := p.eng.r.aim.Action.(game.Cast); ok {
		b.AimArea = p.eng.r.aim.Area + 1
	}
	return b
}

// rowsKey is what the combat rows show: their text and its glyphs.
type rowsKey struct {
	text   string
	glyphs glyphGen
}
