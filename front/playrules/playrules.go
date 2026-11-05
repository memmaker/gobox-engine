// Package playrules runs a play.Game on the old front: New wraps it as a
// game.Rules. Here gobox decides the flow the game no longer has: the asks,
// the aiming, the inventory up, the paging and the pacing of events.
package playrules

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"gobox/game"
	"gobox/game/play"
)

// Dir is where the slot files go, under the game's working directory.
const Dir = "saves"

type rules struct {
	g     play.Game
	dir   string
	items game.ItemsUp
	ask   *game.Ask     // gobox's own question up (a maneuver pick), nil none
	moves []play.Action // the actions ask's options stand for
	aim   *aiming
	read  *game.Read    // the handout up, kept so the pointer holds while it does
	sheet int           // the member whose sheet is up, -1 none
	rest  time.Duration // camp's rest still to go, a chunk per Step (PaceRest)
	walk  []image.Point // the auto-walk's squares still to go, one per Step (PaceWalk)
	seq   uint64        // counts the changes: the front redraws on a new one
}

// aiming is an Aimed action whose squares the player is picking.
type aiming struct {
	play.Aimed
	ats []image.Point
}

// New is g as the old front's Rules.
func New(g play.Game) game.Rules { return &rules{g: g, dir: Dir, sheet: -1} }

// Map is the game's walkable map while the party explores and no question
// is up (game.Mapped): gobox walks on it.
func (r *rules) Map() game.Map {
	if t, ok := r.g.State().(play.Exploring); ok && r.ask == nil && t.Map.Walk != nil {
		return t.Map.Walk
	}
	return nil
}

// Step goes on with what gobox paces: a chunk of camp's rest or a square
// of the auto-walk. The game never acts alone; it is sent the next action.
func (r *rules) Step() game.Result {
	switch {
	case r.rest > 0:
		d := min(r.rest, restChunk)
		r.rest -= d
		res := r.keep(play.AdvanceTimeBy{D: d})
		if res.Err != nil || len(res.Events) > 0 { // something happened meanwhile: the rest stops
			r.rest = 0
		}
		return res
	case len(r.walk) > 0:
		return r.walkStep()
	}
	return game.Result{}
}

// restChunk is the game time one paced step of camp's rest takes.
const restChunk = 10 * time.Minute

func (r *rules) Pending() *game.Ask { return r.pending(r.g.State()) }

func (r *rules) pending(ps play.State) *game.Ask {
	if r.ask != nil {
		return r.ask
	}
	if f, ok := ps.(play.ShipFight); ok && len(f.Menu) > 0 {
		a := &game.Ask{Kind: game.Choice}
		for _, t := range f.Menu {
			a.Options = append(a.Options, game.Option{Label: t, Key: firstKey(t)})
		}
		return a
	}
	sc, ok := ps.(play.Scene)
	switch {
	case !ok || sc.Kind == play.SceneHandout || sc.Kind == play.SceneTitle || sc.Kind == play.SceneSequence:
		return nil // the pager, the title and the sequence answer by their own decisions
	case sc.Kind == play.SceneHeading:
		p := ""
		if len(sc.Text) > 0 {
			p = sc.Text[len(sc.Text)-1]
		}
		return &game.Ask{Kind: game.Direction, Prompt: p, Options: []game.Option{{Label: "Done", Key: 'D', Do: game.Leave{}}}}
	case sc.Typed:
		return &game.Ask{Kind: game.Name, Max: 40}
	case len(sc.Replies) > 0:
		a := &game.Ask{Kind: game.Choice}
		for _, t := range sc.Replies {
			a.Options = append(a.Options, game.Option{Label: t, Key: firstKey(t)})
		}
		return a
	}
	return &game.Ask{Kind: game.Continue}
}

func firstKey(s string) rune {
	for _, c := range strings.ToUpper(s) {
		return c
	}
	return 0
}

