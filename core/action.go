package core

// Action is one thing a character or the party does. The set is closed. These
// are the rule verbs of gobox/game/play: at migration play's become aliases
// of them, never a second copy.
type Action interface{ action() }

// Offer is an action allowed now: what it costs (shops, temples, training)
// and, on an action offered but refused, why not.
type Offer struct {
	Action Action
	Price  uint64
	Coin   Coin
	Not    string
	Name   string // its label, in the rules' words ("Cure Light Wounds")
}

// Aimed is an action that takes squares (play.Aimed).
type Aimed struct {
	Action  Action
	Squares []BoardPos
	Targets int
	Area    int
}

// Allowed is what a situation lets anyone do now (play.Allowed). Moves is the
// squares the actor may walk to, with the path, computed on Fight.Board.
type Allowed struct {
	Offers []Offer
	Aimed  []Aimed
	Moves  []Reach
}

// Reach is a square the actor can walk to.
type Reach struct {
	At   BoardPos
	Path []BoardPos
}

type (
	// In a fight (the actor is the Fight's).
	Move   struct{ To BoardPos }
	Attack struct {
		Target CharID
		At     *BoardPos
	}
	Shoot struct{ Ats []BoardPos }
	Cast  struct {
		Caster CharID
		Spell  SpellID
		Target CharID
		Ats    []BoardPos
		Aims   []Aim // the aims the player took, in the order the spell asks for them
		Begun  bool  // the spell begun in an earlier turn goes off
	}
	UseItem struct {
		User   CharID
		Item   ItemSlot
		Held   ItemHandle // the item, for a ruleset that hands out handles (Item unused)
		Target CharID
		Ats    []BoardPos
		Aims   []Aim   // as Cast's
		Spell  SpellID // a scroll's spell
	}
	Maneuver struct {
		ID     ManeuverID
		Ats    []BoardPos
		Target CharID // the one aimed at (NoChar: none)
		// Cancel: an armed maneuver's attack was not taken; it is disarmed and
		// its move undone (Ats[0]: where the actor began).
		Cancel bool
	}
	TurnUndead struct{}
	Guard      struct{}
	Delay      struct{}
	Quick      struct{ All bool }
	Flee       struct{}
	EndTurn    struct{}
	Aid        struct{ Target CharID } // BR
	Sprint     struct{}                // BR
	Dodge      struct{}                // BR
	Intimidate struct{}                // BR
	Jetpack    struct{ To BoardPos }   // BR
	// AutoTurn is a whole computer turn the rules play at once (por
	// computerTurn); Key is the player's key its polls take, the engine's input.
	AutoTurn struct{ Key AutoKey }
	// FightPlaced: the engine has built the board and placed everyone
	// (after BeginFight); the rules' start-of-fight effects run.
	FightPlaced struct{}
	// PlaceFight: the engine has built the board (after BeginFight); the rules
	// give everyone a combat block and place them on it (por combatBlocks and
	// placeCombatants). Walls nil: the board is open country, no walls between
	// the team blocks. One that finds no room Left.
	PlaceFight struct{ Walls MapSides }

	// Confirmed sends Action again after the rules' Outcome.Confirm got a
	// yes (por: "its Yes makes the call again sure", combatflow.go).
	Confirmed struct{ Action Action }

	// Goods, anywhere.
	Ready struct {
		Who   CharID
		Item  ItemSlot
		Count int32
	}
	Unready struct {
		Who  CharID
		Item ItemSlot
	}
	DropItem struct {
		Who  CharID
		Item ItemSlot
	}
	GiveItem struct {
		From, To CharID
		Item     ItemSlot
	}
	Halve struct {
		Who  CharID
		Item ItemSlot
	}
	JoinStack struct {
		Who  CharID
		Item ItemSlot
	}

	// Character creation (Creating, and Party for NewCharacter and Modify).
	NewCharacter struct{}
	Modify       struct{ Who CharID }
	ChooseRace   struct{ Race Race }
	ChooseClass  struct{ Classes ClassSet }
	ChooseSex    struct{ Sex Sex }
	ChooseAlign  struct{ Alignment Alignment }
	SetScore     struct {
		Ability Ability
		To      int32
	}
	PointBuy    struct{ Scores map[Ability]int32 }
	Reroll      struct{ Minimal bool }
	PickSkill   struct{ Skill Skill } // BR's skills step
	Rename      struct{ Name string }
	KeepDraft   struct{}
	DropDraft   struct{}
	ChangeClass struct {
		Who CharID
		To  Class
	} // dual class

	// Camp.
	Memorize struct {
		Who   CharID
		Spell SpellID
	}
	Forget struct {
		Who   CharID
		Spell SpellID
	}
	Scribe struct {
		Who   CharID
		Spell SpellID
	}
	Plan struct {
		Who   CharID
		Spell SpellID
		Count int32
	}
	Rest       struct{ Minutes int32 }
	Fix        struct{ Who CharID }
	Bandage    struct{ Who CharID }
	LayOnHands struct{ Who CharID }

	// Trade.
	Buy struct {
		Who  CharID
		Item CarriedID
	}
	Sell struct {
		Who  CharID
		Item ItemSlot
	}
	Identify struct {
		Who  CharID
		Item ItemSlot
	}
	Appraise struct {
		Who  CharID
		Item ItemSlot
		Keep bool
	}
	Repair struct {
		Who  CharID
		Item ItemSlot
	} // BR
	Reload struct {
		Who  CharID
		Item ItemSlot
		Kind ReloadKind
	} // BR
	Train struct{ Who CharID }
	Learn struct {
		Who   CharID
		Spell SpellID
	}
	// SetPolicy is Who's standing policy for one of its reactions (por
	// SetPolicy, the sheet's Reactions).
	SetPolicy struct {
		Who      CharID
		Reaction int
		Policy   Policy
	}
	Heal struct {
		Service ServiceID
		Who     CharID
	}

	// Loot.
	TakeItem struct {
		Who  CharID
		Item CarriedID
	}
	TakeMoney struct {
		Who    CharID
		Amount Copper
	}

	// Doors (por TryDoor; the engine moves, the rules roll).
	BashDoor struct{ Who CharID }
	PickLock struct{ Who CharID }

	// What a script does to characters, as commands (por opDamage, sotsb cmdGainExp).
	// ScriptDamage is a script's DAMAGE (ruling 3, cotab's rule for all):
	// Dice d Sides + Bonus; Random: to Count members picked at random, each
	// by a trap's hit roll at HitBonus and the damage rolled again after
	// each; else to every member (All) or to one, picked at random or
	// the Selected (BySelected), each with a saving throw unless Save is
	// "". A made save spares unless SaveHurts. The rules pick and roll;
	// the Outcome has a Hurt per member hit.
	ScriptDamage struct {
		Selected           CharID
		Dice, Sides, Bonus int32
		Random             bool
		Count, HitBonus    int32
		All, BySelected    bool
		Save               Save
		SaveBonus          int32
		SaveHurts          bool
	}
	// DuelClone: por DuelSetup's opponent, a copy of Of named Name joining
	// the monsters' side with copies of Of's items; Icon is its combat icon slot.
	DuelClone struct {
		Of   CharID
		Name string
		Icon int32
	}
	ScriptGainExp struct {
		Whom []CharID
		Exp  int32
	}
)

