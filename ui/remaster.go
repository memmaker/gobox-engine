// remaster.go shows the game with the Blobber renderer: the 3D view
// fills the window, and the original's windows are laid over it. It only
// reads the game, through play.State and the UI, so drawing never touches the
// game's state.

package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"gobox/asset"
	"gobox/audio"
	"gobox/ega"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// Presenter draws the remastered screen. Screens it has no remaster for
// (no remastered map, sheets, shops, the title) it draws as the original
// laid them out, from the View's page (page.go).
type Presenter struct {
	pg page // the original's screen, drawn from the View

	calls     []func(*classicGame) // rule calls not yet run (Calls)
	overview  bool                 // the party overview is up (any key or click closes it)
	pager     *pager               // the help or a Read is up (pager.go)
	eng       *classicGame         // for queries (SetEngine)
	pool      *poolPick            // gobox's own pool picker is up (poolpick.go)
	fonts     *asset.Chooser       // the text fonts (fontmenu.go; nil: the game's own)
	fontNote  string               // the font's name, shown for a moment after Ctrl-F
	fontAt    time.Time            // ... since then
	fontLabel string               // the game's own font's name in the list

	pauseFrom time.Time // a script's pause (AskPause) began; zero: none up

	view        *View
	scenes      map[string]*Scene                  // parsed maps by name
	shown       string                             // the scene the view shows
	last        time.Time                          // when the animation last advanced
	ticks       float64                            // animation time not yet run, in 1/60 s ticks
	border      memo[[Places]image.Image, borders] // the game's window border
	bars        *gpu.Image                         // bars.png
	idImg       *gpu.Image                         // its pop-up
	expShown    bool                               // drawParty: every card's experience bar, Ctrl or not (the training hall)
	cardImgs    []*gpu.Image                       // the cards, drawn at scale 1...
	cardKeys    []cardKey                          // ...and what each shows
	faceImgs    map[image.Image]*gpu.Image         // the faces on the GPU
	hiFaces     map[image.Image]*gpu.Image         // the PNG portraits at the cards' window resolution
	labels      map[string]*gpu.Image              // card labels in the game's font
	glyph       func(byte) [8]byte                 // the game's font
	main        bool                               // the main menu is up (mainmenu.go)
	mm          mainUI                             // ... its state
	ask         bool                               // ... with a question over it
	gridTries   int                                // the ActResize sent for this grid
	grid        [2]int                             // the text cells the window holds, as last told to the engine (ActResize)
	gridAge     int                                // frames since
	cards       []image.Rectangle                  // the party's cards on screen in the 3D view, as last drawn (none elsewhere)
	faces       map[image.Image]image.Image        // card art cropped from each paper-doll head
	inside      bool                               // the view has the inside ceiling
	active      bool                               // the last frame was drawn remastered
	menu        bool                               // the vertical menu is open
	menuFor     string                             // the menu it opened for (menuID): another menu arriving closes it, so one key never answers two menus
	sel         int                                // its highlighted item
	asked       string                             // the question menu last opened by itself (menuID); send clears it on a pick, so the same menu back after one (a refusal, back from Browse) opens again
	pageAsked   string                             // the page question menu last opened (drawPageMenu)
	answered    string                             // the text pane under the last question: once it is answered, that text is done
	mapImg      *gpu.Image                         // Map on the GPU
	mapAt       image.Rectangle                    // where drawMap last put the minimap (a click there opens the full map)
	mapSrc      *MapSquares                   // the derived minimap it holds (derived)
	mapInk      [16]color.RGBA                     // the ink it was drawn in
	mapFont     int                                // and the font (fonts.Cur, -1: the game's own)
	menuRects   []image.Rectangle                  // the vertical menu's items on screen, as last drawn
	scrolls     Bars                               // the scroll bars drawn this frame (scrollbar.go) and the drag
	logTop      int                                // the battle Log's first line shown
	sheet       sheetCanvas                        // the character sheet
	journal     journalUI                          // the quest journal (journal.go)
	resultsPage sheetCanvas                        // the after-combat page (drawResults)
	area        AreaMapPainter                          // the overland's map and marker last uploaded (drawAreaMap)
	areaPic     areaPic                            // the overland's picture built from its squares (drawAreaMap)
	fightAskAt  []image.Rectangle                  // the fight pop-up menu's lines on screen
	winRects    []image.Rectangle                  // the Esc menu's lines on screen, as last drawn
	noMenu      bool                               // nothing to pick: any click answers
	pointer     image.Point                        // the mouse, as Hover last saw it (keys move the highlight while it rests)
	mc          MenuCanvas                         // the vertical menu's images (ui.MenuCanvas.Draw)
	battle      CombatPresenter                    // the combat map in 3D
	mapBottom   int                                // screen rows above this are the 3D map
	fight       bool                               // the last frame was the 3D combat map
	inv         inventory                          // the inventory screen (inventory.go)
	fightLists  lists                              // the turn's weapons, spells and maneuvers (drawnBoard)
	board       *play.Combat                       // the fight this frame shows: the live one, or the one a replayed moment holds (nil: none)
	boardParty  []game.Member                      // the party's figures on board
	noCards     bool                               // a fight's moment plays: the cards are not drawn
	fightView   *play.Combat                       // board with the fallen kept until their fall plays (keepFallen)
	fightBoard  *Board                             // fightView as the tabletop draws it, aims put in
	fly         *flight                            // a shot's sprite over board (a replayed missile frame)
	lastBoard   *play.Combat                       // the fight before the moments now playing (keepFallen)
	inkAsk      *FrontAsk                          // the question whose menu is being drawn: its colours (asking)
	replay      replay                             // the fight's events still to play
	logGrey     bool                               // drawMenu's plain text light grey, not green (the battle Log)
	fightLog    bool                               // the battle Log is up (fight.go)
	blowLine    string                             // the last line a blow's replay said, shown on the pane a while
	blowAt      time.Time
	fightCard   image.Rectangle           // the active member's card (fight.go): a click the sheet
	rowsImg     *gpu.Image                // the combat rows, drawn at scale 1
	rowsKey     string                    // what they show
	typing      bool                      // a line is being typed (input.go) ...
	typed       string                    // ... and reads this ...
	typedCur    int                       // ... its cursor here
	cardHover   int                       // 1 + the card the pointer is on (0 none): its name and class show beside it
	dragFrom    int                       // 1 + the card pressed while the party is being ordered (Ordering), it follows the pointer to the card it is dropped on (0 none)
	dragAt      image.Point               // where the pointer is while a card is dragged
	Sound       func(n int, event string) // plays a fight's EvSound n (the sound pack's event) when its turn comes (Engine.PlayEventSound)
	Snd         audio.Sound               // the sound driver EvSound and EvSong play through when Sound is nil (NewFront)
	AdLib       *audio.AdLib              // Data.AdLib's player: EvSong and EvSound go to it first (NewFront)
	dragStart   image.Point               // where it was pressed
	dragged     bool                      // it has moved far enough from dragStart to be a drag
	ordering    bool                      // Ordering
	tipping     bool                      // tipMenu is drawing: drawMenuIcons keeps the box on the window
	ctx         int                       // 1 + the card whose context menu a right click opened (0 none): drawCardTip
	ctxSel      int                       // its highlighted line
	ctxActs     []ActKind                 // its lines' actions, for the member
	behind      *Scene                    // the last 3D scene shown
	ctxKeys     []byte                    // its lines' hotkeys (the first letters)
	campCtx     [][]image.Rectangle       // in camp every member's context menu lines (drawCampMenus)
	campImg     []*gpu.Image              // drawCampMenus: each member's menu image
	campBuf     [][]byte                  // and its pixels
	menuGrey    bool                      // drawMenu's border grey: a menu without the focus
	menuTight   bool                      // drawMenu without the blank column inside the border
	ctxRects    []image.Rectangle         // its lines on screen, as last drawn
	ctxBox      image.Rectangle           // where its box went: the spell pick opens there
	pick        []Spell                   // the Ask (AskSpell) as last drawn: a spell to cast is being picked on this screen
	picBottom   int                       // the bottom of the picture box drawn this frame (0: none): menus go under it
	labelAt     []image.Rectangle         // the top labels drawn this frame (Sweeping, <act>)
	pickAt      image.Rectangle           // the context menu's box the pick list replaces (zero: the top middle, the bar's scale)
	pickSel     int                       // the pick list's highlight ...
	picking     bool                      // ... reset when it opens
	command     bool                      // the party explores (Exploring), as last drawn
	cmds        []command                 // the command menu's commands (explore.go), as last drawn
	camp        campUI                    // the camp screens (camp.go)
	screenWas   bool                      // a camp-style screen was up when last drawn: one opening starts on its main screen
	backAt      still                     // the location as last shown: what a menu screen draws over (backdrop.go)
	back        memo[uint64, *Presenter]  // draws backAt, keyed on its Version
	backdrop    bool                      // this presenter draws a backdrop: no question, no text
	listWas     bool                      // a list was asked when last drawn (openList)
	played      *frame                    // the moment Draw plays (nil: none)
	list        listUI                    // the Ask (AskList)'s highlight (list.go)
	pageMenu    bool                      // a page's question is up (list.go)
	mapPointer  image.Point               // the full map: the pointer's last cell (a move there moves the cursor)
	fullMap     *fullMapUI                // the full map up (fullmap.go)
	resets      int                       // the engine's Resets when own last looked
	lay         layouts                   // the page screens laid out from the state (layout.go)
	gridPick    *PickUI                   // a grid picker up (picker.go)
	creating    bool                      // the creation sheet up (creation.go)
	training    bool                      // the training hall up over the party menu (shop.go serviceAct)
	icon        *IconUI                   // the icon editor up
	ran         int                       // the picker's and editor's calls run: their pages are laid out again
	slots       *SlotsUI                  // the save slots up (slots.go)
	intro       IntroUI                   // the title being played (title.go)
	seqSong     string                    // the song the animation loaded last (vace cue "load")
	seq         SequenceUI                // the full-screen animation being played (Sequence)
	seqImg      srcImage                  // its frame on the GPU
	clue        *clueUI                   // the clue book's chapter list up (cluebook.go)
	clueRead    *Read                     // the clue book last read, at the page left
	clueEnd     bool                      // ... read to its end: B gives the list
	clueSel     int                       // the chapter last read
	tipImg      *gpu.Image                // the card tip's box, drawn at scale 1 (drawMenu's image, swapped in)
	tipBuf      []byte                    // its pixels
	spellTipImg *gpu.Image                // the camp spell lists' rules tooltip (drawSpellTip), its own image
	spellTipBuf []byte                    // its pixels
	scr         screenUI                  // the page screens' questions (screens.go)
	screen      image.Rectangle           // the window, as last drawn
	confirmFor  string                    // the yes/no question (the Ask, AskConfirm) up, as last drawn ...
	paneRect    image.Rectangle           // the page's text pane as drawn this frame: card menus keep clear of it
	menuFloor   int                       // the page's text pane's top: a bottom menu stops above it (drawGeneric; 0 none)
	confirmSel  int                       // ... its highlight (Yes 0, No 1) ...
	confirmAt   []image.Rectangle         // ... and its rows on screen
	noteBtn     image.Rectangle           // the text box's "> Notes" on the window (empty: no text box)
	q           *Question                 // the presenter's question up (pace.go)
	paneFrom    int                       // ... its first line shown ...
	blinkAt     time.Time                 // ... and when they began
	sf          sfUI                      // the ship fight's window page, blink and bar (spacefight.go)
}