// State is the typed state as the old State.
func (r *rules) State() game.State {
	ps := r.g.State()
	c := commonOf(ps)
	s := game.State{Selected: c.Selected, Purse: c.Purse, Where: c.Where, Quests: c.Quests, Lore: c.Lore,
		Allowed: allowedNames(c.Allowed), Items: r.items, Saves: r.saves()}
	for i, m := range c.Party {
		s.Party = append(s.Party, member(m, i == c.Selected, canCast(c.Allowed, i)))
	}
	s.Map.Area = c.Area
	s.PowerPick, s.Picker, s.IconEditor = c.PowerPick, c.Picker, c.IconEditor
	if n := c.Selected; n >= 0 && n < len(c.Party) && (len(c.Party[n].Powers) > 0 || len(c.Party[n].Reactions) > 0) {
		s.Sheet = &game.Sheet{Powers: c.Party[n].Powers, Reactions: c.Party[n].Reactions}
	}
	if n := r.sheet; n >= 0 && n < len(c.Party) {
		s.SheetPage = c.Party[n].Sheet
	}
	switch {
	case r.rest > 0:
		s.Pacing = &game.Pacing{Pace: game.PaceRest, Resting: true}
	case len(r.walk) > 0:
		s.Pacing = &game.Pacing{Pace: game.PaceWalk}
	}
	if r.read != nil && !isHandout(ps) {
		r.read = nil
	}
	switch t := ps.(type) {
	case play.Assembling:
		s.Mode = game.ModeParty
		if d := t.Draft; d != nil {
			s.Mode = game.ModeCreate
			s.Creation = &game.Creation{Draft: member(d.Member, true, false), Fields: d.Fields, Roster: t.Roster,
				MinHP: d.MinHP, MaxHP: d.MaxHP, Minimums: d.Minimums, Step: d.Step, Choices: d.Choices}
		}
	case play.Exploring:
		s.Mode, s.Window, s.Map = game.ModeExplore, game.WindowScene, t.Map.Automap
		if s.Map.Area == "" {
			s.Map.Area = c.Area
		}
		day := t.Day // the game's say, never the hour guessed
		s.Flags = &game.Flags{Exploring: true, Adventure: true, Day: day, Inside: c.Where.Inside, CanCast: canCast(c.Allowed, c.Selected),
			IconDrop: c.IconEditor != nil}
		s.Door = t.Door
		w := t.Map.Walls
		s.Walls = &game.Walls{Key: c.Area, Side: w.Side, Fronts: w.Fronts, Pool: w.Pool, Floor: w.Floor, Sky: w.Sky,
			FloorRGBA: w.FloorRGBA, CeilRGBA: w.CeilRGBA, Sun: w.Outside && day, Mountains: w.Outside && day && w.Mountains,
			SunAz: w.SunAz, SunEl: w.SunEl, Backdrop: w.Backdrop}
	case play.Camping:
		s.Mode = game.ModeCamp
		s.RestHeals = s.Allowed["AdvanceTimeBy"]
		s.RestStudies = s.RestHeals
	case play.Fighting:
		s.Mode, s.Combat = game.ModeCombat, r.combat(t)
	case play.Scene:
		s.Mode, s.Window = game.ModeStory, game.WindowText
		if t.Kind == play.SceneHandout {
			if r.read == nil || r.read.Title != t.Handout || !slices.Equal(r.read.Lines, t.Lines) {
				r.read = &game.Read{Title: t.Handout, Lines: t.Lines}
			}
			s.Read = r.read
			break
		}
		switch t.Kind {
		case play.SceneTitle:
			s.Mode, s.Intro = game.ModeTitle, t.Intro
		case play.SceneSequence:
			s.Sequence = t.Sequence
		case play.SceneResults:
			s.Results = t.Results
		case play.SceneScreen:
			s.Screen = t.Screen
		}
		s.Page = page(t)
		if t.Full {
			s.Flags = &game.Flags{Full: true}
		}
		for _, l := range t.Layers {
			if l.FrameMs > 0 && s.Pacing == nil {
				s.Pacing = &game.Pacing{Pace: game.PaceAnim}
			}
		}
		if len(t.Text) > 0 {
			s.LastText = t.Text[len(t.Text)-1]
		}
	case play.Trading:
		switch {
		case t.Kind == play.TradeVault:
			s.Mode, s.Vault = game.ModeVault, t.Vault
			if s.Vault == nil {
				s.Vault = &game.Vault{}
			}
		case t.Kind == play.TradeTraining, t.Training != nil:
			s.Mode, s.Training = game.ModeTraining, t.Training
		case t.Kind == play.TradeTemple, len(t.Services) > 0:
			s.Mode, s.Temple, s.TempleName, s.TempleOffers = game.ModeTemple, t.Services, t.Name, t.Offers
		default:
			s.Mode = game.ModeShop
			s.Shop = &game.Shop{Name: t.Name, Goods: t.Goods, Prices: t.Prices, Buys: t.Buys, Offers: t.Offers,
				Pooled: t.Pooled, Reloads: s.Allowed["Reload"], IdentifyPrice: t.Identify, Unknown: t.Unknown, Identified: t.Known}
		}
		if r.items != game.ItemsNone {
			s.Items = game.ItemsShop
		}
	case play.Looting:
		s.Mode, s.Loot = game.ModeLoot, &game.Loot{Items: t.Items, Money: t.Money, Status: t.Status}
		if r.items != game.ItemsNone {
			s.Items = game.ItemsLoot
		}
	case play.Over:
		s.Mode, s.LastText = game.ModeTitle, t.Text
	case play.ShipFight:
		f := t.Fight
		s.Mode, s.Screen = game.ModeStory, &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: &f, ShipSheet: t.Sheet}
	}
	if s.Pacing == nil {
		s.Pacing = &game.Pacing{}
	}
	s.Pacing.Seq = r.seq
	s.Pending = r.pending(ps)
	return s
}

func isHandout(ps play.State) bool {
	sc, ok := ps.(play.Scene)
	return ok && sc.Kind == play.SceneHandout
}

func commonOf(ps play.State) play.Common {
	switch t := ps.(type) {
	case play.Assembling:
		return t.Common
	case play.Exploring:
		return t.Common
	case play.Camping:
		return t.Common
	case play.Fighting:
		return t.Common
	case play.Scene:
		return t.Common
	case play.Trading:
		return t.Common
	case play.Looting:
		return t.Common
	case play.Over:
		return t.Common
	case play.ShipFight:
		return t.Common
	}
	return play.Common{}
}

