package rules

// Rules is a ruleset: 1st edition, Buck Rogers XXVc, ... It holds its
// characters and every number of its world (monster, item, spell tables, typed
// and converted at import, never read from a game's files at run time), and only
// answers; the engine owns the loop, the
// board and the script machine. It knows nothing of any Adventure.
//
// Every method carries the call in today's code that needs it. Everything is
// plain data both ways: no callbacks, nothing the rules could call back on.
type Rules interface {
	// Catalog is what this ruleset handles.
	Catalog() Catalog

	CharacterRules
	CombatRules

	// Begin sets what the adventure and the player chose: at start and at load.
	Begin(Campaign, Options)

	// Ships is the ship fights, nil without FeatureSpaceCombat.
	Ships() ShipRules

	// Allowed and Resolve are every other rule: creation, spells, camp,
	// training, shops, temples, loot, doors. Outside a Fight the situation
	// says where. Each play.State's Allowed; play.Game.Do.
	Allowed(Situation) Allowed
	Resolve(Situation, Action, *RNG) (Outcome, error) // error: refused, nothing changed

	// Elapse is time passing for the party: spells run out, poison works,
	// memorizing goes on. por stepRest (camp.go:1020), subRest, memTick (:967);
	// BR rest.go; pod clock.go. The engine keeps the clock words of the script.
	Elapse(p Party, minutes int32, r *RNG) Outcome

	// Save and Load are all of the ruleset's state: the characters. The engine
	// keeps script memory, the board and the RNG in its own part. Canonical
	// format only; originals come in once through the save importer.
	Save() ([]byte, error)
	Load([]byte) error // error: refused, nothing changed
}

// CharacterRules are the characters the ruleset holds.
type CharacterRules interface {
	// Spawn makes one monster of the ruleset's stat block for id, rolling what
	// varies. por addMonsters (eclops.go:206), AddMonsterToParty (files.go:355);
	// BR monster.go. Error: an id the Catalog lacks.
	Spawn(MonsterID, *RNG) (CharID, error)
	// Item is the ruleset's numbers for a placed item. por base table DS:5582, BR ITEMS.
	Item(ItemRef) (Item, error)
	// Dismiss forgets one. por opClearMonsters, RemoveSelected (eclops.go:191).
	Dismiss(CharID)
	// Sheet is its facts for the screens. por members (combatflow.go:470), playfacts.go; BR adv.go:618.
	Sheet(CharID) Sheet
	// Draft is the character being made, nil none. play.Assembling.Draft.
	Draft() *Draft
	// Purse is the coins a character can spend: its own, or the party's one
	// purse under the shared purse. play.Common.Purse; por purseCoin.
	Purse(CharID) Copper
	// Shared is the one slotted inventory under Options.PartyInventory, else nil.
	Shared() []Carried
	// Spell is the registry row: presence does not mean learnable (Campaign).
	Spell(SpellID) (Spell, error)
	// Fact and SetFact are a script's typed reads and writes. Neither has a side
	// effect: a consequence is an Action (ARCHITECTURE.md's table). A Fact the
	// character lacks is an error. por readSpecialGlobal, writeSpecialGlobal
	// (ecl.go:108, :192); cotab charValue, setCharValue (eclvars.go:110, :187);
	// BR charVar, setChar (ecl.go).
	Fact(CharID, Fact) (Value, error)
	SetFact(CharID, Fact, Value) error
	// Check answers a script's question about characters. por opCheckParty,
	// opPartyStrength, opPartySurprise (eclops.go:505..590); BR opSkill
	// (eclops.go:670); cotab FindAffect.
	Check(Check, *RNG) Answer
	// Import and Export carry a character between adventures, canonical format.
	Import(CharacterFile) (CharID, error)
	Export(CharID) (CharacterFile, error)
}

// CombatRules are a fight's sequence. The engine calls these in its loop:
//
//	BeginFight
//	for: BeginRound; for NextActor: BeginTurn; Allowed or Decide; Resolve ...; EndTurn; Verdict
//	     EndRound
//	EndFight
type CombatRules interface {
	// BeginFight: por combatSetup, CombatRules.CombatStart (combatsetup.go:12); BR startFight (combat.go:35).
	BeginFight(Fight, *RNG) Outcome
	// BeginRound: initiative for all: roll for every eligible, modifiers, a
	// deterministic tie-break, sort. Replaces por nextCombatant. BR newRound (combat.go:98).
	BeginRound(Fight, *RNG) Outcome
	// NextActor is who acts now, next in the order, delayed spells included;
	// false: the round is over. pod buildTurnOrder; BR nextTurn (combat.go:313).
	NextActor(Fight, *RNG) (CharID, bool)
	// BeginTurn: por takeTurn (combat.go), TurnStart; BR turnStart (combat.go:327).
	BeginTurn(Fight, CharID, *RNG) Turn
	// Decide is the computer's next action for its actor, called until it
	// returns EndTurn. por computerTurn (combatai.go:8), rules4e ai.go; BR ai.go.
	Decide(Fight, CharID, *RNG) Action
	// EndTurn: por endTurn (combatflow.go:186), TurnEnd; BR endTurn (combat.go:362).
	EndTurn(Fight, CharID, *RNG) Outcome
	// EndRound: BR endRound (combat.go:368); pod and por run effect time down here.
	EndRound(Fight, *RNG) Outcome
	// Verdict: por sidesCheck (combatflow.go:193); BR run on count (combat.go:86, :279).
	Verdict(Fight) Verdict
	// EndFight: por pushAfterCombat (aftercombat.go:406), giveExp (:89); BR endFight (combat.go:401).
	EndFight(Fight, *RNG) FightResult
}
