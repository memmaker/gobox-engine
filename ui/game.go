package ui

import (
	"cmp"
	"errors"
	"fmt"
	"gobox/rules"
	"image"
	"image/color"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"gobox/game"
)

// Game is a *classicGame built from a game's rules and data alone: View comes
// from the classic builders over State, every rule call and Action becomes
// the Decision the game takes. It holds no game code.
type Game struct {
	G game.Game
	D game.Data
	r *frontState // what gobox holds over G

	events     []game.Report // the fight's events since the last Settle
	lastCombat *game.Combat  // the fight before the step: its events play on it (Settle)
	ended      *game.Combat  // that fight as the ending events so far left it
	lastParty  []game.Member // the party before the step: the events' figures are theirs (a fight the party lost leaves none)
	resets     int
	traceMode  game.Mode // the mode last traced
}

// aiming is the squares picked so far for spell, up to n.
type aiming struct {
	spell, n int
	ats      []image.Point
}

// NewGame is the *classicGame over p and d.
func NewGame(p game.Game, d game.Data) *Game {
	GameID = d.ID
	return &Game{G: p, r: newRules(p, d), D: d}
}

// do hands d to the rules, keeping its events; the refusal is the error.
func (g *Game) do(a game.Action) error {
	return g.run(fmt.Sprintf("%T%+v", a, a), func() game.Result { return g.r.apply(a) })
}

// run does one step, the game's or the front's own, and takes its result.
func (g *Game) run(what string, step func() game.Result) error {
	if c := g.combat(); c != nil { // the fight as the decision found it: a Result that ends it plays on it (Settle)
		g.lastCombat = c
	}
	g.lastParty = g.party()
	was := g.pane()
	paneWas, linesWere := paneTail(was), paneCount(was)
	done := g.r.held.Change() // the step may write the saves
	g.r.ran++
	res := step()
	done()
	g.dropLook()
	g.trace(what, res)
	if g.combat() == nil && g.itemsUp() { // over the items the text pane is hidden: its new line (BR2's "Doc buys a ...", "not enough credits") is the toast
		if t := paneTail(g.pane()); t != "" && (t != paneWas || paneCount(g.pane()) != linesWere) { // a new line, or the same one said again
			g.r.toast = FrontNote{Text: t, At: time.Now()}
		}
	}
	g.events = append(g.events, res.Events...)
	for _, e := range res.Events {
		if e.Kind == game.EvReset {
			g.resets++
		}
		if e.Kind == game.EvMessage && strings.TrimSpace(e.Text) != "" && g.combat() == nil && g.itemsUp() { // over the items (the shop's "Brann buys a Broad Sword"): the toast on its bottom row, as a refusal's; the inventory plays no message
			g.r.toast = FrontNote{Text: strings.TrimSpace(e.Text), At: time.Now()}
		}
	}
	if r, ok := res.Err.(*game.Refusal); ok && r.Ask != nil { // the question to put instead
		g.r.refusalAsk = r.Ask
		return nil
	}
	if res.Err != nil && res.Err.Error() != "" { // a refusal always says why, whichever screen asked
		g.Refuse(res.Err.Error())
		if _, ok := res.Err.(*game.Refusal); ok && g.combat() != nil { // in a fight also the message row's, held as a message (POR's "Blocked")
			g.events = append(g.events, game.Report{Kind: game.EvMessage, Text: res.Err.Error()})
		}
	}
	return res.Err
}

// sel is the selected member's index for n < 0.
func (g *Game) sel(n int) int {
	if n < 0 {
		return g.selected()
	}
	return n
}

// Send applies d to the front state and the game, for scripted drives (cmd
// tests); the player's path is Answer.
func (g *Game) Send(a game.Action) game.Result { return g.r.apply(a) }

// Step goes on with what gobox paces, for scripted drives.
func (g *Game) Step() game.Result { return g.r.Step() }

func (g *Game) step() error {
	res := g.r.Step()
	if len(res.Events) > 0 || res.Err != nil {
		g.trace("step", res)
	}
	g.events = append(g.events, res.Events...)
	g.dropLook()
	return res.Err
}

// dropLook ends a look-only sheet the fight no longer holds (it ended, or
// the member is gone), so it does not come back with the next fight.
func (g *Game) dropLook() {
	if _, ok := g.r.lookMember(g.r.g.State()); !ok {
		g.r.lookSheet = 0
	}
}

// View is the page classic builds from State.
// art is the rules' art as it is now, else the Data's.
// Subject is the member whose portrait and icon are on show, as the rules
// state it: the draft, else the selected one.
func (g *Game) Subject() game.Member {
	ps := g.r.g.State()
	return OwnSubject(g.r.modeOf(ps), creationOf(ps), partyOf(ps), g.r.selectedOf(ps))
}

// PickerOf is the game's grid picker, page page on cols by rows cells.
func (g *Game) PickerOf(k *game.Picker, page, cols, rows int) *FrontPicker {
	return Picker(k, page, cols, rows)
}

// TilesOf is the weapon's pool bodies' picker.
func (g *Game) TilesOf(page, cols, rows int) *FrontPicker {
	return TilePicker(g.Subject(), page, cols, rows)
}

// Art is the game's art as it is now.
func (g *Game) Art() map[string]game.Picture { return g.art() }