// Draw shows the game on dst and returns where the whole page went
// (zero when it is not on screen: the mouse then points at nothing).
func (p *Presenter) Draw(dst *gpu.Image, note Note) Rect {
	p.sync()
	if d, ok := p.Snd.(*audio.Driver); ok && p.AdLib != nil && p.AdLib.SID() != Host.SIDMusic { // the Music card option
		d.Do(func() { p.AdLib.SetSID(Host.SIDMusic) })
	}
	if f := SequenceStep(p, &p.seq, p.sequenceUp(), false); p.sequenceUp() != nil { // the animation: full screen, over everything
		p.seqMusic(p.seq.Cues)
		dst.Fill(color.Black)
		if f != nil {
			b, bb := dst.Bounds(), f.Image.Bounds()
			k := min(float64(b.Dx())/float64(bb.Dx()), float64(b.Dy())/float64(bb.Dy())) // fit, keeping the aspect, as the title backdrop
			op := &gpu.DrawImageOptions{}
			op.GeoM.Scale(k, k)
			op.GeoM.Translate((float64(b.Dx())-float64(bb.Dx())*k)/2, (float64(b.Dy())-float64(bb.Dy())*k)/2)
			dst.DrawImage(p.seqImg.of(f.Image), op)
		}
		return Rect{}
	}
	if p.introUp() != nil && p.mm.title == nil { // the picture the main menu comes in over
		p.mm.title, p.mm.painted = layersImage(titleBackdrop), paintedLayers(titleBackdrop)
	}
	p.intro.NoCredits = len(credits) == 0
	IntroStep(p, &p.intro, p.introUp(), false)
	ev := p.playing()
	p.played = ev
	if ev != nil && ev.Kind == game.EvMove && ev.Ms > 0 { // a step's hold is how long the figure takes for it
		p.battle.StepTime = float64(ev.Ms) / 1000
	}
	live, party := p.liveFight()
	p.board, p.boardParty, p.fly, p.noCards = live, party, nil, false
	if p.replaying() { // the fight as it stood at the moment played (a page or screen played is laid, askUp, gameScreen)
		p.board, p.boardParty, p.fly, p.noCards = ev.Board, ev.Party, ev.fly, true // the cards are the live party's: none while the moment plays
	}
	p.scrolls.Reset()
	p.noteBtn = image.Rectangle{}
	p.battle.Asking = p.confirm() != ""
	if !p.pageScreen() { // none up: the next one opens fresh (F10 on the last save)
		p.scr.up = ""
	}
	r := p.draw(dst, note)
	if q := p.confirm(); q != "" { // the engine's yes/no question: over everything, starting on No
		if q != p.confirmFor {
			p.confirmFor, p.confirmSel = q, confirmStart(p.askUp())
		}
		p.keyWindow(dst, q, []string{"Yes", "No"}, yesNoKeys, nil, p.confirmSel, false)
		p.confirmAt = slices.Clone(p.winRects)
	} else {
		p.confirmFor, p.confirmAt = "", nil
	}
	if it := p.identified(); it != nil { // the shop's Id: what it revealed, over everything
		p.drawIdentified(dst, note.Glyph, it)
	}
	p.pick = p.spellPick()
	if up := p.pickList() != nil; up != p.picking { // a pick list opens on its first spell
		p.picking, p.pickSel = up, 0
	}
	if p.pick == nil {
		p.pickAt = image.Rectangle{}
	} else { // the spell pick (the combat's Cast list) over whatever screen is up
		k := scale(dst)
		cx, top := dst.Bounds().Dx()/2, int(k)*16
		if p.picBottom > 0 { // a picture up: under it, not over it, as far as the list fits
			h := int(k) * 8 * (len(p.pick) + 4) // drawMenu's box: the border, the title and a blank line
			top = max(top, min(p.picBottom+int(k)*4, dst.Bounds().Dy()-h))
		}
		if at := p.pickAt; at != (image.Rectangle{}) { // from a card's context menu: in its place at its scale (the card's), the edge next to the card kept
			k = cardScale(dst)
			w := 0
			for _, s := range p.pick {
				w = max(w, int(k)*8*(len(s.Name)+4)) // as drawMenu's box: a border and a blank column each side
			}
			cx, top = at.Min.X+w/2, at.Min.Y
			if at.Min.X > dst.Bounds().Dx()/2 { // the right column: the box grows to the left
				cx = at.Max.X - w/2
			}
			cx = min(max(cx, w/2), dst.Bounds().Dx()-w/2)
			top = max(0, min(top, p.listFloor(dst)-int(k)*8*(len(p.pick)+4))) // the row: above the cards, not down over them
		}
		p.drawPickList(dst, note.Glyph, k, cx, top)
	}
	return r
}

