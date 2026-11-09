package core

// Rules is a ruleset: 1st edition, Buck Rogers XXVc, ... It holds its
// characters and its dice, and only answers; the engine owns the loop, the
// board and the script machine. It knows nothing of any Adventure.
//
// Every method carries the call in today's code that needs it. Everything is
// plain data both ways: no callbacks, nothing the rules could call back on.
type Rules interface {
	// Catalog is what this ruleset handles, for Mismatches at load.
	Catalog() Catalog

	// Random is 0..n-1 of the one dice stream, so that a script's RANDOM and
	// the rules' rolls keep the original's order. ecl.Host.Random; por eclHost.
	Random(n int) int

	CharacterRules
	CombatRules

	// Allowed and Resolve are every other rule: creation, spells, camp,
	// training, shops, temples, loot, doors. Outside a Fight the situation
	// says where. Each play.State's Allowed; play.Game.Do.
	Allowed(Situation) Allowed
	Resolve(Situation, Action) (Outcome, error) // error: refused, nothing changed

	// Elapse is time passing for the party: spells run out, poison works,
	// memorizing goes on. por stepRest (camp.go:1020), subRest, memTick (:967);
	// BR rest.go; pod clock.go. The engine keeps the clock words of the script.
	Elapse(p Party, minutes int32) Outcome

	// Save and Load are all of the ruleset's state: the characters and the
	// dice. play.Game.Save, Load. The engine keeps script memory and the
	// board's absence in its own part (ARCHITECTURE.md on save files).
	Save() ([]byte, error)
	Load([]byte) error // error: refused, nothing changed
}

// CharacterRules are the characters the ruleset holds.
type CharacterRules interface {
	// Spawn makes a character from an adventure's Creature. por addMonsters
	// (eclops.go:206), AddMonsterToParty (files.go:355); BR monster.go.
	Spawn(Creature) (CharID, error)
	// Dismiss forgets one. por opClearMonsters, RemoveSelected (eclops.go:191).
	Dismiss(CharID)
	// Sheet is its facts for the screens. por members (combatflow.go:470), playfacts.go; BR adv.go:618.
	Sheet(CharID) Sheet
	// Draft is the character being made, nil none. play.Assembling.Draft.
	Draft() *Draft
	// Purse is the party's one purse. play.Common.Purse; por purseCoin.
	Purse() Purse
	// Fact and SetFact are a script's reads and writes of the selected
	// character. por readSpecialGlobal, writeSpecialGlobal (ecl.go:108, :192);
	// cotab charValue, setCharValue (eclvars.go:110, :187); BR charVar, setChar
	// (ecl.go). A write may have effects (status 0x87 stones it).
	Fact(CharID, Fact) (Value, error)
	SetFact(CharID, Fact, Value) ([]Effect, error)
	// Check answers a script's question about characters. por opCheckParty,
	// opPartyStrength, opPartySurprise (eclops.go:505..590); BR opSkill
	// (eclops.go:670); cotab FindAffect.
	Check(Check) Answer
	// Import and Export carry a character over files. pod, cotab, sotsb
	// charimport.go, charfiles.go; game.Data.Imports; play.AddCharacter.From.
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
	BeginFight(Fight) Outcome
	// BeginRound: initiative. por turnSetup (combatturn1.go:18); pod and sotsb
	// buildTurnOrder (combat.go:75, :74); BR newRound (combat.go:98).
	BeginRound(Fight) Outcome
	// NextActor is who acts now; false: the round is over. por nextCombatant
	// (combat.go:32) rolls here, pod walks its sorted order, BR nextTurn (combat.go:313).
	NextActor(Fight) (CharID, bool)
	// BeginTurn: por takeTurn (combat.go), TurnStart; BR turnStart (combat.go:327).
	BeginTurn(Fight, CharID) Turn
	// Decide is the computer's next action for its actor, called until it
	// returns EndTurn. por computerTurn (combatai.go:8), rules4e ai.go; BR ai.go.
	Decide(Fight, CharID) Action
	// EndTurn: por endTurn (combatflow.go:186), TurnEnd; BR endTurn (combat.go:362).
	EndTurn(Fight, CharID) Outcome
	// EndRound: BR endRound (combat.go:368); pod and por run effect time down here.
	EndRound(Fight) Outcome
	// Verdict: por sidesCheck (combatflow.go:193); BR run on count (combat.go:86, :279).
	Verdict(Fight) Verdict
	// EndFight: por pushAfterCombat (aftercombat.go:406), giveExp (:89); BR endFight (combat.go:401).
	EndFight(Fight) FightResult
}
