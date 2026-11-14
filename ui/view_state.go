// view_state.go: what the screens read of the state up, each from play's types.

package ui

import (
	"cmp"
	"image"
	"reflect"

	"gobox/game"
	"gobox/rules"
)

// partyOf is ps's party as the screens show it.
func partyOf(ps game.State) []game.Member {
	c := game.CommonOf(ps)
	var out []game.Member
	for i, m := range c.Party {
		m.CanCast = m.CanCast || canCast(c.Allowed, i)
		out = append(out, m)
	}
	return out
}

// memberCard is what only the front knows of a member's card.
type memberCard struct {
	Chosen bool         // the card is marked: the selected member
	Menu   []game.Entry // its context menu out of combat
}

// cardsOf are ps's party's cards, by member.
func cardsOf(ps game.State) []memberCard {
	c := game.CommonOf(ps)
	var out []memberCard
	for i, m := range partyOf(ps) {
		out = append(out, memberCard{Chosen: i == c.Selected, Menu: card(m, c.Allowed, i)})
	}
	return out
}

// selectedOf is the member the screens act on: the one whose look-only sheet
// is up in a fight, else with the items up whose items they are.
func (r *frontState) selectedOf(ps game.State) int {
	if n, ok := r.lookMember(ps); ok { // a fight's look-only sheet: that member's, and its items
		return n
	}
	c := game.CommonOf(ps)
	if f, ok := ps.(game.Fighting); ok {
		if m := cmp.Or(r.itemsOf+1, activeMember(f.Field)+1) - 1; r.items != game.ItemsNone && m >= 0 && m < len(c.Party) {
			return m
		}
	}
	return c.Selected
}

// allowedOf is what ps allows by action name, and NoRest where camp has no
// timed rest.
func allowedOf(ps game.State) map[string]bool {
	m := allowedNames(game.CommonOf(ps).Allowed)
	if _, camp := ps.(game.Camping); camp && !m["Rest"] {
		m["NoRest"] = true
	}
	return m
}

// automapOf is the map of where the party is; camp keeps the explored one.
func automapOf(ps game.State, explored game.Automap) game.Automap {
	c := game.CommonOf(ps)
	a := c.Automap
	if x, ok := ps.(game.Exploring); ok && a.Size == (rules.Size{}) {
		a = x.Map.Automap
	}
	if _, ok := ps.(game.Camping); ok && len(a.Known) == 0 && len(explored.Known) > 0 {
		a = explored
	}
	if a.Area == "" {
		a.Area = c.Area
	}
	return a
}

// storyOf is the speaker's portrait, head and body.
func storyOf(c game.Common) []game.Picture {
	if pt := c.Portrait; pt != nil {
		return []game.Picture{pt.Head, pt.Body}
	}
	return nil
}

// pageOf is the text page up.
func (r *frontState) pageOf(ps game.State) *TextPage {
	c := game.CommonOf(ps)
	p := &TextPage{}
	for _, d := range c.Layers {
		p.Layers = append(p.Layers, drawnLayer(d))
	}
	if c.Picture != nil && len(c.Layers) == 0 { // the front draws a page's layers: the picture is the page's only one
		p.Layers = []Picture{drawnPicture(*c.Picture, image.Rectangle{})}
	}
	for _, t := range c.Text {
		p.Pane = append(p.Pane, r.run(t))
	}
	if c.Status != "" {
		p.Prompt = r.run(c.Status)
	}
	if t, ok := ps.(game.Scene); ok {
		pg := r.page(t)
		if t.Kind == game.SceneTitle && !r.title.credits {
			if st := scriptOf(t.Intro).Step(r.titleStep(t)); st != nil {
				pg.Layers = append(pg.Layers, scriptPictures(st, t.Intro)...)
				for _, line := range st.Text { // the step's words, over its picture
					pg.Pane = append(pg.Pane, r.run(line))
				}
			}
		}
		p.Pane, p.View, p.Layers = append(p.Pane, pg.Pane...), pg.View, append(p.Layers, pg.Layers...)
	}
	return p
}

