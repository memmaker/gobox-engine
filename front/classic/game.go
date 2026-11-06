package classic

import (
	"cmp"
	"fmt"
	"image"
	"os"
	"slices"
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

	grid       image.Point  // the page's size in cells (ActResize); 40x25 at least
	events     []game.Event // the fight's events since the last Settle
	lastCombat *game.Combat // the fight the last Settle left up (a Result that ends it plays its events on it)
	ended      *game.Combat // that fight as the ending events so far left it
	note       front.Note   // the front's own refusal toast
	lore       int          // journal notes seen (State.Lore), +1; 0: none counted yet
	resets     int
	traceMode  game.Mode // the mode last traced
	tiles      bool      // the weapon picker up is the pool bodies by original ref (TilePicker)
	asked      *game.Ask // a Refusal's question put instead (answered by its options' Do)
	aim        *aiming   // a combat spell's squares being picked (Combat.Aims)
	speed      int       // the game speed (SetSpeed), 0 fastest .. 9 slowest: a fight message's hold
	sheet      int       // the look-only sheet up (a fight's View): member+1, 0 none
	found      bool      // the treasure's "You found some loot!" was acknowledged (cleared when it goes)
}

// aiming is the squares picked so far for spell, up to n.
type aiming struct {
	spell, n int
	ats      []image.Point
}

var _ front.Game = (*Game)(nil)

// NewGame is the front.Game over r and d.
func NewGame(r game.Rules, d game.Data) *Game {
	GameID, Flow = d.ID, d.CreationFlow
	return &Game{R: r, D: d, grid: image.Pt(40, 25), speed: 4}
}

