package engine

import (
	"gobox/adventure"
	"gobox/rules"
	"slices"
	"strings"

	"gobox/asset"
	"gobox/game"
)

// Grid pickers (extra): the saved characters are picked from all of them.
// OpenPicker readies the candidates (game.Picker), ui lays them out in pages,
// Pick takes one.
// Portrait and combat-icon parts are gobox's shared-pool picker
// (front/present/poolpick.go), answered by SetChoice.

// picker is the grid picker readied (OpenPicker).
type picker struct {
	kind, n, cur int
	pic          func(i int) (im *asset.Image, mask int) // picture i (nil: none)
	labels       []string                                // the pictures' names (saved characters)
	chars        []rules.CharID                          // ... the characters (core.NoChar where the files are gone)
	ask          string                                  // asked before a pick (%s: the label)

	state *game.Picker // pickState's pictures, built once (nil: again)
}

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
		p = &picker{n: len(chars), labels: names, chars: chars}
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
	e.picker = p
	return true
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
		e.as(e.sel) // ImportCharacter's member
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
		return count > 5 || e.memberCount() > 7
	}
	return e.memberCount() > 7
}