// draw shows the screen the View is: the page, the main menu, a tall page or the 3D view.
func (p *Presenter) draw(dst *gpu.Image, note Note) Rect {
	p.screen = dst.Bounds()
	p.picBottom = 0
	p.setup(note)
	p.command = p.exploring()
	if up := p.screenUp(); up && !p.screenWas { // camp, the temple, the shop, training start on their main screen
		p.camp.screen, p.camp.sel, p.camp.rows = campMain, 0, nil
		if t := p.trainingUp(); t != nil && len(t.Learn) == 0 && t.Why == "" { // the training question starts on No
			p.camp.sel = 1
		}
	}
	p.screenWas = p.screenUp()
	p.keepBackdrop()
	p.openList()
	p.mm.titleDrawn = false
	p.main = p.mainMenuUp() // before the pages, which return early: a stale main menu would take their keys
	p.pace()
	p.pageMenu = false
	if !p.exploring() {
		p.fullMap = nil
	}
	p.cmds = p.commands()
	if !p.active {
		p.ctx = 0
	}
	p.fight = p.board != nil && !(p.sheetUp() && p.sheetPage() != nil) // a member's sheet viewed mid-fight (View) is drawn and keyed over it, as a shop's; the battle view is the only one (the engine has no combat page); under a modal (help, F10 ...) the engine leaves Battle out
	p.noMenu = p.askMenu() == nil || len(p.askMenu().Commands) == 1    // a one-item bar (PRESS <ENTER> TO CONTINUE) has nothing to pick either
	p.resize(dst)
	p.syncRead()
	p.overview = p.overview && p.exploring()
	if p.overview || p.pager != nil || p.fullMap != nil || p.slots != nil { // the party overview, help, a Read, the full map, the slots: a page of its own over the whole window
		p.shown, p.active, p.menu, p.cards = "", false, false, nil
		r := p.drawPage(dst, note)
		p.scr.at = r
		p.scr.pw, p.scr.ph = p.pageSize()
		if p.fullMap != nil || p.slots != nil {
			p.drawPageMenu(dst, note)
		}
		return r
	}
	if p.fight {
		p.shown, p.active, p.menu = "", false, false
		p.drawnBoard()
		p.battle.Glyph, p.battle.K = note.Glyph, int(scale(dst))
		p.syncBattle()
		p.battle.Draw(dst, p.fightBoard)
		p.drawFight(dst, note)
		if p.view != nil {
			p.view.DrawMenu(dst) // the Esc menu over the battle
		}
		return Rect{}
	}
	if openAt != "" {
		p.drawOpen(dst)
		return Rect{}
	}
	if p.main && p.mainMenu().Hall { // the training hall: its menu over the 3D view
		p.active, p.menu = false, false
		if sc := p.scene(); sc != nil {
			p.draw3D(dst, sc, note)
		} else {
			dst.Fill(color.Black)
		}
		p.mm.drawMainUI(p, dst, p.mainMenu())
		return Rect{}
	}
	if p.main {
		p.shown, p.active, p.menu = "", false, false
		p.mm.drawMain(p, dst, p.mainMenu())
		return Rect{}
	}
	p.journal.open = p.journal.open && p.exploring()
	if p.journal.open { // the quest journal: a page of its own over the whole window
		p.shown, p.active, p.menu = "", true, false
		dst.Fill(color.Black)
		p.drawJournal(dst, note)
		if j := &p.journal; j.ask {
			p.keyWindow(dst, "Remove this note?", []string{"Yes", "No"}, yesNoKeys, nil, j.askSel, false)
			j.askAt = slices.Clone(p.winRects)
		}
		return Rect{}
	}
	if !p.exploring() {
		p.clue = nil
	}
	if p.clue != nil { // the clue book's chapters, over the 3D view
		p.shown, p.active, p.menu, p.cards = "", true, false, nil
		if sc := p.scene(); sc != nil {
			p.draw3D(dst, sc, note)
		} else {
			dst.Fill(color.Black)
		}
		p.drawClue(dst)
		return Rect{}
	}
	sc := p.scene()
	area := p.areaMapUp()                                                                                                                                                                                                                                                                                                                                       // the overland: its map over the whole window (also under its pop-ups), the 3D view's cards and boxes over it
	if sc == nil && !area && !p.campUp() || p.flags().Full && len(p.pictures()) > 0 && p.introUp() == nil /* a script's whole-screen picture (SotSB's big pictures): the picture, not the 3D view */ || p.sfOf() != nil || p.results() != nil || p.iconEditor(p.ink()) != nil || !p.choosing() && !p.flags().Adventure || p.sheetUp() && p.sheetPage() != nil { // the sheet: over the 3D view too (Tab from the items, View) // a member pick (Choosing) shows the view and the cards whatever page asked
		if p.screenUp() { // training from the hall's party menu: over the 3D view
			p.active, p.cards = true, nil // its lines take clicks (Click), no card under them
			switch {                      // over the location as it stands, never black
			case p.backP() != nil: // as last shown: its picture, the 3D view, the cards; the menu takes the text's place
				p.backP().draw(dst, note)
			case sc != nil:
				p.draw3D(dst, sc, note)
			case len(p.pictures()) > 0:
				p.drawGeneric(dst, note)
			default:
				dst.Fill(color.Black)
			}
			b := dst.Bounds()
			p.drawCamp(dst, note.Glyph, scale(dst), b.Dx()/2, b.Dy()/4, b.Max.Y)
			return Rect{}
		}
		p.shown, p.active, p.menu, p.cards = "", false, false, nil            // back from elsewhere: no tween from where the view was
		p.asked = ""                                                          // and a question still waiting on return (View from the loot menu) pops up again
		if sp := p.sheetPage(); p.sheetUp() && sp != nil && sp.Spells == "" { // the sheet: the presenter's, at the inventory's scale
			dst.Fill(color.Black)
			p.drawSheet(dst, note)
			return Rect{}
		}
		if p.results() != nil {
			return p.drawResults(dst, note)
		}
		if !p.ownLayout() { // nothing of the engine's page: the page's pictures and text in the remaster's layout
			return p.drawGeneric(dst, note)
		}
		r := p.drawPage(dst, note)
		p.scr.at = r
		p.scr.pw, p.scr.ph = p.pageSize()
		p.drawPageMenu(dst, note)
		return r
	}
	p.active = true
	if area {
		p.shown = ""
		p.drawAreaMap(dst)
	} else if sc != nil {
		p.draw3D(dst, sc, note)
	} else { // camp with no 3D view (the original shows none there): black under the cards, the picture and the menus
		p.shown = ""
		dst.Fill(color.Black)
	}
	p.labelAt = p.labelAt[:0]
	if p.flags().Sweeping { // the reminder: under the minimap, else in its corner
		k := scale(dst)
		const s = "Sweeping"
		w, gap := int(k*float64(8*len(s)+2)), int(k*cardGap)
		y := gap
		if p.mapAt.Max.Y > 0 {
			y = p.mapAt.Max.Y + gap/2
		}
		p.label(dst, s, dst.Bounds().Dx()-gap-w, y, k)
		p.labelAt = append(p.labelAt, image.Rect(dst.Bounds().Dx()-gap-w, y, dst.Bounds().Dx()-gap, y+int(k*10)))
	}
	if p.flags().Interact { // Explicit Interactions: the square's event waits for Space (or the command menu's Interact)
		k := scale(dst)
		const s = "<act>"
		w, gap := int(k*float64(8*len(s)+2)), int(k*cardGap)
		p.label(dst, s, (dst.Bounds().Dx()-w)/2, gap, k)
		p.labelAt = append(p.labelAt, image.Rect((dst.Bounds().Dx()-w)/2, gap, (dst.Bounds().Dx()+w)/2, gap+int(k*10)))
	}
	p.ordering = p.flags().Ordering
	if !p.ordering {
		p.dragFrom = 0
	}
	at := cardAt(dst)
	if p.dragFrom > 0 && p.dragged { // the dragged card hangs under the pointer
		at = func(i, n int) image.Rectangle {
			r := cardAt(dst)(i, n)
			if i == p.dragFrom-1 {
				r = r.Sub(r.Min).Add(p.dragAt).Sub(image.Pt(r.Dx()/2, r.Dy()/2))
			}
			return r
		}
	}
	p.drawParty(dst, p.shownParty(), p.cardSelected(), at)
	p.view.DrawMenu(dst) // over the map and cards
	p.layCards(dst)
	if p.ctx == 0 && !p.campHome() { // the tooltip (camp shows the members' menus); a context menu goes over everything (below)
		p.drawCardTip(dst)
	}
	k := scale(dst)
	b := dst.Bounds()
	if n := len(p.cards); n > 0 && Host.PartyRow { // pop-ups go over the card row
		b.Max.Y = cardAt(dst)(0, n).Min.Y
	} else if n > 0 { // pop-ups go between the card columns
		b.Min.X = cardAt(dst)(0, n).Max.X
		if n > 1 { // a lone card has no column on the right: the pop-ups run to the edge
			b.Max.X = cardAt(dst)(n-1, n).Min.X
		}
	}
	// the picture at the top, the text in the middle, an answer's menu right under it
	pic, txt := innerWindow.Framed(), paneCells(p.pane()).Framed()
	if p.window() == game.WindowText { // text all over the view window, not just the picture frame: its height places it
		pic = viewWindow.Framed()
	}
	ph, th := float64(pic.Pixels().Dy()), float64(txt.Pixels().Dy())
	pk := min(k, max(1, float64(min(int(float64(b.Dy())/2/(ph+th/2)), b.Dx()/ega.W))))
	// the text box and the menus at their sizes (Esc, Graphics), as far as they fit between the cards
	bw, bh := float64(b.Dx()), float64(b.Dy())
	tk := fitScale(k, 1, func(k float64) bool { return k*float64(txt.Pixels().Dx()) <= bw && k*th <= bh/2 })
	mk := fitScale(k, 1, func(k float64) bool { return k*28*8 <= bw && k*12*8 <= bh/2 })
	cx, cy := (b.Min.X+b.Max.X)/2, b.Dy()/2
	picBottom := 0                                                                                                               // the bottom of the picture box, when one is up
	boxUp := p.window() == game.WindowPicture && len(p.pictures()) > 0 || p.window() == game.WindowText && len(p.viewText()) > 0 // never an empty box
	if p.flags().Adventure && boxUp && !p.spellList() {                                                                          // the spell lists' rules go there; not the page a member pick came from
		var pp float64
		var py int
		pp, py, cy = picPlace(pic, pk, tk, mk, th, b)
		var inside *gpu.Image
		if p.window() == game.WindowPicture {
			inside = pictureBox(&p.pg.pic, p.pictures(), pic.Pixels().Inset(8))
		} else {
			inside = viewTextImage(&p.pg.pic, p.viewText(), pic.Pixels().Inset(8), note) // text over the whole view window (WinText)
		}
		drawBox(dst, inside, p.border.val.tiles, pic, pp, cx, py)
		picBottom = py + int(pp*ph/2)
		p.picBottom = picBottom
	}
	pane, asking := p.textPane(), p.askMenu() != nil && !p.exploring() && !p.noChoice()
	if asking {
		p.answered = pane
	}
	shown := p.flags().Adventure && p.pane() != nil && p.pane().Shown() && (asking || pane != p.answered) // back from a question with its text unchanged: the encounter is over
	ask := p.typingAsk() != nil && p.typingAsk().Prompt != ""                                             // a script's input: its question and the line in one pop-up, no text box
	shown = shown && !ask && !(p.askList() != nil && p.askMenu().Prompt != "")                            // a vertical menu's prompt is the list's own title
	shown = shown && !(p.shopUp() && p.camp.screen != campMain)                                           // the shop's submenus (identify, appraise) stand alone, the greeting gone
	p.paneRect = image.Rectangle{}
	if shown && p.exploring() && p.window() != game.WindowPicture && p.window() != game.WindowText { // the text under the command bar: along the bottom edge
		p.paneBox(dst, note, tk, cx, b.Dy()-int(tk*th/2))
		p.menuFloor = b.Dy() - int(tk*th) // the camp's member menus stop above the text
		defer func() { p.menuFloor = 0 }()
	} else if shown { // a question, or a message to confirm (Enter, a click): a box in the middle, its choices under it
		p.paneBox(dst, note, tk, cx, cy)
		picBottom = max(picBottom, cy+int(tk*th/2)) // the menus go under the text too
		if p.picBottom > 0 {
			p.picBottom = picBottom // a yes/no and the spell pick too: the bottom lines, not over the picture or its text
		}
	}
	if p.askMenu() != nil && !p.exploring() && !p.noChoice() && !p.choosing() { // a question: its choices pop up at once
		if id := menuID(p.askMenu()); id != p.asked { // not the menu just answered
			p.asked, p.menu, p.menuFor, p.sel = id, true, id, p.askDefault()
		}
	} else {
		p.asked = ""
	}
	if p.choosing() && p.askMenu() != nil { // a member pick: the question alone in the middle, the cards round it take the pick (a click, F1-F8, the arrows and Enter; Esc or a right click give up)
		p.menu = false
		p.drawMenu(dst, note.Glyph, "", []string{strings.TrimSpace(p.askMenu().Prompt)}, nil, nil, -1, mk, cx, b.Dy()/2-int(mk*12))
		p.menuRects = p.menuRects[:0]
	} else if ask { // the question over the line being typed, its cursor blinking
		p.drawTyping(dst, note, mk, cx, picBottom+int(pk*4))
	} else if p.askList() != nil && p.askMenu() != nil { // a list and its question (list.go)
		p.menu = false
		rows := len(TitleLines(p.askMenu().Prompt)) + min(len(p.askList()), listRows) + len(p.askMenu().Commands) + 4 // drawMenu's: the title, a blank, the lines, a blank, the options, the border
		top := max(b.Dy()/2-int(mk*float64(rows)*4), picBottom+int(pk*4))
		p.asking(func() { p.drawList(dst, note.Glyph, mk, cx, max(0, min(top, b.Max.Y-int(mk*float64(rows)*8)))) }) // all of it on screen, over the picture's edge if it must
	} else if p.screenUp() && p.spellPick() == nil { // camp (camp.go), the temple, the shop, training
		p.menu = false
		p.drawCamp(dst, note.Glyph, mk, cx, max(b.Dy()/4, picBottom+int(pk*4)), p.listFloor(dst))
		if p.campHome() {
			p.drawCampMenus(dst)
		}
	} else if p.exploring() && p.menu { // the command menu (explore.go)
		cmds := p.commands()
		lines, keys := make([]string, len(cmds)), make([]byte, len(cmds))
		for i, c := range cmds {
			lines[i], keys[i] = c.text, c.key
		}
		p.sel = min(p.sel, len(lines)-1)
		top := b.Dy()/2 - int(mk*float64(len(lines)+4)*4)
		if picBottom > 0 {
			top = max(top, picBottom+int(pk*4))
		}
		p.drawMenu(dst, note.Glyph, "", lines, keys, nil, p.sel, mk, cx, top)
	} else if m := p.askMenu(); m != nil && len(m.Commands) == 0 && strings.TrimSpace(m.Prompt) != "" && !shown { // a message with no choice ("You found some loot!"), no text box to hold it: the prompt alone until a key, under the picture
		p.menu = false
		top := b.Dy()/2 - int(mk*12)
		if picBottom > 0 {
			top = max(top, picBottom+int(pk*4))
		}
		p.drawMenu(dst, note.Glyph, "", []string{strings.TrimSpace(m.Prompt)}, nil, nil, -1, mk, cx, top)
	} else if p.askMenu() == nil { // "press a key" prompts: nothing to show (the original's prompt line stays off the remastered screens)
		p.menu = false
	} else if p.menu && menuID(p.askMenu()) != p.menuFor { // the menu the popup was opened for is gone (answered): the next one opens on its own key
		p.menu = false
	} else if p.menu {
		lines, keys := make([]string, len(p.askMenu().Commands)), make([]byte, len(p.askMenu().Commands))
		for i, it := range p.askMenu().Commands {
			lines[i], keys[i] = it.Text, it.Key
		}
		p.sel = min(p.sel, len(lines)-1)
		top := cy + int(tk*th/2) + int(pk*8)
		if p.exploring() || !shown { // no text box above: in the middle
			top = b.Dy()/2 - int(mk*float64(len(lines)+4)*4)
			if picBottom > 0 { // never over the picture: right under it
				top = max(top, picBottom+int(pk*4))
			}
		}
		p.asking(func() { p.drawMenu(dst, note.Glyph, p.askMenu().Prompt, lines, keys, nil, p.sel, mk, cx, top) })
	}
	if note.Show { // the toast: pageImage's bottom line (noteLine) on a black band along the very bottom of the screen, no border
		h := int(k * 8) // at K, as every glyph
		dst.SubImage(image.Rect(b.Min.X, b.Max.Y-h, b.Max.X, b.Max.Y)).(*gpu.Image).Fill(color.Black)
		pw, _ := p.pageSize()
		op := &gpu.DrawImageOptions{}
		op.GeoM.Scale(k, k)
		op.GeoM.Translate(float64(cx)-k*float64(pw)/2, float64(b.Max.Y-h))
		m := image.NewRGBA(image.Rect(0, 0, pw, 8))
		noteLine(m, note, 0)
		dst.DrawImage(upload(&p.pg.toast, m), op)
	}
	if p.ctx > 0 && !p.campHome() { // the context menu: on top, nothing dimmed
		p.drawCardTip(dst)
	}
	return Rect{}
}