// oldNames are the old Allowed keys of the renamed actions.
var oldNames = map[string]string{"Give": "Trade", "Take": "TakeLoot", "AdvanceTimeBy": "Rest",
	"MakeCamp": "Camp", "BreakCamp": "LeaveCamp", "SetChoice": "SetIcon", "KeepCharacter": "KeepDraft",
	"AddCharacter": "AddChar", "MoveTo": "Move", "Shoot": "Ranged", "DoManeuver": "Maneuver", "Delay": "Wait"}

func name(a play.Action) string { return reflect.TypeOf(a).Name() }

// allowedNames is al by action name, new and old, plus gobox's own Select.
func allowedNames(al play.Allowed) map[string]bool {
	m := map[string]bool{"Select": true}
	add := func(a play.Action) {
		n := name(a)
		m[n] = true
		if o := oldNames[n]; o != "" {
			m[o] = true
		}
	}
	for _, a := range al.Plain {
		add(a)
	}
	for _, a := range al.Aimed {
		add(a.Action)
	}
	return m
}

func canCast(al play.Allowed, n int) bool {
	for _, a := range al.Plain {
		if c, ok := a.(play.Cast); ok && c.Member == n {
			return true
		}
	}
	for _, a := range al.Aimed {
		if c, ok := a.Action.(play.Cast); ok && c.Member == n {
			return true
		}
	}
	return false
}

func member(m play.Member, chosen, cast bool) game.Member {
	g := game.Member{Name: m.Name, Race: m.Race, Class: m.Class, Alignment: m.Alignment, Status: m.Status,
		Levels: m.Levels, Stats: m.Stats, BaseStats: m.BaseStats, Str00: m.Str00, HP: m.HP, MaxHP: m.MaxHP,
		AC: m.AC, THAC0: m.THAC0, Move: m.Move, Exp: uint32(m.Exp), LevelExp: m.LevelExp, NextExp: m.NextExp,
		Items: m.Items, ItemCap: m.ItemCap, Spells: m.Spells, Slots: m.Slots, SlotsMax: m.SlotsMax,
		PortraitHeadID: m.PortraitHeadID, PortraitBodyID: m.PortraitBodyID, IconHeadID: m.IconHeadID,
		IconBodyID: m.IconBodyID, IconColors: m.IconColours, Okay: m.Okay, Foe: m.Foe, Chosen: chosen, CanCast: cast,
		ClassText: m.ClassText, LevelText: m.LevelText, SpellsReady: m.SpellsReady, Memo: m.Memo}
	for _, e := range m.Effects {
		g.Effects = append(g.Effects, game.Line{Text: e})
	}
	return g
}

// Text colours are gobox's: plain ink, and one per Tone (EGA white, light
// cyan, light red, yellow, light magenta, white). A game's own Ink is not
// taken.
var ink = color.RGBA{255, 255, 255, 255}

var tones = map[game.Tone]color.RGBA{game.ToneText: ink, game.ToneOwn: {85, 255, 255, 255}, game.ToneBad: {255, 85, 85, 255},
	game.TonePrompt: {255, 255, 85, 255}, game.ToneWarn: {255, 85, 255, 255}, game.ToneLoud: ink}

func page(sc play.Scene) *game.Page {
	p := &game.Page{}
	if sc.Speaker != "" {
		p.Pane = append(p.Pane, []game.TextRun{{Text: sc.Speaker + ":", Fg: ink}})
	}
	for _, t := range sc.Text {
		p.Pane = append(p.Pane, []game.TextRun{{Text: t, Fg: ink}})
	}
	for _, pic := range sc.Pictures {
		p.Layers = append(p.Layers, game.Layer{Picture: pic})
	}
	p.Layers = append(p.Layers, sc.Layers...)
	for _, t := range sc.Placed {
		for len(p.View) < t.At.Y {
			p.View = append(p.View, nil)
		}
		p.View[t.At.Y-1] = append(p.View[t.At.Y-1], game.TextRun{X: t.At.X, Text: t.Text, Fg: tones[t.Tone]})
	}
	return p
}

// turnOptions are the old combat menu's decisions for the new actions.
var turnOptions = []struct {
	new string
	do  game.Decision
}{
	{"MoveTo", game.Move{}}, {"Attack", game.Attack{}}, {"Shoot", game.Ranged{}}, {"Guard", game.Guard{}},
	{"Cast", game.Cast{}}, {"UseItem", game.Items{}}, {"DoManeuver", game.Maneuver{}},
	{"TurnUndead", game.TurnUndead{}}, {"Bandage", game.Bandage{}}, {"Delay", game.Wait{}},
	{"Quick", game.Quick{}}, {"Flee", game.Flee{}}, {"EndTurn", game.EndTurn{}},
}

