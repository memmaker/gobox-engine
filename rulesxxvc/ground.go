package rulesxxvc

import (
	"fmt"
	"image"
	"math"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

func errNoAreaWeapon(id rules.CharID) error {
	return fmt.Errorf("AttackGroundTarget %d: no readied area weapon with its projectile kind set (special 2 and 3)", id)
}

// saveReduction is C_Save: what a made save does to a projectile's damage or
// effect.
type saveReduction uint8

const (
	saveNoEffect saveReduction = iota
	saveCancels
	saveHalves
)

// projectileKind is gdconsts.pas projectileKinds' row (TProjectileKindInfo).
type projectileKind struct {
	effect    EffectID // given to each creature hit, 0 none
	radius    int
	lingering bool
	save      int
	reduction saveReduction
}

// The projectile kinds MakeAttack launches (a weapon's special 2, while its
// special 3 is set); 4 is in the table but never launched.
const (
	firstProjectileKind = 5
	lastProjectileKind  = 12
	dazzleKind          = 5
	gasKind             = 7
	chaffKind           = 8
	mistKind            = 9
	stunKind            = 12
)

// projectileKinds is gdconsts.pas projectileKinds[4..12].
var projectileKinds = [lastProjectileKind + 1]projectileKind{
	4:  {EffHelpless, 1, true, 1, saveCancels},
	5:  {EffBlind, 1, false, 3, saveCancels},
	6:  {0, 1, false, 2, saveHalves},
	7:  {EffHelpless, 1, true, 1, saveCancels},
	8:  {0, 1, true, 7, saveNoEffect},
	9:  {0, 1, true, 7, saveNoEffect},
	10: {0, 3, false, 2, saveHalves},
	11: {0, 2, false, 2, saveHalves},
	12: {EffStunned, 1, false, 0, saveCancels},
}

// effectStatusNames is gdconsts.pas effectStatusNames: "is <name>" when the
// effect takes.
var effectStatusNames = map[EffectID]string{EffHelpless: "Helpless", EffScared: "Scared", EffBlind: "Blind", EffStunned: "Stunned", EffDizzy: "Dizzy"}

// blockSteps is gdconsts.pas blockCellDirections over dirStepX/Y: a 3x3
// block, centre first, then east clockwise.
var blockSteps = [9]rules.BoardPos{{}, {X: 1}, {X: 1, Y: 1}, {Y: 1}, {X: -1, Y: 1}, {X: -1}, {X: -1, Y: -1}, {Y: -1}, {X: 1, Y: -1}}

// cloud is types.pas TAreaEffect: a lingering cloud's marker over the 3x3
// block round at, on the cells that were passable, over the terrain beneath.
type cloud struct {
	at        rules.BoardPos
	active    [9]bool
	under     [9]rules.TerrainID
	countdown int
	marker    rules.TerrainID
}

// corpse is a downedCreatures record: the terrain beneath its cell.
type corpse struct {
	id      rules.CharID
	at      rules.BoardPos
	terrain rules.TerrainID
}

// arena is the rules' part of GDVars.combatArena for the fight: the cells
// changed from Board.Terrain (the overlay), areaEffects (newest first) and
// downedCreatures. BeginFight and EndFight clear it.
type arena struct {
	cells   map[rules.BoardPos]rules.TerrainID
	clouds  []*cloud
	corpses []corpse
}

// board is THE seam for the fight's board: b with the overlay laid over its
// terrain, a copy the action may change. A Board with no size is none handed
// in (the shared engine passes none): the ruleset's own, r.ground, is the
// board then. A Board with a size is how the unit tests hand one in.
func (r *Rules) board(b rules.Board) rules.Board {
	if noBoard(b) {
		b = r.ground
	}
	t := make([]rules.TerrainID, b.Size.X*b.Size.Y)
	if len(b.Terrain) == len(t) {
		copy(t, b.Terrain)
	} else {
		for i := range t {
			t[i] = openTerrain
		}
	}
	for p, v := range r.arena.cells {
		if onBoard(b, p) {
			t[p.Y*b.Size.X+p.X] = v
		}
	}
	b.Terrain = t
	return b
}

// ownsBoard is whether the fight f runs on the ruleset's own board.
func ownsBoard(f rules.Fight) bool { return noBoard(f.Board) }

// noBoard is a Board nobody laid: no size, no units, not zero-g (the engine's
// Fight carries only its Straight flag).
func noBoard(b rules.Board) bool { return b.Size == (image.Point{}) && len(b.Units) == 0 && !b.ZeroG }

// keepOnGround writes what an action's Effects did to the units back into
// r.ground (the one place the units on it change): Walked ends on its path's
// last square, Turned faces, Placed stands, Left and Fled are gone.
func (r *Rules) keepOnGround(effects []rules.Effect) {
	us := slices.Clone(r.ground.Units)
	at := func(id rules.CharID) int {
		return slices.IndexFunc(us, func(u rules.BoardUnit) bool { return u.ID == id })
	}
	for _, e := range effects {
		switch e := e.(type) {
		case rules.Walked:
			if i := at(charOf(e.Unit)); i >= 0 && len(e.Path) > 0 {
				us[i].At = e.Path[len(e.Path)-1]
			}
		case rules.Turned:
			if i := at(charOf(e.Unit)); i >= 0 {
				us[i].Facing = e.Facing
			}
		case rules.Placed:
			if i := at(e.Unit); i >= 0 {
				us[i].At = e.At
			}
		case rules.Left:
			us = slices.DeleteFunc(us, func(u rules.BoardUnit) bool { return u.ID == e.Unit })
		case rules.Fled:
			us = slices.DeleteFunc(us, func(u rules.BoardUnit) bool { return u.ID == e.Unit })
		}
	}
	r.ground.Units = us
}

func onBoard(b rules.Board, p rules.BoardPos) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < b.Size.X && p.Y < b.Size.Y
}