// textPane is the text pane's text, to tell a pane apart from the one
// answered.
func (p *Presenter) textPane() string {
	if p.pane() == nil {
		return ""
	}
	var s strings.Builder
	for _, l := range p.pane().Lines {
		for _, r := range l {
			fmt.Fprintf(&s, "%d:%s|", r.X, r.Text)
		}
		s.WriteByte('\n')
	}
	return s.String()
}

// Flow paces a held move key (code, 0 for none) with fluid movement: it
// reports whether to send code now (ready: the game waits for a key), and
// whether the view paces held keys at all (else they repeat as typed).
func (p *Presenter) Flow(code uint16, ready bool) (send, paced bool) {
	if !p.active || p.view == nil || p.menu || p.view.OptionsOpen() {
		if p.view != nil {
			p.view.Flow(false, false)
		}
		return false, false
	}
	return p.view.Flow(code != 0, ready), p.view.Fluid()
}

// Options reports whether the renderer's options page is up; while it is,
// UpdateOptions reads the keyboard for it and the game gets no keys.
func (p *Presenter) Options() bool { return p.view != nil && p.view.OptionsOpen() }

// UpdateOptions runs the options page for one frame.
func (p *Presenter) UpdateOptions() { p.view.UpdateOptions() }

// optionsKeys: the pad's acts as the options page reads them.
var optionsKeys = map[Act]gpu.Key{Up: gpu.KeyUp, Down: gpu.KeyDown, InputLeft: gpu.KeyLeft, InputRight: gpu.KeyRight,
	PgUp: gpu.KeyPageUp, PgDn: gpu.KeyPageDown, Home: gpu.KeyHome, End: gpu.KeyEnd,
	PadA: gpu.KeyEnter, Enter: gpu.KeyEnter, Esc: gpu.KeyEscape}