// do hands d to the rules, keeping its events; the refusal is the error.
func (g *Game) do(d game.Decision) error {
	if s, ok := d.(game.SetSpeed); ok {
		g.speed = min(max(s.N, 0), 9)
	}
	if c := g.R.State().Combat; c != nil { // the fight as the decision found it: a Result that ends it plays on it (Settle)
		g.lastCombat = c
	}
	res := g.R.Apply(d)
	g.trace(fmt.Sprintf("%T%+v", d, d), res)
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
	if res.Err != nil && res.Err.Error() != "" { // a refusal always says why, whichever screen asked
		g.Refuse(res.Err.Error())
		if _, ok := res.Err.(*game.Refusal); ok && g.R.State().Combat != nil { // in a fight also the message row's, held as a message (POR's "Blocked")
			g.events = append(g.events, game.Event{Kind: game.EvMessage, Text: res.Err.Error()})
		}
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
	if len(res.Events) > 0 || res.Err != nil {
		g.trace("step", res)
	}
	g.events = append(g.events, res.Events...)
	return res.Err
}

// View is the page classic builds from State.
// art is the rules' art as it is now, else the Data's.
func (g *Game) art() map[string]game.Picture {
	if r, ok := g.R.(interface {
		Art() map[string]game.Picture
	}); ok {
		return r.Art()
	}
	return g.D.Art
}

func (g *Game) View() front.View { return g.ViewOf(g.state()) }

// ViewOf is View over s, the rules' State already read (a builder that
// reads it too need not build it twice).
func (g *Game) ViewOf(s game.State) front.View {
	if g.asked != nil {
		s.Pending, s.Mode = g.asked, game.ModeAsk
	}
	art := g.art()
	v := front.View{Exploring: s.Mode == game.ModeExplore, NoMap: s.Map.NoMap}
	if s.Map.Size == image.Pt(16, 16) {
		v.Map = &front.MapSquares{}
	}
	switch s.Items { // the builders refine a screen the game has up: it is up when State says so
	case game.ItemsOwn:
		v.Inventory = &front.Inventory{}
	case game.ItemsShop, game.ItemsLoot:
		v.Inventory = &front.Inventory{Wares: []front.InvItem{}}
	case game.ItemsNone: // the treasure is up: its screen is the loot inventory, whatever the game's own View/Take did
		if s.Mode == game.ModeLoot && s.Loot != nil && s.Sheet == nil && g.found {
			v.Inventory = &front.Inventory{Wares: []front.InvItem{}}
		}
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
	g.found = g.found && s.Loot != nil                        // a new treasure is announced again (a question over this one keeps it)
	if s.Mode == game.ModeLoot && s.Loot != nil && !g.found { // the treasure's picture under a message; Enter or Space opens the loot screen
		v.Ask = &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "You found some loot!"}}
	}
	if v.AreaMap != nil {
		v.AreaMap = AreaMap(v.AreaMap, s, art)
	}
	if v.MainMenu != nil && !v.Hall && v.Powers == nil {
		v.Logo, v.Backdrop = Title(v.Logo, v.Backdrop, art)
	}
	v.Battle = BattleOf(s, art)
	if m := g.aim; m != nil && v.Battle != nil && s.Combat != nil && len(v.Battle.AreaAt) == 0 { // the aimed spell's area, previewed under the pointer
		for _, sp := range s.Combat.Aims {
			if sp.ID == m.spell {
				v.Battle.AimArea = sp.Area + 1
			}
		}
	}
	v.Battle = ViewVerb(v.Battle, g.D)
	v = Page(v, s)
	v = Flags(v, s)
	if m := g.walkMap(); m != nil {
		v.Open = openSides(m, s.Where)
	}
	v = Place(v, s)
	v.Quests, v.Lore = Quests(s), Lore(s)
	v = Waits(v, s)
	v.Party, v.Purse = Party(s, art), Purse(s)
	v = Sheet(v, s)
	v = Border(v, art)
	v = Walls(v, s, art)
	v = Intro(v, s)
	v = Sequence(v, s, g.D.Sequences, g.D.Files)
	v = Note(v, s)
	v = Results(v, s)
	v = Screen(v, s)
	v = Read(v, s)
	v = Pacing(v, s)
	if n := len(s.Lore) + 1; g.lore > 0 && n > g.lore { // the journal gained a note: gobox says so, the game prints nothing
		g.note = front.Note{Text: "Added to journal", At: time.Now()}
		g.lore = n
	} else if g.lore == 0 || n < g.lore { // the first look, a load or a new game: no news
		g.lore = n
	}
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

// ViewVerb is b with the series' View (Data.VerbKeys) on the turn's bar,
// at its VerbOrder place, when the game offers none: the active member's
// sheet (ActView, as a portrait's ActMember in a turn).
func ViewVerb(b *front.Battle, d game.Data) *front.Battle {
	k, ok := d.VerbKeys["View"]
	if b == nil || b.Turn == nil || !ok || len(b.Turn.Options) == 0 || b.Turn.Aim || slices.ContainsFunc(b.Turn.Options, func(o front.TurnOption) bool { return o.Label == "View" || o.Kind == front.ActView }) {
		return b
	}
	c, t := *b, *b.Turn
	at := len(t.Options)
	if vi := slices.Index(d.VerbOrder, "View"); vi >= 0 {
		at = slices.IndexFunc(t.Options, func(o front.TurnOption) bool { i := slices.Index(d.VerbOrder, o.Label); return i > vi })
		if at < 0 {
			at = len(t.Options)
		}
	}
	t.Options = slices.Insert(slices.Clone(t.Options), at, front.TurnOption{Label: cmp.Or(d.Texts["View"], "View"), Key: byte(k), Kind: front.ActView})
	t.Can = slices.Insert(slices.Clone(t.Can), min(at, len(t.Can)), front.ActView)
	c.Turn = &t
	return &c
}

// Settle goes on to the next wait; the View carries the fight's events since.
func (g *Game) Settle() front.View {
	if p := g.R.State().Pacing; p == nil || p.Pace == game.PaceNone { // a paced step is the front's call (CombatStep ..)
		g.step()
	}
	v := g.View()
	posed := g.posedBattle
	if v.Battle == nil && g.lastCombat != nil { // the events ended the fight: they play on it, each on the board the ones before it left, the page after them
		c := *g.lastCombat
		c.Units = slices.Clone(c.Units)
		g.ended = &c
		posed = g.endedBattle
	}
	v.Events = combatEvents(g.events, v, posed, g.messageMs())
	g.events, g.lastCombat = nil, g.R.State().Combat
	return v
}

// messageMs is how long a fight's message holds: the game speed × 100 ms,
// the rules' speed where they keep one (Speed), else g.speed (SetSpeed).
func (g *Game) messageMs() int {
	n := g.speed
	if r, ok := g.R.(interface{ Speed() int }); ok {
		n = r.Speed()
	}
	return n * 100
}

// endedBattle is the fight the Result ended (g.ended), for each of its
// events in turn: every one is held on the board the events before it left
// (a walk moved, a fall dropped), with its unit in its pose.
func (g *Game) endedBattle(e game.Event) *front.Battle {
	s := g.state()
	c := *g.ended
	c.Units = slices.Clone(c.Units)
	for k := range c.Units {
		if c.Units[k].ID == e.Unit && e.Pose > 0 {
			c.Units[k].Pose = e.Pose
		}
	}
	s.Combat = &c
	b := BattleOf(s, g.art())
	for k := range g.ended.Units { // the board after e, for the events that follow
		switch u := &g.ended.Units[k]; {
		case u.ID == e.Unit && e.Kind == game.EvMove:
			u.At = e.To
		case u.ID == e.Unit && e.Kind == game.EvDown:
			u.Dead, u.Gone, u.HP = true, true, 0 // off the map, as EvDied plays it
		}
	}
	return b
}

// combatEvents are the fight's events as the presenter plays them; the
// kinds with no moment to play (pictures) are left out. Outside fights too:
// every step's events come through here (songs, sounds).
func combatEvents(es []game.Event, v front.View, posed func(game.Event) *front.Battle, holdMs int) (out []front.CombatEvent) {
	kinds := map[game.EventKind]front.EventKind{
		game.EvAttack: front.EvAttack, game.EvHit: front.EvHit, game.EvMiss: front.EvMiss, game.EvDamage: front.EvDamage,
		game.EvDown: front.EvDied, game.EvSpell: front.EvSpell, game.EvMove: front.EvMoved, game.EvMissile: front.EvMissile,
		game.EvMessage: front.EvMessage, game.EvSound: front.EvSound, game.EvSong: front.EvSong, game.EvSaves: front.EvSaves,
	}
	for _, e := range es {
		if e.Screen != nil { // a timed typed screen: the scene with the screen as it stood
			sc := v
			sc.Screen = e.Screen
			out = append(out, front.CombatEvent{Kind: front.EvFrame, Ms: e.Ms, Scene: &sc})
			continue
		}
		if e.Page != nil { // a timed page: the scene as the page stood
			sc := Page(v, game.State{Page: e.Page})
			out = append(out, front.CombatEvent{Kind: front.EvFrame, Ms: e.Ms, Scene: &sc})
			continue
		}
		if k, ok := kinds[e.Kind]; ok {
			ce := front.CombatEvent{Kind: k, Unit: e.Unit, Target: e.Target, Amount: e.N, From: e.At, To: e.To, Text: e.Text}
			if n, err := strconv.Atoi(e.Sound); err == nil && ce.Amount == 0 && (k == front.EvSound || k == front.EvSong) { // the sound or song ref as a number
				ce.Amount = n
			}
			if v.Battle == nil && posed != nil { // the fight over: every event on the board the ones before it left
				if b := posed(e); e.Ms > 0 {
					ce.Ms, ce.Battle = e.Ms, b
				}
			} else if e.Pose > 0 && e.Ms > 0 && posed != nil { // the moment held with the unit in its pose
				ce.Ms, ce.Battle = e.Ms, posed(e)
			} else if e.Text != "" && e.Ms == 0 && v.Battle != nil && posed != nil { // a fight's message stays up the game-speed delay before what follows, as the originals'
				ce.Ms, ce.Battle = holdMs, posed(e)
			}
			if k == front.EvMissile && len(e.Sprite) > 0 && ce.Battle != nil { // the shot flies: its sprite square by square from the shooter, in thirds, over the moment
				out = append(out, missileFlight(ce, e.Sprite[0])...)
				continue
			}
			out = append(out, ce)
		}
	}
	return out
}

// missileFlight is the moment ce (a missile held on its Battle) as the
// sprite's flight from ce.Unit's square (else From) to To, a frame a third
// of a square, the moment's time shared among them.
func missileFlight(ce front.CombatEvent, art game.Picture) []front.CombatEvent {
	from := ce.From
	for _, u := range ce.Battle.Units {
		if u.ID == ce.Unit {
			from = image.Pt(u.X, u.Y)
		}
	}
	d := ce.To.Sub(from)
	n := 3 * max(d.X, -d.X, d.Y, -d.Y, 1)
	out := make([]front.CombatEvent, 0, n+1)
	for i := 0; i <= n; i++ {
		f := ce
		b := *ce.Battle
		b.Effects = append(slices.Clone(b.Effects), front.Effect{X: float64(from.X) + float64(d.X*i)/float64(n), Y: float64(from.Y) + float64(d.Y*i)/float64(n)})
		b.EffectPix = append(slices.Clone(b.EffectPix), pix(art))
		f.Battle, f.Ms, f.From = &b, max(ce.Ms/(n+1), 1), from
		if i > 0 {
			f.Kind = front.EvFrame
		}
		out = append(out, f)
	}
	return out
}

// posedBattle is the fight as it stands with e's Unit in e's Pose (nil:
// no fight).
func (g *Game) posedBattle(e game.Event) *front.Battle {
	s := g.state()
	if s.Combat == nil {
		return nil
	}
	c := *s.Combat
	c.Units = slices.Clone(c.Units)
	for k := range c.Units {
		if c.Units[k].ID == e.Unit {
			c.Units[k].Pose = e.Pose
		}
	}
	s.Combat = &c
	return BattleOf(s, g.art())
}

// Open starts the rules on save slot slot ("" none) when they take it.
func (g *Game) Open(slot string) {
	if r, ok := g.R.(interface{ Open(string) }); ok {
		r.Open(slot)
	}
}

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
func (g *Game) PointBuy(scores []int) error   { return g.do(game.PointBuy{Scores: scores}) }
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
func (g *Game) TitlePicture(n int)    { g.do(game.TitlePicture{N: n}) }
func (g *Game) TitleCredits(page int) { g.do(game.TitleCredits{Page: page}) }
func (g *Game) TitleFlash(c int)      { g.do(game.TitleFlash{C: c}) }
func (g *Game) TitleDone()            { g.do(game.TitleDone{}) }

// SequenceDone: the presenter played State.Sequence to its end, or it was skipped.
func (g *Game) SequenceDone(name string) { g.do(game.SequenceDone{Name: name}) }

// Camp.
func (g *Game) LeaveCamp() error { return g.do(game.LeaveCamp{}) }
func (g *Game) PlanSpell(n, spell, count int) error {
	return g.do(game.Plan{Member: n, Spell: spell, Count: count})
}
func (g *Game) Scribe(n, spell int) error { return g.do(game.Scribe{Member: n, Spell: spell}) }
func (g *Game) CampRest(kind int) error   { return g.do(game.CampRest{Kind: kind}) }

// RestFor rests minutes long (a RestKind.Duration rest's picked time).
func (g *Game) RestFor(minutes int) error { return g.do(game.Rest{Minutes: minutes}) }
func (g *Game) StopRest() error           { return g.do(game.StopRest{}) }

// Powers.
func (g *Game) UsePower(n int, k front.ActKind, spell, target int, sure bool) error {
	return g.asking(g.do(game.UsePower{Member: n, Kind: int(k), Spell: spell, Target: target, Sure: sure}))
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
func (g *Game) LootView(n int) error { return g.do(game.LootLook{Member: n}) }
func (g *Game) LootTake(n int) error { return g.do(game.LootTake{Member: n}) }
func (g *Game) LootDetect() error    { return g.do(game.LootDetect{}) }
func (g *Game) LootShare() error     { return g.do(game.LootShare{}) }
func (g *Game) LootStatus() error    { return g.do(game.LootStatus{}) }
func (g *Game) LootPool() error      { return g.do(game.PoolCoins{Member: g.sel(-1)}) }
func (g *Game) LootLeave(sure bool) error {
	err := g.do(game.LootLeave{Sure: sure})
	if a := g.asked; a != nil && (a.Kind == game.YesNo || a.Kind == game.Choice) { // the presenter asks it as a Confirm (Silver Blades and Pools of Darkness state it as a Yes/No Choice)
		g.asked = nil
		return front.Confirm(a.Prompt)
	}
	return err
}
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
		if Drafting(s) {
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
	inGame := len(st.Party) > 0 && st.Mode != game.ModeTitle && st.Mode != game.ModeCreate && st.Mode != game.ModeCombat
	canSave := !loadOnly && inGame && (st.Flags == nil || !st.Flags.NoSave)
	if !canSave && st.Flags != nil && st.Flags.NoLoad { // neither offered here: no slot screen
		return nil
	}
	v := &front.Slots{CanSave: canSave, LoadOnly: loadOnly}
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
// OpenSheet: in a fight the sheet is look-only, gobox's from Member.Page
// (n < 0: the turn's member); the game is not called. Else the game's own.
func (g *Game) OpenSheet(n int) error {
	s := g.R.State()
	if s.Combat == nil {
		return g.do(game.OpenSheet{Member: n})
	}
	if n < 0 {
		n = s.Selected
		if t := s.Combat.Turn; t != nil {
			for _, u := range s.Combat.Units {
				if u.ID == t.Unit && u.Member >= 0 {
					n = u.Member
				}
			}
		}
	}
	if n < 0 || n >= len(s.Party) {
		return nil
	}
	g.sheet = n + 1
	return nil
}

func (g *Game) CloseSheet() error {
	if g.sheet > 0 {
		g.sheet = 0
		return nil
	}
	return g.do(game.CloseSheet{})
}
func (g *Game) ShowMember(n int) error {
	if g.sheet > 0 && n >= 0 && n < len(g.R.State().Party) { // another member's look-only sheet
		g.sheet = n + 1
		return nil
	}
	return g.do(game.ShowMember{Member: n})
}
func (g *Game) OpenItems(n int, direct bool) error {
	return g.do(game.OpenItems{Member: n, Direct: direct})
}
// CloseItems: Tab (all false) is the sheet, as out of a fight; in a fight
// that is the look-only sheet of the member whose items were up.
func (g *Game) CloseItems(all bool) error {
	if s := g.R.State(); s.Combat != nil && !all {
		if err := g.do(game.CloseItems{All: true}); err != nil {
			return err
		}
		return g.OpenSheet(s.Selected)
	}
	return g.do(game.CloseItems{All: all})
}
func (g *Game) Ready(i, x int) error      { return g.do(game.Ready{Member: g.sel(-1), Item: i, Count: x}) }
func (g *Game) UseItem(i, spell, target int, sure bool) error {
	return g.asking(g.do(game.UseItem{Member: g.sel(-1), Item: i, Spell: spell, Target: target}))
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
func (g *Game) ViewMember(n int) error        { return g.do(game.ViewMember{Member: g.sel(n)}) } // -1: the selected member
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

// SpeakPage has the game voice the page, in place of the page voiced before.
func (g *Game) SpeakPage(lines []string, p int) {
	g.do(game.Speak{Keys: SpeakKeys(lines, p, g.grid.X, g.grid.Y)})
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
	if traceOn {
		fmt.Fprintf(os.Stderr, "trace act %+v\n", a)
	}
	if a.Kind == front.ActResize {
		err := g.do(game.Resize{Cols: a.X, Rows: a.Y})
		g.grid = image.Pt(40, 25) // the page the game shows: its Cells, else the original's (a game that keeps its page lays nothing out wider)
		if c := g.R.State().Cells; c.X > 0 && c.Y > 0 {
			g.grid = image.Pt(max(c.X, 40), max(c.Y, 25))
		}
		return err
	}
	if m := g.aim; m != nil { // up to n squares, then the cast; Enter casts on those picked, Esc gives up
		switch {
		case a.Kind == front.ActAim:
			if m.ats = append(m.ats, image.Pt(a.X, a.Y)); len(m.ats) < m.n {
				return nil
			}
		case a.Kind == front.ActOK && len(m.ats) > 0:
		case a.Kind == front.ActClose:
			g.aim, g.asked = nil, nil
			return nil
		default:
			return nil
		}
		g.aim, g.asked = nil, nil
		return g.do(game.Cast{Member: -1, Spell: m.spell, At: m.ats[0], Ats: m.ats})
	}
	if s := g.R.State(); s.Mode == game.ModeLoot && s.Loot != nil && !g.found && g.asked == nil {
		g.found = a.Kind == front.ActOK
		return nil
	}
	if a.Kind == front.ActCast {
		if c := g.R.State().Combat; c != nil {
			for _, sp := range c.Aims {
				if sp.ID == a.N {
					g.aim = &aiming{spell: sp.ID, n: max(sp.Targets, 1)}
					g.asked = &game.Ask{Kind: game.Square, Prompt: "Aim " + sp.Name}
					return nil
				}
			}
			return g.do(game.Cast{Member: -1, Spell: a.N})
		}
	}
	if q := g.asked; q != nil { // a Refusal's question: an option's Do, or anything else drops it
		g.asked = nil
		i := yesNoOption(q, a.Kind) // a Yes/No with its own options: the one labelled so
		if (a.Kind == front.ActChoose || a.Kind == front.ActOK) && a.N >= 0 && a.N < len(q.Options) {
			i = a.N
		}
		if i >= 0 { // an option without a Do (No) takes nothing
			if d := q.Options[i].Do; d != nil {
				return g.do(d)
			}
			return nil
		}
		if a.Kind == front.ActClose {
			return nil
		}
	}
	if _, ok := walks[a.Kind]; ok { // gobox moves the party on the game's Map
		return g.walk(a.Kind)
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
	if q := g.state().Pending; q != nil && q.Kind == game.Square && a.Kind == front.ActAim { // the square answers the square question, in any mode (a fight's Target asks in ModeCombat)
		return game.Answer{At: at}, true
	}
	if q := g.state().Pending; q != nil && g.state().Mode == game.ModeAsk { // a direction answers the question
		switch {
		case q.View && a.Kind == front.ActView: // the highlighted member (N, or the selected) is viewed; the ask stays
			return game.ViewMember{Member: m}, true
		case q.Kind == game.Direction && a.Kind == front.ActDir && a.N >= 1 && a.N <= 9 && a.N != 5:
			return game.Answer{N: [10]int{1: 5, 2: 4, 3: 3, 4: 6, 6: 2, 7: 7, 8: 0, 9: 1}[a.N]}, true // keypad 8 up: facing 0 north, 2 east ...
		}
	}
	switch a.Kind {
	case front.ActOption: // the turn's option N, as the game offers it
		if c := g.state().Combat; c != nil && c.Turn != nil && a.N >= 0 && a.N < len(c.Turn.Options) && c.Turn.Options[a.N].Do != nil {
			return c.Turn.Options[a.N].Do, true
		}
		return nil, false
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
	case front.ActYes, front.ActNo:
		if i := yesNoOption(g.state().Pending, a.Kind); i >= 0 { // the ask's own Yes and No options: their indexes
			return game.Answer{N: i}, true
		}
		return game.Answer{N: map[front.ActKind]int{front.ActYes: 1}[a.Kind]}, true // no options: Yes 1, No 0
	case front.ActText:
		return game.Answer{Text: a.Text}, true
	case front.ActSearch:
		return game.Search{}, true
	case front.ActCamp:
		return game.Camp{}, true
	case front.ActLead:
		return game.Select{N: a.N}, true
	case front.ActMember:
		if s := g.state(); s.Mode == game.ModeTemple || s.Mode == game.ModeShop || s.Temple != nil || s.Shop != nil { // the shop's and the temple's View (up whatever the mode)
			return game.ViewMember{Member: m}, true
		}
		return game.ShowMember{Member: a.N}, true
	case front.ActInventory:
		return game.OpenItems{Member: a.N}, true
	case front.ActSheet:
		return game.CloseItems{}, true
	case front.ActTrade:
		return game.Trade{From: g.sel(-1), To: a.X, Item: a.N}, true
	case front.ActHalve:
		return game.Halve{Member: g.sel(-1), Item: a.N}, true
	case front.ActJoin:
		return game.Join{Member: g.sel(-1), Item: a.N}, true
	case front.ActReady:
		return game.Ready{Member: g.sel(-1), Item: a.N, Count: a.X}, true
	case front.ActUse:
		return game.UseItem{Member: g.sel(-1), Item: a.N}, true
	case front.ActDrop:
		return game.Drop{Member: g.sel(-1), Item: a.N}, true
	case front.ActSell:
		return game.Sell{Member: g.sel(-1), Item: a.N}, true
	case front.ActReload:
		return game.Reload{Member: g.sel(-1), Item: a.N}, true
	case front.ActBuy:
		if g.state().Loot != nil {
			return game.TakeLoot{Item: a.N}, true
		}
		return game.Buy{Member: g.sel(-1), Item: a.N}, true
	case front.ActTakeAll:
		return game.TakeLoot{Item: -1, All: true}, true
	case front.ActTakeMoney:
		return game.TakeLoot{Item: -1}, true
	case front.ActLootPool: // a shop's Pool (the treasure's goes through Gate)
		return game.PoolCoins{Member: g.sel(-1)}, true // the selected member's coins
	case front.ActLootShare:
		return game.LootShare{}, true
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
		if a.X > 0 { // a set duration (RestKind.Duration): X minutes
			return game.Rest{Minutes: a.X}, true
		}
		return game.CampRest{Kind: a.N}, true
	case front.ActPlan:
		return game.Plan{Member: m, Spell: a.X, Count: a.Y}, true
	case front.ActScribe:
		return game.Scribe{Member: m, Spell: a.X}, true
	case front.ActDropMember:
		return game.DropChar{Member: m}, true
	case front.ActIcon, front.ActPortrait: // the rules say which takes it: the icon, or the sheet's portrait
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
	case front.ActCommand: // a game's own party command (Data.Commands)
		if a.N >= 0 && a.N < len(g.D.Commands) && g.D.Commands[a.N].Do != nil {
			return g.D.Commands[a.N].Do, true
		}
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
		if Drafting(g.R.State()) { // the creation sheet's Keep
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

// Shown is the state the screens show: the rules' with a Refusal's
// question pending (a View built on the rules' own State loses it).
func (g *Game) Shown() game.State { return g.state() }

// state is the rules' state with a Refusal's question pending.
func (g *Game) state() game.State {
	s := g.R.State()
	if g.asked != nil {
		s.Pending, s.Mode = g.asked, game.ModeAsk
	}
	if n := g.sheet - 1; n >= 0 && n < len(s.Party) && s.Combat != nil { // the look-only sheet: the member's page, no powers (the game's sheet menu is not up)
		s.Selected, s.Sheet, s.SheetPage = n, &game.Sheet{}, cmp.Or(s.Party[n].Page, game.MemberPage(s.Party[n]))
	} else {
		g.sheet = 0
	}
	return s
}

// asking is a Refusal's question of UsePower or UseItem as the front's
// pick or question (PickSpell, PickMember, Confirm), which the presenter
// asks before making the call again with the answer.
func (g *Game) asking(err error) error {
	a := g.asked
	if a == nil {
		return err
	}
	switch a.Kind {
	case game.PickSpell:
		var ps front.PickSpell
		for _, o := range a.Options {
			id := 0
			switch d := o.Do.(type) {
			case game.UsePower:
				id = d.Spell
			case game.UseItem:
				id = d.Spell
			}
			ps = append(ps, front.Spell{ID: id, Name: o.Label})
		}
		g.asked = nil
		return ps
	case game.PickMember:
		g.asked = nil
		return front.PickMember(a.Prompt)
	case game.YesNo:
		g.asked = nil
		return front.Confirm(a.Prompt)
	}
	return err
}

// yesNoOption is the index of q's option labelled Yes (for ActYes) or No
// (ActNo); -1 when q has none so labelled.
func yesNoOption(q *game.Ask, k front.ActKind) int {
	if q == nil || k != front.ActYes && k != front.ActNo {
		return -1
	}
	want := map[front.ActKind]string{front.ActYes: "yes", front.ActNo: "no"}[k]
	return slices.IndexFunc(q.Options, func(o game.Option) bool { return strings.EqualFold(strings.TrimSpace(o.Label), want) })
}