// setCell writes the arena cell (combatArena^.cells[y, x] :=); a change is a
// TerrainSet.
func (r *Rules) setCell(b *rules.Board, p rules.BoardPos, t rules.TerrainID, o *rules.Outcome) {
	if !onBoard(*b, p) || terrainAt(*b, p) == t {
		return
	}
	if r.arena.cells == nil {
		r.arena.cells = map[rules.BoardPos]rules.TerrainID{}
	}
	r.arena.cells[p] = t
	b.Terrain[p.Y*b.Size.X+p.X] = t
	o.Effects = append(o.Effects, rules.TerrainSet{At: p, Terrain: t})
}

func isCloudMarker(t rules.TerrainID) bool {
	return t == chaffTerrain || t == gasTerrain || t == aerosolMistTerrain
}

// resolveTerrain is cbtmap.pas ResolveTerrain: the terrain beneath a cloud
// marker (the last cloud in the list that covers p) or a corpse.
func (r *Rules) resolveTerrain(b rules.Board, p rules.BoardPos) rules.TerrainID {
	t := terrainAt(b, p)
	switch {
	case isCloudMarker(t):
		for _, c := range r.arena.clouds {
			for i, on := range c.active {
				if on && c.at.Add(blockSteps[i]) == p {
					t = c.under[i]
				}
			}
		}
	case t == corpseTerrain:
		for _, d := range r.arena.corpses {
			if d.at == p {
				t = d.terrain
			}
		}
	}
	return t
}

// setTerrain is cbtmap.pas SetTerrain: beneath every cloud cell and corpse on
// p, or on the arena when there is neither.
func (r *Rules) setTerrain(b *rules.Board, p rules.BoardPos, t rules.TerrainID, o *rules.Outcome) {
	found := false
	for _, c := range r.arena.clouds {
		for i, on := range c.active {
			if on && c.at.Add(blockSteps[i]) == p {
				c.under[i], found = t, true
			}
		}
	}
	for i := range r.arena.corpses {
		if r.arena.corpses[i].at == p {
			r.arena.corpses[i].terrain, found = t, true
		}
	}
	if !found {
		r.setCell(b, p, t, o)
	}
}