// flagsOf are the screen's plain status flags.
func (r *frontState) flagsOf(ps game.State, party []game.Member) *game.Flags {
	c := game.CommonOf(ps)
	_, camp := ps.(game.Camping)
	f := &game.Flags{Adventure: c.Adventure, Ordering: camp && allowedOf(ps)["Order"], Sweeping: c.Where.Searching,
		Inside: c.Where.Inside, CanCast: c.Selected >= 0 && c.Selected < len(party) && party[c.Selected].CanCast,
		IconDrop: c.IconEditor != nil, Day: c.Day, NoSave: c.NoSave, NoLoad: c.NoLoad}
	switch t := ps.(type) {
	case game.Exploring:
		f.Exploring, f.Day, f.Interact = c.Ask == nil, t.Day || c.Day, t.Interact && c.Ask == nil // the game's say, never the hour guessed
	case game.Scene:
		f.Full = t.Full
		if t.Kind == game.SceneTitle {
			f.Credits, f.CreditPage = r.title.credits, r.title.page
			st := scriptOf(t.Intro).Step(r.titleStep(t))
			f.CreditsPic = !r.title.credits && st != nil && st.Credits && scriptOf(t.Intro).Plays(r.titleStep(t), amigaShown())
		}
	}
	return f
}

// windowOf is what the view window shows.
func windowOf(ps game.State) game.Window {
	c := game.CommonOf(ps)
	pictured := c.Picture != nil || len(c.Layers) > 0 || c.Portrait != nil
	switch t := ps.(type) {
	case game.Exploring:
		switch {
		case c.Picture != nil || c.Portrait != nil:
			return game.WindowPicture
		case t.Map.Overland: // the map screen of its own: the page's layers are the original's small map of it, which gobox draws whole
			return game.WindowMap
		case pictured:
			return game.WindowPicture
		}
		return game.WindowScene
	case game.Scene:
		switch {
		case t.Kind == game.SceneTitle, t.Kind == game.SceneSequence:
			return game.WindowScene
		case len(t.Pictures)+len(t.Layers) > 0:
			return game.WindowPicture
		case len(t.Placed) > 0: // text over the view window (opClearBox, the runes)
			return game.WindowText
		}
		return game.WindowScene // nothing over the view: the 3D view under the text (Rolf's tour between his pictures)
	}
	if pictured {
		return game.WindowPicture
	}
	var none game.Window
	return none
}

// sheetOf is what the selected member's sheet offers (nil: none up).
func (r *frontState) sheetOf(c game.Common) *game.Sheet {
	if n := c.Selected; n >= 0 && n < len(c.Party) && (len(c.Party[n].SheetPowers) > 0 || len(c.Party[n].Reactions) > 0) {
		return &game.Sheet{Powers: c.Party[n].SheetPowers, Reactions: c.Party[n].Reactions}
	}
	if n := r.sheet; n >= 0 && n < len(c.Party) {
		return &game.Sheet{} // up though the member has no powers: the front shows a page only with one
	}
	return nil
}

// sheetPageOf is the sheet's page up: the rules', else gobox's from the member.
func (r *frontState) sheetPageOf(ps game.State, party []game.Member) *game.SheetPage {
	c := game.CommonOf(ps)
	var p *game.SheetPage
	if n := r.sheet; n >= 0 && n < len(c.Party) {
		if p = c.Party[n].Page; p == nil {
			p = memberPage(party[n])
		}
	}
	if a, ok := ps.(game.Assembling); ok && a.Draft != nil && a.Draft.Member.Page != nil {
		p = a.Draft.Member.Page // the rules' figures for the sheet being made
	}
	return p
}

// creationOf is the draft up (nil: none).
func creationOf(ps game.State) *game.Creation {
	a, ok := ps.(game.Assembling)
	if !ok || a.Draft == nil {
		return nil
	}
	d := a.Draft
	cr := &game.Creation{Draft: d.Member, Fields: d.Fields, Roster: a.Roster,
		MinHP: d.MinHP, MaxHP: d.MaxHP, Minimums: d.Minimums, Step: d.Step, Stats: d.Stats, Derived: d.Derived}
	if d.Ask != nil {
		cr.Choices = d.Ask.Replies
	}
	return cr
}

// pacingOf is what goes on with time.
func (r *frontState) pacingOf(ps game.State) *game.Pacing {
	p := &game.Pacing{Seq: r.seq}
	switch {
	case r.rest > 0:
		p.Pace, p.Resting = game.PaceRest, true
	case r.walk != nil:
		p.Pace = game.PaceWalk
	}
	return p
}