// What a script used to cause by writing bytes.
type (
	// RemoveMember: POR off 0 and 0x100 written 0, then the drop bit of
	// SelectMember (eclops.go:178-194); COTAB nameCleared, activeCleared.
	RemoveMember struct{ Who CharID }
	// SpringTrap: a script's trap, as data (por, cotab, sotsb Damage).
	SpringTrap struct {
		Whom []CharID
		Trap TrapDefinition
	}
	// SetStatusAction: a script's status with its consequences, 0x87 stones.
	SetStatusAction struct {
		Who    CharID
		Status Status
	}
	// NewGroup starts a group of monsters: the next Spawn rolls a fresh one,
	// later Spawns of the same MonsterID copy it (por addMonsters clones a
	// template without re-rolling).
	NewGroup struct{}
)

func (RemoveMember) action()    {}
func (SpringTrap) action()      {}
func (SetStatusAction) action() {}
func (NewGroup) action()        {}

// Aim is one aim the player took for a spell: Found false is a cancelled
// aim; Target (NoChar for bare ground) stands At.
type Aim struct {
	Found  bool
	Target CharID
	At     BoardPos
}

// ItemSlot is an item on a character by its id on that character. CarriedID is
// the id of one item the ruleset holds or Situation lists. ManeuverID is a
// ruleset's own maneuver; ServiceID one of a Temple's services.
type (
	ItemSlot   int32
	CarriedID  int32
	ManeuverID int32
	ServiceID  int32
)

