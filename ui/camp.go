package ui

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"time"
	"unicode"

	"gobox/asset"
	"gobox/game"
	"gobox/gpu"
	"gobox/ui/input"
)

// Camp (Camp): the presenter's own screens over camp's state, each
// choice a decision for the selected member.

// Camp screens.
const (
	campMain     = iota
	campMemorize // the spell levels
	campLevel    // one level's spells to memorize
	campScribe
	campDrop // the drop's confirmation

	templeHeal // the temple's services (temple.go)
	templePay  // pay for the cure?
	appraise   // the gems and jewelry (temple and shop)
	shopID     // the items to identify (shop.go)
	shopDeal   // identify this one?
)

type campUI struct {
	screen   int
	sel      int
	top      int            // the first line shown of a list too long for the screen
	pointer  image.Point    // the pointer on the window (Hover), for the wheel
	wheeled  time.Time      // the last line the wheel moved a list
	pick     int            // the temple's cure or the shop's item asked about
	rowsWere []string       // the rows' texts as last shown (follow)
	rowsOf   int            // ... and their screen
	ids      []int          // the spell lists' lines: each one's spell (0: a level, Exit)
	unset    []bool         // the spell lists' lines with none memorized (grey)
	level    string         // the level campLevel lists; the power pick's member and pick (powerItems)
	dropAsk  int            // the drop's question asked (Data.Confirms["DropMember"].Asks)
	rows     map[int][2]int // each screen left for a submenu: its highlighted row and first line shown, for back
	idSeen   game.ItemRef   // the shop's Id reveal closed last (identify.go) ...
	idShut   bool           // ... while this holds
}

// to opens screen n (a submenu, or a question on row sel), keeping the
// row that opened it for back.
func (c *campUI) to(n, sel int) func() {
	return func() {
		if c.rows == nil {
			c.rows = map[int][2]int{}
		}
		c.rows[c.screen] = [2]int{c.sel, c.top}
		c.screen, c.sel, c.top = n, sel, 0
	}
}

// back returns to screen n on the row that opened the submenu left.
func (c *campUI) back(n int) func() {
	return func() { c.screen, c.sel, c.top = n, c.rows[n][0], c.rows[n][1] }
}

// follow keeps the highlight on its row when rows come or go above it
// (the shop's Take once credits are pooled): the same text, not the same
// index.
func (c *campUI) follow(items []MenuRow) {
	if c.rowsOf == c.screen && c.sel >= 0 && c.sel < len(c.rowsWere) && (c.sel >= len(items) || items[c.sel].Text != c.rowsWere[c.sel]) {
		if i := slices.IndexFunc(items, func(r MenuRow) bool { return r.Text == c.rowsWere[c.sel] }); i >= 0 {
			c.sel = i
		}
	}
	c.rowsWere, c.rowsOf = c.rowsWere[:0], c.screen
	for _, r := range items {
		c.rowsWere = append(c.rowsWere, r.Text)
	}
}

