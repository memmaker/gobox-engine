package rules

// The fight: aims, the board, units, loot and the turn (what the combat screens and script commands ask of the rules).

// AtFact is a fighter's square on the fight's board (por CombatX, CombatY);
// its Value is a BoardPos.
type AtFact struct{}

func (AtFact) fact() {}

func (BoardPos) value() {}

// BankLoot: the fight's loot coins go into the party's purse (por afterShared:
// MoneySound, Purse.Add, Loot cleared).
type BankLoot struct{}

func (BankLoot) action() {}

// FightCleared: the fight is over and its rule state goes (por combatEnd:
// the computer's loadouts restored, fight state, clouds and combat map reset).
type FightCleared struct{}

func (FightCleared) action() {}

// DetectLoot: the actor casts its memorized Detect Magic on the fight's
// treasure (por LootDetect); ErrNotNow when there is no treasure or no such spell.
type DetectLoot struct{}

func (DetectLoot) action() {}

// IconFact is a fighter's combat icon slot (por Character.Order); its
// Value is a Number.
type IconFact struct{}

func (IconFact) fact() {}

// UnitFact is a fighter's unit number on the fight's board (por
// CombatIndex, 1-based; 0 off the board); its Value is a Number.
type UnitFact struct{}

// WeaponFact is the weapon a fighter holds ready: Handles, empty or one.
// Ranged (extra): the one its Ranged action shoots or throws instead.
type WeaponFact struct{ Ranged bool }

// WeaponRulesFact is the rule lines of the readied weapon (por
// WeaponRules); its Value is Lines, nil without a weapon. Ranged: as WeaponFact.
type WeaponRulesFact struct{ Ranged bool }

// TargetWhyFact is why the fighter cannot attack Target this turn, as the
// status row words it (por TargetWhy in weapon range); its Value is Text.
type TargetWhyFact struct{ Target CharID }

// SquaresFact is the squares a walk of Move squares reaches (por Squares),
// in board order; its Value is Squares.
type SquaresFact struct{ Move int }

type Lines []string

type Squares []BoardPos

func (UnitFact) fact() {}

func (WeaponFact) fact() {}

func (WeaponRulesFact) fact() {}

func (TargetWhyFact) fact() {}

func (SquaresFact) fact() {}

func (Lines) value() {}

func (Squares) value() {}

// ManeuverAim is what a maneuver aims at.
type ManeuverAim int32

const (
	AimWeapon ManeuverAim = iota // a foe, with the next weapon attack
	AimSelf                      // no aim
	AimFoe                       // a foe within Range
	AimAlly                      // a member of the fighter's side within Range (itself too)
	AimArea                      // a square within Range: who stands within Area squares of it
)

// ManeuverCard is one maneuver as the fighter's turn offers it.
type ManeuverCard struct {
	ID          ManeuverID
	Name, Tip   string
	Aim         ManeuverAim
	Range, Area int // squares: AimFoe, AimAlly, AimArea; AimArea's burst
	Move        int // AimWeapon: squares the fighter may move first
}

// ManeuversFact is the fighter's maneuvers; its Value is ManeuverCards.
type ManeuversFact struct{}

// ManeuverCards are maneuvers in the ruleset's order.
type ManeuverCards []ManeuverCard

func (ManeuversFact) fact() {}

func (ManeuverCards) value() {}

// Square is one square of the fight's ground as the battle view draws it.
type Square struct {
	Tile   int
	Wall   bool // nobody walks onto it
	Height int  // blocking height
}

// Step is one step of a walk: where it ends and whether a foe gets a free
// attack when the walker leaves it.
type Step struct {
	At   BoardPos
	Free bool
}

// MayActFact: the fighter may act against Target (por MayActAgainst); Yes.
type MayActFact struct{ Target CharID }

// RangeToFact is the fighter's distance to Target in squares (por RangeTo); Number.
type RangeToFact struct{ Target CharID }

// SizeFact is the fighter's footprint size on the board (por CombatSize); Number.
type SizeFact struct{}

// FightCardFact is the fighter's turn record (FightCard).
type FightCardFact struct{}

