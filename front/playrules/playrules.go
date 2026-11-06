// Package playrules runs a play.Game on the old front: New wraps it as a
// game.Rules. Here gobox decides the flow the game no longer has: the asks,
// the aiming, the inventory up, the paging and the pacing of events.
package playrules

import (
	"cmp"
	"errors"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gobox/game"
	"gobox/game/play"
)

// Dir is the slot folder when the game names none and the user has no
// config dir: under the working directory.
const Dir = "saves"

// saveDir is where slot files go: the game's SaveDir, else the user's
// config dir (gobox/<ID>/saves), never the game's data folder. A run with
// its own HOME so gets its own saves.
func saveDir(d game.Data) string {
	if d.SaveDir != "" {
		return d.SaveDir
	}
	c, err := os.UserConfigDir()
	if err != nil {
		return Dir
	}
	return filepath.Join(c, "gobox", cmp.Or(d.ID, "game"), "saves")
}

type rules struct {
	g       play.Game
	dir     string
	items   game.ItemsUp
	ask     *game.Ask     // gobox's own question up (a maneuver pick), nil none
	moves   []play.Action // the actions ask's options stand for
	aim     *aiming
	cursor  image.Point   // the free cursor's square
	read    *game.Read    // the handout up, kept so the pointer holds while it does
	sheet   int           // the member whose sheet is up, -1 none
	itemsOf int           // the member whose items are up (OpenItems), -1: the acting or selected one
	speedUp bool          // the Speed menu is up (SpeedMenu)
	speed   int           // the game speed (SetSpeed, 0-9): a fight's moment holds speed × 100 ms, as the originals' message delay
	rest    time.Duration // camp's rest still to go, a chunk per Step (PaceRest)
	walk    []image.Point // the auto-walk's squares still to go, one per Step (PaceWalk)
	seq     uint64        // counts the changes: the front redraws on a new one
	data    game.Data     // the series' data (verb keys, texts)
	log     []string      // the fight's message log, from the outcomes' text
	note    game.Note     // the status line's toast: the last refusal
	walls   *game.Walls   // the 3D view last explored: a scene over it keeps it
	camMap  game.Automap  // the map last explored: camp keeps it
	title   struct {      // the title as gobox steps it
		n       int
		credits bool
		page    int
	}
}

// aiming is an Aimed action whose squares the player is picking.
type aiming struct {
	play.Aimed
	ats    []image.Point
	byUnit map[image.Point]int // an Attack aimed at a unit the game offers by id: its squares' ids
}

// New is g as the old front's Rules; d is the game's Data
// (its series' verb keys and texts).
func New(g play.Game, d game.Data) game.Rules {
	r := &rules{g: g, dir: saveDir(d), sheet: -1, data: d, speed: defaultSpeed}
	return r
}

// who is n, or the selected member when n is -1.
func who(n int, ps play.State) int {
	if n < 0 {
		return commonOf(ps).Selected
	}
	return n
}

// restCap is the longest rest al allows at once (0: any), false none.
func restCap(al play.Allowed) (time.Duration, bool) {
	for _, a := range al.Plain {
		if t, ok := a.(play.AdvanceTimeBy); ok {
			return t.D, true
		}
	}
	return 0, false
}

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
		if m, _ := restCap(commonOf(r.g.State()).Allowed); m > 0 {
			d = min(d, m)
		}
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

// pending is the question up; where none is built and Continue is the
// only way on, a Continue ask, so the player is never stuck.
func (r *rules) pending(ps play.State) *game.Ask {
	if _, fight := ps.(play.Fighting); fight && r.items != game.ItemsNone { // the items opened over a question (a View's Gear) are the top layer; closing them shows it again
		return nil
	}
	if r.speedUp {
		return &game.Ask{Kind: game.Choice, Prompt: "Speed " + strconv.Itoa(r.speed),
			Options: []game.Option{{Label: "Faster", Key: 'F'}, {Label: "Slower", Key: 'S'}, {Label: "Exit", Key: 'E'}}}
	}
	if a := r.asked(ps); a != nil {
		return a
	}
	if sc, ok := ps.(play.Scene); ok && (sc.Kind == play.SceneHandout || sc.Kind == play.SceneTitle || sc.Kind == play.SceneSequence) {
		return nil
	}
	al := commonOf(ps).Allowed
	if f, ok := ps.(play.Fighting); ok && r.loneAim(f) {
		return &game.Ask{Kind: game.Square, Prompt: f.Status}
	}
	if len(al.Plain) == 1 && len(al.Aimed) == 0 && al.Plain[0] == (play.Continue{}) {
		return &game.Ask{Kind: game.Continue}
	}
	return nil
}

func (r *rules) asked(ps play.State) *game.Ask {
	if r.ask != nil {
		return r.ask
	}
	c := commonOf(ps)
	if q := c.Ask; q != nil {
		return r.question(q, q.List)
	}
	if a, ok := ps.(play.Assembling); ok {
		if d := a.Draft; d != nil {
			return r.draftAsk(d)
		}
		return r.partyMenu(a)
	}
	if f, ok := ps.(play.ShipFight); ok && len(f.Menu) > 0 {
		o := r.options(f.Menu, false)
		for i := range o {
			if i < len(f.Keys) && f.Keys[i] != 0 {
				o[i].Key = f.Keys[i] // the game's key (BR2's C for Character), not the label's
			}
		}
		return &game.Ask{Kind: game.Choice, Options: o}
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
		return &game.Ask{Kind: game.Name, Prompt: c.Status, Max: cmp.Or(sc.Max, 40), Default: sc.Default}
	case len(sc.Replies) > 0:
		return &game.Ask{Kind: game.Choice, Prompt: c.Status, Default: sc.Default, NoPick: sc.NoPick, NoEsc: sc.NoEsc, Options: r.options(sc.Replies, c.Status != "" || sc.Kind == play.SceneScreen)} // a screen's or the roster's pick is a list; a menu over the view keeps its keys
	}
	return &game.Ask{Kind: game.Continue}
}