func (r *rules) combat(f play.Fighting) *game.Combat {
	b, al := f.Field, f.Allowed
	c := &game.Combat{Size: b.Size, Squares: b.Squares, Tiles: b.Tiles, FloorTile: b.FloorTile, Floor: b.Floor,
		Round: b.Round, Dungeon: b.Dungeon, Reach: al.Moves, Effects: b.Clouds}
	var spells []game.SpellRef
	for _, u := range b.Units {
		c.Units = append(c.Units, game.Unit{ID: u.ID, Name: u.Name, At: u.At, Facing: u.Facing, Foe: u.Foe,
			HP: u.HP, MaxHP: u.MaxHP, Art: u.Art, Dead: u.Dead, Gone: u.Gone, Size: u.Size, Squares: u.Squares,
			AC: u.AC, Status: u.Status, Dmg: u.Dmg, Member: u.Member, Active: u.ID == b.Turn, Hit: u.Hit, Pose: u.Pose})
		if u.ID == b.Turn && u.Member >= 0 && u.Member < len(f.Party) {
			spells = f.Party[u.Member].Spells
		}
	}
	t := &game.Turn{Unit: b.Turn, Aim: r.aim != nil, Moves: b.Moves, Casting: b.Casting, Weapon: b.Weapon, Missile: b.Missile}
	c.Turn, c.Panel = t, b.Turn
	if r.aim != nil {
		c.Marks = r.aim.ats
		if cs, ok := r.aim.Action.(play.Cast); ok {
			t.Spell = spellRef(spells, cs.Spell).Name
		}
	}
	allowed := allowedNames(al)
	for _, o := range turnOptions {
		if allowed[o.new] && o.new != "MoveTo" {
			t.Options = append(t.Options, game.Option{Label: o.new, Key: firstKey(o.new), Do: o.do})
		}
	}
	for _, v := range brVerbs { // as Buck Rogers' rules put them: an answer by hotkey
		if allowed[name(v.a)] {
			t.Options = append(t.Options, game.Option{Label: v.label, Key: v.key, Do: game.Answer{N: int(v.key)}})
		}
	}
	named := map[int]game.TurnManeuver{}
	for _, m := range b.Maneuvers {
		named[m.ID] = m
	}
	for _, a := range al.Plain {
		switch a := a.(type) {
		case play.Attack:
			c.Targets = append(c.Targets, a.Target)
		case play.Cast:
			c.Spells = append(c.Spells, spellRef(spells, a.Spell))
		case play.DoManeuver:
			t.Maneuvers = append(t.Maneuvers, maneuver(named, a.ID, 0))
		}
	}
	for _, a := range al.Aimed {
		switch x := a.Action.(type) {
		case play.Cast:
			sp := spellRef(spells, x.Spell)
			sp.Targets, sp.Area = a.Targets, a.Area
			c.Spells, c.Aims = append(c.Spells, sp), append(c.Aims, sp)
		case play.DoManeuver:
			t.Maneuvers = append(t.Maneuvers, maneuver(named, x.ID, a.Area))
		}
	}
	return c
}

// maneuver is the rules' maneuver id, named as they name it.
func maneuver(named map[int]game.TurnManeuver, id, area int) game.TurnManeuver {
	if m, ok := named[id]; ok {
		return m
	}
	return game.TurnManeuver{ID: id, Name: fmt.Sprint("Maneuver ", id), Area: area}
}

// brVerbs are Buck Rogers' turn verbs with the hotkeys its menu answers by.
var brVerbs = []struct {
	label string
	key   rune
	a     play.Action
}{
	{"Target", 'T', play.Target{}}, {"View", 'V', play.ViewTarget{}}, {"Intimidate", 'I', play.Intimidate{}},
	{"Aid", 'A', play.Aid{}}, {"Sprint", 'S', play.Sprint{}}, {"Dodge", 'D', play.Dodge{}}, {"Jetpack", 'J', play.Jetpack{}},
}

func spellRef(sp []game.SpellRef, id int) game.SpellRef {
	for _, s := range sp {
		if s.ID == id {
			return s
		}
	}
	return game.SpellRef{ID: id}
}