// FightCard is a fighter's turn record: Delay > 0 while it acts this
// round, Moves its half-squares left, Casting the spell number it began
// casting last round (0: none).
type FightCard struct {
	Delay, Moves, Casting int
	Auto                  bool // the computer runs it
}

// AimFact is what an aim of the fighter offers (por playerAim's opening):
// Range 0xFF takes the readied weapon's; FoesOnly offers foes only. AimOffer.
type AimFact struct {
	Range, ShowRange int
	FoesOnly         bool
}

// AimOffer: the resolved Range, the fighters in sight (Near, the order
// they were found in), and those the aim offers (Targets); Hits[i] is the
// chance in percent to hit Targets[i], -1 when not shown.
type AimOffer struct {
	Range   int
	Near    []CharID
	Targets []CharID
	Hits    []int
}

// AimCheckFact is aiming at At from Near (an AimOffer's) (por AimCheck); AimCheck.
type AimCheckFact struct {
	Near             []CharID
	At               BoardPos
	Range, ShowRange int
	Ground           bool // a bare square may be taken
}

// AimCheck: who stands at the square (NoChar: nobody, or a body's owner),
// why it cannot be taken ("" it can), and whether a fighter stands on it.
type AimCheck struct {
	Target   CharID
	Why      string
	Occupied bool
}

// BoardUp: the fight's board is laid (Yes).
type BoardUp struct{}

// GroundFact is the fight's ground (Ground); Who is NoChar.
type GroundFact struct{}

// Ground is the fight's 50x25 squares and the squares a footprint of each
// size covers (offsets from its top left).
type Ground struct {
	Squares    [25][50]Square
	Footprints [5]Squares
}

// UnitID is a fighter's board unit number (1-based, the original's combat
// index): never a CharID. A ruleset whose board numbers its units by
// CharID converts at the effect (rulesxxvc unitOf).
type UnitID int32

// UnitsFact is the fighters on the board in unit order (Units); Who is NoChar.
type UnitsFact struct{}

// UnitCharFact is who board unit Unit (1-based, an effect's Unit) is
// (UnitChar; Who NoChar: no such unit). The Fact's own id is NoChar.
type UnitCharFact struct{ Unit UnitID }

// UnitChar is UnitCharFact's answer.
type UnitChar struct{ Who CharID }

func (UnitCharFact) fact() {}

func (UnitChar) value() {}

// Unit is one fighter on the board.
type Unit struct {
	ID              CharID
	Number          int // unit number, 1-based
	At              BoardPos
	Size, Facing    int
	Party, Helpless bool
	Active          bool // not down
	Icon            int  // its combat icon slot
	Name            string
	HP, MaxHP, AC   int // AC as the sheet shows it
	Damage          string
	Squares         Squares
}

// Units are fighters in unit order.
type Units []Unit

// ClearCombat: the fighter's turn record is cleared (por ClearCombatInfo);
// Outcome.TurnOver is ClearCombatInfo's answer.
type ClearCombat struct{ Who CharID }

// CastingTaken: the spell the fighter began goes off now (por: the turn
// record's spell cleared).
type CastingTaken struct{ Who CharID }

func (MayActFact) fact() {}

func (RangeToFact) fact() {}

func (SizeFact) fact() {}

func (FightCardFact) fact() {}

func (AimFact) fact() {}

func (AimCheckFact) fact() {}

func (GroundFact) fact() {}

func (UnitsFact) fact() {}

func (BoardUp) check() {}

func (FightCard) value() {}

func (AimOffer) value() {}

func (AimCheck) value() {}

func (Ground) value() {}

func (Units) value() {}

func (ClearCombat) action() {}

func (CastingTaken) action() {}

// Settings are a ruleset's own settings a frontend hands in (por
// rules1e.Settings: the combat rule set, path search); the engine passes
// them on without reading them.
type Settings interface{ RulesetSettings() }

// Configure: the ruleset takes With as its settings from now on.
type Configure struct{ With Settings }

func (Configure) action() {}

// FightRound: the fight's round number (Value; por CombatRound).
type FightRound struct{}

func (FightRound) check() {}

