package engine

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"gobox/rules"
	"image"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"gobox/game"
)

// The engine is gobox's typed reactive contract (gobox/game): the
// core parks wherever the original waits for the player, State reads where
// as one of play's states, and Do turns one action into the adapter's
// decisions and runs it to the next such point (in combat, the monsters'
// turns with it).
var _ game.Game = (*Engine)(nil)

// Art is the wall art loaded so far (playrules forwards classic's live reads here).
func (e *Engine) Art() map[string]game.Picture { return e.overlandArt() }

var errNotAllowed = errors.New("not allowed now")

// State is where the core is parked, as one of play's states: the one
// built after the last change (rebuild), shared, read-only.
func (e *Engine) State() game.State {
	if e.ps == nil {
		e.rebuild()
	}
	return e.ps
}

// Version is the build number of the State held: it changes whenever State
// is built anew (rebuild), and only then.
func (e *Engine) Version() uint64 {
	if e.ps == nil {
		e.rebuild()
	}
	return uint64(e.StateBuilds)
}

// rebuild builds State anew: every call that changes the engine (Do, Load,
// LoadSlot, Start, Step) ends with it.
func (e *Engine) rebuild() {
	e.ps = e.build()
	if _, ok := e.ps.(game.Looting); !ok { // a new treasure is announced again
		e.lootSeen = false
	}
	e.StateBuilds++
}

// build is where the core is parked now, as one of play's states.
func (e *Engine) build() game.State {
	s := e.known()
	cm := e.common(s)
	switch {
	case e.ext.ending && e.title.end != nil: // the end's pictures, as the title's steps
		return e.endScene(cm)
	case s.Intro != nil:
		sc := e.title.short
		if e.FullIntro.Load() {
			sc = e.title.full
		}
		sc.Common = cm
		return sc
	case s.Sequence != "":
		return game.Scene{Common: cm, Kind: game.SceneSequence, Sequence: s.Sequence}
	case s.Results != nil:
		return game.Scene{Common: cm, Kind: game.SceneResults, Results: s.Results}
	}
	if s.Read != nil { // a handout over any state: its Scene, Leave puts it away
		cm.Ask = nil
		return game.Scene{Common: cm, Kind: game.SceneHandout, Handout: s.Read.Title, Texts: s.Read.Texts}
	}
	mode := s.Mode
	if mode == game.ModeAsk && s.Combat != nil { // a question in a fight is the Fighting's Ask over the field (classic draws it over the map)
		mode = game.ModeCombat
	}
	switch mode {
	case game.ModeNone, game.ModeParty, game.ModeCreate:
		a := game.Assembling{Common: cm}
		for _, c := range e.gameData.Commands { // the game's own party commands: adding from another game
			if c.From != "" {
				a.Commands = append(a.Commands, game.Option{Label: c.Label, Key: c.Key, Do: game.AddCharacter{Roster: -1, From: c.From}})
			}
		}
		a.Roster = e.BuildCharList()
		if cr := s.Creation; cr != nil && s.Mode == game.ModeCreate {
			a.Draft = &game.Draft{Member: cr.Draft, Fields: cr.Fields, MinHP: cr.MinHP, MaxHP: cr.MaxHP, Minimums: cr.Minimums,
				Step: cr.Step}
			if len(cr.Choices) > 0 {
				a.Draft.Ask = &game.Question{Replies: cr.Choices}
			}
			if c := e.creation; c != nil && c.naming { // Keep's name question (BUG-CLASSES 59): a Reply with the name keeps it
				a.Draft.Ask = &game.Question{Text: "Character name:", Typed: true, Max: 0xF}
			}
		}
		return a
	case game.ModeExplore:
		x := game.Exploring{Common: cm, Map: game.ExploreMap{Automap: s.Map, Walk: e.walkMap(), Overland: e.viewMode != 1}, Door: s.Door,
			Day: s.Flags != nil && s.Flags.Day, Interact: s.Flags != nil && s.Flags.Interact} // the sky drawView paints, not the clock
		if w := walls(s); w != nil {
			x.Map.Walls = *w
		}
		return x
	case game.ModeCombat:
		f := game.Fighting{Common: cm}
		if b := s.Combat; b != nil {
			f.Field = game.Battlefield{Size: b.Size, Squares: b.Squares, Tiles: b.Tiles, FloorTile: b.FloorTile, Floor: b.Floor,
				Clouds: b.Effects, Round: b.Round, Dungeon: b.Dungeon}
			if t := b.Turn; t != nil {
				f.Field.Turn, f.Field.Moves, f.Field.Casting, f.Field.Weapon, f.Field.Missile, f.Field.Maneuvers = t.Unit, t.Moves, t.Casting, t.Weapon, t.Missile, t.Maneuvers
			}
			f.Field.Units = append(f.Field.Units, b.Units...)
		}
		return f
	case game.ModeShop, game.ModeTemple, game.ModeTraining, game.ModeVault:
		t := game.Trading{Common: cm, Kind: tradeKinds[s.Mode], Name: s.TempleName, Services: s.Temple, Training: s.Training, Offers: s.TempleOffers, Vault: s.Vault}
		if sh := s.Shop; sh != nil {
			t.Name, t.Goods, t.Buys, t.Offers, t.Pooled = sh.Name, sh.Goods, sh.Buys, sh.Offers, sh.Pooled
			t.Identify, t.Unknown, t.Known = sh.IdentifyPrice, sh.Unknown, sh.Identified
		}
		return t
	case game.ModeLoot:
		l := game.Looting{Common: cm}
		if s.Loot != nil {
			l.Items, l.Money, l.Status = s.Loot.Items, s.Loot.Money, s.Loot.Status
			l.Announce = !e.lootSeen
		}
		return l
	case game.ModeCamp:
		return game.Camping{Common: cm}
	}
	cm.Ask = nil // a Scene's own question (below) carries it
	// the pane, the picture and its layers are Common's (Text, Picture, Layers): the Scene adds only its own
	sc := game.Scene{Common: cm, Kind: game.SceneText, Full: s.Flags != nil && s.Flags.Full}
	if pg := s.Page; pg != nil {
		for row, l := range pg.View {
			for _, t := range l {
				sc.Placed = append(sc.Placed, game.Placed{At: image.Pt(t.X, row+1), Text: t.Text})
			}
		}
	}
	// a Scene's picture and layers are its own (gobox's window is the Scene's), once: a layer already shows the picture
	sc.Layers, sc.Common.Layers, sc.Common.Picture = cm.Layers, nil, nil
	if c := cm.Picture; c != nil && !slices.ContainsFunc(sc.Layers, func(l game.Layer) bool { return l.Picture.File == c.File && l.Picture.Block == c.Block }) {
		sc.Pictures = append(sc.Pictures, *c)
	}
	sc.Pictures = append(sc.Pictures, s.Story...)
	sc.Walls = walls(s) // the 3D view under the page (the opening's come before any exploring)
	if s.Read != nil {
		sc.Kind, sc.Handout, sc.Texts = game.SceneHandout, s.Read.Title, s.Read.Texts
	} else if a := s.Pending; a != nil {
		if a.Text != "" {
			sc.Text = append(sc.Text, a.Text)
		}
		q := &game.Question{Text: sc.Status, List: sc.Status != "" || a.List, NoPick: a.NoPick, NoEsc: a.NoEsc, Esc: a.Esc, Keys: a.Keys} // bug class 49 on the Scene path
		switch {
		case a.Continue && a.Ms > 0: // a timed pause: the Scene carries it, so the presenter paces it
			sc.Ask, q.Continue, q.Ms = q, true, a.Ms
		case a.Continue:
		case a.Typed:
			sc.Kind, sc.Ask, q.Typed, q.Max = game.SceneTalk, q, true, 40
		case a.Number:
			sc.Kind, sc.Ask, q.Number, q.Max = game.SceneTalk, q, true, a.Max
		default: // a yes/no (No, Yes), a menu, a list, a member pick
			sc.Kind, sc.Ask = game.SceneTalk, q
			for _, r := range a.Replies {
				q.Replies = append(q.Replies, strings.TrimSpace(r))
			}
		}
	}
	return sc
}