// Apply turns an old decision into a play action, or acts on gobox's own
// state for the decisions that only steer the screen.
func (r *rules) Apply(d game.Decision) game.Result {
	ps := r.g.State()
	switch d := d.(type) {
	case game.Answer:
		if f, ok := ps.(play.Fighting); ok && r.ask == nil {
			for _, v := range brVerbs {
				if d.N == int(v.key) && allowedNames(f.Allowed)[name(v.a)] {
					return r.do(v.a)
				}
			}
		}
		return r.answer(ps, d)
	case game.Cancel:
		r.ask, r.moves, r.aim = nil, nil, nil
		return game.Result{}
	case game.Items, game.OpenItems:
		r.items = game.ItemsOwn
		return game.Result{}
	case game.CloseItems:
		r.items = game.ItemsNone
		return game.Result{}
	case game.OpenSheet:
		r.sheet = d.Member
		return game.Result{}
	case game.CloseSheet:
		r.sheet = -1
		return game.Result{}
	case game.SetSpeed, game.Resize, game.Browse, game.ViewMember, game.TitlePicture, game.TitleCredits, game.TitleFlash, game.LootLook:
		return game.Result{} // the screen's
	case game.TitleDone, game.SequenceDone:
		return r.do(play.Continue{})
	case game.Speak: // gobox voices the page itself
		var res game.Result
		for _, k := range d.Keys {
			res.Events = append(res.Events, game.Event{Kind: game.EvSound, Sound: k})
		}
		return res
	case game.StopRest:
		r.rest = 0
		return game.Result{}
	case game.Rest:
		if _, ok := ps.(play.Camping); ok { // gobox paces the rest in chunks, Step by Step
			r.rest = time.Duration(d.Minutes) * time.Minute
			if r.rest == 0 {
				r.rest = 8 * time.Hour
			}
			r.seq++
			return game.Result{}
		}
	case game.WalkTo:
		return r.walkTo(ps, d.At)
	case game.Select:
		return r.do(play.SelectMember{Member: d.N})
	case game.ShowMember:
		return r.do(play.SelectMember{Member: d.Member})
	case game.Maneuver:
		return r.maneuvers(commonOf(ps).Allowed)
	case game.Ranged:
		return r.aimOrDo(commonOf(ps).Allowed, play.Shoot{})
	case game.Cast:
		a := play.Cast{Member: d.Member, Spell: d.Spell, Target: d.Target, Ats: d.Ats}
		if len(a.Ats) == 0 && d.At != (image.Point{}) {
			a.Ats = []image.Point{d.At}
		}
		return r.aimOrDo(commonOf(ps).Allowed, a)
	case game.UseItem:
		return r.aimOrDo(commonOf(ps).Allowed, play.UseItem{Member: d.Member, Item: d.Item, Target: d.Target, Ats: d.Ats, Spell: d.Spell})
	case game.Aim:
		return r.aimAt(d.At)
	case game.Save:
		return game.Result{Err: r.save(d.Name)}
	case game.Load:
		return r.load(d.Slot)
	case game.DeleteSlot:
		return game.Result{Err: r.remove(d.Name)}
	case game.SetField:
		if a, ok := ps.(play.Assembling); ok && a.Draft != nil && d.Field >= 0 && d.Field < len(a.Draft.Fields) {
			return r.do(play.SetChoice{Field: a.Draft.Fields[d.Field].Name, N: d.Option})
		}
		return game.Result{Err: errors.New("no such field")}
	case game.EndCreation:
		if d.Keep {
			return r.do(play.KeepCharacter{})
		}
		return r.do(play.DiscardCharacter{})
	}
	if a := action(ps, d); a != nil {
		return r.do(a)
	}
	return game.Result{Err: fmt.Errorf("%T is not in the typed contract", d)}
}