func (g *Game) art() map[string]game.Picture {
	return g.r.Art()
}

// Settle goes on to the next wait and returns what happened since, as the
// moments the presenter plays.
func (g *Game) Settle() []Moment {
	if g.pace() == game.PaceNone { // a paced step is the front's call (CombatStep ..)
		if c := g.combat(); c != nil && g.lastCombat == nil { // a fight already up before its first step (a dev start): its events play from it too
			g.lastCombat = c
		}
		g.lastParty = g.party()
		g.step()
	}
	posed := g.posedBoard
	for _, e := range g.events {
		if e.Kind == game.EvEngaged && e.Combat != nil { // the fight began in the step: its events play from its first board
			g.lastCombat, g.lastParty = e.Combat, e.Party
		}
	}
	if g.lastCombat != nil { // the events play from the board before the step, each on the board the ones before it left (one step is many turns: the dead stand until they fall, a walk starts where it starts); a fight they ended: the page after them
		c := *g.lastCombat
		c.Units, c.Squares = slices.Clone(c.Units), slices.Clone(c.Squares)
		g.ended = &c
		posed = g.stepBoard
	}
	ms := g.moments(g.events, g.combat() != nil, posed, g.messageMs())
	g.events, g.lastCombat = nil, g.combat()
	g.loreNews()
	return ms
}

// loreNews puts "Added to journal" up when the journal gained a note: gobox
// says so, the game prints nothing. The first look, a load or a new game is
// no news.
func (g *Game) loreNews() {
	n := len(game.CommonOf(g.r.g.State()).Lore) + 1
	if g.r.lore > 0 && n > g.r.lore {
		g.r.toast = FrontNote{Text: "Added to journal", At: time.Now()}
	}
	g.r.lore = n
}

// A combat step holds minStepMs and stepShare of the message delay: the
// game speed paces a walk as it paces the messages, the fastest still
// walks (extra: the originals redrew the figure on its next square).
const (
	minStepMs = 60
	stepShare = 0.3
)

// stepMs is how long a combat step holds, for a message delay of holdMs.
func stepMs(holdMs int) int { return minStepMs + int(float64(holdMs)*stepShare) }

// messageMs is how long a fight's message holds: the game speed × 100 ms,
// the rules' speed where they keep one (Speed), else g.speed (SetSpeed).
func (g *Game) messageMs() int {
	return g.r.Speed() * 100
}

// stepBoard is the fight as the step found it (g.ended), for each of its
// events in turn: every one is held on the board the events before it left
// (a walk moved, a fall dropped), with its unit in its pose; and the
// party's figures then.
func (g *Game) stepBoard(e game.Report) (*game.Combat, []game.Member) {
	if e.Kind == game.EvMove && !e.Turn { // a step: held on the square it reaches
		for k := range g.ended.Units {
			if u := &g.ended.Units[k]; u.ID == e.Unit {
				u.At = e.To
			}
		}
	}
	if b := e.Picture; e.Kind == game.EvDown && hasPic(b) { // the body the game leaves (Event.Picture): the fall's moment shows it in the unit's place (standing on it, the figure hid it)
		for k := range g.ended.Units {
			u := &g.ended.Units[k]
			if u.ID != e.Unit {
				continue
			}
			g.ended.Squares = withOverlay(g.ended.Squares, u.At, b)
			u.Dead, u.Gone, u.HP = true, true, 0
		}
	}
	c := *g.ended
	c.Units = slices.Clone(c.Units)
	for k := range c.Units {
		if c.Units[k].ID == e.Unit && e.Pose > 0 {
			c.Units[k].Pose = e.Pose
		}
	}
	party := g.lastParty
	for k := range g.ended.Units { // the board after e, for the events that follow
		switch u := &g.ended.Units[k]; {
		case u.ID == e.Unit && e.Kind == game.EvMove && e.Turn:
			u.Facing = e.N
		case u.ID == e.Unit && e.Kind == game.EvMove:
			u.At = e.To
		case u.ID == e.Unit && e.Kind == game.EvDown:
			u.Dead, u.Gone, u.HP = true, true, 0 // off the map, as EvDied plays it
		}
	}
	return &c, party
}

// hasPic is whether p is a picture at all.
func hasPic(p game.Picture) bool { return p.File != "" || p.Name != "" || p.Pixels != nil }

// withOverlay is sq with o lying on the square at.
func withOverlay(sq []game.Ground, at image.Point, o game.Picture) []game.Ground {
	if i := slices.IndexFunc(sq, func(q game.Ground) bool { return q.At == at }); i >= 0 {
		sq[i].Overlay = o
		return sq
	}
	return append(sq, game.Ground{At: at, Overlay: o})
}

