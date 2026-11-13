package engine

import (
	"gobox/adventure"
	"gobox/rules"
	"image"
	"slices"
	"strings"

	"gobox/asset"
	"gobox/game"
)

// Grid pickers (extra): the saved characters are picked from a grid of all of them, as many
// columns and rows as the whole page holds (a larger window shows more; no
// bar: the box takes the prompt row too), centred in the box. What does not
// fit is on further pages: a pager line "< page 1/2 >" at the bottom of the
// box says so. The picker is the presenter's screen: OpenPicker readies
// one, PickerView lays a page of it out with its pictures, Pick takes one.
// Portrait and combat-icon parts are gobox's shared-pool picker
// (front/present/poolpick.go), answered by SetChoice.

// picker is the grid picker readied (OpenPicker).
type picker struct {
	kind, n, cw, ch, cur int
	pic                  func(i int) (im *asset.Image, mask int) // picture i (nil: none)
	labels               []string                                // the pictures' names (saved characters)
	chars                []rules.CharID                          // ... the characters (core.NoChar where the files are gone)
	ask                  string                                  // asked before a pick (%s: the label)

	cols, rows, gap, x0, y0, bot, pager, page int // the layout, from the page's size

	state    *game.Picker // pickState's pictures, built once (nil: again)
	stateDOS bool         // the art set they were built in
}

func (p *picker) layout() {
	p.bot = promptRow - 1 // the box's inner bottom: the page's last row holds its border
	p.gap, p.cols = 1, winRight/(p.cw+1)
	if winRight/p.cw > p.cols { // a blank between the columns only when it costs none
		p.gap, p.cols = 0, winRight/p.cw
	}
	p.cols = max(1, min(p.cols, p.n))
	p.rows, p.pager = max(1, p.bot/p.ch), 0
	if (p.n+p.cols-1)/p.cols > p.rows { // paged: the pager line takes the box's last row
		p.rows, p.pager = max(1, (p.bot-1)/p.ch), p.bot
	}
	p.rows = min(p.rows, (p.n+p.cols-1)/p.cols)
	p.x0 = 1 + (winRight-(p.cols*(p.cw+p.gap)-p.gap))/2
	p.y0 = 1 + (p.bot-p.pager/p.bot-p.rows*p.ch)/2
	p.page = min(max(p.page, 0), p.pages()-1)
}

func (p *picker) perPage() int { return p.cols * p.rows }
func (p *picker) pages() int   { return max(1, (p.n+p.perPage()-1)/p.perPage()) }

// OpenPicker readies picker kind for the presenter: false when it has
// nothing to pick now (no creation or icon editor up, no saved
// characters).
func (e *Engine) OpenPicker(kind int) bool {
	var p *picker
	head := func(n byte) (*asset.Image, int) {
		heads := e.portraitPics("HEAD")
		if n == 0 || int(n) >= len(heads) {
			return nil, -1
		}
		return e.LoadPic(heads[n].Set.File("HEAD"), int(heads[n].Id)), -1
	}
	switch kind {
	case PickAdd, PickForget, PickAddFrom:
		// pickSaved (extra): the Add and Forget lists are cards, the head of
		// the character's portrait (a dropped PNG fitted into its place) with
		// the name and class beside it; a character whose files are gone
		// (nil) is its name alone.
		var names []string
		var chars []rules.CharID
		if kind == PickAddFrom {
			if names, chars = e.foreignChars(e.addFrom); len(names) == 0 {
				return false
			}
		} else {
			if names = e.BuildCharList(); len(names) == 0 { // ovr18:0010
				return false
			}
			chars = e.loadSaved(names)
		}
		w := 0
		for i, c := range chars {
			w = max(w, len(names[i]))
			if c != rules.NoChar {
				w = max(w, len(e.word(c, rules.FactClasses)))
			}
		}
		p = &picker{n: len(chars), cw: 12 + w, ch: 5, labels: names, chars: chars}
		p.pic = func(i int) (*asset.Image, int) {
			if pt, ok := e.portrait(p.chars[i]); ok && pt.PNG == nil { // a PNG, the name and class are the presenter's
				return head(byte(pt.Head))
			}
			return nil, -1
		}
		if kind == PickForget {
			p.ask = "Forget %s forever?"
		}
	default:
		return false
	}
	p.kind = kind
	p.cur = min(max(p.cur, 0), max(p.n-1, 0))
	p.layout()
	p.page = p.cur / p.perPage()
	e.picker = p
	return true
}