// campItems are the lines of the camp screen up, and its title.
// The temple, the shop and training share the camp's screens.
func (p *Presenter) campItems() (string, []MenuRow) {
	switch {
	case p.templeUp():
		return p.templeItems()
	case p.shopUp():
		return p.shopItems()
	case p.trainingUp() != nil:
		return p.trainItems()
	case p.powerPick() != nil:
		return p.powerItems()
	}
	c, s := &p.camp, p.state()
	party, sel := partyOf(s), p.selected()
	var m game.Member // the selected member: whose spells the lists show
	if sel >= 0 && sel < len(party) {
		m = party[sel]
	}
	book := campBook(m)
	heals, studies := restOf(s)
	noRest := allowedOf(s)["NoRest"] // camp with no AdvanceTimeBy
	act := func(k ActKind, n int) func() { return func() { p.send(k, n) } }
	switch c.screen {
	case campMemorize: // the levels, each its own submenu
		var it []MenuRow
		c.ids, c.unset = c.ids[:0], c.unset[:0]
		for i, sp := range book {
			if level, planned, slots := bookLine(m, book, sp); i == 0 || levelName(book[i-1].Level) != level {
				it, c.ids = append(it, MenuRow{Text: fmt.Sprintf("%-26s %d/%d", level, planned, slots), Key: 0, Do: func() { c.level = level; c.to(campLevel, 0)() }}), append(c.ids, 0)
			}
		}
		return "Spells to memorize", append(it, MenuRow{Text: "Exit", Key: 0, Do: c.back(campMain)})
	case campLevel, campScribe:
		var list []game.SpellRef // the scribe list is empty: the state names no scrolls' spells
		title := "Spells to scribe"
		if c.screen == campLevel {
			title = c.level
			for _, sp := range book {
				if level, planned, slots := bookLine(m, book, sp); level == c.level {
					list, title = append(list, sp), fmt.Sprintf("%c%s%c %d/%d", InkHead, level, InkHead, planned, slots)
				}
			}
		}
		var it []MenuRow
		c.ids, c.unset = c.ids[:0], c.unset[:0]
		for _, sp := range list {
			text := fmt.Sprintf("%-26s x%d", sp.Name, sp.Planned)
			if c.screen == campScribe {
				text = sp.Name
				if sp.Planned > 0 {
					text += " *"
				}
			}
			it, c.ids, c.unset = append(it, MenuRow{Text: text, Key: 0, Do: func() { p.campSpell(sp.ID, 1) }}), append(c.ids, sp.ID), append(c.unset, c.screen == campLevel && sp.Planned == 0)
		}
		back := c.back(campMain)
		if c.screen == campLevel {
			back = c.back(campMemorize)
		}
		return title, append(it, MenuRow{Text: "Exit", Key: 0, Do: back}) // no hotkeys: the lines are spells
	case campDrop:
		cf := confirmOf("DropMember", game.Confirm{Asks: []string{"Drop %s from the party?"}})
		n := min(c.dropAsk, len(cf.Asks)-1)
		yes := func() { // the next question, else the drop
			if c.dropAsk = n + 1; c.dropAsk < len(cf.Asks) {
				c.sel = map[bool]int{false: 1}[cf.Yes]
				return
			}
			c.dropAsk = 0
			p.send(ActDropMember, p.selectedIndex())
			c.back(campMain)()
		}
		no := func() { c.dropAsk = 0; c.back(campMain)() }
		return confirmText(cf.Asks[n], p.selectedName(), ""), []MenuRow{{Text: "Yes", Key: 'Y', Do: yes}, {Text: "No", Key: 'N', Do: no}}
	}
	var it []MenuRow // the party's, the rests first: each member's own are on their card's menu (drawCampMenus)
	for _, r := range rests {
		if r.Fix && !heals && !studies || r.Duration && noRest {
			continue
		}
		text, key := r.Text, r.Key
		if !p.flags().Day && r.Night != "" {
			text, key = r.Night, r.NightKey
		}
		do := act(ActRest, r.N)
		if r.Fix { // its own moment: the fade first, the rest sent once the screen is black
			do = func() { p.startFix(r.N) }
		}
		if r.Duration {
			do = func() { p.askRest(r.N) }
		}
		it = append(it, MenuRow{Text: text, Key: key, Do: do})
	}
	it = append(it, MenuRow{Text: "Save", Key: 'S', Do: act(ActSaveGame, 0)}) // the slot screen says whether the rules let it save
	return "Camp", append(it, MenuRow{Text: "Exit", Key: 'E', Do: act(ActClose, 0)})
}

// campBook is m's spells to memorize at the next rest, in their order.
func campBook(m game.Member) []game.SpellRef {
	return slices.DeleteFunc(slices.Clone(m.Spells), func(sp game.SpellRef) bool { return sp.Memorized })
}

// bookLine is where spell sp of m's book is listed: its level's heading,
// the spells planned in its group and the group's slots.
func bookLine(m game.Member, book []game.SpellRef, sp game.SpellRef) (level string, planned, slots int) {
	for _, x := range book {
		if x.Group == sp.Group {
			planned += x.Planned
		}
	}
	return levelName(sp.Level), planned, m.SlotsMax[sp.Group]
}

// campHome: camp's own screen is up, the members' menus beside their
// cards (drawCampMenus).
func (p *Presenter) campHome() bool {
	return p.campUp() && p.camp.screen == campMain && p.screenUp()
}

