package games

import "gobox/engine"

// porQuests are the quest journal's quests (gametables.json Quests): the clerk's
// commissions (area8 OfferLoop), Cadorna's, Urslingen's and the council's chambers, area0
// the bishop's study. Done is the quest table the clerk reads (QuestTable 4AA6-4ABF: 254
// reward due, 255 rewarded) or the word the quest's own scripts test.
const clerk, cityHall = "The council clerk", "City Hall, New Phlan"

// quests are the commissions of the scripts (area8 OfferLoop, Cadorna's,
// Urslingen's and the council's chambers; area0 the bishop's study).
var porQuests = []engine.QuestDef{
	{ID: 0, Name: "Clear the slums", Giver: clerk, Place: cityHall, Brief: []string{"THE SLUMS IMMEDIATELY TO OUR WEST NEED TO BE CLEARED OF MONSTERS."}, Done: engine.QuestTest{{Field: "QuestSlums", Min: 254}}},
	{ID: 1, Name: "Sokal Keep", Giver: clerk, Place: cityHall, Brief: []string{"SOKAL KEEP ON THORN ISLAND MUST BE CLEARED."}, Done: engine.QuestTest{{Field: "QuestSokalKeep", Min: 254}}},
	{ID: 2, Name: "Lore of old Phlan", Giver: clerk, Place: cityHall, Brief: []string{"THE COUNCIL IS OFFERING A REWARD FOR BOOKS, MAPS, TOMES, ETC. WHICH PROVIDE USEFUL INFORMATION ABOUT PHLAN BEFORE THE FALL.  THE REWARD IS TIED TO THE VALUE OF THE INFORMATION."}, Done: engine.QuestTest{{Field: "LibraryRewardGiven", Min: 255}}},
	{ID: 3, Name: "Podal Plaza auction", Giver: clerk, Place: cityHall, Brief: []string{"A WEAPON OF GREAT POWER IS TO BE AUCTIONED TO OUR ENEMIES. THIS AUCTION IS TO BE HELD IN PODAL PLAZA.  FIND OUT WHAT THE WEAPON IS AND RETURN."}, Done: engine.QuestTest{{Field: "QuestPodalAuction", Min: 254}}, Given: engine.QuestTest{{Field: "QuestPodalAuction", Min: 1}}}, // 1 AuctionOn: offered
	{ID: 4, Name: "Cadorna's treasure", Giver: "Junior Councilman Cadorna", Place: "His chambers, east of the City Hall clerk",
		Brief: []string{"WHEN OLD PHLAN WAS OVERRUN, A FAMILY TREASURE WAS HIDDEN IN THE WESTERN BUILDING OF THE TEXTILE COMPLEX.", "THE FAITHFUL SERVANT SENT TO FETCH IT NEVER RETURNED.  THE COMPLEX IS JUST SOUTH OF PODOL PLAZA.  BRING THE TREASURE TO ME AND YOU WILL BE WELL REWARDED."},
		Done:  engine.QuestTest{{Field: "QuestCadornaPayment", Min: 254}, {Field: "QuestCadornaTraitor", Min: 254}}, // paid (area0, area8: IronBoxState HandedIn goes with it) or Cadorna exposed (area4)
		Given: engine.QuestTest{{Field: "CadornaCommission", Min: 254}}},                                            // CadornaCommission: offered by the clerk
	{ID: 5, Name: "Temple of Bane", Giver: "Bishop Braccio of Tyr", Place: "The Temple of Tyr, New Phlan",
		Brief: []string{"HE IS BOUND TO RECOVER THE TEMPLE WHICH HAS BEEN DESECRATED TO BANE.  GO WITH HIM ACROSS THE RIVER AND HELP HIM TO CLEANSE THE TEMPLE.  IN PAYMENT, YOU MAY KEEP THE HIDDEN TREASURES."}, Done: engine.QuestTest{{Field: "QuestTempleOfBane", Min: 254}}, Given: engine.QuestTest{{Field: "BishopSummons", Min: 254}}}, // BishopSummons
	{ID: 6, Name: "Kovel Mansion", Giver: clerk, Place: cityHall, Brief: []string{"A LARGE GROUP OF THIEVES OPERATES OUT OF THE OLD KOVEL MANSION. THE COUNCIL WILL PAY TO HAVE THEM CLEANED OUT."}, Done: engine.QuestTest{{Field: "QuestKovelThieves", Min: 254}}},
	{ID: 7, Name: "The nomads", Giver: clerk, Place: cityHall, Brief: []string{"FIND THE NOMADS AND STOP THEM FROM JOINING FORCES WITH OUR ENEMIES."}, Done: engine.QuestTest{{Field: "QuestNomads", Min: 254}}},
	{ID: 8, Name: "The kobold tribe", Giver: clerk, Place: cityHall, Brief: []string{"A LARGE TRIBE OF KOBOLDS IS BEING RECRUITED BY THE ENEMY.  MAKE SURE THEY DON'T JOIN THE ENEMY FORCES."}, Done: engine.QuestTest{{Field: "QuestKobolds", Min: 254}}},
	{ID: 9, Name: "The poisoned river", Giver: clerk, Place: cityHall, Brief: []string{"FIND THE SOURCE OF THE RIVER'S POLLUTION AND END IT."}, Done: engine.QuestTest{{Field: "Boss", Min: 254}}}, // 4AB3: Boss (vars.go), 254 Yarash beaten (area23)
	{ID: 10, Name: "The lizardmen", Giver: clerk, Place: cityHall, Brief: []string{"A TRIBE OF LIZARDMEN IS PREPARING TO JOIN THE ENEMY.  PREVENT THIS."}, Done: engine.QuestTest{{Field: "QuestLizardmen", Min: 254}}},
	{ID: 11, Name: "The heir of Bivant", Giver: clerk, Place: cityHall, Brief: []string{"THE HEIR TO THE HOUSE OF BIVANT MUST BE RESCUED.  WE WILL PAY GENEROUSLY FOR HIS SAFE RETURN."}, Done: engine.QuestTest{{Field: "QuestBivant", Min: 128}}}, // 128 bungled, 254 rescued (area1): over either way
	{ID: 12, Name: "Zhentil envoys", Giver: "Lord Cadorna", Place: "His chambers, south of the City Hall clerk",
		Brief: []string{"IT IS IMPERATIVE THAT A MESSAGE BE DELIVERED TO THE ZHENTIL OUTPOST TO THE WEST.  I AM SENDING YOU AS DIPLOMATIC ENVOYS."},
		Done:  engine.QuestTest{{Field: "QuestZhentilOutpost", Min: 1}}, Given: engine.QuestTest{{Field: "ZhentilMission", Min: 254}}}, // area28's own test: 1 commandant killed, 253 left, 254 escorted (area25); ZhentilMission
	{ID: 13, Name: "Stojanow Gate", Giver: "Lord Urslingen", Place: "Behind Cadorna's old quarters, City Hall",
		Brief: []string{"THANK YOU FOR COMING.  I MUST BE BRIEF FOR OUR NEED IS URGENT.  OUR ONE CHANCE TO DEFEAT THE ENEMY IS THROUGH AN ASSAULT ON THE CASTLE. TO DO THIS, WE MUST FIRST SECURE STOJANOW GATE."}, Done: engine.QuestTest{{Field: "QuestGateTaken", Min: 254}}},
	{ID: 14, Name: "Valjevo Castle", Giver: "Lord High Mayor Ulrich Eberhard", Place: "The council chambers, City Hall",
		Brief: []string{"IT IS THE UNANIMOUS REQUEST OF THE COUNCIL THAT YOU LEAD THE ATTACK. YOUR SKILL AND BRAVERY HAVE WON THE ADMIRATION OF ALL PHLAN.  WE LOOK TO YOU TO HELP US WIN OUR FREEDOM."}, Done: engine.QuestTest{{Field: "QuestTyranthraxus", Min: 254}}},
	{ID: 15, Name: "Valhingen Graveyard", Giver: clerk, Place: cityHall, Brief: []string{"THE UNDEAD ARE CROSSING THE RIVER IN HORDES AND KILLING OR CARRYING OFF OUR SETTLERS.  THE COUNCIL DEMANDS THESE ATTACKS BE STOPPED. A GENEROUS REWARD IS OFFERED."}, Done: engine.QuestTest{{Field: "QuestGraveyard", Min: 254}}},
}
