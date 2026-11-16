package rules1e

// The lingering clouds (fold T3f): Stinking Cloud (POR ovr23:1B11, four
// squares) and Cloud Kill (cotab O0C2:4F95, nine), one record list
// (Board.Clouds) told apart by Kind.

// CloudKind is which cloud a record is.
type CloudKind uint8

const (
	CloudStinking CloudKind = iota
	CloudPoison
	CloudDelayedBlast // a Delayed Blast Fireball waiting on its square (spells_sotsbfirst.go)
)

// GroundIDs are the grounds the rules lay themselves, per game (rules.json
// groundIds; zero: POR's, PorGroundIDs).
type GroundIDs struct {
	Floor         Ground `json:"floor"`         // bare floor
	PoisonCloud   Ground `json:"poisonCloud"`   // Cloud Kill
	StinkingCloud Ground `json:"stinkingCloud"` // Stinking Cloud
	Body          Ground `json:"body"`          // a fallen combatant's body
	DelayedBlast  Ground `json:"delayedBlast"`  // a waiting Delayed Blast Fireball (SotSB only)
}

// The games' grounds (c_terr.pas of each -re; POR ovr32). In POR 0x1C is the
// clutter ground, so the poison test reads the records, not the ground.
var (
	PorGroundIDs   = GroundIDs{Floor: 0x17, PoisonCloud: 0x1C, StinkingCloud: 0x1E, Body: 0x1F}
	CotabGroundIDs = GroundIDs{Floor: 0x17, PoisonCloud: 28, StinkingCloud: 30, Body: 31}
	SotsbGroundIDs = GroundIDs{Floor: 0x17, PoisonCloud: 27, StinkingCloud: 29, Body: 30, DelayedBlast: 26}
	PodGroundIDs   = GroundIDs{Floor: 0x17, PoisonCloud: 28, StinkingCloud: 29, Body: 31}
)

// grounds is the game's GroundIDs.
func (e *RuleState) grounds() GroundIDs { return e.T().Grounds }

// cloudSquaresMax is the most squares a cloud covers (Cloud Kill's nine).
const cloudSquaresMax = 9

// poisonCloudDirs: Cloud Kill's nine squares from its centre (cotab cloudSquares9; 8 is the centre).
var poisonCloudDirs = [cloudSquaresMax]Dir{8, 0, 1, 2, 3, 4, 5, 6, 7}

// cloudSpec is what a kind of cloud covers, lays and is ended by.
type cloudSpec struct {
	dirs    []Dir
	ground  func(GroundIDs) Ground
	affect  AffectType
	message string
	// firstOlder: under an older cloud the first saved ground is kept, skipping
	// saved poison grounds too (cotab olderGround stopFirst); else the last.
	firstOlder bool
}

var cloudSpecs = [...]cloudSpec{
	CloudStinking:     {cloudDirs[1:], func(g GroundIDs) Ground { return g.StinkingCloud }, affStinkingCloud, "Creates a noxious cloud", false},
	CloudPoison:       {poisonCloudDirs[:], func(g GroundIDs) Ground { return g.PoisonCloud }, affCloudkill, "Creates a poisonous cloud", true},
	CloudDelayedBlast: {delayedBlastDirs[:], func(g GroundIDs) Ground { return g.DelayedBlast }, affDelayedBlastSotsb, "Sets a delayed blast fireball.", true},
}

// square is the map square of a cloud's k-th square (0-based).
func (cl *cloud) square(k int) (int, int) {
	d := cloudSpecs[cl.Kind].dirs[k]
	return int(int8(cl.X)) + DirDX[d], int(int8(cl.Y)) + DirDY[d]
}

// covers: a square of the cloud lies on x, y.
func (cl *cloud) covers(x, y int) bool {
	for k := range cloudSpecs[cl.Kind].dirs {
		if cl.Covered[k] != 0 {
			if sx, sy := cl.square(k); sx == x && sy == y {
				return true
			}
		}
	}
	return false
}

