// view_state.go: what the screens read of the state up, each from play's types.
package classic

import (
	"cmp"
	"image"
	"slices"

	"gobox/game"
	"gobox/game/play"
)

// partyOf is ps's party as the screens show it.
func partyOf(ps play.State) []game.Member {
	c := commonOf(ps)
	var out []game.Member
	for i, m := range c.Party {
		m.Chosen, m.CanCast = i == c.Selected, m.CanCast || canCast(c.Allowed, i)
		m.Card = card(m, c.Allowed, i)
		out = append(out, m)
	}
	return out
}

// selectedOf is the member the screens act on: in a fight with the items up,
// whose items they are.
func (r *rules) selectedOf(ps play.State) int {
	c := commonOf(ps)
	if f, ok := ps.(play.Fighting); ok {
		if m := cmp.Or(r.itemsOf+1, activeMember(f.Field)+1) - 1; r.items != game.ItemsNone && m >= 0 && m < len(c.Party) {
			return m
		}
	}
	return c.Selected
}

// allowedOf is what ps allows by action name, and NoRest where camp has no
// timed rest.
func allowedOf(ps play.State) map[string]bool {
	m := allowedNames(commonOf(ps).Allowed)
	if _, camp := ps.(play.Camping); camp && !m["Rest"] {
		m["NoRest"] = true
	}
	return m
}

// automapOf is the map of where the party is; camp keeps the one explored.
func (r *rules) automapOf(ps play.State) game.Automap {
	c := commonOf(ps)
	a := c.Automap
	if x, ok := ps.(play.Exploring); ok && a.Size == (image.Point{}) {
		a = x.Map.Automap
	}
	if _, ok := ps.(play.Exploring); ok {
		r.camMap = a
	} else if _, ok := ps.(play.Camping); ok && len(a.Known) == 0 && len(r.camMap.Known) > 0 {
		a = r.camMap
	}
	if a.Area == "" {
		a.Area = c.Area
	}
	return a
}

// storyOf is the speaker's portrait, head and body.
func storyOf(c play.Common) []game.Picture {
	if pt := c.Portrait; pt != nil {
		return []game.Picture{pt.Head, pt.Body}
	}
	return nil
}

// pageOf is the text page up.
func (r *rules) pageOf(ps play.State) *game.Page {
	c := commonOf(ps)
	p := &game.Page{Layers: c.Layers}
	if c.Picture != nil && len(c.Layers) == 0 { // the front draws a page's layers: the picture is the page's only one
		p.Layers = []game.Layer{{Picture: *c.Picture}}
	}
	for _, t := range c.Text {
		p.Pane = append(p.Pane, r.run(t))
	}
	if c.Status != "" {
		p.Prompt = r.run(c.Status)
	}
	if t, ok := ps.(play.Scene); ok {
		pg := r.page(t)
		if t.Kind == play.SceneTitle && !r.title.credits {
			n := r.title.n
			if n == 0 && t.Intro != nil && len(t.Intro.Pictures) > 0 { // before gobox's first step: that step's picture already
				n = t.Intro.Pictures[0]
			}
			pg.Layers = append(pg.Layers, t.Steps[n].Layers...)
		}
		p.Pane, p.View, p.Layers = append(p.Pane, pg.Pane...), pg.View, append(p.Layers, pg.Layers...)
	}
	return p
}

// flagsOf are the screen's plain status flags.
func (r *rules) flagsOf(ps play.State, party []game.Member) *game.Flags {
	c := commonOf(ps)
	_, camp := ps.(play.Camping)
	f := &game.Flags{Adventure: c.Adventure, Ordering: camp && allowedOf(ps)["Order"], Sweeping: c.Where.Searching,
		Inside: c.Where.Inside, CanCast: c.Selected >= 0 && c.Selected < len(party) && party[c.Selected].CanCast,
		IconDrop: c.IconEditor != nil, Day: c.Day, NoSave: c.NoSave, NoLoad: c.NoLoad}
	switch t := ps.(type) {
	case play.Exploring:
		f.Exploring, f.Day = c.Ask == nil, t.Day || c.Day // the game's say, never the hour guessed
	case play.Scene:
		f.Full = t.Full
		if t.Kind == play.SceneTitle {
			f.Credits, f.CreditPage = r.title.credits, r.title.page
		}
	}
	return f
}