// moments are the step's reports as the presenter plays them; the kinds
// with no moment to play (pictures) are left out. Outside fights too: every
// step's events come through here (songs, sounds). posed is the board an
// event holds on, and the party's figures then.
func (g *Game) moments(es []game.Report, fight bool, posed func(game.Report) (*game.Combat, []game.Member), holdMs int) (out []Moment) {
	for _, e := range es {
		if e.Screen != nil { // a timed typed screen: the scene as it stood
			out = append(out, Moment{Report: e, Ms: e.Ms})
			continue
		}
		if !momentKind(e.Kind) {
			continue
		}
		m := Moment{Report: e}
		hold := func(ms int) { m.Ms = ms; m.Board, m.Party = posed(e) }
		switch {
		case !fight: // the fight over: every event on the board the ones before it left
			if b, p := posed(e); e.Ms > 0 {
				m.Ms, m.Board, m.Party = e.Ms, b, p
			}
		case e.Kind == game.EvDown && hasPic(e.Picture): // a fall that leaves a body is held with the body down (with no moment, the shot before it kept the figure standing up to the page)
			hold(cmp.Or(e.Ms, holdMs))
		case e.Kind == game.EvMove && !e.Turn: // a step held on the board it leaves, so a walk plays square by square (the figure walks it: combat.Presenter.StepTime)
			hold(stepMs(holdMs))
		case e.Pose > 0 && e.Ms > 0: // the moment held with the unit in its pose
			hold(e.Ms)
		case e.Text != "" && e.Ms == 0: // a fight's message stays up the game-speed delay before what follows, as the originals'
			hold(holdMs)
		}
		if e.Kind == game.EvMissile && len(e.Sprite) > 0 && m.Board == nil && fight { // a shot the game gave no moment of its own: it flies on the board as it stands
			hold(cmp.Or(e.Ms, flightMs))
		}
		out = append(out, m)
	}
	return out
}

// momentKind is whether a report of kind k is a moment to play.
func momentKind(k game.EventKind) bool {
	switch k {
	case game.EvAttack, game.EvHit, game.EvMiss, game.EvDamage, game.EvDown, game.EvSpell, game.EvMove, game.EvMissile,
		game.EvMessage, game.EvSound, game.EvSong, game.EvSaves:
		return true
	}
	return false
}

// posedBoard is the fight as it stands with e's Unit in e's Pose, and the
// party now (nil: no fight).
func (g *Game) posedBoard(e game.Report) (*game.Combat, []game.Member) {
	now := g.combat()
	if now == nil {
		return nil, nil
	}
	c := *now
	c.Units, c.Squares = slices.Clone(c.Units), slices.Clone(c.Squares)
	for k := range c.Units {
		if c.Units[k].ID == e.Unit {
			c.Units[k].Pose = e.Pose
		}
	}
	if e.Kind == game.EvDown && hasPic(e.Picture) && g.lastCombat != nil { // the body on the square the unit fell on, it off the board
		for _, u := range g.lastCombat.Units {
			if u.ID == e.Unit {
				c.Squares = withOverlay(c.Squares, u.At, e.Picture)
			}
		}
		c.Units = slices.DeleteFunc(c.Units, func(u game.Unit) bool { return u.ID == e.Unit })
	}
	return &c, g.party()
}

// Open starts the rules on save slot slot ("" none) when they take it.
func (g *Game) Open(slot string) {
	if o, ok := g.r.g.(interface{ Open(string) }); ok {
		o.Open(slot)
	}
}

func (g *Game) Resets() int       { return g.resets }
func (g *Game) Refuse(s string)   { g.r.toast = FrontNote{Text: s, At: time.Now()} }
func (g *Game) CombatStep() error { return g.step() }
func (g *Game) RestStep() error   { return g.step() }
func (g *Game) StoryStep() error  { return g.step() }
func (g *Game) WalkStep() error   { return g.step() }
func (g *Game) StopWalk() error   { return g.run("stopWalk", g.r.stopWalk) }

// Party menu.
func (g *Game) EndGame() error { return g.do(game.EndGame{}) }

// Creation sheet.

// CreateCharacter, Roll and Rename stay exported for the Buck Rogers repos' front tests (buckrogers-go, buckrogers2-go), which have no other way to send an action and read its refusal.
func (g *Game) CreateCharacter() error { return g.do(game.NewCharacter{}) }
func (g *Game) Roll(minimal bool, minimums string) error {
	return g.do(game.Reroll{Minimal: minimal})
}
func (g *Game) Rename(name string) error { return g.do(game.Rename{Name: name}) }
func (g *Game) SetField(f, o int) error {
	return g.run("setField", func() game.Result { return g.r.setField(f, o) })
}
func (g *Game) EndCreation(keep bool) error { return g.endCreation(keep) }
func (g *Game) CreationView() *FrontCreation {
	return g.CreationOn(g.r.grid.X, g.r.grid.Y)
}

// Title: the presenter plays it from State.Intro, the game draws it.
func (g *Game) TitlePicture(n int) {
	g.run("titlePicture", func() game.Result { return g.r.titlePicture(n) })
}
func (g *Game) TitleCredits(page int) {
	g.run("titleCredits", func() game.Result { return g.r.titleCredits(page) })
}
func (g *Game) TitleDone() { g.run("titleDone", g.r.titleDone) }

// SequenceDone: the presenter played State.Sequence to its end, or it was skipped.
func (g *Game) SequenceDone(name string) { g.run("sequenceDone", g.r.sequenceDone) }

// Camp.

// RestFor rests minutes long (a RestKind.Duration rest's picked time).
func (g *Game) RestFor(minutes int) error {
	return g.run("rest", func() game.Result { return g.r.restFor(minutes) })
}

// Powers.
func (g *Game) UsePower(n int, k ActKind, spell, target int, sure bool) error {
	return g.asking(g.do(game.UsePower{Member: n, Kind: frontPower[int(k)], Spell: spell, Target: target}))
}
func (g *Game) SelectedMember() int { return g.selected() }

