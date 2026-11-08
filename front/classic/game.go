package classic

import (
	"cmp"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// Game is a front.Game built from a game's rules and data alone: View comes
// from the classic builders over State, every rule call and Action becomes
// the Decision the game takes. It holds no game code.
type Game struct {
	G play.Game
	D game.Data
	r *rules // what gobox holds over G

	grid       image.Point   // the page's size in cells (ActResize); 40x25 at least
	events     []game.Event  // the fight's events since the last Settle
	lastCombat *game.Combat  // the fight before the step: its events play on it (Settle)
	ended      *game.Combat  // that fight as the ending events so far left it
	lastParty  []game.Member // the party before the step: the events' figures are theirs (a fight the party lost leaves none)
	note       front.Note    // the front's own refusal toast
	lore       int           // journal notes seen (State.Lore), +1; 0: none counted yet
	resets     int
	traceMode  game.Mode // the mode last traced
	tiles      bool      // the weapon picker up is the pool bodies by original ref (TilePicker)
	asked      *game.Ask // a Refusal's question put instead (answered by its options' Do)
	aim        *aiming   // a combat spell's squares being picked (Combat.Aims)
	sheet      int       // the look-only sheet up (a fight's View): member+1, 0 none
	found      bool      // the treasure's "You found some loot!" was acknowledged (cleared when it goes)
}

// aiming is the squares picked so far for spell, up to n.
type aiming struct {
	spell, n int
	ats      []image.Point
}

var _ front.Game = (*Game)(nil)

// NewGame is the front.Game over p and d.
func NewGame(p play.Game, d game.Data) *Game {
	GameID = d.ID
	return &Game{G: p, r: newRules(p, d), D: d, grid: image.Pt(40, 25)}
}

// do hands d to the rules, keeping its events; the refusal is the error.
func (g *Game) do(d game.Decision) error {
	if c := g.combat(); c != nil { // the fight as the decision found it: a Result that ends it plays on it (Settle)
		g.lastCombat = c
	}
	g.lastParty = g.party()
	was := g.pane()
	paneWas, linesWere := paneTail(was), paneCount(was)
	res := g.r.Apply(d)
	g.trace(fmt.Sprintf("%T%+v", d, d), res)
	if g.combat() == nil && g.itemsUp() { // over the items the text pane is hidden: its new line (BR2's "Doc buys a ...", "not enough credits") is the toast
		if t := paneTail(g.pane()); t != "" && (t != paneWas || paneCount(g.pane()) != linesWere) { // a new line, or the same one said again
			g.note = front.Note{Text: t, At: time.Now()}
		}
	}
	g.events = append(g.events, res.Events...)
	for _, e := range res.Events {
		if e.Kind == game.EvReset {
			g.resets++
		}
		if e.Kind == game.EvMessage && strings.TrimSpace(e.Text) != "" && g.combat() == nil && g.itemsUp() { // over the items (the shop's "Brann buys a Broad Sword"): the toast on its bottom row, as a refusal's; the inventory plays no message
			g.note = front.Note{Text: strings.TrimSpace(e.Text), At: time.Now()}
		}
	}
	if r, ok := res.Err.(*game.Refusal); ok && r.Ask != nil { // the question to put instead
		g.asked = r.Ask
		return nil
	}
	if res.Err != nil && res.Err.Error() != "" { // a refusal always says why, whichever screen asked
		g.Refuse(res.Err.Error())
		if _, ok := res.Err.(*game.Refusal); ok && g.combat() != nil { // in a fight also the message row's, held as a message (POR's "Blocked")
			g.events = append(g.events, game.Event{Kind: game.EvMessage, Text: res.Err.Error()})
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
func (g *Game) Send(d game.Decision) game.Result { return g.r.Apply(d) }

// Step goes on with what gobox paces, for scripted drives.
func (g *Game) Step() game.Result { return g.r.Step() }

func (g *Game) step() error {
	res := g.r.Step()
	if len(res.Events) > 0 || res.Err != nil {
		g.trace("step", res)
	}
	g.events = append(g.events, res.Events...)
	return res.Err
}

// View is the page classic builds from State.
// art is the rules' art as it is now, else the Data's.
// Subject is the member whose portrait and icon are on show, as the rules
// state it: the draft, else the selected one.
func (g *Game) Subject() game.Member {
	ps := g.r.g.State()
	return OwnSubject(g.r.modeOf(ps), creationOf(ps), partyOf(ps), g.r.selectedOf(ps))
}

// SheetIcon is v with the draft's combat icon in the creation sheet's box.
func (g *Game) SheetIcon(v front.View) front.View {
	return PoolIcon(v, g.Subject(), g.r.sheetPageOf(g.r.g.State(), g.party()))
}

// PickerOf is the game's grid picker, page page on cols by rows cells.
func (g *Game) PickerOf(k *game.Picker, page, cols, rows int) *front.Picker {
	return Picker(k, image.Point{}, page, cols, rows)
}

// TilesOf is the weapon's pool bodies' picker.
func (g *Game) TilesOf(page, cols, rows int) *front.Picker {
	return TilePicker(g.Subject(), image.Point{}, page, cols, rows)
}

// Art is the game's art as it is now.
func (g *Game) Art() map[string]game.Picture { return g.art() }

func (g *Game) art() map[string]game.Picture {
	return g.r.Art()
}

func (g *Game) View() front.View {
	r, ps := g.r, g.r.g.State()
	c := commonOf(ps)
	art := g.art()
	mode, pending := r.modeOf(ps), r.pending(ps)
	if g.asked != nil {
		mode, pending = game.ModeAsk, g.asked
	}
	party, sel := partyOf(ps), r.selectedOf(ps)
	sheet, page := r.sheetOf(c), r.sheetPageOf(ps, party)
	combat := r.combatOf(ps)
	if n := g.sheet - 1; n >= 0 && n < len(party) && combat != nil { // the look-only sheet: the member's page, no powers
		sel, sheet, page = n, &game.Sheet{}, cmp.Or(party[n].Page, game.MemberPage(party[n]))
	} else {
		g.sheet = 0
	}
	amap, allowed, loot, shop := r.automapOf(ps), allowedOf(ps), lootOf(ps), shopOf(ps)
	services, offers := templeOf(ps)
	training, draft := trainingOf(ps), creationOf(ps)
	own := OwnSubject(mode, draft, party, sel)
	v := front.View{Exploring: mode == game.ModeExplore, NoMap: amap.NoMap}
	if amap.Size == image.Pt(16, 16) {
		v.Map = &front.MapSquares{}
	}
	switch r.itemsUp(ps) { // the builders refine a screen the game has up
	case game.ItemsOwn:
		v.Inventory = &front.Inventory{}
	case game.ItemsShop, game.ItemsLoot:
		v.Inventory = &front.Inventory{Wares: []front.InvItem{}}
	case game.ItemsNone: // the treasure is up: its screen is the loot inventory
		if mode == game.ModeLoot && loot != nil && sheet == nil && g.found {
			v.Inventory = &front.Inventory{Wares: []front.InvItem{}}
		}
	}
	if len(amap.Terrain) > 0 {
		v.AreaMap = &front.AreaMap{}
	}
	if shop != nil {
		v.Shop = &front.Shop{}
	}
	if services != nil {
		v.Temple = &front.Temple{}
	}
	if mode == game.ModeCamp {
		v.Camp = &front.Camp{}
	}
	heals, studies := restOf(ps)
	v.MainMenu, v.Hall = PartyMenu(mode, training, pending, c.Where.Hall)
	v.Map, v.Camp = Map(v.Map, amap, c.Where), Camp(v.Camp, party, sel, heals, studies, allowed["NoRest"], nil)
	v.Temple, v.Shop = Temple(v.Temple, services, offers, c.Purse, allowed, party, sel), Shop(v.Shop, shop, c.Purse, allowed, party)
	v.Inventory = Inventory(v.Inventory, mode, party, sel, c.PartyCap, c.PoolItems, c.Purse, allowed, shop, loot)
	v = askOver(v, mode, pending, loot != nil || training != nil)
	g.found = g.found && loot != nil                     // a new treasure is announced again
	if mode == game.ModeLoot && loot != nil && !g.found { // the treasure's picture under a message; Enter or Space opens it
		v.Ask = &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: "You found some loot!"}}
	}
	if v.AreaMap != nil {
		v.AreaMap = AreaMap(v.AreaMap, amap, c.Where, art)
	}
	if v.MainMenu != nil && !v.Hall && v.Powers == nil {
		v.Logo, v.Backdrop = Title(v.Logo, v.Backdrop, art)
	}
	v.Battle = Battle(BattleOf(combat, party, art), combat, party)
	if m := g.aim; m != nil && v.Battle != nil && combat != nil && len(v.Battle.AreaAt) == 0 { // the aimed spell's area, previewed
		for _, sp := range combat.Aims {
			if sp.ID == m.spell {
				v.Battle.AimArea = sp.Area + 1
			}
		}
	}
	v.Battle = ViewVerb(v.Battle, g.D)
	v = Page(v, r.pageOf(ps), storyOf(c))
	v = Flags(v, r.flagsOf(ps, party), OverMenu(mode, pending))
	if m := g.walkMap(); m != nil {
		v.Open = openSides(m, c.Where)
	}
	v = Place(v, c.Where, windowOf(ps))
	v.Quests, v.Lore = Quests(c.Quests), Lore(c.Lore)
	v = Waits(v, game.Hold(holding(ps)), doorOf(ps), c.PowerPick)
	v.Party, v.Purse = Party(party, sel, art), Purse(c.Purse, c.PoolItems, c.PartyCap)
	v = Sheet(v, sheet, image.Rectangle{}, image.Rectangle{}, page, Drafting(mode, draft), draft, own)
	v = PoolPortrait(v, own)
	v = Border(v, art)
	v = Walls(v, r.wallsOf(ps), art)
	v = Intro(v, introOf(ps))
	v = Sequence(v, sequenceOf(ps), g.D.Sequences, g.D.Files)
	v = Note(v, r.note)
	v = Results(v, resultsOf(ps))
	v = Screen(v, image.Point{}, screenOf(ps), [16]color.RGBA{})
	v.Read = r.readOf(ps)
	v = Pacing(v, r.pacingOf(ps))
	if n := len(c.Lore) + 1; g.lore > 0 && n > g.lore { // the journal gained a note: gobox says so, the game prints nothing
		g.note = front.Note{Text: "Added to journal", At: time.Now()}
		g.lore = n
	} else if g.lore == 0 || n < g.lore { // the first look, a load or a new game: no news
		g.lore = n
	}
	if v.Note.Text == "" {
		v.Note = g.note
	}
	v.Creation = Creation(mode, draft, pending, g.grid.X, g.grid.Y)
	v = PoolIcon(v, own, page)
	v.Training = Training(training, party, sel, c.Purse)
	v.Picker = Picker(c.Picker, image.Point{}, 0, g.grid.X, g.grid.Y)
	v.IconEditor = IconEditor(c.IconEditor, nil, own, &v.Ink, g.grid.X, g.grid.Y)
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
	if g.pace() == game.PaceNone { // a paced step is the front's call (CombatStep ..)
		if c := g.combat(); c != nil && g.lastCombat == nil { // a fight already up before its first step (a dev start): its events play from it too
			g.lastCombat = c
		}
		g.lastParty = g.party()
		g.step()
	}
	v := g.View()
	posed := g.posedBattle
	for _, e := range g.events {
		if e.Kind == game.EvEngaged && e.Combat != nil { // the fight began in the step: its events play from its first board
			g.lastCombat, g.lastParty = e.Combat, e.Party
		}
	}
	if g.lastCombat != nil { // the events play from the board before the step, each on the board the ones before it left (one step is many turns: the dead stand until they fall, a walk starts where it starts); a fight they ended: the page after them
		c := *g.lastCombat
		c.Units, c.Squares = slices.Clone(c.Units), slices.Clone(c.Squares)
		g.ended = &c
		posed = g.stepBattle
	}
	v.Events = combatEvents(g.events, v, posed, g.messageMs())
	g.events, g.lastCombat = nil, g.combat()
	return v
}

// messageMs is how long a fight's message holds: the game speed × 100 ms,
// the rules' speed where they keep one (Speed), else g.speed (SetSpeed).
func (g *Game) messageMs() int {
	return g.r.Speed() * 100
}

// stepBattle is the fight as the step found it (g.ended), for each of its
// events in turn: every one is held on the board the events before it left
// (a walk moved, a fall dropped), with its unit in its pose.
func (g *Game) stepBattle(e game.Event) *front.Battle {
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
	b := BattleOf(&c, g.lastParty, g.art())
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
	return b
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
			sc := Page(v, e.Page, nil)
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
			} else if e.Kind == game.EvDown && hasPic(e.Picture) && posed != nil { // a fall that leaves a body is held with the body down (with no moment, the shot before it kept the figure standing up to the page)
				ce.Ms, ce.Battle = cmp.Or(e.Ms, holdMs), posed(e)
			} else if e.Pose > 0 && e.Ms > 0 && posed != nil { // the moment held with the unit in its pose
				ce.Ms, ce.Battle = e.Ms, posed(e)
			} else if e.Text != "" && e.Ms == 0 && v.Battle != nil && posed != nil { // a fight's message stays up the game-speed delay before what follows, as the originals'
				ce.Ms, ce.Battle = holdMs, posed(e)
			}
			if k == front.EvMissile && len(e.Sprite) > 0 && ce.Battle == nil && v.Battle != nil && posed != nil { // a shot the game gave no moment of its own: it flies on the board as it stands
				ce.Ms, ce.Battle = cmp.Or(e.Ms, flightMs), posed(e)
			}
			if k == front.EvMissile && len(e.Sprite) > 0 && ce.Battle != nil { // the shot flies: its sprite square by square from the shooter, in thirds, over the moment
				out = append(out, missileFlight(ce, e.Sprite[0], e.StepMs)...)
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
func missileFlight(ce front.CombatEvent, art game.Picture, step int) []front.CombatEvent {
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
		if step > 0 { // the game's own pace, whatever the game speed
			f.Ms = step
		}
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
	now := g.combat()
	if now == nil {
		return nil
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
	return BattleOf(&c, g.party(), g.art())
}

// Open starts the rules on save slot slot ("" none) when they take it.
func (g *Game) Open(slot string) {
	if o, ok := g.G.(interface{ Open(string) }); ok {
		o.Open(slot)
	}
}

func (g *Game) Resets() int       { return g.resets }
func (g *Game) Refuse(s string)   { g.note = front.Note{Text: s, At: time.Now()} }
func (g *Game) Animate()          { g.step() }
func (g *Game) CombatStep() error { return g.step() }
func (g *Game) RestStep() error   { return g.step() }
func (g *Game) StoryStep() error  { return g.step() }
func (g *Game) WalkStep() error   { return g.step() }
func (g *Game) StopWalk() error   { return g.do(game.StopWalk{}) }

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
func (g *Game) CreateCharacter() error      { return g.do(game.NewCharacter{}) }
func (g *Game) ModifyCharacter(n int) error { return g.do(game.Modify{Member: n}) }
func (g *Game) Rename(name string) error    { return g.do(game.Rename{Name: name}) }
func (g *Game) SetField(f, o int) error     { return g.do(game.SetField{Field: f, Option: o}) }
func (g *Game) EndCreation(keep bool) error { return g.do(game.EndCreation{Keep: keep}) }
func (g *Game) CreationView() *front.Creation {
	return g.CreationOn(g.grid.X, g.grid.Y)
}
func (g *Game) PointBuy(scores []int) error { return g.do(game.PointBuy{Scores: scores}) }
func (g *Game) Roll(minimal bool, minimums string) error {
	return g.do(game.Roll{Minimal: minimal, Minimums: minimums})
}

// SetStat types ability i: "18/50" is strength 18 and exceptional 50.
func (g *Game) SetStat(i int, typed string) error {
	c := creationOf(g.r.g.State())
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
func (g *Game) SelectedMember() int { return g.selected() }

// EffectsList is member n's Effects lines from State.
func (g *Game) EffectsList(n int) (out []front.Entry) {
	party := g.party()
	if n = g.sel(n); n < 0 || n >= len(party) {
		return nil
	}
	for _, l := range party[n].Effects {
		out = append(out, front.Entry{Text: l.Text, Header: l.Header, Dim: l.Dim, Off: l.Off, Spell: l.Spell})
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
	if a := g.asked; a != nil && slices.ContainsFunc(a.Options, func(o game.Option) bool { return o.Do != nil }) { // the game's own answers (POR: Yes stays, No leaves Sure): its question as it is
		return nil
	}
	if a := g.asked; a != nil && (a.Kind == game.YesNo || a.Kind == game.Choice) { // no answers of its own: the presenter asks it as a Confirm, Yes leaving (Silver Blades and Pools of Darkness state it as a Yes/No Choice)
		g.asked = nil
		return front.Confirm(a.Prompt)
	}
	var r *game.Refusal
	if !sure && errors.As(err, &r) && r.Text != "" { // the game's own question as its text (POR: "...go back and claim your treasure?"): Yes stays, No leaves
		return front.ConfirmNo(r.Text)
	}
	return err
}
func (g *Game) TryDoor(k front.DoorWay) error {
	return g.do(game.TryDoor{Way: int(k)})
}

// Pickers and the icon designer.
func (g *Game) OpenPicker(kind int) bool {
	ok := g.do(game.OpenPicker{Kind: kind}) == nil
	if kind == front.PickWeapon && commonOf(g.r.g.State()).Picker == nil {
		if n, _ := BodyTiles(g.Subject()); len(n) > 0 { // no pictures from the game: the pool's, by original ref
			g.tiles = true
			return true
		}
	}
	return ok
}
func (g *Game) PickerView(page int) *front.Picker {
	if g.tiles {
		return g.TilesOf(page, g.grid.X, g.grid.Y)
	}
	return g.PickerOf(commonOf(g.r.g.State()).Picker, page, g.grid.X, g.grid.Y)
}
func (g *Game) Pick(i int) bool {
	if g.tiles {
		t := PickTile(g.Subject(), i)
		if t == "" || g.SetChoice("iconbody", t) != nil {
			return false
		}
		g.tiles = false
		return true
	}
	g.do(game.Pick{N: i})
	return commonOf(g.r.g.State()).Picker == nil
}

// SetChoice sends gobox's pool pick field (the portrait's or the icon's) to
// the game: it takes it for the selected member or the draft.
func (g *Game) SetChoice(field, id string) error {
	_, err := g.G.Do(play.SetChoice{Field: field, Text: id})
	return err
}
func (g *Game) ClosePicker()   { g.tiles = false; g.do(game.ClosePicker{}) }
func (g *Game) EditIcon(n int) { g.do(game.EditIcon{Member: n}) }
func (g *Game) IconEditorView() *front.IconEditor {
	ink := g.View().Ink
	return g.IconEditorOn(&ink, g.grid.X, g.grid.Y)
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
	ps := g.r.g.State()
	c := commonOf(ps)
	if introOf(ps) != nil {
		return nil
	}
	mode := g.r.modeOf(ps)
	inGame := len(c.Party) > 0 && mode != game.ModeTitle && mode != game.ModeCreate && mode != game.ModeCombat
	canSave := !loadOnly && inGame && !c.NoSave
	if !canSave && c.NoLoad { // neither offered here: no slot screen
		return nil
	}
	v := &front.Slots{CanSave: canSave, LoadOnly: loadOnly, NoLoad: c.NoLoad}
	for _, sl := range g.r.saves() {
		v.Names, v.Times, v.Levels = append(v.Names, sl.Name), append(v.Times, sl.Time), append(v.Levels, sl.Level)
		v.Overwrites = append(v.Overwrites, sl.Overwrite)
	}
	if len(v.Names) > 0 {
		v.Current = v.Names[0]
	}
	return v
}
func (g *Game) Save(name string) error       { return g.do(game.Save{Name: name}) }
func (g *Game) DeleteSlot(name string) error { return g.do(game.DeleteSlot{Name: name}) }
func (g *Game) Load(name string, loadOnly bool) bool {
	for i, sl := range g.r.saves() {
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
	c := g.combat()
	if c == nil {
		return g.do(game.OpenSheet{Member: n})
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
	if g.sheet > 0 && n >= 0 && n < len(g.party()) { // another member's look-only sheet
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
	if g.combat() != nil && !all {
		if err := g.do(game.CloseItems{All: true}); err != nil {
			return err
		}
		return g.OpenSheet(g.selected())
	}
	return g.do(game.CloseItems{All: all})
}
func (g *Game) Ready(i, x int) error { return g.do(game.Ready{Member: g.sel(-1), Item: i, Count: x}) }
func (g *Game) UseItem(i, spell, target int, sure bool) error {
	return g.asking(g.do(game.UseItem{Member: g.sel(-1), Item: i, Spell: spell, Target: target}))
}
func (g *Game) Drop(i int) error { return g.do(game.Drop{Member: g.sel(-1), Item: i}) }
func (g *Game) Sell(i int) error { return g.do(game.Sell{Member: g.sel(-1), Item: i}) }
func (g *Game) Buy(w int) error {
	if lootOf(g.r.g.State()) != nil { // the loot screen's Take, as the act mapping does
		return g.do(game.TakeLoot{Item: w})
	}
	return g.do(game.Buy{Member: g.sel(-1), Item: w})
}
func (g *Game) TakeAll() error { return g.do(game.TakeLoot{Item: -1, All: true}) }
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
func (g *Game) TrainingView() *front.Training {
	ps := g.r.g.State()
	return Training(trainingOf(ps), partyOf(ps), g.r.selectedOf(ps), commonOf(ps).Purse)
}
func (g *Game) Train() error           { return g.do(game.Train{Member: g.sel(-1)}) }
func (g *Game) Learn(id int) error     { return g.do(game.Learn{Spell: id}) }
func (g *Game) PickPower(id int) error { return g.do(game.PickPower{ID: id}) }

// The full map: its cursor and label are the presenter's, classic fills the rest.
func (g *Game) FullMapView(cursor image.Point, label *string) *front.FullMap {
	return g.FullMapOf(&front.FullMap{Cursor: cursor})
}
func (g *Game) Mark(x, y int, text string) { g.do(game.Mark{At: image.Pt(x, y), Text: text}) }
func (g *Game) Unmark(x, y int)            { g.do(game.Unmark{At: image.Pt(x, y)}) }
func (g *Game) WalkTo(x, y int) bool       { return g.do(game.WalkTo{At: image.Pt(x, y)}) == nil }

// JournalPage is the journal text laid out on the page (Data.Scans has its pictures).
func (g *Game) JournalPage(title string, lines []string, p, off int) (*front.Pages, int) {
	return JournalPage(title, lines, p, off, g.grid.X, g.grid.Y, [16]color.RGBA{}, g.D.Scans)
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
		fmt.Fprintf(os.Stderr, "trace %s act %+v\n", traceStamp(), a)
	}
	if a.Kind == front.ActResize {
		err := g.do(game.Resize{Cols: a.X, Rows: a.Y})
		g.grid = image.Pt(40, 25) // the page the game shows: its Cells, else the original's (a game that keeps its page lays nothing out wider)
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
	if ps := g.r.g.State(); g.r.modeOf(ps) == game.ModeLoot && !g.found && g.asked == nil {
		g.found = a.Kind == front.ActOK || a.Kind == front.ActClose
		if a.Kind != front.ActClose { // Esc passes the found page and leaves at once, as Enter then Esc
			return nil
		}
	}
	if a.Kind == front.ActCast {
		if c := g.combat(); c != nil {
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
	if q := g.pending(); q != nil && q.Kind == game.Square && a.Kind == front.ActAim { // the square answers the square question, in any mode (a fight's Target asks in ModeCombat)
		return game.Answer{At: at}, true
	}
	if q := g.pending(); q != nil && g.mode() == game.ModeAsk { // a direction answers the question
		switch {
		case q.View && a.Kind == front.ActView: // the highlighted member (N, or the selected) is viewed; the ask stays
			return game.ViewMember{Member: a.N}, true
		case q.Kind == game.Direction && a.Kind == front.ActDir && a.N >= 1 && a.N <= 9 && a.N != 5:
			return game.Answer{N: [10]int{1: 5, 2: 4, 3: 3, 4: 6, 6: 2, 7: 7, 8: 0, 9: 1}[a.N]}, true // keypad 8 up: facing 0 north, 2 east ...
		}
	}
	switch a.Kind {
	case front.ActOption: // the turn's option N, as the game offers it
		if c := g.combat(); c != nil && c.Turn != nil && a.N >= 0 && a.N < len(c.Turn.Options) && c.Turn.Options[a.N].Do != nil {
			return c.Turn.Options[a.N].Do, true
		}
		return nil, false
	case front.ActClose:
		return game.Cancel{}, true
	case front.ActMainMenu:
		return game.EndGame{}, true
	case front.ActNote:
		return game.AddNote{Text: paneText(g.pane())}, true
	case front.ActNoteRemove:
		return game.RemoveNote{ID: a.N}, true
	case front.ActOrder:
		return game.Order{From: a.N, To: a.X}, true
	case front.ActOK, front.ActChoose:
		return game.Answer{N: a.N}, true
	case front.ActYes, front.ActNo:
		if i := yesNoOption(g.pending(), a.Kind); i >= 0 { // the ask's own Yes and No options: their indexes
			return game.Answer{N: i}, true
		}
		return game.Answer{N: map[front.ActKind]int{front.ActYes: 1}[a.Kind]}, true // no options: Yes 1, No 0
	case front.ActText: // a number typed is N as well (a Number ask's amount), the Text kept
		n, _ := strconv.Atoi(strings.TrimSpace(a.Text))
		return game.Answer{N: n, Text: a.Text}, true
	case front.ActSearch:
		return game.Search{}, true
	case front.ActCamp:
		return game.Camp{}, true
	case front.ActLead:
		return game.Select{N: a.N}, true
	case front.ActMember:
		if up := g.mode(); up == game.ModeTemple || up == game.ModeShop || shopOf(g.r.g.State()) != nil { // the shop's and the temple's View (up whatever the mode)
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
		if lootOf(g.r.g.State()) != nil {
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
		if g.Subject(); Drafting(g.mode(), creationOf(g.r.g.State())) { // the creation sheet's Keep
			return game.KeepDraft{}, true
		}
		return game.KeepAppraised{Keep: a.N == 1}, true
	}
	return nil, false
}

// paneText is the message pane's words on one line: the note ActNote takes.
func paneText(p *game.Page) string {
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
func (g *Game) pane() *game.Page     { return g.r.pageOf(g.r.g.State()) }
func (g *Game) pace() game.Pace      { return g.r.pacingOf(g.r.g.State()).Pace }
func (g *Game) itemsUp() bool        { return g.r.itemsUp(g.r.g.State()) != game.ItemsNone }

// pending is the question up, a Refusal's first.
func (g *Game) pending() *game.Ask {
	if g.asked != nil {
		return g.asked
	}
	return g.r.pending(g.r.g.State())
}

// mode is the screen up, a Refusal's question counted.
func (g *Game) mode() game.Mode {
	if g.asked != nil {
		return game.ModeAsk
	}
	return g.r.modeOf(g.r.g.State())
}

// CreationOn is the creation sheet on a page of cols by rows.
func (g *Game) CreationOn(cols, rows int) *front.Creation {
	ps := g.r.g.State()
	return Creation(g.r.modeOf(ps), creationOf(ps), g.r.pending(ps), cols, rows)
}

// IconEditorOn is the icon designer on a page of cols by rows.
func (g *Game) IconEditorOn(ink *[16]color.RGBA, cols, rows int) *front.IconEditor {
	return IconEditor(commonOf(g.r.g.State()).IconEditor, nil, g.Subject(), ink, cols, rows)
}

// FullMapOf is f with the map of where the party is.
func (g *Game) FullMapOf(f *front.FullMap) *front.FullMap {
	ps := g.r.g.State()
	return FullMap(f, g.r.automapOf(ps), commonOf(ps).Where)
}

// Where is the party's place; Mode the screen up; Dungeon a fight's ground.
func (g *Game) Where() game.Where { return commonOf(g.r.g.State()).Where }
func (g *Game) Mode() game.Mode   { return g.mode() }

// Asked is the question up (nil: none).
func (g *Game) Asked() *game.Ask { return g.pending() }

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

// flightMs is a shot's flight when the game sets no time (Ms, StepMs).
const flightMs = 300

// paneTail is the text pane's last non-empty line (the game's latest message), "" with none.
func paneTail(p *game.Page) string {
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

func paneCount(p *game.Page) int {
	if p == nil {
		return 0
	}
	return len(p.Pane)
}

// PickerOn is page page of the picker up on a page of cols by rows (nil: none).
func (g *Game) PickerOn(page, cols, rows int) *front.Picker {
	if p := g.PickerOf(commonOf(g.r.g.State()).Picker, page, cols, rows); p != nil {
		return p
	}
	return g.PickerView(page)
}

// Draft is the character being made (nil: none).
func (g *Game) Draft() *game.Creation { return creationOf(g.r.g.State()) }

// askOver is v with the question up. A place's mode with no screen of its
// own yet draws its question, never a blank frame.
func askOver(v front.View, mode game.Mode, pending *game.Ask, screen bool) front.View {
	if q, ok := Ask(mode, pending); ok && q != nil {
		v.Ask = q
	}
	if v.MainMenu == nil && v.Ask == nil && v.Camp == nil && v.Temple == nil && v.Shop == nil && v.Inventory == nil && !screen &&
		pending != nil && mode >= game.ModeCamp && mode <= game.ModeLoot {
		if q, ok := Ask(game.ModeAsk, pending); ok {
			v.Ask = q
		}
	}
	return v
}