// OptionsPad hands a gamepad key (as PadKeys sends it) to the options page,
// for the next UpdateOptions.
func (p *Presenter) OptionsPad(k uint16) {
	if gk, ok := optionsKeys[actOf(k)]; ok && p.view != nil {
		p.view.PadKey(gk)
	}
}

// OptionsMouse hands the pointer to the Esc menu: the line under (x, y)
// highlights, a left click picks or steps it, a right click goes back,
// the wheel steps a setting.
func (p *Presenter) OptionsMouse(x, y int, left, right bool, wheel float64) {
	line := -1
	for i, r := range p.winRects {
		if image.Pt(x, y).In(r) {
			line = i
		}
	}
	p.view.PointOptions(line, left, right, int(math.Copysign(math.Ceil(math.Abs(wheel)), wheel)))
}

// noChoice reports a question with one answer ("press enter to continue"):
// a message, not a choice.
func (p *Presenter) noChoice() bool {
	return p.askMenu() != nil && !p.exploring() && len(p.askMenu().Commands) <= 1 && !p.choosing() // a member pick's lone Select is still a choice: whom
}

// Tick advances the renderer's animation one tick (1/60 s) under a script.
func (p *Presenter) Tick() {
	frames++
	if p.view != nil {
		p.view.Update()
	}
}

