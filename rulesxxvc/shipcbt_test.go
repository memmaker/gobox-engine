package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// shipTables are gdconsts.pas shipWeapons and the BR1 SHIPS records 40
// (Hubble) and 50 (Maelstrom Rider), as importbr types them.
func shipTables() xxvc.Tables {
	w := map[xxvc.ShipWeaponKind]xxvc.ShipWeapon{
		1: {MaxRange: 6, Damage: 10, TargetBonus: 1, Name: "Beam laser", Short: "Laser"},
		2: {MaxRange: 4, Damage: 20, Name: "Pumped laser", Short: "P-laser"},
		3: {MaxRange: 4, ReloadAmmo: 50, Damage: 10, TargetBonus: 1, Name: "Gyrocannon", Short: "Gyrocann"},
		4: {MaxRange: 6, ReloadAmmo: 4, Damage: 40, Name: "Missile mount", Short: "Missile"},
		5: {MaxRange: 6, ReloadAmmo: 4, Damage: 60, TargetBonus: -1, Name: "Heavy missile mount", Short: "Hv-missl"},
		6: {MaxRange: 4, ReloadAmmo: 6, Damage: 100, TargetBonus: -3, Name: "K-cannon", Short: "K-cannon"},
		7: {MaxRange: 2, ReloadAmmo: 15, Damage: 20, TargetBonus: -1, Name: "Light acceleration gun", Short: "acl-Cann"},
		8: {MaxRange: 2, ReloadAmmo: 10, Damage: 30, TargetBonus: -2, Name: "Heavy acceleration gun", Short: "hv-Accel"},
	}
	full := func(s ...int16) [6]int16 { return [6]int16(s) }
	return xxvc.Tables{ShipWeapons: w, Ships: map[rules.ShipID]xxvc.ShipRecord{
		40: {Name: "Hubble", Tonnage: 50, Status: full(200, 60, 65, 100, 160, 150), Maximum: full(200, 60, 65, 100, 160, 150),
			Crew: 20, ArmorClass: 6, AlternateSkill: 100, SystemSkill: 60, SystemValue: 50, BaseSpeed: 3, Reaction: -1,
			Maneuver: 254, Type: "Scout cruiser", Race: xxvc.Martian,
			Mounts: [8]xxvc.ShipMount{{Kind: 1, Crew: true, MaxStatus: 2, Status: 2}, {Kind: 4, MaxStatus: 2, Status: 2, Ammo: 4, Reloads: 7}}},
		50: {Name: "Maelstrom Rider", Tonnage: 150, Status: full(600, 150, 150, 300, 450, 450), Maximum: full(600, 150, 150, 300, 450, 450),
			Crew: 6, ArmorClass: 6, AlternateSkill: 100, SystemSkill: 70, SystemValue: 50, BaseSpeed: 3, Type: "Md cruiser", Race: xxvc.Terran,
			Mounts: [8]xxvc.ShipMount{{Kind: 6, MaxStatus: 2, Status: 2, Ammo: 6, Reloads: 5}, {Kind: 4, MaxStatus: 2, Status: 2, Ammo: 4, Reloads: 7},
				{Kind: 1, Crew: true, MaxStatus: 5, Status: 5}}},
	}}
}

func shipFightOf(t *testing.T, seed uint32) (*Rules, rules.ShipFight, rules.CharID, rules.CharID) {
	t.Helper()
	r := New()
	r.Begin(rules.Campaign{Tables: shipTables()}, rules.Options{})
	sc := map[rules.Ability]int{rules.Dexterity: 16}
	a := r.Add(Character{Name: "Ace", Enabled: true, Status: StOkay, HP: 30, MaxHP: 30, BaseCombat: 10, Level: 3,
		Career: xxvc.RocketJock, Scores: sc, Pilot: true})
	b := r.Add(Character{Name: "Bo", Enabled: true, Status: StOkay, HP: 30, MaxHP: 30, BaseCombat: 8, Scores: sc, Pilot: true})
	rng := rules.RNG{State: seed}
	sf, err := r.Ships().BeginShips(rules.ShipEncounter{Foe: 40, PilotSkill: 80, ToHit: 15}, &rng)
	if err != nil {
		t.Fatal(err)
	}
	return r, sf, a, b
}

// ShipCombat's set-up: the first marked member flies, a later mark is
// cleared; the ECL arguments set the foe's alternateSkill and systemValue
// (60 - to-hit); the pilot contest is 1d10 - 1d10 + 0 - (-1) > 0.
func TestBeginShips(t *testing.T) {
	for seed := uint32(1); seed < 40; seed++ {
		r, sf, a, b := shipFightOf(t, seed)
		f := r.space
		if f.pilot != a || r.chars[b].Pilot || !r.chars[a].Pilot {
			t.Fatalf("pilot %d, marks %v %v", f.pilot, r.chars[a].Pilot, r.chars[b].Pilot)
		}
		if f.ships[foeSide].AlternateSkill != 80 || f.ships[foeSide].SystemValue != 45 || f.rangeNow != 8 || sf.Range != 8 {
			t.Fatalf("foe %+v range %d", f.ships[foeSide], f.rangeNow)
		}
		x := rules.RNG{State: seed}
		want := rules.RNGDice(&x, 1, 10)-rules.RNGDice(&x, 1, 10)+1 > 0
		if f.enemyFirst != want {
			t.Fatalf("seed %d: enemyFirst %v", seed, f.enemyFirst)
		}
	}
}