// action is d's play action where it is one to one (renamed or not), nil none.
func action(ps play.State, d game.Decision) play.Action {
	switch d := d.(type) {
	case game.Arrive:
		return play.Arrive{From: d.From, To: d.To, Facing: d.Facing}
	case game.RunEvent:
		return play.RunEvent{Before: d.Before, Side: d.Side}
	case game.Face:
		return play.Face{Facing: d.Facing}
	case game.Search:
		return play.Search{}
	case game.Look:
		return play.Look{}
	case game.Move:
		return play.MoveTo{To: d.To}
	case game.Attack:
		return play.Attack{Target: d.Target}
	case game.Ready:
		return play.Ready{Member: d.Member, Item: d.Item, Count: d.Count}
	case game.Unready:
		return play.Unready{Member: d.Member, Item: d.Item}
	case game.Drop:
		return play.Drop{Member: d.Member, Item: d.Item}
	case game.Trade:
		return play.Give{From: d.From, To: d.To, Item: d.Item}
	case game.Halve:
		return play.Halve{Member: d.Member, Item: d.Item}
	case game.Join:
		return play.Join{Member: d.Member, Item: d.Item}
	case game.Buy:
		return play.Buy{Member: d.Member, Item: d.Item}
	case game.Sell:
		return play.Sell{Member: d.Member, Item: d.Item}
	case game.Identify:
		return play.Identify{Member: d.Member, Item: d.Item}
	case game.Appraise:
		return play.Appraise{Member: d.Member, Item: d.Item, Keep: d.Keep}
	case game.Repair:
		return play.Repair{Member: d.Member, Item: d.Item}
	case game.Reload:
		return play.Reload{Member: d.Member, Item: d.Item}
	case game.PoolCoins:
		return play.PoolCoins{Member: d.Member}
	case game.TakeLoot:
		return play.Take{Item: d.Item, All: d.All}
	case game.Leave, game.LootLeave:
		return play.Leave{}
	case game.QuickStart:
		return play.QuickStart{}
	case game.Quit:
		return play.Quit{}
	case game.Modify:
		return play.Modify{Member: d.Member}
	case game.Roll:
		return play.Reroll{Minimal: d.Minimal}
	case game.Rename:
		return play.Rename{Name: d.Name}
	case game.OpenPicker:
		return play.OpenPicker{Kind: d.Kind}
	case game.Pick:
		return play.Pick{N: d.N}
	case game.ClosePicker:
		return play.ClosePicker{}
	case game.EditIcon:
		return play.EditIcon{Member: d.Member}
	case game.IconSize:
		return play.IconSize{}
	case game.IconColour:
		return play.IconColour{Part: d.Part, Shade: d.Shade, Value: d.Value}
	case game.IconRemove:
		return play.IconRemove{}
	case game.IconPNG:
		return play.IconPNG{Path: d.Path}
	case game.EndIcon:
		return play.EndIcon{Keep: d.Keep}
	case game.UsePower:
		return play.UsePower{Member: d.Member, Kind: d.Kind, Spell: d.Spell, Target: d.Target}
	case game.PickPower:
		return play.PickPower{ID: d.ID}
	case game.SetPolicy:
		return play.SetPolicy{ID: d.ID, Policy: d.Policy}
	case game.TryDoor:
		return play.TryDoor{Way: d.Way}
	case game.Mark:
		return play.Mark{At: d.At, Text: d.Text}
	case game.Unmark:
		return play.Unmark{At: d.At}
	case game.Plan:
		return play.Plan{Member: d.Member, Spell: d.Spell, Count: d.Count}
	case game.CampRest:
		return play.CampRest{Kind: d.Kind}
	case game.OpenTraining:
		return play.OpenTraining{Member: d.Member}
	case game.CloseTraining:
		return play.CloseTraining{}
	case game.Learn:
		return play.Learn{Spell: d.Spell}
	case game.LootDetect:
		return play.LootDetect{}
	case game.LootShare:
		return play.LootShare{}
	case game.LootStatus:
		return play.LootStatus{}
	case game.Rest:
		m := time.Duration(d.Minutes) * time.Minute
		if m == 0 {
			m = 8 * time.Hour // gobox's default rest
		}
		return play.AdvanceTimeBy{D: m}
	case game.Camp:
		return play.MakeCamp{}
	case game.LeaveCamp:
		return play.BreakCamp{}
	case game.Train:
		return play.Train{Member: d.Member}
	case game.Heal:
		return play.Heal{Service: d.Service, Member: d.Member}
	case game.Cure:
		return play.Heal{Service: d.Service, Member: d.Member}
	case game.Create:
		return play.SetChoice{Field: d.Field, N: d.N, Text: d.Text}
	case game.SetIcon:
		return play.SetChoice{Field: d.Field, N: d.N, Text: d.ID}
	case game.SetStat:
		return play.SetStat{Stat: d.Stat, Text: d.Text}
	case game.Adjust:
		return play.Adjust{Stat: d.Stat, By: d.By}
	case game.PointBuy:
		return play.PointBuy{Scores: d.Scores}
	case game.KeepDraft:
		return play.KeepCharacter{}
	case game.AddChar:
		return play.AddCharacter{Roster: d.Roster, From: d.From}
	case game.DropChar:
		return play.DropMember{Member: d.Member}
	case game.DropMember:
		return play.DropMember{Member: d.Member}
	case game.NewCharacter:
		return play.NewCharacter{}
	case game.Reroll:
		return play.Reroll{}
	case game.DiscardCharacter:
		return play.DiscardCharacter{}
	case game.RemoveCharacter:
		return play.RemoveCharacter{Roster: d.Roster}
	case game.ChangeClass:
		return play.ChangeClass{Member: d.Member, Class: d.Class}
	case game.SetDifficulty:
		return play.SetDifficulty{N: d.N}
	case game.Begin:
		return play.Begin{}
	case game.Wait:
		return play.Delay{}
	case game.Guard:
		return play.Guard{}
	case game.EndTurn:
		return play.EndTurn{}
	case game.Quick:
		return play.Quick{All: d.All}
	case game.Flee:
		return play.Flee{}
	case game.TurnUndead:
		return play.TurnUndead{}
	case game.Bandage:
		return play.Bandage{Member: d.Member}
	case game.LayOnHands:
		return play.LayOnHands{Member: d.Member}
	case game.Fix:
		return play.Fix{Member: d.Member}
	case game.Memorize:
		return play.Memorize{Member: d.Member, Spell: d.Spell}
	case game.Forget:
		return play.Forget{Member: d.Member, Spell: d.Spell}
	case game.Scribe:
		return play.Scribe{Member: d.Member, Spell: d.Spell}
	case game.Order:
		return play.Order{From: d.From, To: d.To}
	case game.AddNote:
		return play.AddNote{Text: d.Text}
	case game.RemoveNote:
		return play.RemoveNote{ID: d.ID}
	case game.Call:
		return play.Call{Member: d.Member}
	case game.EndGame:
		return play.EndGame{}
	}
	return nil
}

// facing is the step from a to b as a facing, else f (a step off the map).

// answer answers gobox's own ask, or the scene's.
func (r *rules) answer(ps play.State, d game.Answer) game.Result {
	if r.ask != nil {
		moves := r.moves
		r.ask, r.moves = nil, nil
		if d.N < 0 || d.N >= len(moves) {
			return game.Result{Err: errors.New("no such choice")}
		}
		return r.aimOrDo(commonOf(ps).Allowed, moves[d.N])
	}
	if _, ok := ps.(play.ShipFight); ok {
		return r.do(play.Crew{N: d.N})
	}
	sc, ok := ps.(play.Scene)
	switch {
	case !ok:
		return game.Result{Err: errors.New("nothing is asked")}
	case sc.Kind == play.SceneHeading:
		return r.do(play.Heading{Dir: d.N})
	case sc.Kind == play.SceneHandout:
		return r.do(play.Continue{})
	case sc.Typed:
		return r.do(play.Reply{Text: d.Text})
	case len(sc.Replies) > 0:
		return r.do(play.Reply{N: d.N})
	}
	return r.do(play.Continue{})
}