// EffectsList is member n's Effects lines from State.
func (g *Game) EffectsList(n int) (out []FrontEntry) {
	party := g.party()
	if n = g.sel(n); n < 0 || n >= len(party) {
		return nil
	}
	for _, l := range party[n].Effects {
		out = append(out, FrontEntry{Text: l.Text, Header: l.Header, Dim: l.Dim, Off: l.Off, Spell: l.Spell})
	}
	return out
}

// Waits.
func (g *Game) LootView(n int) error { return error(nil) }
func (g *Game) LootTake(n int) error { return errors.New("no such action") }
func (g *Game) LootLeave(sure bool) error {
	err := g.do(game.Leave{})
	if a := g.r.refusalAsk; a != nil && slices.ContainsFunc(a.Options, func(o game.Option) bool { return o.Do != nil }) { // the game's own answers (POR: Yes stays, No leaves Sure): its question as it is
		return nil
	}
	if a := g.r.refusalAsk; a != nil && (a.Kind == game.YesNo || a.Kind == game.Choice) { // no answers of its own: the presenter asks it as a Confirm, Yes leaving (Silver Blades and Pools of Darkness state it as a Yes/No Choice)
		g.r.refusalAsk = nil
		return Confirm(a.Prompt)
	}
	var r *game.Refusal
	if !sure && errors.As(err, &r) && r.Text != "" { // the game's own question as its text (POR: "...go back and claim your treasure?"): Yes stays, No leaves
		return ConfirmNo(r.Text)
	}
	return err
}

// Pickers and the icon designer.
func (g *Game) OpenPicker(kind int) bool {
	ok := g.do(game.OpenPicker{Kind: kind}) == nil
	if kind == PickWeapon && game.CommonOf(g.r.g.State()).Picker == nil {
		if n, _ := BodyTiles(g.Subject()); len(n) > 0 { // no pictures from the game: the pool's, by original ref
			g.r.tiles = true
			return true
		}
	}
	return ok
}
func (g *Game) PickerView(page int) *FrontPicker {
	if g.r.tiles {
		return g.TilesOf(page, g.r.grid.X, g.r.grid.Y)
	}
	return g.PickerOf(game.CommonOf(g.r.g.State()).Picker, page, g.r.grid.X, g.r.grid.Y)
}
func (g *Game) Pick(i int) bool {
	if g.r.tiles {
		t := PickTile(g.Subject(), i)
		if t == "" || g.SetChoice("iconbody", t) != nil {
			return false
		}
		g.r.tiles = false
		return true
	}
	g.do(game.Pick{N: i})
	return game.CommonOf(g.r.g.State()).Picker == nil
}

// SetChoice sends gobox's pool pick field (the portrait's or the icon's) to
// the game: it takes it for the selected member or the draft.
func (g *Game) SetChoice(field, id string) error {
	_, err := g.r.g.Do(game.SetChoice{Field: field, Text: id})
	return err
}
func (g *Game) ClosePicker()         { g.r.tiles = false; g.do(game.ClosePicker{}) }
func (g *Game) EditIcon(n int) error { return g.do(game.EditIcon{Member: n}) }
func (g *Game) IconSize()            { g.do(game.IconSize{}) }
func (g *Game) IconColour(part, shade, v int) {
	g.do(game.IconColour{Part: part, Shade: shade, Value: v})
}
func (g *Game) IconRemove()         { g.do(game.IconRemove{}) }
func (g *Game) IconPNG(path string) { g.do(game.IconPNG{Path: path}) }
func (g *Game) EndIcon(keep bool)   { g.do(game.EndIcon{Keep: keep}) }

// SlotsView is State.Saves as the slot list (nil on the title).
func (g *Game) SlotsView(loadOnly bool) *Slots {
	ps := g.r.g.State()
	c := game.CommonOf(ps)
	if introOf(ps) != nil {
		return nil
	}
	mode := g.r.modeOf(ps)
	inGame := len(c.Party) > 0 && mode != game.ModeTitle && mode != game.ModeCreate && mode != game.ModeCombat
	canSave := !loadOnly && inGame && !c.NoSave
	if !canSave && c.NoLoad { // neither offered here: no slot screen
		return nil
	}
	v := &Slots{CanSave: canSave, LoadOnly: loadOnly, NoLoad: c.NoLoad}
	for _, sl := range g.r.saves() {
		v.Names, v.Times, v.Levels = append(v.Names, sl.Name), append(v.Times, sl.Time), append(v.Levels, sl.Level)
		v.Overwrites = append(v.Overwrites, sl.Overwrite)
	}
	if len(v.Names) > 0 {
		v.Current = v.Names[0]
	}
	return v
}
func (g *Game) Save(name string) error {
	return g.run("save", func() game.Result { return g.r.saveNamed(name) })
}
func (g *Game) DeleteSlot(name string) error {
	return g.run("deleteSlot", func() game.Result { return game.Result{Err: g.r.remove(name)} })
}
func (g *Game) Load(name string, loadOnly bool) bool {
	for i, sl := range g.r.saves() {
		if sl.Name == name {
			return g.run("load", func() game.Result { return g.r.loadSlot(i) }) == nil
		}
	}
	return false
}