// groundUnder is the ground a cloud or a body lying on x, y covers (what
// lifting them gives back: clearCloud, the body's Ground), and whether one
// lies there.
func (e *RuleState) groundUnder(x, y int) (Ground, bool) {
	g, ids := e.Cmap.Ground[y][x], e.grounds()
	covered := false
	for range len(e.Clouds) + e.Nbodies + 1 { // a body under a cloud under an older one: each lift once
		under, found := g, false
		for i := len(e.Clouds) - 1; i >= 0 && !found; i-- { // the newest cloud on top
			cl := e.Clouds[i]
			for k := range cloudSpecs[cl.Kind].dirs {
				if sx, sy := cl.square(k); cl.Covered[k] != 0 && sx == x && sy == y && g == cloudSpecs[cl.Kind].ground(ids) && cl.Ground[k] != g {
					under, found = cl.Ground[k], true
				}
			}
		}
		for b := 1; b <= e.Nbodies && !found && g == ids.Body && ids.Body != 0; b++ {
			if bd := e.Bodies[b]; int(int8(bd.X)) == x && int(int8(bd.Y)) == y && Ground(bd.Ground) != g {
				under, found = Ground(bd.Ground), true
			}
		}
		if !found {
			break
		}
		g, covered = under, true
	}
	return g, covered
}

// olderGround: the ground an older cloud of the kind saved at x, y, else g.
func (e *RuleState) olderGround(kind CloudKind, self *cloud, x, y int, g Ground, firstOlder bool) Ground {
	for _, o := range e.Clouds {
		if o == self || o.Kind != kind {
			continue
		}
		for j := range cloudSpecs[kind].dirs {
			if o.Covered[j] == 0 {
				continue
			}
			if ox, oy := o.square(j); ox != x || oy != y {
				continue
			}
			og := o.Ground[j]
			if og == e.grounds().StinkingCloud || firstOlder && og == e.grounds().PoisonCloud {
				continue
			}
			g = og
			if firstOlder {
				return g
			}
		}
	}
	return g
}

// createCloud = POR ovr23:1B11 / cotab O0C2:27EB, O0C2:4F95: the cloud is
// laid over the target square; who stands under it breathes it.
func (e *RuleState) createCloud(kind CloudKind) {
	spec := cloudSpecs[kind]
	e.areaSpell = 1
	sel := e.Acting()
	lvl := e.casterLevel(int(e.CurSpell))
	var same []*cloud
	for _, o := range e.Clouds {
		if o.Kind == kind {
			same = append(same, o)
		}
	}
	n := 0
	// the last cloud of the list is not counted
	for i := 0; i < len(same)-1; i++ {
		if same[i].Ptr00 == sel {
			n++
		}
	}
	cl := &cloud{Kind: kind}
	e.Clouds = append(e.Clouds, cl)
	e.AddAffect(sel, spec.affect, lvl, byte(lvl+n<<4), 1)
	cl.Ptr00 = sel
	cl.Num = byte(n)
	cl.X = byte(e.TgtX)
	cl.Y = byte(e.TgtY)
	who := make([]Slot, len(spec.dirs))
	for k := range spec.dirs {
		x, y := cl.square(k)
		w, g := e.MapAt(x, y)
		who[k] = w
		if g > 0 && e.GroundInfo(g)[0] < 0xFF {
			cl.Covered[k] = 1
		}
		switch {
		case g == e.grounds().StinkingCloud:
			g = e.olderGround(CloudStinking, cl, x, y, g, spec.firstOlder)
		case g == e.grounds().PoisonCloud && kind == CloudPoison:
			g = e.olderGround(CloudPoison, cl, x, y, g, true)
		case g == e.grounds().Body:
			// a body lies there: keep the ground under it
			for b := 1; b <= e.Nbodies; b++ {
				if int(int8(e.Bodies[b].X)) == x && int(int8(e.Bodies[b].Y)) == y {
					g = Ground(e.Bodies[b].Ground)
				}
			}
		}
		cl.Ground[k] = g
		if cl.Covered[k] != 0 {
			e.Cmap.Ground[y][x] = spec.ground(e.grounds())
		}
	}
	e.CombatSay(sel, spec.message, true)
	for _, w := range who {
		if w > 0 {
			e.inClouds(e.Combatant(byte(w)))
		}
	}
}