// wallsOf is the 3D view up (shown) and the one last explored (kept): camp
// a scene and the party menu keep it (the party menu: the engine's own
// when it gives one, a loaded game's hall); a fight drops it.
func wallsOf(ps game.State, kept *game.Walls) (shown, keep *game.Walls) {
	switch t := ps.(type) {
	case game.Exploring:
		if w := t.Map.Walls; w.Side != nil { // no squares: no 3D view (the overland)
			kept = gameWalls(w)
		} else {
			kept = nil
		}
	case game.Scene:
		if t.Walls != nil && t.Walls.Side != nil {
			return gameWalls(*t.Walls), kept
		}
	case game.Assembling: // the party menu never drops the 3D view
		if t.Walls != nil && t.Walls.Side != nil {
			kept = gameWalls(*t.Walls)
		}
	case game.Camping:
	case game.Trading: // a shop, temple, hall or vault stays over the 3D view it was entered from
	default:
		kept = nil
	}
	return kept, kept
}

// readOf is the handout up, kept so the pointer holds while it does.
func (r *frontState) readOf(ps game.State) *game.Read {
	t, ok := ps.(game.Scene)
	if !ok || t.Kind != game.SceneHandout {
		r.read = nil
		return nil
	}
	if r.read == nil || r.read.Title != t.Handout || !reflect.DeepEqual(r.read.Texts, t.Texts) {
		r.read = &game.Read{Title: t.Handout, Texts: t.Texts}
	}
	return r.read
}

// modeOf is the screen ps puts up, a question over it counted.
func (r *frontState) modeOf(ps game.State) game.Mode {
	if r.refusalAsk != nil { // a Refusal's question is up
		return game.ModeAsk
	}
	var m game.Mode
	c := game.CommonOf(ps)
	switch t := ps.(type) {
	case game.Assembling:
		m = game.ModeParty
		if t.Draft != nil {
			m = game.ModeCreate
		}
	case game.Exploring:
		m = game.ModeExplore
		if c.Ask != nil && !(c.Ask.Continue && len(c.Ask.Replies) == 0) || t.Door != nil { // a question (or the door menu) over the view
			m = game.ModeAsk
		}
	case game.Camping:
		m = game.ModeCamp
	case game.Fighting:
		m = game.ModeCombat
	case game.Scene:
		m = game.ModeStory
		if t.Kind == game.SceneTitle {
			m = game.ModeTitle
		}
	case game.Trading:
		switch {
		case t.Kind == game.TradeVault:
			m = game.ModeVault
		case t.Kind == game.TradeTraining, t.Training != nil:
			m = game.ModeTraining
		case t.Kind == game.TradeTemple, len(t.Services) > 0:
			m = game.ModeTemple
		default:
			m = game.ModeShop
		}
	case game.Looting:
		m = game.ModeLoot
	case game.Over:
		m = game.ModeTitle
	case game.ShipFight:
		m = game.ModeStory
	}
	return askMode(m, r.pending(ps), ps)
}

// itemsOf is whose goods the inventory screen shows.
func (r *frontState) itemsUp(ps game.State) game.ItemsUp {
	if r.items == game.ItemsNone {
		return r.items
	}
	switch ps.(type) {
	case game.Trading:
		return game.ItemsShop
	case game.Looting:
		return game.ItemsLoot
	}
	return r.items
}

// shopOf is the shop up (nil: none); a vault or hall is no shop.
func shopOf(ps game.State) *game.Shop {
	t, ok := ps.(game.Trading)
	if !ok || t.Kind == game.TradeVault || t.Kind == game.TradeTraining || t.Training != nil || t.Kind == game.TradeTemple || len(t.Services) > 0 {
		return nil
	}
	return &game.Shop{Name: t.Name, Goods: t.Goods, Buys: t.Buys, Offers: t.Offers, Pooled: t.Pooled,
		Reloads: allowedOf(ps)["Reload"], IdentifyPrice: t.Identify, Unknown: t.Unknown, Identified: t.Known}
}