// windowOf is what the view window shows.
func windowOf(ps play.State) game.Window {
	c := commonOf(ps)
	pictured := c.Picture != nil || len(c.Layers) > 0 || c.Portrait != nil
	switch t := ps.(type) {
	case play.Exploring:
		switch {
		case pictured:
			return game.WindowPicture
		case t.Map.Overland:
			return game.WindowMap
		}
		return game.WindowScene
	case play.Scene:
		switch {
		case t.Kind == play.SceneTitle, t.Kind == play.SceneSequence:
			return game.WindowScene
		case len(t.Pictures)+len(t.Layers) > 0:
			return game.WindowPicture
		}
		return game.WindowText
	}
	if pictured {
		return game.WindowPicture
	}
	var none game.Window
	return none
}

// sheetOf is what the selected member's sheet offers (nil: none up).
func (r *rules) sheetOf(c play.Common) *game.Sheet {
	if n := c.Selected; n >= 0 && n < len(c.Party) && (len(c.Party[n].SheetPowers) > 0 || len(c.Party[n].Reactions) > 0) {
		return &game.Sheet{Powers: c.Party[n].SheetPowers, Reactions: c.Party[n].Reactions}
	}
	if n := r.sheet; n >= 0 && n < len(c.Party) {
		return &game.Sheet{} // up though the member has no powers: the front shows a page only with one
	}
	return nil
}

// sheetPageOf is the sheet's page up: the rules', else gobox's from the member.
func (r *rules) sheetPageOf(ps play.State, party []game.Member) *game.SheetPage {
	c := commonOf(ps)
	var p *game.SheetPage
	if n := r.sheet; n >= 0 && n < len(c.Party) {
		if p = c.Party[n].Page; p == nil {
			p = game.MemberPage(party[n])
		}
	}
	if a, ok := ps.(play.Assembling); ok && a.Draft != nil && a.Draft.Member.Page != nil {
		p = a.Draft.Member.Page // the rules' figures for the sheet being made
	}
	return p
}

// creationOf is the draft up (nil: none).
func creationOf(ps play.State) *game.Creation {
	a, ok := ps.(play.Assembling)
	if !ok || a.Draft == nil {
		return nil
	}
	d := a.Draft
	draft := d.Member
	draft.Chosen = true
	cr := &game.Creation{Draft: draft, Fields: d.Fields, Roster: a.Roster,
		MinHP: d.MinHP, MaxHP: d.MaxHP, Minimums: d.Minimums, Step: d.Step, Stats: d.Stats, Derived: d.Derived}
	if d.Ask != nil {
		cr.Choices = d.Ask.Replies
	}
	return cr
}

// pacingOf is what goes on with time.
func (r *rules) pacingOf(ps play.State) *game.Pacing {
	p := &game.Pacing{Seq: r.seq}
	switch {
	case r.rest > 0:
		p.Pace, p.Resting = game.PaceRest, true
	case r.walk != nil:
		p.Pace = game.PaceWalk
	default:
		if t, ok := ps.(play.Scene); ok && slices.ContainsFunc(t.Layers, func(l game.Layer) bool { return l.FrameMs > 0 }) {
			p.Pace = game.PaceAnim
		}
	}
	return p
}