// ask is the question over a state (a Scene clears it: its Replies);
// the door, the party menu and creation's steps are play's Door,
// Assembling and Draft.
func ask(s facts) *game.Question {
	a := s.Pending
	if s.ShopMenu { // the shop's own Buy, Leave: Trading's Buy and Leave
		return nil
	}
	if a == nil || s.Door != nil || a.Continue && a.Ms == 0 || s.Mode == game.ModeNone || s.Mode == game.ModeParty || s.Mode == game.ModeCreate {
		return nil
	}
	q := &game.Question{Text: a.Text, Default: a.Default, Max: a.Max, NoEsc: a.NoEsc, NoPick: a.NoPick, YesNo: a.YesNo,
		Typed: a.Typed, Number: a.Number, Keys: a.Keys, Esc: a.Esc, Continue: a.Continue, Ms: a.Ms}
	if !a.Typed && !a.Number {
		for _, r := range a.Replies {
			q.Replies = append(q.Replies, strings.TrimSpace(r))
		}
	}
	return q
}

// sceneless: the mode is a state of its own, not a Scene.
func sceneless(m game.Mode) bool {
	switch m {
	case game.ModeExplore, game.ModeCombat, game.ModeShop, game.ModeTemple, game.ModeTraining, game.ModeVault, game.ModeLoot, game.ModeCamp:
		return true
	}
	return false
}

// runText is a row's runs as plain text.
func runText(rs []game.TextRun) (t string) {
	for _, r := range rs {
		t += r.Text
	}
	return t
}

// paneText is the page's message lines as plain strings.
func paneText(pg *game.Page) (out []string) {
	if pg == nil {
		return nil
	}
	for _, l := range pg.Pane {
		var b strings.Builder
		for _, r := range l {
			b.WriteString(r.Text)
		}
		out = append(out, b.String())
	}
	return out
}

// tradeKinds are the trading places by mode.
var tradeKinds = map[game.Mode]game.TradeKind{game.ModeShop: game.TradeShop, game.ModeTemple: game.TradeTemple,
	game.ModeTraining: game.TradeTraining, game.ModeVault: game.TradeVault}

// common is what every state shows, with the actions allowed now.
func (e *Engine) common(s facts) game.Common {
	cm := game.Common{Selected: s.Selected, Purse: s.Purse, PartyCap: s.PartyCap, PoolItems: s.PoolItems, Where: s.Where, Area: s.Map.Area, Lore: s.Lore, Quests: s.Quests,
		PowerPick: s.PowerPick, Picker: s.Picker, IconEditor: s.IconEditor, Automap: s.Map, Text: paneText(s.Page)}
	if len(cm.Text) == 0 && s.LastText != "" {
		cm.Text = []string{s.LastText}
	}
	if pg := s.Page; pg != nil {
		cm.Status = runText(pg.Prompt)
	}
	if s.Picture.Name != "" || s.Picture.File != "" {
		cm.Picture = &s.Picture
	}
	cm.Ask = ask(s)
	if pq := e.ext.pq; pq != nil {
		q := pq.q
		q.Replies = slices.Clone(q.Replies)
		cm.Ask = &q
	}
	cm.Day = s.Flags != nil && s.Flags.Day
	cm.Adventure = s.Flags != nil && s.Flags.Adventure                                        // the old rule (gamepage.go): explore, an ask, camp
	cm.NoSave, cm.NoLoad = s.Flags != nil && s.Flags.NoSave, s.Flags != nil && s.Flags.NoLoad // gobox drops Save or Load (and their hint keys) where the rules refuse them
	if pg := s.Page; pg != nil {
		cm.Layers = pg.Layers
	}
	for i, m := range s.Party {
		pm := m
		pm.Page = m.Page // the game's own sheet of each member: what View shows (BUG-CLASSES 36)
		if i == s.Selected && s.SheetPage != nil {
			if sh := s.Sheet; sh != nil {
				pm.SheetPowers, pm.Reactions = sh.Powers, sh.Reactions
			}
			pm.Page = s.SheetPage
		} else if i == s.Selected && s.Sheet != nil {
			pm.SheetPowers, pm.Reactions = s.Sheet.Powers, s.Sheet.Reactions
		}
		cm.Party = append(cm.Party, pm)
	}
	cm.Allowed = e.allowed(s)
	return cm
}

// readyAct is member n's item it (from State's Items, its card already
// read) as Ready, or Unready when it is readied (a fight's item menu
// toggles, so Ready only).
func (e *Engine) readyAct(n int, it game.Item) game.Action {
	if it.Readied && !e.battleUp() {
		return game.Unready{Member: n, Item: it.Handle}
	}
	return game.Ready{Member: n, Item: it.Handle}
}