// TurnChoice is the combat Action as its Decision (the unsure ask is an Ask).
func (g *Game) TurnChoice(a Action, sure bool) error { return g.Answer(a) }

// Sheet and inventory.
// OpenSheet: in a fight the sheet is look-only, gobox's from Member.Page
// (n < 0: the turn's member); the game is not called. Else the game's own.
func (g *Game) OpenSheet(n int) error {
	c := g.combat()
	if c == nil {
		return g.run("openSheet", func() game.Result { return g.r.openSheet(n) })
	}
	if n < 0 {
		n = g.selected()
		if t := c.Turn; t != nil {
			for _, u := range c.Units {
				if u.ID == t.Unit && u.Member >= 0 {
					n = u.Member
				}
			}
		}
	}
	if n < 0 || n >= len(g.party()) {
		return nil
	}
	g.r.lookSheet = n + 1
	return nil
}

func (g *Game) CloseSheet() error {
	if g.r.lookSheet > 0 {
		g.r.lookSheet = 0
		return nil
	}
	return g.run("closeSheet", g.r.closeSheet)
}
func (g *Game) ShowMember(n int) error {
	if g.r.lookSheet > 0 && n >= 0 && n < len(g.party()) { // another member's look-only sheet
		g.r.lookSheet = n + 1
		return nil
	}
	return g.run("showMember", func() game.Result { return g.r.showMember(n) })
}

// CloseItems: Tab (all false) is the sheet, as out of a fight; in a fight
// that is the look-only sheet of the member whose items were up.
func (g *Game) CloseItems(all bool) error {
	if g.combat() != nil && !all {
		if err := g.run("closeItems", func() game.Result { return g.r.closeItems(true) }); err != nil {
			return err
		}
		return g.OpenSheet(g.selected())
	}
	return g.run("closeItems", func() game.Result { return g.r.closeItems(all) })
}
func (g *Game) UseItem(h rules.ItemHandle, spell, target int, sure bool) error {
	return g.asking(g.do(game.UseItem{Member: g.sel(-1), Item: h, Spell: spell, Target: target}))
}
func (g *Game) Buy(h rules.ItemHandle) error {
	if lootOf(g.r.g.State()) != nil { // the loot screen's Take, as the act mapping does
		return g.do(game.Take{Item: h})
	}
	return g.do(game.Buy{Member: g.sel(-1), Item: h})
}

// Shops, temples, training.

// Appraise values the purse's gem or piece of jewelry: the offer of that kind.
func (g *Game) Appraise(jewelry bool) error {
	kind := game.Gems
	if jewelry {
		kind = game.Jewelry
	}
	for _, o := range g.offers() {
		if o.Kind == kind {
			return g.do(game.Appraise{Member: o.Member, Offer: o.Offer})
		}
	}
	return errors.New("no " + kind + " to appraise")
}

// KeepAppraised keeps or sells the one valued: the offer with a value.
func (g *Game) KeepAppraised(keep bool) error {
	for _, o := range g.offers() {
		if o.Value > 0 {
			return g.do(game.Appraise{Member: o.Member, Offer: o.Offer, Keep: keep})
		}
	}
	return errors.New("nothing valued to keep or sell")
}

// offers are the shop's or temple's appraisals.
func (g *Game) offers() []game.Offer {
	if t, ok := g.r.g.State().(game.Trading); ok {
		return t.Offers
	}
	return nil
}

func (g *Game) TrainingView() *FrontTraining {
	ps := g.r.g.State()
	return Training(trainingOf(ps), partyOf(ps), g.r.selectedOf(ps), game.CommonOf(ps).Purse)
}

// The full map: its cursor and label are the presenter's, classic fills the rest.
func (g *Game) FullMapView(cursor image.Point, label *string) *FrontFullMap {
	return g.FullMapOf(&FrontFullMap{Cursor: cursor})
}
func (g *Game) Mark(x, y int, text string) { g.do(game.Mark{At: image.Pt(x, y), Text: text}) }
func (g *Game) Unmark(x, y int)            { g.do(game.Unmark{At: image.Pt(x, y)}) }
func (g *Game) WalkTo(x, y int) bool {
	return g.run("walkTo", func() game.Result { return g.r.walkTo(g.r.g.State(), image.Pt(x, y)) }) == nil
}

// JournalPage is the journal text laid out on a page of cells (Data.Scans has its pictures).
func (g *Game) JournalPage(title string, texts []game.ReadText, p, off int, cells image.Point) (*Pages, int) {
	return JournalPage(title, texts, p, off, cells.X, cells.Y, [16]color.RGBA{}, g.D.Scans)
}

// SpeakPage has the game voice page p of the text laid out on a page of cells, in place of the page voiced before.
func (g *Game) SpeakPage(texts []game.ReadText, p int, cells image.Point) {
	g.run("speak", func() game.Result { return g.r.speak(SpeakKeys(texts, g.D.Scans, p, cells.X, cells.Y)) })
}

// ClueChapters and ClueRead are Data.Clues.
func (g *Game) ClueChapters() (out []string) {
	for _, c := range g.D.Clues {
		out = append(out, c.Title)
	}
	return out
}
func (g *Game) ClueRead(i int) *Read {
	if len(g.D.Clues) == 0 {
		return nil
	}
	r := g.D.Clues[min(max(i, 0), len(g.D.Clues)-1)].Read
	return &Read{Title: r.Title, Texts: r.Texts, Page: r.Page}
}