// maneuvers asks which of the turn's maneuvers.
func (r *rules) maneuvers(al play.Allowed) game.Result {
	r.ask, r.moves = &game.Ask{Kind: game.Choice, Prompt: "Maneuver"}, nil
	add := func(a play.Action) {
		if m, ok := a.(play.DoManeuver); ok {
			r.moves = append(r.moves, m)
			r.ask.Options = append(r.ask.Options, game.Option{Label: fmt.Sprint("Maneuver ", m.ID)})
		}
	}
	for _, a := range al.Plain {
		add(a)
	}
	for _, a := range al.Aimed {
		add(a.Action)
	}
	if len(r.moves) == 0 {
		r.ask = nil
		return game.Result{Err: errors.New("no maneuvers")}
	}
	return game.Result{}
}

// aimOrDo starts aiming a when it is an Aimed action sent without squares,
// else sends it.
func (r *rules) aimOrDo(al play.Allowed, a play.Action) game.Result {
	if len(ats(a)) == 0 {
		for _, x := range al.Aimed {
			if reflect.DeepEqual(withAts(x.Action, nil), withAts(a, nil)) {
				r.aim = &aiming{Aimed: x}
				return game.Result{}
			}
		}
	}
	return r.do(a)
}

func (r *rules) aimAt(at image.Point) game.Result {
	m := r.aim
	if m == nil {
		return game.Result{Err: errors.New("not aiming")}
	}
	if !slices.Contains(m.Squares, at) {
		return game.Result{Err: errors.New("out of range")}
	}
	m.ats = append(m.ats, at)
	if len(m.ats) < max(m.Targets, 1) {
		return game.Result{}
	}
	r.aim = nil
	return r.do(withAts(m.Action, m.ats))
}

func ats(a play.Action) []image.Point {
	switch a := a.(type) {
	case play.Cast:
		return a.Ats
	case play.UseItem:
		return a.Ats
	case play.Shoot:
		return a.Ats
	case play.DoManeuver:
		return a.Ats
	}
	return nil
}

func withAts(a play.Action, at []image.Point) play.Action {
	switch x := a.(type) {
	case play.Cast:
		x.Ats = at
		return x
	case play.UseItem:
		x.Ats = at
		return x
	case play.Shoot:
		x.Ats = at
		return x
	case play.DoManeuver:
		x.Ats = at
		return x
	}
	return a
}

// do sends a to the game, and the player's new move ends what gobox was
// pacing (a rest, an auto-walk).
func (r *rules) do(a play.Action) game.Result {
	r.rest, r.walk = 0, nil
	return r.keep(a)
}

// keep sends a to the game: a refusal is Err, else the outcome as events.
func (r *rules) keep(a play.Action) game.Result {
	out, err := r.g.Do(a)
	if err != nil {
		return game.Result{Err: err}
	}
	r.ask, r.moves, r.aim = nil, nil, nil
	r.seq++
	return game.Result{Events: events(out.Events)}
}

// sides are the facings of the four steps.
var sides = map[image.Point]int{{0, -1}: 0, {1, 0}: 2, {0, 1}: 4, {-1, 0}: 6}

// walkTo plans the auto-walk to at over the game's map, shortest first;
// Step walks it a square at a time.
func (r *rules) walkTo(ps play.State, at image.Point) game.Result {
	x, ok := ps.(play.Exploring)
	if !ok || x.Map.Walk == nil {
		return game.Result{Err: errors.New("not walking")}
	}
	m, from := x.Map.Walk, image.Pt(x.Where.X, x.Where.Y)
	prev := map[image.Point]image.Point{from: from}
	for q := []image.Point{from}; len(q) > 0; q = q[1:] {
		for d, f := range sides {
			n := q[0].Add(d)
			if _, seen := prev[n]; seen || !n.In(image.Rectangle{Max: m.Size()}) || m.Edge(q[0], f).Kind == game.EdgeWall {
				continue
			}
			prev[n] = q[0]
			q = append(q, n)
		}
	}
	if _, ok := prev[at]; !ok || at == from {
		return game.Result{Err: errors.New("no way there")}
	}
	var path []image.Point
	for p := at; p != from; p = prev[p] {
		path = append([]image.Point{p}, path...)
	}
	r.walk = path
	r.seq++
	return game.Result{}
}

// walkStep walks the next square: the script ahead, the step, the square's
// event, as gobox's own walk sends them. It stops where the party does not
// stand exploring after, or where the game says anything.
func (r *rules) walkStep() game.Result {
	x, ok := r.g.State().(play.Exploring)
	if !ok {
		r.walk = nil
		return game.Result{}
	}
	from, to := image.Pt(x.Where.X, x.Where.Y), r.walk[0]
	r.walk = r.walk[1:]
	side, ok := sides[to.Sub(from)]
	if !ok {
		r.walk = nil
		return game.Result{Err: errors.New("off the walk")}
	}
	var res game.Result
	for _, a := range []play.Action{play.RunEvent{Before: true, Side: side}, play.Arrive{From: from, To: to, Facing: side}, play.RunEvent{}} {
		st := r.keep(a)
		res.Events, res.Err = append(res.Events, st.Events...), st.Err
		if _, still := r.g.State().(play.Exploring); st.Err != nil || !still || len(st.Events) > 0 && !onlyMoved(st.Events) {
			r.walk = nil
			break
		}
	}
	return res
}

func onlyMoved(es []game.Event) bool {
	for _, e := range es {
		if e.Kind != game.EvMove {
			return false
		}
	}
	return true
}

