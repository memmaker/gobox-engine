package pool

import (
	"fmt"
	"gobox/core"
	"gobox/core/grid"
	"image"
	"slices"
	"strconv"
	"strings"

	"gobox/game"
	"gobox/game/play"
)

// facts are the engine's domain facts, read live, as gobox's data types:
// what State builds play's states from (play.go).
type facts struct {
	Mode                game.Mode
	ShopMenu            bool           // Pending is the shop's own choices, nothing over them
	Pending             *play.Question // the question up, its replies in play's order (a yes/no: No, Yes)
	Party               []game.Member
	Selected            int
	Where               game.Where
	Picture             game.Picture
	Story               []game.Picture
	LastText            string
	Purse               map[string]int
	Quests              []game.Quest
	Map                 game.Automap
	Combat              *play.Combat
	Shop                *game.Shop
	Temple              []game.Service
	TempleName          string
	Training            *game.Training
	Loot                *game.Loot
	Door                *game.Door
	Vault               *game.Vault
	Creation            *game.Creation
	TempleOffers        []game.Offer
	PartyCap, PoolItems int
	Picker              *game.Picker
	IconEditor          *game.IconEditor
	Page                *game.Page
	Flags               *game.Flags
	Walls               *game.Walls
	Intro               *game.Intro
	Sequence            string
	Results             *game.Results
	Read                *game.Read
	Lore                []game.Lore
	PowerPick           *game.PowerPick
	Sheet               *game.Sheet
	SheetPage           *game.SheetPage
}

