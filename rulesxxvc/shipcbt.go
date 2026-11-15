package rulesxxvc

import (
	"errors"
	"slices"
	"strconv"
	"unicode"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// The ship fight: shipcbt.pas ShipCombat (BR1; BR2's copy differs only in
// DamageCrew's explosion, xxvc.Tables.Explosion, and in typing that is equal
// for the ships' records: TargetValue's armour class as a Byte, Word casts
// on tonnage and crew; every SHIPS armour class, tonnage and crew is >= 0).
//
// The engine's loop (core.ShipRules):
//
//	BeginShips; for ShipVerdict = FightOn: ShipRound; each Crew member in
//	order: while Allowed offers, Resolve; ShipFoe; then EndShips.
//
// ShipRound starts a round (shotsFired cleared) and, when the party won the
// pilot contest, runs the enemy's turn first; ShipFoe runs the enemy's turn
// otherwise, then the round's end (range, coma). A member's turn ends with
// an action the original counts as taken (fire, reload, jury-rig, aid,
// boost, sensors, ram, board, quit); Close, Withdraw and Command leave the
// member at the menu, so Allowed offers again. An Outcome.Confirm from
// ShipFoe or ShipRound asks "Try to dodge enemy ship?": the engine answers
// Resolve(ShipDodge{Yes}), which plays the rest of the enemy's turn.

// The systems by TShip.systemStatus' index (c_sys.pas), and the targets.
const (
	sysHull = iota
	sysSensors
	sysControl
	sysLifeSupport
	sysFuel
	sysEngines
	sysWeapons // C_Sys.WEAPONS: a mount, no gauge
	anySystem  // ShipCbt ANY_SYSTEM: systemHitChances picks
)

const (
	partySide = 0 // shipFight rows: systemJuryRigs[0], the party's ship
	foeSide   = 1

	scannerRange = 8 // SCANNER_RANGE
	pumpedLaser  = 2 // PUMPED_LASER: the kinds above fire ammunition
	mounts       = 8

	sensorOperation SkillNo = 10 // SKILL_SENSOR_OPERATION
	pilotRocket     SkillNo = 51 // C_Skill.PILOT_ROCKET
	repairWeapon    SkillNo = 14 // C_Skill.REPAIR_WEAPON
)

// The ship combat's end (zone0 shipCombatOutcome and postCombatMode).
type shipEnd int8

const (
	shipOn       shipEnd = iota
	foeDestroyed         // shipCombatOutcome 0
	boarded              // C_PostCb.BOARDED, outcome 0
	hullLost             // shipCombatOutcome 1
	receded              // outcome 1, no experience
)

// foeStage is where the enemy's turn stands: the dodge question pauses it.
type foeStage int8

const (
	foeIdle foeStage = iota
	foeAsking
)

// points counts a mount's shots or rigged points (a Byte).
type points uint8

// shipFight is ShipCombat's locals and the GDVars it uses.
type shipFight struct {
	ships        [2]xxvc.ShipRecord // the party's (zone0 partyShip) and the enemy's
	boosts       [2]int8            // playerBoosts, enemyBoosts
	systemRigs   [2][6]int16        // systemJuryRigs
	mountRigs    [2][mounts]points  // weaponJuryRigs
	shots        [2][mounts]points  // shotsFired
	rangeNow     int8               // GDVars.shipCombatRange: the round's
	combatRange  int8               // where the moves put it
	enemyFirst   bool
	lifeFailed   bool // hasLifeSupportFailed
	enemyCanMove bool // canEnemyManeuver
	retreating   bool // GDVars.isEnemyRetreating
	pilot        rules.CharID
	crew         []rules.CharID // partyHead's order
	dying        map[rules.CharID]int
	exp          uint16 // shipCombatExperience
	end          shipEnd
	foeID        rules.ShipID

	// TEncounterState.
	reporting     bool
	firingSlot    int
	damageSlot    int
	damage        uint16
	reportPending bool
	enemyRigged   bool
	commandTaken  bool

	// The crew's turn (PlayerTurnBody): the member at the menu.
	crewOn      bool
	turn        int
	canManeuver bool

	// The enemy's turn paused at the dodge question.
	stage        foeStage
	strategy     int // EnemyTurnBody's, kept over the question
	boostPending bool
	tailPending  bool // ShipFoe's round end still to run after the answer
	crewAfter    bool // ShipRound's: the crew acts after the answer

	rng *rules.RNG
	out []rules.Effect
}

// savedFoe is zone0's record of the last enemy ship for the scripts
// (savedSystemStatus, savedShipType, savedCrew, savedSystemValue,
// savedWeaponCount); zero after one lost off the scanners.
type savedFoe struct {
	Status      [6]int16
	Type        string
	Crew        int16
	SystemValue uint8
	WeaponCount uint8
}

// errNoShips is what the ship calls answer outside a ship fight.
var errNoShips = errors.New("rulesxxvc: no ship fight")

func (r *Rules) Ships() rules.ShipRules { return shipRules{r} }

// shipRules is Rules' core.ShipRules.
type shipRules struct{ r *Rules }

func (f *shipFight) say(text string) { f.out = append(f.out, rules.Narrated{Text: text}) }

func (f *shipFight) flush() rules.Outcome {
	o := rules.Outcome{Effects: f.out}
	f.out = nil
	return o
}

// partyShip is zone0's partyShip: GameInit loads SHIPS record 50 into it.
func (r *Rules) partyShip() xxvc.ShipRecord {
	if r.ship == nil {
		s := r.tables.Ships[xxvc.PartyShip]
		r.ship = &s
	}
	return *r.ship
}

// shipExperience is ShipCombat's experience by enemy ship id (ECL argument 1 less 1).
var shipExperience = map[rules.ShipID]uint16{0: 300, 1: 3000, 2: 12600, 3: 2500, 4: 4600}

// firstShipRecord is SaveLoad FIRST_SHIP_RECORD: ship id n is SHIPS record 40 + n.
const firstShipRecord = 40

// BeginShips is ShipCombat's set-up: Foe is the SHIPS record (40 + the ECL's
// argument 1 less 1), PilotSkill and ToHit arguments 3 and 4; the first
// enabled member marked as pilot flies, later marks are cleared; the pilot
// contest decides who moves first.
func (s shipRules) BeginShips(e rules.ShipEncounter, rng *rules.RNG) (rules.ShipFight, error) {
	r := s.r
	foe, ok := r.tables.Ships[e.Foe]
	if !ok {
		return rules.ShipFight{}, errors.New("rulesxxvc: no ship " + strconv.Itoa(int(e.Foe)))
	}
	f := &shipFight{ships: [2]xxvc.ShipRecord{r.partyShip(), foe}, reporting: true, firingSlot: -1, damageSlot: -1,
		dying: map[rules.CharID]int{}, foeID: e.Foe, rng: rng}
	f.exp = shipExperience[e.Foe-firstShipRecord]
	for _, id := range r.order {
		if c := r.chars[id]; c != nil && c.team() {
			f.crew = append(f.crew, id)
		}
	}
	for _, id := range f.crew {
		c := r.chars[id]
		if c.Pilot {
			if f.pilot != rules.NoChar {
				c.Pilot = false
			} else if c.Enabled {
				f.pilot = id
			}
		}
	}
	f.ships[foeSide].AlternateSkill = uint8(e.PilotSkill)
	f.ships[foeSide].SystemValue = uint8(60 - e.ToHit)
	f.rangeNow, f.combatRange = scannerRange, scannerRange
	f.enemyFirst = f.pilotContest()
	r.space = f
	return r.shipSnapshot(), nil
}

// shipSnapshot is the fight as core.ShipFight.
func (r *Rules) shipSnapshot() rules.ShipFight {
	f := r.space
	ship := func(i int) rules.Ship {
		sh := rules.Ship{Name: f.ships[i].Name, Systems: map[rules.ShipSystem]rules.Gauge{}}
		for n, sys := range xxvc.ShipSystems {
			sh.Systems[sys] = rules.Gauge{Now: int32(f.ships[i].Status[n]), Max: int32(f.ships[i].Maximum[n])}
		}
		return sh
	}
	sf := rules.ShipFight{Own: ship(partySide), Foe: ship(foeSide), Range: int32(f.rangeNow), Crew: f.crew, Actor: r.crewActor()}
	sf.Own.ID, sf.Foe.ID = xxvc.PartyShip, f.foeID
	return sf
}

// ShipRound is the round's start: shotsFired cleared; the enemy first when
// the party won the pilot contest.
func (s shipRules) ShipRound(sf rules.ShipFight, rng *rules.RNG) (rules.ShipFight, rules.Outcome) {
	f := s.r.space
	if f == nil {
		return sf, rules.Outcome{}
	}
	f.rng = rng
	if len(sf.Crew) > 0 {
		f.crew = slices.Clone(sf.Crew)
	}
	f.shots = [2][mounts]points{}
	f.crewOn = false
	if f.enemyFirst {
		f.crewAfter = true
		if s.r.enemyTurn() {
			return s.r.shipSnapshot(), f.flush()
		}
		f.crewAfter = false
	}
	s.r.startCrew()
	return s.r.shipSnapshot(), f.flush()
}

// ShipFoe is the enemy's turn (when it did not go first) and the round's end.
func (s shipRules) ShipFoe(sf rules.ShipFight, rng *rules.RNG) (rules.ShipFight, rules.Outcome) {
	f := s.r.space
	if f == nil {
		return sf, rules.Outcome{}
	}
	f.rng = rng
	f.crewOn = false
	if !f.enemyFirst {
		f.tailPending = true
		if s.r.enemyTurn() {
			return s.r.shipSnapshot(), f.flush()
		}
		f.tailPending = false
	}
	s.r.roundEnd()
	return s.r.shipSnapshot(), f.flush()
}

func (s shipRules) ShipVerdict(rules.ShipFight) rules.Verdict {
	f := s.r.space
	if f == nil {
		return rules.FightOn
	}
	switch f.end {
	case foeDestroyed, boarded:
		return rules.PartyWon
	case hullLost:
		return rules.PartyLost
	case receded:
		return rules.PartyFled
	}
	return rules.FightOn
}

// EndShips is ShipCombat's end: the party's jury-rigs come off, zone 0 keeps
// the enemy ship for the scripts, and unless the party boarded, RepairShip
// shares the experience and repairs.
// ponytail: PostCombat's DetermineCombatResults and TreatInjuries (before
// RepairShip, drawing from the same RandSeed) are not ported for ship fights.
func (s shipRules) EndShips(_ rules.ShipFight, rng *rules.RNG) rules.FightResult {
	r := s.r
	f := r.space
	if f == nil {
		return rules.FightResult{}
	}
	f.rng = rng
	f.removeJuryRigs()
	if f.exp == 0 {
		r.foeSaved = savedFoe{}
	} else {
		foe := f.ships[foeSide]
		r.foeSaved = savedFoe{Status: foe.Status, Type: foe.Type, Crew: foe.Crew, SystemValue: foe.SystemValue}
		for _, m := range foe.Mounts {
			r.foeSaved.WeaponCount += m.Status
		}
	}
	var res rules.FightResult
	if f.end != boarded && !r.partyDefeated(f) {
		res.Exp = int32(r.repairShip(f))
	}
	ship := f.ships[partySide]
	r.ship = &ship
	res.Effects = f.out
	r.space = nil
	return res
}

// partyDefeated is GDVars.isPartyDefeated: no member left standing.
// ponytail: DetermineCombatResults' own test is not ported; this is "no
// member okay".
func (r *Rules) partyDefeated(f *shipFight) bool {
	return !slices.ContainsFunc(f.crew, func(id rules.CharID) bool { return r.chars[id].Status == StOkay })
}

// pilotContest is PilotContest: 1d10 - 1d10 + the reactions' difference > 0.
func (f *shipFight) pilotContest() bool {
	score := int8(rules.RNGDice(f.rng, 1, 10) - rules.RNGDice(f.rng, 1, 10))
	score += f.ships[partySide].Reaction
	score -= f.ships[foeSide].Reaction
	return score > 0
}

// removeJuryRigs is RemoveJuryRigs: the party's rigged points come off, a
// system only when it has more left; a mount at worst destroyed.
func (f *shipFight) removeJuryRigs() {
	s := &f.ships[partySide]
	for i := range s.Status {
		if f.systemRigs[0][i] < s.Status[i] {
			s.Status[i] -= f.systemRigs[0][i]
		}
	}
	for i := range s.Mounts {
		if uint8(f.mountRigs[0][i]) > s.Mounts[i].Status {
			s.Mounts[i].Status = 0
		} else {
			s.Mounts[i].Status -= uint8(f.mountRigs[0][i])
		}
	}
}

// raceNames is gdconsts.pas raceNames (RaceMonster: "").
var raceNames = map[rules.Race]string{"": "Monster", xxvc.Terran: "Terran", xxvc.Martian: "Martian", xxvc.Venusian: "Venusian",
	xxvc.Mercurian: "Mercurian", xxvc.Tinker: "Tinker", xxvc.DesertRunner: "Desert Runner", xxvc.Lunarian: "lunarian", xxvc.Lowlander: "lowlander"}

// systemNames is gdconsts.pas shipSystemNames.
var systemNames = [6]string{"Hull", "Sensors", "Control systems", "Life support", "Fuel", "Engines"}

// formatShipName is FormatShipName: the actor's name, the party ship's, or
// for the enemy a d4 pick of "Enemy ship", race and type, or its name.
func (r *Rules) formatShipName(side int, actor *Character) string {
	f := r.space
	switch {
	case actor != nil:
		return actor.Name
	case side == partySide:
		return f.ships[side].Name
	}
	switch rules.RNGDice(f.rng, 1, 4) {
	case 1:
		return "Enemy ship"
	case 2:
		return raceNames[f.ships[side].Race] + " " + f.ships[side].Type
	}
	return f.ships[side].Name
}

// shipSkillCheck is ShipCombat's SkillCheck: source's skill, or the ship's
// own value (systemSkill for Repair Engine and Jury Rig, alternateSkill for
// Pilot Rocket), plus bonus, scaled (0 x4 .. 4 /4), a percentile roll.
func (r *Rules) shipSkillCheck(n SkillNo, scale rules.Hardness, side int, source *Character, bonus int) bool {
	f := r.space
	var value uint8
	switch {
	case source != nil:
		value = uint8(skillValue(source, n))
	case n == RepairEngine || n == JuryRig:
		value = f.ships[side].SystemSkill
	case n == pilotRocket:
		value = f.ships[side].AlternateSkill
	}
	return ShipSkillCheck(f.rng, int(value), bonus, scale)
}

// canMove is CharHlp.CanShipMove: control, fuel and engines all work.
func (f *shipFight) canMove(side int) bool {
	s := f.ships[side].Status
	return s[sysControl] > 0 && s[sysFuel] > 0 && s[sysEngines] > 0
}

// speed is CharHlp.ShipSpeed.
func (f *shipFight) speed(side int) int8 {
	if !f.canMove(side) {
		return 0
	}
	s := f.ships[side]
	speed := s.BaseSpeed + f.boosts[side]
	if s.Maximum[sysHull]>>1 > s.Status[sysHull] {
		speed = int8(uint16(int16(speed)) >> 1)
	}
	if s.Maximum[sysEngines]>>2 > s.Status[sysEngines] {
		speed -= 3
	} else if s.Maximum[sysEngines]>>1 > s.Status[sysEngines] {
		speed--
	}
	return max(speed, 1)
}

// weapon is shipWeapons[kind].
func (r *Rules) weapon(m xxvc.ShipMount) xxvc.ShipWeapon { return r.tables.ShipWeapons[m.Kind] }

// mountReady is WeaponStatus = WEAPON_READY for the pilot (crew false) or the crew.
func (r *Rules) mountReady(side int, crew bool, slot int) bool {
	f := r.space
	m := f.ships[side].Mounts[slot]
	return m.Kind != 0 && m.Crew == crew && m.Status != 0 && (m.Kind <= pumpedLaser || m.Ammo != 0) &&
		uint8(f.shots[side][slot]) < m.Status && int(f.rangeNow) <= int(r.weapon(m).MaxRange)
}

// bestWeapon is BestWeapon: the ready mount of the most damage, the first
// of equals, or -1.
func (r *Rules) bestWeapon(side int, crew bool) int {
	best, priority := -1, uint8(0)
	for i, m := range r.space.ships[side].Mounts {
		if r.mountReady(side, crew, i) && r.weapon(m).Damage > priority {
			best, priority = i, r.weapon(m).Damage
		}
	}
	return best
}

// animateHit is AnimateHit's draws: a d8 and two d3 for where it flashes.
func (f *shipFight) animateHit() {
	rules.RNGDice(f.rng, 1, 8)
	rules.RNGDice(f.rng, 1, 3)
	rules.RNGDice(f.rng, 1, 3)
}

// damageCrew is DamageCrew: each member who is okay is hurt chance percent
// of the time (a save 2 leaps clear, unless chance is 100) or, with
// explosion, always.
func (r *Rules) damageCrew(chance uint8, explosion bool) {
	f := r.space
	for _, id := range f.crew {
		c := r.chars[id]
		if c.Status != StOkay {
			continue
		}
		if !(PercentileRoll(f.rng, int(chance)) || explosion) {
			continue
		}
		if rollSave(c, 2, 0, f.rng) && !explosion && chance < 100 {
			f.say(c.Name + " leaps clear of an explosion.")
			continue
		}
		var damage uint8
		switch {
		case explosion && r.tables.Explosion == xxvc.ExplosionFixed:
			damage = uint8(c.MaxHP + 30)
		case explosion:
			damage = uint8(rules.RNGDice(f.rng, 5, 20) + c.MaxHP)
		default:
			damage = uint8(rules.RNGDice(f.rng, 1, 9) + 1)
		}
		InflictDamage(c, int(damage))
		f.say(c.Name + " takes " + strconv.Itoa(int(damage)) + " points of damage!")
		if !explosion {
			switch c.Status {
			case StDead:
				f.say(c.Name + " is dead!!")
			case StDying:
				f.say(c.Name + " is dying!")
			case StUnconscious:
				f.say(c.Name + " is unconscious!")
			default:
				if c.HP < 10 {
					f.say(c.Name + " is seriously injured.")
				}
			}
		}
		if id == f.pilot && c.Status != StOkay {
			c.Pilot = false
			f.pilot = rules.NoChar
			f.commandTaken = false
		}
	}
}

// reportEnemyDamage is ReportEnemyDamage: the pending report of the enemy's
// mount damageSlot, then DamageCrew at a fifth of the damage.
func (r *Rules) reportEnemyDamage() {
	f := r.space
	if !f.reportPending {
		return
	}
	name := "" // ponytail: damageSlot -1 reads shipWeapons' memory before the table in the original
	if f.damageSlot >= 0 {
		name = r.weapon(f.ships[foeSide].Mounts[f.damageSlot]).Name
	}
	msg := " fires " + name + "s and "
	if f.damage > 0 {
		msg += "hits for " + strconv.Itoa(int(f.damage)) + " points of damage!"
	} else {
		msg += "misses."
	}
	f.say(r.formatShipName(foeSide, nil) + msg)
	r.damageCrew(uint8(f.damage/5), false)
	f.damage = 0
	f.reportPending = false
}

// damageShip is DamageShip: value off system sys (a destroyed one passes to
// the hull), or a point off a mount picked by status for sysWeapons; losses
// reported; then a crew member lost on PercentileRoll(value / 2).
func (r *Rules) damageShip(side, sys int, value uint16) {
	f := r.space
	s := &f.ships[side]
	exploding := false
	if sys == sysWeapons {
		var total int8
		for _, m := range s.Mounts {
			total += int8(m.Status)
		}
		if total <= 0 {
			return
		}
		total = int8(rules.RNGDice(f.rng, 1, int(uint8(total))))
		for i := range s.Mounts {
			if total <= 0 {
				continue
			}
			if total -= int8(s.Mounts[i].Status); total < 1 {
				if side == partySide {
					r.reportEnemyDamage()
				}
				f.say(r.weapon(s.Mounts[i]).Name + " destroyed on " + r.formatShipName(side, nil) + "!")
				s.Mounts[i].Status--
			}
		}
	} else {
		if s.Status[sys] == 0 {
			sys = sysHull
		}
		if int32(value) > int32(s.Status[sys]) {
			value = uint16(s.Status[sys])
		}
		if value == 0 {
			return
		}
		s.Status[sys] -= int16(value)
		if s.Status[sys] == 0 {
			if side == partySide {
				r.reportEnemyDamage()
			}
			f.say(systemNames[sys] + " destroyed!")
			switch {
			case sys == sysHull:
				if side == foeSide {
					f.end = foeDestroyed
					for range 6 {
						f.animateHit()
					}
				} else {
					f.end = hullLost
					exploding = true
				}
			case sys == sysLifeSupport && side == partySide && !f.lifeFailed:
				r.damageCrew(100, false)
				value = 0
				f.lifeFailed = true
			}
		}
	}
	if exploding {
		r.damageCrew(100, true)
	}
	if CrewLost(f.rng, int(value), int(s.Crew)) {
		s.Crew--
	}
}

// targetValue is TargetValue for side's mount slot against the other ship.
func (r *Rules) targetValue(side, slot int, source *Character) int {
	f := r.space
	a, d := f.ships[side], f.ships[1-side]
	sh := Shooter{SystemValue: int(a.SystemValue), PartyShip: side == partySide, TargetBonus: int(r.weapon(a.Mounts[slot]).TargetBonus),
		ControlDown: a.Status[sysControl] == 0, TargetControlDown: d.Status[sysControl] == 0,
		TargetManeuver: int(d.Maneuver), TargetArmorClass: uint8(d.ArmorClass), Range: int(f.rangeNow)}
	if source != nil {
		sh.Crew, sh.BaseCombat, sh.Dex = true, source.BaseCombat, source.Scores[rules.Dexterity]
		sh.RocketJock, sh.Level = source.Career == xxvc.RocketJock, source.Level
	}
	return TargetValue(sh)
}

// dispatchWeaponHit is DispatchWeaponHit: kind's damage to the target
// system, or one systemHitChances picks.
func (r *Rules) dispatchWeaponHit(side int, m xxvc.ShipMount, target int) {
	f := r.space
	damage := uint16(r.weapon(m).Damage)
	if side == foeSide {
		f.animateHit()
	}
	if target == anySystem {
		target = HitSystem(f.rng)
	}
	r.damageShip(side, target, damage)
}

// fireWeapon is FireWeapon: a shot of side's mount slot; an aimed shot
// first needs a 50% roll, then the d20 against TargetValue.
func (r *Rules) fireWeapon(side, slot int, source *Character, target int) {
	f := r.space
	m := f.ships[side].Mounts[slot]
	f.reportPending = true
	hit := (target == anySystem || PercentileRoll(f.rng, 50)) && rules.RNGDice(f.rng, 1, 20) >= r.targetValue(side, slot, source)
	msg := "misses."
	if hit {
		damage := uint16(r.weapon(m).Damage)
		msg = "hits for " + strconv.Itoa(int(damage)) + " points of damage!"
		f.damage += damage
	}
	if f.reporting {
		f.say(r.formatShipName(side, source) + " attacks with a " + r.weapon(m).Name + " and " + msg)
	}
	if hit {
		r.dispatchWeaponHit(1-side, m, target)
	}
	if m.Kind > pumpedLaser {
		f.ships[side].Mounts[slot].Ammo--
	}
	f.shots[side][slot]++
}

// attemptSpeedBoost is AttemptSpeedBoost: a halved Repair Engine check; a
// failure does 1d100 to the engines and gives no boost.
func (r *Rules) attemptSpeedBoost(side int, source *Character) {
	f := r.space
	msg := r.formatShipName(side, source) + " attempts a speed boost and "
	var damage uint16
	if r.shipSkillCheck(RepairEngine, rules.Hard, side, source, 0) {
		msg += "succeeds!"
		f.boosts[side]++
	} else {
		msg += "fails!"
		damage = uint16(rules.RNGDice(f.rng, 1, 100))
	}
	f.say(msg)
	if damage > 0 {
		r.damageShip(side, sysEngines, damage)
	}
}

// canJuryRig is AttemptJuryRig's CanJuryRig: a mount below its maxStatus
// with fewer rigged points than that.
func (f *shipFight) canJuryRig(side int) bool {
	for i, m := range f.ships[side].Mounts {
		if m.Status < m.MaxStatus && uint8(f.mountRigs[side][i]) < m.MaxStatus {
			return true
		}
	}
	return false
}

// rigSystems are the party's jury-rig menu's systems: below maximum, not
// rigged, not fuel.
func (f *shipFight) rigSystems(side int) []int {
	var out []int
	s := f.ships[side]
	for i := range s.Status {
		if s.Status[i] < s.Maximum[i] && f.systemRigs[side][i] == 0 && i != sysFuel {
			out = append(out, i)
		}
	}
	return out
}

// hotkey is MenuHorizontal's key of a menu word: its first capital or digit.
func hotkey(word string) rune {
	for _, c := range word {
		if unicode.IsUpper(c) || unicode.IsDigit(c) {
			return c
		}
	}
	return 0
}

// pickMount is the menus' weapon pick (AttemptJuryRig, ReloadWeapon): the
// last mount whose weaponShortNames' first letter is the key, whatever its
// state. A short name whose first letter is not its key ('acl-Cann',
// 'hv-Accel') is never picked: the menu stays up. -1 none.
func (r *Rules) pickMount(side int, key rune) int {
	slot := -1
	for i, m := range r.space.ships[side].Mounts {
		if s := r.weapon(m).Short; m.Kind != 0 && s != "" && rune(s[0]) == key {
			slot = i
		}
	}
	return slot
}

// attemptJuryRig is AttemptJuryRig. The party's pick is sys and slot (the
// menu's choice, checked by Allowed); the enemy picks its best mount, else
// its worst system. It answers false when the enemy finds nothing.
func (r *Rules) attemptJuryRig(side int, source *Character, sys, slot int) bool {
	f := r.space
	s := &f.ships[side]
	if source == nil {
		sys, slot = anySystem, -1
		best := 0
		if f.canJuryRig(side) {
			// The original also tests an unset localLoop against FUEL here; taken as never FUEL.
			for i, m := range s.Mounts {
				if m.Status < m.MaxStatus && int(r.weapon(m).Damage) > best {
					slot, best = i, int(r.weapon(m).Damage)
				}
			}
			if slot >= 0 {
				sys = sysWeapons
			}
		}
		if sys == anySystem {
			for i := range s.Status {
				if s.Status[i] != s.Maximum[i] && f.systemRigs[1][i] == 0 {
					score := 10 - int(10*s.Status[i]/s.Maximum[i])
					if s.Status[i] == 0 {
						score += 5
					}
					score *= 2 // systemJuryRigs[1] = 0, tested again
					if score > best {
						best, sys = score, i
					}
				}
			}
			if sys == anySystem {
				// The original leaves the result unset; either way the repair passes do nothing more.
				return false
			}
		}
	}
	what := ""
	if sys == sysWeapons {
		what = r.weapon(s.Mounts[slot]).Name
	} else {
		what = systemNames[sys]
	}
	if r.shipSkillCheck(JuryRig, rules.Normal, side, source, 0) {
		if side == partySide {
			f.say(r.formatShipName(side, source) + " attempts to jury-rig " + what + " and succeeds!")
		} else {
			f.enemyRigged = true
		}
		if sys == sysWeapons {
			f.mountRigs[side][slot]++
			s.Mounts[slot].Status++
		} else {
			amount := uint8(rules.RNGDice(f.rng, 2, 10))
			if s.Status[sys]+int16(amount) > s.Maximum[sys] {
				amount = uint8(s.Maximum[sys] - s.Status[sys])
			}
			s.Status[sys] += int16(amount)
			f.systemRigs[side][sys] += int16(amount)
		}
	} else if side == partySide {
		f.say(r.formatShipName(side, source) + " attempts to jury-rig " + what + " and fails.")
	}
	return true
}

// ramStart is RamAttempt up to DodgesRam's question: it answers whether
// the question is asked (the party's moving ship rammed).
func (r *Rules) ramStart(rammer int, actor *Character) bool {
	f := r.space
	f.say(r.formatShipName(rammer, actor) + " attempts to ram!")
	rammed := 1 - rammer
	return rammed == partySide && f.speed(rammed) > 0
}

// ramEnd is the rest of RamAttempt: the dodge (tried: a Pilot Rocket check
// with the speeds' difference x10) or both hulls take the other's tonnage.
func (r *Rules) ramEnd(rammer int, tries bool) {
	f := r.space
	rammed := 1 - rammer
	var source *Character
	if rammed == partySide && f.pilot != rules.NoChar {
		source = r.chars[f.pilot]
	}
	if rammed == foeSide {
		tries = f.speed(rammed) > 0 && f.ships[rammed].Status[sysHull] < f.ships[rammer].Status[sysHull]
	}
	if tries && r.shipSkillCheck(pilotRocket, rules.Normal, rammed, source, int(f.speed(rammed)-f.speed(rammer))*10) {
		name := "Your ship"
		if source == nil {
			name = r.formatShipName(rammed, nil)
		}
		f.say(name + " dodges out of the way!")
		return
	}
	f.say("Ramming successful!")
	r.damageShip(rammer, sysHull, uint16(f.ships[rammed].Tonnage))
	r.damageShip(rammed, sysHull, uint16(f.ships[rammer].Tonnage))
}

// closeShip is CloseShip: in by the ship's speed.
func (r *Rules) closeShip(side int) {
	f := r.space
	f.say(r.formatShipName(side, nil) + " is closing.")
	f.combatRange -= f.speed(side)
	if side == foeSide {
		f.retreating = false
	}
}

// retreatShip is RetreatShip: out by amount.
func (r *Rules) retreatShip(side int, amount int8) {
	f := r.space
	f.say(r.formatShipName(side, nil) + " is retreating.")
	f.combatRange += amount
	if side == foeSide {
		f.retreating = true
	}
}

// canReload is CanReload: an empty ammunition mount with reloads left.
func (f *shipFight) canReload(side, slot int) bool {
	m := f.ships[side].Mounts[slot]
	return m.Kind > pumpedLaser && m.Reloads > 0 && m.Ammo == 0
}

func (f *shipFight) hasReloadable(side int) bool {
	for i := range mounts {
		if f.canReload(side, i) {
			return true
		}
	}
	return false
}

// reloadWeapon is ReloadWeapon's reload of slot (the enemy's: its last
// reloadable mount).
func (r *Rules) reloadWeapon(side int, source *Character, slot int) bool {
	f := r.space
	if side == foeSide {
		slot = -1
		for i := range mounts {
			if f.canReload(side, i) {
				slot = i
			}
		}
	}
	if slot < 0 {
		return false
	}
	m := &f.ships[side].Mounts[slot]
	m.Reloads--
	m.Ammo += r.weapon(*m).ReloadAmmo
	f.say(r.formatShipName(side, source) + " reloads a " + r.weapon(*m).Name + ".")
	return true
}

// needsRepair is ShipNeedsRepair: a system or mount to jury-rig; for the
// party's ship only when source can Jury Rig.
func (r *Rules) needsRepair(side int, source *Character) bool {
	f := r.space
	if side == partySide && skillValue(source, JuryRig) <= 0 {
		return false
	}
	if len(f.rigSystems(side)) > 0 {
		return true
	}
	return f.canJuryRig(side)
}

// combatValue is ShipCombatValue: hull, a third of control, fuel and
// engines when it can move, and the mounts' damage times status; 0 unarmed.
func (r *Rules) combatValue(side int) int16 {
	f := r.space
	s := f.ships[side]
	base := s.Status[sysHull]
	if f.canMove(side) {
		base += (s.Status[sysControl] + s.Status[sysFuel] + s.Status[sysEngines]) / 3
	}
	var total int16
	for _, m := range s.Mounts {
		if m.Kind != 0 || m.Crew {
			total += int16(r.weapon(m).Damage) * int16(m.Status)
		}
	}
	if total == 0 {
		return 0
	}
	return base + total
}

// fireWorthwhile is EnemyFireWorthwhile: fire rather than ram, weighing the
// enemy's pilot mounts against the party pilot's Pilot Rocket.
// ponytail: with no party pilot the original reads skillLevels through nil
// (DOS memory 0); taken as 0.
func (r *Rules) fireWorthwhile() bool {
	f := r.space
	chosen := r.bestWeapon(foeSide, false)
	if chosen < 0 {
		return false
	}
	score := int16(f.speed(partySide)-f.speed(foeSide)) * 10
	if f.pilot != rules.NoChar {
		score += int16(skillValue(r.chars[f.pilot], pilotRocket))
	}
	if score >= 90 {
		return true
	}
	var power uint8
	for i, m := range f.ships[foeSide].Mounts {
		if r.mountReady(foeSide, false, i) {
			power += m.Status * r.weapon(m).Damage
		}
	}
	return int32(100-score)*int32(uint16(f.ships[foeSide].Tonnage)/10) <
		int32(int16(21-r.targetValue(foeSide, chosen, nil))*5)*int32(int8(power)/10)
}

// enemyTurn is EnemyTurnBody up to the dodge question: it answers true when
// paused there.
func (r *Rules) enemyTurn() bool {
	f := r.space
	if f.end != shipOn {
		return false
	}
	active, enemy := r.combatValue(partySide), r.combatValue(foeSide)
	strategy := 1
	switch {
	case enemy*2 < active:
		strategy = 0
	case enemy > active:
		strategy = 2
	}
	f.boostPending = false
	if f.canMove(foeSide) {
		foe, party := f.ships[foeSide], f.ships[partySide]
		switch {
		case f.rangeNow <= 1 && !r.fireWorthwhile() && foe.Status[sysHull] > party.Status[sysHull] &&
			int32(foe.Status[sysHull]) > int32(uint16(party.Tonnage)):
			if r.ramStart(foeSide, nil) {
				f.stage, f.strategy = foeAsking, strategy
				return true
			}
			r.ramEnd(foeSide, false)
		case strategy == 0 && f.enemyCanMove:
			r.retreatShip(foeSide, f.speed(foeSide))
		case strategy == 1 && f.rangeNow <= 2:
			r.retreatShip(foeSide, min(f.speed(foeSide), 3))
		case strategy == 2 || f.rangeNow >= 5:
			r.closeShip(foeSide)
		}
		f.boostPending = r.boostTest(strategy)
	}
	r.enemyFire()
	return false
}

// boostTest is EnemyTurnBody's boost decision after the move.
func (r *Rules) boostTest(strategy int) bool {
	f := r.space
	return strategy == 0 && f.speed(foeSide) < f.speed(partySide) && PercentileRoll(f.rng, int(f.ships[foeSide].SystemSkill)) && f.enemyCanMove
}

// dodgeAnswer is the rest of the enemy's turn after the dodge question.
func (r *Rules) dodgeAnswer(yes bool) {
	f := r.space
	f.stage = foeIdle
	r.ramEnd(foeSide, yes)
	f.boostPending = r.boostTest(f.strategy)
	r.enemyFire()
}

// enemyFire is EnemyTurnBody from the pilot's shots on: the pilot mounts
// fire out, the crew reloads or fires (a tenth of it, rounded up, repairs),
// then the repair passes boost or jury-rig.
func (r *Rules) enemyFire() {
	f := r.space
	f.damageSlot, f.damage, f.reporting, f.reportPending = -1, 0, false, false
	for {
		f.firingSlot = r.bestWeapon(foeSide, false)
		if f.damageSlot >= 0 && f.damageSlot != f.firingSlot {
			r.reportEnemyDamage()
		}
		if f.firingSlot >= 0 {
			r.fireWeapon(foeSide, f.firingSlot, nil, anySystem)
		}
		f.damageSlot = f.firingSlot
		if f.firingSlot < 0 || f.end != shipOn {
			break
		}
	}
	crew := uint16(f.ships[foeSide].Crew)
	repairs := int8((crew + 9) / 10)
	fires := int8(crew - uint16(repairs))
	for fires > 0 && f.end == shipOn {
		fires--
		if f.hasReloadable(foeSide) {
			r.reloadWeapon(foeSide, nil, -1)
			continue
		}
		f.firingSlot = r.bestWeapon(foeSide, true)
		if f.damageSlot >= 0 && f.damageSlot != f.firingSlot {
			r.reportEnemyDamage()
		}
		if f.firingSlot >= 0 {
			r.fireWeapon(foeSide, f.firingSlot, nil, anySystem)
		} else {
			repairs += fires
			fires = 0
		}
		f.damageSlot = f.firingSlot
	}
	f.reporting = true
	if f.damageSlot >= 0 {
		r.reportEnemyDamage()
	}
	for repairs > 0 && f.end == shipOn {
		repairs--
		if f.boostPending {
			r.attemptSpeedBoost(foeSide, nil)
			f.boostPending = false
		} else if !r.attemptJuryRig(foeSide, nil, 0, 0) {
			repairs = 0
		}
	}
	if f.enemyRigged {
		f.say("The enemy jury-rigs some systems.")
	}
	f.enemyRigged = false
}

// roundEnd is ShipCombat's round end: the range clamped at 0, the enemy free
// to manoeuvre under 4, lost past the scanners; the dying count on
// (ProcessDyingCharacters); the round's range is the new one.
func (r *Rules) roundEnd() {
	f := r.space
	if f.end == shipOn {
		switch {
		case f.combatRange < 0:
			f.combatRange = 0
		case f.combatRange < 4:
			f.enemyCanMove = true
		case f.combatRange > scannerRange:
			f.say("The enemy ship recedes off your scanners.")
			f.end, f.exp = receded, 0
		}
	}
	count := 0
	for _, id := range f.crew {
		c := r.chars[id]
		if c.Status != StDying {
			continue
		}
		if f.dying[id]++; f.dying[id] > 9 {
			c.Status = StComatose
			f.say(c.Name + " slips into a coma!")
		} else {
			count++
		}
	}
	switch {
	case count > 1:
		f.say("Your Teammates are Dying!")
	case count > 0:
		f.say("Your Teammate is Dying!")
	}
	f.rangeNow = f.combatRange
}

// startCrew is PlayerTurnBody's start.
func (r *Rules) startCrew() {
	f := r.space
	f.crewOn, f.turn = true, 0
	f.canManeuver = f.canMove(partySide)
	f.commandTaken = false
}

// crewActor is the member at the menu: the next one in order who is okay,
// while the fight goes on.
func (r *Rules) crewActor() rules.CharID {
	f := r.space
	if f == nil || !f.crewOn || f.end != shipOn || f.stage != foeIdle {
		return rules.NoChar
	}
	for ; f.turn < len(f.crew); f.turn++ {
		if r.chars[f.crew[f.turn]].Status == StOkay {
			return f.crew[f.turn]
		}
	}
	return rules.NoChar
}

// shipPatients is FindPatients outside a fight: every team member in play.
func (r *Rules) shipPatients(c *Character) []rules.CharID {
	var ps []rules.CharID
	for _, id := range r.space.crew {
		if r.isPatient(c, r.chars[id]) {
			ps = append(ps, id)
		}
	}
	return ps
}

// shipAllowed is BuildPlayerMenu for the member at the menu.
func (r *Rules) shipAllowed(sf rules.ShipFight) rules.Allowed {
	id := r.crewActor()
	if id == rules.NoChar || id != sf.Actor {
		return rules.Allowed{}
	}
	f := r.space
	c := r.chars[id]
	pilot := id == f.pilot
	var a rules.Allowed
	offer := func(act rules.Action, name string) {
		if !slices.ContainsFunc(a.Offers, func(o rules.Offer) bool { return o.Action == act }) {
			a.Offers = append(a.Offers, rules.Offer{Action: act, Name: name})
		}
	}
	if r.bestWeapon(partySide, !pilot) >= 0 {
		offer(rules.ShipFire{}, "Fire")
		for i, sys := range append(slices.Clone(xxvc.ShipSystems), xxvc.Weapons) {
			offer(rules.ShipFire{Target: sys}, "Target "+[7]string{"Hull", "Sens", "Ctrl", "Life", "Fuel", "Engn", "Weap"}[i])
		}
	}
	if !pilot {
		if f.hasReloadable(partySide) {
			for i := range mounts {
				if f.canReload(partySide, i) {
					if slot := r.pickMount(partySide, hotkey(r.weapon(f.ships[partySide].Mounts[i]).Short)); slot >= 0 {
						offer(rules.ShipReload{Weapon: int32(slot)}, "Load "+r.weapon(f.ships[partySide].Mounts[i]).Short)
					}
				}
			}
		}
		if r.needsRepair(partySide, c) {
			if f.canJuryRig(partySide) {
				for i, m := range f.ships[partySide].Mounts {
					if m.Status < m.MaxStatus && uint8(f.mountRigs[0][i]) < m.MaxStatus {
						if slot := r.pickMount(partySide, hotkey(r.weapon(m).Short)); slot >= 0 {
							offer(rules.ShipJuryRig{System: xxvc.Weapons, Mount: int32(slot)}, "Juryrig "+r.weapon(m).Short)
						}
					}
				}
			}
			for _, sys := range f.rigSystems(partySide) {
				offer(rules.ShipJuryRig{System: xxvc.ShipSystems[sys]}, "Juryrig "+systemNames[sys])
			}
		}
	}
	offer(rules.EndTurn{}, "Quit")
	if pilot {
		if f.canManeuver {
			if f.rangeNow > 1 {
				offer(rules.ShipClose{}, "Close")
			}
			offer(rules.ShipRetreat{}, "Withdraw")
			if f.rangeNow <= 1 {
				offer(rules.ShipRam{}, "Ram")
				if f.speed(foeSide) == 0 {
					offer(rules.ShipBoard{}, "Board")
				}
			}
		}
	} else {
		if !f.commandTaken {
			offer(rules.ShipCommand{}, "Command")
		}
		if len(r.shipPatients(c)) > 0 {
			offer(rules.Aid{}, "Aid")
		}
		if skillValue(c, RepairEngine) > 0 && f.ships[partySide].Status[sysEngines] > 0 && f.boosts[partySide] < 2 {
			offer(rules.ShipBoost{System: xxvc.Engines}, "Boost engines")
		}
		if f.ships[partySide].Status[sysSensors] > 0 && c.Skills.Level(sensorOperation) > 0 {
			offer(rules.ShipSensors{}, "Sensors")
		}
	}
	return a
}

// errNotOffered refuses an action the menu does not hold.
var errNotOffered = errors.New("rulesxxvc: not on the ship menu")

// systemIndex is a ShipSystem's index (sysWeapons for xxvc.Weapons, anySystem for "").
func systemIndex(s rules.ShipSystem) int {
	if s == "" {
		return anySystem
	}
	if s == xxvc.Weapons {
		return sysWeapons
	}
	return slices.Index(xxvc.ShipSystems, s)
}

// shipResolve is PlayerTurnBody's case for the member at the menu, and the
// answer to the dodge question.
func (r *Rules) shipResolve(sf rules.ShipFight, act rules.Action, rng *rules.RNG) (rules.Outcome, error) {
	f := r.space
	f.rng = rng
	if d, ok := act.(rules.ShipDodge); ok {
		if f.stage != foeAsking {
			return rules.Outcome{}, errNotOffered
		}
		r.dodgeAnswer(d.Yes)
		switch {
		case f.tailPending:
			f.tailPending = false
			r.roundEnd()
		case f.crewAfter:
			f.crewAfter = false
			r.startCrew()
		}
		return f.flush(), nil
	}
	confirmed := false
	if c, ok := act.(rules.Confirmed); ok {
		act, confirmed = c.Action, true
	}
	allowed := r.shipAllowed(sf)
	if !slices.ContainsFunc(allowed.Offers, func(o rules.Offer) bool { return o.Action == act }) {
		return rules.Outcome{}, errNotOffered
	}
	id := sf.Actor
	c := r.chars[id]
	taken := true
	switch a := act.(type) {
	case rules.Aid:
		for _, p := range r.shipPatients(c) {
			f.out = append(f.out, r.treatOne(c, p, nil, rng)...)
		}
	case rules.ShipBoard:
		f.end = boarded
		f.say("You close with the injured vessel and attempt to board!")
	case rules.ShipBoost:
		r.attemptSpeedBoost(partySide, c)
	case rules.ShipClose:
		r.closeShip(partySide)
		f.canManeuver, taken = false, false
	case rules.ShipCommand:
		f.say(c.Name + " takes over as pilot.")
		f.pilot, c.Pilot, f.commandTaken = id, true, true
		taken = false
	case rules.ShipFire:
		target := systemIndex(a.Target)
		if target != anySystem {
			f.say("Target which system?")
		}
		for {
			f.firingSlot = r.bestWeapon(partySide, id != f.pilot)
			if f.firingSlot >= 0 {
				r.fireWeapon(partySide, f.firingSlot, c, target)
			}
			if f.firingSlot < 0 || f.end != shipOn || id != f.pilot {
				break
			}
		}
	case rules.ShipJuryRig:
		r.attemptJuryRig(partySide, c, systemIndex(a.System), int(a.Mount))
	case rules.ShipReload:
		r.reloadWeapon(partySide, c, int(a.Weapon))
	case rules.EndTurn:
	case rules.ShipRam:
		if !confirmed { // YesNoMenu's question: its ship name draws each time it is asked
			o := f.flush()
			o.Confirm = "Ram " + r.formatShipName(foeSide, nil) + "? "
			return o, nil
		}
		r.ramStart(partySide, c) // the enemy rammed: no question
		r.ramEnd(partySide, false)
	case rules.ShipRetreat:
		r.retreatShip(partySide, f.speed(partySide))
		f.canManeuver, taken = false, false
	case rules.ShipSensors:
		g, _ := SkillCheck([]int8{skillValue(c, sensorOperation)}, rules.Easy, rng)
		outcome := "fails"
		if g >= rules.Made {
			outcome = "succeeds"
		}
		f.say(c.Name + " " + outcome + " at sensor operation!")
	}
	if taken {
		f.turn++
	}
	return f.flush(), nil
}

// repairShip is postcbt.pas RepairShip: the experience shared over the
// enabled members; each system but fuel below its maximum repaired by the
// first okay member in order to make its repair skill (Normal), the dice
// repairDice d10; each missing mount point fixed by the first to make
// Repair Weapon; then fuel plundered from the enemy when experience was
// shared. It answers the share.
func (r *Rules) repairShip(f *shipFight) uint16 {
	var enabled []*Character
	for _, id := range f.crew {
		if c := r.chars[id]; c.Enabled {
			enabled = append(enabled, c)
		}
	}
	var share uint16
	if len(enabled) > 0 {
		share = f.exp / uint16(len(enabled))
		for _, c := range enabled {
			c.Exp += int64(share)
		}
	}
	f.exp = 0
	f.say("  Repairing Ship")
	s := &f.ships[partySide]
	repairSkills := [6]SkillNo{RepairHull, RepairElectrical, RepairMechanical, RepairLifeSupport, 0, RepairEngine}
	repairDice := [6]int{20, 4, 4, 8, 0, 12}
	okay := func(id rules.CharID) bool { return r.chars[id].Status == StOkay }
	made := func(c *Character, n SkillNo) bool {
		g, _ := SkillCheck([]int8{skillValue(c, n)}, rules.Normal, f.rng)
		return g >= rules.Made
	}
	for i := range s.Status {
		if i == sysFuel || s.Status[i] >= s.Maximum[i] {
			continue
		}
		for _, id := range f.crew {
			if !okay(id) || !made(r.chars[id], repairSkills[i]) {
				continue
			}
			amount := uint8(rules.RNGDice(f.rng, repairDice[i], 10))
			if int16(amount)+s.Status[i] > s.Maximum[i] {
				amount = uint8(s.Maximum[i] - s.Status[i])
			}
			s.Status[i] += int16(amount)
			f.say(r.chars[id].Name + " " + systemNames[i] + " " + strconv.Itoa(int(amount)))
			break
		}
	}
	for i := range s.Mounts {
		m := &s.Mounts[i]
		for remain := m.MaxStatus - m.Status; remain > 0; remain-- {
			for _, id := range f.crew {
				if okay(id) && made(r.chars[id], repairWeapon) {
					f.say(r.chars[id].Name + " " + r.weapon(*m).Name + " fixed")
					m.Status++
					break
				}
			}
		}
	}
	if share != 0 {
		foeFuel := int16(f.ships[foeSide].Status[sysFuel])
		fuel := uint16(int16(rules.RNGDice(f.rng, 2, 4)+7)*foeFuel) >> 4
		if fuel > 0 {
			fuel += uint16(rules.RNGDice(f.rng, 1, 16))
		}
		if room := int32(s.Maximum[sysFuel] - s.Status[sysFuel]); int32(fuel) > room {
			fuel = uint16(room)
		}
		if fuel > 0 {
			f.say(strconv.Itoa(int(fuel)) + " units of fuel plundered from enemy ship!")
			s.Status[sysFuel] += int16(fuel)
		}
	}
	f.ships[foeSide].Status[sysFuel] = 0
	f.say("  All repairs completed.")
	return share
}