// Answer does the Decision a names; ActResize is the front's own (the page's size).
func (g *Game) Answer(a Action) error {
	if traceOn {
		fmt.Fprintf(os.Stderr, "trace %s act %+v\n", traceStamp(), a)
	}
	if a.Kind == ActResize {
		err := error(nil)
		g.r.grid = image.Pt(max(40, a.X), max(25, a.Y)) // the window's grid at K: the screens the presenter lays out fill it
		return err
	}
	if m := g.r.spellAim; m != nil { // up to n squares, then the cast; Enter casts on those picked, Esc gives up
		switch {
		case a.Kind == ActAim:
			if m.ats = append(m.ats, image.Pt(a.X, a.Y)); len(m.ats) < m.n {
				return nil
			}
		case a.Kind == ActOK && len(m.ats) > 0:
		case a.Kind == ActClose:
			g.r.spellAim, g.r.refusalAsk = nil, nil
			return nil
		default:
			return nil
		}
		g.r.spellAim, g.r.refusalAsk = nil, nil
		return g.do(game.Cast{Member: -1, Spell: m.spell, Ats: m.ats})
	}
	if announced(g.r.g.State()) && g.r.refusalAsk == nil {
		if a.Kind != ActOK && a.Kind != ActClose {
			return nil
		}
		if err := g.do(game.ContinueAction{}); err != nil || a.Kind == ActOK { // Esc passes the found page and leaves at once, as Enter then Esc
			return err
		}
	}
	if a.Kind == ActCast {
		if c := g.combat(); c != nil {
			for _, sp := range c.Aims {
				if sp.ID == a.N {
					g.r.spellAim = &aiming{spell: sp.ID, n: max(sp.Targets, 1)}
					g.r.refusalAsk = &game.Ask{Kind: game.Square, Prompt: "Aim " + sp.Name}
					return nil
				}
			}
			return g.do(game.Cast{Member: -1, Spell: a.N})
		}
	}
	if q := g.r.refusalAsk; q != nil { // a Refusal's question: an option's Do, or anything else drops it
		g.r.refusalAsk = nil
		i := yesNoOption(q, a.Kind) // a Yes/No with its own options: the one labelled so
		if (a.Kind == ActChoose || a.Kind == ActOK) && a.N >= 0 && a.N < len(q.Options) {
			i = a.N
		}
		if i >= 0 { // an option without a Do (No) takes nothing
			if d := q.Options[i].Do; d != nil {
				return g.do(d)
			}
			return nil
		}
		if a.Kind == ActClose {
			return nil
		}
	}
	if _, ok := walks[a.Kind]; ok { // gobox moves the party on the game's Map
		return g.walk(a.Kind)
	}
	if a.Kind == ActGo {
		return g.goWay(a.N)
	}
	step := g.act(a)
	if step == nil {
		Unhandled(a)
		return nil
	}
	return step()
}