// known is the engine's facts now.
func (e *Engine) known() facts {
	note, _ := e.Note()
	pick := e.powersView() // a pick owed stands in for the commands and the party menu
	exploring, sweeping := e.exploring(), e.exploring() && e.worldWord(core.WordSearch)&1 != 0
	px, py := e.posX, e.posY
	inside := e.MapCell(px, py) > 0x7F
	s := facts{
		Mode:     e.stateMode(),
		Pending:  e.pendingAsk(),
		ShopMenu: e.shopMenu(),
		LastText: note,
		Page:     e.statePage(),
		Flags: &game.Flags{Exploring: exploring && pick == nil, Sweeping: sweeping, Interact: exploring && e.interactHere(),
			CanCast: exploring && pick == nil && e.canCastNow(), Adventure: e.adventure(), Day: e.flagDay(inside), Inside: inside,
			Ordering: e.campOn(), Full: e.snap.full, IconDrop: e.snap.iconDrop != nil && e.snap.iconBox == (CellRect{})},
		Where: game.Where{
			Area: strconv.Itoa(int(AreaID(e.worldWord(core.WordGeo)))), X: px, Y: py, Facing: int(Dir(e.facing/2&3)) * 2,
			Inside: inside, Searching: sweeping,
			Mode: map[bool]string{true: "dungeon", false: "wilderness"}[e.worldWord(core.WordIn3D) != 0],
		},
	}
	members, quests := e.members(), e.questsView()
	for i, m := range members {
		gm := game.Member{Name: m.Name, Class: m.Class, Status: m.Status, HP: m.HP, MaxHP: m.MaxHP, AC: m.AC, Exp: uint32(m.Exp), Dropped: m.Portrait}
		if i < e.partySize() {
			id := e.partyIDs()[i]
			poolState(e.fact(id, core.PoolPicsFact{}).(core.PoolPics), &gm)
			e.iconState(id, &gm)
			gm.Stats = map[string]int{}
			scores := e.ruleset().Sheet(id).Scores
			for n, a := range map[string]core.Ability{"str": core.Strength, "int": core.Intelligence, "wis": core.Wisdom,
				"dex": core.Dexterity, "con": core.Constitution, "cha": core.Charisma} {
				gm.Stats[n] = int(scores[a])
			}
			book := e.fact(id, core.BookFact{}).(core.Book)
			for _, sp := range book.Known { // the book: the spells id knows
				gm.Spells = append(gm.Spells, game.SpellRef{ID: int(sp.Number), Name: sp.Name, Level: int(sp.Level),
					Planned: int(sp.Planned), Group: spellGroup(sp.Class, byte(sp.Level))})
			}
			gm.SlotsMax = map[string]int{}
			for cls, levels := range book.Slots {
				for lvl, n := range levels {
					gm.SlotsMax[spellGroup(core.SpellClass(cls), byte(lvl+1))] = int(n)
				}
			}
			sel := e.sel
			e.sel = i
			gm.CanCast = e.canCastNow()
			e.sel = sel
			for _, sp := range book.Memorized {
				gm.Spells = append(gm.Spells, game.SpellRef{ID: int(sp.Number), Name: sp.Name, Memorized: true})
			}
			gm.Items = e.itemsState(id)
			for _, l := range listLines(e.EffectsList(i)) {
				gm.Effects = append(gm.Effects, game.Line{Text: l.Text, Header: l.Header, Dim: l.Dim, Off: l.Off, Spell: l.Spell})
			}
			for _, p := range e.powers(id) {
				o := game.PowerOffer{Label: p.Name}
				switch p.Kind {
				case PowerSpells:
					o.Kind = game.PowerCast
				case PowerMagic:
					o.Kind = game.PowerDisplay // the spell effects, not a cast
				case PowerHeal:
					o.Kind = game.PowerHeal
				case PowerCure:
					o.Kind = game.PowerCure
				}
				gm.Powers = append(gm.Powers, o)
			}
			for _, l := range book.Levels {
				gm.Levels = append(gm.Levels, int(l))
			}
		}
		if m.Selected {
			s.Selected = i
		}
		s.Party = append(s.Party, gm)
	}
	if e.partySize() > 0 {
		s.Purse = purseMoney(e.purse(false))
	}
	for _, q := range quests {
		s.Quests = append(s.Quests, game.Quest{ID: q.Name, Text: fmt.Sprint(q.Text), Done: q.Done})
	}
	s.PartyCap = int(e.ruleset().Check(core.PoolCap{}, nil).Value)
	e.placeState(&s)
	e.mapState(&s)
	e.combatState(&s)
	s.Where.Hour, s.Where.Minute = e.worldWord(core.WordClockMinutes)/60, e.worldWord(core.WordClockMinutes)%60
	s.Where.Hall = e.ext.hall // the party menu is the training hall's
	if cr := e.creation; cr != nil {
		s.Creation = e.creationState(cr)
		if e.question() == nil {
			s.Mode = game.ModeCreate // the sheet is over the party menu
		}
	}
	if e.waiting(PendDoor) {
		s.Door = &game.Door{Locked: true}
	}
	e.pickState(&s)
	e.state3(&s)
	e.fillState2(&s, members, quests, pick)
	return s
}

