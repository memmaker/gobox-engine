// shown.go: what the presenter shows, read from the game's State and UI
// (frame.go). The question up and the laid-out text page are derived once
// per build (Version) and the front's inputs they read; only the clock's
// part (a blink, an animation's frame) is applied as they are drawn.

package ui

import (
	"image/color"
	"slices"
	"time"

	"gobox/asset"
	"gobox/game"
)

const (
	blinkSwaps = 11                    // how many times a Blink run swaps between its ink and the highlight
	blinkMs    = 70 * time.Millisecond // the time between swaps
)

// laidPage is the state's text page laid out for drawing: the pane
// wrapped, the prompt row, the view window's text and the pictures decoded.
type laidPage struct {
	pane     *Pane // all its lines (nil: none)
	prompt   []PaneRun
	viewText [][]PaneRun
	pictures []Picture
	blinks   string // the blinking runs' text ("": none)
}

// laidKey is what the text page is laid out from: the build, the moment
// played, the creation sheet's boxes the pool pictures go in and the ink.
type laidKey struct {
	version  stateBuild
	creation *FrontCreation
	ink      [16]color.RGBA
}

// moment is the moment being played (nil: none), as Draw took it.
func (p *Presenter) moment() *frame { return p.played }

// replaying is whether a fight's moment holds the screen: its board, not
// the live page or question.
func (p *Presenter) replaying() bool {
	ev := p.moment()
	return ev != nil && ev.Screen == nil
}

// ink is the page's text colours: the EGA's, the title's lightning in
// entry 1 while it flashes (ovr02:0626 title_screen).
func (p *Presenter) ink() [16]color.RGBA {
	k := egaInk
	if p.introUp() != nil && len(p.intro.Flash) > 0 {
		k[1] = asset.EGARGBA(p.intro.Flash[0])
	}
	return k
}

// laid is the text page up, laid out: a played moment's page, else the
// state's (with a script portrait when it has no pictures of its own).
func (p *Presenter) laid() *laidPage {
	if p.eng == nil {
		return &laidPage{}
	}
	cr := p.layoutCreation()
	if !p.creating {
		cr = nil
	}
	k := laidKey{p.eng.r.build(), cr, p.ink()}
	return p.lay.page.update(k, func(old *laidPage) *laidPage {
		s := p.state()
		pg, portrait := p.eng.r.pageOf(s), storyOf(game.CommonOf(s))
		l := layPage(pg, portrait, k.ink)
		if l.blinks != "" && (old == nil || old.blinks != l.blinks) {
			p.blinkAt = now()
		}
		return l
	})
}

// layPage lays the page pg out (portrait: a script portrait's pictures
// when pg has none of its own) in ink.
func layPage(pg *TextPage, portrait []game.Picture, ink [16]color.RGBA) *laidPage {
	l := &laidPage{}
	if pg == nil {
		return l
	}
	if pg.Pane != nil {
		l.pane = wrapPane(&Pane{Lines: paneLines(pg.Pane, ink)})
	}
	l.prompt, l.viewText = paneRuns(pg.Prompt, ink), paneLines(pg.View, ink)
	l.pictures = append(l.pictures, pg.Layers...)
	if len(pg.Layers) == 0 {
		l.pictures = story(portrait)
	}
	l.pictures = indexedPictures(l.pictures, ink)
	if !slices.ContainsFunc(l.pictures, animated) { // the stills decoded once; an animation's frames as the clock shows them (pictures)
		l.pictures = daxPictures(l.pictures)
	}
	for _, ls := range [][][]PaneRun{{l.prompt}, l.viewText, linesOf(l.pane)} {
		for _, rs := range ls {
			for _, r := range rs {
				if r.Blink {
					l.blinks += r.Text + "\x00"
				}
			}
		}
	}
	return l
}

// linesOf is the pane's lines (none: nil).
func linesOf(pn *Pane) [][]PaneRun {
	if pn == nil {
		return nil
	}
	return pn.Lines
}

// linesText is the pane's lines as one string (a change of text).
func linesText(pn *Pane) string {
	var b []byte
	for _, l := range linesOf(pn) {
		for _, r := range l {
			b = append(b, r.Text...)
		}
		b = append(b, '\n')
	}
	return string(b)
}

