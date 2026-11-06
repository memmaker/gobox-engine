// question.go: the questions the presenter asks itself (drawn and
// answered as the game's, View.Ask): the yes/no of a rule call's
// Confirm, a member or a spell a rule call needs (PickMember, PickSpell),
// a list to read (Magic); and the powers out of combat as rule calls.

package front

import (
	"errors"
	"slices"
)

// AskYes puts a yes/no question q up (AskConfirm, starting on No); yes
// runs on Yes.
func AskYes(h Host, q string, yes func()) {
	h.Ask(&Question{Ask: Ask{Kind: AskConfirm, Menu: Menu{Prompt: q}}, Answer: func(a Action) bool {
		switch a.Kind {
		case ActYes:
			yes()
		case ActNo, ActClose:
		default:
			return false
		}
		return true
	}})
}

// AskForMember asks for a party member (AskMember: Select, Exit), starting
// on member who; then gets the one picked (Exit: nothing).
func AskForMember(h Host, prompt string, who int, then func(n int)) {
	q := &Question{Ask: Ask{Kind: AskMember, Menu: Menu{Prompt: prompt,
		Items: []MenuItem{{Text: "Select", Key: 'S'}, {Text: "Exit", Key: 'E'}}}}, Cand: max(who, 0)}
	q.Answer = func(a Action) bool {
		n := h.Shown().Party
		switch a.Kind {
		case ActLead:
			if a.N >= 0 && a.N < n {
				q.Cand = a.N
			}
			return false
		case ActMember:
			if a.N < 0 || a.N >= n {
				return false
			}
			then(a.N)
		case ActOK:
			then(q.Cand)
		case ActChoose:
			if a.N == 0 {
				then(q.Cand)
			}
		case ActClose, ActNo:
		default:
			return false
		}
		return true
	}
	h.Ask(q)
}

// AskForSpell asks for one of spells (AskSpell); then gets its ID (Close:
// nothing).
func AskForSpell(h Host, spells []Spell, then func(id int)) {
	h.Ask(&Question{Ask: Ask{Kind: AskSpell, Spells: spells}, Answer: func(a Action) bool {
		switch a.Kind {
		case ActCast:
			if slices.ContainsFunc(spells, func(s Spell) bool { return s.ID == a.N }) {
				then(a.N)
				return true
			}
		case ActClose:
			return true
		}
		return false
	}})
}

// ShowList shows lines (AskList) until Exit.
func ShowList(h Host, lines []ListLine) {
	h.Ask(&Question{Ask: Ask{Kind: AskList, List: lines, Menu: Menu{Items: []MenuItem{{Text: "Exit", Key: 'E'}}}},
		Answer: func(a Action) bool {
			return a.Kind == ActChoose || a.Kind == ActClose || a.Kind == ActOK
		}})
}

// AskAct takes a while the host's question is up (every other decision
// waits), and asks "Stop resting?" for one made while camp rests; done
// false leaves a to the others.
func AskAct(h Host, a Action) (taken, done bool) {
	s := h.Shown()
	switch q := h.Asking(); {
	case q != nil:
		if a.Kind == ActResize || a.Kind == ActFont || a.Kind == ActSound {
			return false, true
		}
		if a.Kind != ActPoint && q.Answer(a) && h.Asking() == q {
			h.Ask(nil)
		}
		return true, true
	case s.Loot != nil || s.Door != nil:
		return GateAct(h, a), true
	case s.Treasure != nil:
		if lootAct(h, s.Treasure, a) {
			return true, true
		}
	case s.Resting && !s.Battle && a.Kind != ActPoint && a.Kind != ActResize && a.Kind != ActMainMenu: // a fight's turn takes its own decisions
		AskYes(h, "Stop resting?", func() { h.Call(func(e Game) { e.StopRest() }) })
		return true, true
	}
	return false, false
}

// PowerAct makes a member's power out of combat (the sheet's, the
// inventory's, a card's, camp's, the bar's Cast) up a rule call; Magic
// lists the effects; false: a is no power.
func PowerAct(h Host, a Action) bool {
	k := a.Kind
	switch k {
	case ActCastPick:
		k = ActSpells
	case ActSpells, ActMagic, ActHeal, ActCure:
	default:
		return false
	}
	v := h.Shown()
	if v.Ask || v.Powers {
		return false
	}
	n := a.N
	if a.Kind != ActCastPick && (v.Sheet || v.Inventory) {
		n = -1 // the member shown
	}
	switch {
	case k == ActMagic:
		h.Call(func(e Game) { ShowList(h, e.EffectsList(n)) })
	case v.Battle:
		return false
	default:
		PowerCall(h, n, k, 0, -1, false)
	}
	return true
}

