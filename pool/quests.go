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

// questDef is a quest: its id, journal name (at most 19 letters: the list's
// width), giver and place, the giver's words, its done test and its offered
// test (nil: no flag).
type questDef struct {
	id                 questID
	name, giver, place string
	brief              []string // the giver's own words: the scripts' texts that hand the quest over
	done               func(e *Engine) bool
	given              func(e *Engine) bool
}

// varAtLeast: the game's field name (layout/) is at least v.
func varAtLeast(name string, v int64) func(e *Engine) bool {
	return func(e *Engine) bool { return e.savedWord(name) >= v }
}

// questOver: the quest table's field a says the deed is
// done (254 reward due, 255 rewarded).
func questOver(a string) func(e *Engine) bool {
	return varAtLeast(a, 254)
}

const clerk, cityHall = "The council clerk", "City Hall, New Phlan"

// quests are the commissions of the scripts (area8 OfferLoop, Cadorna's,
// Urslingen's and the council's chambers; area0 the bishop's study).
var quests = []questDef{
	{0, "Clear the slums", clerk, cityHall, []string{"THE SLUMS IMMEDIATELY TO OUR WEST NEED TO BE CLEARED OF MONSTERS."}, questOver("QuestSlums"), nil},
	{1, "Sokal Keep", clerk, cityHall, []string{"SOKAL KEEP ON THORN ISLAND MUST BE CLEARED."}, questOver("QuestSokalKeep"), nil},
	{2, "Lore of old Phlan", clerk, cityHall, []string{"THE COUNCIL IS OFFERING A REWARD FOR BOOKS, MAPS, TOMES, ETC. WHICH PROVIDE USEFUL INFORMATION ABOUT PHLAN BEFORE THE FALL.  THE REWARD IS TIED TO THE VALUE OF THE INFORMATION."}, varAtLeast("LibraryRewardGiven", 255), nil},
	{3, "Podal Plaza auction", clerk, cityHall, []string{"A WEAPON OF GREAT POWER IS TO BE AUCTIONED TO OUR ENEMIES. THIS AUCTION IS TO BE HELD IN PODAL PLAZA.  FIND OUT WHAT THE WEAPON IS AND RETURN."}, questOver("QuestPodalAuction"), varAtLeast("QuestPodalAuction", 1)}, // 1 AuctionOn: offered
	{4, "Cadorna's treasure", "Junior Councilman Cadorna", "His chambers, east of the City Hall clerk",
		[]string{"WHEN OLD PHLAN WAS OVERRUN, A FAMILY TREASURE WAS HIDDEN IN THE WESTERN BUILDING OF THE TEXTILE COMPLEX.", "THE FAITHFUL SERVANT SENT TO FETCH IT NEVER RETURNED.  THE COMPLEX IS JUST SOUTH OF PODOL PLAZA.  BRING THE TREASURE TO ME AND YOU WILL BE WELL REWARDED."},
		func(e *Engine) bool { // paid (area0, area8: IronBoxState HandedIn goes with it) or Cadorna exposed (area4)
			return e.savedWord("QuestCadornaPayment") >= 254 || e.savedWord("QuestCadornaTraitor") >= 254
		}, varAtLeast("CadornaCommission", 254)}, // CadornaCommission: offered by the clerk
	{5, "Temple of Bane", "Bishop Braccio of Tyr", "The Temple of Tyr, New Phlan",
		[]string{"HE IS BOUND TO RECOVER THE TEMPLE WHICH HAS BEEN DESECRATED TO BANE.  GO WITH HIM ACROSS THE RIVER AND HELP HIM TO CLEANSE THE TEMPLE.  IN PAYMENT, YOU MAY KEEP THE HIDDEN TREASURES."}, questOver("QuestTempleOfBane"), varAtLeast("BishopSummons", 254)}, // BishopSummons
	{6, "Kovel Mansion", clerk, cityHall, []string{"A LARGE GROUP OF THIEVES OPERATES OUT OF THE OLD KOVEL MANSION. THE COUNCIL WILL PAY TO HAVE THEM CLEANED OUT."}, questOver("QuestKovelThieves"), nil},
	{7, "The nomads", clerk, cityHall, []string{"FIND THE NOMADS AND STOP THEM FROM JOINING FORCES WITH OUR ENEMIES."}, questOver("QuestNomads"), nil},
	{8, "The kobold tribe", clerk, cityHall, []string{"A LARGE TRIBE OF KOBOLDS IS BEING RECRUITED BY THE ENEMY.  MAKE SURE THEY DON'T JOIN THE ENEMY FORCES."}, questOver("QuestKobolds"), nil},
	{9, "The poisoned river", clerk, cityHall, []string{"FIND THE SOURCE OF THE RIVER'S POLLUTION AND END IT."}, questOver("Boss"), nil}, // 4AB3: Boss (vars.go), 254 Yarash beaten (area23)
	{10, "The lizardmen", clerk, cityHall, []string{"A TRIBE OF LIZARDMEN IS PREPARING TO JOIN THE ENEMY.  PREVENT THIS."}, questOver("QuestLizardmen"), nil},
	{11, "The heir of Bivant", clerk, cityHall, []string{"THE HEIR TO THE HOUSE OF BIVANT MUST BE RESCUED.  WE WILL PAY GENEROUSLY FOR HIS SAFE RETURN."}, varAtLeast("QuestBivant", 128), nil}, // 128 bungled, 254 rescued (area1): over either way
	{12, "Zhentil envoys", "Lord Cadorna", "His chambers, south of the City Hall clerk",
		[]string{"IT IS IMPERATIVE THAT A MESSAGE BE DELIVERED TO THE ZHENTIL OUTPOST TO THE WEST.  I AM SENDING YOU AS DIPLOMATIC ENVOYS."},
		varAtLeast("QuestZhentilOutpost", 1), varAtLeast("ZhentilMission", 254)}, // area28's own test: 1 commandant killed, 253 left, 254 escorted (area25); ZhentilMission
	{13, "Stojanow Gate", "Lord Urslingen", "Behind Cadorna's old quarters, City Hall",
		[]string{"THANK YOU FOR COMING.  I MUST BE BRIEF FOR OUR NEED IS URGENT.  OUR ONE CHANCE TO DEFEAT THE ENEMY IS THROUGH AN ASSAULT ON THE CASTLE. TO DO THIS, WE MUST FIRST SECURE STOJANOW GATE."}, questOver("QuestGateTaken"), nil},
	{14, "Valjevo Castle", "Lord High Mayor Ulrich Eberhard", "The council chambers, City Hall",
		[]string{"IT IS THE UNANIMOUS REQUEST OF THE COUNCIL THAT YOU LEAD THE ATTACK. YOUR SKILL AND BRAVERY HAVE WON THE ADMIRATION OF ALL PHLAN.  WE LOOK TO YOU TO HELP US WIN OUR FREEDOM."}, questOver("QuestTyranthraxus"), nil},
	{15, "Valhingen Graveyard", clerk, cityHall,
		[]string{"THE UNDEAD ARE CROSSING THE RIVER IN HORDES AND KILLING OR CARRYING OFF OUR SETTLERS.  THE COUNCIL DEMANDS THESE ATTACKS BE STOPPED. A GENEROUS REWARD IS OFFERED."}, questOver("QuestGraveyard"), nil},
}