// drawCampMenus draws every member's context menu at once beside their
// card; the one with the keys (ctx) shows its highlight.
func (p *Presenter) drawCampMenus(dst *gpu.Image) {
	p.campCtx = p.campCtx[:0]
	for i := range min(len(p.party()), len(p.cards)) {
		lines, acts, keys := powerLines(p.cardPowers()[i])
		sel := -1
		if p.ctx == 1+i {
			sel = p.ctxSel
		}
		p.menuGrey = p.ctx != 1+i // the focus's border red, the others' grey
		for len(p.campImg) <= i {
			p.campImg, p.campBuf = append(p.campImg, nil), append(p.campBuf, nil)
		}
		p.tipImg, p.tipBuf, p.campImg[i], p.campBuf[i] = p.campImg[i], p.campBuf[i], p.tipImg, p.tipBuf // each menu its own image: one image redrawn in a frame shows its last pixels everywhere
		shown := keys                                                                                   // the white hotkeys only on the focus's menu
		if p.ctx != 1+i {
			shown = nil
		}
		rs := p.cardMenu(dst, i, lines, shown, sel)
		p.tipImg, p.tipBuf, p.campImg[i], p.campBuf[i] = p.campImg[i], p.campBuf[i], p.tipImg, p.tipBuf
		p.menuGrey = false
		p.campCtx = append(p.campCtx, rs)
		if p.ctx == 1+i {
			p.ctxActs, p.ctxKeys, p.ctxRects = acts, keys, append(p.ctxRects[:0], rs...)
		}
	}
}

// campFocus moves camp's one focus for key k: F1-F8 to member m's menu,
// F9 to the party's menu, Left/Right (the pad's d-pad) and Tab round them
// in the screen's order: the first half of the members (the left column,
// or the row's left half), the party's menu, the rest.
func (p *Presenter) campFocus(k uint16, m int) {
	order := []int{} // ctx values: 0 the party's menu, 1+i member i's
	for i := range p.party() {
		if i == (len(p.party())+1)/2 {
			order = append(order, 0)
		}
		order = append(order, 1+i)
	}
	if len(p.party()) < 2 {
		order = append(order, 0)
	}
	n, at := len(order), slices.Index(order, p.ctx)
	switch a := actOf(k); a {
	case input.InputPartyMenu:
		p.ctx = 0
		return
	case input.InputLeft, input.PrevMember:
		p.ctx = order[(at+n-1)%n]
	case input.InputRight, input.NextMember, input.Minimap: // Minimap: Tab
		p.ctx = order[(at+1)%n]
	default:
		p.ctx = 1 + m
	}
	p.ctxSel = 0
	if p.ctx > 0 {
		p.send(ActLead, p.ctx-1)
	}
}

// campHover gives the focus to the camp window under the pointer (x, y)
// and its line the highlight: the party's menu, or a member's (who is
// then selected).
func (p *Presenter) campHover(x, y int) {
	if i := slices.IndexFunc(p.menuRects, image.Pt(x, y).In); i >= 0 {
		p.ctx, p.camp.sel = 0, p.camp.top+i
		return
	}
	for i, rs := range p.campCtx {
		if j := slices.IndexFunc(rs, image.Pt(x, y).In); j >= 0 {
			if p.ctx != 1+i {
				p.send(ActLead, i)
			}
			p.ctx, p.ctxSel = 1+i, j
			return
		}
	}
}

// memberAct carries out power k of member n's context menu: camp's
// Memorize, Scribe and Drop open the presenter's screens for them.
func (p *Presenter) memberAct(k ActKind, n int) {
	screen := map[ActKind]int{ActPlan: campMemorize, ActScribe: campScribe, ActDropMember: campDrop}
	if s, ok := screen[k]; ok {
		p.send(ActLead, n)
		p.camp.to(s, map[bool]int{true: 1}[s == campDrop && !confirmOf("DropMember", game.Confirm{}).Yes])() // the drop's question on No, unless the game starts it on Yes
		p.camp.dropAsk = 0
		return
	}
	p.send(k, n)
}

// campSpell plans one more (n > 0) or one less of spell id, or toggles
// its scribing.
func (p *Presenter) campSpell(id, n int) {
	if p.camp.screen == campScribe {
		p.inv.acts = append(p.inv.acts, Action{Kind: ActScribe, N: -1, X: id})
		return
	}
	p.inv.acts = append(p.inv.acts, Action{Kind: ActPlan, N: -1, X: id, Y: n})
}

