package rules

// A ship fight follows the pattern: this data here, the adventure says who and
// when (a script's SpaceFight names a ShipEncounter), the rules decide, the
// engine takes input and presents. BR spacefight.go, game.SpaceFight.

// ShipSystem is a part of a ship with a gauge: core/xxvc names them.
type ShipSystem string

// ShipEncounter is what a script's SpaceFight says (spaceFight v1b, v4b, v3b).
type ShipEncounter struct {
	Foe   ShipID
	Range int32 // B4A1F
	// The script's pilot skill and to-hit number for the foe (BR ShipCombat's
	// ECL arguments 3 and 4).
	PilotSkill, ToHit int32
}

// Gauge is a system's state.
type Gauge struct{ Now, Max int32 }

// Ship is one ship as the fight holds it (eclbr.Ship, game.ShipSheet).
type Ship struct {
	ID      ShipID
	Name    string
	Systems map[ShipSystem]Gauge
	Weapons []ItemID // the mounts
}

// ShipFight is a fight's snapshot, and the Situation for Allowed and Resolve.
// Crew takes turns aboard Own (teamStart, memberNext).
type ShipFight struct {
	Own, Foe Ship
	Range    int32
	Round    int32
	Crew     Party
	Actor    CharID // NoChar: the foe's part of the round
}

// The crew's actions (fire, boost, juryRig, ram, dodge, closeIn, retreat, sReload).
type (
	ShipFire struct {
		Weapon int32 // a mount of Own.Weapons
		Target ShipSystem
	}
	ShipBoost   struct{ System ShipSystem }
	ShipJuryRig struct {
		System ShipSystem
		Mount  int32 // a mount of Own.Weapons, for the weapons system
	}
	ShipRam     struct{}
	ShipDodge   struct{ Yes bool } // the answer to the rules' dodge question (Outcome.Confirm)
	ShipCommand struct{}           // take over as pilot (BR)
	ShipBoard   struct{}           // board a stopped foe (BR)
	ShipSensors struct{}           // sensor operation (BR)
	ShipClose   struct{}
	ShipRetreat struct{}
	ShipReload  struct{ Weapon int32 }
)

func (ShipFire) action()    {}
func (ShipBoost) action()   {}
func (ShipJuryRig) action() {}
func (ShipRam) action()     {}
func (ShipDodge) action()   {}
func (ShipClose) action()   {}
func (ShipRetreat) action() {}
func (ShipReload) action()  {}
func (ShipCommand) action() {}
func (ShipBoard) action()   {}
func (ShipSensors) action() {}

// ShipRules are a ship fight's sequence. Rules.Ships gives them, nil when the
// ruleset's Catalog lacks FeatureSpaceCombat: such an adventure never loads.
//
//	BeginShips; for: ShipRound; each Crew: Allowed, Resolve; ShipFoe; ShipVerdict; EndShips
type ShipRules interface {
	BeginShips(ShipEncounter, *RNG) (ShipFight, error) // spaceFight, loadFoe
	ShipRound(ShipFight, *RNG) (ShipFight, Outcome)    // roundEnd, sReload
	ShipFoe(ShipFight, *RNG) (ShipFight, Outcome)      // foeTurn, foeShoots, foeFire
	ShipVerdict(ShipFight) Verdict                     // spaceEnd
	EndShips(ShipFight, *RNG) FightResult              // spaceEnd
}
