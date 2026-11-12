package pool

import (
	"fmt"
	"slices"
)

// Quest journal (extra): the commissions the party has been given
// (e.ext.quests, kept with the save), each with who gave it where and its
// giver's words from the original dialogue (brief), published as
// View.Quests. A quest is recorded when a script hands it over (Quest,
// the import's questFixes at each handover): most of the clerk's
// commissions set no flag when offered (area8 OfferLoop). It is also
// listed once a script flag says it was offered (given) or it is done.
// Done is the quest table the clerk reads
// (data/assets/scripts/names/global.txt, QuestTable 4AA6-4ABF: 254 reward
// due, 255 rewarded) or the word the quest's own scripts test.

// questID is a quest's id: its Quest value (names/quests.txt) and a save's questNote.
type questID int

// QuestDef is a quest (GameTables.Quests, the import's): its id (Quest's value), journal
// name (at most 19 letters: the list's width), giver and place, the giver's own words (the
// scripts' texts that hand it over), its done and its offered tests.
type QuestDef struct {
	ID                 int
	Name, Giver, Place string
	Brief              []string
	Done, Given        QuestTest
}

// QuestTest holds when any of its fields is at least its Min; empty never holds.
type QuestTest []FieldAtLeast

// FieldAtLeast is the game's saved field (layout/) at least Min.
type FieldAtLeast struct {
	Field string
	Min   int64
}

func (t QuestTest) holds(e *Engine) bool {
	return slices.ContainsFunc(t, func(f FieldAtLeast) bool { return e.savedWord(f.Field) >= f.Min })
}

// questNote is a quest the party was given.
type questNote struct {
	ID questID
}

// questByID is the quest of id.
func (e *Engine) questByID(id questID) (QuestDef, bool) {
	qs := e.game().Quests
	i := slices.IndexFunc(qs, func(q QuestDef) bool { return questID(q.ID) == id })
	if i < 0 {
		return QuestDef{}, false
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
		out = append(out, Quest{Name: q.Name, Giver: q.Giver, Place: q.Place, Text: q.Brief, Done: q.Done.holds(e)})
	}
	for _, q := range e.game().Quests { // the flags say so, not handed over (a save from before the journal)
		noted := slices.ContainsFunc(e.ext.quests, func(n questNote) bool { return n.ID == questID(q.ID) })
		if done := q.Done.holds(e); !noted && (done || q.Given.holds(e)) {
			out = append(out, Quest{Name: q.Name, Giver: q.Giver, Place: q.Place, Text: q.Brief, Done: done})
		}
	}
	return out
}
