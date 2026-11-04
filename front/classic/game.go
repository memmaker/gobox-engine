package classic

import (
	"image"
	"strconv"
	"strings"
	"time"

	"gobox/front"
	"gobox/game"
)

// Game is a front.Game built from a game's rules and data alone: View comes
// from the classic builders over State, every rule call and Action becomes
// the Decision the game takes. It holds no game code.
type Game struct {
	R game.Rules
	D game.Data

	grid   image.Point  // the page's size in cells (ActResize); 40x25 at least
	events []game.Event // the fight's events since the last Settle
	note   front.Note   // the front's own refusal toast
	resets int
	tiles  bool      // the weapon picker up is the pool bodies by original ref (TilePicker)
	asked  *game.Ask // a Refusal's question put instead (answered by its options' Do)
}

var _ front.Game = (*Game)(nil)

// NewGame is the front.Game over r and d.
func NewGame(r game.Rules, d game.Data) *Game {
	GameID = d.ID
	return &Game{R: r, D: d, grid: image.Pt(40, 25)}
}

// do hands d to the rules, keeping its events; the refusal is the error.
func (g *Game) do(d game.Decision) error {
	res := g.R.Apply(d)
	g.events = append(g.events, res.Events...)
	for _, e := range res.Events {
		if e.Kind == game.EvReset {
			g.resets++
		}
	}
	if r, ok := res.Err.(*game.Refusal); ok && r.Ask != nil { // the question to put instead
		g.asked = r.Ask
		return nil
	}
	return res.Err
}

// sel is the selected member's index for n < 0.
func (g *Game) sel(n int) int {
	if n < 0 {
		return g.R.State().Selected
	}
	return n
}

func (g *Game) step() error {
	res := g.R.Step()
	g.events = append(g.events, res.Events...)
	return res.Err
}

// View is the page classic builds from State.
func (g *Game) View() front.View {
	s := g.state()
	art := g.D.Art
	if r, ok := g.R.(interface {
		Art() map[string]game.Picture
	}); ok { // the rules' art as it is now
		art = r.Art()
	}
	v := front.View{Exploring: s.Mode == game.ModeExplore}
	if s.Map.Size == image.Pt(16, 16) {
		v.Map = &front.MapSquares{}
	}
	switch s.Items { // the builders refine a screen the game has up: it is up when State says so
	case game.ItemsOwn:
		v.Inventory = &front.Inventory{}
	case game.ItemsShop, game.ItemsLoot:
		v.Inventory = &front.Inventory{Wares: []front.InvItem{}}
	}
	if len(s.Map.Terrain) > 0 {
		v.AreaMap = &front.AreaMap{}
	}
	if s.Shop != nil {
		v.Shop = &front.Shop{}
	}
	if s.Temple != nil {
		v.Temple = &front.Temple{}
	}
	if s.Mode == game.ModeCamp {
		v.Camp = &front.Camp{}
	}
	v = View(v, s)
	if v.AreaMap != nil {
		v.AreaMap = AreaMap(v.AreaMap, s, art)
	}
	if v.MainMenu != nil && !v.Hall && v.Powers == nil {
		v.Logo, v.Backdrop = Title(v.Logo, v.Backdrop, art)
	}
	v.Battle = BattleOf(s, art)
	v = Page(v, s)
	v = Flags(v, s)
	v = Place(v, s)
	v.Quests, v.Lore = Quests(s), Lore(s)
	v = Waits(v, s)
	v.Party, v.Purse = Party(s, art), Purse(s)
	v = Sheet(v, s)
	v = Border(v, art)
	v = Walls(v, s, art)
	v = Intro(v, s)
	v = Note(v, s)
	v = Results(v, s)
	v = Screen(v, s)
	v = Read(v, s)
	v = Pacing(v, s)
	if v.Note.Text == "" {
		v.Note = g.note
	}
	v.Creation = Creation(s, g.grid.X, g.grid.Y)
	v = PoolIcon(v, s)
	v.Training = Training(s)
	v.Picker = Picker(s, 0, g.grid.X, g.grid.Y)
	v.IconEditor = IconEditor(s, &v.Ink, g.grid.X, g.grid.Y)
	return v
}

