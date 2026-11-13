package engine

import (
	"slices"
	"strconv"
	"strings"

	"gobox/ecl/script"
	"gobox/game"
	"gobox/game/play"
	"gobox/rules"
)

// Questions (docs/engine-redesign.md, M4): an instruction or rule that
// needs the player does not wait. It records what it asks (the Ask)
// and how the answer completes it (eclWait), and returns; RunECL then
// opens it (openAsk), answer completes the instruction on the answering
// decision, and the VM goes on with the next one. A completion may ask
// again (the encounter's WAIT, a native script's next question): the
// question stays open until nothing is pending.

// eclWait is the pending question: what it asks (a play.Question) and how
// the answer (a play.Reply, Continue or SelectMember) completes it.
type eclWait struct {
	ask    *play.Question
	lines  []game.Line // a list's rows with their marks (play.Question has the text only)
	member bool        // a member pick: the roster highlights its candidate
	ms     int         // a pause: the presenter paces it
	anim   bool        // the picture animates while it waits
	shape  askShape    // how its replies answer it
	ids    []int       // a choice: the menu place of each reply
	list   []*ListNode // a list's nodes
	enter  bool        // a choice: Enter answers it (place 0)
	clear  bool        // the prompt row is cleared once it is answered
	maxlen int         // a typed line: at most so long
	done   askDone     // what the answer completes
	seq    int         // when it opened (openAsk; 0: not yet)
	save   *animSave   // what ending it puts back (openAsk)
}

// askShape is how the replies answer a question.
type askShape int

const (
	shChoice askShape = iota // a hotkey menu: the reply is the menu place
	shYesNo                  // no, yes
	shList                   // a line of list, or -1
	shLine                   // a typed line (or number)
	shMember                 // a party member
	shPause                  // a key, after the presenter's pause
)

// doneKind is what answering a question completes (askDone): the
// instruction that asked, or the step of a flow it goes on with. The
// question is plain data, as the answer is.
type doneKind int

const (
	dNone          doneKind = iota // nothing more
	dNumber                        // InputNumber: the low word to var dst
	dString                        // InputString: the line to var dst
	dList                          // VMenu: the line's index to var dst
	dChoiceVar                     // HMenu: the menu place to var dst
	dEncounter                     // the encounter menu's choice
	dParlay                        // Parlay: r[choice] to var dst
	dWho                           // Who: the member is the selected one (who: where it started)
	dPauseKey                      // a pause, then a key
	dYesNo                         // YesNo: equal when the answer is Yes
	dTavernWager                   // the tavern: the key after "YOU HAVE", then the wager
	dTavernSpun                    // the tavern: the wager is taken, the spin told
	dTavernAsk                     // the tavern: the spin had its key
	dTavernAgain                   // the tavern: go again or not
	dContinueFight                 // "Continue Battle:"
	dDevParlay                     // dev scene: the encounter menu's choice
	dDevPassword                   // dev scene: the word said
)

// askDone is what the answer completes, with what it needs.
type askDone struct {
	kind doneKind
	dst  script.Ref
	r    [5]byte
	who  rules.CharID
}

// ansv is the answer: a menu place or index, a line, a member.
type ansv struct {
	n    int
	s    string
	who  rules.CharID // NoChar: none
	list []*ListNode  // a list's nodes
}

// choosing: a member pick waits (the roster highlights its candidate).
func (e *Engine) choosing() bool { return e.eclWait != nil && e.eclWait.member }

// yesNo is a yes/no question (the highlight starts on Yes); strict: only Y
// and N answer it, not Enter or Esc (ovr26:10D1).
func yesNo(text string, strict bool) *play.Question {
	return &play.Question{Text: text, YesNo: true, Default: 1, Replies: []string{"No", "Yes"}, NoEsc: strict, NoPick: strict}
}

// askYesNo asks q (yesNo's); d gets the answer (n 1 for yes).
func (e *Engine) askYesNo(q *play.Question, anim bool, d askDone) {
	e.eclWait = &eclWait{ask: q, anim: anim, shape: shYesNo, done: d}
}