// question is the rules' question as the old Ask.
func (r *rules) question(q *play.Question, list bool) *game.Ask {
	a := &game.Ask{Kind: game.Choice, Prompt: q.Text, Default: q.Default, Max: q.Max, NoEsc: q.NoEsc, NoPick: q.NoPick}
	switch {
	case q.Continue:
		a.Kind = game.Continue
	case q.Heading:
		a.Kind, a.Options = game.Direction, []game.Option{{Label: "Done", Key: 'D', Do: game.Leave{}}}
		return a
	case q.YesNo:
		a.Kind, a.Yes = game.YesNo, q.Default == 1
	case q.Typed:
		a.Kind = game.Name
	case q.Number:
		a.Kind = game.Number
	}
	if a.Kind != game.YesNo { // a yes/no is the front's confirm: Options on it would be extra moves (game.Ask), as the old path had none
		a.Options = r.options(q.Replies, list)
	}
	return a
}

// draftAsk is the creation step's question (nil: the sheet's fields).
func (r *rules) draftAsk(d *play.Draft) *game.Ask {
	if !asks(d) {
		return nil
	}
	return r.question(&play.Question{Text: d.Prompt, Replies: d.Choices, YesNo: d.YesNo, Typed: d.Typed, Number: d.Number, Default: d.Default, Max: d.Max}, true)
}

// asks says whether d puts a question (else the sheet's fields are up):
// a pick, a yes/no, a typed line or a number.
func asks(d *play.Draft) bool { return len(d.Choices) > 0 || d.YesNo || d.Typed || d.Number }

// partyFuncs are the party menu's functions: label, hotkey and the action
// that opens it ("": gobox's own).
var partyFuncs = []struct {
	label string
	key   rune
	act   string
}{
	{"Create Character", 'C', "NewCharacter"}, {"Add Character", 'A', "AddCharacter"},
	{"Remove Character", 'R', "RemoveCharacter"}, {"Forget Character", 'F', "Forget"}, {"Modify Character", 'M', "Modify"},
	{"Train Character", 'T', "OpenTraining"}, {"Human Change Class", 'H', "ChangeClass"},
	{"View Character", 'V', ""}, {"Icon Select", 'I', "EditIcon"}, {"Drop Character", 'D', "DropMember"},
	{"Load Saved Game", 'L', ""}, {"Save Current Game", 'S', ""},
	{"Quick Start", 'Q', "QuickStart"}, {"Begin Adventuring", 'B', "Begin"}, {"Exit to DOS", 'E', "Quit"},
}

// partyMenu is the party menu gobox builds from what a allows; the labels
// are gobox's, a series renames one in Data.Texts ("Remove Character":
// "Purge Character").
func (r *rules) partyMenu(a play.Assembling) *game.Ask {
	al := allowedNames(a.Allowed)
	acts := map[string]string{}
	order := r.data.PartyMenu
	for _, f := range partyFuncs {
		acts[f.label] = f.act
		if r.data.PartyMenu == nil {
			order = append(order, game.MenuKey{Label: f.label, Key: f.key})
		}
	}
	q := &game.Ask{Kind: game.Choice}
	for _, f := range order {
		act, known := acts[f.Label]
		on := known && al[act]
		switch f.Label {
		case "Forget Character": // gobox's roster picker over OpenPicker{PickForget}
			on = slices.Contains(a.Allowed.Plain, play.Action(play.OpenPicker{Kind: game.PickForget}))
		case "View Character":
			on = len(a.Party) > 0
		case "Save Current Game":
			on = len(a.Party) > 0 && !a.NoSave
		case "Load Saved Game":
			on = !a.NoLoad // gobox's slot screen, empty or not
		}
		if on {
			q.Options = append(q.Options, game.Option{Label: cmp.Or(r.data.Texts[f.Label], f.Label), Key: f.Key})
		}
	}
	return q
}

// key is the hotkey gobox invents for a label: its first letter, none
// when used already (a clash never happens).
func key(s string, used map[rune]bool) rune {
	for _, c := range strings.ToUpper(s) {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			continue // a mark such as "* FLAVIUS" is no key
		}
		if used[c] {
			return 0
		}
		used[c] = true
		return c
	}
	return 0
}

// options are labels as options with invented keys; a vertical list
// (list) has none where the series' lists have none (Data.NoHotkeys), but
// its way out ("Done", "Exit", "Leave") keeps one.
func (r *rules) options(labels []string, list bool) []game.Option {
	used := map[rune]bool{}
	var o []game.Option
	for _, t := range labels {
		k := rune(0)
		if !list || !r.data.NoHotkeys || exitLabel[strings.ToLower(t)] {
			k = key(t, used)
		}
		o = append(o, game.Option{Label: t, Key: k})
	}
	return o
}