// layCloudMarkers is FxEngine.LayAreaEffectMarkers.
func (r *Rules) layCloudMarkers(b *rules.Board, o *rules.Outcome) {
	for _, c := range r.arena.clouds {
		for i, on := range c.active {
			if on {
				r.setCell(b, c.at.Add(blockSteps[i]), c.marker, o)
			}
		}
	}
}

// countDownClouds is FxEngine.CountDownAreaEffects (EndOfRound): a cloud at
// countdown 1 or less clears (its cells back to the terrain beneath, or a
// corpse), the others count down; after a clearing the rest are laid again.
func (r *Rules) countDownClouds(b *rules.Board, o *rules.Outcome) {
	cleared := false
	kept := r.arena.clouds[:0:0]
	for _, c := range r.arena.clouds {
		if c.countdown > 1 {
			c.countdown--
			kept = append(kept, c)
			continue
		}
		for i, on := range c.active {
			if !on {
				continue
			}
			p := c.at.Add(blockSteps[i])
			t := c.under[i]
			if slices.ContainsFunc(r.arena.corpses, func(d corpse) bool { return d.at == p }) {
				t = corpseTerrain
			}
			r.setCell(b, p, t, o)
		}
		cleared = true
	}
	r.arena.clouds = kept
	if cleared {
		r.layCloudMarkers(b, o)
		o.Effects = append(o.Effects, rules.Narrated{Text: "The air clears a little..."})
	}
}

// layCorpses is CbtMap.CreatureDowned's and CombatSt.PlaceAllCreatures'
// corpse: each fighter down (not out of play) that has none yet leaves
// CORPSE_TERRAIN on its cell, over the terrain there, unless a cloud's marker
// lies on it. at gives the fallen's cell.
func (r *Rules) layCorpses(b *rules.Board, ids []rules.CharID, at func(rules.CharID) (rules.BoardPos, bool), o *rules.Outcome) {
	for _, id := range ids {
		c := r.chars[id]
		p, ok := at(id)
		if c == nil || c.Enabled || c.Status == StFleeing || !ok ||
			slices.ContainsFunc(r.arena.corpses, func(d corpse) bool { return d.id == id }) {
			continue
		}
		r.arena.corpses = append(r.arena.corpses, corpse{id, p, r.resolveTerrain(*b, p)})
		if !isCloudMarker(terrainAt(*b, p)) {
			r.setCell(b, p, corpseTerrain, o)
		}
	}
}

// layFallen lays the corpses of the units an outcome felled (Fell), in
// order, each on its cell after the outcome's walks.
func (r *Rules) layFallen(b *rules.Board, o *rules.Outcome) {
	var fell []rules.CharID
	cells := map[rules.CharID]rules.BoardPos{}
	for _, u := range b.Units {
		cells[u.ID] = u.At
	}
	for _, e := range o.Effects {
		switch e := e.(type) {
		case rules.Walked:
			if len(e.Path) > 0 {
				cells[charOf(e.Unit)] = e.Path[len(e.Path)-1]
			}
		case rules.Fell:
			fell = append(fell, charOf(e.Unit))
		}
	}
	r.layCorpses(b, fell, func(id rules.CharID) (rules.BoardPos, bool) { p, ok := cells[id]; return p, ok }, o)
}