// placeState is the shop, temple, treasure or training hall the party is
// at: what the script waits on (e.pending), the hall while it is open.
func (e *Engine) placeState(s *facts) {
	var goods []game.Item // as the wares list them, Wrong for the selected member
	sel := e.selected()
	for _, h := range e.goods() {
		none := e.itemCard(core.NoChar, h) // the goods are named for no one
		g := game.Item{Handle: h, Name: none.Name, Price: int(none.Price)}
		if sel != core.NoChar {
			card := e.itemCard(sel, h)
			g, g.Wrong = gameItem(h, card), card.CannotUse
			g.Name = e.itemCard(core.NoChar, h).Name
		}
		goods = append(goods, g)
	}
	switch e.pending {
	case PendShop:
		for i := range goods {
			goods[i].Price = int(e.shopPrice(int32(goods[i].Price), e.itemCard(core.NoChar, goods[i].Handle).Count))
		}
		s.Shop = &game.Shop{Goods: goods, Buys: true}
		s.Shop.Offers, s.Shop.IdentifyPrice = e.offers(s.Selected), 200
		identified := e.snap.identIt
		for i, id := range e.partyIDs() {
			l := e.listedItems(id)
			for _, h := range e.fact(id, core.IdentifiableFact{}).(core.Handles) {
				if !e.itemCard(id, h).Identified && i < len(s.Party) && slices.Contains(l, h) {
					s.Shop.Unknown = append(s.Shop.Unknown, game.ItemRef{Member: i, Item: h})
				}
			}
			if identified != core.NoItem && slices.Contains(e.fact(id, core.IdentifiableFact{}).(core.Handles), identified) {
				s.Shop.Identified = &game.ItemRef{Member: i, Item: identified}
			}
		}
		s.Shop.Identifies = e.snap.identN
	case PendTemple:
		for i := 0; ; i++ {
			c, ok := e.templeService(core.NoChar, i)
			if !ok {
				break
			}
			sv := game.Service{Name: c.Name, Price: int(c.Price), Not: map[int]string{}, Declines: c.Declines, Only: c.Only}
			for k, id := range e.partyIDs() {
				if e.fact(id, core.TempleCureFact{Cure: i}).(core.TempleCure).Needed {
					sv.Needs = append(sv.Needs, k)
				}
				sv.Not[k] = c.Not // the temple words it so for anyone it would not help
			}
			s.Temple = append(s.Temple, sv)
		}
		s.TempleOffers = e.offers(s.Selected)
	case PendLoot:
		for i := range goods {
			goods[i].Price = 0 // treasure is taken for nothing
		}
		l := e.purse(true)
		s.Loot = &game.Loot{Items: goods, Money: purseMoney(l)}
	}
	open, learning := e.hallOpen()
	if !open {
		return
	}
	t := &game.Training{Cost: int(e.ruleset().Check(core.TrainFee{}, nil).Value), Trains: map[int]string{}, Becomes: map[int][]string{}, Why: map[int]string{}, Learning: learning}
	s.Training = t
	for i, id := range e.partyIDs() {
		card := e.fact(id, core.TrainFact{}).(core.TrainCard)
		if t.Becomes[i] = card.Becomes; card.Why != "" {
			t.Why[i] = card.Why
			continue
		}
		t.Ready = append(t.Ready, i)
		t.Trains[i] = strings.Join(card.Becomes, " and ")
	}
	if learning {
		for _, sp := range e.fact(e.selected(), core.LearnableFact{}).(core.SpellCards) {
			t.Learn = append(t.Learn, game.SpellRef{ID: int(sp.Number), Name: sp.Name, Level: int(sp.Level)})
		}
	}
}

// purseMoney is p as gobox's coin map.
func purseMoney(p core.Purse) map[string]int {
	return map[string]int{game.Gold: int(p.Gold), game.Copper: int(p.Copper), game.Gems: int(p.Gems), game.Jewelry: int(p.Jewelry)}
}

// offers are the purse's gem and jewel to value for member m, or the one
// valued, with its value and whether there is room to keep it (Appraise).
func (e *Engine) offers(m int) []game.Offer {
	id := e.selected()
	if it := e.valued; it != core.NoItem {
		e.applyItems(id) // Overloaded applies the items first
		o := game.Offer{Member: m, Offer: game.OfferGem, Kind: game.Gems, Value: int(e.itemCard(id, it).Price), CanKeep: e.ruleset().Check(core.HasRoom{Who: id, Item: it}, nil).Yes}
		if e.ruleset().Check(core.ValuedItem{}, nil).Yes { // a jewel's name word
			o.Offer, o.Kind = game.OfferJewel, game.Jewelry
		}
		return []game.Offer{o}
	}
	var l []game.Offer
	purse := e.purse(false)
	if purse.Gems > 0 {
		l = append(l, game.Offer{Member: m, Offer: game.OfferGem, Kind: game.Gems})
	}
	if purse.Jewelry > 0 {
		l = append(l, game.Offer{Member: m, Offer: game.OfferJewel, Kind: game.Jewelry})
	}
	return l
}