// templeOf is the temple up: its services, offers (nil services: none).
func templeOf(ps game.State) ([]game.Service, []game.Offer) {
	t, ok := ps.(game.Trading)
	if !ok || t.Kind == game.TradeVault || t.Kind == game.TradeTraining || t.Training != nil || t.Kind != game.TradeTemple && len(t.Services) == 0 {
		return nil, nil
	}
	return t.Services, t.Offers
}

func trainingOf(ps game.State) *game.Training {
	if t, ok := ps.(game.Trading); ok && t.Kind != game.TradeVault {
		return t.Training
	}
	return nil
}

func lootOf(ps game.State) *game.Loot {
	if t, ok := ps.(game.Looting); ok {
		return &game.Loot{Items: t.Items, Money: t.Money, Status: t.Status}
	}
	return nil
}

// combatOf is the fight up (nil: none).
func (r *frontState) combatOf(ps game.State) *game.Combat {
	if f, ok := ps.(game.Fighting); ok {
		return r.combat(f)
	}
	return nil
}

// doorOf, introOf, sequenceOf, resultsOf, screenOf: the scene's or view's part.
func doorOf(ps game.State) *game.Door {
	if x, ok := ps.(game.Exploring); ok {
		return x.Door
	}
	return nil
}

func introOf(ps game.State) *game.Intro {
	if t, ok := ps.(game.Scene); ok && t.Kind == game.SceneTitle {
		return t.Intro
	}
	return nil
}

func sequenceOf(ps game.State) string {
	if t, ok := ps.(game.Scene); ok && t.Kind == game.SceneSequence {
		return t.Sequence
	}
	return ""
}

func resultsOf(ps game.State) *game.Results {
	if t, ok := ps.(game.Scene); ok && t.Kind == game.SceneResults {
		return t.Results
	}
	return nil
}

func screenOf(ps game.State) *game.Screen {
	switch t := ps.(type) {
	case game.Scene:
		if t.Kind == game.SceneScreen {
			return t.Screen
		}
	case game.ShipFight:
		f := t.Fight
		return &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: &f, ShipSheet: t.Sheet}
	}
	return nil
}

// rest is what resting now does.
func restOf(ps game.State) (heals, studies bool) {
	if t, ok := ps.(game.Camping); ok {
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
	return MusicCue{Where: game.CommonOf(ps).Where, Mode: m, Title: sequenceOf(ps) != "" || introOf(ps) != nil || m == game.ModeTitle,
		Creating: creationOf(ps) != nil || m == game.ModeCreate, Results: resultsOf(ps) != nil,
		Fighting: c != nil || m == game.ModeCombat, Dungeon: c != nil && c.Dungeon}
}

// What the sibling screens and a game's front tests read of the state up.
func (g *Game) Party() []game.Member     { return g.party() }
func (g *Game) Fight() *game.Combat      { return g.combat() }
func (g *Game) Pace() game.Pace          { return g.pace() }
func (g *Game) Common() game.Common      { return game.CommonOf(g.r.g.State()) }
func (g *Game) Walls() *game.Walls       { return g.r.derived().walls }
func (g *Game) Loot() *game.Loot         { return lootOf(g.r.g.State()) }
func (g *Game) Automap() game.Automap    { return g.r.derived().automap }
func (g *Game) Page() *TextPage          { return g.pane() }
func (g *Game) Results() *game.Results   { return resultsOf(g.r.g.State()) }
func (g *Game) Flags() *game.Flags       { return g.r.flagsOf(g.r.g.State(), g.party()) }
func (g *Game) Window() game.Window      { return windowOf(g.r.g.State()) }
func (g *Game) SheetUp() *game.SheetPage { return g.r.sheetPageOf(g.r.g.State(), g.party()) }
func (g *Game) Shop() *game.Shop         { return shopOf(g.r.g.State()) }
func (g *Game) Training() *game.Training { return trainingOf(g.r.g.State()) }
func (g *Game) ItemsUp() game.ItemsUp    { return g.r.itemsUp(g.r.g.State()) }

// titleStep is the script's step shown (adventure.Sequence.Step): before gobox's first
// step, that step's picture already.
func (r *frontState) titleStep(t game.Scene) int {
	if n := r.title.n; n != 0 || t.Intro == nil || len(t.Intro.Pictures) == 0 {
		return n
	}
	return t.Intro.Pictures[0]
}