// allowed is what the state lets the player do now.
func (e *Engine) allowed(s facts) game.Allowed {
	var a game.Allowed
	add := func(x ...game.Action) { a.Plain = append(a.Plain, x...) }
	if pq := e.ext.pq; pq != nil { // a play action's question: its replies, Leave backs out
		for i := range pq.calls {
			add(game.Reply{N: i})
		}
		add(game.Leave{})
		return a
	}
	items := func() {
		for n, m := range s.Party {
			for _, it := range m.Items {
				add(e.readyAct(n, it), game.Drop{Member: n, Item: it.Handle})
				if len(e.ruleset().Allowed(rules.UsingItem{Who: e.partyIDs()[n], Item: it.Handle}).Offers) > 0 { // an item with no use is not offered Use (UseItem refused it "not now")
					add(game.UseItem{Member: n, Item: it.Handle})
				}
			}
			add(game.SelectMember{Member: n})
		}
	}
	if s.Read != nil {
		add(game.Leave{})
		return a
	}
	switch {
	case s.Intro != nil, s.Sequence != "", e.ext.ending, s.Results != nil && s.Pending == nil: // results with a question answer it, as the Scene's Replies say
		add(game.ContinueAction{})
		return a
	case s.Door != nil: // the Locked. menu: its ways, by the old front's numbering (doorWays)
		for w := range game.DoorWay(4) { // exit, bash, pick, knock: TryDoor checks which are left
			add(game.TryDoor{Way: w})
		}
		return a
	}
	mode := s.Mode
	if mode == game.ModeAsk && s.Combat != nil { // a question in a fight: its replies over the field (as the view)
		mode = game.ModeCombat
	}
	if mode == game.ModeTraining && s.Training == nil { // the hall's party menu before OpenTraining (gobox's PartyMenu reads it so)
		mode = game.ModeParty
	}
	if s.Training != nil && s.Mode == game.ModeTraining {
		add(game.CloseTraining{}) // the training box's Exit (and Esc), back to the hall's menu, over any question the box asks
	}
	if q := ask(s); q != nil && sceneless(mode) { // a question over the state: its replies and only them (bug class 18)
		if q.Typed || q.Number {
			add(game.Reply{})
		}
		for i := range q.Replies {
			add(game.Reply{N: i})
		}
		return a
	}
	if s.PowerPick != nil {
		add(game.PickPower{})
	}
	if s.Picker != nil {
		for n := range max(len(s.Picker.Pics), len(s.Picker.Labels)) { // every row on every page
			add(game.Pick{N: n})
		}
		add(game.ClosePicker{})
	}
	if s.IconEditor != nil {
		add(game.IconSize{}, game.IconColour{}, game.IconPNG{}, game.EndIcon{}, game.EndIcon{Keep: true})
	}
	switch mode {
	case game.ModeNone, game.ModeParty, game.ModeCreate:
		if s.Creation != nil && s.Mode == game.ModeCreate {
			if c := e.creation; c != nil && c.naming {
				add(game.Reply{}) // the name
				break
			}
			add(game.Reroll{}, game.KeepCharacter{}, game.DiscardCharacter{})
			for _, f := range s.Creation.Fields {
				for j := range f.Options {
					add(game.SetChoice{Field: f.Name, N: j})
				}
			}
			break
		}
		for _, k := range e.partyFunctions() {
			switch k {
			case partyBegin:
				add(game.Begin{})
			case partyCreate:
				add(game.NewCharacter{})
			case partyAdd:
				add(game.OpenPicker{Kind: PickAdd})
				for i := range e.BuildCharList() {
					add(game.AddCharacter{Roster: i})
				}
				for _, c := range e.gameData.Commands { // adding from another game: its list, then each of its characters
					if c.From == "" || e.Cfg.Foreign == nil {
						continue
					}
					add(game.AddCharacter{Roster: -1, From: c.From})
					if p := e.picker; p != nil && p.kind == PickAddFrom && e.addFrom == c.From {
						for i := range p.n {
							add(game.AddCharacter{Roster: i, From: c.From})
						}
					}
				}
			case partyForget:
				add(game.OpenPicker{Kind: PickForget})
			case partyQuickStart:
				add(game.QuickStart{})
			case partyQuit:
				add(game.Quit{})
			}
			for n := range s.Party {
				switch k {
				case partyDrop:
					add(game.DropMember{Member: n})
				case partyRemove:
					add(game.RemoveCharacter{Roster: n})
				case partyModify:
					if e.ruleset().Check(rules.Modifiable{Who: e.partyIDs()[n]}, nil).Yes { // greyed (not allowed) for a character with experience of its own
						add(game.Modify{Member: n})
					}
					add(game.EditIcon{Member: n})
				case partyTraining:
					add(game.OpenTraining{Member: n})
				}
			}
		}
	case game.ModeExplore:
		if m := e.walkMap(); m != nil { // gobox walks: the step's parts as the engine waits for them
			at, f := image.Pt(s.Where.X, s.Where.Y), s.Where.Facing
			if e.ext.owed != nil {
				add(game.RunEvent{})
			}
			for _, d := range [4]int{0, 2, 4, 6} {
				add(game.RunEvent{Before: true, Side: d})
				if p := e.ext.pre; p != nil && p.ran && p.side == d && m.Edge(at, d).Kind != game.EdgeWall {
					add(game.Arrive{From: at, To: at.Add(sideSteps[d]), Facing: f})
				}
				if d != f && e.viewMode == 1 { // the 3D view turns; the map's walk has no facing to change (verbTurn* there do nothing)
					add(game.Face{Facing: d})
				}
			}
		}
		add(game.Search{}, game.Look{}, game.MakeCamp{})
		items()
	case game.ModeCamp:
		add(game.AdvanceTimeBy{D: time.Hour}, game.BreakCamp{}, game.Order{})
		for k := range 3 { // RestFix, RestWait, RestUntil
			add(game.CampRest{Kind: k})
		}
		for n, m := range s.Party {
			if m.CanCast {
				add(game.Cast{Member: n}, game.Memorize{Member: n}, game.Scribe{Member: n})
			}
			if n < e.partySize() { // a paladin's powers, wherever casting is; they ask whom (UsePower)
				for _, p := range e.powers(e.partyIDs()[n]) {
					switch p.Kind {
					case PowerHeal:
						add(game.UsePower{Member: n, Kind: game.PowerHeal, Target: -1})
					case PowerCure:
						add(game.UsePower{Member: n, Kind: game.PowerCure, Target: -1})
					}
				}
			}
		}
		items()
	case game.ModeCombat:
		b := s.Combat
		if aim := e.cs.aim; aim != nil && b != nil && aim.turn == nil { // an aim up on its own (a spell cast over rounds going off): its squares
			var act game.Action = game.Shoot{}
			if sa := e.cs.sa; aim.then.kind == atSpell && sa != nil {
				act = game.Cast{Member: s.Selected, Spell: int(sa.spell)}
			}
			a.Aimed = append(a.Aimed, game.Aimed{Action: act, Squares: inRange(b, max(b.Size.X, b.Size.Y)), Targets: 1})
			break
		}
		if e.battleUp() && e.ruleset().Check(rules.PartyAuto{}, nil).Yes && (b == nil || b.Turn == nil) { // the keys a computer turn polls (aiKeys): Space and Alt-M
			add(game.Manual{}, game.QuickMagic{})
		}
		if b == nil || b.Turn == nil || !e.battleUp() || e.battle().Turn == nil {
			break
		}
		can := e.battle().Turn.Can
		a.Moves = b.Reach
		for _, r := range b.Reach {
			add(game.MoveTo{To: r.At})
		}
		add(game.Flee{}) // a step off the map, as moveAttack's; the rules refuse it where there is no edge
		if slices.Contains(can, VerbAttack) {
			for _, t := range b.Targets {
				add(game.Attack{Target: t})
			}
			a.Aimed = append(a.Aimed, game.Aimed{Action: game.Shoot{}, Squares: inRange(b, max(b.Size.X, b.Size.Y)), Targets: 1})
		}
		if slices.Contains(can, VerbCast) {
			for _, sp := range b.Spells { // an aimed spell is only Aimed: cast at a square
				if !slices.ContainsFunc(b.Aims, func(a game.SpellRef) bool { return a.ID == sp.ID }) {
					add(game.Cast{Member: s.Selected, Spell: sp.ID})
				}
			}
			for _, sp := range b.Aims {
				a.Aimed = append(a.Aimed, game.Aimed{Action: game.Cast{Member: s.Selected, Spell: sp.ID}, Squares: inRange(b, sp.Range), Targets: max(sp.Targets, 1), Area: sp.Area})
			}
		}
		if s.Selected >= 0 && s.Selected < len(s.Party) { // the acting member's items (the sheet's View, O089's Use)
			if slices.Contains(can, VerbItems) {
				add(game.UseItem{Member: s.Selected, Item: rules.NoItem})
			}
			for _, it := range s.Party[s.Selected].Items {
				if slices.Contains(can, VerbItems) {
					add(game.UseItem{Member: s.Selected, Item: it.Handle})
				}
				add(e.readyAct(s.Selected, it), game.Drop{Member: s.Selected, Item: it.Handle}) // the fight's View: Ready and Drop (BUG-CLASSES 47)
			}
		}
		for _, v := range []struct { // in a fixed order: State is the same on every call (BUG-CLASSES 33)
			k   TurnVerb
			act game.Action
		}{{VerbTurn, game.TurnUndead{}}, {VerbQuick, game.Quick{}}, {VerbGuard, game.Guard{}},
			{VerbWait, game.Delay{}}, {VerbEnd, game.EndTurn{}}, {VerbBandage, game.Bandage{Member: max(s.Selected, 0)}}} {
			if slices.Contains(can, v.k) {
				add(v.act)
			}
		}
	case game.ModeShop, game.ModeTemple, game.ModeTraining, game.ModeVault:
		add(game.Leave{})
		if s.Shop != nil {
			add(game.AppraiseActions(s.Shop.Offers)...)
		}
		add(game.AppraiseActions(s.TempleOffers)...)
		if s.Shop != nil {
			for _, u := range s.Shop.Unknown { // only what the shop can identify (Identify refused the rest "not now")
				add(game.Identify{Member: u.Member, Item: u.Item})
			}
		}
		for n, m := range s.Party {
			if s.Shop != nil {
				for _, g := range s.Shop.Goods {
					add(game.Buy{Member: n, Item: g.Handle})
				}
				for _, it := range m.Items {
					add(game.Sell{Member: n, Item: it.Handle})
				}
			}
			for k, sv := range s.Temple {
				if slices.Contains(sv.Needs, n) {
					add(game.Heal{Service: k, Member: n})
				}
			}
			if t := s.Training; t != nil && slices.Contains(t.Ready, n) {
				add(game.Train{Member: n})
			}
		}
		items()
	case game.ModeLoot:
		if s.Loot != nil && !e.lootSeen { // "You found some loot!" first: Continue opens the treasure
			add(game.ContinueAction{})
			return a
		}
		add(game.Leave{}, game.Take{Item: rules.NoItem}, game.Take{Item: rules.NoItem, All: true})
		if s.Loot != nil {
			for _, g := range s.Loot.Items {
				add(game.Take{Item: g.Handle})
			}
		}
		items()
		if e.vault == nil { // Drop is the vault's Deposit: no other treasure takes the party's items
			a.Plain = slices.DeleteFunc(a.Plain, func(x game.Action) bool { _, ok := x.(game.Drop); return ok })
		}
	default: // a Scene
		switch {
		case s.Read != nil:
			add(game.Leave{})
		case s.Pending == nil:
		case s.Pending.Continue:
			add(game.ContinueAction{})
		case s.Pending.Typed, s.Pending.Number:
			add(game.Reply{})
		default:
			for i := range s.Pending.Replies {
				add(game.Reply{N: i})
			}
		}
	}
	return a
}