// considerTerrain is FxEngine.ConsiderTerrainEffects after a step: an
// enabled c on gas, not helpless, without breathing protection, failing save
// 1 is helpless for 1d6+1 and its actions end.
func (r *Rules) considerTerrain(b rules.Board, id rules.CharID, c *Character, s *combatState, rng *rules.RNG) []rules.Effect {
	i := slices.IndexFunc(b.Units, func(u rules.BoardUnit) bool { return u.ID == id })
	if !c.Enabled || i < 0 || !slices.ContainsFunc(cellsOf(b.Units[i]), func(p rules.BoardPos) bool { return terrainAt(b, p) == gasTerrain }) ||
		helpless(c) || c.Armor.Breathing || rollSave(c, 1, 0, rng) {
		return nil
	}
	var out []rules.Effect
	if r.addEffect(c, c, Effect{ID: EffHelpless, Minutes: rules.RNGDice(rng, 1, 6) + 1}, false) {
		out = append(out, rules.Narrated{Text: c.Name + " is helpless"})
	}
	if helpless(c) {
		endActions(s)
	}
	return out
}

// scatterOffset is CombatLaunchProjectile's ScatterOffset: d4 div 2 cells
// further out, or a random way beyond 8 cells or at 0. Both dice are rolled.
func scatterOffset(v int, rng *rules.RNG) int {
	amount := rules.RNGDice(rng, 1, 4) >> 1
	way := rules.RNGDice(rng, 1, 2)*2 - 3
	switch {
	case v < -8 || v > 8:
		return v + amount*way
	case v < 0:
		return v - amount
	case v > 0:
		return v + amount
	}
	return v + way
}

// The scatter roll's limit: attackBonus + 1d20 - ProjectileRangePenalty at
// or below it scatters the shot.
const (
	scatterLimit = 50
	scatterTries = 4
	scanReach    = 3 // ScanProjectileArea's 7x7 block round the target
)