// act is the step a asks for (nil: none).
func (g *Game) act(a Action) func() error {
	send := func(x game.Action) func() error { return func() error { return g.do(x) } }
	own := func(what string, step func() game.Result) func() error {
		return func() error { return g.run(what, step) }
	}
	at := image.Pt(a.X, a.Y)
	m := g.sel(a.N)
	if q := g.pending(); q != nil && q.Kind == game.Square && a.Kind == ActAim { // the square answers the square question, in any mode (a fight's Target asks in ModeCombat)
		return own("aim", func() game.Result { return g.r.answer(g.r.g.State(), game.Reply{}, at) })
	}
	if q := g.pending(); q != nil && g.mode() == game.ModeAsk { // a direction answers the question
		switch {
		case q.View && a.Kind == ActView: // the highlighted member (N, or the selected) is viewed; the ask stays
			return send(game.ViewMember{Member: a.N})
		case q.Kind == game.Direction && a.Kind == ActDir && a.N >= 1 && a.N <= 9 && a.N != 5:
			return send(game.Reply{N: [10]int{1: 5, 2: 4, 3: 3, 4: 6, 6: 2, 7: 7, 8: 0, 9: 1}[a.N]}) // keypad 8 up: facing 0 north, 2 east ...
		}
	}
	switch a.Kind {
	case ActOption: // the turn's option N, as the game offers it
		if ops := turnOptions(g.combat(), g.D); a.N >= 0 && a.N < len(ops) && ops[a.N].Do != nil {
			return send(ops[a.N].Do)
		}
		return nil
	case ActClose:
		return send(game.Cancel{})
	case ActMainMenu:
		return send(game.EndGame{})
	case ActNote:
		return send(game.AddNote{Text: paneText(g.pane())})
	case ActNoteRemove:
		return send(game.RemoveNote{ID: a.N})
	case ActOrder:
		return send(game.Order{From: a.N, To: a.X})
	case ActOK, ActChoose:
		return send(game.Reply{N: a.N})
	case ActYes, ActNo:
		if i := yesNoOption(g.pending(), a.Kind); i >= 0 { // the ask's own Yes and No options: their indexes
			return send(game.Reply{N: i})
		}
		return send(game.Reply{N: map[ActKind]int{ActYes: 1}[a.Kind]}) // no options: Yes 1, No 0
	case ActText: // a number typed is N as well (a Number ask's amount), the Text kept
		n, _ := strconv.Atoi(strings.TrimSpace(a.Text))
		return send(game.Reply{N: n, Text: a.Text})
	case ActSearch:
		return send(game.Search{})
	case ActCamp:
		return send(game.MakeCamp{})
	case ActLead:
		return send(game.SelectMember{Member: a.N})
	case ActMember:
		if up := g.mode(); up == game.ModeTemple || up == game.ModeShop || shopOf(g.r.g.State()) != nil { // the shop's and the temple's View (up whatever the mode)
			return send(game.ViewMember{Member: m})
		}
		return own("showMember", func() game.Result { return g.r.showMember(a.N) })
	case ActInventory:
		return send(game.OpenItems{Member: a.N})
	case ActSheet:
		return own("closeItems", func() game.Result { return g.r.closeItems(false) })
	case ActTrade:
		return send(game.Give{From: g.sel(-1), To: a.X, Item: a.Item})
	case ActHalve:
		return send(game.Halve{Member: g.sel(-1), Item: a.Item})
	case ActJoin:
		return send(game.Join{Member: g.sel(-1), Item: a.Item})
	case ActReady:
		return send(game.Ready{Member: g.sel(-1), Item: a.Item, Count: a.X})
	case ActUse:
		return send(game.UseItem{Member: g.sel(-1), Item: a.Item})
	case ActDrop:
		return send(game.Drop{Member: g.sel(-1), Item: a.Item})
	case ActSell:
		return send(game.Sell{Member: g.sel(-1), Item: a.Item})
	case ActReload:
		return send(game.Reload{Member: g.sel(-1), Item: a.Item})
	case ActBuy:
		if lootOf(g.r.g.State()) != nil {
			return send(game.Take{Item: a.Item})
		}
		return send(game.Buy{Member: g.sel(-1), Item: a.Item})
	case ActTakeAll:
		return send(game.Take{Item: rules.NoItem, All: true})
	case ActTakeMoney:
		return send(game.Take{Item: rules.NoItem})
	case ActLootPool: // a shop's Pool (the treasure's goes through Gate)
		return send(game.PoolCoins{Member: g.sel(-1)}) // the selected member's coins
	case ActLootShare:
		return send(game.LootShare{})
	case ActWalk:
		return send(game.MoveTo{To: at})
	case ActAttack:
		return send(game.Attack{Target: a.N})
	case ActAim:
		return send(game.Aim{At: at})
	case ActGuard:
		return send(game.Guard{})
	case ActWait:
		return send(game.Delay{})
	case ActEnd:
		return send(game.EndTurn{})
	case ActFlee:
		return send(game.Flee{})
	case ActQuick:
		return send(game.Quick{All: a.N == 1})
	case ActBandage:
		return send(game.Bandage{Member: m})
	case ActTurn:
		return send(game.TurnUndead{})
	case ActRanged:
		return send(game.Shoot{})
	case ActItems:
		return send(game.OpenItems{Member: -1})
	case ActManeuver:
		return send(game.DoManeuver{})
	case ActRest:
		if a.X > 0 { // a set duration (RestKind.Duration): X minutes
			return own("rest", func() game.Result { return g.r.restFor(a.X) })
		}
		return send(game.CampRest{Kind: a.N})
	case ActPlan:
		return send(game.Plan{Member: m, Spell: a.X, Count: a.Y})
	case ActScribe:
		return send(game.Scribe{Member: m, Spell: a.X})
	case ActDropMember:
		return send(game.DropMember{Member: m})
	case ActIcon, ActPortrait: // the rules say which takes it: the icon, or the sheet's portrait
		return send(game.IconPNG{Path: a.Text})
	case ActIconEdit:
		return send(game.EditIcon{Member: a.N})
	case ActCreate:
		return send(game.NewCharacter{})
	case ActModify:
		return send(game.Modify{Member: m})
	case ActRemove:
		return send(game.RemoveCharacter{Roster: m})
	case ActTraining:
		return send(game.OpenTraining{Member: m})
	case ActNewClass:
		return send(game.ChangeClass{Member: m, Class: game.ClassAsked}) // the game asks which (its Refusal's Ask)
	case ActQuickStart:
		return send(game.QuickStart{})
	case ActCommand: // a game's own party command (Assembling.Commands)
		if as, ok := g.r.g.State().(game.Assembling); ok && a.N >= 0 && a.N < len(as.Commands) && as.Commands[a.N].Do != nil {
			return send(as.Commands[a.N].Do)
		}
	case ActBegin:
		return send(game.Begin{})
	case ActQuit:
		return send(game.Quit{})
	case ActRoll:
		return send(game.Reroll{Minimal: a.N == 1})
	case ActName:
		return send(game.Rename{Name: a.Text})
	case ActSetField:
		return own("setField", func() game.Result { return g.r.setField(a.N, a.X) })
	case ActStat:
		return send(game.SetStat{Stat: a.N, Text: a.Text})
	case ActKeep:
		if g.Subject(); Drafting(g.mode(), creationOf(g.r.g.State())) { // the creation sheet's Keep
			return send(game.KeepCharacter{})
		}
		return nil
	}
	return nil
}