// animSave is the picture animation as it was when a question opened;
// under is the one a question answered under it put back after its own.
type animSave struct {
	anim  bool
	last  int64
	under *animSave
}

// openAsk opens the pending ask (the Ask).
func (e *Engine) openAsk() {
	w := e.eclWait
	w.save = &animSave{anim: e.animInput, last: e.pic.last}
	w.seq = e.opens()
	e.pic.last = e.Ticks() // the animation timer starts over, as in DisplayInput
	e.showAsk()
}

// restoreAnim puts back the animation as s and what was under it had it.
func (e *Engine) restoreAnim(s *animSave) {
	for ; s != nil; s = s.under {
		e.animInput, e.pic.last = s.anim, s.last
		e.viewAnim = 0
	}
}

// answer gives the reply r (a play.Reply, Continue, Leave or SelectMember)
// to the open question, which it completes or not. Asked again (the
// completion set another question), that one goes on in its place; else
// what follows runs.
func (e *Engine) answer(r play.Action) {
	w := e.eclWait
	v, ok := e.accept(w, r)
	if !ok {
		e.showAsk()
		return
	}
	if w.clear {
		e.clearPrompt()
	}
	e.complete(w.done, v)
	if w.member {
		e.rosterHi, e.rosterNone = rules.NoChar, false
	}
	switch n := e.eclWait; {
	case n == w:
		e.eclWait = nil
		e.restoreAnim(w.save)
	case n != nil && n.seq == 0: // asked again
		n.seq, n.save = w.seq, w.save
		e.showAsk()
		return
	case n != nil: // a question of its own: what it puts back goes on to w's
		s := n.save
		for s != nil && s.under != nil {
			s = s.under
		}
		if s != nil {
			s.under = w.save
		}
		return
	}
	e.closed()
}

// leaveAsk backs out of the open question q (Esc): the reply its Esc names
// (a script menu's exit= row, scriptExit), a typed line as it is, a page's
// Continue. A yes/no, a member pick and a menu naming no exit take no Esc: the
// choice is forced. Felix's rule; the original's MenuInput (ovr08:1684) took
// Esc on no menu.
// (Extra: a page's lone press-Enter reply, which the original also holds, takes Esc as Enter.)
func (e *Engine) leaveAsk(q *play.Question) error {
	if i, ok := q.Esc.Reply(); ok {
		e.answer(play.Reply{N: i})
		return nil
	}
	var r play.Action
	switch {
	case e.choosing(), q.YesNo, q.NoEsc:
		return rules.ErrNotNow
	case q.Typed, q.Number:
		r = play.Reply{}
	case q.Continue, e.eclWait.enter:
		r = play.Continue{}
	default:
		return rules.ErrNotNow
	}
	e.answer(r)
	return nil
}

// showAsk shows the pending ask: the picture's animation, a member
// pick's candidate in the roster.
func (e *Engine) showAsk() {
	w := e.eclWait
	e.animInput = w.anim
}

// listLines are the lines of list as the Ask shows them.
func listLines(list []*ListNode) []game.Line {
	l := make([]game.Line, len(list))
	for i, n := range list {
		l[i] = game.Line{Text: n.Text, Header: n.Flag != 0, Dim: n.Dim || n.Grey, Off: n.skipped(), Spell: n.Spell}
	}
	return l
}

// hotMenu is MenuInput's bar for menu as a question: each word with a hotkey
// is a reply (its hotkey lit and typed to pick it), and ids[N] the choice
// of the hotkey's place in the menu.
func hotMenu(prompt, menu string) (q *play.Question, ids []int) {
	hot, shown := extractHotkeys(menu)
	q = &play.Question{Text: strings.TrimSpace(prompt)}
	for _, w := range strings.Split(shown, " ") {
		if i := strings.IndexFunc(w, func(r rune) bool { return r < 0x80 && isHotkey(byte(r)) }); i >= 0 {
			q.Replies, q.Keys = append(q.Replies, optionText(w)), append(q.Keys, rune(w[i]))
			ids = append(ids, strings.IndexByte(hot, w[i]))
		}
	}
	return q, ids
}