// sideSteps are a step's offset by side (0 N, 2 E, 4 S, 6 W).
var sideSteps = map[int]image.Point{0: {0, -1}, 2: {1, 0}, 4: {0, 1}, 6: {-1, 0}}

// inRange are the battlefield's squares within r of the turn's unit.
func inRange(b *game.Combat, r int) []image.Point {
	var me image.Point
	for _, u := range b.Units {
		if b.Turn != nil && u.ID == b.Turn.Unit {
			me = u.At
		}
	}
	var out []image.Point
	for y := range b.Size.Y {
		for x := range b.Size.X {
			if d := image.Pt(x, y).Sub(me); max(d.X, -d.X, d.Y, -d.Y) <= r {
				out = append(out, image.Pt(x, y))
			}
		}
	}
	return out
}

// allows: the state publishes a (Plain by value; Aimed by its action with
// the squares cleared; a typed Reply by kind; the free-form ones the rules
// check themselves).
func allows(al game.Allowed, a game.Action) bool {
	base := a
	switch v := a.(type) {
	case game.Reply:
		if v.Text != "" || !slices.Contains(al.Plain, a) && !slices.Contains(al.Plain, game.Action(game.Reply{N: 1})) { // typed, or a number as N (bug class 51): the one open Reply (a menu lists its N's)
			base = game.Reply{}
		}
	case game.Cast:
		v.Ats, v.Target = nil, 0
		base = v
		if v.Member == -1 { // a fight's cast is the turn's (classic sends Member -1): any of its Aimed or Plain Casts of the spell
			for _, x := range al.Aimed {
				if c, ok := x.Action.(game.Cast); ok && c.Spell == v.Spell {
					return true
				}
			}
		}
	case game.SetChoice: // a pool pick's id: the rules check the field
		if v.Text != "" {
			return true
		}
	case game.Shoot:
		base = game.Shoot{}
	case game.Attack: // at a square (a bump, an ally): any Attack offered, the rules check the square
		if v.At != nil {
			return slices.ContainsFunc(al.Plain, func(x game.Action) bool { _, ok := x.(game.Attack); return ok })
		}
	case game.IconColour: // any part, shade and value: the core checks them
		base = game.IconColour{}
	case game.IconPNG:
		base = game.IconPNG{}
	case game.AdvanceTimeBy: // any length: gobox paces camp's rest in chunks
		base = game.AdvanceTimeBy{D: time.Hour}
	case game.UseItem:
		v.Ats, v.Target, v.Spell = nil, 0, 0
		base = v
	case game.Ready:
		v.Count = 0
		base = v
	case game.EndGame: // gobox's Back to Main Menu, any time: Do refuses it where there is no game to drop
		return true
	case game.UsePower, game.SetStat, game.Adjust, game.PointBuy, game.Order, game.AddNote, game.RemoveNote, game.Give, game.ChangeClass,
		game.Memorize, game.Forget, game.Scribe, game.LayOnHands, game.Halve, game.Join, game.PoolCoins, game.SetDifficulty:
		return true
	}
	if c, ok := base.(game.Cast); ok && c.Spell != 0 { // camp offers a caster's Cast{Member}, the spell the rules' check
		c.Spell = 0
		if slices.ContainsFunc(al.Plain, func(x game.Action) bool { return reflect.DeepEqual(x, game.Action(c)) }) {
			return true
		}
	}
	for _, x := range al.Plain {
		if reflect.DeepEqual(x, a) || reflect.DeepEqual(x, base) {
			return true
		}
	}
	for _, x := range al.Aimed {
		if reflect.DeepEqual(x.Action, base) {
			return true
		}
	}
	return false
}