// groundAttack is cbtacts.pas AttackGroundTarget, MakeAttack with no
// defender and CombatLaunchProjectile for c's readied area weapon at to.
func (r *Rules) groundAttack(f rules.Fight, id rules.CharID, c *Character, s *combatState, to rules.BoardPos, rng *rules.RNG) (rules.Outcome, error) {
	var o rules.Outcome
	b := &f.Board
	r.combatEndRound = r.combatRound + idleRounds // MakeAttack
	s.target = rules.NoChar                       // MakeAttack: target := defender
	wi := r.readiedIndex(c, xxvc.SlotWeapon)
	if wi < 0 {
		return o, errNoAreaWeapon(id)
	}
	ci := c.Items[wi]
	wt, wty, _ := r.itemType(ci)
	kind := int(wt.Special[1])
	if wty.WeaponClass < xxvc.RocketLauncher || wt.Special[2] == 0 || kind < firstProjectileKind || kind > lastProjectileKind {
		return o, errNoAreaWeapon(id) // MakeAttack: 'Weapon specials incorrectly set!'
	}
	pk := projectileKinds[kind]
	damage := int(uint8(rules.RNGDice(rng, int(wty.Damage.N), int(wty.Damage.Sides)) + int(wty.Damage.Bonus) + plus(ci, wt)))
	minutes := 255
	switch kind {
	case dazzleKind:
		minutes = rules.RNGDice(rng, 1, 6) + 1
	case gasKind, stunKind:
		minutes, damage = damage, 0
	}
	launched := true
	switch wty.WeaponClass {
	case xxvc.RocketLauncher, xxvc.Plasma:
		c.Items[wi].Readied = false
	case xxvc.Grenade:
		launched = false
		if li := r.readiedIndex(c, xxvc.SlotGrenadeLauncher); li >= 0 {
			c.Items[li].Readied, launched = false, true
		}
	}
	if launched {
		o.Effects = append(o.Effects, rules.Narrated{Text: c.Name + " launches a missile..."})
	} else {
		o.Effects = append(o.Effects, rules.Narrated{Text: c.Name + " throws a grenade..."})
	}
	o.Effects = append(o.Effects, rules.Used{From: unitOf(id), What: r.itemName(ci.Ref)})
	// ATTACK_ROLL_ATTACKER's BLIND lowers c's armour until its next
	// RecalcStats; every attack on c recalculates first, so it changes nothing.
	r.RecalcStats(c)
	src, _ := unitAt(f, id)
	dx, dy := rules.Abs(to.X-src.X), rules.Abs(to.Y-src.Y)
	if dx > dy {
		dx += dy >> 1
	} else {
		dx = dy + dx>>1
	}
	msg := ""
	if c.AttackBonus+rules.RNGDice(rng, 1, 20)-ProjectileRangePenalty(dx, r.readiedWeaponRange(c)) <= scatterLimit {
		tries, dist := 0, 4 // WalkLine's dist is a var: each try starts from the last one's
		var p rules.BoardPos
		for {
			p.X = scatterOffset(to.X-src.X, rng) + src.X
			p.Y = scatterOffset(to.Y-src.Y, rng) + src.Y
			tries++
			reached := false
			if onBoard(*b, p) {
				_, dist, reached, _ = walkLine(*b, src, p, dist, true, false)
			}
			if reached || tries >= scatterTries {
				break
			}
		}
		if tries < scatterTries { // a hit on the last try is dropped, as in the original
			to, msg = p, "It goes astray"
		}
	}
	if stop, _, ok, seen := walkLine(*b, src, to, 255, true, true); !ok && seen.chaff {
		if msg == "" { // sic: the original's test is the wrong way round
			msg = " and"
		} else {
			msg = "It"
		}
		msg += " hits chaff"
		to = stop
	}
	o.Effects = append(o.Effects, rules.Shot{From: unitOf(id), Start: src, To: to, Item: ci.Ref.ID})

	leaper, leapFrom, leapTo := rules.NoChar, rules.BoardPos{}, rules.BoardPos{X: -1, Y: -1}
	if kind == dazzleKind || kind == 6 || kind == 10 || kind == 11 {
		for _, fid := range r.fighters(f) {
			if p := r.chars[fid]; p.Enabled && hasEffect(p, EffUnkillable) {
				leaper = fid
				leapFrom, _ = unitAt(f, fid)
				break
			}
		}
		if kind != dazzleKind {
			leapTo = r.scanArea(b, to, pk.radius, leaper, leapFrom, rng, &o)
		}
	}
	var targets []rules.CharID
	for _, t := range r.lineTargets(*b, []rules.BoardPos{to}, pk.radius, true, func(*Character) bool { return true }) {
		if t.id != leaper || leapTo.X == -1 {
			targets = append(targets, t.id)
			continue
		}
		// Felix: the leap is a walk along the line to the chosen cell, not a jump.
		var path []rules.BoardPos
		for l := rules.NewLine(leapFrom.X, leapFrom.Y, leapTo.X, leapTo.Y); l.Step(); {
			path = append(path, rules.BoardPos{X: l.X, Y: l.Y})
		}
		o.Effects = append(o.Effects, rules.Narrated{Text: r.chars[leaper].Name + " leaps for cover!"}, rules.Walked{Unit: unitOf(leaper), Path: path, Pace: rules.PaceFast})
		if u := slices.IndexFunc(b.Units, func(u rules.BoardUnit) bool { return u.ID == leaper }); u >= 0 {
			b.Units = slices.Clone(b.Units)
			b.Units[u].At = leapTo
		}
	}
	animateDraws(kind, pk.radius, to, rng)
	if msg != "" {
		o.Effects = append(o.Effects, rules.Narrated{Text: msg})
	}
	if pk.lingering {
		r.layCloud(b, kind, to, f.Board.ZeroG, &o)
	}
	for _, did := range targets {
		d := r.chars[did]
		if !d.Enabled {
			continue
		}
		saved := kind == dazzleKind && d.Armor.Goggles || kind == gasKind && d.Armor.Breathing
		if !saved {
			saved = rollSave(d, pk.save, 0, rng)
		}
		if damage > 0 {
			o.Effects = append(o.Effects, r.inflictNonWeapon(c, id, did, d, damage, pk.reduction, r.itemName(ci.Ref), saved)...)
		}
		if d.Enabled && pk.effect > 0 && r.addEffect(d, c, Effect{ID: pk.effect, Minutes: minutes}, saved && pk.reduction == saveCancels) {
			o.Effects = append(o.Effects, rules.Narrated{Text: d.Name + " is " + effectStatusNames[pk.effect]})
		}
	}
	endActions(s)
	if c.Items[wi].Ref.Count > 0 { // wi still the weapon fired: nothing was removed since
		c.Items[wi].Ref.Count--
		if c.Items[wi].Ref.Count == 0 && wty.Consumable {
			c.Items = slices.Delete(c.Items, wi, wi+1)
		}
	}
	return o, nil
}

