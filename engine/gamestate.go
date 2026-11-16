package engine

import (
	"fmt"
	"gobox/adventure"
	"gobox/rules"
	"image"
	"slices"
	"strconv"
	"strings"

	"gobox/game"
)

// partyState is the party's cards as State shows them, and the selected
// member's index: the selected one carries the sheet up (BUG-CLASSES 36).
func (e *Engine) partyState() (party []game.Member, selected int) {
	members, chars := e.members(), e.partySide()
	for i, m := range members {
		gm := game.Member{Name: m.Name, Class: m.Class, Status: m.Status, Health: m.Health, HP: m.HP, MaxHP: m.MaxHP, AC: m.AC, Exp: uint32(m.Exp), Dropped: m.Portrait,
			LevelExp: m.LevelExp, NextExp: m.NextExp, SpellsReady: m.Spells, Memo: m.Memo}
		if i < e.partySize() {
			id := e.partyIDs()[i]
			poolState(e.poolPics(id), &gm)
			e.iconState(id, &gm)
			gm.Stats = map[string]int{}
			scores := e.ruleset().Sheet(id).Scores
			for n, a := range map[string]rules.Ability{"str": rules.Strength, "int": rules.Intelligence, "wis": rules.Wisdom,
				"dex": rules.Dexterity, "con": rules.Constitution, "cha": rules.Charisma} {
				gm.Stats[n] = int(scores[a])
			}
			book := e.fact(id, rules.BookFact{}).(rules.Book)
			for _, sp := range book.Known { // the book: the spells id knows
				gm.Spells = append(gm.Spells, game.SpellRef{ID: int(sp.Number), Name: sp.Name, Level: int(sp.Level),
					Planned: int(sp.Planned), Group: spellGroup(sp.Class, byte(sp.Level))})
			}
			gm.SlotsMax = map[string]int{}
			if slices.Contains(e.ruleset().Catalog().Features, rules.FeatureSpells) { // a ruleset without spells has no slots to list
				for cls, levels := range book.Slots {
					for lvl, n := range levels {
						gm.SlotsMax[spellGroup(rules.SpellClass(cls), byte(lvl+1))] = int(n)
					}
				}
			}
			gm.CanCast = e.canCast(id)
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
		if i < len(chars) {
			gm.Page = gameSheetPage(e.sheetPage(chars[i])) // BUG-CLASSES 36: gobox's look-only sheet (a fight's, a shop's View)
		}
		if m.Selected {
			selected = i
		}
		party = append(party, gm)
	}
	if selected < len(party) { // the selected member's sheet up: its powers, reactions and page
		pm := &party[selected]
		if e.sheetUp() {
			if sh := e.sheetView(); sh != nil {
				pm.SheetPowers = gameEntries(sh.Powers)
				for _, r := range sh.Reactions {
					pm.Reactions = append(pm.Reactions, game.Reaction{ID: r.ID, Name: r.Name, Policy: int(r.Policy)})
				}
			}
		}
		if pg := gameSheetPage(e.snap.sheet); pg != nil {
			pm.Page = pg
		}
	}
	return party, selected
}

// whereState is where the party is and the map as it knows it.
func (e *Engine) whereState() (game.Where, game.Automap) {
	px, py := e.posX, e.posY
	inside := e.MapCell(px, py) > 0x7F
	w := game.Where{
		Area: strconv.Itoa(int(AreaID(int(e.world.Geo)))), X: px, Y: py, Facing: int(Dir(e.facing/2&3)) * 2,
		Inside: inside, Searching: e.exploring() && e.world.Search&1 != 0,
		Mode: map[bool]string{true: "dungeon", false: "wilderness"}[e.world.In3D != 0],
	}
	m := e.mapState(&w)
	w.Hour, w.Minute = e.clockPart(rules.ClockMinuteOfDay)/60, e.clockPart(rules.ClockMinuteOfDay)%60
	w.Hall = e.ext.hall // the party menu is the training hall's
	return w, m
}

// questsState is the journal's quests.
func (e *Engine) questsState() []game.Quest {
	var l []game.Quest
	for _, q := range e.questsView() {
		l = append(l, game.Quest{ID: q.Name, Text: fmt.Sprint(q.Text), Done: q.Done, Giver: q.Giver, Place: q.Place, Brief: q.Text})
	}
	return l
}

// loreState is the journal's lore.
func (e *Engine) loreState() []game.Lore {
	var l []game.Lore
	for _, x := range e.loreView() {
		l = append(l, game.Lore{Title: x.Title, Place: x.Place, Text: x.Text, ID: x.ID})
	}
	return l
}

// powerPickState is the power pick owed, nil for none.
func (e *Engine) powerPickState() *game.PowerPick {
	p := e.powersView()
	if p == nil {
		return nil
	}
	k := &game.PowerPick{Name: p.Name, What: p.What}
	for _, m := range p.Options {
		k.Options = append(k.Options, game.PowerOption{ID: int(m.ID), Name: m.Name, Tip: m.Tip, Aim: int(m.Aim), Range: m.Range, Area: m.Area, Move: m.Move})
	}
	return k
}

// doorState is the locked door waiting (Exploring's Door), nil for none.
func (e *Engine) doorState() *game.Door {
	if !e.waiting(PendDoor) {
		return nil
	}
	d := e.ext.door.opts
	return &game.Door{Locked: true, Bash: d.Bash, Pick: d.Pick, Knock: d.Knock}
}

// goodsState are the wares as they list them, Wrong for the selected member.
func (e *Engine) goodsState() []game.Item {
	var goods []game.Item
	sel := e.selected()
	for _, h := range e.goods() {
		none := e.itemCard(rules.NoChar, h) // the goods are named for no one
		g := game.Item{Handle: h, Name: none.Name, Price: int(none.Price)}
		if sel != rules.NoChar {
			card := e.itemCard(sel, h)
			g, g.Wrong = gameItem(h, card), card.CannotUse
			g.Name = e.itemCard(rules.NoChar, h).Name
		}
		goods = append(goods, g)
	}
	return goods
}

// identifyFee is the Pool shop's price to identify an item (core's Identify, ovr07; the rules take it at IdentifyItem).
const identifyFee = 200

// tradeState is the shop or temple the script waits on (e.pending) and the
// training hall while it is open, as Trading.
func (e *Engine) tradeState(mode game.Mode, cm game.Common) game.Trading {
	t := game.Trading{Common: cm, Kind: tradeKinds[mode]}
	switch e.pending {
	case PendShop:
		goods := e.goodsState()
		for i := range goods {
			goods[i].Price = int(e.shopPrice(int32(goods[i].Price), e.itemCard(rules.NoChar, goods[i].Handle).Count))
		}
		t.Goods, t.Buys, t.Offers = goods, true, e.offers(cm.Selected)
		identified := e.snap.identIt
		for i, id := range e.partyIDs() {
			l := e.listedItems(id)
			for _, h := range e.fact(id, rules.IdentifiableFact{}).(rules.Handles) {
				if !e.itemCard(id, h).Identified && i < len(cm.Party) && slices.Contains(l, h) {
					t.Unknown = append(t.Unknown, game.ItemRef{Member: i, Item: h})
				}
			}
			if identified != rules.NoItem && slices.Contains(e.fact(id, rules.IdentifiableFact{}).(rules.Handles), identified) {
				t.Known = &game.ItemRef{Member: i, Item: identified}
			}
		}
		if len(t.Unknown) > 0 || t.Known != nil { // the shop identifies only where the ruleset lists items it can (IdentifiableFact)
			t.Identify = identifyFee
		}
	case PendTemple:
		for i := 0; ; i++ {
			c, ok := e.templeService(rules.NoChar, i)
			if !ok {
				break
			}
			sv := game.Service{Name: c.Name, Price: int(c.Price)}
			for k, id := range e.partyIDs() {
				if e.fact(id, rules.TempleCureFact{Cure: i}).(rules.TempleCure).Needed {
					sv.Needs = append(sv.Needs, k)
				}
			}
			t.Services = append(t.Services, sv)
		}
		t.Offers = e.offers(cm.Selected)
	}
	t.Training = e.trainingState()
	return t
}

// lootState is the treasure the party is at, as Looting.
func (e *Engine) lootState(cm game.Common) game.Looting {
	goods := e.goodsState()
	for i := range goods {
		goods[i].Price = 0 // treasure is taken for nothing
	}
	return game.Looting{Common: cm, Items: goods, Money: e.purseMoney(e.purse(true)), Announce: !e.lootSeen}
}

// trainingState is the training hall while it is open, nil when shut.
func (e *Engine) trainingState() *game.Training {
	open, learning := e.hallOpen()
	if !open {
		return nil
	}
	t := &game.Training{Cost: int(e.ruleset().Check(rules.TrainFee{}, nil).Value), Trains: map[int]string{}, Becomes: map[int][]string{}, Why: map[int]string{}, Learning: learning}
	for i, id := range e.partyIDs() {
		card := e.fact(id, rules.TrainFact{Hall: e.hallTerms()}).(rules.TrainCard)
		if t.Becomes[i] = card.Becomes; card.Why != "" {
			t.Why[i] = card.Why
			continue
		}
		t.Ready = append(t.Ready, i)
		t.Trains[i] = strings.Join(card.Becomes, " and ")
	}
	if learning {
		for _, sp := range e.fact(e.selected(), rules.LearnableFact{}).(rules.SpellCards) {
			t.Learn = append(t.Learn, game.SpellRef{ID: int(sp.Number), Name: sp.Name, Level: int(sp.Level)})
		}
	}
	return t
}

// purseMoney is p as gobox's coin map.
func (e *Engine) purseMoney(p rules.Purse) map[string]int {
	if c := e.gameData.Coin; c != "" { // the game counts in one coin of its own (game.Data.Coin: BR's credits): the purse's base unit (rules.Purse: Copper), and no gold or copper beside it
		return map[string]int{c: int(p.Copper), game.Gems: int(p.Gems), game.Jewelry: int(p.Jewelry)}
	}
	return map[string]int{game.Gold: int(p.Gold), game.Copper: int(p.Copper), game.Gems: int(p.Gems), game.Jewelry: int(p.Jewelry)}
}

// offers are the purse's gem and jewel to value for member m, or the one
// valued, with its value and whether there is room to keep it (Appraise).
func (e *Engine) offers(m int) []game.Offer {
	id := e.selected()
	if it := e.valued; it != rules.NoItem {
		e.applyItems(id) // Overloaded applies the items first
		o := game.Offer{Member: m, Offer: game.OfferGem, Kind: game.Gems, Value: int(e.itemCard(id, it).Price), CanKeep: e.ruleset().Check(rules.HasRoom{Who: id, Item: it}, nil).Yes}
		if e.ruleset().Check(rules.ValuedItem{}, nil).Yes { // a jewel's name word
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
func (e *Engine) itemsState(id rules.CharID) []game.Item {
	var l []game.Item
	for _, h := range e.listedItems(id) {
		card := e.itemCard(id, h)
		g := gameItem(h, card)
		g.Wrong = g.Kind != "misc" && card.CannotUse
		g.Sale = e.purseMoney(card.Sale)
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
func gameItem(h rules.ItemHandle, card rules.ItemCard) game.Item {
	return game.Item{Handle: h, Name: card.Name, Readied: card.Readied, Identified: card.Identified, Price: int(card.Price), Weight: int(card.Weight),
		Type: int(card.Type), Base: card.Kind, Count: int(card.Count), Text: strings.Join(card.Look, "\n"), Usable: card.Usable,
		Kind: card.Tab, Slot: card.Slot, Words: card.Words, Hidden: card.Hidden}
}

// mapState is the current map as the party knows it, in the 3D view
// (mapView; NoMap where the game refuses its map, gobox's call): the squares seen, walls and doors (Kind 1 a door, 2 a locked
// one; a wall's Kind is its wall set), the labels (built in and the
// player's) as Marked and Notes, Area the map's name.
func (e *Engine) mapState(w *game.Where) game.Automap {
	if !e.mapView() {
		if e.world.In3D == 0 && e.adventure() { // the overland: its terrain
			m := game.Automap{Area: e.mapName(), Size: rules.Size{W: 44, H: 36}}
			marks := e.areaMarks()
			for y := range 36 {
				for x := range 44 {
					m.Known = append(m.Known, rules.MapPos{X: x, Y: y}) // the whole map is shown
					m.Terrain = append(m.Terrain, int(e.ruleset().Check(rules.AreaTileIs{X: int32(x), Y: int32(y), Marks: marks}, nil).Value))
				}
			}
			w.X, w.Y = e.overlandSquare().X, e.overlandSquare().Y
			w.Square = &rules.MapPos{X: e.posX, Y: e.posY} // the view's own (the dungeon's, stale here)
			return m
		}
		return game.Automap{}
	}
	m := game.Automap{Area: e.mapName(), Size: rules.Size{W: 16, H: 16}, NoMap: e.world.NoCoords != 0}
	seen, mine, built := e.ext.seen[e.areaKey()], e.ext.labels[e.areaKey()], e.builtinLabels(e.areaKey())
	secret := e.secretSides()
	for y := range 16 {
		for x := range 16 {
			at := rules.MapPos{X: x, Y: y}
			if seen != nil && seen[y<<4|x] != 0 {
				m.Known = append(m.Known, at)
			}
			for _, d := range [4]int{0, 2, 4, 6} {
				w, door := e.WallAt(x, y, d), e.DoorAt(x, y, d)
				switch {
				case w == 0:
				case secret[doorSide{X: x, Y: y, Dir: d}]: // a secret door not found yet: the wall it looks like
					m.Walls = append(m.Walls, game.Wall{At: at, Side: d, Kind: int(w)})
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
	return m
}

// creationState is the creation sheet's draft and fields.
func (e *Engine) creationState(cr *creation) *game.Draft {
	id := cr.c
	sh := e.ruleset().Sheet(id)
	d := game.Member{Name: sh.Name, Race: e.word(id, rules.FactRace), Class: e.word(id, rules.FactClasses), Alignment: e.word(id, rules.FactAlignment),
		HP: int(sh.HP), MaxHP: int(sh.MaxHP), Stats: map[string]int{}, Str00: int(sh.Scores[rules.StrengthPercent])}
	for k, n := range game.StatKeys[:6] {
		d.Stats[n] = int(sh.Scores[abilities[k]])
	}
	var sp SheetPage // the figures as the sheet's redraw works them out (combatStats), on every roll
	e.applyItems(id)
	e.combatStats(id, &sp)
	d.AC, d.THAC0, d.Move, d.Age = sp.AC, sp.THAC0, sp.Move, int(sh.Age)
	poolState(e.poolPics(id), &d)
	e.iconState(id, &d)
	cl := e.Creation()
	mins := map[string]int{}
	for a, v := range cl.Minimums(cr.d.D) {
		if k := slices.Index(abilities[:], a); k >= 0 {
			mins[statKeys[k]] = int(v)
		}
	}
	g := &game.Draft{Member: d, Minimums: mins}
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
	partyDrop: "Drop Character", partyNewClass: game.ChangeClassFunction.Label, partyQuickStart: "Quick Start",
	partyForget: "Forget Character",
}

// shopReplies are the shop's own choices when nothing is asked.
var shopReplies = []string{"Buy", "Leave"}

// shopMenu: the shop's own choices are up, nothing over them.
func (e *Engine) shopMenu() bool {
	return e.question() == nil && e.creation == nil && !e.menuUp() && e.waiting(PendShop)
}

// pendingAsk is the engine's question, the party menu or the shop's choices.
func (e *Engine) pendingAsk() *game.Question {
	if q := e.question(); q != nil {
		return q
	}
	if cr := e.creation; cr != nil { // Esc or Exit on the sheet asked this before abandoning it; the open sheet asks nothing (class 50)
		if cr.naming {
			return &game.Question{Text: "Character name:", Typed: true, Max: 0xF} // Rename cuts at 15
		}
		if !cr.asking {
			return nil
		}
		return &game.Question{Text: cr.abandon, YesNo: true, Replies: []string{"No", "Yes"}, Default: 1}
	}
	if e.menuUp() {
		q := &game.Question{Text: "Choose a function"}
		for _, k := range e.partyFunctions() { // the engine's, not the View's: the front builds its menu from these
			q.Replies = append(q.Replies, mainMenuLabels[k])
		}
		return q
	}
	if e.shopMenu() {
		return &game.Question{Replies: slices.Clone(shopReplies), Esc: game.EscTo(slices.Index(shopReplies, "Leave"))}
	}
	return nil
}

// dirOf is the direction (0 2 4 6) of a one-square step v, or -1.
func dirOf(v image.Point) int {
	for d := 0; d < 8; d += 2 {
		if v == image.Pt(rules.StepX[d], rules.StepY[d]) {
			return d
		}
	}
	return -1
}

// refusal is err as a game.Refusal.
func refusal(err error) *game.Refusal {
	if f, ok := err.(*game.Refusal); ok {
		return f
	}
	return &game.Refusal{Text: err.Error()}
}

// question is the question the engine waits on (the Ask), nil for none: a
// list's and a member pick's rows are its replies, a menu with none a Continue.
func (e *Engine) question() *game.Question {
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
	if w.shape == shPause { // DELAY and the other timed pauses: the presenter paces them (game.Question.Ms)
		q.Ms = w.ms
	}
	return &q
}

// labelKinds name the label kinds (game.Label.Kind); the player's are "".
var labelKinds = [...]string{adventure.PoiUser: "", adventure.PoiShop: "shop", adventure.PoiTemple: "temple", adventure.PoiInn: "inn", adventure.PoiCivic: "civic"}

// spellGroup is the level group a spell of class cls and level lvl is
// listed under (SpellRef.Group, the SlotsMax keys): "cleric 1", "mage 2".
func spellGroup(cls rules.SpellClass, lvl byte) string {
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
	if p.Alt != nil {
		if a := gamePicture(*p.Alt); a.Pixels != nil {
			g.Alt = &a
		}
	}
	return g
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
func (e *Engine) iconState(id rules.CharID, m *game.Member) {
	ic := e.fact(id, rules.IconPicsFact{}).(rules.IconPics)
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
	return image.Pt(int(ViewModeShift[e.viewMode])+int(e.world.MapX), int(e.world.MapY))
}