// drawCamp draws the camp screen up at scale k, centred on cx, its top at
// top; a bestiary entry's icon goes over its box.
func (p *Presenter) drawCamp(dst *gpu.Image, glyph func(byte) [8]byte, k float64, cx, top, bottom int) {
	title, items := p.campItems()
	p.camp.follow(items)
	lines, keys, off := make([]string, len(items)), p.campKeys(items), RowsOff(items)
	for i, it := range items {
		lines[i] = it.Text
		if i < len(p.camp.unset) && p.camp.unset[i] && p.camp.screen == campLevel { // none memorized: grey, still picked
			off[i] = true
		}
	}
	p.camp.sel = min(max(p.camp.sel, 0), len(items)-1)
	var icons []image.Image // the rows' icons: the items' (the shop's Identify)
	switch {
	case p.shopUp() && p.camp.screen == shopID:
		s := p.state()
		for _, u := range p.eng.r.derived().unknown {
			it := itemOf(partyOf(s), u)
			icons = append(icons, p.inv.iconImage(it.Base))
		}
	}
	if icons != nil {
		icons = append(icons, make([]image.Image, max(0, len(lines)-len(icons)))...)[:len(lines)]
	}
	c := &p.camp
	lh := 1 // a line's cells
	if len(icons) > 0 {
		lh = iconRow
	}
	extra := 2 // the border; the title and a blank line under it
	if title != "" {
		extra += len(TitleLines(title)) + 1
	}
	fit := max(3, (int(float64(bottom-top)/(k*8))-extra)/lh) // the lines that fit under top (never over the text box above): more scroll, the scale stays
	n := len(lines)
	c.top = min(max(c.top, c.sel-fit+1), c.sel, max(0, n-fit))
	if n > fit {
		end := c.top + fit
		lines, keys, off = lines[c.top:end], keys[c.top:end], off[c.top:end]
		if icons != nil {
			icons = icons[c.top:end]
		}
	}
	rows := extra + lh*len(lines)
	top = max(0, min(top, bottom-int(k*float64(rows*8)))) // moved up to fit above bottom
	if yesNo(keys) {                                      // a question: where every one goes (keyWindow)
		p.keyWindow(dst, title, lines, keys, off, c.sel, false)
		p.menuRects = slices.Clone(p.winRects)
		return
	}
	sel := c.sel - c.top
	if p.campHome() && p.ctx > 0 { // a member's menu has the focus: no highlight, no white hotkeys here
		sel, p.menuGrey, keys = -1, true, nil
	}
	p.drawMenuIcons(dst, glyph, title, lines, keys, off, icons, sel, k, cx, top)
	p.menuGrey = false
	p.menuBar(dst, p.menuRects, n, len(lines), &c.top, &c.sel, k)
	if (p.spellList() || len(p.learning()) > 0) && c.sel < len(c.ids) && len(p.menuRects) > 0 {
		p.drawTip(dst, mapTips(c.ids, spellTip), c.sel, k, cx)
	}
	if pp := p.powerPick(); pp != nil && c.sel < len(pp.Options) && len(p.menuRects) > 0 {
		p.drawTip(dst, mapTips(pp.Options, func(o game.PowerOption) []string { return WrapWords(o.Tip, 36) }), c.sel, k, cx)
	}
}

// spellTip is spell id's rules as the tip's lines (nil: none).
func spellTip(id int) []string {
	if id > 0 && id < len(spellRules) {
		return WrapWords(spellRules[id], 36)
	}
	return nil
}

// selectOutline marks picture r: fill paints a frame of light green t
// pixels thick along the inside of r.
func selectOutline(r image.Rectangle, t int, fill func(image.Rectangle, color.RGBA)) {
	c := asset.EGARGBA(10)
	fill(image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+t), c)
	fill(image.Rect(r.Min.X, r.Max.Y-t, r.Max.X, r.Max.Y), c)
	fill(image.Rect(r.Min.X, r.Min.Y, r.Min.X+t, r.Max.Y), c)
	fill(image.Rect(r.Max.X-t, r.Min.Y, r.Max.X, r.Max.Y), c)
}

// mapTips is the tip of each of a list's lines.
func mapTips[T any](l []T, tip func(T) []string) [][]string {
	tips := make([][]string, len(l))
	for i, x := range l {
		tips[i] = tip(x)
	}
	return tips
}

