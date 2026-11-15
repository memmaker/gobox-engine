package engine

import (
	"gobox/rules"
	"strconv"
	"strings"
)

// ClearScreen = 0709:0CAB.
func (e *Engine) ClearScreen() {
	e.clearPage()
}

// CharacterSheet = ovr20:004A: the selected member's sheet (SheetPage,
// laid out by the presenter).
func (e *Engine) CharacterSheet() {
	id := e.selected()
	e.ClearScreen()
	e.applyItems(id)
	e.snap.sheet = e.sheetPage(id)
	if e.viewBit0 != 0 { // the original: only while set 3 is loaded (its HEAD/BODY); extra: everywhere, from charArt
		e.SheetPortrait()
	}
}

// sheetPage is c's sheet page as CharacterSheet shows it, its portrait
// aside (SheetPortrait); with no side effect, so State builds it for every
// member (game.Member.Page, BUG-CLASSES 36).
func (e *Engine) sheetPage(id rules.CharID) *SheetPage {
	sh := e.ruleset().Sheet(id)
	sp := &SheetPage{Name: sh.Name, NameColour: e.nameColour(id), NPC: e.fact(id, rules.FactControl).(rules.Number) > 0x7F,
		Sex: e.word(id, rules.FactSex), Race: e.word(id, rules.FactRace), Age: int(sh.Age),
		Alignment: e.word(id, rules.FactAlignment), Class: e.word(id, rules.FactClasses), Str00: e.word(id, rules.FactExceptional), Creating: e.ext.creating}
	for i, a := range [...]rules.Ability{rules.Strength, rules.Intelligence, rules.Wisdom, rules.Dexterity, rules.Constitution, rules.Charisma} {
		sp.Stats[i] = int(sh.Scores[a])
	}
	e.combatStats(id, sp)
	if !e.ext.creating { // extra: the creation's sheet: no level, experience or status yet (create.go)
		for _, l := range e.fact(id, rules.BookFact{}).(rules.Book).Levels[:8] {
			if int8(l) > 0 {
				if sp.Levels != "" {
					sp.Levels += "/"
				}
				sp.Levels += strconv.Itoa(int(int8(l)))
			}
		}
		sp.Exp = int(int32(sh.Exp))
		if need := sh.NextExp; need > 0 { // extra: the experience the next level takes, green once it is there
			sp.Next = int(need)
		}
		for _, v := range e.ruleset().Catalog().Saves { // extra: the saving throws, as many as the ruleset has
			sp.Saves = append(sp.Saves, int(e.fact(id, rules.SaveFact{Save: v}).(rules.Number)))
		}
		sp.Status = e.statusName(id)
		sp.Effects, _ = e.effectsLine(id)
		sp.DualClass = sh.DualClass
	}
	return sp
}

// Extra: the sheet lists the spell effects on the character (the camp
// menu's Display, which is not offered) on the free row under HP. When they
// do not all fit, the View menu offers Magic with the whole list.
func (e *Engine) effectsLine(id rules.CharID) (string, bool) {
	names := e.ruleset().Sheet(id).Effects
	if len(names) == 0 {
		return "", true
	}
	s := strings.Join(names, ", ")
	if len(s) <= winRight-9 {
		return s, true
	}
	n := 0
	for len(strings.Join(names[:n+1], ", "))+len(", +9 more") <= winRight-9 {
		n++
	}
	if n == 0 {
		return strconv.Itoa(len(names)) + " spell effects", false
	}
	return strings.Join(names[:n], ", ") + ", +" + strconv.Itoa(len(names)-n) + " more", false
}

func (e *Engine) effectsFit(id rules.CharID) bool { _, ok := e.effectsLine(id); return ok }

// combatStats = showCombatStats (ovr20:0608): AC, HP, THAC0, damage,
// party items, movement.
func (e *Engine) combatStats(id rules.CharID, sp *SheetPage) {
	sh := e.ruleset().Sheet(id)
	sp.AC, sp.RearAC, sp.HP, sp.MaxHP = int(sh.AC), int(sh.RearAC), int(sh.HP), int(sh.MaxHP)
	sp.THAC0, sp.Damage, sp.Move = int(sh.THAC0), rules.Cut(string(e.fact(id, rules.DamageFact{}).(rules.Text)), textCols), int(sh.Move)
	if sp.Rules = e.fact(id, rules.FiguresFact{}).(rules.Figures); sp.Rules != nil { // extra: the rule set's modifiers and figures
		for a := range sp.Mods {
			sp.Mods[a] = e.mod(id, a)
		}
	}
}

