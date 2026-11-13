// rules.go: the front state over a game.Game: what gobox decides itself (the
// asks, the aiming, the inventory up, the paging and the pacing).

package ui

import (
	"cmp"
	"errors"
	"fmt"
	"hash/fnv"
	"image"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gobox/game"
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

type frontState struct {
	g    game.Game
	dir  string
	now  func() time.Time
	held heldStore              // the slot list read while the game waits (saves); run drops it
	data game.Data              // the series' data (verb keys, texts)
	ran  uint64                 // the steps done: each drops the held saves (Game.run)
	der  memo[uint64, *derived] // what is made from the State, at its Version (derived)
	UI
}

// squareAim is an Aimed action whose squares the player is picking.
type squareAim struct {
	game.Aimed
	ats    []image.Point
	byUnit map[image.Point]int // an Attack aimed at a unit the game offers by id: its squares' ids
}

// newRules is the front state over g; d is the game's Data.
func newRules(g game.Game, d game.Data) *frontState {
	return &frontState{g: g, dir: saveDir(d), data: d, now: time.Now, UI: UI{grid: image.Pt(40, 25), sheet: -1, speed: defaultSpeed}}
}

// who is n, or the selected member when n is -1.
func who(n int, ps game.State) int {
	if n < 0 {
		return game.CommonOf(ps).Selected
	}
	return n
}

// restCap is the longest rest al allows at once (0: any), false none.
func restCap(al game.Allowed) (time.Duration, bool) {
	for _, a := range al.Plain {
		if t, ok := a.(game.AdvanceTimeBy); ok {
			return t.D, true
		}
	}
	return 0, false
}

// Map is the game's walkable map while the party explores and no question
// is up  gobox walks on it.
func (r *frontState) Map() game.Map { return r.mapOf(r.g.State()) }

// mapOf is Map in state ps (a caller that has the State already: State is
// the rules' whole state, built anew on each call).
func (r *frontState) mapOf(ps game.State) game.Map {
	if t, ok := ps.(game.Exploring); ok && r.ask == nil && t.Map.Walk != nil {
		return t.Map.Walk
	}
	return nil
}

// Step goes on with what gobox paces: a chunk of camp's rest or a square
// of the auto-walk. The game never acts alone; it is sent the next action.
func (r *frontState) Step() game.Result {
	switch {
	case r.rest > 0:
		d := min(r.rest, restChunk)
		if m, _ := restCap(game.CommonOf(r.g.State()).Allowed); m > 0 {
			d = min(d, m)
		}
		r.rest -= d
		res := r.keep(game.AdvanceTimeBy{D: d})
		if res.Err != nil || len(res.Events) > 0 { // something happened meanwhile: the rest stops
			r.rest = 0
		}
		return res
	case r.walk != nil:
		return r.walkStep()
	}
	return game.Result{}
}

// restChunk is the game time one paced step of camp's rest takes.
const restChunk = 10 * time.Minute

// pending is the question up; where none is built and Continue is the
// only way on, a Continue ask, so the player is never stuck.
func (r *frontState) pending(ps game.State) *game.Ask {
	if _, fight := ps.(game.Fighting); fight && r.items != game.ItemsNone { // the items opened over a question (a View's Gear) are the top layer; closing them shows it again
		return nil
	}
	if r.speedUp {
		return &game.Ask{Kind: game.Choice, Prompt: "Speed " + strconv.Itoa(r.speed),
			Options: []game.Option{{Label: "Faster", Key: 'F'}, {Label: "Slower", Key: 'S'}, {Label: "Exit", Key: 'E'}}}
	}
	if a := r.asked(ps); a != nil {
		return a
	}
	if sc, ok := ps.(game.Scene); ok && (sc.Kind == game.SceneHandout || sc.Kind == game.SceneTitle || sc.Kind == game.SceneSequence) {
		return nil
	}
	al := game.CommonOf(ps).Allowed
	if f, ok := ps.(game.Fighting); ok && r.loneAim(f) {
		return &game.Ask{Kind: game.Square, Prompt: f.Status}
	}
	if len(al.Plain) == 1 && len(al.Aimed) == 0 && al.Plain[0] == (game.ContinueAction{}) {
		return &game.Ask{Kind: game.Continue}
	}
	return nil
}

func (r *frontState) asked(ps game.State) *game.Ask {
	if r.ask != nil {
		return r.ask
	}
	c := game.CommonOf(ps)
	if q := c.Ask; q != nil {
		a := r.question(q, q.List)
		a.View = slices.ContainsFunc(c.Allowed.Plain, func(x game.Action) bool { _, ok := x.(game.ViewMember); return ok }) // a pick the rules let the player look over first
		return a
	}
	if a, ok := ps.(game.Assembling); ok {
		if d := a.Draft; d != nil {
			return r.draftAsk(d)
		}
		return r.partyMenu(a)
	}
	if f, ok := ps.(game.ShipFight); ok && f.Menu != nil {
		return r.question(f.Menu, false)
	}
	sc, ok := ps.(game.Scene)
	switch {
	case !ok || sc.Kind == game.SceneHandout || sc.Kind == game.SceneTitle || sc.Kind == game.SceneSequence:
		return nil // the pageCut, the title and the sequence answer by their own decisions
	case sc.Kind == game.SceneHeading:
		p := ""
		if len(sc.Text) > 0 {
			p = sc.Text[len(sc.Text)-1]
		}
		return &game.Ask{Kind: game.Direction, Prompt: p, Options: []game.Option{{Label: "Done", Key: 'D', Do: game.Leave{}}}}
	}
	return &game.Ask{Kind: game.Continue}
}

// question is the rules' question as the old Ask.
func (r *frontState) question(q *game.Question, list bool) *game.Ask {
	a := &game.Ask{Kind: game.Choice, Prompt: q.Text, Default: q.Default, Max: q.Max, NoEsc: q.NoEsc, NoPick: q.NoPick}
	switch {
	case q.Continue:
		a.Kind, a.Ms = game.Continue, paced(q.Ms, r.speed*speedStepMs)
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
	a.Lines = q.Rows
	if a.Kind != game.YesNo { // a yes/no is the front's confirm: Options on it would be extra moves (game.Ask), as the old path had none
		a.Options = r.options(q.Replies, list)
		for i := range a.Options {
			if i < len(q.Keys) && q.Keys[i] != 0 {
				a.Options[i].Key = q.Keys[i] // the game's key (BR2's C for Character), not the label's
			}
		}
	}
	return a
}

// draftAsk is the creation step's question (nil: the sheet's fields).
func (r *frontState) draftAsk(d *game.Draft) *game.Ask {
	if d.Ask == nil {
		return nil
	}
	return r.question(d.Ask, true)
}

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
func (r *frontState) partyMenu(a game.Assembling) *game.Ask {
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
			on = slices.Contains(a.Allowed.Plain, game.Action(game.OpenPicker{Kind: game.PickForget}))
		case "View Character":
			on = len(a.Party) > 0
		case "Save Current Game":
			on = len(a.Party) > 0 && !a.NoSave
		case "Load Saved Game":
			on = !a.NoLoad // gobox's slot screen, empty or not
		}
		if on {
			q.Options = append(q.Options, game.Option{Label: cmp.Or(r.data.Texts[f.Label], f.Label), Key: f.Key, Func: f.Label})
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
// (list) has none, as the Gold Box lists are picked by arrows, but its way
// out ("Done", "Exit", "Leave") keeps one.
func (r *frontState) options(labels []string, list bool) []game.Option {
	used := map[rune]bool{}
	var o []game.Option
	for _, t := range labels {
		k := rune(0)
		if !list || exitLabel[strings.ToLower(t)] {
			k = key(t, used)
		}
		o = append(o, game.Option{Label: t, Key: k})
	}
	return o
}

// State is the typed state as the old State.
// askMode is ModeAsk where a question is up over a scene (a menu, a list, a
// name) or over the creation sheet (Common.Ask): the old front shows a
// question only in ModeAsk, so in ModeStory or ModeCreate it stayed hidden.
func askMode(m game.Mode, a *game.Ask, ps game.State) game.Mode {
	if a == nil || a.Kind == game.Continue && a.Ms == 0 {
		return m
	}
	if a.Kind == game.Continue { // a timed pause: the presenter's to pace (AskPause)
		return game.ModeAsk
	}
	switch {
	case m == game.ModeStory:
		if sc, ok := ps.(game.Scene); ok && sc.Kind == game.SceneScreen && sc.Screen != nil {
			return m // a typed screen answers by its own keys
		}
		return game.ModeAsk
	case m == game.ModeCreate && game.CommonOf(ps).Ask != nil:
		return game.ModeAsk
	}
	return m
}

// wallsKey names w's look: a new one has the 3D view built again.
func wallsKey(w game.MapWalls) string {
	h := fnv.New64a()
	fmt.Fprint(h, w.Side, w.Fronts, w.Pool, w.Floor, w.Sky, w.FloorInk, w.CeilInk, w.Backdrop)
	return strconv.FormatUint(h.Sum64(), 36)
}

func isHandout(ps game.State) bool {
	sc, ok := ps.(game.Scene)
	return ok && sc.Kind == game.SceneHandout
}

// oldNames are the old Allowed keys of the renamed actions.
var oldNames = map[string]string{"Give": "Trade", "Take": "TakeLoot", "AdvanceTimeBy": "Rest",
	"MakeCamp": "Camp", "BreakCamp": "LeaveCamp", "SetChoice": "SetIcon", "KeepCharacter": "KeepDraft",
	"AddCharacter": "AddChar", "MoveTo": "Move", "Shoot": "Ranged", "DoManeuver": "Maneuver", "Delay": "Wait"}

func name(a game.Action) string { return reflect.TypeOf(a).Name() }

// allowedNames is al by action name, new and old, plus gobox's own Select.
func allowedNames(al game.Allowed) map[string]bool {
	m := map[string]bool{"Select": true}
	add := func(a game.Action) {
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
func card(m game.Member, al game.Allowed, n int) []game.Entry {
	es := []game.Entry{{Name: "View", Verb: "view"}}
	has := map[string]bool{"items": len(m.Items) > 0}
	for _, a := range append(slices.Clone(al.Plain), aimedActions(al)...) {
		switch t := a.(type) {
		case game.Ready:
			has["items"] = has["items"] || t.Member == n
		case game.Drop:
			has["items"] = has["items"] || t.Member == n
		case game.UseItem:
			has["items"] = has["items"] || t.Member == n
		case game.Cast:
			has["cast"] = has["cast"] || t.Member == n
		case game.Memorize:
			has["memorize"] = has["memorize"] || t.Member == n
		case game.Plan:
			has["memorize"] = has["memorize"] || t.Member == n
		case game.Scribe:
			has["scribe"] = has["scribe"] || t.Member == n
		case game.LayOnHands:
			has["heal"] = has["heal"] || t.Member == n
		case game.EditIcon:
			has["icon"] = has["icon"] || t.Member == n
		case game.DropMember:
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

func aimedActions(al game.Allowed) []game.Action {
	var out []game.Action
	for _, a := range al.Aimed {
		out = append(out, a.Action)
	}
	return out
}

func canCast(al game.Allowed, n int) bool {
	for _, a := range al.Plain {
		if c, ok := a.(game.Cast); ok && c.Member == n {
			return true
		}
	}
	for _, a := range al.Aimed {
		if c, ok := a.Action.(game.Cast); ok && c.Member == n {
			return true
		}
	}
	return false
}

// run is a line of the rules' text: its colour unset, so the front gives
// it gobox's body text colour, in the text glyphs (Upcase: fonts.Glyph
// keeps a font's own lower case, else capitals). Upcase is not the case: a
// false one (once set for Data.Lowercase) drew PoD's lower-case letters as
// the charset's graphics.
func (r *frontState) run(t string) []game.TextRun {
	return []game.TextRun{{Text: t, Upcase: true}}
}

func gameWalls(w game.MapWalls) *game.Walls {
	return &game.Walls{Key: wallsKey(w), Side: w.Side, Fronts: w.Fronts, Pool: w.Pool, Floor: w.Floor, Sky: w.Sky,
		FloorInk: w.FloorInk, CeilInk: w.CeilInk, Sun: w.Sun, Mountains: w.Mountains,
		SunAz: w.SunAz, SunEl: w.SunEl, Backdrop: w.Backdrop}
}

func (r *frontState) page(sc game.Scene) *TextPage {
	p := &TextPage{View: sc.Placed}
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
	return p
}

// turnVerbs are the old combat menu's decisions for the new actions.
// turnVerbs are the combat menu's verbs: the action, gobox's label and
// the old decision (a series relabels one in Data.Texts, keys it in
// Data.VerbKeys).
// verbWord is the Gold Box combat menus' word for a gobox verb label, so
// a VerbOrder in the original's words (Use, Delay, Done) places them.
var verbWord = map[string]string{"Items": "Use", "Wait": "Delay", "End": "Done"}

var turnVerbs = []struct {
	new, label string
	do         game.Action
}{
	{"MoveTo", "Move", game.MoveTo{}}, {"Attack", "Attack", game.Attack{}}, {"Shoot", "Ranged", game.Shoot{}},
	{"Guard", "Guard", game.Guard{}}, {"Cast", "Cast", game.Cast{}}, {"UseItem", "Items", game.OpenItems{Member: -1}},
	{"DoManeuver", "Maneuver", game.DoManeuver{}}, {"TurnUndead", "Turn", game.TurnUndead{}},
	{"Bandage", "Bandage", game.Bandage{}}, {"Delay", "Wait", game.Delay{}}, {"Quick", "Quick", game.Quick{}},
	{"Manual", "Manual", game.Manual{}}, {"QuickMagic", "Magic", game.QuickMagic{}},
	{"Flee", "Flee", game.Flee{}}, {"EndTurn", "End", game.EndTurn{}},
}

// activeMember is the party index of the unit whose turn it is, -1 none.
func activeMember(b game.Battlefield) int {
	for _, u := range b.Units {
		if u.ID == b.Turn {
			return u.Member
		}
	}
	return -1
}

// loneAim starts the pick of squares when an Aimed action is all the turn
// allows (a jetpack's flight, a throw): gobox asks for the square at once.
func (r *frontState) loneAim(f game.Fighting) bool {
	al := f.Allowed
	if len(al.Aimed) != 1 || slices.ContainsFunc(al.Plain, func(a game.Action) bool { return a != game.Action(game.Leave{}) }) || len(al.Moves) > 0 {
		return false // Leave beside it is Esc (BR2's T)
	}
	if r.aim == nil {
		r.aim = &squareAim{Aimed: al.Aimed[0]}
		if sq := al.Aimed[0].Squares; len(sq) > 0 {
			r.cursor = sq[0] // the cursor on the first target
		}
	}
	return true
}

func (r *frontState) combat(f game.Fighting) *game.Combat {
	r.loneAim(f)
	b, al := f.Field, f.Allowed
	c := &game.Combat{Size: b.Size, Squares: b.Squares, Tiles: b.Tiles, FloorTile: b.FloorTile,
		Round: b.Round, Dungeon: b.Dungeon, Reach: al.Moves, Effects: b.Clouds}
	var spells []game.SpellRef
	for _, u := range b.Units {
		g := u
		g.Active = u.ID == b.Turn
		c.Units = append(c.Units, g)
		if u.ID == b.Turn && u.Member >= 0 && u.Member < len(f.Party) {
			spells = f.Party[u.Member].Spells
		}
	}
	t := &game.Turn{Unit: b.Turn, Aim: r.aim != nil, Moves: b.Moves, Casting: b.Casting, Weapon: b.Weapon, Missile: b.Missile}
	c.Turn, c.Panel, c.Log = t, b.Turn, r.log
	if slices.ContainsFunc(f.Text, func(l string) bool { return strings.TrimSpace(l) != "" }) { // the rules' message pane is the fight's log; a blank one never hides the blows
		c.Log = f.Text
	}
	if r.aim != nil {
		c.Marks = r.aim.ats
		if cs, ok := r.aim.Action.(game.Cast); ok {
			t.Spell = spellRef(spells, cs.Spell).Name
		}
	}
	allowed := allowedNames(al)
	used := map[rune]bool{} // the series' keys first; gobox's own invented around them
	for _, k := range r.data.VerbKeys {
		used[k] = true
	}
	for _, v := range answerVerbs {
		if allowed[name(v.a)] {
			used[r.verbKey(v.label)] = true
		}
	}
	for _, o := range turnVerbs {
		if allowed[o.new] || o.new == "MoveTo" && len(al.Moves) > 0 {
			l := cmp.Or(r.data.Texts[o.label], o.label)
			k, ok := r.data.VerbKeys[o.label]
			if !ok {
				k = key(l, used)
			}
			t.Options = append(t.Options, game.Option{Label: l, Key: k, Do: o.do})
		}
	}
	if slices.Contains(al.Plain, game.Action(game.Quick{All: true})) { // the original's Quick for the whole party: its own entry, keyed by the series (VerbKeys, VerbChords) or clicked
		t.Options = append(t.Options, game.Option{Label: cmp.Or(r.data.Texts["Quick All"], "Quick All"), Key: r.data.VerbKeys["Quick All"], Do: game.Quick{All: true}})
	}
	for _, v := range answerVerbs { // the rules' own verbs, keyed as the series keys them
		if allowed[name(v.a)] {
			t.Options = append(t.Options, game.Option{Label: v.label, Key: r.verbKey(v.label), Do: v.a})
		}
	}
	if k, ok := r.data.VerbKeys["View"]; ok && activeMember(b) >= 0 && len(t.Options) > 0 && !allowed["ViewTarget"] { // the series' View (keyed in VerbKeys): the active member's sheet, front state (BR's View is its rules' ViewTarget)
		l, m := cmp.Or(r.data.Texts["View"], "View"), activeMember(b)
		t.Options = append(t.Options, game.Option{Label: l, Key: k, Do: game.ViewMember{Member: m}})
	}
	called := r.data.AimsShots && slices.ContainsFunc(t.Options, func(o game.Option) bool { _, shoot := o.Do.(game.Shoot); return shoot }) // the series' Data calls its Ranged verb Aim (SotSB), and it is offered
	if _, keyed := r.data.VerbKeys["Aim"]; len(t.Options) > 0 && !called && (keyed || slices.Contains(r.data.VerbOrder, "Aim")) {          // the series names Aim (keyed or placed): gobox's free cursor; no menu ("Continue battle?"): no Aim either
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
		case game.Attack:
			c.Targets = append(c.Targets, a.Target)
		case game.Cast:
			c.Spells = append(c.Spells, spellRef(spells, a.Spell))
		case game.DoManeuver:
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
		cur, best, kept := r.cursor, -1, len(r.aim.ats) > 0 || slices.Contains(r.aim.Squares, r.cursor)
		for _, u := range live {
			c.Targets = append(c.Targets, u.ID)
			kept = kept || covers(u, []image.Point{r.cursor})
			if d := u.At.Sub(from); best < 0 || d.X*d.X+d.Y*d.Y < best {
				best, cur = d.X*d.X+d.Y*d.Y, u.At
			}
		}
		if kept || best < 0 {
			cur = r.cursor
		}
		r.cursor = cur
	}
	for _, a := range al.Aimed {
		switch x := a.Action.(type) {
		case game.Cast:
			sp := spellRef(spells, x.Spell)
			sp.Targets, sp.Area = a.Targets, a.Area
			c.Spells, c.Aims = append(c.Spells, sp), append(c.Aims, sp)
		case game.DoManeuver:
			t.Maneuvers = append(t.Maneuvers, maneuver(named, x.ID, a.Area))
		}
	}
	return c
}

// verbKey is the series' hotkey for a combat verb (Data.VerbKeys), else k.
func (r *frontState) verbKey(label string) rune {
	if x, ok := r.data.VerbKeys[label]; ok {
		return x
	}
	return rune(label[0]) // the label's first letter
}

// maneuver is the rules' maneuver id, named as they name it.
func maneuver(named map[int]game.TurnManeuver, id, area int) game.TurnManeuver {
	if m, ok := named[id]; ok {
		return m
	}
	return game.TurnManeuver{ID: id, Name: fmt.Sprint("Maneuver ", id), Area: area}
}

// answerVerbs are turn verbs the rules answer by hotkey: the key is the
// game's VerbKeys entry, else the label's first letter.
var answerVerbs = []struct {
	label string
	a     game.Action
}{
	{"Target", game.Target{}}, {"View", game.ViewTarget{}}, {"Intimidate", game.Intimidate{}},
	{"Aid", game.Aid{}}, {"Sprint", game.Sprint{}}, {"Dodge", game.Dodge{}}, {"Jetpack", game.Jetpack{}},
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
func (r *frontState) apply(a game.Action) game.Result {
	ps := r.g.State()
	al := game.CommonOf(ps).Allowed
	switch d := a.(type) {
	case game.Reply:
		if f, ok := ps.(game.Fighting); ok && r.ask == nil && f.Ask == nil {
			for _, v := range answerVerbs {
				if d.N == int(r.verbKey(v.label)) && allowedNames(f.Allowed)[name(v.a)] {
					return r.do(v.a)
				}
			}
		}
		return r.answer(ps, d, image.Point{})
	case game.Cancel:
		return r.cancel()
	case game.Aim:
		return r.aimAt(d.At)
	case game.OpenItems:
		return r.openItems(d.Member)
	case game.SpeedMenu:
		r.speedUp = true
		r.seq++
		return game.Result{}
	case game.Arrive: // the step the rules allow from and to there, with their facing: gobox's guess of it (kept on a step back) is not the rules' say
		return r.do(allowedArrive(al, d.From, d.To, d.Facing))
	case game.ViewMember: // a shop's, temple's or pool's View: the member's sheet, front state; closing it goes back to the screen under it
		if slices.Contains(al.Plain, a) { // a pick's View the rules offer: theirs
			return r.do(d)
		}
		r.sheet = who(d.Member, ps)
		return game.Result{}
	case game.StopRest:
		r.rest = 0
		if slices.ContainsFunc(al.Plain, func(x game.Action) bool { _, ok := x.(game.StopRest); return ok }) {
			return r.do(d) // the game lists it: it hears of the stop
		}
		return game.Result{}
	case game.DoManeuver:
		if d.ID == 0 && len(d.Ats) == 0 { // the bar's: gobox asks which
			return r.maneuvers(al)
		}
	case game.Shoot:
		if len(d.Ats) == 0 {
			return r.aimOrDo(al, d)
		}
	case game.Cast:
		return r.aimOrDo(al, d)
	case game.UseItem:
		return r.aimOrDo(al, d)
	case game.Attack:
		if !slices.Contains(al.Plain, a) {
			for _, x := range al.Aimed {
				if _, ok := x.Action.(game.Attack); ok { // no foe to hit by id: gobox asks for the square
					r.aim = &squareAim{Aimed: x}
					return game.Result{}
				}
			}
			if m := r.attackAim(al.Plain); m != nil { // Attack before a target (the bar's, F): gobox aims at the foes the game offers
				r.aim, r.cursor = m, m.Squares[0] // the cursor on the first foe: Enter takes it, Tab cycles
				return game.Result{}
			}
		}
	case game.Ready: // on a readied item it unreadies it, as the originals' Ready toggles
		if pt := game.CommonOf(ps).Party; d.Member >= 0 && d.Member < len(pt) && slices.ContainsFunc(pt[d.Member].Items, func(it game.Item) bool { return it.Handle == d.Item && it.Readied }) {
			return r.do(game.Unready{Member: d.Member, Item: d.Item})
		}
	}
	return r.do(a)
}

// cancel is Esc: the rules' reply it names, else what gobox asked is dropped.
func (r *frontState) cancel() game.Result {
	ps := r.g.State()
	if r.speedUp {
		r.speedUp = false
		r.seq++
		return game.Result{}
	}
	if r.ask == nil && r.aim == nil && isHandout(ps) { // Esc on a handout leaves it, as reading to its end does
		return r.do(onward(game.CommonOf(ps).Allowed))
	}
	lone := r.aim != nil && r.ask == nil && slices.Contains(game.CommonOf(ps).Allowed.Plain, game.Action(game.Leave{}))
	if q := game.CommonOf(ps).Ask; q != nil && r.ask == nil && r.aim == nil { // Esc on the rules' menu: the reply it names (its Exit, as the originals')
		if i, ok := q.Esc.Reply(); ok && i < len(q.Replies) {
			return r.answer(ps, game.Reply{N: i}, image.Point{})
		}
	}
	bare := r.ask == nil && r.aim == nil
	r.ask, r.moves, r.aim = nil, nil, nil
	if _, fight := ps.(game.Fighting); fight && lone { // Esc on the aim the game asks: its Leave
		return r.do(game.Leave{})
	}
	if _, loot := ps.(game.Looting); loot && bare && slices.Contains(game.CommonOf(ps).Allowed.Plain, game.Action(game.Leave{})) { // Esc on the treasure list: its Leave
		return r.do(game.Leave{})
	}
	return game.Result{}
}

// openItems puts up the inventory, member's (-1: the selected or acting one).
func (r *frontState) openItems(member int) game.Result {
	r.items, r.itemsOf = game.ItemsOwn, member
	return game.Result{}
}

// closeItems takes the inventory down; without all, to the member's sheet.
func (r *frontState) closeItems(all bool) game.Result {
	ps := r.g.State()
	r.items = game.ItemsNone
	if _, fight := ps.(game.Fighting); !all && !fight {
		r.sheet = who(-1, ps)
	}
	return game.Result{}
}

func (r *frontState) openSheet(member int) game.Result {
	r.sheet = who(member, r.g.State())
	return game.Result{}
}

func (r *frontState) closeSheet() game.Result {
	r.sheet = -1
	return game.Result{}
}

// showMember selects the member and opens their sheet.
func (r *frontState) showMember(member int) game.Result {
	n := who(member, r.g.State())
	res := r.do(game.SelectMember{Member: n})
	if res.Err == nil {
		r.sheet = n
	}
	return res
}

// titlePicture steps the title over the scene's Steps.
func (r *frontState) titlePicture(n int) game.Result {
	r.title.n, r.title.credits = n, false
	r.seq++
	if sc, ok := r.g.State().(game.Scene); ok && sc.Steps[n].Song != "" {
		return game.Result{Events: []game.Report{{Kind: game.EvSong, Sound: sc.Steps[n].Song}}}
	}
	return game.Result{}
}

func (r *frontState) titleCredits(page int) game.Result {
	r.title.n, r.title.credits, r.title.page = 0, true, page
	r.seq++
	return game.Result{}
}

func (r *frontState) titleDone() game.Result {
	r.title.n, r.title.credits, r.title.page = 0, false, 0
	return r.do(game.ContinueAction{})
}

func (r *frontState) sequenceDone() game.Result {
	return r.do(onward(game.CommonOf(r.g.State()).Allowed))
}

// speak voices a page: gobox's own.
func (r *frontState) speak(keys []string) game.Result {
	var res game.Result
	for _, k := range keys {
		res.Events = append(res.Events, game.Report{Kind: game.EvSound, Sound: k})
	}
	return res
}

// restFor rests minutes (0: the study's time, else 8 hours): gobox paces it
// in chunks, Step by Step, where the game allows; else the game's own advance.
func (r *frontState) restFor(minutes int) game.Result {
	ps := r.g.State()
	if _, ok := restCap(game.CommonOf(ps).Allowed); ok {
		r.rest = time.Duration(minutes) * time.Minute
		if c, ok := ps.(game.Camping); ok && r.rest == 0 {
			r.rest = c.Study
		}
		if r.rest == 0 {
			r.rest = 8 * time.Hour
		}
		r.seq++
		return game.Result{}
	}
	m := time.Duration(minutes) * time.Minute
	if m == 0 {
		m = 8 * time.Hour // gobox's default rest
	}
	return r.do(game.AdvanceTimeBy{D: m})
}

func (r *frontState) stopWalk() game.Result {
	r.walk = nil
	r.seq++
	return game.Result{}
}

func (r *frontState) saveNamed(name string) game.Result {
	if game.CommonOf(r.g.State()).NoSave {
		return game.Result{Err: errors.New("cannot save here")}
	}
	return game.Result{Err: r.save(name)}
}

func (r *frontState) loadSlot(slot int) game.Result {
	if game.CommonOf(r.g.State()).NoLoad {
		return game.Result{Err: errors.New("cannot load here")}
	}
	return r.load(slot)
}

// setField is a creation sheet field's choice, by the field's place.
func (r *frontState) setField(field, option int) game.Result {
	if a, ok := r.g.State().(game.Assembling); ok && a.Draft != nil && field >= 0 && field < len(a.Draft.Fields) {
		return r.do(game.SetChoice{Field: a.Draft.Fields[field].Name, N: option})
	}
	return game.Result{Err: errors.New("no such field")}
}

// facing is the step from a to b as a facing, else f (a step off the map).

// answer answers gobox's own ask, or the scene's.
func (r *frontState) answer(ps game.State, d game.Reply, at image.Point) game.Result {
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
		return r.aimAt(at)
	}
	if r.ask != nil {
		moves := r.moves
		r.ask, r.moves = nil, nil
		if d.N < 0 || d.N >= len(moves) {
			return game.Result{Err: errors.New("no such choice")}
		}
		return r.aimOrDo(game.CommonOf(ps).Allowed, moves[d.N])
	}
	if q := game.CommonOf(ps).Ask; q != nil {
		if d.N >= 0 && d.N < len(q.Do) && q.Do[d.N] != nil { // the rules' front-state reply (Gear: the items screen)
			return r.apply(q.Do[d.N])
		}
		switch {
		case q.Continue:
			return r.do(game.ContinueAction{})
		case q.Heading:
			return r.do(game.Heading{Dir: d.N})
		case q.YesNo: // the front's yes (1) or no (0): the reply so labelled, whatever its index; unlabelled: the play convention (no, yes)
			want := map[bool]string{true: "yes", false: "no"}[d.N == 1]
			if i := slices.IndexFunc(q.Replies, func(l string) bool { return strings.EqualFold(strings.TrimSpace(l), want) }); i >= 0 {
				return r.do(game.Reply{N: i})
			}
		}
		return r.do(game.Reply{N: d.N, Text: d.Text})
	}
	if a, ok := ps.(game.Assembling); ok && a.Draft != nil && a.Draft.Ask != nil {
		return r.do(game.SetChoice{Field: a.Draft.Step, N: d.N, Text: d.Text})
	}
	if _, ok := ps.(game.ShipFight); ok {
		return r.do(game.Crew{N: d.N})
	}
	sc, ok := ps.(game.Scene)
	switch {
	case !ok:
		return game.Result{Err: &game.Refusal{}} // nothing is asked: still refused, but silently (an empty Refusal raises no toast)
	case sc.Kind == game.SceneHeading:
		return r.do(game.Heading{Dir: d.N})
	case sc.Kind == game.SceneHandout:
		return r.do(onward(sc.Allowed))
	}
	return r.do(game.ContinueAction{})
}

// maneuvers asks which of the turn's maneuvers.
func (r *frontState) maneuvers(al game.Allowed) game.Result {
	r.ask, r.moves = &game.Ask{Kind: game.Choice, Prompt: "Maneuver"}, nil
	add := func(a game.Action) {
		if m, ok := a.(game.DoManeuver); ok {
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
func (r *frontState) aimOrDo(al game.Allowed, a game.Action) game.Result {
	if len(ats(a)) == 0 {
		for _, x := range al.Aimed {
			if reflect.DeepEqual(withAts(x.Action, nil), withAts(a, nil)) {
				r.aim = &squareAim{Aimed: x}
				return game.Result{}
			}
		}
	}
	return r.do(a)
}

func (r *frontState) aimAt(at image.Point) game.Result {
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
		return r.do(game.Attack{Target: m.byUnit[at]})
	}
	m.ats = append(m.ats, at)
	if len(m.ats) < max(m.Targets, 1) {
		return game.Result{}
	}
	r.aim = nil
	return r.do(withAts(m.Action, m.ats))
}

// attackAim is the aim over the units plain's Attacks name (nil: none).
func (r *frontState) attackAim(plain []game.Action) *squareAim {
	c := r.combatOf(r.g.State())
	if c == nil {
		return nil
	}
	m := &squareAim{Aimed: game.Aimed{Action: game.Attack{}, Targets: 1}, byUnit: map[image.Point]int{}}
	for _, a := range plain {
		if at, ok := a.(game.Attack); ok {
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

func ats(a game.Action) []image.Point {
	switch a := a.(type) {
	case game.Cast:
		return a.Ats
	case game.UseItem:
		return a.Ats
	case game.Shoot:
		return a.Ats
	case game.DoManeuver:
		return a.Ats
	case game.Attack:
		if a.At != nil {
			return []image.Point{*a.At}
		}
	}
	return nil
}

func withAts(a game.Action, at []image.Point) game.Action {
	switch x := a.(type) {
	case game.Cast:
		x.Ats = at
		return x
	case game.UseItem:
		x.Ats = at
		return x
	case game.Shoot:
		x.Ats = at
		return x
	case game.DoManeuver:
		x.Ats = at
		return x
	case game.Attack:
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
func (r *frontState) do(a game.Action) game.Result {
	r.rest, r.walk = 0, nil
	return r.keep(a)
}

// keep sends a to the game: a refusal is Err, else the outcome as events.
func (r *frontState) keep(a game.Action) game.Result {
	out, err := r.g.Do(a)
	return r.outcome(a, out, err)
}

// outcome is what a's Do gave as the result the front plays.
func (r *frontState) outcome(a game.Action, out game.Outcome, err error) game.Result {
	if traceOn {
		fmt.Fprintf(os.Stderr, "trace %s game.Do %T%+v err=%v events=%d\n", time.Now().Format("15:04:05.000"), a, a, err, len(out.Events))
	}
	if err != nil { // no toast: the old path's refusals were silent unless the rules said (Refused)
		return game.Result{Err: err}
	}
	r.ask, r.moves, r.aim = nil, nil, nil
	r.seq++
	_, fight := r.g.State().(game.Fighting)
	if !fight {
		r.log = nil
	}
	for _, e := range out.Events {
		switch e := e.(type) {
		case game.Refused:
			r.note = game.Note{Text: e.Why, At: time.Now()}
		case game.Narrated:
			if fight {
				r.log = append(r.log, e.Text)
			}
		case game.Said:
			if fight {
				r.log = append(r.log, e.Who+": "+e.Text)
			}
		}
	}
	return game.Result{Events: eventsAt(out.Events, r.speed*speedStepMs, r.engaged)}
}

// walkSides are the four steps as sides (N E S W) with their offsets, in the
// order the walk's search tries them.
var walkSides = [4]struct {
	side int
	d    image.Point
}{{0, image.Pt(0, -1)}, {2, image.Pt(1, 0)}, {4, image.Pt(0, 1)}, {6, image.Pt(-1, 0)}}

// walkPace is the auto-walk's time per square.
const walkPace = 150 * time.Millisecond

// walking is the auto-walk under way: where to, in which area, and when the
// last square was walked (time.Time's zero: none yet, the first goes at once).
type walking struct {
	to   image.Point
	area string
	last time.Time
}

// walkBlocked is whether the walk's search counts e as a wall: a locked door
// does, the map shows every wall and so does the way.
func walkBlocked(e game.Edge) bool {
	return e.Kind == game.EdgeWall || e.Kind == game.EdgeDoor && (e.State == game.DoorLocked || e.State == game.DoorWizardLocked)
}

// walkSide is the side of from the shortest way to goal leaves by, false: none.
func walkSide(m game.Map, from, goal image.Point) (int, bool) {
	first := map[image.Point]int{from: -1} // each square reached: the side the way to it left from by
	for q := []image.Point{from}; len(q) > 0; q = q[1:] {
		if q[0] == goal {
			return first[goal], goal != from
		}
		for _, s := range walkSides {
			n := q[0].Add(s.d)
			if _, seen := first[n]; seen || !n.In(image.Rectangle{Max: m.Size()}) || walkBlocked(m.Edge(q[0], s.side)) {
				continue
			}
			first[n] = first[q[0]]
			if q[0] == from {
				first[n] = s.side
			}
			q = append(q, n)
		}
	}
	return 0, false
}

// walkOffset is side's step.
func walkOffset(side int) image.Point {
	for _, s := range walkSides {
		if s.side == side {
			return s.d
		}
	}
	return image.Point{}
}

// walkTo sets off for at: there already is done, no known way refuses.
func (r *frontState) walkTo(ps game.State, at image.Point) game.Result {
	x, ok := ps.(game.Exploring)
	if !ok || x.Map.Walk == nil || x.Ask != nil {
		return game.Result{Err: errors.New("not walking")}
	}
	from := image.Pt(x.Where.X, x.Where.Y)
	if at == from {
		return game.Result{}
	}
	if _, ok := walkSide(x.Map.Walk, from, at); !ok {
		return game.Result{Err: errors.New("no way there")}
	}
	r.walk = &walking{to: at, area: x.Where.Area}
	r.seq++
	return game.Result{}
}

// stands is whether the party still walks after a step's action: exploring,
// nothing asked, no door up, on square at in the walk's area, facing facing.
func (w *walking) stands(ps game.State, at image.Point, facing int) bool {
	x, ok := ps.(game.Exploring)
	return ok && x.Ask == nil && x.Door == nil && x.Map.Walk != nil && x.Where.Area == w.area && image.Pt(x.Where.X, x.Where.Y) == at && x.Where.Facing == facing
}

// walked is the map's squares walked on so far.
func walked(x game.Exploring) []image.Point {
	if x.Automap.Size != (image.Point{}) {
		return x.Automap.Walked
	}
	return x.Map.Automap.Walked
}

// walkHalts is whether what a step's scripts did stops the walk. Moving, time
// and sound do not; plain text does only on a square not visited before (a
// visited square's text is flavour; blank text only clears the window); any
// other fact does.
func walkHalts(es []game.Event, visited bool) bool {
	text := func(t string) bool { return !visited && strings.TrimSpace(t) != "" }
	for _, e := range es {
		switch e := e.(type) {
		case game.Moved, game.Turned, game.TimePassed, game.Sounded, game.MusicChange:
		case game.Narrated:
			if text(e.Text) {
				return true
			}
		case game.Said:
			if text(e.Text) {
				return true
			}
		default:
			return true
		}
	}
	return false
}

// newText is whether the rules' message lines changed to something not blank.
func newText(was, now []string) bool {
	return !slices.Equal(was, now) && slices.ContainsFunc(now, func(l string) bool { return strings.TrimSpace(l) != "" })
}

// hasPicture is whether c holds a picture up.
func hasPicture(c game.Common) bool {
	return c.Picture != nil || len(c.Layers) > 0 || c.Portrait != nil
}

// walkStep walks the next square when it is time: the turn, the script ahead
// (the step only if the party stands as it did), the step, the square's event.
// The walk stops at the square, where a step does not land as expected, where
// a question, door or picture comes up, and where walkHalts says.
func (r *frontState) walkStep() game.Result {
	w := r.walk
	x, ok := r.g.State().(game.Exploring)
	from := image.Pt(x.Where.X, x.Where.Y)
	if !ok || !w.stands(x, from, x.Where.Facing) || from == w.to {
		r.walk = nil
		return game.Result{}
	}
	if now := r.now(); now.Sub(w.last) < walkPace {
		return game.Result{}
	} else {
		w.last = now
	}
	side, ok := walkSide(x.Map.Walk, from, w.to)
	if !ok {
		r.walk = nil
		return game.Result{}
	}
	to, visited := from.Add(walkOffset(side)), slices.Contains(walked(x), from.Add(walkOffset(side)))
	var res game.Result
	var facts []game.Event
	act := func(a game.Action) bool {
		out, err := r.g.Do(a)
		st := r.outcome(a, out, err)
		res.Events, res.Err = append(res.Events, st.Events...), st.Err
		facts = append(facts, out.Events...)
		return err == nil
	}
	r.walk = nil // kept only by reaching the end whole
	if x.Where.Facing != side && (!act(game.Face{Facing: side}) || !w.stands(r.g.State(), from, side)) {
		return res
	}
	if x.Map.Walk.Departure(from, side) && (!act(game.RunEvent{Before: true, Side: side}) || !w.stands(r.g.State(), from, side)) {
		return res
	}
	if !act(allowedArrive(game.CommonOf(r.g.State()).Allowed, from, to, side)) {
		return res
	}
	ps := r.g.State()
	if n, ok := ps.(game.Exploring); ok && n.Ask == nil && n.Map.Walk != nil && n.Map.Walk.Event(image.Pt(n.Where.X, n.Where.Y)) { // the square's event, as a manual step's
		if !act(game.RunEvent{}) {
			return res
		}
		ps = r.g.State()
	}
	if !w.stands(ps, to, side) || walkHalts(facts, visited) || hasPicture(game.CommonOf(ps)) && !hasPicture(x.Common) || !visited && newText(x.Text, game.CommonOf(ps).Text) {
		return res
	}
	r.walk = w
	return res
}

// allowedArrive is the step as the rules allow it from and to there, with
// their facing, else with facing.
func allowedArrive(al game.Allowed, from, to image.Point, facing int) game.Action {
	for _, a := range al.Plain {
		if s, ok := a.(game.Arrive); ok && s.From == from && s.To == to {
			return s
		}
	}
	return game.Arrive{From: from, To: to, Facing: facing}
}

// The pacing gobox gives a fight's moments: a blow or a cast holds the
// game speed's delay (speed × speedStepMs, the originals' pauseForMessage,
// SotSB R540:0BF7); a ship's blink is the hardware's ten 35 ms half-steps.
const speedStepMs, defaultSpeed, shipMs = 100, 4, 350

// paced is an original's hold of ms (made at the default speed) at the game speed
// whose message delay is holdMs: the fastest speed keeps 1 ms; ms 0 stays 0.
func paced(ms, holdMs int) int {
	if ms <= 0 {
		return 0
	}
	return max(ms*holdMs/(defaultSpeed*speedStepMs), 1)
}

// fallHoldMs is how long the replay holds on the batch's last fall.
const fallHoldMs = 500

// events are the outcome's facts as the old events, at the default speed.
// engaged is f, a fight as it began, as the EvEngaged event: its board and party.
func (r *frontState) engaged(f game.Fighting) game.Report {
	return game.Report{Kind: game.EvEngaged, Combat: r.combat(f), Party: f.Party}
}

// eventsAt are the outcome's facts as the old events, a blow or a cast held
// ms; engaged gives an Engaged fight as the old Combat and party (nil: no EvEngaged).
func eventsAt(es []game.Event, ms int, engaged func(game.Fighting) game.Report) (out []game.Report) {
	strikeMs, castMs, lastFell := ms, ms, -1
	for _, e := range es {
		switch e := e.(type) {
		case game.Narrated:
			out = append(out, game.Report{Kind: game.EvMessage, Text: e.Text, Wait: true})
		case game.Said:
			out = append(out, game.Report{Kind: game.EvSaid, Role: e.Who, Text: e.Text, Wait: true})
		case game.Shown:
			out = append(out, game.Report{Kind: game.EvPicture, Picture: e.Picture})
		case game.Sounded:
			out = append(out, game.Report{Kind: game.EvSound, Sound: e.Sound})
		case game.MusicChange:
			out = append(out, game.Report{Kind: game.EvSong, Sound: e.Song})
		case game.Walked:
			for _, st := range e.Path {
				out = append(out, game.Report{Kind: game.EvMove, Unit: e.Unit, To: st.At})
			}
		case game.Turned:
			out = append(out, game.Report{Kind: game.EvMove, Unit: e.Unit, N: e.Facing, Turn: true})
		case game.Attacked:
			out = append(out, game.Report{Kind: game.EvAttack, Unit: e.From, Target: e.To, Pose: 1, Ms: strikeMs})
			if !e.Hit {
				out = append(out, game.Report{Kind: game.EvMiss, Unit: e.From, Target: e.To})
				break
			}
			out = append(out, game.Report{Kind: game.EvHit, Unit: e.From, Target: e.To},
				game.Report{Kind: game.EvDamage, Unit: e.To, Target: e.To, N: e.Damage}) // Unit: who takes it (its number floats there)
		case game.Shot:
			ev := game.Report{Kind: game.EvMissile, Unit: e.From, To: e.To, Pose: 1, Ms: strikeMs, StepMs: paced(e.StepMs, ms)}
			if e.Art.File != "" || e.Art.Pixels != nil || e.Art.Name != "" { // a shot with no art flies nothing
				ev.Sprite = []game.Picture{e.Art}
			}
			out = append(out, ev)
		case game.CastSpell:
			ev := game.Report{Kind: game.EvSpell, Unit: e.From, N: e.Spell, Pose: 1, Ms: castMs}
			if len(e.Ats) > 0 {
				ev.To = e.Ats[0]
			}
			out = append(out, ev)
		case game.Engaged:
			if engaged != nil {
				out = append(out, engaged(e.At))
			}
		case game.Struck:
			out = append(out, game.Report{Kind: game.EvDamage, Unit: e.Unit, Target: e.Unit, N: e.Damage})
		case game.Healed:
			out = append(out, game.Report{Kind: game.EvDamage, Unit: e.Unit, Target: e.Unit, N: -e.HP})
		case game.Fell:
			lastFell = len(out)
			out = append(out, game.Report{Kind: game.EvDown, Unit: e.Unit, Target: e.Unit, Picture: e.Body})
		case game.Fled: // a unit that flees or surrenders leaves the field as one that falls does
			out = append(out, game.Report{Kind: game.EvDown, Unit: e.Unit, Target: e.Unit})
		case game.Moved:
			out = append(out, game.Report{Kind: game.EvMove, Unit: -1, To: e.To, N: e.Facing})
		case game.Refused:
			out = append(out, game.Report{Kind: game.EvRefused, Text: e.Why})
		case game.Ended:
			out = append(out, game.Report{Kind: game.EvReset})
		case game.SaveOffered:
			out = append(out, game.Report{Kind: game.EvSaves})
		case game.ShipShown:
			f := e.Fight
			out = append(out, game.Report{Kind: game.EvPicture, Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: &f}, Ms: shipMs})
		}
		// TimePassed, Gained, Lost, JournalNote: the state shows them.
	}
	if lastFell >= 0 { // the batch's last fall holds, so the body the last blow leaves shows before the fight's page (bug class 55)
		out[lastFell].Ms = fallHoldMs
	}
	return out
}

// Slot files: one per name, Dir/<name>.sav, the game's Save bytes.

func (r *frontState) path(name string) (string, error) {
	if name == "" || name != filepath.Base(name) || strings.HasPrefix(name, ".") {
		return "", errors.New("bad slot name")
	}
	return filepath.Join(r.dir, name+".sav"), nil
}

func (r *frontState) save(name string) error {
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

func (r *frontState) remove(name string) error {
	p, err := r.path(name)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

func (r *frontState) load(slot int) game.Result {
	own := r.ownSaves()
	var err error
	switch gs, ok := r.g.(game.Slots); {
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
	return game.Result{Events: []game.Report{{Kind: game.EvReset}}}
}

// saves are gobox's slot files, newest first, then the game's own slots
// (game.Slots).
func (r *frontState) saves() []game.SaveSlot {
	return heldValue(&r.held, slotsKey{}, r.readSaves)
}

// slotsKey is saves' value held: the slot screen lists it every frame.
type slotsKey struct{}

// readSaves is saves read from the disk.
func (r *frontState) readSaves() []game.SaveSlot {
	sl := r.ownSaves()
	if gs, ok := r.g.(game.Slots); ok {
		sl = append(sl, gs.Slots()...)
	}
	return sl
}

func (r *frontState) ownSaves() []game.SaveSlot {
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

// exitLabel are the list entries that leave or page a pick; they keep a key in a list.
var exitLabel = map[string]bool{"done": true, "exit": true, "leave": true, "prev": true, "next": true}

// onward is how a page is left: Continue, or Leave where the state allows
// that and not Continue (a handout).
func onward(al game.Allowed) game.Action {
	if !slices.Contains(al.Plain, game.Action(game.ContinueAction{})) && slices.Contains(al.Plain, game.Action(game.Leave{})) {
		return game.Leave{}
	}
	return game.ContinueAction{}
}

// Art is the game's art read live: what the game's own Art gives (its art
// changes as the party moves), else Data.Art.
func (r *frontState) Art() map[string]game.Picture {
	if a, ok := r.g.(interface {
		Art() map[string]game.Picture
	}); ok {
		return a.Art()
	}
	return r.data.Art
}

// setSpeed sets the one game speed, 0 fastest .. 9 slowest: gobox's
// replay delay, and the rules' own where they take it.
func (r *frontState) setSpeed(n int) {
	r.speed = min(max(n, 0), 9)
	if g, ok := r.g.(interface{ SetSpeed(int) }); ok {
		g.SetSpeed(r.speed)
	}
}

// Speed is the game speed, 0 fastest .. 9 slowest (classic holds a fight's
// messages by it).
func (r *frontState) Speed() int { return r.speed }

// frontPower is the inventory's power act as the game's PowerKind.
var frontPower = map[int]game.PowerKind{int(ActSpells): game.PowerCast, int(ActMagic): game.PowerDisplay, int(ActHeal): game.PowerHeal, int(ActCure): game.PowerCure}