// drawTip draws tips[sel] in the spell tip's box under the menu, or over
// it when the list's longest tip would not fit under it: the box keeps
// its side as the selection moves.
func (p *Presenter) drawTip(dst *gpu.Image, tips [][]string, sel int, k float64, cx int) {
	if sel < 0 || sel >= len(tips) || tips[sel] == nil {
		return
	}
	lines, rows := tips[sel], 0
	for _, t := range tips {
		rows = max(rows, len(t))
	}
	box := p.menuRects[0]
	for _, r := range p.menuRects {
		box = box.Union(r)
	}
	h := int(k) * 8 * (len(lines) + 2)
	top := box.Max.Y + int(k)*16
	if top+int(k)*8*(rows+2) > p.listFloor(dst) {
		top = max(0, box.Min.Y-int(k)*32-h)
	}
	p.tipImg, p.tipBuf, p.spellTipImg, p.spellTipBuf = p.spellTipImg, p.spellTipBuf, p.tipImg, p.tipBuf
	p.tipMenu(dst, "", lines, make([]byte, len(lines)), -1, k, cx, top)
	p.tipImg, p.tipBuf, p.spellTipImg, p.spellTipBuf = p.spellTipImg, p.spellTipBuf, p.tipImg, p.tipBuf
}

// spellList: a camp spell list is up (its rules drawn by it, the picture
// hidden under them).
func (p *Presenter) spellList() bool {
	s := p.camp.screen
	return p.campUp() && (s == campMemorize || s == campLevel || s == campScribe)
}

// campKey takes key k on a camp screen.
func (p *Presenter) campKey(k uint16) bool {
	c := &p.camp
	_, items := p.campItems()
	c.follow(items)
	n := len(items)
	spells := p.spellList()
	m, member := memberKey(k, p.selectedIndex(), len(p.party()), true)
	switch a := actOf(k); {
	case spells && (a == input.Less || a == input.Backspace) && c.sel < n-1: // -, Backspace: one less
		p.campSpellAt(c.sel, -1)
	case spells && a == input.More && c.sel < n-1: // +: one more
		p.campSpellAt(c.sel, 1)
	case actOf(k) == input.InputPartyMenu:
	case (member || actOf(k) == input.Minimap) && p.campHome(): // another menu takes the keys
		p.campFocus(k, m)
	case member && !yesNo(p.campKeys(items)): // the member camp acts for (not under a question about one)
		p.send(ActLead, m)
	default:
		switch i, back, ok := MenuNav(actOf(k), k, &c.sel, p.campKeys(items)); {
		case back:
			p.campBack()
		case i >= 0:
			p.campPick(i)
		case !ok && p.campHome() && p.cardKey(k):
		case !ok:
			return !globalKey(k) // help, the minimap, save/load ... go on (keyAct); other keys would act on the view under the screen
		}
	}
	return true
}

// cardKey acts the selected member's card line whose hotkey is k (camp:
// a card's verbs answer their keys without Tab to its menu first).
func (p *Presenter) cardKey(k uint16) bool {
	n := p.selectedIndex()
	if n < 0 || n >= len(p.party()) || k >= 128 {
		return false
	}
	_, acts, keys := powerLines(p.cardPowers()[n])
	if j := slices.Index(keys, byte(unicode.ToUpper(rune(k)))); j >= 0 && keys[j] != 0 {
		p.memberAct(acts[j], n)
		return true
	}
	return false
}

// campBack goes back one level: the screen's Exit or No row, else the
// screen closes. It never picks a list's line.
func (p *Presenter) campBack() {
	_, items := p.campItems()
	if n := len(items); n > 0 && slices.Contains([]string{"Exit", "No"}, items[n-1].Text) {
		p.campPick(n - 1)
		return
	}
	p.send(ActClose, 0)
}

// campKeys are the camp screen's hotkeys: none on the spell lists (the
// digits would land in the counts), the rest as every menu.
func (p *Presenter) campKeys(items []MenuRow) []byte {
	if p.spellList() || len(p.learning()) > 0 || p.powerPick() != nil { // the lines are names: no hotkeys
		return make([]byte, len(items))
	}
	return RowKeys(items)
}

// campSpellAt changes the plan of the spell on line i.
func (p *Presenter) campSpellAt(i, n int) {
	if c := &p.camp; i < len(c.ids) && c.ids[i] != 0 {
		p.campSpell(c.ids[i], n)
	}
}

// campPick picks line i of the camp screen up.
func (p *Presenter) campPick(i int) {
	if _, items := p.campItems(); i >= 0 && i < len(items) && items[i].Do != nil { // a heading picks nothing
		p.camp.sel = i
		items[i].Do()
	}
}