// blinking is whether the Blink runs show in the highlight colour now: from
// their first showing, swapping blinkSwaps times to end on their own ink.
func (p *Presenter) blinking() bool {
	if p.laid().blinks == "" {
		return false
	}
	n := int(now().Sub(p.blinkAt) / blinkMs)
	return n < blinkSwaps && n%2 == 0
}

// blinked is rs with its Blink runs in the highlight colour while they blink.
func (p *Presenter) blinked(rs []PaneRun) []PaneRun {
	if !p.blinking() {
		return rs
	}
	c := slices.Clone(rs)
	for i := range c {
		if c[i].Blink {
			c[i].Fg = col(inkWhite)
		}
	}
	return c
}

// blinkedLines is blinked over each line.
func (p *Presenter) blinkedLines(ls [][]PaneRun) [][]PaneRun {
	if !p.blinking() {
		return ls
	}
	c := make([][]PaneRun, len(ls))
	for i, l := range ls {
		c[i] = p.blinked(l)
	}
	return c
}

// pane is the text pane under the 3D view as shown (nil: none).
func (p *Presenter) pane() *Pane {
	if p.backdrop { // the menu over the backdrop takes the text's place
		return nil
	}
	pn := p.laid().pane
	if pn == nil {
		return nil
	}
	c := *pn
	c.Lines = p.blinkedLines(c.Lines)
	return &c
}

// prompt is the prompt row's text (the last row).
func (p *Presenter) prompt() []PaneRun { return p.blinked(p.laid().prompt) }

// viewText is the text over the view window (WinText), rows from 1.
func (p *Presenter) viewText() [][]PaneRun { return p.blinkedLines(p.laid().viewText) }

// animated is whether pc cycles through frames by the clock.
func animated(pc Picture) bool {
	_, ok := spriteFrame(pc, now())
	return ok || len(pc.Frames) > 0 && pc.FrameMs > 0 || len(pc.Anim) > 1
}

// pictures is the page's pictures, bottom first, the animated ones at the
// frame the clock is on.
func (p *Presenter) pictures() []Picture {
	ps := p.laid().pictures
	if !slices.ContainsFunc(ps, animated) {
		return ps
	}
	for i, pc := range ps {
		if f, ok := gameFrame(pc, now()); ok { // the game's own frames (game.Drawn.Anim)
			ps = append(ps[:i:i], ps[i:]...)
			ps[i] = f
			pc = f
		}
		if f, ok := spriteFrame(pc, now()); ok { // a DAX animation block (PIC, SPRIT, FINAL): its frames at their own delays
			ps = append(ps[:i:i], ps[i:]...)
			ps[i].Frame = f
			pc = ps[i]
		}
		if len(pc.Frames) > 0 && pc.FrameMs > 0 {
			ps = append(ps[:i:i], ps[i:]...) // not the laid page's slice
			for j := i; j < len(ps); j++ {
				if q := ps[j]; len(q.Frames) > 0 && q.FrameMs > 0 {
					ps[j].Frame = q.Frames[int(now().UnixMilli()/int64(q.FrameMs))%len(q.Frames)]
				}
			}
			break
		}
	}
	return daxPictures(indexedPictures(ps, p.ink())) // a game frame swapped in comes as pixels
}

// askUp is the question up: the presenter's own (Ask), else the state's, else
// the locked door's menu (nil: none, or a fight's moment holds the screen).
func (p *Presenter) askUp() *FrontAsk {
	if p.backdrop { // the backdrop draws no question
		return nil
	}
	if p.q != nil {
		return &p.q.Ask
	}
	if p.eng == nil || p.replaying() {
		return nil
	}
	return p.lay.ask.get(p.uiKey(), func() *FrontAsk {
		if a := p.eng.r.askOf(p.state()); a != nil {
			return a
		}
		if q := GateQuestion(p, p.door()); q != nil {
			return &q.Ask
		}
		return nil
	})
}

// askOf is the state's question as the presenter asks it. A place's mode
// with no screen of its own yet draws its question, never a blank frame;
// the treasure's picture waits under a message (Enter or Space opens it).
func (r *frontState) askOf(ps game.State) *FrontAsk {
	mode, pending := r.modeOf(ps), r.asking(ps)
	if mode == game.ModeLoot && lootOf(ps) != nil && announced(ps) {
		return &FrontAsk{Kind: AskMenu, Menu: Menu{Prompt: "You found some loot!"}}
	}
	if q, ok := Ask(mode, pending); ok && q != nil {
		return q
	}
	services, _ := templeOf(ps)
	menu, _ := PartyMenu(mode, trainingOf(ps), pending, false)
	if menu == nil && mode != game.ModeCamp && services == nil && shopOf(ps) == nil && r.inventoryKind(ps) == invNone && lootOf(ps) == nil && trainingOf(ps) == nil &&
		pending != nil && mode >= game.ModeCamp && mode <= game.ModeLoot {
		if q, ok := Ask(game.ModeAsk, pending); ok {
			return q
		}
	}
	return nil
}

