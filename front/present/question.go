// question.go: the questions the presenter asks itself (drawn and
// answered as the game's, the Ask): the yes/no of a rule call's
// Confirm, a member or a spell a rule call needs (PickMember, PickSpell),
// a list to read (Magic); and the powers out of combat as rule calls.

package present

import (
	"errors"
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"slices"
)

// AskYes puts a yes/no question q up (AskConfirm, starting on No); yes
// runs on Yes.
func AskYes(h RuleHost, q string, yes func()) {
	h.Ask(&front.Question{Ask: front.Ask{Kind: front.AskConfirm, Menu: front.Menu{Prompt: q}}, Answer: func(a front.Action) bool {
		switch a.Kind {
		case front.ActYes:
			yes()
		case front.ActNo, front.ActClose:
		default:
			return false
		}
		return true
	}})
}

// AskForMember asks for a party member (AskMember: Select, Exit), starting
// on member who; then gets the one picked (Exit: nothing).
func AskForMember(h RuleHost, prompt string, who int, then func(n int)) {
	q := &front.Question{Ask: front.Ask{Kind: front.AskMember, Menu: front.Menu{Prompt: prompt,
		Commands: []front.Command{{Text: "Select", Key: 'S'}, {Text: "Exit", Key: 'E'}}}}, Cand: max(who, 0)}
	q.Answer = func(a front.Action) bool {
		n := h.Shown().Party
		switch a.Kind {
		case front.ActLead:
			if a.N >= 0 && a.N < n {
				q.Cand = a.N
			}
			return false
		case front.ActMember:
			if a.N < 0 || a.N >= n {
				return false
			}
			then(a.N)
		case front.ActOK:
			then(q.Cand)
		case front.ActChoose:
			if a.N == 0 {
				then(q.Cand)
			}
		case front.ActClose, front.ActNo:
		default:
			return false
		}
		return true
	}
	h.Ask(q)
}

// spellEntries is spells as AskSpell's entries: the name, the ID as Spell.
func spellEntries(spells []front.Spell) []front.Entry {
	out := make([]front.Entry, len(spells))
	for i, s := range spells {
		out[i] = front.Entry{Text: s.Name, Spell: s.ID}
	}
	return out
}

// AskForSpell asks for one of spells (AskSpell); then gets its ID (Close:
// nothing).
func AskForSpell(h RuleHost, spells []front.Spell, then func(id int)) {
	h.Ask(&front.Question{Ask: front.Ask{Kind: front.AskSpell, Entries: spellEntries(spells)}, Answer: func(a front.Action) bool {
		switch a.Kind {
		case front.ActCast:
			if slices.ContainsFunc(spells, func(s front.Spell) bool { return s.ID == a.N }) {
				then(a.N)
				return true
			}
		case front.ActClose:
			return true
		}
		return false
	}})
}

// ShowList shows lines (AskList) until Exit.
func ShowList(h RuleHost, lines []front.Entry) {
	h.Ask(&front.Question{Ask: front.Ask{Kind: front.AskList, Entries: lines, Menu: front.Menu{Commands: []front.Command{{Text: "Exit", Key: 'E'}}}},
		Answer: func(a front.Action) bool {
			return a.Kind == front.ActChoose || a.Kind == front.ActClose || a.Kind == front.ActOK
		}})
}

// AskAct takes a while the host's question is up (every other decision
// waits), and asks "Stop resting?" for one made while camp rests; done
// false leaves a to the others.
func AskAct(h RuleHost, a front.Action) (taken, done bool) {
	s := h.Shown()
	switch q := h.Asking(); {
	case q != nil:
		if a.Kind == front.ActResize || a.Kind == front.ActFont || a.Kind == front.ActSound {
			return false, true
		}
		if a.Kind != front.ActPoint && q.Answer(a) && h.Asking() == q {
			h.Ask(nil)
		}
		return true, true
	case s.Door != nil:
		return GateAct(h, a), true
	case s.Treasure != nil:
		if lootAct(h, s.Treasure, a) {
			return true, true
		}
	case s.Resting && !s.Battle && a.Kind != front.ActPoint && a.Kind != front.ActResize && a.Kind != front.ActMainMenu: // a fight's turn takes its own decisions
		AskYes(h, "Stop resting?", func() { h.Call(func(e *classicGame) { e.StopRest() }) })
		return true, true
	}
	return false, false
}