// PowerCall calls UsePower; a pick or question it is refused with is
// asked, and the call made again with the answer.
func PowerCall(h Host, n int, k ActKind, spell, target int, sure bool) {
	h.Call(func(e Game) {
		var (
			who  PickMember
			pick PickSpell
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
	do   func(e Game)
}

// Gate is the menu of the treasure screen (loot) or a locked door
// (door), the last item Exit; nil: neither is up.
func Gate(h Host, l *Loot, d *DoorOpts) (string, []GateItem) {
	var items []GateItem
	add := func(on bool, text string, do func(e Game)) {
		if on {
			items = append(items, GateItem{text, do})
		}
	}
	door := func(k DoorWay) func(e Game) { return func(e Game) { e.TryDoor(k) } }
	switch {
	case l != nil:
		add(true, "View", func(e Game) { e.LootView(-1) })
		add(l.Take, "Take", func(e Game) { e.LootTake(-1) })
		add(l.Detect, "Detect", func(e Game) { e.LootDetect() })
		add(l.Share, "Share", func(e Game) {
			if s, ok := e.(interface{ LootShare() error }); ok {
				s.LootShare()
			}
		}) // only a game that offers it sets Share
		add(l.Pool, "Pool", func(e Game) {
			if p, ok := e.(interface{ LootPool() error }); ok {
				p.LootPool()
			}
		}) // only a game that offers it sets Pool
		add(l.Status, "Status", func(e Game) {
			if s, ok := e.(interface{ LootStatus() error }); ok {
				s.LootStatus()
			}
		}) // only a game that offers it sets Status
		add(true, "Exit", func(e Game) {
			AskSure(h, e.LootLeave(false), func() { h.Call(func(e Game) { e.LootLeave(true) }) })
		})
		return "", items
	case d != nil:
		add(d.Bash, "Bash", door(DoorBash))
		add(d.Pick, "Pick", door(DoorPick))
		add(d.Knock, "Knock", door(DoorKnock))
		add(true, "Exit", door(DoorExit))
		return "Locked.", items
	}
	return "", nil
}

// lootActs are the loot inventory's acts that make a Gate call.
var lootActs = map[ActKind]string{ActLootDetect: "Detect", ActLootShare: "Share", ActLootPool: "Pool", ActLootStatus: "Status", ActClose: "Exit"}

// lootAct makes a's Gate call on the loot inventory (Exit asks first while
// treasure is left); false: a is not one.
func lootAct(h Host, l *Loot, a Action) bool {
	_, items := Gate(h, l, nil)
	for _, it := range items {
		if lootActs[a.Kind] == it.text {
			h.Call(it.do)
			return true
		}
	}
	return false
}

// GateQuestion is the treasure screen's or the locked door's menu as a
// question (nil: neither is up).
func GateQuestion(h Host, l *Loot, d *DoorOpts) *Question {
	prompt, items := Gate(h, l, d)
	if items == nil {
		return nil
	}
	q := &Question{Ask: Ask{Kind: AskMenu, Menu: Menu{Prompt: prompt}}}
	for _, it := range items {
		q.Ask.Items = append(q.Ask.Items, MenuItem{Text: it.text, Key: it.text[0]})
	}
	return q
}

// GateAct takes a on the treasure screen or at a locked door: an option
// (ActChoose N, Esc Exit) is its rule call; F1-F8 and the portraits pick
// the member the treasure is for.
func GateAct(h Host, a Action) bool {
	v := h.Shown()
	_, items := Gate(h, v.Loot, v.Door)
	switch a.Kind {
	case ActResize, ActFont, ActSound:
		return false
	case ActChoose:
		if a.N >= 0 && a.N < len(items) {
			h.Call(items[a.N].do)
		}
	case ActClose:
		h.Call(items[len(items)-1].do)
	case ActLead, ActMember:
		if v.Loot != nil {
			h.Call(func(e Game) { e.Lead(a.N) })
		}
	}
	return true
}