// CharHlp.ShipSpeed: base plus boosts, halved under half hull, -3 under a
// quarter of the engines, -1 under half, at least 1; 0 when it cannot move.
func TestShipSpeed(t *testing.T) {
	r, _, _, _ := shipFightOf(t, 1)
	f := r.space
	s := &f.ships[partySide]
	if f.speed(partySide) != 3 {
		t.Fatal(f.speed(partySide))
	}
	f.boosts[partySide] = 2
	s.Status[sysHull] = 299 // < 600 shr 1
	if f.speed(partySide) != 2 {
		t.Fatal("halved", f.speed(partySide))
	}
	s.Status[sysEngines] = 112 // < 450 shr 2
	if f.speed(partySide) != 1 {
		t.Fatal("engines", f.speed(partySide))
	}
	s.Status[sysFuel] = 0
	if f.speed(partySide) != 0 {
		t.Fatal("no fuel")
	}
}

// The pilot's Fire: every ready pilot mount fires, best damage first, at
// range 4 the K-cannon and the missile twice each (status 2; the beam
// laser is the crew's).
// Each shot: d20 against TargetValue; a hit flashes (d8, d3, d3), picks a
// system (d100), and the crew roll (damage / 2) follows. Replayed here.
func TestPilotFire(t *testing.T) {
	for seed := uint32(1); seed < 60; seed++ {
		r, _, a, _ := shipFightOf(t, seed)
		f := r.space
		f.rangeNow = 4
		r.startCrew()
		x := *f.rng
		sf := r.shipSnapshot()
		if sf.Actor != a {
			t.Fatalf("actor %d", sf.Actor)
		}
		f.ships[foeSide].Status[sysHull] = 2000 // the hull holds: all four shots fire
		foe, mountPoints := f.ships[foeSide], 4
		if _, err := r.Resolve(sf, rules.ShipFire{}, f.rng); err != nil {
			t.Fatal(err)
		}
		// TargetValue for Ace: 10 + dex 16's +1 + level 3 + 3 (party) + bonus - 2 (range 3..4),
		// + maneuver 254 (-2), 60 - 6 - that.
		for _, shot := range []struct {
			damage int16
			bonus  int
		}{{100, -3}, {100, -3}, {40, 0}, {40, 0}} { // each mount fires its status' times
			need := min(20, 60-6-(10+1+3+3+shot.bonus-2-2)) // at most 20
			if rules.RNGDice(&x, 1, 20) >= need {
				rules.RNGDice(&x, 1, 8)
				rules.RNGDice(&x, 1, 3)
				rules.RNGDice(&x, 1, 3)
				sys := HitSystem(&x)
				if sys == sysWeapons {
					rules.RNGDice(&x, 1, mountPoints) // which point
					rules.RNGDice(&x, 1, 4)           // the destroyed message's ship name
					mountPoints--
					CrewLost(&x, int(shot.damage), 1)
				} else {
					if foe.Status[sys] == 0 {
						sys = sysHull
					}
					v := min(shot.damage, foe.Status[sys])
					foe.Status[sys] -= v
					CrewLost(&x, int(v), 1)
				}
			}
		}
		if x.State != f.rng.State {
			t.Fatalf("seed %d: the draws differ", seed)
		}
		if f.shots[partySide][0] != 2 || f.shots[partySide][1] != 2 || f.ships[partySide].Mounts[0].Ammo != 4 || f.turn != 1 {
			t.Fatalf("seed %d: shots %v ammo %d turn %d", seed, f.shots[partySide], f.ships[partySide].Mounts[0].Ammo, f.turn)
		}
	}
}

// The menus' weapon pick: the last mount whose short name starts with the
// key; 'acl-Cann' (key C) never matches, so it is never offered.
func TestPickMount(t *testing.T) {
	r, _, _, _ := shipFightOf(t, 1)
	f := r.space
	f.ships[partySide].Mounts[3] = xxvc.ShipMount{Kind: 1, MaxStatus: 1, Status: 1}
	if got := r.pickMount(partySide, hotkey("Laser")); got != 3 {
		t.Errorf("Laser picks %d, want the last, 3", got)
	}
	f.ships[partySide].Mounts[4] = xxvc.ShipMount{Kind: 7, Crew: true, MaxStatus: 2, Status: 1}
	if got := r.pickMount(partySide, hotkey("acl-Cann")); got != -1 {
		t.Errorf("acl-Cann picks %d", got)
	}
}

