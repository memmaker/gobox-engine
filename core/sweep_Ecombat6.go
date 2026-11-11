package core

// Additions of sweep Ecombat6 (fold E-R6): por's combat end.

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
type WeaponFact struct{}

// WeaponRulesFact is the rule lines of the readied weapon (por
// WeaponRules); its Value is Lines, nil without a weapon.
type WeaponRulesFact struct{}

// TargetWhyFact is why the fighter cannot attack Target this turn, as the
// status row words it (por TargetWhy in weapon range); its Value is Text.
type TargetWhyFact struct{ Target CharID }

// SquaresFact is the squares a walk of Move squares reaches (por Squares),
// in board order; its Value is Squares.
type SquaresFact struct{ Move int }

// Lines are lines of text in order; Squares are board squares in order.
type (
	Lines   []string
	Squares []BoardPos
)

func (UnitFact) fact()        {}
func (WeaponFact) fact()      {}
func (WeaponRulesFact) fact() {}
func (TargetWhyFact) fact()   {}
func (SquaresFact) fact()     {}
func (Lines) value()          {}
func (Squares) value()        {}