// wallsOf is the 3D view up: camp and a scene keep the one last explored; a
// fight or the party menu drops it.
func (r *rules) wallsOf(ps play.State) *game.Walls {
	switch t := ps.(type) {
	case play.Exploring:
		r.walls = nil
		if w := t.Map.Walls; w.Side != nil { // no squares: no 3D view (the overland)
			r.walls = walls(w)
		}
	case play.Scene:
		if t.Walls != nil && t.Walls.Side != nil {
			return walls(*t.Walls)
		}
	case play.Camping:
	default:
		r.walls = nil
	}
	return r.walls
}

// readOf is the handout up, kept so the pointer holds while it does.
func (r *rules) readOf(ps play.State) *game.Read {
	t, ok := ps.(play.Scene)
	if !ok || t.Kind != play.SceneHandout {
		r.read = nil
		return nil
	}
	if r.read == nil || r.read.Title != t.Handout || !slices.Equal(r.read.Lines, t.Lines) {
		r.read = &game.Read{Title: t.Handout, Lines: t.Lines}
	}
	return r.read
}

// modeOf is the screen ps puts up, a question over it counted.
func (r *rules) modeOf(ps play.State) game.Mode {
	var m game.Mode
	c := commonOf(ps)
	switch t := ps.(type) {
	case play.Assembling:
		m = game.ModeParty
		if t.Draft != nil {
			m = game.ModeCreate
		}
	case play.Exploring:
		m = game.ModeExplore
		if c.Ask != nil && !(c.Ask.Continue && len(c.Ask.Replies) == 0) || t.Door != nil { // a question (or the door menu) over the view
			m = game.ModeAsk
		}
	case play.Camping:
		m = game.ModeCamp
	case play.Fighting:
		m = game.ModeCombat
	case play.Scene:
		m = game.ModeStory
		if t.Kind == play.SceneTitle {
			m = game.ModeTitle
		}
	case play.Trading:
		switch {
		case t.Kind == play.TradeVault:
			m = game.ModeVault
		case t.Kind == play.TradeTraining, t.Training != nil:
			m = game.ModeTraining
		case t.Kind == play.TradeTemple, len(t.Services) > 0:
			m = game.ModeTemple
		default:
			m = game.ModeShop
		}
	case play.Looting:
		m = game.ModeLoot
	case play.Over:
		m = game.ModeTitle
	case play.ShipFight:
		m = game.ModeStory
	}
	return askMode(m, r.pending(ps), ps)
}

// itemsOf is whose goods the inventory screen shows.
func (r *rules) itemsUp(ps play.State) game.ItemsUp {
	if r.items == game.ItemsNone {
		return r.items
	}
	switch ps.(type) {
	case play.Trading:
		return game.ItemsShop
	case play.Looting:
		return game.ItemsLoot
	}
	return r.items
}

// shopOf is the shop up (nil: none); a vault or hall is no shop.
func shopOf(ps play.State) *game.Shop {
	t, ok := ps.(play.Trading)
	if !ok || t.Kind == play.TradeVault || t.Kind == play.TradeTraining || t.Training != nil || t.Kind == play.TradeTemple || len(t.Services) > 0 {
		return nil
	}
	return &game.Shop{Name: t.Name, Goods: t.Goods, Buys: t.Buys, Offers: t.Offers, Pooled: t.Pooled,
		Reloads: allowedOf(ps)["Reload"], IdentifyPrice: t.Identify, Unknown: t.Unknown, Identified: t.Known}
}

// templeOf is the temple up: its services, offers (nil services: none).
func templeOf(ps play.State) ([]game.Service, []game.Offer) {
	t, ok := ps.(play.Trading)
	if !ok || t.Kind == play.TradeVault || t.Kind == play.TradeTraining || t.Training != nil || t.Kind != play.TradeTemple && len(t.Services) == 0 {
		return nil, nil
	}
	return t.Services, t.Offers
}

func trainingOf(ps play.State) *game.Training {
	if t, ok := ps.(play.Trading); ok && t.Kind != play.TradeVault {
		return t.Training
	}
	return nil
}