// Do applies a and returns what happened, or refuses it.
func (e *Engine) Do(a game.Action) (game.Outcome, error) {
	defer e.rebuild()
	e.takeSettings()
	st := e.State()
	al := game.CommonOf(st).Allowed
	if v, ok := a.(game.SetChoice); ok && v.Text != "" && e.creation == nil && e.sel >= 0 && e.sel < e.partySize() {
		pic, ok := rules.PoolPicNamed(v.Field)
		if !ok {
			return game.Outcome{}, errNotAllowed
		}
		e.do(rules.SetPoolPic{Who: e.partyIDs()[e.sel], Pic: pic, ID: rules.PicID(v.Text)}) // gobox's pool picker on the selected member, any screen
		return game.Outcome{}, nil
	}
	if !allows(al, a) {
		return game.Outcome{}, errNotAllowed
	}
	if l, ok := st.(game.Looting); ok && l.Announce { // allows: Continue only
		e.lootSeen = true
		return game.Outcome{}, nil
	}
	var out game.Outcome
	native := func(f func() error) error { // a play action the engine takes itself, run on to its next wait
		r := e.runNative(f)
		e.facts(&out, r)
		return r.Err
	}
	if ok, err := e.partyDo(a, st, native); ok {
		return out, err
	}
	if ok, err := e.frontDo(a, native); ok {
		return out, err
	}
	if e.ext.pq != nil {
		switch a := a.(type) {
		case game.Reply:
			return out, native(func() error { return e.replyPending(a.N) })
		case game.Leave:
			return out, native(e.leavePending)
		}
	}
	if ok, err := e.itemDo(a, native); ok {
		return out, e.stepOn(&out, err)
	}
	if ok, err := e.campDo(a, &out, native); ok {
		return out, e.stepOn(&out, err)
	}
	if ok, err := e.combatDo(a, al, native); ok {
		return out, e.stepOn(&out, err)
	}
	var err error
	switch a := a.(type) {
	case game.Arrive:
		if err = native(func() error { return e.arrive(a.From, a.To, a.Facing) }); err == nil {
			out.Events = append(out.Events, game.Moved{To: image.Pt(e.posX, e.posY), Facing: int(Dir(e.facing/2&3)) * 2})
		}
	case game.RunEvent:
		err = native(func() error { return e.runEvent(a.Before, a.Side) })
	case game.Face:
		err = native(func() error { return e.faceTo(a.Facing) })
	case game.Search: // the original's Search: the searching pace
		err = native(func() error { return e.exploreVerb(e.sweepCmd) })
	case game.Look: // the original's Look: this square once
		err = native(func() error { return e.exploreVerb(e.lookCmd) })
	case game.AddNote:
		err = native(func() error { return e.noteAdd(a.Text) })
	case game.RemoveNote: // BUG-CLASSES 31: a refusal, not a silent no-op
		err = native(func() error {
			n := len(e.ext.notes)
			e.noteRemove(a.ID)
			return refuseIf(len(e.ext.notes) == n, "no such note")
		})
	case game.Adjust:
		err = native(func() error { return e.adjustStat(a.Stat, a.By) })
	case game.PickPower:
		err = native(func() error { return e.pickPower(a.ID) })
	case game.StopRest:
		err = native(e.stopRest)
	case game.LootDetect:
		err = native(e.lootDetect)
	case game.Learn:
		err = native(func() error { return e.learn(a.Spell) })
	case game.SetPolicy:
		err = native(func() error {
			_, err := e.ruleset().Resolve(nil, rules.SetPolicy{Who: e.selected(), Reaction: a.ID, Policy: rules.Policy(a.Policy)}, nil)
			return err
		})
	case game.ContinueAction:
		if e.ext.intro {
			err = native(func() error { e.titleDone(); return nil })
			break
		}
		if e.sequence != "" || e.ext.ending {
			err = native(e.endDone)
			break
		}
		err = native(func() error { return e.answerPlay(0, "") })
	case game.Reply:
		err = native(func() error { return e.answerPlay(a.N, a.Text) })
	case game.Leave:
		if e.ext.read != nil { // the handout: Leave puts it away (native clears it)
			err = native(func() error { return nil })
			break
		}
		err = native(e.leave)
	default:
		return game.Outcome{}, fmt.Errorf("Pool of Radiance has no %T", a)
	}
	return out, e.stepOn(&out, err)
}

// stepOn runs the monsters' turns and the story's steps up to the next
// wait for the player, after an action taken (err nil); err passes.
func (e *Engine) stepOn(out *game.Outcome, err error) error {
	for n := 0; err == nil && n < maxSteps; n++ {
		if p := e.pace(); p != PaceFight && p != PaceStory && p != PaceRest { // gobox paces the outcome, the game runs on
			break
		}
		e.snap.quiet = true
		e.Step()
		e.advance()
		e.snap.quiet = false
		r := e.noteResult()
		e.facts(out, r)
		if r.Err != nil {
			break
		}
	}
	return err
}