// Settle goes on to the next wait; the View carries the fight's events since.
func (g *Game) Settle() front.View {
	if p := g.R.State().Pacing; p == nil || p.Pace == game.PaceNone { // a paced step is the front's call (CombatStep ..)
		g.step()
	}
	v := g.View()
	v.Events = combatEvents(g.events)
	g.events = nil
	return v
}

// combatEvents are the fight's events as the presenter plays them; the
// kinds a fight has no moment for (pictures, songs) are left out.
func combatEvents(es []game.Event) (out []front.CombatEvent) {
	kinds := map[game.EventKind]front.EventKind{
		game.EvAttack: front.EvAttack, game.EvHit: front.EvHit, game.EvMiss: front.EvMiss, game.EvDamage: front.EvDamage,
		game.EvDown: front.EvDied, game.EvSpell: front.EvSpell, game.EvMove: front.EvMoved, game.EvMissile: front.EvMissile,
		game.EvMessage: front.EvMessage, game.EvSound: front.EvSound,
	}
	for _, e := range es {
		if k, ok := kinds[e.Kind]; ok {
			out = append(out, front.CombatEvent{Kind: k, Unit: e.Unit, Target: e.Target, Amount: e.N, From: e.At, To: e.To, Text: e.Text})
		}
	}
	return out
}

func (g *Game) Open(string)       {}
func (g *Game) Resets() int       { return g.resets }
func (g *Game) Refuse(s string)   { g.note = front.Note{Text: s, At: time.Now()} }
func (g *Game) Animate()          { g.step() }
func (g *Game) CombatStep() error { return g.step() }
func (g *Game) RestStep() error   { return g.step() }
func (g *Game) StoryStep() error  { return g.step() }
func (g *Game) WalkStep() error   { return g.step() }

// Party menu.
func (g *Game) Begin() error             { return g.do(game.Begin{}) }
func (g *Game) Lead(n int) error         { return g.do(game.Select{N: n}) }
func (g *Game) RemoveMember(n int) error { return g.do(game.RemoveCharacter{Roster: n}) }
func (g *Game) DropMember(n int) error   { return g.do(game.DropChar{Member: n}) }
func (g *Game) NewClass(n, class int) error {
	return g.do(game.ChangeClass{Member: n, Class: class})
}
func (g *Game) QuickStart()     { g.do(game.QuickStart{}) }
func (g *Game) QuitGame() error { return g.do(game.Quit{}) }
func (g *Game) EndGame() error  { return g.do(game.EndGame{}) }

// Creation sheet.
func (g *Game) CreateCharacter() error        { return g.do(game.NewCharacter{}) }
func (g *Game) ModifyCharacter(n int) error   { return g.do(game.Modify{Member: n}) }
func (g *Game) Rename(name string) error      { return g.do(game.Rename{Name: name}) }
func (g *Game) SetField(f, o int) error       { return g.do(game.SetField{Field: f, Option: o}) }
func (g *Game) EndCreation(keep bool) error   { return g.do(game.EndCreation{Keep: keep}) }
func (g *Game) CreationView() *front.Creation { return Creation(g.R.State(), g.grid.X, g.grid.Y) }
func (g *Game) Roll(minimal bool, minimums string) error {
	return g.do(game.Roll{Minimal: minimal, Minimums: minimums})
}

// SetStat types ability i: "18/50" is strength 18 and exceptional 50.
func (g *Game) SetStat(i int, typed string) error {
	c := g.R.State().Creation
	if c == nil || i < 0 || i >= 6 {
		return g.do(game.Adjust{Stat: i})
	}
	base, exc, _ := strings.Cut(typed, "/")
	n, err := strconv.Atoi(base)
	if err != nil {
		return &game.Refusal{Text: "not a number"}
	}
	if err := g.do(game.Adjust{Stat: i, By: n - c.Draft.Stats[game.StatKeys[i]]}); err != nil || exc == "" {
		return err
	}
	x, _ := strconv.Atoi(exc)
	if x == 0 {
		x = 100
	}
	return g.do(game.Adjust{Stat: 7, By: x - c.Draft.Str00})
}