// The enemy's jury-rig pick: the damaged mount of the most damage; else
// the system scored (10 - 10*now/max, +5 destroyed) x2, the highest.
func TestEnemyJuryRigPick(t *testing.T) {
	r, _, _, _ := shipFightOf(t, 3)
	f := r.space
	s := &f.ships[foeSide]
	s.Status[sysSensors] = 30 // 10 - 5 = 5, x2 10
	s.Status[sysFuel] = 0     // 10 + 5 = 15, x2 30: fuel too, for the enemy
	x := *f.rng
	if !r.attemptJuryRig(foeSide, nil, 0, 0) {
		t.Fatal("nothing rigged")
	}
	made := ShipSkillCheck(&x, 60, 0, rules.Normal)
	amount := int16(0)
	if made {
		amount = int16(rules.RNGDice(&x, 2, 10))
	}
	if s.Status[sysFuel] != amount || f.systemRigs[1][sysFuel] != amount || x.State != f.rng.State {
		t.Fatalf("fuel %d rigs %d, want %d", s.Status[sysFuel], f.systemRigs[1][sysFuel], amount)
	}
	s.Mounts[1].Status = 1
	f.mountRigs[1] = [mounts]points{}
	before := s.Status
	r.attemptJuryRig(foeSide, nil, 0, 0)
	if s.Status != before { // the missile mount, not a system
		t.Fatal("a system rigged over a damaged mount")
	}
}

// The enemy rams the party's moving ship: the dodge question pauses the
// turn; the answer goes on with the Pilot Rocket check (bonus (1-3)*10);
// a ram does the enemy's 50 tons to the party's hull.
func TestDodgeQuestion(t *testing.T) {
	r, sf, _, _ := shipFightOf(t, 9)
	f := r.space
	f.enemyFirst = false
	f.rangeNow = 1
	f.ships[partySide].Status[sysHull] = 100
	f.ships[partySide].Tonnage = 50
	f.ships[foeSide].Mounts = [8]xxvc.ShipMount{} // nothing to fire: ram
	rng := rules.RNG{State: 77}
	sf, o := r.Ships().ShipFoe(sf, &rng)
	if f.stage != foeAsking || f.tailPending != true {
		t.Fatalf("no question: %+v", o)
	}
	x := rng
	if _, err := r.Resolve(sf, rules.ShipDodge{Yes: true}, &rng); err != nil {
		t.Fatal(err)
	}
	dodged := ShipSkillCheck(&x, int(skillValue(r.chars[f.pilot], pilotRocket)), (1-3)*10, rules.Normal) // speeds: 3 halved under half hull, 3
	if !dodged && f.ships[partySide].Status[sysHull] != 50 {
		t.Fatalf("not dodged, hull %d", f.ships[partySide].Status[sysHull])
	}
	if f.stage != foeIdle || f.tailPending {
		t.Fatal("still asking")
	}
}

// The round's end: past the scanners the enemy is lost with the experience.
func TestRecede(t *testing.T) {
	r, sf, _, _ := shipFightOf(t, 2)
	f := r.space
	f.enemyFirst = true
	f.combatRange = 9
	rng := rules.RNG{State: 5}
	r.Ships().ShipFoe(sf, &rng)
	if v := r.Ships().ShipVerdict(sf); v != rules.PartyFled || f.exp != 0 {
		t.Fatalf("verdict %v exp %d", v, f.exp)
	}
}

// RepairShip: 300 experience shared by two; a hull hit repaired by the first
// to make Repair Hull, 20d10 capped at the maximum; no enemy fuel, nothing
// plundered.
func TestRepairShip(t *testing.T) {
	r, sf, a, b := shipFightOf(t, 4)
	f := r.space
	f.end = foeDestroyed
	f.ships[partySide].Status[sysHull] = 590
	f.ships[foeSide].Status[sysFuel] = 0
	r.chars[a].Skills = SkillLevels{}
	rng := rules.RNG{State: 11}
	x := rng
	res := r.Ships().EndShips(sf, &rng)
	if res.Exp != 150 || r.chars[a].Exp != 150 || r.chars[b].Exp != 150 {
		t.Fatalf("exp %d", res.Exp)
	}
	hull := int16(590)
	for range 2 { // each member's Repair Hull roll until one makes it
		if g, _ := SkillCheck([]int8{0}, rules.Normal, &x); g >= rules.Made {
			hull = min(600, hull+int16(rules.RNGDice(&x, 20, 10)))
			break
		}
	}
	rules.RNGDice(&x, 2, 4) // fuel: (2d4+7)*0 shr 4 = 0
	if r.ship.Status[sysHull] != hull || x.State != rng.State {
		t.Fatalf("hull %d, want %d", r.ship.Status[sysHull], hull)
	}
	if r.space != nil {
		t.Fatal("fight kept")
	}
}