// State is the typed state as the old State.
func (r *rules) State() game.State {
	ps := r.g.State()
	c := commonOf(ps)
	s := game.State{Selected: c.Selected, Purse: c.Purse, Where: c.Where, Quests: c.Quests, Lore: c.Lore,
		Allowed: allowedNames(c.Allowed), Items: r.items, Saves: r.saves()}
	if _, camp := ps.(play.Camping); camp && !s.Allowed["Rest"] {
		s.Allowed["NoRest"] = true // no timed rest here: camp offers none (classic Camp.NoRest)
	}
	for i, m := range c.Party {
		s.Party = append(s.Party, member(m, i == c.Selected && !r.data.Unmarked, m.CanCast || canCast(c.Allowed, i)))
		s.Party[i].Card = card(m, c.Allowed, i)
		s.Party[i].Page = m.Sheet // the rules' page, for gobox's look-only sheet (a fight's View)
	}
	s.Map = c.Automap
	if x, ok := ps.(play.Exploring); ok && c.Automap.Size == (image.Point{}) {
		s.Map = x.Map.Automap
	}
	if _, ok := ps.(play.Exploring); ok {
		r.camMap = s.Map
	} else if _, ok := ps.(play.Camping); ok && len(s.Map.Known) == 0 && len(r.camMap.Known) > 0 { // none, or a bare one (the party's square): the map explored
		s.Map = r.camMap
	}
	if s.Map.Area == "" {
		s.Map.Area = c.Area
	}
	if c.Picture != nil {
		s.Picture = *c.Picture
	}
	if pt := c.Portrait; pt != nil { // gobox composes it from the head and body
		s.Story = []game.Picture{pt.Head, pt.Body}
	}
	s.Page = &game.Page{Layers: c.Layers}
	if c.Picture != nil && len(c.Layers) == 0 { // the front draws a page's layers, never State.Picture: the picture is the page's one layer
		s.Page.Layers = []game.Layer{{Picture: *c.Picture}}
	}
	for _, t := range c.Text {
		s.Page.Pane = append(s.Page.Pane, r.run(t))
		s.LastText = t
	}
	if c.Status != "" {
		s.Page.Prompt = r.run(c.Status)
	}
	s.Note = r.note
	_, camp := ps.(play.Camping)
	s.Flags = &game.Flags{Adventure: c.Adventure, Ordering: camp && s.Allowed["Order"], Sweeping: c.Where.Searching,
		Inside: c.Where.Inside, CanCast: c.Selected >= 0 && c.Selected < len(s.Party) && s.Party[c.Selected].CanCast,
		IconDrop: c.IconEditor != nil, Day: c.Day, NoSave: c.NoSave, NoLoad: c.NoLoad}
	if c.Picture != nil || len(c.Layers) > 0 || c.Portrait != nil {
		s.Window = game.WindowPicture
	}
	s.PowerPick, s.Picker, s.IconEditor = c.PowerPick, c.Picker, c.IconEditor
	if n := c.Selected; n >= 0 && n < len(c.Party) && (len(c.Party[n].Powers) > 0 || len(c.Party[n].Reactions) > 0) {
		s.Sheet = &game.Sheet{Powers: c.Party[n].Powers, Reactions: c.Party[n].Reactions}
	}
	if n := r.sheet; n >= 0 && n < len(c.Party) {
		if s.SheetPage = c.Party[n].Sheet; s.SheetPage == nil { // no page from the rules: gobox's own from the member
			s.SheetPage = game.MemberPage(s.Party[n])
		}
		if s.Sheet == nil && s.SheetPage != nil { // the sheet is up though the member has no powers: the front shows a sheet only with one
			s.Sheet = &game.Sheet{}
		}
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
			if d.Member.Sheet != nil {
				s.SheetPage = d.Member.Sheet // the rules' figures (age, ...) for the sheet being made
			}
		}
	case play.Exploring:
		s.Mode, s.Window = game.ModeExplore, game.WindowScene
		if c.Ask != nil && !(c.Ask.Continue && len(c.Ask.Replies) == 0) || t.Door != nil { // a question (or the door menu) over the view, as the old path asks it; a bare Continue (no reply to show) stays ModeExplore, where any key goes on (ModeAsk drew an empty menu that took no Enter)
			s.Mode = game.ModeAsk
		}
		switch {
		case c.Picture != nil || len(c.Layers) > 0 || c.Portrait != nil: // a speaker's portrait over the view too
			s.Window = game.WindowPicture
		case t.Map.Overland:
			s.Window = game.WindowMap
		}
		s.Flags.Exploring, s.Flags.Day = c.Ask == nil, t.Day || c.Day // the game's say, never the hour guessed
		s.Door = t.Door
		if w := t.Map.Walls; w.Side != nil { // no squares: no 3D view (the overland)
			s.Walls = walls(w)
		}
		r.walls = s.Walls
	case play.Camping:
		s.Mode, s.Walls = game.ModeCamp, r.walls          // camp is a menu over the 3D view last explored
		s.RestHeals, s.RestStudies = t.Heals, t.Study > 0 // the rules' say, not whether resting is allowed
	case play.Fighting:
		s.Mode, s.Combat = game.ModeCombat, r.combat(t)
		s.Items = r.items                                                                                                   // the turn member's items (Gear), front state over the fight; their actions are Allowed's
		if m := cmp.Or(r.itemsOf+1, activeMember(t.Field)+1) - 1; r.items != game.ItemsNone && m >= 0 && m < len(s.Party) { // the acting member's own, never another's (OpenItems{-1})
			s.Selected = m
		}
	case play.Scene:
		s.Mode, s.Window = game.ModeStory, game.WindowText
		if len(t.Pictures)+len(t.Layers) > 0 {
			s.Window = game.WindowPicture
		}
		s.Walls = r.walls // the 3D view stays under the page, as on the old path
		if t.Walls != nil && t.Walls.Side != nil {
			s.Walls = walls(*t.Walls)
		}
		switch t.Kind {
		case play.SceneHandout: // read over the page: its pictures and layers stay
			if r.read == nil || r.read.Title != t.Handout || !slices.Equal(r.read.Lines, t.Lines) {
				r.read = &game.Read{Title: t.Handout, Lines: t.Lines}
			}
			s.Read = r.read
		case play.SceneTitle:
			s.Mode, s.Intro, s.Window = game.ModeTitle, t.Intro, game.WindowScene
			s.Flags.Credits, s.Flags.CreditPage = r.title.credits, r.title.page
		case play.SceneSequence:
			s.Sequence, s.Window = t.Sequence, game.WindowScene // full screen: no text window (the old zero)
		case play.SceneResults:
			s.Results = t.Results
		case play.SceneScreen:
			s.Screen = t.Screen
		}
		pg := r.page(t)
		if t.Kind == play.SceneTitle && !r.title.credits {
			n := r.title.n
			if n == 0 && t.Intro != nil && len(t.Intro.Pictures) > 0 { // before gobox's first step: that step's picture already, no black frame
				n = t.Intro.Pictures[0]
			}
			pg.Layers = append(pg.Layers, t.Steps[n].Layers...)
		}
		s.Page.Pane, s.Page.View, s.Page.Layers = append(s.Page.Pane, pg.Pane...), pg.View, append(s.Page.Layers, pg.Layers...)
		s.Flags.Full = t.Full
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
			s.Shop = &game.Shop{Name: t.Name, Goods: t.Goods, Buys: t.Buys, Offers: t.Offers,
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
	switch ps.(type) {
	case play.Exploring, play.Scene, play.Camping:
	default:
		r.walls = nil // a fight, the party menu: no 3D view to go back to
	}
	s.Holding = holding(ps)
	if s.Pacing == nil {
		s.Pacing = &game.Pacing{}
	}
	s.Pacing.Seq = r.seq
	s.Pending = r.pending(ps)
	s.Mode = askMode(s.Mode, s.Pending, ps)
	return s
}

// askMode is ModeAsk where a question is up over a scene (a menu, a list, a
// name) or over the creation sheet (Common.Ask): the old front shows a
// question only in ModeAsk, so in ModeStory or ModeCreate it stayed hidden.
func askMode(m game.Mode, a *game.Ask, ps play.State) game.Mode {
	if a == nil || a.Kind == game.Continue {
		return m
	}
	switch {
	case m == game.ModeStory:
		if sc, ok := ps.(play.Scene); ok && sc.Kind == play.SceneScreen && sc.Screen != nil {
			return m // a typed screen answers by its own keys
		}
		return game.ModeAsk
	case m == game.ModeCreate && commonOf(ps).Ask != nil:
		return game.ModeAsk
	}
	return m
}

// holding is what the old front waits on in ps.
func holding(ps play.State) game.Hold {
	switch t := ps.(type) {
	case play.Exploring:
		if t.Door != nil && t.Ask == nil {
			return game.HoldDoor
		}
	case play.Trading:
		if t.Kind == play.TradeTemple || t.Kind == play.TradeShop && len(t.Services) > 0 {
			return game.HoldTemple
		}
		if t.Kind == play.TradeShop || t.Kind == play.TradeVault {
			return game.HoldShop
		}
	case play.Looting:
		return game.HoldLoot
	case play.Fighting:
		return game.HoldFight
	}
	return game.HoldNone // a story page waits on its question, not a hold (HoldStory is the finale's pictures)
}

// wallsKey names w's look: a new one has the 3D view built again.
func wallsKey(w play.Walls) string {
	h := fnv.New64a()
	fmt.Fprint(h, w.Side, w.Fronts, w.Pool, w.Floor, w.Sky, w.FloorRGBA, w.CeilRGBA, w.Backdrop)
	return strconv.FormatUint(h.Sum64(), 36)
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

// card is member n's context menu from what al lets n do, as POR's:
// View, Items, then the allowed member actions in a fixed order. No game
// builds it, so every game's cards get the same menu.
func card(m play.Member, al play.Allowed, n int) []game.Entry {
	es := []game.Entry{{Name: "View", Verb: "view"}}
	has := map[string]bool{"items": len(m.Items) > 0}
	for _, a := range append(slices.Clone(al.Plain), aimedActions(al)...) {
		switch t := a.(type) {
		case play.Ready:
			has["items"] = has["items"] || t.Member == n
		case play.Drop:
			has["items"] = has["items"] || t.Member == n
		case play.UseItem:
			has["items"] = has["items"] || t.Member == n
		case play.Cast:
			has["cast"] = has["cast"] || t.Member == n
		case play.Memorize:
			has["memorize"] = has["memorize"] || t.Member == n
		case play.Plan:
			has["memorize"] = has["memorize"] || t.Member == n
		case play.Scribe:
			has["scribe"] = has["scribe"] || t.Member == n
		case play.LayOnHands:
			has["heal"] = has["heal"] || t.Member == n
		case play.EditIcon:
			has["icon"] = has["icon"] || t.Member == n
		case play.DropMember:
			has["drop"] = has["drop"] || t.Member == n
		}
	}
	for _, e := range []game.Entry{{Name: "Items", Verb: "items"}, {Name: "Cast", Verb: "cast"}, {Name: "Heal", Verb: "heal"}, {Name: "Memorize", Verb: "memorize"},
		{Name: "Scribe", Verb: "scribe"}, {Name: "Icon", Verb: "icon"}, {Name: "Drop", Verb: "drop"}} {
		if has[e.Verb] {
			es = append(es, e)
		}
	}
	return es
}

func aimedActions(al play.Allowed) []play.Action {
	var out []play.Action
	for _, a := range al.Aimed {
		out = append(out, a.Action)
	}
	return out
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
		ClassText: m.ClassText, LevelText: m.LevelText, SpellsReady: m.SpellsReady, Memo: m.Memo, Icon: m.Icon, Portrait: m.Portrait, Age: m.Age}
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

// run is a line of the rules' text: its colour unset, so the front gives
// it gobox's body text colour, in the text glyphs (Upcase: fonts.Glyph
// keeps a font's own lower case, else capitals). Upcase is not the case: a
// false one (once set for Data.Lowercase) drew PoD's lower-case letters as
// the charset's graphics.
func (r *rules) run(t string) []game.TextRun {
	return []game.TextRun{{Text: t, Upcase: true}}
}

func walls(w play.Walls) *game.Walls {
	return &game.Walls{Key: wallsKey(w), Side: w.Side, Fronts: w.Fronts, Pool: w.Pool, Floor: w.Floor, Sky: w.Sky,
		FloorRGBA: w.FloorRGBA, CeilRGBA: w.CeilRGBA, Sun: w.Sun, Mountains: w.Mountains,
		SunAz: w.SunAz, SunEl: w.SunEl, Backdrop: w.Backdrop}
}

func (r *rules) page(sc play.Scene) *game.Page {
	p := &game.Page{}
	if sc.Speaker != "" {
		p.Pane = append(p.Pane, r.run(sc.Speaker+":"))
	}
	for _, t := range sc.Text {
		p.Pane = append(p.Pane, r.run(t))
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
// turnOptions are the combat menu's verbs: the action, gobox's label and
// the old decision (a series relabels one in Data.Texts, keys it in
// Data.VerbKeys).
// verbWord is the Gold Box combat menus' word for a gobox verb label, so
// a VerbOrder in the original's words (Use, Delay, Done) places them.
var verbWord = map[string]string{"Items": "Use", "Wait": "Delay", "End": "Done"}

var turnOptions = []struct {
	new, label string
	do         game.Decision
}{
	{"MoveTo", "Move", game.Move{}}, {"Attack", "Attack", game.Attack{}}, {"Shoot", "Ranged", game.Ranged{}},
	{"Guard", "Guard", game.Guard{}}, {"Cast", "Cast", game.Cast{}}, {"UseItem", "Items", game.Items{}},
	{"DoManeuver", "Maneuver", game.Maneuver{}}, {"TurnUndead", "Turn", game.TurnUndead{}},
	{"Bandage", "Bandage", game.Bandage{}}, {"Delay", "Wait", game.Wait{}}, {"Quick", "Quick", game.Quick{}},
	{"Flee", "Flee", game.Flee{}}, {"EndTurn", "End", game.EndTurn{}},
}

// activeMember is the party index of the unit whose turn it is, -1 none.
func activeMember(b play.Battlefield) int {
	for _, u := range b.Units {
		if u.ID == b.Turn {
			return u.Member
		}
	}
	return -1
}

// loneAim starts the pick of squares when an Aimed action is all the turn
// allows (a jetpack's flight, a throw): gobox asks for the square at once.
func (r *rules) loneAim(f play.Fighting) bool {
	al := f.Allowed
	if len(al.Aimed) != 1 || slices.ContainsFunc(al.Plain, func(a play.Action) bool { return a != play.Action(play.Leave{}) }) || len(al.Moves) > 0 {
		return false // Leave beside it is Esc (BR2's T)
	}
	if r.aim == nil {
		r.aim = &aiming{Aimed: al.Aimed[0]}
		if sq := al.Aimed[0].Squares; len(sq) > 0 {
			r.cursor = sq[0] // the cursor on the first target
		}
	}
	return true
}

func (r *rules) combat(f play.Fighting) *game.Combat {
	r.loneAim(f)
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
	c.Turn, c.Panel, c.Log, c.CursorOn, c.Cursor = t, b.Turn, r.log, r.aim != nil, r.cursor
	if slices.ContainsFunc(f.Text, func(l string) bool { return strings.TrimSpace(l) != "" }) { // the rules' message pane is the fight's log; a blank one never hides the blows
		c.Log = f.Text
	}
	if r.aim != nil {
		c.Marks = r.aim.ats
		if cs, ok := r.aim.Action.(play.Cast); ok {
			t.Spell = spellRef(spells, cs.Spell).Name
		}
	}
	allowed := allowedNames(al)
	used := map[rune]bool{} // the series' keys first; gobox's own invented around them
	for _, k := range r.data.VerbKeys {
		used[k] = true
	}
	for _, v := range brVerbs {
		if allowed[name(v.a)] {
			used[r.verbKey(v.label, v.key)] = true
		}
	}
	for _, o := range turnOptions {
		if allowed[o.new] || o.new == "MoveTo" && len(al.Moves) > 0 {
			l := cmp.Or(r.data.Texts[o.label], o.label)
			k, ok := r.data.VerbKeys[o.label]
			if !ok {
				k = key(l, used)
			}
			t.Options = append(t.Options, game.Option{Label: l, Key: k, Do: o.do})
		}
	}
	if slices.Contains(al.Plain, play.Action(play.Quick{All: true})) { // the original's Quick for the whole party: its own entry, keyed by the series (VerbKeys, VerbChords) or clicked
		t.Options = append(t.Options, game.Option{Label: cmp.Or(r.data.Texts["Quick All"], "Quick All"), Key: r.data.VerbKeys["Quick All"], Do: game.Quick{All: true}})
	}
	for _, v := range brVerbs { // as Buck Rogers' rules put them: an answer by hotkey
		if k := r.verbKey(v.label, v.key); allowed[name(v.a)] {
			t.Options = append(t.Options, game.Option{Label: v.label, Key: k, Do: game.Answer{N: int(k)}})
		}
	}
	if k, ok := r.data.VerbKeys["View"]; ok && activeMember(b) >= 0 && len(t.Options) > 0 && !allowed["ViewTarget"] { // the series' View (keyed in VerbKeys): the active member's sheet, front state (BR's View is its rules' ViewTarget)
		l, m := cmp.Or(r.data.Texts["View"], "View"), activeMember(b)
		t.Options = append(t.Options, game.Option{Label: l, Key: k, Do: game.ViewMember{Member: m}})
	}
	called := slices.ContainsFunc(t.Options, func(o game.Option) bool { return o.Label == "Aim" })                           // a verb the series calls Aim (SotSB's Ranged)
	if _, keyed := r.data.VerbKeys["Aim"]; !r.data.NoFreeAim && len(t.Options) > 0 && !called && (keyed || r.data.FreeAim) { // the series' Aim (keyed or placed), gobox's free cursor; no menu ("Continue battle?"): no Aim either
		t.Options = append(t.Options, game.Option{Label: "Aim", Do: game.Aim{}}) // the free cursor, gobox's own
	}
	if ord := r.data.VerbOrder; len(ord) > 0 { // the series' order; the verbs it leaves out after, as gobox has them
		at := func(o game.Option) int {
			if i := slices.Index(ord, o.Label); i >= 0 {
				return i
			}
			if i := slices.Index(ord, verbWord[o.Label]); i >= 0 { // gobox's label, the Gold Box menus' word for it
				return i
			}
			return len(ord)
		}
		slices.SortStableFunc(t.Options, func(a, b game.Option) int { return at(a) - at(b) })
	}
	for i, o := range t.Options {
		t.Options[i].Chords = r.data.VerbChords[o.Label]
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
	if r.aim != nil { // aiming: the foes on the aim's squares are what Tab cycles (none in reach: every foe), the cursor on the nearest
		c.Targets = nil
		var from image.Point
		for _, u := range c.Units {
			if u.Active {
				from = u.At
			}
		}
		covers := func(u game.Unit, sq []image.Point) bool { // any square it covers (a big foe's At may lie outside the reach)
			return slices.ContainsFunc(append([]image.Point{u.At}, u.Squares...), func(q image.Point) bool { return slices.Contains(sq, q) })
		}
		var live []game.Unit
		for _, u := range c.Units {
			if u.Foe && !u.Dead && !u.Gone {
				live = append(live, u)
			}
		}
		if inReach := slices.DeleteFunc(slices.Clone(live), func(u game.Unit) bool { return !covers(u, r.aim.Squares) }); len(inReach) > 0 {
			live = inReach
		}
		best, kept := -1, len(r.aim.ats) > 0 || slices.Contains(r.aim.Squares, r.cursor)
		for _, u := range live {
			c.Targets = append(c.Targets, u.ID)
			kept = kept || covers(u, []image.Point{r.cursor})
			if d := u.At.Sub(from); best < 0 || d.X*d.X+d.Y*d.Y < best {
				best, c.Cursor = d.X*d.X+d.Y*d.Y, u.At
			}
		}
		if kept || best < 0 {
			c.Cursor = r.cursor
		}
		r.cursor = c.Cursor
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

// verbKey is the series' hotkey for a combat verb (Data.VerbKeys), else k.
func (r *rules) verbKey(label string, k rune) rune {
	if x, ok := r.data.VerbKeys[label]; ok {
		return x
	}
	return k
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
		if f, ok := ps.(play.Fighting); ok && r.ask == nil && f.Ask == nil {
			for _, v := range brVerbs {
				if d.N == int(r.verbKey(v.label, v.key)) && allowedNames(f.Allowed)[name(v.a)] {
					return r.do(v.a)
				}
			}
		}
		return r.answer(ps, d)
	case game.Cancel:
		if r.speedUp {
			r.speedUp = false
			r.seq++
			return game.Result{}
		}
		if r.ask == nil && r.aim == nil && isHandout(ps) { // Esc on a handout leaves it, as reading to its end does
			return r.do(onward(commonOf(ps).Allowed))
		}
		lone := r.aim != nil && r.ask == nil && slices.Contains(commonOf(ps).Allowed.Plain, play.Action(play.Leave{}))
		if q := commonOf(ps).Ask; q != nil && r.ask == nil && r.aim == nil { // Esc on the rules' menu is its Exit, as the originals' (BR2's View menu)
			if i := slices.IndexFunc(q.Replies, func(l string) bool { return strings.EqualFold(strings.TrimSpace(l), "Exit") }); i >= 0 {
				return r.answer(ps, game.Answer{N: i})
			}
		}
		r.ask, r.moves, r.aim = nil, nil, nil
		if _, fight := ps.(play.Fighting); fight && lone { // Esc on the aim the game asks: its Leave
			return r.do(play.Leave{})
		}
		return game.Result{}
	case game.Arrive: // the step the rules allow from and to there, with their facing: gobox's guess of it (kept on a step back) is not the rules' say
		for _, a := range commonOf(ps).Allowed.Plain {
			if s, ok := a.(play.Arrive); ok && s.From == d.From && s.To == d.To {
				return r.do(s)
			}
		}
		return r.do(play.Arrive{From: d.From, To: d.To, Facing: d.Facing})
	case game.Items, game.OpenItems, game.Browse: // Browse: a shop's goods beside the member's items (State.Items = ItemsShop)
		r.items, r.itemsOf = game.ItemsOwn, -1
		if o, ok := d.(game.OpenItems); ok {
			r.itemsOf = o.Member
		}
		return game.Result{}
	case game.CloseItems:
		r.items = game.ItemsNone
		if _, fight := ps.(play.Fighting); !d.All && !fight { // Tab: from the items to the member's sheet, as the old path's
			r.sheet = who(-1, ps)
		}
		return game.Result{}
	case game.OpenSheet:
		r.sheet = who(d.Member, ps)
		return game.Result{}
	case game.CloseSheet:
		r.sheet = -1
		return game.Result{}
	case game.ViewMember: // a shop's, temple's or pool's View: the member's sheet, front state; CloseSheet goes back to the screen under it
		r.sheet = who(d.Member, ps)
		return game.Result{}
	case game.SetSpeed:
		r.setSpeed(d.N)
		return game.Result{}
	case game.SpeedMenu:
		r.speedUp = true
		r.seq++
		return game.Result{}
	case game.Resize, game.TitleFlash, game.LootLook:
		return game.Result{} // the screen's (the title's flash is the front's palette)
	case game.TitlePicture: // gobox steps the title over the scene's Steps
		r.title.n, r.title.credits = d.N, false
		r.seq++
		if sc, ok := ps.(play.Scene); ok && sc.Steps[d.N].Song != "" {
			return game.Result{Events: []game.Event{{Kind: game.EvSong, Sound: sc.Steps[d.N].Song}}}
		}
		return game.Result{}
	case game.TitleCredits:
		r.title.n, r.title.credits, r.title.page = 0, true, d.Page
		r.seq++
		return game.Result{}
	case game.TitleDone:
		r.title.n, r.title.credits, r.title.page = 0, false, 0
		return r.do(play.Continue{})
	case game.SequenceDone:
		return r.do(onward(commonOf(ps).Allowed))
	case game.Speak: // gobox voices the page itself
		var res game.Result
		for _, k := range d.Keys {
			res.Events = append(res.Events, game.Event{Kind: game.EvSound, Sound: k})
		}
		return res
	case game.StopRest:
		r.rest = 0
		if slices.ContainsFunc(commonOf(ps).Allowed.Plain, func(a play.Action) bool { _, ok := a.(play.StopRest); return ok }) {
			return r.do(play.StopRest{}) // the game lists it: it hears of the stop
		}
		return game.Result{}
	case game.Rest:
		if _, ok := restCap(commonOf(ps).Allowed); ok { // gobox paces the rest in chunks, Step by Step
			r.rest = time.Duration(d.Minutes) * time.Minute
			if c, ok := ps.(play.Camping); ok && r.rest == 0 {
				r.rest = c.Study
			}
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
	case game.ShowMember: // select the member and open their sheet, as the old path's View
		n := who(d.Member, ps)
		res := r.do(play.SelectMember{Member: n})
		if res.Err == nil {
			r.sheet = n
		}
		return res
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
	case game.Attack:
		al := commonOf(ps).Allowed
		if !slices.Contains(al.Plain, play.Action(play.Attack{Target: d.Target})) {
			for _, x := range al.Aimed {
				if _, ok := x.Action.(play.Attack); ok { // no foe to hit by id: gobox asks for the square
					r.aim = &aiming{Aimed: x}
					return game.Result{}
				}
			}
			if m := r.attackAim(al.Plain); m != nil { // Attack before a target (the bar's, F): gobox aims at the foes the game offers
				r.aim, r.cursor = m, m.Squares[0] // the cursor on the first foe: Enter takes it, Tab cycles
				return game.Result{}
			}
		}
	case game.Save:
		if commonOf(ps).NoSave {
			return game.Result{Err: errors.New("cannot save here")}
		}
		return game.Result{Err: r.save(d.Name)}
	case game.Load:
		if commonOf(ps).NoLoad {
			return game.Result{Err: errors.New("cannot load here")}
		}
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
	case game.Ready: // on a readied item it unreadies it, as the originals' Ready toggles
		if pt := commonOf(ps).Party; d.Member >= 0 && d.Member < len(pt) && d.Item >= 0 && d.Item < len(pt[d.Member].Items) && pt[d.Member].Items[d.Item].Readied {
			return play.Unready{Member: d.Member, Item: d.Item}
		}
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
		return play.TryDoor{Way: doorWays[d.Way]}
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

// doorWays are the front's door ways (front.DoorExit ..) as play's.
var doorWays = map[int]play.DoorWay{0: play.DoorExit, 1: play.DoorBash, 2: play.DoorPick, 3: play.DoorKnock}

// facing is the step from a to b as a facing, else f (a step off the map).

// answer answers gobox's own ask, or the scene's.
func (r *rules) answer(ps play.State, d game.Answer) game.Result {
	if r.speedUp { // Faster, Slower (one step, the menu stays), Exit
		switch d.N {
		case 0:
			r.setSpeed(r.speed - 1)
		case 1:
			r.setSpeed(r.speed + 1)
		default:
			r.speedUp = false
		}
		r.seq++
		return game.Result{}
	}
	if r.aim != nil && r.ask == nil { // the square ask: the square answers the aim
		return r.aimAt(d.At)
	}
	if r.ask != nil {
		moves := r.moves
		r.ask, r.moves = nil, nil
		if d.N < 0 || d.N >= len(moves) {
			return game.Result{Err: errors.New("no such choice")}
		}
		return r.aimOrDo(commonOf(ps).Allowed, moves[d.N])
	}
	if q := commonOf(ps).Ask; q != nil {
		if d.N >= 0 && d.N < len(q.Do) && q.Do[d.N] != nil { // the rules' front-state reply (Gear: the items screen)
			return r.Apply(q.Do[d.N])
		}
		switch {
		case q.Continue:
			return r.do(play.Continue{})
		case q.Heading:
			return r.do(play.Heading{Dir: d.N})
		case q.YesNo: // the front's yes (1) or no (0): the reply so labelled, whatever its index; unlabelled: the play convention (no, yes)
			want := map[bool]string{true: "yes", false: "no"}[d.N == 1]
			if i := slices.IndexFunc(q.Replies, func(l string) bool { return strings.EqualFold(strings.TrimSpace(l), want) }); i >= 0 {
				return r.do(play.Reply{N: i})
			}
		}
		return r.do(play.Reply{N: d.N, Text: d.Text})
	}
	if a, ok := ps.(play.Assembling); ok && a.Draft != nil && asks(a.Draft) {
		return r.do(play.SetChoice{Field: a.Draft.Step, N: d.N, Text: d.Text})
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
		return r.do(onward(sc.Allowed))
	case sc.Typed:
		return r.do(play.Reply{N: d.N, Text: d.Text})
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
	if m == nil { // free cursor (looking): gobox's own, the game is not told
		r.cursor = at
		return game.Result{}
	}
	if !slices.Contains(m.Squares, at) { // not offered: the aim stays, no word (the original's Center)
		return game.Result{}
	}
	if m.byUnit != nil {
		r.aim = nil
		return r.do(play.Attack{Target: m.byUnit[at]})
	}
	m.ats = append(m.ats, at)
	if len(m.ats) < max(m.Targets, 1) {
		return game.Result{}
	}
	r.aim = nil
	return r.do(withAts(m.Action, m.ats))
}

// attackAim is the aim over the units plain's Attacks name (nil: none).
func (r *rules) attackAim(plain []play.Action) *aiming {
	var c *game.Combat
	if s := r.State(); s.Combat != nil {
		c = s.Combat
	}
	if c == nil {
		return nil
	}
	m := &aiming{Aimed: play.Aimed{Action: play.Attack{}, Targets: 1}, byUnit: map[image.Point]int{}}
	for _, a := range plain {
		if at, ok := a.(play.Attack); ok {
			for _, u := range c.Units {
				if u.ID == at.Target {
					for _, sq := range append([]image.Point{u.At}, u.Squares...) {
						m.byUnit[sq] = u.ID
						m.Squares = append(m.Squares, sq)
					}
				}
			}
		}
	}
	if len(m.Squares) == 0 {
		return nil
	}
	return m
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
	case play.Attack:
		if a.At != nil {
			return []image.Point{*a.At}
		}
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
	case play.Attack:
		x.At = nil
		if len(at) > 0 {
			x.At = &at[0]
		}
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
// traceOn (GOBOX_TRACE=1) logs every play action sent to the game.
var traceOn = os.Getenv("GOBOX_TRACE") != ""

func (r *rules) keep(a play.Action) game.Result {
	out, err := r.g.Do(a)
	if traceOn {
		fmt.Fprintf(os.Stderr, "trace %s play.Do %T%+v err=%v events=%d\n", time.Now().Format("15:04:05.000"), a, a, err, len(out.Events))
	}
	if err != nil { // no toast: the old path's refusals were silent unless the rules said (Refused)
		return game.Result{Err: err}
	}
	r.ask, r.moves, r.aim = nil, nil, nil
	r.seq++
	_, fight := r.g.State().(play.Fighting)
	if !fight {
		r.log = nil
	}
	for _, e := range out.Events {
		switch e := e.(type) {
		case play.Refused:
			r.note = game.Note{Text: e.Why, At: time.Now()}
		case play.Narrated:
			if fight {
				r.log = append(r.log, e.Text)
			}
		case play.Said:
			if fight {
				r.log = append(r.log, e.Who+": "+e.Text)
			}
		}
	}
	return game.Result{Events: eventsAt(out.Events, r.speed*speedStepMs)}
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
	steps := []play.Action{play.RunEvent{Before: true, Side: side}, play.Arrive{From: from, To: to, Facing: side}, play.RunEvent{}}
	if x.Where.Facing != side { // the party turns the way it walks first, as a player would: games allow Arrive facing the way it goes
		steps = append([]play.Action{play.Face{Facing: side}}, steps...)
	}
	for _, a := range steps {
		st := r.keep(a)
		res.Events, res.Err = append(res.Events, st.Events...), st.Err
		_, arrived := a.(play.Arrive) // what the step itself says, the square's event still runs, as a manual step's does
		before := false
		if e, ok := a.(play.RunEvent); ok && e.Before {
			before = true // the script ahead said something: the step it holds is still made, never left waiting on an Arrive
		}
		if _, still := r.g.State().(play.Exploring); st.Err != nil || !still || !arrived && !before && len(st.Events) > 0 && !onlyMoved(st.Events) {
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

// The pacing gobox gives a fight's moments: a blow or a cast holds the
// game speed's delay (speed × speedStepMs, the originals' pauseForMessage,
// SotSB R540:0BF7); a ship's blink is the hardware's ten 35 ms half-steps.
const speedStepMs, defaultSpeed, shipMs = 100, 4, 350

// events are the outcome's facts as the old events, at the default speed.
func events(es []play.Event) []game.Event { return eventsAt(es, defaultSpeed*speedStepMs) }

// eventsAt are the outcome's facts as the old events, a blow or a cast held ms.
func eventsAt(es []play.Event, ms int) (out []game.Event) {
	strikeMs, castMs := ms, ms
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
				game.Event{Kind: game.EvDamage, Unit: e.To, Target: e.To, N: e.Damage}) // Unit: who takes it (its number floats there)
		case play.Shot:
			ev := game.Event{Kind: game.EvMissile, Unit: e.From, To: e.To, Pose: 1, Ms: strikeMs, StepMs: e.StepMs}
			if e.Art.File != "" || e.Art.Pixels != nil || e.Art.Name != "" { // a shot with no art flies nothing
				ev.Sprite = []game.Picture{e.Art}
			}
			out = append(out, ev)
		case play.CastSpell:
			ev := game.Event{Kind: game.EvSpell, Unit: e.From, N: e.Spell, Pose: 1, Ms: castMs, StepMs: e.StepMs}
			if len(e.Ats) > 0 {
				ev.To = e.Ats[0]
			}
			out = append(out, ev)
		case play.Struck:
			out = append(out, game.Event{Kind: game.EvDamage, Unit: e.Unit, Target: e.Unit, N: e.Damage})
		case play.Healed:
			out = append(out, game.Event{Kind: game.EvDamage, Unit: e.Unit, Target: e.Unit, N: -e.HP})
		case play.Fell:
			out = append(out, game.Event{Kind: game.EvDown, Unit: e.Unit, Target: e.Unit, Picture: e.Body})
		case play.Fled: // a unit that flees or surrenders leaves the field as one that falls does
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
		// TimePassed, Gained, Lost, JournalNote: the state shows them.
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
	if gs, ok := r.g.(play.SlotSaver); ok { // the game's slots are the store
		sl := gs.Slots()
		return gs.SaveSlot(cmp.Or(slices.IndexFunc(sl, func(s game.SaveSlot) bool { return s.Name == name })+1, len(sl)+1) - 1)
	}
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
	if _, ok := r.g.(play.SlotSaver); ok {
		return errors.New("the game's slots are not removed")
	}
	p, err := r.path(name)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

func (r *rules) load(slot int) game.Result {
	own := r.ownSaves()
	var err error
	switch gs, ok := r.g.(play.Slots); {
	case slot >= 0 && slot < len(own):
		var p string
		var b []byte
		if p, err = r.path(own[slot].Name); err == nil {
			if b, err = os.ReadFile(p); err == nil {
				err = r.g.Load(b)
			}
		}
	case ok && slot >= len(own) && slot < len(own)+len(gs.Slots()):
		err = gs.LoadSlot(slot - len(own))
	default:
		err = errors.New("no such slot")
	}
	if err != nil {
		return game.Result{Err: err}
	}
	r.ask, r.moves, r.aim, r.items, r.read = nil, nil, nil, game.ItemsNone, nil
	return game.Result{Events: []game.Event{{Kind: game.EvReset}}}
}

// saves are gobox's slot files, newest first, then the game's own slots
// (play.Slots).
func (r *rules) saves() []game.SaveSlot {
	sl := r.ownSaves()
	if gs, ok := r.g.(play.Slots); ok {
		sl = append(sl, gs.Slots()...)
	}
	return sl
}

func (r *rules) ownSaves() []game.SaveSlot {
	if _, ok := r.g.(play.SlotSaver); ok {
		return nil // one store: the game's
	}
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

// exitLabel are the list entries that leave or page a pick; they keep a key under NoHotkeys.
var exitLabel = map[string]bool{"done": true, "exit": true, "leave": true, "prev": true, "next": true}

// onward is how a page is left: Continue, or Leave where the state allows
// that and not Continue (a handout).
func onward(al play.Allowed) play.Action {
	if !slices.Contains(al.Plain, play.Action(play.Continue{})) && slices.Contains(al.Plain, play.Action(play.Leave{})) {
		return play.Leave{}
	}
	return play.Continue{}
}

// Art is the game's art read live: what the game's own Art gives (its art
// changes as the party moves), else Data.Art.
func (r *rules) Art() map[string]game.Picture {
	if a, ok := r.g.(interface {
		Art() map[string]game.Picture
	}); ok {
		return a.Art()
	}
	return r.data.Art
}

// setSpeed sets the one game speed, 0 fastest .. 9 slowest: gobox's
// replay delay, and the rules' own where they take it.
func (r *rules) setSpeed(n int) {
	r.speed = min(max(n, 0), 9)
	if g, ok := r.g.(interface{ SetSpeed(int) }); ok {
		g.SetSpeed(r.speed)
	}
}

// Speed is the game speed, 0 fastest .. 9 slowest (classic holds a fight's
// messages by it).
func (r *rules) Speed() int { return r.speed }