// Title: the presenter plays it from State.Intro, the game draws it.
func (g *Game) TitlePicture(n int) { g.do(game.TitlePicture{N: n}) }
func (g *Game) TitleCredits()      { g.do(game.TitleCredits{}) }
func (g *Game) TitleFlash(c int)   { g.do(game.TitleFlash{C: c}) }
func (g *Game) TitleDone()         { g.do(game.TitleDone{}) }

// Camp.
func (g *Game) LeaveCamp() error { return g.do(game.LeaveCamp{}) }
func (g *Game) PlanSpell(n, spell, count int) error {
	return g.do(game.Plan{Member: n, Spell: spell, Count: count})
}
func (g *Game) Scribe(n, spell int) error { return g.do(game.Scribe{Member: n, Spell: spell}) }
func (g *Game) CampRest(kind int) error   { return g.do(game.CampRest{Kind: kind}) }
func (g *Game) StopRest() error           { return g.do(game.StopRest{}) }

// Powers.
func (g *Game) UsePower(n int, k front.ActKind, spell, target int, sure bool) error {
	return g.do(game.UsePower{Member: n, Kind: int(k), Spell: spell, Target: target, Sure: sure})
}
func (g *Game) SelectedMember() int { return g.R.State().Selected }

// EffectsList is member n's Effects lines from State.
func (g *Game) EffectsList(n int) (out []front.ListLine) {
	s := g.R.State()
	if n = g.sel(n); n < 0 || n >= len(s.Party) {
		return nil
	}
	for _, l := range s.Party[n].Effects {
		out = append(out, front.ListLine{Text: l.Text, Header: l.Header, Dim: l.Dim, Off: l.Off, Spell: l.Spell})
	}
	return out
}

// Waits.
func (g *Game) LootView(n int) error      { return g.do(game.LootLook{Member: n}) }
func (g *Game) LootTake(n int) error      { return g.do(game.LootTake{Member: n}) }
func (g *Game) LootDetect() error         { return g.do(game.LootDetect{}) }
func (g *Game) LootLeave(sure bool) error { return g.do(game.LootLeave{Sure: sure}) }
func (g *Game) TryDoor(k front.DoorWay) error {
	return g.do(game.TryDoor{Way: int(k)})
}

// Pickers and the icon designer.
func (g *Game) OpenPicker(kind int) bool {
	ok := g.do(game.OpenPicker{Kind: kind}) == nil
	if s := g.R.State(); kind == front.PickWeapon && s.Picker == nil {
		if n, _ := BodyTiles(s); len(n) > 0 { // no pictures from the game: the pool's, by original ref
			g.tiles = true
			return true
		}
	}
	return ok
}
func (g *Game) PickerView(page int) *front.Picker {
	if g.tiles {
		return TilePicker(g.R.State(), page, g.grid.X, g.grid.Y)
	}
	return Picker(g.R.State(), page, g.grid.X, g.grid.Y)
}
func (g *Game) Pick(i int) bool {
	if g.tiles {
		s := g.R.State()
		var d game.Decision = game.SetIcon{Member: s.Selected, Field: "iconbody", ID: PickTile(s, i)}
		if s.Creation != nil {
			d = game.Create{Field: "iconbody", Text: PickTile(s, i)}
		}
		if PickTile(s, i) == "" || g.R.Apply(d).Err != nil {
			return false
		}
		g.tiles = false
		return true
	}
	g.do(game.Pick{N: i})
	return g.R.State().Picker == nil
}
func (g *Game) ClosePicker()   { g.tiles = false; g.do(game.ClosePicker{}) }
func (g *Game) EditIcon(n int) { g.do(game.EditIcon{Member: n}) }
func (g *Game) IconEditorView() *front.IconEditor {
	ink := g.View().Ink
	return IconEditor(g.R.State(), &ink, g.grid.X, g.grid.Y)
}
func (g *Game) IconSize() { g.do(game.IconSize{}) }
func (g *Game) IconColour(part, shade, v int) {
	g.do(game.IconColour{Part: part, Shade: shade, Value: v})
}
func (g *Game) IconRemove()         { g.do(game.IconRemove{}) }
func (g *Game) IconPNG(path string) { g.do(game.IconPNG{Path: path}) }
func (g *Game) EndIcon(keep bool)   { g.do(game.EndIcon{Keep: keep}) }