func (Move) action()          {}
func (AutoTurn) action()      {}
func (FightPlaced) action()   {}
func (PlaceFight) action()    {}
func (Attack) action()        {}
func (Shoot) action()         {}
func (Cast) action()          {}
func (UseItem) action()       {}
func (Maneuver) action()      {}
func (TurnUndead) action()    {}
func (Guard) action()         {}
func (Delay) action()         {}
func (Quick) action()         {}
func (Flee) action()          {}
func (EndTurn) action()       {}
func (Aid) action()           {}
func (Sprint) action()        {}
func (Dodge) action()         {}
func (Intimidate) action()    {}
func (Jetpack) action()       {}
func (Confirmed) action()     {}
func (Ready) action()         {}
func (Unready) action()       {}
func (DropItem) action()      {}
func (GiveItem) action()      {}
func (Halve) action()         {}
func (JoinStack) action()     {}
func (NewCharacter) action()  {}
func (Modify) action()        {}
func (ChooseRace) action()    {}
func (ChooseClass) action()   {}
func (ChooseSex) action()     {}
func (ChooseAlign) action()   {}
func (SetScore) action()      {}
func (PointBuy) action()      {}
func (Reroll) action()        {}
func (PickSkill) action()     {}
func (Rename) action()        {}
func (KeepDraft) action()     {}
func (DropDraft) action()     {}
func (ChangeClass) action()   {}
func (Memorize) action()      {}
func (Forget) action()        {}
func (Scribe) action()        {}
func (Plan) action()          {}
func (Rest) action()          {}
func (Fix) action()           {}
func (Bandage) action()       {}
func (LayOnHands) action()    {}
func (Buy) action()           {}
func (Sell) action()          {}
func (Identify) action()      {}
func (Appraise) action()      {}
func (Repair) action()        {}
func (Reload) action()        {}
func (Train) action()         {}
func (Learn) action()         {}
func (SetPolicy) action()     {}
func (Heal) action()          {}
func (TakeItem) action()      {}
func (TakeMoney) action()     {}
func (BashDoor) action()      {}
func (PickLock) action()      {}
func (ScriptDamage) action()  {}
func (DuelClone) action()     {}
func (ScriptGainExp) action() {}

// ReloadKind is BR's ammunition choice (shop.pas ShopAmmoMenu 'C'/'F'): a
// new clip, or the weapon filled up.
type ReloadKind int8

const (
	ReloadClip ReloadKind = iota + 1
	ReloadFill
)

// AutoKey is a key the player pressed for a computer turn (por aiKeys).
type AutoKey uint8

const (
	AutoKeyNone       AutoKey = iota
	AutoKeyQuickMagic         // the computer casts for the party, or stops
	AutoKeyTakeOver           // every player character back under manual control
)

// Policy is when a reaction fires: a member's standing choice per reaction.
type Policy uint8

const (
	PolicyAlways Policy = iota // the zero value
	PolicyNever
	PolicyOnHit // only when it turns a hit (the ruleset says which way)
)