// pictureOf is frame f of im at page pixel (x, y) as the page would show
// it (pageView), without drawing it on the page.
func (e *Engine) pictureOf(x, y int, im *asset.Image, f, mask int) Picture {
	return e.layerPicture(&layer{at: image.Rect(x, y, x+im.Width*8, y+im.Height), d: im.Frame(f), w: im.Width * 8, mask: mask, src: im, frame: f})
}

// Pick takes picture i of the picker readied; done when the picker goes
// (Add: the party is full; Forget: none are left).
func (e *Engine) Pick(i int) (done bool) {
	p := e.picker
	if p == nil || i < 0 || i >= p.n {
		return p == nil
	}
	p.cur, p.state = i, nil
	switch p.kind {
	case PickAdd:
		done = e.addSaved(p, i)
	case PickAddFrom:
		done = e.addForeign(p, i)
	case PickForget:
		base := adventure.MakeFileName(rules.RemoveSubstr(p.labels[i], "."))
		for _, ext := range []string{".cha", ".itm", ".spc", ".mem", ".dual", ".pow", ".png", ".icon.png"} {
			e.removeSaved(e.savePath(base + ext))
		}
		p.labels, p.chars = slices.Delete(p.labels, i, i+1), slices.Delete(p.chars, i, i+1)
		p.n = len(p.labels)
		e.WriteCharList(p.labels)
		done = p.n == 0
	}
	if done {
		e.picker = nil
	}
	return done
}

// ClosePicker puts the picker readied away.
func (e *Engine) ClosePicker() { e.picker = nil }

// loadSaved loads the saved characters named (core.NoChar where the files are gone).
func (e *Engine) loadSaved(names []string) []rules.CharID {
	chars := make([]rules.CharID, len(names))
	for i, n := range names {
		chars[i] = e.loadChar(strings.TrimPrefix(n, "* ")) // ovr18:0AEC
	}
	return chars
}

// iconRect is the page rectangle of a combat icon pose drawn at cell (x, y).
func iconRect(x, y int) image.Rectangle { return image.Rect(x*8, y*8, x*8+24, y*8+24) }

// foreignChars are the characters another game saved (Config.Foreign) taken into the
// ruleset, flagged as added from another game (core.MarkImported), with their names.
func (e *Engine) foreignChars(from string) (names []string, chars []rules.CharID) {
	if e.Cfg.Foreign == nil {
		return nil, nil
	}
	files, err := e.Cfg.Foreign(from)
	if err != nil {
		e.fileError("read", from)
		return nil, nil
	}
	for _, f := range files {
		e.setWorldWord(rules.WordWho, e.sel) // ImportCharacter's member
		id, err := e.ruleset().Import(f)
		if err != nil {
			e.fileError("read", from)
			continue
		}
		e.do(rules.MarkImported{Who: id})
		names, chars = append(names, e.ruleset().Sheet(id).Name), append(chars, id)
	}
	return names, chars
}

// addForeign is the pick of character i of the PickAddFrom picker: it joins the party as a saved one does;
// the list marks it ("* "), done when the party is full.
func (e *Engine) addForeign(p *picker, i int) bool {
	if id := p.chars[i]; id != rules.NoChar && !strings.HasPrefix(p.labels[i], "*") {
		p.labels[i] = rules.Cut("* "+p.labels[i], textCols)
		refused, count := e.joinPicked(id)
		if refused {
			p.labels[i] = p.labels[i][2:]
		}
		return count > 5 || e.worldWord(rules.WordPartyCount) > 7
	}
	return e.worldWord(rules.WordPartyCount) > 7
}