// SlotsView is State.Saves as the slot list (nil on the title).
func (g *Game) SlotsView(loadOnly bool) *front.Slots {
	st := g.R.State()
	if st.Intro != nil {
		return nil
	}
	inGame := len(st.Party) > 0 && st.Mode != game.ModeTitle && st.Mode != game.ModeParty && st.Mode != game.ModeCreate && st.Mode != game.ModeCombat
	v := &front.Slots{CanSave: !loadOnly && inGame, LoadOnly: loadOnly}
	for _, sl := range st.Saves {
		v.Names, v.Times, v.Levels = append(v.Names, sl.Name), append(v.Times, sl.Time), append(v.Levels, sl.Level)
	}
	if len(v.Names) > 0 {
		v.Current = v.Names[0]
	}
	return v
}
func (g *Game) Save(name string) error       { return g.do(game.Save{Name: name}) }
func (g *Game) DeleteSlot(name string) error { return g.do(game.DeleteSlot{Name: name}) }
func (g *Game) Load(name string, loadOnly bool) bool {
	for i, sl := range g.R.State().Saves {
		if sl.Name == name {
			return g.do(game.Load{Slot: i}) == nil
		}
	}
	return false
}

// TurnChoice is the combat Action as its Decision (the unsure ask is an Ask).
func (g *Game) TurnChoice(a front.Action, sure bool) error { return g.Answer(a) }

// Sheet and inventory.
func (g *Game) OpenSheet(n int) error  { return g.do(game.OpenSheet{Member: n}) }
func (g *Game) CloseSheet() error      { return g.do(game.CloseSheet{}) }
func (g *Game) ShowMember(n int) error { return g.do(game.ShowMember{Member: n}) }
func (g *Game) OpenItems(n int, direct bool) error {
	return g.do(game.OpenItems{Member: n, Direct: direct})
}
func (g *Game) CloseItems(all bool) error { return g.do(game.CloseItems{All: all}) }
func (g *Game) Ready(i, x int) error      { return g.do(game.Ready{Member: g.sel(-1), Item: i, Count: x}) }
func (g *Game) UseItem(i, spell, target int, sure bool) error {
	return g.do(game.UseItem{Member: g.sel(-1), Item: i, Spell: spell, Target: target})
}
func (g *Game) Drop(i int) error { return g.do(game.Drop{Member: g.sel(-1), Item: i}) }
func (g *Game) Sell(i int) error { return g.do(game.Sell{Member: g.sel(-1), Item: i}) }
func (g *Game) Buy(w int) error  { return g.do(game.Buy{Member: g.sel(-1), Item: w}) }
func (g *Game) TakeAll() error   { return g.do(game.TakeLoot{Item: -1, All: true}) }
func (g *Game) SetPolicy(id int, p front.Policy) error {
	return g.do(game.SetPolicy{ID: id, Policy: int(p)})
}

// Shops, temples, training.
func (g *Game) Browse(n int) error            { return g.do(game.Browse{Member: n}) }
func (g *Game) Identify(n, x int) error       { return g.do(game.Identify{Member: n, Item: x}) }
func (g *Game) Appraise(jewelry bool) error   { return g.do(game.Value{Jewelry: jewelry}) }
func (g *Game) KeepAppraised(keep bool) error { return g.do(game.KeepAppraised{Keep: keep}) }
func (g *Game) Cure(n, x int) error           { return g.do(game.Cure{Member: n, Service: x}) }
func (g *Game) ViewMember(n int) error        { return g.do(game.ViewMember{Member: n}) }
func (g *Game) Leave() error                  { return g.do(game.Leave{}) }
func (g *Game) OpenTraining(n int) error      { return g.do(game.OpenTraining{Member: n}) }
func (g *Game) CloseTraining() error          { return g.do(game.CloseTraining{}) }
func (g *Game) TrainingView() *front.Training { return Training(g.R.State()) }
func (g *Game) Train() error                  { return g.do(game.Train{Member: g.sel(-1)}) }
func (g *Game) Learn(id int) error            { return g.do(game.Learn{Spell: id}) }
func (g *Game) PickPower(id int) error        { return g.do(game.PickPower{ID: id}) }