// combatDo takes the fight's verbs and the item and power uses (false: a
// is none of them); a question one asks is the pending one (playask.go).
func (e *Engine) combatDo(a game.Action, al game.Allowed, native func(func() error) error) (bool, error) {
	turn := func(as ...TurnAct) (bool, error) {
		return true, native(func() error { return e.turnActs(as, false) })
	}
	inFight := e.battleUp()
	switch a := a.(type) {
	case game.MoveTo:
		return true, native(func() error {
			as, err := e.moveActs(a.To, al.Moves)
			if err != nil {
				return err
			}
			return e.turnActs(as, false)
		})
	case game.Attack:
		if a.At != nil {
			return turn(TurnAct{Verb: VerbAttack}, TurnAct{Verb: VerbAim, X: a.At.X, Y: a.At.Y})
		}
		return turn(TurnAct{Verb: VerbAttack, N: a.Target})
	case game.Shoot:
		if len(a.Ats) == 0 {
			return true, &game.Refusal{Text: "shoot: pick a square"}
		}
		if e.cs.aim != nil { // the aim up already
			return turn(TurnAct{Verb: VerbAim, X: a.Ats[0].X, Y: a.Ats[0].Y})
		}
		return turn(TurnAct{Verb: VerbAttack}, TurnAct{Verb: VerbAim, X: a.Ats[0].X, Y: a.Ats[0].Y})
	case game.Guard:
		return turn(TurnAct{Verb: VerbGuard})
	case game.Delay:
		return turn(TurnAct{Verb: VerbWait})
	case game.Quick:
		return turn(TurnAct{Verb: VerbQuick})
	case game.Manual:
		return true, native(func() error { e.aiKey = rules.AutoKeyTakeOver; return nil })
	case game.QuickMagic:
		return true, native(func() error { e.aiKey = rules.AutoKeyQuickMagic; return nil })
	case game.Bandage:
		return turn(TurnAct{Verb: VerbBandage})
	case game.EndTurn:
		return turn(TurnAct{Verb: VerbEnd})
	case game.Flee:
		return true, native(func() error {
			actor, turn := e.cs.actor, e.cs.turn
			if err := e.turnActs([]TurnAct{{Verb: VerbFlee}}, false); err != nil {
				return err
			}
			if e.ext.pq == nil && e.question() == nil && e.cs.actor == actor && e.cs.turn == turn { // no edge beside the actor: Flee: never asked (BUG-CLASSES 31)
				return &game.Refusal{Text: "Not at the edge"}
			}
			return nil
		})
	case game.TurnUndead:
		return turn(TurnAct{Verb: VerbTurn})
	case game.Cast:
		if !inFight { // camp's Cast: the member it goes on asked (castAsk), Target 0 not given
			t := a.Target
			if t == 0 {
				t = -1
			}
			return true, native(func() error {
				return e.runCall(pendingCall{kind: callPower, member: a.Member, power: game.PowerCast, spell: a.Spell, target: t})
			})
		}
		return true, native(func() error { // the spell, then a square for each target
			if e.cs.aim == nil { // up already: a spell cast over rounds going off
				if err := e.turnActs([]TurnAct{{Verb: VerbCast, N: a.Spell}}, false); err != nil || e.cs.aim == nil || e.ext.pq != nil {
					return err // a spell cast over rounds: its aim opens when it goes off (combatflow's playerAim)
				}
			}
			var as []TurnAct
			for _, at := range a.Ats {
				as = append(as, TurnAct{Verb: VerbAim, X: at.X, Y: at.Y})
			}
			return e.turnActs(as, false)
		})
	case game.UseItem:
		return true, native(func() error {
			return e.runCall(pendingCall{kind: callItem, member: a.Member, item: a.Item, spell: a.Spell, target: a.Target})
		})
	case game.UsePower:
		return true, native(func() error {
			return e.runCall(pendingCall{kind: callPower, member: a.Member, power: a.Kind, spell: a.Spell, target: a.Target})
		})
	case game.Leave:
		if e.stateMode() == game.ModeLoot && e.question() == nil { // O072 asks while treasure is left
			return true, native(func() error { return e.runCall(pendingCall{kind: callLoot}) })
		}
	}
	return false, nil
}

// browsing runs f with the shop's Buy list up beside member n, opened for
// it when it was not up and closed after.
func (e *Engine) browsing(n int, f func() error) error {
	if e.snap.invOn {
		if err := e.ShowMember(n); err != nil {
			return err
		}
		return f()
	}
	if err := e.browse(n); err != nil {
		return err
	}
	defer e.CloseItems(false)
	return f()
}

// facts hands out what the engine did: its note (r, a message), then the
// facts it recorded (e.events), the last fall held (gobox 0fa83f2).
func (e *Engine) facts(out *game.Outcome, r game.Result) {
	for _, ev := range r.Events {
		if ev.Text != "" {
			out.Events = append(out.Events, game.Narrated{Text: ev.Text})
		}
	}
	e.present()
	out.Events = append(out.Events, e.events...)
	e.events = nil
	for i := len(out.Events) - 1; i >= 0; i-- {
		if f, ok := out.Events[i].(game.Fell); ok {
			if f.Ms == 0 {
				f.Ms = fallHoldMs
				out.Events[i] = f
			}
			break
		}
	}
}

// maxSteps bounds Do's run of the engine's own steps.
const maxSteps = 10000

// fallHoldMs is how long the replay holds on the last fall.
const fallHoldMs = 500

// playSlot is the save slot Save and Load go through.
const playSlot = "PLAYSAVE"

// Save is the whole game as one slot's files.
func (e *Engine) Save() (_ []byte, err error) {
	if err := e.SaveNamed(playSlot); err != nil {
		return nil, err
	}
	st := e.store()
	defer st.DeleteJoin(playSlot, &err)
	files := map[string][]byte{}
	dir := st.Path(playSlot)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, en := range ents {
		if b, err := os.ReadFile(filepath.Join(dir, en.Name())); err == nil && !en.IsDir() {
			files[en.Name()] = b
		}
	}
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(files)
	return buf.Bytes(), err
}