// clearCloud = POR ovr13:0CC3 / cotab O064:0C1D, O064:1AB9 (char, affect):
// the caster's cloud of that number lifts; the squares get back their ground
// (or a body's), and the other clouds of the kind are laid again.
func (e *RuleState) clearCloud(kind CloudKind, c *Character, a *Affect) {
	num := a.Data >> 4
	ci := -1
	for i, cl := range e.Clouds {
		if cl.Kind == kind && cl.Ptr00 == c && cl.Num == num {
			ci = i
			break
		}
	}
	if ci < 0 {
		return
	}
	cl := e.Clouds[ci]
	e.notice("The air clears a little...")
	set := func(x, y int, v Ground) { // the original does not check the bounds
		if x >= 0 && y >= 0 && y < len(e.Cmap.Ground) && x < len(e.Cmap.Ground[0]) {
			e.Cmap.Ground[y][x] = v
		}
	}
	spec := cloudSpecs[kind]
	for k := range spec.dirs {
		if cl.Covered[k] == 0 {
			continue
		}
		x, y := cl.square(k)
		body := false
		for i := 1; i <= e.Nbodies; i++ {
			if b := &e.Bodies[i]; b.C != nil && int(int8(b.X)) == x && int(int8(b.Y)) == y {
				body = true
			}
		}
		if body {
			set(x, y, e.grounds().Body)
		} else {
			set(x, y, cl.Ground[k])
		}
	}
	e.Clouds = append(e.Clouds[:ci:ci], e.Clouds[ci+1:]...)
	for _, p := range e.Clouds {
		if p.Kind != kind {
			continue
		}
		for k := range spec.dirs {
			if p.Covered[k] != 0 {
				x, y := p.square(k)
				set(x, y, spec.ground(e.grounds()))
			}
		}
	}
}

// clearStinkingCloud = ovr13:0CC3.
func (e *RuleState) clearStinkingCloud(c *Character, a *Affect, flag byte) {
	e.clearCloud(CloudStinking, c, a)
}

// clearPoisonCloud = cotab O064:1AB9 (the "cloudkill" handler).
func (e *RuleState) clearPoisonCloud(c *Character, a *Affect, flag byte) {
	e.clearCloud(CloudPoison, c, a)
}

// inClouds: c breathes the clouds it stands in.
func (e *RuleState) inClouds(c *Character) {
	e.StinkingCloud(c)
	e.inPoisonCloud(c)
}

// underPoisonCloud: a square c covers lies under a poison cloud.
func (e *RuleState) underPoisonCloud(c *Character) bool {
	idx := e.CombatIndex(c)
	for part := 0; part <= 3; part++ {
		dx, dy, ok := e.Footprint(int(e.Cpos[idx].Size), part)
		if !ok {
			continue
		}
		for _, cl := range e.Clouds {
			if cl.Kind == CloudPoison && cl.covers(int(e.Cpos[idx].X)+dx, int(e.Cpos[idx].Y)+dy) {
				return true
			}
		}
	}
	return false
}

// Cloud Kill by hit dice (Felix's ruling): up to cloudKillSlain slain; at
// cloudKillSaveMinus a save vs poison at cloudKillMalus, at cloudKillSavePlain
// a plain one, or slain; above it cloudKillDamageSides poison damage each round.
const (
	cloudKillSlain       = 4
	cloudKillSaveMinus   = 5
	cloudKillSavePlain   = 6
	cloudKillMalus       = -4
	cloudKillDamageSides = 10
)

// inPoisonCloud = cotab O0E3:0E47's poison half: in a poison cloud the weak
// die and the strong are hurt.
func (e *RuleState) inPoisonCloud(c *Character) {
	if c == nil || c.Okay == 0 || !e.underPoisonCloud(c) {
		return
	}
	switch hd := int(c.MaxLevel); {
	case hd <= cloudKillSlain:
	case hd == cloudKillSaveMinus:
		if e.SavingThrow(c, SaveVsParalysis, cloudKillMalus) {
			return
		}
	case hd == cloudKillSavePlain:
		if e.SavingThrow(c, SaveVsParalysis, 0) {
			return
		}
	default:
		e.Damage(c, e.Dice(1, cloudKillDamageSides), 0, false)
		return
	}
	e.CombatSay(c, "is Poisoned", true)
	e.AddAffect(c, AffPoisoned, 0, 0xFF, 0)
	e.killWithMessage(c, StatusDead, "is killed")
}