// The full map: its cursor and label are the presenter's, classic fills the rest.
func (g *Game) FullMapView(cursor image.Point, label *string) *front.FullMap {
	return FullMap(&front.FullMap{Cursor: cursor}, g.R.State())
}
func (g *Game) Mark(x, y int, text string) { g.do(game.Mark{At: image.Pt(x, y), Text: text}) }
func (g *Game) Unmark(x, y int)            { g.do(game.Unmark{At: image.Pt(x, y)}) }
func (g *Game) WalkTo(x, y int) bool       { return g.do(game.WalkTo{At: image.Pt(x, y)}) == nil }

// JournalPage is the journal text laid out on the page (Data.Scans has its pictures).
func (g *Game) JournalPage(title string, lines []string, p, off int) (*front.Pages, int) {
	return JournalPage(title, lines, p, off, g.grid.X, g.grid.Y, g.R.State().Ink, g.D.Scans)
}

// SpeakPage has the game voice the page's clips.
func (g *Game) SpeakPage(lines []string, p int) {
	for _, k := range SpeakKeys(lines, p, g.grid.X, g.grid.Y) {
		g.do(game.Speak{Key: k})
	}
}

// ClueChapters and ClueRead are Data.Clues.
func (g *Game) ClueChapters() (out []string) {
	for _, c := range g.D.Clues {
		out = append(out, c.Title)
	}
	return out
}
func (g *Game) ClueRead(i int) *front.Read {
	if len(g.D.Clues) == 0 {
		return nil
	}
	r := g.D.Clues[min(max(i, 0), len(g.D.Clues)-1)].Read
	return &front.Read{Title: r.Title, Lines: r.Lines, Page: r.Page}
}

// Answer does the Decision a names; ActResize is the front's own (the page's size).
func (g *Game) Answer(a front.Action) error {
	if a.Kind == front.ActResize {
		g.grid = image.Pt(max(a.X, 40), max(a.Y, 25))
		return g.do(game.Resize{Cols: a.X, Rows: a.Y})
	}
	if q := g.asked; q != nil { // a Refusal's question: an option's Do, or anything else drops it
		g.asked = nil
		if (a.Kind == front.ActChoose || a.Kind == front.ActOK) && a.N >= 0 && a.N < len(q.Options) && q.Options[a.N].Do != nil {
			return g.do(q.Options[a.N].Do)
		}
		if a.Kind == front.ActClose {
			return nil
		}
	}
	d, ok := g.decision(a)
	if !ok {
		front.Unhandled(a)
		return nil
	}
	return g.do(d)
}