// askChoice asks menu (MenuInput's hotkey text) after prompt: d gets the
// choice's place in the menu (0 for Enter with allowEnter).
func (e *Engine) askChoice(prompt, menu string, allowEnter, anim bool, d askDone) {
	q, ids := hotMenu(prompt, menu)
	q.NoEsc = !allowEnter // leaveAsk: MenuInput holds on Esc
	e.eclWait = &eclWait{ask: q, anim: anim, shape: shChoice, ids: ids, enter: allowEnter, done: d}
}

// scriptExit gives the menu just asked by the current instruction its
// exit row (Instr.Exit); a row that names none leaves the choice forced.
func (e *Engine) scriptExit() {
	if e.vm.Script == nil || e.vm.Cur() >= len(e.vm.Script.Code) {
		return
	}
	row, ok := e.vm.Script.Code[e.vm.Cur()].Exit.Reply()
	if !ok {
		return
	}
	w := e.eclWait
	if w.shape == shChoice { // the reply whose hotkey's place is the row
		reply := slices.Index(w.ids, row)
		if reply < 0 {
			return
		}
		row = reply
	}
	w.ask.Esc, w.ask.NoEsc = play.EscTo(row), false
}

// askContinue is the scripts' one-item "PRESS <RETURN>" menu; the prompt
// row is cleared with the key.
func (e *Engine) askContinue(d askDone) {
	e.askChoice("", "~PRESS <ENTER>/<RETURN> TO CONTINUE", true, false, d)
	e.eclWait.clear = true
}

// keyWait waits for a key outside a fight: the scripts' Continue, opened
// at once so what follows the caller waits too; c (pushed first) goes on
// with the key. The one key wait that works everywhere (pause's key is
// a fight event only).
func (e *Engine) keyWait(c cont) {
	e.push(c)
	e.askContinue(askDone{})
	e.openAsk()
}

// askList asks for a line of list (ListMenu without options): d gets
// its index, -1 when the player backs out.
func (e *Engine) askList(list []*ListNode, d askDone) {
	e.eclWait = &eclWait{ask: &play.Question{List: true}, lines: listLines(list), shape: shList, list: list, done: d}
}

// askLine asks for a typed line of at most maxlen (InputString's, upcased;
// with number, InputNumber's: asked again until it is 0..65536).
func (e *Engine) askLine(maxlen int, number bool, d askDone) {
	e.clearPrompt()
	q := &play.Question{Text: strings.TrimSpace(e.ext.printed), Typed: !number, Number: number, Max: maxlen} // the question is the text the script printed
	e.eclWait = &eclWait{ask: q, shape: shLine, maxlen: maxlen, done: d}
}

// askNumber is askLine for a number (InputNumber): d gets its low word.
func (e *Engine) askNumber(d askDone) { e.askLine(6, true, d) }

// askMember asks for a party member (SelectCharacter without Exit),
// starting at who; the candidate is the selected member as it steps (the
// var of opWho is the selection itself). With no party d completes at once.
func (e *Engine) askMember(prompt string, whoID rules.CharID, d askDone) {
	ids := e.partyIDs()
	if len(ids) == 0 {
		e.complete(d, ansv{})
		return
	}
	if whoID == rules.NoChar {
		whoID = ids[0]
	}
	e.rosterHi, e.rosterNone = whoID, false
	q := &play.Question{Text: strings.TrimSpace(rules.Cut(prompt, textCols) + " "), List: true}
	for _, id := range ids {
		q.Replies = append(q.Replies, e.ruleset().Sheet(id).Name)
	}
	e.eclWait = &eclWait{ask: q, member: true, shape: shMember, anim: e.beginFlag() == 2 || e.beginFlag() == 6, done: d}
}

// askPause is a pause of ms the presenter paces (the engine never sleeps);
// none, d completes at once.
func (e *Engine) askPause(ms int, d askDone) {
	if ms <= 0 {
		e.complete(d, ansv{})
		return
	}
	e.eclWait = &eclWait{ask: &play.Question{Continue: true}, ms: ms, shape: shPause, done: d}
}