// CombatSay is Who's line in the fight's text (NoChar: no name), a fresh
// line (por CombatSay).
type CombatSay struct {
	Who  CharID
	Text string
}

func (CombatSay) action() {}

// ClearTreasure empties the fight's items, the coins kept (the loot taken).
type ClearTreasure struct{}

// DevTreasure makes the treasure one random item of the ruleset's dev
// scene (por's dev loot: gauntlets).
type DevTreasure struct{}

// AddTreasureBlock puts the data set's item block Block at the head of the
// treasure, its last item leading (por opTreasure's ITEM block). Error:
// the data set lacks the block.
type AddTreasureBlock struct{ Block int32 }

// AddLoot adds Loot to the fight's money (TreasureOf; Felix's S ruling 4: its coins add).
// Money is gold and copper only (Felix 2026-11-12): a caller converts other coins by value.
type AddLoot struct{ Loot Purse }

// AddLibraryItems puts the data set's library items IDs at the head of the
// treasure in turn, so the last one leads (pod TREASURE, ITEM SET):
// RandomScrollItem makes a random scroll, an empty entry adds nothing.
// Known: their names are known. Error: the ruleset has no library, or no
// random scroll when one is asked for.
type AddLibraryItems struct {
	IDs   []int32
	Known bool
}

// RandomScrollItem is AddLibraryItems' id of a random scroll.
const RandomScrollItem int32 = 255

func (AddLoot) action() {}

func (AddLibraryItems) action() {}

// SetSpeed sets Who's attacks for the round from its weapon and affects (por SetSpeed).
type SetSpeed struct{ Who CharID }

func (ClearTreasure) action() {}

func (DevTreasure) action() {}

func (AddTreasureBlock) action() {}

func (SetSpeed) action() {}

// ClearLoot empties the fight's coins and items (por opClearMonsters).
type ClearLoot struct{}

// SetLoot replaces the fight's coins by the seven kinds of coin (por opTreasure).
type SetLoot struct{ Coins [7]uint16 }

// AddRandomTreasure puts N random items at the head of the treasure and
// names each treasure item for Who (por opTreasure's random form).
type AddRandomTreasure struct {
	N   int32
	Who CharID
}

// Rob takes Pct percent of the purse times the robbed share of the party,
// and each robbed member's items with chance ItemPct (por opRob).
type Rob struct {
	Robbed       Party
	Pct, ItemPct int32
}

// SetInactive makes Who unable to act (por DuelSetup: the other members sit out).
type SetInactive struct{ Who CharID }

// SetOrder sets Who's place in its monster group (por addMonsters: the icon).
type SetOrder struct {
	Who   CharID
	Order int32
}

func (ClearLoot) action() {}

func (SetLoot) action() {}

func (AddRandomTreasure) action() {}

func (Rob) action() {}

func (SetInactive) action() {}

func (SetOrder) action() {}

// InFight asks whether a fight is on (Yes).
type InFight struct{}

func (InFight) check() {}

// ArcDir asks the first direction (Value, 0..7) whose arc from Start holds To on the board.
type ArcDir struct{ Start, To BoardPos }

// UnitBody asks about a fallen combat Unit: Yes its body is shown on the
// ground, Value that ground's tile, Low and High the unit's square.
type UnitBody struct{ Unit UnitID }

func (ArcDir) check() {}

func (UnitBody) check() {}

// OpenTurn: the Fight's Actor's turn menu is shown (por combatMenu's
// opening, before its offer): Resolve runs what the menu runs first, then
// Allowed(Fight) lists the offer without changing anything.
type OpenTurn struct{ Actor CharID }

func (OpenTurn) action() {}

// StepFact is what lies a step of Dir (0..7) from the fighter on the fight's board (StepInfo).
type StepFact struct{ Dir int }

// StepInfo is a square a fighter may step on: Unit the combatant on it (0 none),
// Ground its terrain (0 and no unit: off the map), Foe whether that unit is
// on the other side.
type StepInfo struct {
	Unit, Ground int32
	Foe          bool
}

func (StepFact) fact()  {}
func (StepInfo) value() {}