// decision is the Decision for a (false: the game takes none).
func (g *Game) decision(a front.Action) (game.Decision, bool) {
	at := image.Pt(a.X, a.Y)
	m := g.sel(a.N)
	switch a.Kind {
	case front.ActClose:
		return game.Cancel{}, true
	case front.ActMainMenu:
		return game.EndGame{}, true
	case front.ActNote:
		return game.AddNote{Text: paneText(g.R.State())}, true
	case front.ActNoteRemove:
		return game.RemoveNote{ID: a.N}, true
	case front.ActOrder:
		return game.Order{From: a.N, To: a.X}, true
	case front.ActOK, front.ActChoose:
		return game.Answer{N: a.N}, true
	case front.ActYes:
		return game.Answer{N: 1}, true
	case front.ActNo:
		return game.Answer{N: 0}, true
	case front.ActText:
		return game.Answer{Text: a.Text}, true
	case front.ActForward:
		return game.Forward{}, true
	case front.ActTurnRight:
		return game.TurnBy{By: 2}, true
	case front.ActTurnLeft:
		return game.TurnBy{By: 6}, true
	case front.ActSearch:
		return game.Search{}, true
	case front.ActCamp:
		return game.Camp{}, true
	case front.ActLead:
		return game.Select{N: a.N}, true
	case front.ActMember:
		return game.ShowMember{Member: a.N}, true
	case front.ActInventory:
		return game.OpenItems{Member: a.N}, true
	case front.ActSheet:
		return game.CloseItems{}, true
	case front.ActReady:
		return game.Ready{Member: g.sel(-1), Item: a.N, Count: a.X}, true
	case front.ActUse:
		return game.UseItem{Member: g.sel(-1), Item: a.N}, true
	case front.ActDrop:
		return game.Drop{Member: g.sel(-1), Item: a.N}, true
	case front.ActSell:
		return game.Sell{Member: g.sel(-1), Item: a.N}, true
	case front.ActBuy:
		return game.Buy{Member: g.sel(-1), Item: a.N}, true
	case front.ActTakeAll:
		return game.TakeLoot{Item: -1, All: true}, true
	case front.ActWalk:
		return game.Move{To: at}, true
	case front.ActAttack:
		return game.Attack{Target: a.N}, true
	case front.ActAim:
		return game.Aim{At: at}, true
	case front.ActGuard:
		return game.Guard{}, true
	case front.ActWait:
		return game.Wait{}, true
	case front.ActEnd:
		return game.EndTurn{}, true
	case front.ActFlee:
		return game.Flee{}, true
	case front.ActQuick:
		return game.Quick{All: a.N == 1}, true
	case front.ActBandage:
		return game.Bandage{Member: m}, true
	case front.ActTurn:
		return game.TurnUndead{}, true
	case front.ActRanged:
		return game.Ranged{}, true
	case front.ActItems:
		return game.Items{}, true
	case front.ActManeuver:
		return game.Maneuver{}, true
	case front.ActRest:
		return game.CampRest{Kind: a.N}, true
	case front.ActPlan:
		return game.Plan{Member: m, Spell: a.X, Count: a.Y}, true
	case front.ActScribe:
		return game.Scribe{Member: m, Spell: a.X}, true
	case front.ActDropMember:
		return game.DropChar{Member: m}, true
	case front.ActIcon:
		return game.IconPNG{Path: a.Text}, true
	case front.ActIconEdit:
		return game.EditIcon{Member: a.N}, true
	case front.ActCreate:
		return game.NewCharacter{}, true
	case front.ActModify:
		return game.Modify{Member: m}, true
	case front.ActRemove:
		return game.RemoveCharacter{Roster: m}, true
	case front.ActTraining:
		return game.OpenTraining{Member: m}, true
	case front.ActNewClass:
		return game.ChangeClass{Member: m, Class: -1}, true // the game asks which (its Refusal's Ask)
	case front.ActQuickStart:
		return game.QuickStart{}, true
	case front.ActBegin:
		return game.Begin{}, true
	case front.ActQuit:
		return game.Quit{}, true
	case front.ActRoll:
		return game.Roll{Minimal: a.N == 1, Minimums: a.Text}, true
	case front.ActName:
		return game.Rename{Name: a.Text}, true
	case front.ActSetField:
		return game.SetField{Field: a.N, Option: a.X}, true
	case front.ActStat:
		return game.SetStat{Stat: a.N, Text: a.Text}, true
	case front.ActKeep:
		if g.R.State().Creation != nil { // the creation sheet's Keep
			return game.KeepDraft{}, true
		}
		return game.KeepAppraised{Keep: a.N == 1}, true
	}
	return nil, false
}

// paneText is the message pane's words on one line: the note ActNote takes.
func paneText(s game.State) string {
	if s.Page == nil {
		return ""
	}
	var w []string
	for _, l := range s.Page.Pane {
		for _, r := range l {
			w = append(w, strings.Fields(r.Text)...)
		}
	}
	return strings.Join(w, " ")
}

// State is the rules' state.
func (g *Game) State() game.State { return g.R.State() }

// Rules is the game's rules (present.WithRules wraps the game with gobox's pool pickers).
func (g *Game) Rules() game.Rules { return g.R }

// state is the rules' state with a Refusal's question pending.
func (g *Game) state() game.State {
	s := g.R.State()
	if g.asked != nil && s.Pending == nil {
		s.Pending = g.asked
	}
	return s
}