// paneText is the message pane's words on one line: the note ActNote takes.
func paneText(p *TextPage) string {
	if p == nil {
		return ""
	}
	var w []string
	for _, l := range p.Pane {
		for _, r := range l {
			w = append(w, strings.Fields(r.Text)...)
		}
	}
	return strings.Join(w, " ")
}

func (g *Game) party() []game.Member { return partyOf(g.r.g.State()) }
func (g *Game) selected() int        { return g.r.selectedOf(g.r.g.State()) }
func (g *Game) combat() *game.Combat { return g.r.combatOf(g.r.g.State()) }
func (g *Game) pane() *TextPage      { return g.r.pageOf(g.r.g.State()) }
func (g *Game) pace() game.Pace      { return g.r.pacingOf(g.r.g.State()).Pace }
func (g *Game) itemsUp() bool        { return g.r.itemsUp(g.r.g.State()) != game.ItemsNone }

// pending is the question up, a Refusal's first.
func (g *Game) pending() *game.Ask {
	if g.r.refusalAsk != nil {
		return g.r.refusalAsk
	}
	return g.r.pending(g.r.g.State())
}

// mode is the screen up, a Refusal's question counted.
func (g *Game) mode() game.Mode {
	return g.r.modeOf(g.r.g.State())
}

// CreationOn is the creation sheet on a page of cols by rows.
func (g *Game) CreationOn(cols, rows int) *FrontCreation {
	ps := g.r.g.State()
	return Creation(g.r.modeOf(ps), creationOf(ps), g.r.pending(ps), cols, rows)
}

// IconEditorOn is the icon designer on a page of cols by rows.
func (g *Game) IconEditorOn(ink *[16]color.RGBA, cols, rows int) *FrontIconEditor {
	return IconEditor(game.CommonOf(g.r.g.State()).IconEditor, nil, g.Subject(), ink, cols, rows)
}

// FullMapOf is f with the map of where the party is.
func (g *Game) FullMapOf(f *FrontFullMap) *FrontFullMap {
	ps := g.r.g.State()
	return FullMap(f, g.r.derived().automap, game.CommonOf(ps).Where)
}

// Where is the party's place; Mode the screen up; Dungeon a fight's ground.
func (g *Game) Where() game.Where { return game.CommonOf(g.r.g.State()).Where }
func (g *Game) Mode() game.Mode   { return g.mode() }

// Asked is the question up (nil: none).
func (g *Game) Asked() *game.Ask { return g.pending() }

// asking is a Refusal's question of UsePower or UseItem as the front's
// pick or question (PickSpell, PickMember, Confirm), which the presenter
// asks before making the call again with the answer.
func (g *Game) asking(err error) error {
	a := g.r.refusalAsk
	if a == nil {
		return err
	}
	switch a.Kind {
	case game.PickSpell:
		var ps PickSpell
		for _, o := range a.Options {
			id := 0
			switch d := o.Do.(type) {
			case game.UsePower:
				id = d.Spell
			case game.UseItem:
				id = d.Spell
			}
			ps = append(ps, Spell{ID: id, Name: o.Label})
		}
		g.r.refusalAsk = nil
		return ps
	case game.PickMember:
		g.r.refusalAsk = nil
		return PickMember(a.Prompt)
	case game.YesNo:
		g.r.refusalAsk = nil
		return Confirm(a.Prompt)
	}
	return err
}

// yesNoOption is the index of q's option labelled Yes (for ActYes) or No
// (ActNo); -1 when q has none so labelled.
func yesNoOption(q *game.Ask, k ActKind) int {
	if q == nil || k != ActYes && k != ActNo {
		return -1
	}
	want := map[ActKind]string{ActYes: "yes", ActNo: "no"}[k]
	return slices.IndexFunc(q.Options, func(o game.Option) bool { return strings.EqualFold(strings.TrimSpace(o.Label), want) })
}

// flightMs is a shot's flight when the game sets no time (Ms, StepMs).
const flightMs = 300

// paneTail is the text pane's last non-empty line (the game's latest message), "" with none.
func paneTail(p *TextPage) string {
	if p == nil {
		return ""
	}
	for i := len(p.Pane) - 1; i >= 0; i-- {
		var b strings.Builder
		for _, r := range p.Pane[i] {
			b.WriteString(r.Text)
		}
		if t := strings.TrimSpace(b.String()); t != "" {
			return t
		}
	}
	return ""
}

func paneCount(p *TextPage) int {
	if p == nil {
		return 0
	}
	return len(p.Pane)
}

// PickerOn is page page of the picker up on a page of cols by rows (nil: none).
func (g *Game) PickerOn(page, cols, rows int) *FrontPicker {
	if p := g.PickerOf(game.CommonOf(g.r.g.State()).Picker, page, cols, rows); p != nil {
		return p
	}
	return g.PickerView(page)
}

// Draft is the character being made (nil: none).
func (g *Game) Draft() *game.Creation { return creationOf(g.r.g.State()) }

// endCreation keeps or discards the character being made.
func (g *Game) endCreation(keep bool) error {
	if keep {
		return g.do(game.KeepCharacter{})
	}
	return g.do(game.DiscardCharacter{})
}

// announced is whether the treasure's "You found some loot!" is up (game.Looting.Announce).
func announced(ps game.State) bool {
	l, ok := ps.(game.Looting)
	return ok && l.Announce
}
