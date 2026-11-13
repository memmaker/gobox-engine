package engine

import (
	"gobox/ecl/script"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestQuestGiven: a quest handed over makes one entry, once, with its
// giver's words.
func TestQuestGiven(t *testing.T) {
	e := &Engine{}
	e.adv = "../data/por/sources/assets" // the quest table's fields (layout/)
	e.initVM()
	e.questGiven(1)
	e.questGiven(4)
	e.questGiven(1)
	n := e.ext.quests
	if len(n) != 2 || n[0].ID != 1 || n[1].ID != 4 {
		t.Fatalf("got %+v", n)
	}
	if q := e.questsView(); len(q) != 2 || q[0].Name != "Cadorna's treasure" || len(q[0].Text) != 2 || q[1].Text[0] != "SOKAL KEEP ON THORN ISLAND MUST BE CLEARED." {
		t.Fatalf("journal %+v", q)
	}
}

// TestQuestSitesNamed (bug class: a quest recorded by matching printed
// text): names/quests.txt names exactly the quests of gametables.json, and every
// quest is handed over by a Quest command of the imported scripts.
func TestQuestSitesNamed(t *testing.T) {
	vals, err := script.ReadValues(os.DirFS("../data/por/sources/assets/scripts"), "names/quests.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	named := map[questID]string{}
	for _, v := range vals {
		if _, ok := (&Engine{}).questByID(questID(v.N)); v.Kind != "Quest" || !ok {
			t.Errorf("quests.txt: %+v is no quest", v)
		}
		named[questID(v.N)] = v.Name
	}
	files, _ := filepath.Glob("../data/por/sources/assets/scripts/area*.txt")
	if len(files) == 0 {
		t.Skip("no imported scripts")
	}
	var all strings.Builder
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		all.Write(b)
	}
	for _, q := range porTables.Quests {
		if n, ok := named[questID(q.ID)]; !ok || !strings.Contains(all.String(), "Quest("+n+")") {
			t.Errorf("quest %d %s (%q): no Quest command", q.ID, q.Name, n)
		}
	}
}

// A quest its script flags say was offered, or is done, is in the journal
// without its text seen (a save from before the journal); done follows the
// quest table.
func TestQuestFromFlags(t *testing.T) {
	e := quickStarted(t)
	if q := e.questsView(); len(q) != 0 {
		t.Fatalf("a new party's journal: %+v", q)
	}
	e.setSavedWord("BishopSummons", 254)
	if q := e.questsView(); len(q) != 1 || q[0].Name != "Temple of Bane" || q[0].Done || len(q[0].Text) != 1 {
		t.Fatalf("summoned by the bishop: %+v", q)
	}
	e.setSavedWord("QuestTempleOfBane", 255)
	e.setSavedWord("Boss", 254)
	if q := e.questsView(); len(q) != 2 || !q[0].Done || q[1].Name != "The poisoned river" || !q[1].Done {
		t.Fatalf("both done: %+v", q)
	}
}

// Each quest's words are the original dialogue's, verbatim: every line is
// a text of the scripts (data/assets/scripts).
func TestQuestBriefsVerbatim(t *testing.T) {
	files, _ := filepath.Glob("../data/por/sources/assets/scripts/area*.txt")
	if len(files) == 0 {
		t.Skip("no scripts")
	}
	var all strings.Builder
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		all.Write(b)
	}
	for _, q := range porTables.Quests {
		if len(q.Brief) == 0 {
			t.Errorf("%s: no words", q.Name)
		}
		for _, l := range q.Brief {
			if !strings.Contains(all.String(), l) {
				t.Errorf("%s: %q is no script text", q.Name, l)
			}
		}
	}
}