// animate runs the renderer's 60 ticks a second, whatever the frame rate.
func (p *Presenter) animate() {
	if Scripted {
		return
	}
	now := time.Now()
	if !p.last.IsZero() {
		p.ticks = min(p.ticks+now.Sub(p.last).Seconds()*60, 6) // after a pause, no long catch-up
	}
	p.last = now
	for ; p.ticks >= 1; p.ticks-- {
		p.view.Update()
	}
}

// scale is K, the UI's one glyph cell size over 8 (CLAUDE.md): the largest
// whole factor the inventory's invW x invH canvas fits dst at (at least
// 1). Every glyph and ribbon tile is 8*K window pixels; a larger window
// gives more cells, not bigger ones.
func scale(dst *gpu.Image) float64 {
	return Scale(dst.Bounds())
}

// window is ui.Window: every window's frame, the ribbon border in tiles t.
func window(dst *gpu.Image, t Ribbon, g gpu.GeoM, cw, ch int, bg color.Color) image.Rectangle {
	r := Window(dst, t, g, cw, ch, bg)
	if m := edgesOf(t); m != nil { // the theme's own tiles (or their grey): its cracked edges too
		WindowEdges(dst, sizedEdges(m, cw, ch), g, cw, ch)
	}
	return r
}

// divider redraws column 0 of the window at g, ch cells high, as the
// theme's cracked divider (when t is the theme's tiles).
func divider(dst *gpu.Image, t Ribbon, g gpu.GeoM, ch int) {
	if m := edgesOf(t); m != nil {
		WindowDivider(dst, sizedEdges(m, 0, ch), g, ch)
	}
}