// screenUp: one of the camp-style screens is up (camp, temple, shop, training, a power pick).
func (p *Presenter) screenUp() bool {
	return (p.campUp() || p.templeUp() || p.shopUp() || p.trainingUp() != nil || p.powerPick() != nil) && p.askMenu() == nil && p.sheetPage() == nil // a member viewed: the sheet over the screen
}

// sendX sends a decision with an X.
func (p *Presenter) sendX(k ActKind, n, x int) { p.sendA(Action{Kind: k, N: n, X: x}) }

// appraiseItems is the appraise screen; back goes to the screen before.
func (p *Presenter) appraiseItems(purse map[string]int, offers []game.Offer, back func()) (string, []MenuRow) {
	if len(offers) > 0 && offers[0].Value > 0 {
		o := offers[0]
		what := "Gem"
		if o.Kind == game.Jewelry {
			what = "Jewel"
		}
		var it []MenuRow
		if o.CanKeep {
			it = append(it, MenuRow{Text: "Keep", Key: 'K', Do: func() { p.call(func(e *classicGame) { e.KeepAppraised(true) }) }})
		}
		sell := func() { p.call(func(e *classicGame) { e.KeepAppraised(false) }) }
		return fmt.Sprintf("The %s is valued at %s. Leaving sells it.", what, price(Money{Gold: o.Value})), append(it, MenuRow{Text: "Sell", Key: 'S', Do: sell},
			MenuRow{Text: "Exit", Key: 'E', Do: sell}) // back to the list: the original's leaving sells (Esc too, campBack), the shop stays
	}
	var it []MenuRow
	if n := purse[game.Gems]; n > 0 {
		it = append(it, MenuRow{Text: fmt.Sprintf("Gems (%d)", n), Key: 'G', Do: func() { p.call(func(e *classicGame) { e.Appraise(false) }) }})
	}
	if n := purse[game.Jewelry]; n > 0 {
		it = append(it, MenuRow{Text: fmt.Sprintf("Jewelry (%d)", n), Key: 'J', Do: func() { p.call(func(e *classicGame) { e.Appraise(true) }) }})
	}
	title := "You have a fine collection"
	if len(it) == 0 {
		title = "Nothing to appraise"
	}
	return title, append(it, MenuRow{Text: "Exit", Key: 'E', Do: back})
}

// askRest asks how long rest N goes on, as the original's one "Rest Time:"
// picker: Days, Hours or Mins picks the field (Mins first, by 5), Add and
// Subtract (Up, Down) change it, Rest rests so long (ActRest N, X minutes),
// Exit gives up.
func (p *Presenter) askRest(n int) {
	t, field := [3]int{}, 2 // days, hours, minutes
	most, step := [3]int{99, 23, 55}, [3]int{1, 1, 5}
	q := &Question{Ask: FrontAsk{Kind: AskMenu, Menu: Menu{Commands: restItems}, Default: 3}}
	set := func() {
		q.Ask.Prompt = fmt.Sprintf("Rest Time: %d Days %d Hours %d Mins", t[0], t[1], t[2])
	}
	set()
	q.Answer = func(a Action) bool {
		switch {
		case a.Kind == ActChoose && a.N < 3:
			field = a.N
		case a.Kind == ActPage: // Up (-1) adds, Down subtracts
			t[field] = min(most[field], max(0, t[field]-step[field]*a.N))
			set()
			return false
		case a.Kind == ActChoose && a.N < 5:
			t[field] = min(most[field], max(0, t[field]+step[field]*(7-2*a.N)))
		case a.Kind == ActChoose && a.N == 5, a.Kind == ActOK:
			p.inv.acts = append(p.inv.acts, Action{Kind: ActRest, N: n, X: t[0]*1440 + t[1]*60 + t[2]})
			return true
		case a.Kind == ActChoose, a.Kind == ActClose:
			return true
		default:
			return false
		}
		q.Ask.Default = a.N
		set()
		return false
	}
	p.Ask(q)
}

// restItems are the rest picker's options.
var restItems = []Command{{Text: "Days", Key: 'D'}, {Text: "Hours", Key: 'H'}, {Text: "Mins", Key: 'M'},
	{Text: "Add", Key: 'A'}, {Text: "Subtract", Key: 'S'}, {Text: "Rest", Key: 'R'}, {Text: "Exit", Key: 'E'}}