// itemsState is what id carries (the party inventory after its own) in
// the order the inventory lists it: names as known, the body slot it is
// readied in, what Look tells of it, what a shop pays.
func (e *Engine) itemsState(id core.CharID) []game.Item {
	var l []game.Item
	for _, h := range e.listedItems(id) {
		card := e.itemCard(id, h)
		g := gameItem(h, card)
		g.Wrong = g.Kind != "misc" && card.CannotUse
		g.Sale = map[string]int{game.Gold: int(card.Sale / 200), game.Copper: int(card.Sale % 200)}
		if len(card.Scroll) > 0 {
			g.Spell, g.Scribe = int(card.Scroll[0].Number), card.Scribe
		}
		for _, sp := range card.Scroll {
			g.Scroll = append(g.Scroll, game.ScrollSpell{ID: int(sp.Number), Scribing: sp.Scribing})
		}
		l = append(l, g)
	}
	return l
}

// gameItem is card (of h) as gobox's item: its tab ("ring" for a ring, else as
// the item's kind), the slot it is readied in, its name's words, what Look
// tells of it, whether it is used, and its real name while unidentified.
func gameItem(h core.ItemHandle, card core.ItemCard) game.Item {
	return game.Item{Handle: h, Name: card.Name, Readied: card.Readied, Identified: card.Identified, Price: int(card.Price), Weight: int(card.Weight),
		Type: int(card.Type), Count: int(card.Count), Text: strings.Join(card.Look, "\n"), Usable: card.Usable,
		Kind: card.Tab, Slot: card.Slot, Words: card.Words, Hidden: card.Hidden}
}

// mapState is the current map as the party knows it, in the 3D view
// (mapView; NoMap where the game refuses its map, gobox's call): the squares seen, walls and doors (Kind 1 a door, 2 a locked
// one; a wall's Kind is its wall set), the labels (built in and the
// player's) as Marked and Notes, Area the map's name.
func (e *Engine) mapState(s *facts) {
	if !e.mapView() {
		if e.worldWord(core.WordIn3D) == 0 && e.adventure() { // the overland: its terrain
			m := game.Automap{Area: e.mapName(), Size: image.Pt(44, 36)}
			for y := range 36 {
				for x := range 44 {
					m.Known = append(m.Known, image.Pt(x, y)) // the whole map is shown
					m.Terrain = append(m.Terrain, int(e.ruleset().Check(core.AreaTileIs{X: int32(x), Y: int32(y)}, nil).Value))
				}
			}
			s.Map = m
			s.Where.X, s.Where.Y = e.overlandSquare().X, e.overlandSquare().Y
			s.Where.Square = &image.Point{X: e.posX, Y: e.posY} // the view's own (the dungeon's, stale here)
		}
		return
	}
	m := game.Automap{Area: e.mapName(), Size: image.Pt(16, 16), NoMap: e.worldWord(core.WordNoCoords) != 0}
	seen, mine, built := e.ext.seen[e.areaKey()], e.userLabels(), e.builtinLabels(e.areaKey())
	for y := range 16 {
		for x := range 16 {
			at := image.Pt(x, y)
			if seen != nil && seen[y<<4|x] != 0 {
				m.Known = append(m.Known, at)
			}
			for _, d := range [4]int{0, 2, 4, 6} {
				w, door := e.WallAt(x, y, d), e.DoorAt(x, y, d)
				switch {
				case w == 0:
				case door == 1:
					m.Doors = append(m.Doors, game.Wall{At: at, Side: d, Kind: int(w), State: game.DoorOpen})
				case door == 2:
					m.Doors = append(m.Doors, game.Wall{At: at, Side: d, Kind: int(w), State: game.DoorLocked})
				case door == 3:
					m.Doors = append(m.Doors, game.Wall{At: at, Side: d, Kind: int(w), State: game.DoorWizardLocked})
				default:
					m.Walls = append(m.Walls, game.Wall{At: at, Side: d, Kind: int(w)})
				}
			}
			if seen != nil && seen[y<<4|x] == seenVisited {
				m.Walked = append(m.Walked, at)
			}
			if l, ok := mine[y<<4|x]; ok {
				m.Marked, m.Notes, m.NoteKinds = append(m.Marked, at), append(m.Notes, l.text), append(m.NoteKinds, "")
			} else if l, ok := built[y<<4|x]; ok {
				m.Labels = append(m.Labels, game.Label{At: at, Text: l.text, Kind: labelKinds[l.kind]})
			}
		}
	}
	s.Map = m
}

