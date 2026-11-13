package engine

import (
	"fmt"
	"gobox/adventure"
	"slices"
)

// Quest journal (extra): the commissions the party has been given
// (e.ext.quests, kept with the save), each with who gave it where and its
// giver's words from the original dialogue (brief), published as
// Quests. A quest is recorded when a script hands it over (Quest,
// the import's questFixes at each handover): most of the clerk's
// commissions set no flag when offered (area8 OfferLoop). It is also
// listed once a script flag says it was offered (given) or it is done.
// Done is the quest table the clerk reads
// (data/assets/scripts/names/global.txt, QuestTable 4AA6-4ABF: 254 reward
// due, 255 rewarded) or the word the quest's own scripts test.

// questID is a quest's id: its Quest value (names/quests.txt) and a save's questNote.
type questID int

// questHolds is whether any of t's fields is at least its Min.
func (e *Engine) questHolds(t adventure.QuestTest) bool {
	return slices.ContainsFunc(t, func(f adventure.FieldAtLeast) bool { return e.savedWord(f.Field) >= f.Min })
}

// questNote is a quest the party was given.
type questNote struct {
	ID questID
}

// questByID is the quest of id.
func (e *Engine) questByID(id questID) (adventure.QuestDef, bool) {
	qs := e.game().Quests
	i := slices.IndexFunc(qs, func(q adventure.QuestDef) bool { return questID(q.ID) == id })
	if i < 0 {
		return adventure.QuestDef{}, false
	}
	return qs[i], true
}

// opQuest is Quest(quest) (gobox ecl, the port's): a script hands the quest over.
func (e *Engine) opQuest() {
	e.vm.ReadArgs(1)
	e.questGiven(questID(e.vm.Arg(1, e.host)))
}

// questGiven records quest id as given, once.
func (e *Engine) questGiven(id questID) {
	if _, ok := e.questByID(id); !ok {
		panic(fmt.Sprintf("script: no quest %d (names/quests.txt)", id)) // a script naming no quest: a broken game folder
	}
	if !slices.ContainsFunc(e.ext.quests, func(n questNote) bool { return n.ID == id }) {
		e.ext.quests = append(e.ext.quests, questNote{ID: id})
		e.say("Quest added to journal")
	}
}

// questsView is the journal: the quests given, newest first.
func (e *Engine) questsView() []Quest {
	var out []Quest
	for _, n := range slices.Backward(e.ext.quests) {
		q, ok := e.questByID(n.ID)
		if !ok {
			continue
		}
		out = append(out, Quest{Name: q.Name, Giver: q.Giver, Place: q.Place, Text: q.Brief, Done: e.questHolds(q.Done)})
	}
	for _, q := range e.game().Quests { // the flags say so, not handed over (a save from before the journal)
		noted := slices.ContainsFunc(e.ext.quests, func(n questNote) bool { return n.ID == questID(q.ID) })
		if done := e.questHolds(q.Done); !noted && (done || e.questHolds(q.Given)) {
			out = append(out, Quest{Name: q.Name, Giver: q.Giver, Place: q.Place, Text: q.Brief, Done: done})
		}
	}
	return out
}