// Load goes back to what Save gave (the party menu's Load: only there).
func (e *Engine) Load(b []byte) (err error) {
	defer e.rebuild()
	var files map[string][]byte
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&files); err != nil {
		return err
	}
	st := e.store()
	if err := st.Write(playSlot, func(dir string) error {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for n, f := range files {
			if err := os.WriteFile(filepath.Join(dir, n), f, 0o644); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	defer st.DeleteJoin(playSlot, &err)
	if !e.LoadNamed(playSlot, e.menuUp()) {
		return errors.New("load: refused")
	}
	return nil
}

// Slots are the original's save slots (game.Slots): gobox lists them after its own.
func (e *Engine) Slots() []game.SaveSlot {
	var l []game.SaveSlot
	for _, sl := range e.slots() {
		l = append(l, game.SaveSlot{Name: sl.Name, Time: sl.At.Format("Jan 02 15:04"), Level: e.slotLevelOf(sl)})
	}
	return l
}

// LoadSlot loads slot n of Slots.
func (e *Engine) LoadSlot(n int) error {
	defer e.rebuild()
	name := "Slot " + strconv.Itoa(n)
	if names := e.SlotsView(false).Names; n >= 0 && n < len(names) {
		name = names[n]
	}
	if !e.LoadNamed(name, e.menuUp()) {
		return &game.Refusal{Text: "load: refused"}
	}
	return nil
}

// walls is the state's 3D view as play's, nil none.
func walls(s facts) *game.MapWalls {
	w := s.Walls
	if w == nil {
		return nil
	}
	return &game.MapWalls{Side: w.Side, Fronts: w.Fronts, Pool: w.Pool, Floor: w.Floor, Sky: w.Sky,
		FloorRGBA: w.FloorRGBA, CeilRGBA: w.CeilRGBA, Backdrop: w.Backdrop, Outside: !s.Where.Inside, Sun: w.Sun, Mountains: w.Mountains, SunAz: w.SunAz, SunEl: w.SunEl}
}

// native runs f, an engine call a play action names, as Apply runs a
// decision: the adapter's own asks dropped, then on to the next wait.
func (e *Engine) runNative(f func() error) game.Result {
	e.snap.quiet = true
	defer func() { e.snap.quiet = false }()
	e.ext.read = nil
	if err := f(); err != nil {
		return game.Result{Err: refusal(err)}
	}
	e.advance()
	return e.noteResult()
}

// noteResult is the engine's note as a message, once.
func (e *Engine) noteResult() game.Result {
	note, at := e.Note()
	if note == "" || at.Equal(e.ext.noteSent) {
		return game.Result{}
	}
	e.ext.noteSent = at
	return game.Result{Events: []game.Report{{Kind: game.EvMessage, Text: note}}}
}

// partyDo takes the party menu's and the creation sheet's actions (false:
// a is none of them). gobox asks any confirmation before it sends them.
func (e *Engine) partyDo(a game.Action, st game.State, native func(func() error) error) (bool, error) {
	switch a := a.(type) {
	case game.Begin:
		return true, native(e.begin)
	case game.NewCharacter:
		return true, native(e.createCharacter)
	case game.QuickStart:
		return true, native(func() error {
			if e.partySize() > 0 { // BUG-CLASSES 64: a party already there would be doubled
				return rules.ErrNotNow
			}
			return e.QuickStart()
		})
	case game.Quit:
		return true, native(e.QuitGame)
	case game.EndGame:
		return true, native(func() error {
			if e.inParty() {
				return rules.ErrNotNow
			}
			e.mainMenu()
			return nil
		})
	case game.AddCharacter:
		return true, native(func() error {
			if a.From != "" { // another game's characters: Roster < 0 opens the list, else adds that one
				if !e.menuUp() || !slices.Contains(e.gameData.Imports, a.From) {
					return rules.ErrNotNow
				}
				e.addFrom = a.From
				if !e.OpenPicker(PickAddFrom) || a.Roster >= e.picker.n {
					e.ClosePicker()
					return &game.Refusal{Text: "no such character"}
				}
				if a.Roster < 0 {
					return nil
				}
			} else if !e.menuUp() || !e.OpenPicker(PickAdd) || a.Roster < 0 || a.Roster >= e.picker.n {
				e.ClosePicker()
				return &game.Refusal{Text: "no such character"}
			}
			n := e.partySize()
			e.Pick(a.Roster)
			e.ClosePicker()
			return refuseIf(e.partySize() == n, "not added") // already in, full, or refused to join (BUG-CLASSES 31)
		})
	case game.RemoveCharacter: // Roster: the party index
		return true, native(func() error { return e.removeMember(a.Roster) })
	case game.DropMember: // ovr16:196E in camp, ovr17:1F5A on the menu: gobox asks first
		return true, native(func() error {
			if a.Member < 0 || a.Member >= e.partySize() || !e.inCamp() && !e.inParty() {
				return rules.ErrNotNow
			}
			e.selectMember(a.Member)
			return e.DropMember(a.Member)
		})
	case game.Modify:
		return true, native(func() error { return e.modifyCharacter(a.Member) })
	case game.ChangeClass:
		return true, native(func() error { return e.runCall(pendingCall{kind: callClass, member: a.Member, target: a.Class}) })
	case game.Reroll:
		return true, native(func() error { return e.roll(false, "") })
	case game.PointBuy:
		return true, native(func() error { return e.pointBuy(a.Scores) })
	case game.SetStat:
		return true, native(func() error { return e.setStat(a.Stat, a.Text) })
	case game.KeepCharacter:
		return true, native(func() error {
			if cr := e.creation; cr != nil && cr.unnamed { // no name typed yet: ask it before saving (BUG-CLASSES 59)
				cr.naming = true
				return nil
			}
			return e.EndCreation(true)
		})
	case game.DiscardCharacter: // gobox asked abandon
		return true, native(func() error { return e.EndCreation(false) })
	case game.SetChoice:
		as, _ := st.(game.Assembling)
		if as.Draft == nil {
			return false, nil
		}
		if i := slices.IndexFunc(as.Draft.Fields, func(f game.Field) bool { return f.Name == a.Field }); i >= 0 && a.Text == "" {
			return true, native(func() error { return e.SetField(i, a.N) })
		}
		return true, native(func() error {
			return e.creationEdit(func(cr *creation) error {
				dst := cr.d.D.Pics()[a.Field]
				if dst == nil {
					return &game.Refusal{Text: "no such field"}
				}
				*dst = rules.PicID(a.Text)
				return nil
			})
		})
	}
	return false, nil
}

// frontDo takes the pickers', the icon editor's, the training hall's and
// the locked door's actions (false: a is none of them).
func (e *Engine) frontDo(a game.Action, native func(func() error) error) (bool, error) {
	switch a := a.(type) {
	case game.OpenPicker:
		return true, native(func() error { return refuseIf(!e.OpenPicker(a.Kind), "no such picker") })
	case game.Pick:
		return true, native(func() error {
			if e.picker == nil || a.N < 0 || a.N >= e.picker.n {
				return &game.Refusal{Text: "no such pick"}
			}
			e.Pick(a.N)
			return nil
		})
	case game.ClosePicker:
		return true, native(func() error { e.ClosePicker(); return nil })
	case game.EditIcon:
		return true, native(func() error {
			if e.iconEd != nil || e.partySize() == 0 {
				return rules.ErrNotNow
			}
			e.EditIcon(a.Member)
			return nil
		})
	case game.IconSize:
		return true, native(func() error { e.IconSize(); return nil })
	case game.IconColour:
		return true, native(func() error { e.IconColour(a.Part, a.Shade, a.Value); return nil })
	case game.IconPNG:
		return true, native(func() error {
			if e.iconEd == nil && e.snap.iconDrop == nil { // no icon takes it: a PNG dropped on the sheet is the portrait
				e.portraitPNG(a.Path)
				return nil
			}
			e.IconPNG(a.Path)
			return nil
		})
	case game.EndIcon:
		return true, native(func() error { e.EndIcon(a.Keep); return nil })
	case game.OpenTraining:
		return true, native(func() error { return e.openTraining(a.Member) })
	case game.CloseTraining:
		return true, native(e.closeTraining)
	case game.Train:
		return true, native(func() error {
			if !e.trainingOpen() {
				if err := e.openTraining(a.Member); err != nil {
					return err
				}
			}
			return e.train()
		})
	case game.TryDoor: // the old front's numbering, 0 exit
		return true, native(func() error { return e.TryDoor(a.Way) })
	}
	return false, nil
}

// answerPlay answers what is asked: the engine's question by its kind (n the
// option or line, text the typed line), the creation's name and abandon
// questions, the shop's Buy or Leave.
func (e *Engine) answerPlay(n int, text string) error {
	if q := e.question(); q != nil {
		var r game.Action = game.Reply{N: n}
		switch {
		case q.Typed, q.Number:
			if q.Number && text == "" {
				text = strconv.Itoa(n)
			}
			r = game.Reply{Text: text}
		case q.Continue:
			e.stopVoice() // Enter silences it (voice.go)
			r = game.ContinueAction{}
		case n < 0 || n >= len(q.Replies):
			return &game.Refusal{Text: "no such option"}
		}
		e.answer(r)
		return nil
	}
	if cr := e.creation; cr != nil && cr.naming { // Keep's name question: a name keeps the character, a blank asks again
		if e.rename(text); cr.unnamed {
			return &game.Refusal{Text: "a name is needed"}
		}
		cr.naming = false
		return e.EndCreation(true)
	}
	if cr := e.creation; cr != nil && cr.asking { // the abandon question: Yes (1) abandons, No back to the sheet
		cr.asking = false
		if n == 1 {
			return e.EndCreation(false)
		}
		return nil
	}
	if e.waiting(PendShop) {
		switch n {
		case 0:
			return e.browse(-1)
		case 1:
			return e.Leave()
		}
		return &game.Refusal{Text: "no such option"}
	}
	return rules.ErrNotNow
}

// leave backs out of what is up: the aim (back to the turn's menu,
// ovr14:3212), the name question, the sheet (asking abandon), the
// question, the handout, or the shop, temple or hall.
func (e *Engine) leave() error {
	e.stopVoice() // Esc silences it (voice.go)
	q := e.question()
	switch cr := e.creation; {
	case q == nil && e.cs.aim != nil:
		return e.turnActs([]TurnAct{{Verb: VerbClose}}, false)
	case cr != nil && cr.naming:
		cr.naming = false
	case cr != nil && q == nil:
		cr.asking = true
	case q != nil:
		return e.leaveAsk(q)
	case e.waiting(PendShop, PendTemple), e.trainingOpen(): // the temple's Exit too (Leave takes both)
		return e.Leave()
	default:
		return rules.ErrNotNow
	}
	return nil
}

// campDo takes camp's verbs (false: a is none of them): making and
// breaking camp, the rest, the marching order, memorizing and scribing.
func (e *Engine) campDo(a game.Action, out *game.Outcome, native func(func() error) error) (bool, error) {
	switch a := a.(type) {
	case game.MakeCamp:
		return true, native(func() error {
			return e.exploreVerb(func() (bool, bool) { return true, true })
		})
	case game.BreakCamp:
		return true, native(e.LeaveCamp)
	case game.AdvanceTimeBy: // the original's Rest: until all are healed and the spells learned (RestFix)
		err := native(func() error {
			return e.campRest(RestFix)
		})
		if err == nil {
			out.Events = append(out.Events, game.TimePassed{D: a.D})
		}
		return true, err
	case game.Fix:
		return true, native(func() error { return e.campRest(RestFix) })
	case game.CampRest:
		return true, native(func() error { return e.campRest(a.Kind) })
	case game.Order:
		return true, native(func() error {
			if err := e.moveMember(a.From, a.To); err != nil {
				return err
			}
			e.selectMember(a.To)
			return nil
		})
	case game.Memorize:
		return true, native(func() error { return e.planSpell(a.Member, a.Spell, 1) })
	case game.Plan:
		return true, native(func() error { return e.planSpell(a.Member, a.Spell, a.Count) })
	case game.Scribe:
		return true, native(func() error { return e.scribe(a.Member, a.Spell) })
	}
	return false, nil
}

// itemDo takes the items' and the trading places' verbs (false: a is none
// of them): ready, drop, the member shown, buy, sell, identify, heal, take.
func (e *Engine) itemDo(a game.Action, native func(func() error) error) (bool, error) {
	shown := func(n int, f func() error) (bool, error) {
		return true, native(func() error { return e.withItems(n, f) })
	}
	switch a := a.(type) {
	case game.Ready:
		return shown(a.Member, func() error { return e.Ready(a.Item, a.Count) })
	case game.Unready:
		return shown(a.Member, func() error { return e.Unready(a.Item) })
	case game.Drop: // ovr20's "Drop It?" is gobox's (BUG-CLASSES 47)
		return shown(a.Member, func() error { return e.drop(a.Item) })
	case game.SelectMember:
		return true, native(func() error {
			if a.Member < 0 || a.Member >= e.partySize() {
				return rules.ErrNotNow
			}
			if e.question() != nil && (!e.battleUp() || e.choosing()) { // a question's own step: a member pick's candidate, the selected member (not in a fight)
				e.answer(game.SelectMember{Member: a.Member})
				return nil
			}
			if !e.battleUp() { // camp's menu, exploring: the member is the selected one
				e.selectMember(a.Member)
				return nil
			}
			if e.top() == onWait { // the fight's turn: the member shown
				e.actor(a.Member)
			}
			return nil
		})
	case game.Buy: // the Buy list opens for the buy and closes after: the state stays Trading
		return true, native(func() error {
			return e.browsing(a.Member, func() error {
				gold := e.purse(false).Gold
				if err := e.Buy(a.Item); err != nil {
					return err
				}
				return refuseIf(e.purse(false).Gold == gold, e.ext.note)
			})
		})
	case game.Sell:
		return true, native(func() error { return e.browsing(a.Member, func() error { return e.sell(a.Item) }) })
	case game.Identify:
		return true, native(func() error { return e.identify(a.Member, a.Item) })
	case game.Appraise: // values one; with one valued, keeps or sells it
		return true, native(func() error {
			if e.valued != rules.NoItem {
				return e.KeepAppraised(a.Keep)
			}
			return e.Appraise(a.Offer == game.OfferJewel)
		})
	case game.Heal:
		return true, native(func() error { return e.cure(a.Member, a.Service) })
	case game.Take:
		return true, native(func() error { return e.TakeLoot(a.Item, a.All) })
	}
	return false, nil
}