// sheetView is the sheet's decisions for the View.
func (e *Engine) sheetView() *Sheet {
	id := e.selected()
	return &Sheet{Powers: e.powers(id), Reactions: e.fact(id, rules.ReactionsFact{}).(rules.Reactions)}
}

// SheetPortrait = ovr29:04F9: the character's head and body in a box at
// the top right of the sheet (SheetPage.PortraitAt; the presenter frames it).
func (e *Engine) SheetPortrait() {
	pt := e.fact(e.selected(), rules.PortraitFact{}).(rules.Portrait)
	if pt.PNG == nil && (pt.Head == 0 || pt.Body == 0) {
		return
	}
	box := CellRect{X1: winRight - 10, Y1: 1, X2: winRight, Y2: 11} // against the right border, whatever the page's width
	if e.creation != nil {
		box = sheetPortraitBox() // larger as the page has room (the presenter scales the pictures)
	}
	e.snap.portrait = box // Shift-click and a dropped PNG land here
	if sp := e.snap.sheet; sp != nil {
		sp.PortraitAt, sp.Portrait = box, pt.PNG // extra: the dropped PNG, fitted into the box at the window's resolution; the head and body the presenter's (sheetPortrait)
	}
}

// identifyItem = ovr20:1F52 (item): 200 gp shows the hidden words (the
// player took the deal).
func (e *Engine) identifyItem(id rules.CharID, h rules.ItemHandle) {
	for _, f := range e.resolve(rules.IdentifyItem{Who: id, Item: h}).Effects {
		done, ok := f.(rules.ItemIdentified)
		if !ok {
			continue
		}
		card := e.itemCard(id, h)
		if !done.Revealed {
			e.say("I can't tell anything new about your " + card.Name)
			return
		}
		e.snap.identN, e.snap.identIt = e.snap.identN+1, h // (extra) the reveal in a pop-up of its own (Shop.Identified)
		e.say("It looks like some sort of " + card.Name)
	}
}

// powers is what c can do besides items: the View menu's Spells, Magic,
// Heal and Cure (the inventory offers them too).
func (e *Engine) powers(id rules.CharID) []Power {
	var ps []Power
	if e.ruleset().Sheet(id).CanCast { // memorized, not still to be learned
		ps = append(ps, Power{Name: "Cast", Kind: PowerSpells})
	}
	if !e.effectsFit(id) {
		ps = append(ps, Power{Name: "Magic", Kind: PowerMagic}) // extra: all of the spell effects
	}
	if e.paladinHeals(id) {
		ps = append(ps, Power{Name: "Heal", Kind: PowerHeal})
	}
	if e.paladinCures(id) {
		ps = append(ps, Power{Name: "Cure", Kind: PowerCure})
	}
	return ps
}

// memberPowers (extra) is a card's context menu out of combat: View, Items
// when c carries any, then powers; in camp also what c does there
// (Memorize and Scribe open the presenter's lists, Drop its question).
func (e *Engine) memberPowers(id rules.CharID) []Power {
	if e.inFight() {
		return nil
	}
	ps := []Power{{Name: "View", Kind: PowerView}}
	if e.carryCount(id) > 0 {
		ps = append(ps, Power{Name: "Items", Kind: PowerItems})
	}
	ps = append(ps, e.powers(id)...)
	if !e.campOn() {
		return ps
	}
	study := e.fact(id, rules.StudyFact{}).(rules.Study)
	if study.Book > 0 {
		ps = append(ps, Power{Name: "Memorize", Kind: PowerPlan})
	}
	if study.Scrolls > 0 {
		ps = append(ps, Power{Name: "Scribe", Kind: PowerScribe})
	}
	return append(ps, Power{Name: "Icon", Kind: PowerIcon}, Power{Name: "Drop", Kind: PowerDrop})
}

type viewState struct {
	on    bool // the sheet is up (OpenSheet)
	leave bool // the sheet was opened on the inventory (F1-F8): leaving that leaves it
	chain bool // the game opened them (the loot, Browse): the cont on top of the stack goes on once both are closed
}

// portrait is id's portrait; false when id is no character.
func (e *Engine) portrait(id rules.CharID) (rules.Portrait, bool) {
	v, err := e.ruleset().Fact(id, rules.PortraitFact{})
	if err != nil {
		return rules.Portrait{}, false
	}
	return v.(rules.Portrait), true
}