// scanArea is CombatLaunchProjectile's ScanProjectileArea (distMode 1): each
// cell of the 7x7 block round to that a line from to reaches within radius
// takes its blastResult (250..253: rubble 146 on a 25% roll, with 147 or 153
// below on another; else 152, 149, 150, 253 kept); with a leaper, the free
// passable cell outside the blast its line reaches soonest is chosen
// ({-1, -1} none).
func (r *Rules) scanArea(b *rules.Board, to rules.BoardPos, radius int, leaper rules.CharID, from rules.BoardPos, rng *rules.RNG, o *rules.Outcome) rules.BoardPos {
	best, sel := 255, rules.BoardPos{X: -1, Y: -1}
	for ox := -scanReach; ox <= scanReach; ox++ {
		for oy := -scanReach; oy <= scanReach; oy++ {
			p := rules.BoardPos{X: to.X + ox, Y: to.Y + oy}
			if !onBoard(*b, p) {
				continue
			}
			if _, _, ok, _ := walkLine(*b, to, p, radius, false, false); ok {
				t := terrainRow(r.resolveTerrain(*b, p)).blast
				if t == 0 {
					continue
				}
				var below rules.TerrainID
				if t >= 250 && t <= 253 {
					if PercentileRoll(rng, 25) {
						t, below = 146, 153
						if PercentileRoll(rng, 25) {
							below = 147
						}
					} else {
						switch t {
						case 250:
							t = 152
						case 251:
							t = 149
						case 252:
							t = 150
						}
					}
				}
				r.setTerrain(b, p, t, o)
				// The original writes the cell below even past the arena's last row
				// (cells[25, x], past the heap block; tracks/X.md X10); here it is dropped.
				if q := p.Add(rules.BoardPos{Y: 1}); below != 0 && onBoard(*b, q) {
					r.setTerrain(b, q, below, o)
				}
			} else if leaper != rules.NoChar {
				if _, taken := r.occupant(*b, p); taken || terrainRow(terrainAt(*b, p)).cost == impassable {
					continue
				}
				if _, d, ok, _ := walkLine(*b, from, p, 255, false, false); ok && d < best {
					best, sel = d, p
				}
			}
		}
	}
	return sel
}

// layCloud is CombatLaunchProjectile's lingering cloud: its marker on each
// passable cell of the 3x3 block round to, for 4 rounds (1 in zero gravity).
func (r *Rules) layCloud(b *rules.Board, kind int, to rules.BoardPos, zeroG bool, o *rules.Outcome) {
	c := &cloud{at: to, countdown: 4}
	if zeroG {
		c.countdown = 1
	}
	switch kind {
	case gasKind:
		c.marker = gasTerrain
	case chaffKind:
		c.marker = chaffTerrain
	case mistKind:
		c.marker = aerosolMistTerrain
	}
	for i, d := range blockSteps {
		p := to.Add(d)
		if !onBoard(*b, p) {
			continue
		}
		t := r.resolveTerrain(*b, p)
		c.under[i] = t
		c.active[i] = t > 0 && terrainRow(t).cost < impassable
		if c.active[i] {
			r.setCell(b, p, c.marker, o)
		}
	}
	r.arena.clouds = append([]*cloud{c}, r.arena.clouds...)
}