// The pacing gobox gives a fight's moments.
const strikeMs, castMs, shipMs = 300, 500, 350 // shipMs: ten 35 ms half-steps of a blink

// events are the outcome's facts as the old events, paced by gobox.
func events(es []play.Event) (out []game.Event) {
	for _, e := range es {
		switch e := e.(type) {
		case play.Narrated:
			out = append(out, game.Event{Kind: game.EvMessage, Text: e.Text, Wait: true})
		case play.Said:
			out = append(out, game.Event{Kind: game.EvSaid, Role: e.Who, Text: e.Text, Wait: true})
		case play.Shown:
			out = append(out, game.Event{Kind: game.EvPicture, Picture: e.Picture})
		case play.Sounded:
			out = append(out, game.Event{Kind: game.EvSound, Sound: e.Sound})
		case play.MusicChange:
			out = append(out, game.Event{Kind: game.EvSong, Sound: e.Song})
		case play.Walked:
			for _, st := range e.Path {
				out = append(out, game.Event{Kind: game.EvMove, Unit: e.Unit, To: st.At})
			}
		case play.Turned:
			out = append(out, game.Event{Kind: game.EvMove, Unit: e.Unit, N: e.Facing})
		case play.Attacked:
			out = append(out, game.Event{Kind: game.EvAttack, Unit: e.From, Target: e.To, Pose: 1, Ms: strikeMs})
			if !e.Hit {
				out = append(out, game.Event{Kind: game.EvMiss, Unit: e.From, Target: e.To})
				break
			}
			out = append(out, game.Event{Kind: game.EvHit, Unit: e.From, Target: e.To},
				game.Event{Kind: game.EvDamage, Unit: e.From, Target: e.To, N: e.Damage})
		case play.Shot:
			out = append(out, game.Event{Kind: game.EvMissile, Unit: e.From, To: e.To, Sprite: []game.Picture{e.Art}, Pose: 1, Ms: strikeMs})
		case play.CastSpell:
			ev := game.Event{Kind: game.EvSpell, Unit: e.From, N: e.Spell, Pose: 1, Ms: castMs}
			if len(e.Ats) > 0 {
				ev.To = e.Ats[0]
			}
			out = append(out, ev)
		case play.Struck:
			out = append(out, game.Event{Kind: game.EvDamage, Target: e.Unit, N: e.Damage})
		case play.Healed:
			out = append(out, game.Event{Kind: game.EvDamage, Target: e.Unit, N: -e.HP})
		case play.Fell:
			out = append(out, game.Event{Kind: game.EvDown, Unit: e.Unit, Target: e.Unit})
		case play.Moved:
			out = append(out, game.Event{Kind: game.EvMove, Unit: -1, To: e.To, N: e.Facing})
		case play.Refused:
			out = append(out, game.Event{Kind: game.EvRefused, Text: e.Why})
		case play.Ended:
			out = append(out, game.Event{Kind: game.EvReset})
		case play.SaveOffered:
			out = append(out, game.Event{Kind: game.EvSaves})
		case play.ShipShown:
			f := e.Fight
			out = append(out, game.Event{Kind: game.EvPicture, Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: &f}, Ms: shipMs})
		}
		// Fled, TimePassed, Gained, Lost, JournalNote: the state shows them.
	}
	return out
}

// Slot files: one per name, Dir/<name>.sav, the game's Save bytes.

func (r *rules) path(name string) (string, error) {
	if name == "" || name != filepath.Base(name) || strings.HasPrefix(name, ".") {
		return "", errors.New("bad slot name")
	}
	return filepath.Join(r.dir, name+".sav"), nil
}

func (r *rules) save(name string) error {
	p, err := r.path(name)
	if err != nil {
		return err
	}
	b, err := r.g.Save()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return err
	}
	tmp := p + ".tmp" // a crash mid-write leaves the old slot whole
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (r *rules) remove(name string) error {
	p, err := r.path(name)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

func (r *rules) load(slot int) game.Result {
	sl := r.saves()
	if slot < 0 || slot >= len(sl) {
		return game.Result{Err: errors.New("no such slot")}
	}
	p, err := r.path(sl[slot].Name)
	if err != nil {
		return game.Result{Err: err}
	}
	b, err := os.ReadFile(p)
	if err == nil {
		err = r.g.Load(b)
	}
	if err != nil {
		return game.Result{Err: err}
	}
	r.ask, r.moves, r.aim, r.items, r.read = nil, nil, nil, game.ItemsNone, nil
	return game.Result{Events: []game.Event{{Kind: game.EvReset}}}
}

// saves are the slot files, newest first.
func (r *rules) saves() []game.SaveSlot {
	es, _ := os.ReadDir(r.dir)
	type slot struct {
		game.SaveSlot
		t time.Time
	}
	var sl []slot
	for _, e := range es {
		n, ok := strings.CutSuffix(e.Name(), ".sav")
		if !ok || e.IsDir() {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		sl = append(sl, slot{game.SaveSlot{Name: n, Time: fi.ModTime().Format("2006-01-02 15:04")}, fi.ModTime()})
	}
	slices.SortFunc(sl, func(a, b slot) int { return b.t.Compare(a.t) })
	out := make([]game.SaveSlot, len(sl))
	for i, s := range sl {
		out[i] = s.SaveSlot
	}
	return out
}
