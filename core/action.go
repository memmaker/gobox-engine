package core

import "image"

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
	Squares []image.Point
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
	At   image.Point
	Path []image.Point
}

type (
	// In a fight (the actor is the Fight's).
	Move   struct{ To image.Point }
	Attack struct {
		Target CharID
		At     *image.Point
	}
	Shoot struct{ Ats []image.Point }
	Cast  struct {
		Caster CharID
		Spell  SpellID
		Target CharID
		Ats    []image.Point
	}
	UseItem struct {
		User   CharID
		Item   ItemSlot
		Target CharID
		Ats    []image.Point
		Spell  SpellID // a scroll's spell
	}
	Maneuver struct {
		ID  ManeuverID
		Ats []image.Point
	}
	TurnUndead struct{}
	Guard      struct{}
	Delay      struct{}
	Quick      struct{ All bool }
	Flee       struct{}
	EndTurn    struct{}
	Aid        struct{ Target CharID }  // BR
	Sprint     struct{}                 // BR
	Dodge      struct{}                 // BR
	Intimidate struct{}                 // BR
	Jetpack    struct{ To image.Point } // BR

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
		Item ItemID
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
	} // BR
	Train struct{ Who CharID }
	Learn struct {
		Who   CharID
		Spell SpellID
	}
	Heal struct {
		Service ServiceID
		Who     CharID
	}

	// Loot.
	TakeItem struct {
		Who  CharID
		Item ItemID
	}
	TakeMoney struct{}

	// Doors (por TryDoor; the engine moves, the rules roll).
	BashDoor  struct{ Who CharID }
	PickLock  struct{ Who CharID }
	KnockDoor struct{ Who CharID }

	// What a script does to characters, as commands (por opDamage, sotsb cmdGainExp).
	ScriptDamage struct {
		Whom               []CharID
		Dice, Sides, Bonus int32
		Save               Save // "": no saving throw
		SaveBonus          int32
	}
	ScriptGainExp struct {
		Whom []CharID
		Exp  int32
	}
)

// ItemSlot is an item on a character by its id on that character. ItemID is
// the id of one item the ruleset holds or Situation lists. ManeuverID is a
// ruleset's own maneuver; ServiceID one of a Temple's services.
type (
	ItemSlot   int32
	ItemID     int32
	ManeuverID int32
	ServiceID  int32
)

func (Move) action()          {}
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
func (Heal) action()          {}
func (TakeItem) action()      {}
func (TakeMoney) action()     {}
func (BashDoor) action()      {}
func (PickLock) action()      {}
func (KnockDoor) action()     {}
func (ScriptDamage) action()  {}
func (ScriptGainExp) action() {}