// edgeTiles are Theme.Border.Edges on the GPU (the menus'); edgeSets all
// of them by key, the "edge@n" tables too, for the colour ribbon and its grey.
var (
	edgeTiles Edges
	edgeSets  []edgeSet
	// edgePrefix names the box being drawn's own edges ("view.": the 3D
	// view's inner box), tried before the general ones.
	edgePrefix string
)

type edgeSet struct {
	for_  Ribbon
	edges map[string][]*gpu.Image
}

// edgesOf is the theme's edges drawn with ribbon t (nil: t is none of the theme's).
func edgesOf(t Ribbon) map[string][]*gpu.Image {
	for _, s := range edgeSets {
		if t[Top] != nil && t[Top] == s.for_[Top] {
			return s.edges
		}
	}
	return nil
}

// sizedEdges is edgeTiles with the box's own edges (edgePrefix+"edge"),
// else each edge's table for its length ("edge@n", the original's
// per-length arrays), where the theme has them.
func sizedEdges(edgeMap map[string][]*gpu.Image, cw, ch int) Edges {
	e := Edges{Top: edgeMap["top"], Bottom: edgeMap["bottom"], Left: edgeMap["left"], Right: edgeMap["right"], Divider: edgeMap["divider"]}
	for _, s := range []struct {
		seq *[]*gpu.Image
		key string
		n   int
	}{{&e.Top, "top", cw}, {&e.Bottom, "bottom", cw}, {&e.Left, "left", ch}, {&e.Right, "right", ch}, {&e.Divider, "divider", ch}} {
		if t := edgeMap[edgePrefix+s.key]; edgePrefix != "" && len(t) > 0 {
			*s.seq = t
		} else if t := edgeMap[s.key+"@"+strconv.Itoa(s.n)]; len(t) > 0 {
			*s.seq = t
		}
	}
	return e
}

// resize tells the engine the text cells the window holds (grid) when
// they change, and again now and then while the page is not yet of that
// size (the action was dropped): the engine lays its screens out to them.
func (p *Presenter) resize(dst *gpu.Image) {
	cols, rows := grid(dst)
	pw, ph := p.pageSize()
	p.gridAge++
	g := [2]int{cols, rows}
	if g != p.grid {
		p.gridTries = 0
	}
	if g != p.grid || p.gridAge > 30 && p.gridTries < 3 && (pw != cols*8 || ph != rows*8) { // a game that never lays out to it (it does not apply ActResize) is asked three times, not for ever
		p.grid, p.gridAge = g, 0
		p.gridTries++
		p.inv.acts = append(p.inv.acts, Action{Kind: ActResize, X: cols, Y: rows})
	}
}