// creationState is the creation sheet's draft and fields.
func (e *Engine) creationState(cr *creation) *game.Creation {
	id := cr.c
	sh := e.ruleset().Sheet(id)
	d := game.Member{Name: sh.Name, Race: e.word(id, core.FactRace), Class: e.word(id, core.FactClasses), Alignment: e.word(id, core.FactAlignment),
		HP: int(sh.HP), MaxHP: int(sh.MaxHP), Stats: map[string]int{}, Str00: int(sh.Scores[core.StrengthPercent])}
	for k, n := range game.StatKeys[:6] {
		d.Stats[n] = int(sh.Scores[abilities[k]])
	}
	var sp SheetPage // the figures as the sheet's redraw works them out (combatStats), on every roll
	e.applyItems(id)
	e.combatStats(id, &sp)
	d.AC, d.THAC0, d.Move, d.Age = sp.AC, sp.THAC0, sp.Move, int(sh.Age)
	poolState(e.fact(id, core.PoolPicsFact{}).(core.PoolPics), &d)
	e.iconState(id, &d)
	cl := e.Creation()
	mins := map[string]int{}
	for a, v := range cl.Minimums(cr.d.D) {
		if k := slices.Index(abilities[:], a); k >= 0 {
			mins[statKeys[k]] = int(v)
		}
	}
	g := &game.Creation{Draft: d, Minimums: mins, Roster: e.BuildCharList()}
	for _, f := range cl.Fields(cr.d.D) { // the four, then the power picks (setPower)
		fl := game.Field{Name: f.Label, Options: slices.Clone(f.Options), Pick: cr.d.D.Picks[f.ID]}
		for i := range f.Options {
			fl.Fits = append(fl.Fits, cl.Fits(cr.d.D, f.ID, i))
		}
		g.Fields = append(g.Fields, fl)
	}
	return g
}

// mainMenuLabels name the party menu's functions.
var mainMenuLabels = map[partyFn]string{
	partyCreate: "Create Character", partyAdd: "Add Character", partyRemove: "Remove Character",
	partyModify: "Modify Character", partyTraining: "Train Character", partyView: "View Character",
	partyLoad: "Load Saved Game", partyBegin: "Begin Adventuring", partyQuit: "Exit to DOS",
	partyDrop: "Drop Character", partyNewClass: "Human Change Class", partyQuickStart: "Quick Start",
	partyForget: "Forget Character",
}

// shopReplies are the shop's own choices when nothing is asked.
var shopReplies = []string{"Buy", "Leave"}

// shopMenu: the shop's own choices are up, nothing over them.
func (e *Engine) shopMenu() bool {
	return e.question() == nil && e.creation == nil && !e.menuUp() && e.waiting(PendShop)
}

// pendingAsk is the engine's question, the party menu or the shop's choices.
func (e *Engine) pendingAsk() *play.Question {
	if q := e.question(); q != nil {
		return q
	}
	if cr := e.creation; cr != nil { // Esc or Exit on the sheet asked this before abandoning it; the open sheet asks nothing (class 50)
		if cr.naming {
			return &play.Question{Text: "Character name:", Typed: true, Max: 0xF} // Rename cuts at 15
		}
		if !cr.asking {
			return nil
		}
		return &play.Question{Text: cr.abandon, YesNo: true, Replies: []string{"No", "Yes"}, Default: 1}
	}
	if e.menuUp() {
		q := &play.Question{Text: "Choose a function"}
		for _, k := range e.partyFunctions() { // the engine's, not the View's: the front builds its menu from these
			q.Replies = append(q.Replies, mainMenuLabels[k])
		}
		return q
	}
	if e.shopMenu() {
		return &play.Question{Replies: slices.Clone(shopReplies), Esc: play.EscTo(slices.Index(shopReplies, "Leave"))}
	}
	return nil
}