// door is the locked door the party stands at (nil: none).
func (p *Presenter) door() *game.Door {
	if p.eng == nil {
		return nil
	}
	return doorOf(p.state())
}

// powerPick is the power a member chooses (nil: none, or the training hall over it).
func (p *Presenter) powerPick() *game.PowerPick {
	if p.eng == nil || p.training {
		return nil
	}
	return game.CommonOf(p.state()).PowerPick
}

// campUp is whether camp is up, no picker or icon editor over it.
func (p *Presenter) campUp() bool {
	return p.eng != nil && !p.pickerOver() && p.eng.r.modeOf(p.state()) == game.ModeCamp
}

// shopUp and templeUp: a shop's, a temple's services are up.
func (p *Presenter) shopUp() bool { return p.eng != nil && shopOf(p.state()) != nil }
func (p *Presenter) templeUp() bool {
	if p.eng == nil {
		return false
	}
	services, _ := templeOf(p.state())
	return services != nil
}

// trainingUp is the training hall's offer to the selected member (nil: none).
func (p *Presenter) trainingUp() *FrontTraining {
	if p.eng == nil {
		return nil
	}
	return p.lay.training.get(p.uiKey(), func() *FrontTraining {
		s := p.state()
		return Training(trainingOf(s), partyOf(s), p.selected(), game.CommonOf(s).Purse)
	})
}

// results is the after-combat page (nil: none).
func (p *Presenter) results() *game.Results {
	if p.eng == nil {
		return nil
	}
	return resultsOf(p.state())
}

// introUp is the title waiting to be played (nil: none).
func (p *Presenter) introUp() *game.Intro {
	if p.eng == nil {
		return nil
	}
	return introOf(p.state())
}

// sequenceUp is the full-screen animation playing: the state's, else a
// game.MenuItem's (playSeq, gobox's own; "": none, or the adventure has no
// such animation).
func (p *Presenter) sequenceUp() string {
	name := playSeq
	if p.eng != nil {
		if n := sequenceOf(p.state()); n != "" {
			name = n
		}
	}
	if seqAnim(name) == nil {
		return ""
	}
	return name
}

// gameScreen is the game's typed screen up: a played moment's, else the state's (nil: none).
func (p *Presenter) gameScreen() *game.Screen {
	if ev := p.moment(); ev != nil && ev.Screen != nil {
		return ev.Screen
	}
	if p.eng == nil {
		return nil
	}
	return screenOf(p.state())
}

// note is the status line's toast: the rules' last refusal, else the front's own.
func (p *Presenter) note() FrontNote {
	if p.eng == nil {
		return FrontNote{}
	}
	if n := p.eng.r.note; n.Text != "" {
		return FrontNote{Text: n.Text, At: n.At}
	}
	return p.eng.r.toast
}

// readUp is the journal text the game waits to have read (nil: none).
func (p *Presenter) readUp() *game.Read {
	if p.eng == nil {
		return nil
	}
	return p.eng.r.readOf(p.state())
}

// sync settles what the presenter keeps over the state before it is read:
// a load or Back to Main Menu closes what was over the game; the creation
// sheet and the training hall close when the game ended them.
func (p *Presenter) sync() {
	if p.eng == nil {
		return
	}
	if p.eng.Resets() != p.resets {
		p.resets = p.eng.Resets()
		p.gridPick, p.icon, p.creating, p.training, p.q, p.clue = nil, nil, false, false, nil, nil
	}
	p.creating = p.creating && p.layoutCreation() != nil // the game ended it itself (a step after Keep, a save): the party menu is back
	p.training = p.training && p.hallUp()                // a game loaded closes it
}

// areaMapUp is whether the overland map is the screen: the whole area map
// over the window (drawAreaMap).
func (p *Presenter) areaMapUp() bool {
	return p.eng != nil && len(p.eng.r.derived().automap.Terrain) > 0
}