// Cell is the original's text cell under screen pixel (x, y) on a
// remastered screen, which shows none of its rows: "press a key" takes a
// click anywhere, as on the original's screen.
func (p *Presenter) Cell(x, y int) (int, int, bool) { return 0, 0, p.noMenu }

// pace ends a script's pause (AskPause) once its time is up: the engine
// never sleeps, the presenter keeps the time.
func (p *Presenter) pace() {
	if p.askUp() == nil || p.askUp().Kind != AskPause {
		p.pauseFrom = time.Time{}
		return
	}
	switch {
	case p.pauseFrom.IsZero():
		p.pauseFrom = now()
	case now().Sub(p.pauseFrom) >= time.Duration(p.askUp().Ms)*time.Millisecond:
		p.send(ActOK, 0)
		p.pauseFrom = time.Time{}
	}
}

// srcImage is a picture on the GPU, uploaded again only when its source changes.
type srcImage struct {
	img *gpu.Image
	src image.Image
}

// of is src on the GPU.
func (s *srcImage) of(src image.Image) *gpu.Image {
	if src != s.src {
		if s.img != nil {
			s.img.Deallocate()
		}
		s.img, s.src = gpu.NewImageFromImage(src), src
	}
	return s.img
}

// draw3D draws the 3D view of the area and its minimap.
func (p *Presenter) draw3D(dst *gpu.Image, sc *Scene, note Note) {
	where := p.where()
	cam := cameraOf(where)
	area, _ := strconv.Atoi(where.Area)
	name := areaMaps[AreaID(area)]
	w := p.walls()
	if w != nil {
		name = p.sceneName()
	}
	if name != p.shown {
		p.view.Show(sc, cam)
		p.shown, p.inside = name, !p.flags().Inside // a new scene: set its flats below, inside or out
	} else {
		p.view.Follow(cam)
	}
	if p.flags().Inside != p.inside { // indoors the original shows the inside sky: here, a ceiling
		p.inside = p.flags().Inside
		if p.inside {
			SetFlats(sc.Floor, "")
		} else {
			SetFlats(sc.Floor, sc.Ceiling)
		}
	}
	p.view.Night = !p.flags().Day
	if w != nil { // the rules' own map: night only outside, the sun and the mountains by day outside
		day := p.flags().Day && !p.flags().Inside
		p.view.Night = !p.flags().Inside && !p.flags().Day
		p.view.SunUp, p.view.SunAz, p.view.SunEl, p.view.Mountains = w.Sun && day, w.SunAz, w.SunEl, w.Mountains && day
	}
	p.animate()
	p.view.Draw(dst)
	p.drawMap(dst, p.ink(), note.Font)
}

// ownLayout: v is a page whose screen lays itself out on the page canvas
// (the special screens, the pagers and maps, the after-combat page, the
// credits); every other page is drawGeneric's.
func (p *Presenter) ownLayout() bool {
	return p.pageScreen() || p.sfOf() != nil || p.modal() || p.sheetPage() != nil || p.creation() != nil || p.flags().Credits
}

// seqMusic plays an animation's music cues on the AdLib player's songs
// (AdLibData.Named): load picks the song, play starts it, stop ends it.
func (p *Presenter) seqMusic(cues []Cue) {
	for _, c := range cues {
		switch c.Op {
		case "load":
			p.seqSong = c.Song
		case "play":
			if p.AdLib != nil && p.Snd != nil {
				p.Snd.Song(p.AdLib.Named(p.seqSong), 0, true)
			}
		case "stop":
			if p.AdLib != nil && p.Snd != nil {
				p.Snd.Song(0, 0, false)
			}
		}
	}
}

// drawTyping draws the question over the line being typed, its cursor
// blinking, in the middle of dst, never above floor (under a picture).
func (p *Presenter) drawTyping(dst *gpu.Image, note Note, mk float64, cx, floor int) {
	p.menu = false
	t, w := p.typingAsk().Text, min(p.typingAsk().Max, 24)
	c := len(t)
	if p.typing {
		t, c = p.typed, p.typedCur
	}
	top := max(dst.Bounds().Dy()/2-int(mk*float64(len(WrapWords(p.typingAsk().Prompt, 24))+5)*4), floor)
	p.drawMenu(dst, note.Glyph, p.typingAsk().Prompt, []string{caretLine(t, c, w, p.blink())}, nil, nil, 0, mk, cx, top)
	p.menuRects = p.menuRects[:0]
}

// picPlace places the picture box pic (with the text box th high at tk and
// a menu at mk under it) in b at most at scale pk: its scale pp, its centre row py and the text
// box's centre row cy. The picture sits in the middle of the space over the
// text, else from b's top with the text moved down under it; it never runs
// off b's top.
func picPlace(pic CellRect, pk, tk, mk, th float64, b image.Rectangle) (pp float64, py, cy int) {
	ph, pw := float64(pic.Pixels().Dy()), float64(pic.Pixels().Dx())
	bw, bh := float64(b.Dx()), float64(b.Dy())
	gap := pk * 4
	pp = fitScale(pk, Host.Pic, func(k float64) bool { return k*pw <= bw && 3*gap+k*ph+tk*th+mk*7*8 <= bh })
	cy = b.Dy() / 2
	py = int(float64(cy)-tk*th/2) / 2
	if float64(py)-pp*ph/2 < gap || float64(py)+pp*ph/2+gap > float64(cy)-tk*th/2 {
		py = int(gap + pp*ph/2)
		cy = py + int(pp*ph/2+gap+tk*th/2)
	}
	return pp, py, cy
}