// PowerAct makes a member's power out of combat (the sheet's, the
// inventory's, a card's, camp's, the bar's Cast) up a rule call; Magic
// lists the effects; false: a is no power.
func PowerAct(h RuleHost, a front.Action) bool {
	k := a.Kind
	switch k {
	case front.ActCastPick:
		k = front.ActSpells
	case front.ActSpells, front.ActMagic, front.ActHeal, front.ActCure:
	default:
		return false
	}
	v := h.Shown()
	if v.Ask || v.Powers {
		return false
	}
	n := a.N
	if a.Kind != front.ActCastPick && (v.Sheet || v.Inventory) {
		n = -1 // the member shown
	}
	switch {
	case k == front.ActMagic:
		h.Call(func(e *classicGame) { ShowList(h, e.EffectsList(n)) })
	case v.Battle:
		return false
	default:
		PowerCall(h, n, k, 0, -1, false)
	}
	return true
}

// PowerCall calls UsePower; a pick or question it is refused with is
// asked, and the call made again with the answer.
func PowerCall(h RuleHost, n int, k front.ActKind, spell, target int, sure bool) {
	h.Call(func(e *classicGame) {
		var (
			who  front.PickMember
			pick front.PickSpell
		)
		switch err := e.UsePower(n, k, spell, target, sure); {
		case errors.As(err, &who):
			AskForMember(h, string(who), e.SelectedMember(), func(t int) { PowerCall(h, n, k, spell, t, sure) })
		case errors.As(err, &pick):
			AskForSpell(h, pick, func(id int) { PowerCall(h, n, k, id, target, sure) })
		default:
			AskSure(h, err, func() { PowerCall(h, n, k, spell, target, true) })
		}
	})
}

// GateItem is an option of the treasure screen or a locked door and the
// rule call it makes.
type GateItem struct {
	text string
	do   func(e *classicGame)
}

// Gate is the menu of the treasure screen (loot) or a locked door
// (door), the last item Exit; nil: neither is up.
func Gate(h RuleHost, l *front.Loot, d *game.Door) (string, []GateItem) {
	var items []GateItem
	add := func(on bool, text string, do func(e *classicGame)) {
		if on {
			items = append(items, GateItem{text, do})
		}
	}
	door := func(k play.DoorWay) func(e *classicGame) { return func(e *classicGame) { e.TryDoor(k) } }
	switch {
	case l != nil:
		add(true, "View", func(e *classicGame) { e.LootView(-1) })
		add(l.Take, "Take", func(e *classicGame) { e.LootTake(-1) })
		add(l.Detect, "Detect", func(e *classicGame) { e.LootDetect() })
		add(l.Share, "Share", func(e *classicGame) {
			e.LootShare()
		})
		add(l.Pool, "Pool", func(e *classicGame) {
			e.LootPool()
		})
		add(l.Status, "Status", func(e *classicGame) {
			e.LootStatus()
		})
		add(true, "Exit", func(e *classicGame) {
			AskSure(h, e.LootLeave(false), func() { h.Call(func(e *classicGame) { e.LootLeave(true) }) })
		})
		return "", items
	case d != nil:
		add(d.Bash, "Bash", door(play.DoorBash))
		add(d.Pick, "Pick", door(play.DoorPick))
		add(d.Knock, "Knock", door(play.DoorKnock))
		add(true, "Exit", door(play.DoorExit))
		return "Locked.", items
	}
	return "", nil
}

// lootActs are the loot inventory's acts that make a Gate call.
var lootActs = map[front.ActKind]string{front.ActLootDetect: "Detect", front.ActLootShare: "Share", front.ActLootPool: "Pool", front.ActLootStatus: "Status", front.ActClose: "Exit"}

// lootAct makes a's Gate call on the loot inventory (Exit asks first while
// treasure is left); false: a is not one.
func lootAct(h RuleHost, l *front.Loot, a front.Action) bool {
	_, items := Gate(h, l, nil)
	for _, it := range items {
		if lootActs[a.Kind] == it.text {
			h.Call(it.do)
			return true
		}
	}
	return false
}

// GateQuestion is the locked door's menu as a question (nil: none is up).
func GateQuestion(h RuleHost, d *game.Door) *front.Question {
	prompt, items := Gate(h, nil, d)
	if items == nil {
		return nil
	}
	q := &front.Question{Ask: front.Ask{Kind: front.AskMenu, Menu: front.Menu{Prompt: prompt}}}
	for _, it := range items {
		q.Ask.Commands = append(q.Ask.Commands, front.Command{Text: it.text, Key: it.text[0]})
	}
	return q
}

// GateAct takes a at a locked door: an option (ActChoose N, Esc Exit) is
// its rule call.
func GateAct(h RuleHost, a front.Action) bool {
	_, items := Gate(h, nil, h.Shown().Door)
	switch a.Kind {
	case front.ActResize, front.ActFont, front.ActSound:
		return false
	case front.ActChoose:
		if a.N >= 0 && a.N < len(items) {
			h.Call(items[a.N].do)
		}
	case front.ActClose:
		h.Call(items[len(items)-1].do)
	}
	return true
}