// dirOf is the direction (0 2 4 6) of a one-square step v, or -1.
func dirOf(v image.Point) int {
	for d := 0; d < 8; d += 2 {
		if v == image.Pt(grid.StepX[d], grid.StepY[d]) {
			return d
		}
	}
	return -1
}

// refusal is err as a play.Refusal.
func refusal(err error) *play.Refusal {
	if f, ok := err.(*play.Refusal); ok {
		return f
	}
	return &play.Refusal{Text: err.Error()}
}

// question is the question the engine waits on (the Ask), nil for none: a
// list's and a member pick's rows are its replies, a menu with none a Continue.
func (e *Engine) question() *play.Question {
	w := e.eclWait
	if e.top() != onAsk || w == nil {
		return nil
	}
	q := *w.ask
	q.Replies, q.Keys = slices.Clone(q.Replies), slices.Clone(q.Keys)
	if q.List && !w.member {
		q.Replies = nil
		for _, l := range w.lines {
			q.Replies = append(q.Replies, l.Text)
		}
	}
	if len(q.Replies) == 0 && !q.Typed && !q.Number {
		q.Continue = true
	}
	if w.shape == shPause { // DELAY and the other timed pauses: the presenter paces them (play.Question.Ms)
		q.Ms = w.ms
	}
	return &q
}

// labelKinds name the label kinds (game.Label.Kind); the player's are "".
var labelKinds = [...]string{poiUser: "", poiShop: "shop", poiTemple: "temple", poiInn: "inn", poiCivic: "civic"}

// spellGroup is the level group a spell of class cls and level lvl is
// listed under (SpellRef.Group, the SlotsMax keys): "cleric 1", "mage 2".
func spellGroup(cls core.SpellClass, lvl byte) string {
	return cls.String() + " " + strconv.Itoa(int(lvl))
}

// gamePicture is p (the game's pixels) as a game.Picture.
func gamePicture(p Picture) game.Picture {
	if p.Pix == nil || p.PixW == 0 {
		return game.Picture{}
	}
	g := game.Picture{W: p.PixW, H: len(p.Pix) / p.PixW, Pixels: p.Pix, Palette: p.Palette, Clear: -1}
	if p.Mask != nil {
		g.Clear = int(*p.Mask)
	}
	return g
}

func gameRuns(l []PaneRun) []game.TextRun {
	var r []game.TextRun
	for _, t := range l {
		r = append(r, game.TextRun{X: t.X, Text: t.Text, Fg: t.Fg, Bg: t.Bg, Upcase: t.Upcase})
	}
	return r
}

func gameLines(ls [][]PaneRun) [][]game.TextRun {
	if ls == nil {
		return nil
	}
	r := make([][]game.TextRun, len(ls))
	for i, l := range ls {
		r[i] = gameRuns(l)
	}
	return r
}

// Open runs the game (Run: Init and the title) to its first wait; with
// Explore it starts from save slot slot ("" a Quick Start).
func (e *Engine) Open(slot string) {
	if e.Explore {
		e.ExploreSlot = slot
	}
	e.Start(e.Run)
}

// iconState puts c's original combat icon numbers (the CHEAD and CBODY blocks) and colours
// in a State member: gobox composes the icon from them (or from its pool ids).
func (e *Engine) iconState(id core.CharID, m *game.Member) {
	ic := e.fact(id, core.IconPicsFact{}).(core.IconPics)
	off := 0
	if ic.Size == 2 { // the T files' blocks follow the S ones' by 0x40 (LoadIconPic)
		off = 0x40
	}
	m.Icon, m.IconColors = map[string]int{"head": ic.Head + off, "body": ic.Body + off}, nil
	for _, col := range ic.Colours { // the record's bytes: low nibble and high nibble
		m.IconColors = append(m.IconColors, col.High<<4|col.Low)
	}
}

// overlandSquare is the party's square on the overland (Where.X, Y there);
// PosX, PosY stay the 3D view's.
func (e *Engine) overlandSquare() image.Point {
	return image.Pt(int(ViewModeShift[e.viewMode])+e.worldWord(core.WordMapX), e.worldWord(core.WordMapY))
}