func lootOf(ps play.State) *game.Loot {
	if t, ok := ps.(play.Looting); ok {
		return &game.Loot{Items: t.Items, Money: t.Money, Status: t.Status}
	}
	return nil
}

// combatOf is the fight up (nil: none).
func (r *rules) combatOf(ps play.State) *play.Combat {
	if f, ok := ps.(play.Fighting); ok {
		return r.combat(f)
	}
	return nil
}

// doorOf, introOf, sequenceOf, resultsOf, screenOf: the scene's or view's part.
func doorOf(ps play.State) *game.Door {
	if x, ok := ps.(play.Exploring); ok {
		return x.Door
	}
	return nil
}

func introOf(ps play.State) *game.Intro {
	if t, ok := ps.(play.Scene); ok && t.Kind == play.SceneTitle {
		return t.Intro
	}
	return nil
}

func sequenceOf(ps play.State) string {
	if t, ok := ps.(play.Scene); ok && t.Kind == play.SceneSequence {
		return t.Sequence
	}
	return ""
}

func resultsOf(ps play.State) *game.Results {
	if t, ok := ps.(play.Scene); ok && t.Kind == play.SceneResults {
		return t.Results
	}
	return nil
}

func screenOf(ps play.State) *game.Screen {
	switch t := ps.(type) {
	case play.Scene:
		if t.Kind == play.SceneScreen {
			return t.Screen
		}
	case play.ShipFight:
		f := t.Fight
		return &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: &f, ShipSheet: t.Sheet}
	}
	return nil
}

// rest is what resting now does.
func restOf(ps play.State) (heals, studies bool) {
	if t, ok := ps.(play.Camping); ok {
		return t.Heals, t.Study > 0
	}
	return false, false
}

// MusicCue is what the music follows.
type MusicCue struct {
	Where    game.Where
	Mode     game.Mode
	Title    bool // the title, or a sequence, is up
	Creating bool
	Results  bool
	Fighting bool
	Dungeon  bool // the fight is on dungeon ground
}

// MusicCue is what the music should follow now.
func (g *Game) MusicCue() MusicCue {
	ps, m := g.r.g.State(), g.mode()
	c := g.r.combatOf(ps)
	return MusicCue{Where: commonOf(ps).Where, Mode: m, Title: sequenceOf(ps) != "" || introOf(ps) != nil || m == game.ModeTitle,
		Creating: creationOf(ps) != nil || m == game.ModeCreate, Results: resultsOf(ps) != nil,
		Fighting: c != nil || m == game.ModeCombat, Dungeon: c != nil && c.Dungeon}
}

// What the sibling screens and a game's front tests read of the state up.
func (g *Game) Party() []game.Member     { return g.party() }
func (g *Game) Fight() *play.Combat      { return g.combat() }
func (g *Game) Pace() game.Pace          { return g.pace() }
func (g *Game) Common() play.Common      { return commonOf(g.r.g.State()) }
func (g *Game) Walls() *game.Walls       { return g.r.wallsOf(g.r.g.State()) }
func (g *Game) Loot() *game.Loot         { return lootOf(g.r.g.State()) }
func (g *Game) Automap() game.Automap    { return g.r.automapOf(g.r.g.State()) }
func (g *Game) Page() *game.Page         { return g.pane() }
func (g *Game) Results() *game.Results   { return resultsOf(g.r.g.State()) }
func (g *Game) Flags() *game.Flags       { return g.r.flagsOf(g.r.g.State(), g.party()) }
func (g *Game) SheetUp() *game.SheetPage { return g.r.sheetPageOf(g.r.g.State(), g.party()) }
func (g *Game) Shop() *game.Shop         { return shopOf(g.r.g.State()) }
func (g *Game) Training() *game.Training { return trainingOf(g.r.g.State()) }
func (g *Game) ItemsUp() game.ItemsUp    { return g.r.itemsUp(g.r.g.State()) }