// inflictNonWeapon is FxEngine.InflictNonWeaponDamage from attacker a:
// COND_NON_WEAPON_DAMAGE (IMMUNE_TO_HEAT_AND_PLASMA against a's readied heat
// or plasma weapon, UNKILLABLE's cap), then a made save cancels or halves.
func (r *Rules) inflictNonWeapon(a *Character, aid, did rules.CharID, d *Character, damage int, red saveReduction, by string, saved bool) []rules.Effect {
	if _, _, wty, armed := r.readied(a, xxvc.SlotWeapon); armed && hasEffect(d, EffImmuneToHeatAndPlasma) &&
		(wty.WeaponClass == xxvc.Heat || wty.WeaponClass == xxvc.Plasma) {
		damage = 0
	}
	if hasEffect(d, EffUnkillable) && damage > d.HP {
		damage = int(uint8(d.HP))
	}
	if saved {
		switch red {
		case saveCancels:
			damage = 0
		case saveHalves:
			damage /= 2
		}
	}
	if damage <= 0 || !d.Enabled {
		return nil
	}
	out := []rules.Effect{rules.Struck{Unit: unitOf(did), Damage: int32(damage), From: unitOf(aid), By: by}}
	if InflictDamage(d, damage) {
		ds := r.fight[did]
		if ds != nil {
			ds.initiative = 0
		}
		removeIncapacitating(d, ds)
		out = append(out, rules.Fell{Unit: unitOf(did)})
	}
	return out
}

// arenaCols and arenaRows are types.pas TArena.cells [0..24, 0..49]: the
// original's fixed arena, whatever the Board's size.
const (
	arenaCols = 50
	arenaRows = 25
)

// burstOffset is AnimateProjectileArea's offsetX/offsetY: the burst pushed
// three tiles per cell the radius overhangs an arena edge (size cells).
func burstOffset(v, size, radius int) int {
	switch {
	case v < radius:
		return (v - radius) * 3
	case v > size-1-radius:
		return (v - (size - 1) + radius) * 3
	}
	return 0
}

// animateDraws is AnimateProjectileArea's draws from the one RandSeed: per
// ring 1..radius*3+1 of the burst, 2*ring-1 particles, each a Random real
// (one step; the angle is State/2^32 turns, as Turbo Pascal's Random) and,
// for an animated kind drawn inside the 19x19 view, a RollDice(1, 4).
func animateDraws(kind, radius int, to rules.BoardPos, rng *rules.RNG) {
	animated := false
	switch kind {
	case dazzleKind, 6, 10, 11:
		animated = true
	case gasKind, chaffKind, mistKind, stunKind:
	default:
		return
	}
	ox, oy := burstOffset(to.X, arenaCols, radius), burstOffset(to.Y, arenaRows, radius)
	for ring := 1; ring <= radius*3+1; ring++ {
		for n := 0; n < ring*2-1; n++ {
			rng.N(0)
			angle := float64(rng.State) / (1 << 32) * 2 * math.Pi
			x := int(math.Cos(angle)*float64(ring)) + ox
			y := int(math.Sin(angle)*float64(ring)) + oy
			if animated && x >= -9 && x <= 9 && y >= -9 && y <= 9 {
				rules.RNGDice(rng, 1, 4)
			}
		}
	}
}

// unitOf is the board unit of id: rulesxxvc numbers its board units by
// CharID, so the effects' UnitID is the id (charOf the way back).
func unitOf(id rules.CharID) rules.UnitID { return rules.UnitID(id) }

// charOf is who board unit u is (unitOf).
func charOf(u rules.UnitID) rules.CharID { return rules.CharID(u) }
