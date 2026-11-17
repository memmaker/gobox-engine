package engine

import (
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
	blocks       []int                                   // PickWeapon: the sheet's icon blocks (memberIconBlocks)
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
	case PickWeapon: // a game of whole-picture icons: the sheet's (BR1: CustomizeIcon's CHARS list); the Pool's parts are gobox's pool
		blocks := e.memberIconBlocks()
		if e.iconEd == nil || len(blocks) == 0 {
			return false
		}
		p = &picker{n: len(blocks), blocks: blocks, cur: max(0, slices.Index(blocks, e.iconPics(e.iconEd.who).Body))}
		p.pic = func(i int) (*asset.Image, int) { return e.memberIcon(p.blocks[i], 0), int(e.artNames().IconMask) }
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
	case PickWeapon:
		e.editIcon(func(ic *rules.IconPics) { ic.Body = p.blocks[i] })
		done = true
	case PickAdd:
		done = e.addSaved(p, i)
	case PickAddFrom:
		done = e.addForeign(p, i)
	case PickForget:
		e.removeRosterFiles(p.labels[i])
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

// memberIconBlocks are the ready poses of the whole-picture member icons' sheet (artNames.MemberIcons, BR1's CHARS:
// blocks under 0x80, each with its action pose 0x80 on), as the art has them.
func (e *Engine) memberIconBlocks() []int {
	if e.artNames().MemberIcons == "" {
		return nil
	}
	var out []int
	for b := range memberIconAction {
		if e.memberIcon(b, 0) != nil {
			out = append(out, b)
		}
	}
	return out
}

// memberIconAction is how far a member icon's action pose follows its ready one in the sheet (LoadIconPic: block+0x80).
const memberIconAction = 0x80

// memberIcon is pose p (0 ready, 1 action) of icon block of the member icons' sheet, masked as the fight loads it.
func (e *Engine) memberIcon(block, p int) *asset.Image {
	return e.LoadPicMasked(e.artNames().MemberIcons, block+p*memberIconAction, e.artNames().IconMask)
}

// memberIconPic is block b of the member icons' sheet as a picture id (a draft's Icon.Body).
func (e *Engine) memberIconPic(b byte) rules.PicID {
	return e.pics.ID(asset.DAXRef{File: e.artNames().MemberIcons + ".DAX", Block: int(b), Frame: asset.NoFrame})
}

// memberIconByte is memberIconPic's inverse.
func (e *Engine) memberIconByte(id rules.PicID) (byte, bool) {
	for b := range memberIconAction {
		if id != "" && e.memberIconPic(byte(b)) == id {
			return byte(b), true
		}
	}
	return 0, false
}

// memberIconPoses are icon block's two poses as references into the member icons' sheet: the front loads them (its
// DAX cache, the game's mask), the state carries no pixels.
func (e *Engine) memberIconPoses(block int) (out [2]game.Picture) {
	for p := range out {
		out[p] = game.Picture{File: e.artNames().MemberIcons, Block: block + p*memberIconAction}
	}
	return out
}
