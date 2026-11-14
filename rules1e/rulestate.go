package rules1e

import (
	"gobox/rules"
)

// RuleState is the rules' state, the Go counterpart of the pure Pascal's
// gdvars (por-re src/pure): the party, the combatants and their places, the
// spell and attack scratch, the RNG seed and the effects the rules made
// (present.go). Rule functions take it; the Engine holds it beside the
// presentation, the script machine and the map (ARCHITECTURE.md step S).
// The combat board is its own value (Board).
type RuleState struct {
	Board                                                         // the combat board and the party's place on the map
	Tables                          *Tables                       // the game's rule tables; nil reads POR
	RandSeed                        uint32                        // DS:4428, Turbo Pascal RandSeed
	Party                           []*Character                  // DS:5D96 list head ...
	Pool                            []*Item                       // extra: the party inventory (pool.go)
	Who                             int                           // DS:5D92 the acting character: set per call from the caller's argument (Engine.as) or by the fight's turn
	Items                           []ItemBase                    // DS:5582: base item records by type
	Fx                              []rules.Effect                // what the rules showed, not yet presented (present.go)
	Loot                            Purse                         // DS:67F4 money found, taken after the fight
	UsingItem                       bool                          // DS:6DBF
	UsedScroll                      *Item                         // DS:6B80
	Rv                              ruleVars                      // DS:6814..682A rule globals
	MagicOn, CombatRound            byte                          // DS:6E30/6DE3
	ThrownCopy                      *Item                         // DS:5D9A
	spellTarget                     *Character                    // DS:6C95 target slot 1
	spellTargets                    [MaxCombatants + 1]*Character // DS:6C91+4i targets 2..; slot 1 is spellTarget
	NSpellTargets                   int                           // DS:6C94
	SpellCaster                     *Character                    // DS:6DB5
	CurSpell                        SpellID                       // DS:681B spell being cast
	spellKept                       bool                          // extra: the effect gave the spell back (Stone to Flesh on one not stoned, out of combat)
	areaSpell                       byte                          // DS:6820
	spellValue, spellValue2         byte                          // DS:4844/4845 spell scratch; DS:4845
	turnValue, modMoves, AffectVeto byte                          // DS:681A speed value, 681D, 681E affect scratch; DS:681D/681E
	fleeRoll                        byte                          // DS:6825 flee/surrender roll
	slotHits, SlotShots             [3]Count                      // DS:6822+i / 6E2C+i, i = 1..2 attack slots; DS:6E2C
	aiLastDir, aiBlocked, aiState3  byte                          // DS:443C..443E computer-turn movement state; DS:443D/443E
	aiOpen                          aiOpen                        // extra: aiPathDir's open list, kept to reuse its storage
	AttackedAlly                    byte                          // DS:6DE2
	ageCheckDue                     bool                          // SotSB b7D44 / PoD: someone aged in the fight: see who dies of old age (oldAgeDeaths)
	tiedTarget                      *Character                    // DS:6826 scratch target
	Settings                        Settings                      // extra: the options handed in (combatrules.go)
	BeginFlag                       byte                          // DS:49F3
	Game                            GameVars                      // DS:49D6 record (vars.go)
	Area                            AreaVars                      // DS:49D2
	Treasure                        []*Item                       // DS:6810 list: the items a fight or a script leaves to take
	loadouts                        map[*Character]*loadout       // extra: weapons to give back after the battle (restoreLoadouts)
	drawn                           *Character                    // extra: whose attacks this turn are ranged (Draw)
	Camp                            campState                     // ovr16/ovr21 globals
	Training                        bool                          // the training hall is open (OpenTraining)
	Learning                        bool                          // training gained a mage level: the new spell is chosen next (Learn)
	Purse                           Purse                         // the party's money
	Debug                           bool                          // DS:470C
	MinStats                        [6]int                        // "Minimal stats": per-stat floor, STR as st*100+str00
	EffTick                         [MaxCombatants + 1]Flag       // DS:471D: member i has timed effects to run down
	Fight                           FightState                    // the fight's rule part (combatflow.go)
	FoesHealth                      byte                          // DS:6E2F
}

// Board is the combat board: the combatants, their squares, the map, the
// bodies and the clouds, and the party's place on the area map the board
// starts from (ARCHITECTURE.md step B, B's partition).
type Board struct {
	Near               []NearEntry                  // DS:6716 combatants near a square
	nearList           []NearEntry                  // DS:671B copy kept by checkAffect
	CombatantSlots     []*Character                 // DS:65BD, combatant i (1-based) at [i-1]; filled by combat setup
	Targets            [MaxCombatants]Slot          // DS:6DE4, near list indexes kept by nearEnemies
	Cmap               *CombatMap                   // DS:6716
	Cpos               [MaxCombatants + 1]CombatPos // DS:5F27, combatant i at [i]
	Ncomb              int                          // DS:5F2A combatants on the map list
	Occ                [25][50]Slot                 // DS:60DB combatant index per square
	Bodies             [10]Body                     // DS:66D6, 1-based, count DS:6715
	Nbodies            int                          // DS:6715
	bodyIndex          byte                         // DS:70C7
	Place              placement                    // DS:4440..4658 combat start formations
	Terrain            Terrain                      // DS:465A wilderness square of the battle
	Clouds             []*cloud                     // DS:6DBB lingering spell areas
	Dungeon            bool                         // extra: the fight is in a dungeon, not outdoors (BuildGround)
	TgtX, TgtY         int                          // DS:6DB9/6DBA target square
	PosX, PosY, Facing int                          // DS:6AAD, 6AAE, 6AAF: the party's square and facing as BeginFight gives them (the Engine owns them)
}

// Dice = ovr25:0DE3 Dice(n, sides): sum of n rolls of 1..sides, kept in a
// byte like the original (wraps at 256).
func (e *RuleState) Dice(n, sides int) int { return rules.DiceSum(e, n, sides) }

// Random = Turbo Pascal Random(n) (0AF8:11C0): RandSeed = RandSeed*0x08088405+1,
// then the high word mod n. The seed advances even for n = 0, which returns 0.
func (e *RuleState) Random(n int) int { return rules.Random(&e.RandSeed, n) }

// MaxCombatants is the size of the original's combatant tables; tables
// indexed by combatant number (1-based) have one entry more.
const MaxCombatants = 0x48

// ErrNotNow is the refusal of a rule call the game does not allow now:
// core.ErrNotNow, so the Engine tests it through core.
var ErrNotNow = rules.ErrNotNow