// accept takes the reply a to w's question: the answer when it completes
// it, false when it keeps it waiting.
func (e *Engine) accept(w *eclWait, a play.Action) (ansv, bool) {
	switch w.shape {
	case shYesNo:
		if r, ok := a.(play.Reply); ok && r.N >= 0 && r.N <= 1 {
			return ansv{n: r.N}, true
		}
	case shChoice:
		switch a := a.(type) {
		case play.Reply:
			if a.N >= 0 && a.N < len(w.ids) {
				return ansv{n: w.ids[a.N]}, true
			}
		case play.Continue: // Enter (MenuInput's allowed Enter, 1723), or a lone page's Esc (extra: leaveAsk)
			if w.enter {
				return ansv{}, true
			}
		case play.SelectMember:
			e.selectMember(a.Member)
		}
	case shList:
		switch a := a.(type) {
		case play.Reply:
			if i := a.N; i >= 0 && i < len(w.list) && !w.list[i].skipped() {
				return ansv{n: i, list: w.list}, true
			}
		case play.SelectMember:
			e.selectMember(a.Member)
		}
	case shLine:
		r, ok := a.(play.Reply)
		if !ok {
			return ansv{}, false
		}
		b := make([]byte, 0, w.maxlen)
		for i := 0; i < len(r.Text) && len(b) < w.maxlen; i++ {
			if c := r.Text[i]; c >= 0x20 && c <= 0x7A {
				b = append(b, c)
			}
		}
		s := strings.ToUpper(string(b))
		if n, err := strconv.ParseInt(strings.TrimLeft(s, " "), 10, 32); w.ask.Number && (err != nil || n < 0 || n > 0x10000) {
			return ansv{}, false
		}
		e.clearPrompt()
		return ansv{s: s}, true
	case shMember:
		switch a := a.(type) {
		case play.Continue: // the highlighted member
			return ansv{who: e.rosterHi}, true
		case play.Reply: // member N: a portrait
			if ids := e.partyIDs(); a.N >= 0 && a.N < len(ids) {
				e.rosterHi = ids[a.N]
				return ansv{who: e.rosterHi}, true
			}
		case play.SelectMember: // SelectMember stepped the candidate
			if ids := e.partyIDs(); a.Member >= 0 && a.Member < len(ids) {
				e.rosterHi, e.rosterNone = ids[a.Member], false
				e.selectChar(e.rosterHi)
			}
		}
	case shPause:
		if _, ok := a.(play.Continue); ok {
			return ansv{}, true
		}
	}
	return ansv{}, false
}

// complete does what an answered question completes (d), given its answer v.
func (e *Engine) complete(d askDone, v ansv) {
	switch d.kind {
	case dNumber:
		n, _ := strconv.ParseInt(strings.TrimLeft(v.s, " "), 10, 32)
		e.host.Set(d.dst, int64(uint16(n)))
	case dString:
		if v.s == "" {
			v.s = " "
		}
		e.host.SetText(d.dst, v.s)
	case dList:
		e.listAnswered(d.dst, v.n)
	case dChoiceVar:
		e.host.Set(d.dst, int64(uint16(byte(v.n))))
		e.clearPrompt()
	case dEncounter:
		e.encounterAnswered(v.n)
	case dParlay:
		if v.n >= 0 && v.n < 5 {
			e.host.Set(d.dst, int64(uint16(d.r[v.n])))
		}
	case dWho:
		who := v.who
		if who == rules.NoChar { // Esc then Enter stores nil; the harness pins it to the op's start
			who = d.who
		}
		e.selectChar(who)
	case dYesNo:
		e.vm.Flags = [6]bool{v.n == 1, v.n != 1}
	case dPauseKey: // was PromptRow24's any key
		e.askContinue(askDone{})
	case dTavernWager:
		e.tavernWager()
	case dTavernSpun:
		e.tavernSpun(v.s)
	case dTavernAsk:
		e.tavernAsk()
	case dTavernAgain:
		e.tavernAgain(v.n)
	case dContinueFight:
		e.cs.phase = fightOver
		if v.n == 1 {
			e.cs.phase = fightRound
		}
	case dDevParlay:
		e.devParlayAnswered(v.n)
	case dDevPassword:
		e.scriptText(true, "YOU SAY '"+v.s+"'.")
		e.askContinue(askDone{})
	}
}