// questNote is a quest the party was given.
type questNote struct {
	ID questID
}

// questByID is the quest of id.
func questByID(id questID) (questDef, bool) {
	i := slices.IndexFunc(quests, func(q questDef) bool { return q.id == id })
	if i < 0 {
		return questDef{}, false
	}
	return quests[i], true
}

// opQuest is Quest(quest) (gobox ecl, the port's): a script hands the quest over.
func (e *Engine) opQuest() {
	e.vm.ReadArgs(1)
	e.questGiven(questID(e.vm.Arg(1, e.host)))
}

// questGiven records quest id as given, once.
func (e *Engine) questGiven(id questID) {
	if _, ok := questByID(id); !ok {
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
		q, ok := questByID(n.ID)
		if !ok {
			continue
		}
		out = append(out, Quest{Name: q.name, Giver: q.giver, Place: q.place, Text: q.brief, Done: q.done(e)})
	}
	for _, q := range quests { // the flags say so, not handed over (a save from before the journal)
		noted := slices.ContainsFunc(e.ext.quests, func(n questNote) bool { return n.ID == q.id })
		if done := q.done(e); !noted && (done || q.given != nil && q.given(e)) {
			out = append(out, Quest{Name: q.name, Giver: q.giver, Place: q.place, Text: q.brief, Done: done})
		}
	}
	return out
}
